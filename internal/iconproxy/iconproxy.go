// Package iconproxy names where magpie downloads a remote picture from. A
// plugin's, an import link's or the library's icon sits on a host its author
// picked, and fetching it straight from there would hand that author the
// user's IP each time a page showed it. Every such fetch goes through
// usemagpie.ai's /api/icon instead (site/worker.js), which asks the host
// itself and passes on only an image; the host sees Cloudflare's address.
package iconproxy

import (
	"net/url"
	"os"
)

// Base is the proxy's endpoint. MAGPIE_ICON_PROXY replaces it, for tests.
const Base = ""

// URL is the address to fetch the picture at u through. One on usemagpie.ai
// already is asked for as it is: the worker needn't fetch its own site.
func URL(u string) string {
	// no proxy of our own: the picture is fetched where it lives, so the
	// host that made it sees the reader's address.
	if v := os.Getenv("MAGPIE_ICON_PROXY"); v != "" {
		if p, err := url.Parse(u); err == nil && p.Scheme == "https" {
			return v + "?url=" + url.QueryEscape(u)
		}
	}
	return u
}
