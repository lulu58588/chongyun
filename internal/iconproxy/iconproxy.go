// Package iconproxy names where starling downloads a remote picture from. A
// plugin's, an import link's or the library's icon sits on a host its author
// picked, and fetching it straight from there would hand that author the
// user's IP each time a page showed it. Every such fetch goes through the
// proxy STARLING_ICON_PROXY names (none by default), which asks the host
// itself and passes on only an image.
package iconproxy

import (
	"net/url"
	"os"
)

// Base is the proxy's endpoint. STARLING_ICON_PROXY replaces it, for tests.
const Base = ""

// URL is the address to fetch the picture at u through. Without a proxy of
// our own the picture is fetched where it lives.
func URL(u string) string {
	// no proxy of our own: the picture is fetched where it lives, so the
	// host that made it sees the reader's address.
	if v := os.Getenv("STARLING_ICON_PROXY"); v != "" {
		if p, err := url.Parse(u); err == nil && p.Scheme == "https" {
			return v + "?url=" + url.QueryEscape(u)
		}
	}
	return u
}
