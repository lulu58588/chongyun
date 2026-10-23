package iconproxy

import (
	"net/url"
	"testing"
)

// This build has no icon proxy of its own: the picture is fetched where it
// lives, so the host that made it sees the reader's address.
func TestURLIsThePictureItself(t *testing.T) {
	t.Setenv("STARLING_ICON_PROXY", "")
	pic := "https://example.com/logo.png?size=96&x=a b#c"
	if got := URL(pic); got != pic {
		t.Fatalf("URL(%q) = %q, want it unchanged", pic, got)
	}
}

// A proxy of the user's own, when one is named, is asked instead.
func TestURLUsesTheProxyWhenNamed(t *testing.T) {
	t.Setenv("STARLING_ICON_PROXY", "https://proxy.example/api/icon")
	pic := "https://example.com/logo.png"
	got := URL(pic)
	u, err := url.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if u.Scheme+"://"+u.Host+u.Path != "https://proxy.example/api/icon" {
		t.Fatalf("URL(%q) = %q, not on the proxy", pic, got)
	}
	if q := u.Query(); q.Get("url") != pic {
		t.Fatalf("URL(%q) = %q: query %v", pic, got, q)
	}
}
