package cli

import "testing"

// starling web on loopback behind a reverse proxy (#1479) prints the proxy's
// link with the key when told the proxy's address, so nobody joins
// "/?k=…" onto it by hand: STARLING_WEB_URL, else STARLING_PUBLIC_URL with no
// port (one with a port isn't a proxy). None is printed without either.
func TestWebProxyLink(t *testing.T) {
	key := "/?k=0123456789abcdef"
	for _, c := range []struct{ env, want string }{
		{"https://starling.example.com", "https://starling.example.com/?k=0123456789abcdef"},
		{"https://starling.example.com/", "https://starling.example.com/?k=0123456789abcdef"},
		{"starling.example.com", "http://starling.example.com/?k=0123456789abcdef"},
		{"https://example.com/starling/", "https://example.com/starling/?k=0123456789abcdef"},
		{"http://192.168.1.20:3425", ""},
		{"", ""},
	} {
		t.Setenv("STARLING_PUBLIC_URL", c.env)
		if got := proxyLink(key); got != c.want {
			t.Errorf("STARLING_PUBLIC_URL=%q: %q, want %q", c.env, got, c.want)
		}
	}
	// the page's own proxy address, when the gateway's is elsewhere
	t.Setenv("STARLING_PUBLIC_URL", "https://api.example.com")
	for _, c := range []struct{ env, want string }{
		{"https://starling.example.com/", "https://starling.example.com/?k=0123456789abcdef"},
		{"starling.example.com:8443", "https://starling.example.com:8443/?k=0123456789abcdef"},
		{"https://starling.example.com/?x=1", ""},
		{"ftp://starling.example.com", ""},
	} {
		t.Setenv("STARLING_WEB_URL", c.env)
		if got := proxyLink(key); got != c.want {
			t.Errorf("STARLING_WEB_URL=%q: %q, want %q", c.env, got, c.want)
		}
	}
}
