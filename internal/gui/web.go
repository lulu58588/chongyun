package gui

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"html"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/klauspost/compress/gzhttp"

	"github.com/lulu58588/starling/internal/gateway"
	"github.com/lulu58588/starling/internal/i18ntext"
	"github.com/lulu58588/starling/internal/update"
)

// `starling web` serves the window's page to a browser, for a computer that
// can't show the app — no desktop, WSL, a server reached over SSH — with
// the gateway beside it as the app has. What is the desktop's (the panel,
// the tray, the Dock) does nothing there; a folder is named, not opened.

// webHost is the Windows a browser tab stands for.
type webHost struct{ quit func() }

func (webHost) HidePanel()                       {}
func (webHost) ShowMain(string)                  {}
func (h webHost) Quit()                          { h.quit() }
func (webHost) OpenURL(string) error             { return nil } // the page opens links itself
func (webHost) Copy(string) bool                 { return false }
func (webHost) FitPanel(int, Glide)              {}
func (webHost) TintPanel([4]uint8, int) bool     { return false }
func (webHost) TintTitleBar([4]uint8, bool) bool { return false }
func (webHost) SetTextSize(int)                  {} // the browser zooms its own tab
func (webHost) OpenFolder(path string) error {
	return errors.New("in the browser starling can't open folders: it is " + tilde(path))
}
func (webHost) ChooseFolder(string) (string, error) {
	return "", errors.New("in the browser starling can't show a folder picker: type the folder's path")
}

// Web is a started `starling web`: its address and the link that signs a
// browser in.
type Web struct {
	Addr, Link string
	key        string
	srv        *http.Server
	quit       chan struct{}
}

// webReexec is set when the page's restart to update has put the new
// version in: Wait then runs it in this one's place.
var webReexec atomic.Bool

// webRunKey hands a run's own key to the version it restarts into, so the
// browser's cookie still opens the page (#111).
const webRunKey = "STARLING_WEB_RUNKEY"

// StartWeb serves the page on addr (host:port), as the given version of
// starling, which keeps itself up to date as the app does. Every request needs the
// key the link carries, taken once into a cookie other sites' pages can't
// send with a POST (every change is one), so neither a page elsewhere nor
// anyone else on the network reaches the settings and keys behind it.
func StartWeb(addr, version string) (*Web, error) {
	Version = version
	key, fixed, err := webKey()
	if err != nil {
		return nil, err
	}
	var keep time.Duration
	if fixed {
		keep = webCookieAge
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	// with --tls-cert, the page is HTTPS to other computers, as the
	// gateway is: it carries the settings and keys
	if ln, err = gateway.ServeTLS(ln); err != nil {
		return nil, err
	}
	w := &Web{Addr: ln.Addr().String(), key: key, quit: make(chan struct{})}
	host, port, _ := net.SplitHostPort(w.Addr)
	if ip := net.ParseIP(host); ip == nil || ip.IsUnspecified() {
		host = "127.0.0.1"
	}
	var once sync.Once
	quit := func() { once.Do(func() { close(w.quit) }) }
	// on loopback the page takes plain HTTP with TLS on too, which a
	// browser here opens without a certificate for 127.0.0.1
	scheme := "http"
	if ip := net.ParseIP(host); ip != nil && !ip.IsLoopback() {
		scheme = gateway.Scheme()
	}
	w.Link = scheme + "://" + net.JoinHostPort(host, port) + "/?k=" + url.QueryEscape(key)
	w.srv = &http.Server{
		Handler:           gateway.AllowGuard(webHandler(port, key, keep, Handler(webHost{quit}, startBackend()))),
		ReadHeaderTimeout: 30 * time.Second,
	}
	go w.srv.Serve(ln)
	// Settings' AutoRestart: starling web runs on as the new version (#2096)
	updates.autoRestart = func() bool {
		if restartInto(true, false, "", true) {
			quit()
			return true
		}
		return false
	}
	updates.start()
	news.start()
	return w, nil
}

// ServesGateway says whether this process serves the gateway, which it
// doesn't while another starling (the service, the app, starling serve) has
// it: the page then shows that one's settings and accounts, read from the
// same files, and this one takes the gateway up once that one is gone.
func (w *Web) ServesGateway() bool { return served.Load() != nil }

// Wait returns once the page's Quit is used. After a restart to update it
// runs the new version in this one's place, with the same arguments and
// key, so the page open in the browser finds it where it was.
func (w *Web) Wait() error {
	<-w.quit
	time.Sleep(200 * time.Millisecond) // the answer to the Quit gets out
	w.srv.Close()
	if !webReexec.Load() {
		return nil
	}
	exe, err := update.Executable()
	if err != nil {
		return err
	}
	return update.Reexec(exe, os.Args[1:], append(os.Environ(), webRunKey+"="+w.key))
}

// webKeyMin is the shortest STARLING_WEB_KEY taken: it is all that stands
// between the network and the keys behind the page.
const webKeyMin = 16

// webCookieAge is how long a browser keeps a fixed key's cookie (the most
// Chrome allows); a run's own key lives in a cookie that goes with the
// browser session, as the key goes with the run.
const webCookieAge = 400 * 24 * time.Hour

// webKey is the key the link carries: a new one each run, unless
// STARLING_WEB_KEY names one — for a page kept running as a service, which
// browsers then stay signed in to across restarts. fixed says which.
func webKey() (key string, fixed bool, err error) {
	if k := os.Getenv(webRunKey); k != "" {
		os.Unsetenv(webRunKey) // the run's alone, not its children's
		if len(k) == 32 && strings.Trim(k, "0123456789abcdef") == "" && os.Getenv("STARLING_WEB_KEY") == "" {
			return k, false, nil
		}
	}
	if k := os.Getenv("STARLING_WEB_KEY"); k != "" {
		if err := CheckWebKey(k); err != nil {
			return "", false, err
		}
		return k, true, nil
	}
	return NewWebKey(), false, nil
}

// NewWebKey is a new random key for the page.
func NewWebKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// CheckWebKey says what is wrong with k as a STARLING_WEB_KEY, nil when
// nothing is.
func CheckWebKey(k string) error {
	if len(k) < webKeyMin {
		return fmt.Errorf("STARLING_WEB_KEY is %d characters; it needs at least %d", len(k), webKeyMin)
	}
	// a cookie can't hold a space, ; " \ or a comma: the browser would
	// be sent one that isn't the key, and turned away
	if strings.Trim(k, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~") != "" {
		return errors.New("STARLING_WEB_KEY takes letters, digits and - . _ ~ only (openssl rand -hex 16 makes one)")
	}
	return nil
}

// webGuard lets through requests with the key: in the link's k, which it
// trades for a cookie (kept for keep, or the browser session when 0) and
// a clean address, or in that cookie. A page opened without it gets a
// form to paste the key into (webSignIn), which says why it was turned
// away, so a bookmark doesn't send anyone back to the start-up log.
func webGuard(cookie, key string, keep time.Duration, next http.Handler) http.Handler {
	same := func(v string) bool { return subtle.ConstantTimeCompare([]byte(v), []byte(key)) == 1 }
	signIn := func(rw http.ResponseWriter, r *http.Request, back string) {
		// Priority=High (Chrome's): the page's host is often an address
		// other services share on other ports, and a browser keeps a
		// limited number of cookies for one host, all ports together;
		// past it Chrome drops the low and medium ones first (#2033)
		rw.Header().Add("Set-Cookie", (&http.Cookie{Name: cookie, Value: key, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: int(keep / time.Second)}).String()+"; Priority=High")
		http.Redirect(rw, r, back, http.StatusSeeOther)
	}
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// a plugin's page's files: a sandboxed frame's requests carry no
		// cookie, and its token, which only a signed-in /api hands out,
		// is the key (servePluginPage serves nothing without a good one)
		if strings.HasPrefix(r.URL.Path, pluginPagePrefix) {
			next.ServeHTTP(rw, r)
			return
		}
		if q := r.URL.Query(); q.Has("k") {
			if !same(q.Get("k")) {
				http.Error(rw, "this link's key is not this starling web's: use the link it printed when it started", http.StatusUnauthorized)
				return
			}
			q.Del("k")
			u := *r.URL
			u.RawQuery = q.Encode()
			signIn(rw, r, u.RequestURI())
			return
		}
		if r.URL.Path == webSignInPath && r.Method == http.MethodPost {
			r.Body = http.MaxBytesReader(rw, r.Body, 4096)
			back := localPath(r.PostFormValue("back"))
			if !same(r.PostFormValue("key")) {
				webTurnedAway(rw, r, back, webText(r, "That key is not this starling web's."))
				return
			}
			signIn(rw, r, back)
			return
		}
		// A browser can hold more than one cookie of this name (one set
		// for another path or domain, an old run's): any that carries
		// the key will do, not only the first sent.
		sent := r.CookiesNamed(cookie)
		for _, c := range sent {
			if same(c.Value) {
				next.ServeHTTP(rw, r)
				return
			}
		}
		why, short := webText(r, "This browser sent no sign-in cookie for this address. A browser keeps cookies per host, not per port, so another service on this host's other ports can take it away: by clearing its site data (a sign-out often does), or by filling the host's cookie limit. A page opened at another address than the one signed in at (an IP instead of a name) has none either."), "no cookie sent"
		if len(sent) > 0 {
			why, short = webText(r, "This browser's sign-in cookie holds a key that isn't this starling web's: the key changed after it signed in (STARLING_WEB_KEY changed, or a start without it made a new one)."), "its cookie holds another key"
		}
		if r.Method != http.MethodGet || !strings.Contains(r.Header.Get("Accept"), "text/html") {
			http.Error(rw, webText(r, "open the link starling web printed when it started: it carries the key. ")+why, http.StatusUnauthorized)
			return
		}
		log.Printf("starling web: turned away %s from %s: %s", r.URL.Path, r.RemoteAddr, short)
		webTurnedAway(rw, r, r.URL.RequestURI(), why)
	})
}

// webSignInPath takes the key pasted into the form a page turned away shows.
const webSignInPath = "/_starling/sign-in"

// localPath is back when it is a path on this server, else "/": the form's
// return address never leads to another site, nor to the form's own.
func localPath(back string) string {
	u, err := url.Parse(back)
	if err != nil || u.Scheme != "" || u.Host != "" || !strings.HasPrefix(back, "/") || strings.HasPrefix(back, "//") || strings.HasPrefix(back, "/\\") || u.Path == webSignInPath {
		return "/"
	}
	return back
}

// webTurnedAway is the 401 page a browser opening the page without the key
// sees: why, and a field for the key that signs it in and comes back to
// back. The key itself is never on it.
func webTurnedAway(rw http.ResponseWriter, r *http.Request, back, why string) {
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	rw.Header().Set("Cache-Control", "no-store")
	rw.WriteHeader(http.StatusUnauthorized)
	fmt.Fprintf(rw, `<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>starling web</title>
<style>body{font:15px/1.5 system-ui,sans-serif;max-width:32em;margin:12vh auto;padding:0 16px;color:#222;background:#fafafa}@media(prefers-color-scheme:dark){body{color:#ddd;background:#1c1c1e}}input,button{font:inherit;padding:.4em .6em}input{width:100%%;box-sizing:border-box;margin:.5em 0}</style></head>
<body><h1>starling web</h1><p>%s</p><p>%s</p>
<form method="post" action="%s"><input type="hidden" name="back" value="%s"><input type="password" name="key" autocomplete="current-password" autofocus required><button type="submit">%s</button></form></body></html>`,
		html.EscapeString(why),
		html.EscapeString(webText(r, "Paste the key (STARLING_WEB_KEY, or the k= in the link starling web printed when it started):")),
		webSignInPath, html.EscapeString(back),
		html.EscapeString(webText(r, "Sign in")))
}

// webText is the text of a page starling web serves itself, in the language
// the browser asks for or the one the user set. The pages the app serves are
// translated by the front end (assets/i18n.js); these ones are plain HTML
// with no script, so their text is here.
func webText(r *http.Request, en string) string {
	return webTextIn(i18ntext.Lang(r), en)
}

// webTextIn is webText for a language already decided, so the fallback order
// can be tried without a request.
func webTextIn(lang, en string) string {
	return i18ntext.Pick(lang, en, webTextZh, webTextZhTW, webTextJa, webTextDe)
}

// The text of starling web's own pages, in Chinese. The key is the English
// as written in the code, so a page that says something new falls back to
// English rather than to nothing.
var webTextZh = map[string]string{
	"This browser sent no sign-in cookie for this address. A browser keeps cookies per host, not per port, so another service on this host's other ports can take it away: by clearing its site data (a sign-out often does), or by filling the host's cookie limit. A page opened at another address than the one signed in at (an IP instead of a name) has none either.": "这个浏览器没有为当前地址发送登录 Cookie。浏览器的 Cookie 是按主机名保存的、不区分端口，所以本机其他端口上的服务可能把它挤掉：清理本站数据（退出登录常会这样），或者占满了该主机的 Cookie 额度。用另一个地址打开页面（用 IP 而不是域名）也不会有。",
	"This browser's sign-in cookie holds a key that isn't this starling web's: the key changed after it signed in (STARLING_WEB_KEY changed, or a start without it made a new one).":                                                                                                                                                                                        "这个浏览器的登录 Cookie 里装的不是本次 starling web 的密钥：登录之后密钥变了（STARLING_WEB_KEY 改了，或者某次启动没带上它、生成了新的）。",
	"That key is not this starling web's.":                                                          "这个密钥不是本次 starling web 的。",
	"open the link starling web printed when it started: it carries the key. ":                      "请打开 starling web 启动时打印的链接：里面带着密钥。",
	"Paste the key (STARLING_WEB_KEY, or the k= in the link starling web printed when it started):": "粘贴密钥（STARLING_WEB_KEY，或 starling web 启动时打印的链接里的 k=）：",
	"Sign in": "登录",
	"this link's key is not this starling web's: use the link it printed when it started": "这个链接里的密钥不是本次 starling web 的：请用它启动时打印的链接",
}

// The same, in Traditional Chinese.
var webTextZhTW = map[string]string{
	"this link's key is not this starling web's: use the link it printed when it started": "這個連結裡的密鑰不是本次 starling web 的：請用它啟動時列印的連結",
	"This browser sent no sign-in cookie for this address. A browser keeps cookies per host, not per port, so another service on this host's other ports can take it away: by clearing its site data (a sign-out often does), or by filling the host's cookie limit. A page opened at another address than the one signed in at (an IP instead of a name) has none either.": "這個瀏覽器沒有為目前位址傳送登入 Cookie。瀏覽器的 Cookie 是依主機名儲存的、不分埠號，所以本機其他埠號上的服務可能把它擠掉：清除本站資料（登出常會這樣），或者佔滿了該主機的 Cookie 額度。用另一個位址開啟頁面（用 IP 而不是網域名稱）也不會有。",
	"This browser's sign-in cookie holds a key that isn't this starling web's: the key changed after it signed in (STARLING_WEB_KEY changed, or a start without it made a new one).":                                                                                                                                                                                        "這個瀏覽器的登入 Cookie 裡裝的不是本次 starling web 的密鑰：登入之後密鑰變了（STARLING_WEB_KEY 改了，或某次啟動沒帶上它、產生了新的）。",
	"That key is not this starling web's.":                                                          "這個密鑰不是本次 starling web 的。",
	"open the link starling web printed when it started: it carries the key. ":                      "請開啟 starling web 啟動時列印的連結：裡面帶著密鑰。",
	"Paste the key (STARLING_WEB_KEY, or the k= in the link starling web printed when it started):": "貼上密鑰（STARLING_WEB_KEY，或 starling web 啟動時列印的連結裡的 k=）：",
	"Sign in": "登入",
}

// NetworkLinks are the links to the page, served on port, that other
// machines open; key is the link's "/?k=…". STARLING_PUBLIC_URL with a port
// of its own names the host the container's ports are published on, so
// the page is at that host on its own port; one without (a reverse proxy's
// https://starling.example.com, #372) is where the page is, path and all.
func NetworkLinks(port, key string) []string {
	if u, err := url.Parse(gateway.PublicURL()); err == nil && u.Host != "" && u.Port() == "" {
		return []string{strings.TrimRight(u.String(), "/") + key}
	}
	var out []string
	for _, a := range LANAddrs() {
		out = append(out, gateway.Scheme()+"://"+net.JoinHostPort(a, port)+key)
	}
	return out
}

// LANAddrs are this computer's addresses on the local network, for the
// links to print when the page is served there: STARLING_PUBLIC_URL's host
// when set (in a container, whose own addresses the network can't reach).
func LANAddrs() []string {
	if h := gateway.PublicHost(); h != "" {
		return []string{h}
	}
	var out []string
	as, _ := net.InterfaceAddrs()
	for _, a := range as {
		n, ok := a.(*net.IPNet)
		if !ok || n.IP.IsLoopback() || n.IP.To4() == nil || !n.IP.IsPrivate() {
			continue
		}
		out = append(out, n.IP.String())
	}
	return out
}

// webHandler is what `starling web` on port serves: the page and its API
// behind the key, compressed for the browser.
func webHandler(port, key string, keep time.Duration, page http.Handler) http.Handler {
	return webCompressed(webGuard("starling_web_"+port, key, keep, page))
}

// webCompressed gzips what `starling web` sends a browser that takes it: the
// page's scripts, styles and SVG icons, and the API's JSON. They went out
// as they are, about 4 MB for a first load before any data, and the
// routing trace's JSON grows with every account a request could go to; a
// browser over a 3 Mbps link to a server waited for all of it (akic404 on
// Discord). They shrink to about a quarter, the JSON to a twentieth.
// Pictures already compressed go as they are.
//
// The app's own window gets the page from inside the process, with no
// network between, so only `starling web` compresses.
//
// A gzipped file's ETag says so ("…-gzip"), as a validator of other bytes;
// asked for again with it, the file is still a 304. A few bytes of padding
// keyed to the content keep a response's compressed size from telling
// exactly what it holds (BREACH): the page carries keys.
func webCompressed(next http.Handler) http.Handler {
	wrap, err := gzhttp.NewWrapper(
		gzhttp.ContentTypeFilter(compressible),
		gzhttp.SuffixETag(gzipETag),
		gzhttp.RandomJitter(32, 0, false),
	)
	if err != nil {
		return next
	}
	gz := wrap(next)
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if tag := r.Header.Get("If-None-Match"); strings.Contains(tag, gzipETag+`"`) {
			r.Header.Set("If-None-Match", strings.ReplaceAll(tag, gzipETag+`"`, `"`))
		}
		gz.ServeHTTP(rw, r)
	})
}

// gzipETag ends the ETag of a gzipped response.
const gzipETag = "-gzip"

// compressible: text, scripts, JSON and SVG; not PNG, WebP, ICO or JPEG.
func compressible(ct string) bool {
	ct = strings.ToLower(strings.TrimSpace(ct))
	if i := strings.IndexByte(ct, ';'); i >= 0 {
		ct = strings.TrimSpace(ct[:i])
	}
	return strings.HasPrefix(ct, "text/") || strings.HasSuffix(ct, "javascript") ||
		strings.HasSuffix(ct, "json") || ct == "image/svg+xml"
}

// isWeb: the page is served to a browser by `starling web`.
func isWeb(w Windows) bool { _, ok := w.(webHost); return ok }

// The same, in Japanese.
var webTextJa = map[string]string{
	"this link's key is not this starling web's: use the link it printed when it started": "このリンクのキーは今回の starling web のものではありません：起動時に表示されたリンクを使用してください",
	"This browser sent no sign-in cookie for this address. A browser keeps cookies per host, not per port, so another service on this host's other ports can take it away: by clearing its site data (a sign-out often does), or by filling the host's cookie limit. A page opened at another address than the one signed in at (an IP instead of a name) has none either.": "このブラウザは、このアドレスのサインイン用 Cookie を送っていません。ブラウザの Cookie はポートではなくホスト単位で保存されるため、このホストの別ポートのサービスが/sites データを消去したり（サインアウトでもそうなります）、Cookie の上限に達したりすると、提示されなくなることがあります。サインインしたアドレスとは別のアドレス（名前ではなく IP）で開いたページにもありません。",
	"This browser's sign-in cookie holds a key that isn't this starling web's: the key changed after it signed in (STARLING_WEB_KEY changed, or a start without it made a new one).":                                                                                                                                                                                        "このブラウザのサインイン用 Cookie に入っているのは、今回の starling web のキーではありません：サインイン後にキーが変わりました（STARLING_WEB_KEY が変わった、あるいはそれを指定せずに起動して新しいキーが作られました）。",
	"That key is not this starling web's.":                                                          "そのキーは今回の starling web のものではありません。",
	"open the link starling web printed when it started: it carries the key. ":                      "starling web が起動時に表示したリンクを開いてください：キーが入っています。",
	"Paste the key (STARLING_WEB_KEY, or the k= in the link starling web printed when it started):": "キーを貼り付けてください（STARLING_WEB_KEY、または starling web が起動時に表示したリンクの k=）：",
	"Sign in": "サインイン",
}

// The same, in German.
var webTextDe = map[string]string{
	"this link's key is not this starling web's: use the link it printed when it started": "Der Schlüssel in diesem Link ist nicht der dieses starling-Webs: Verwenden Sie den Link, der beim Start angezeigt wurde",
	"This browser sent no sign-in cookie for this address. A browser keeps cookies per host, not per port, so another service on this host's other ports can take it away: by clearing its site data (a sign-out often does), or by filling the host's cookie limit. A page opened at another address than the one signed in at (an IP instead of a name) has none either.": "Dieser Browser hat kein Anmelde-Cookie für diese Adresse gesendet. Browser speichern Cookies pro Host, nicht pro Port; ein anderer Dienst auf den anderen Ports dieses Hosts kann es wegnehmen: durch Löschen der Website-Daten (das Abmelden tut das oft) oder weil das Cookie-Limit des Hosts gefüllt ist. Eine Seite, die unter einer anderen Adresse geöffnet wurde (eine IP statt eines Namens), hat ebenfalls keines.",
	"This browser's sign-in cookie holds a key that isn't this starling web's: the key changed after it signed in (STARLING_WEB_KEY changed, or a start without it made a new one).":                                                                                                                                                                                        "Das Anmelde-Cookie dieses Browsers enthält einen Schlüssel, der nicht zu diesem starling web gehört: Der Schlüssel hat sich nach der Anmeldung geändert (STARLING_WEB_KEY wurde geändert, oder ein Start ohne ihn hat einen neuen erzeugt).",
	"That key is not this starling web's.":                                                          "Dieser Schlüssel gehört nicht zu diesem starling web.",
	"open the link starling web printed when it started: it carries the key. ":                      "Öffnen Sie den Link, den starling web beim Start ausgegeben hat: Er enthält den Schlüssel. ",
	"Paste the key (STARLING_WEB_KEY, or the k= in the link starling web printed when it started):": "Schlüssel einfügen (STARLING_WEB_KEY oder das k= im Link, den starling web beim Start ausgegeben hat):",
	"Sign in": "Anmelden",
}
