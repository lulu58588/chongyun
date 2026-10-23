package gui

import (
	"slices"
	"testing"
)

// starling web behind a reverse proxy (#372): STARLING_PUBLIC_URL with no port
// of its own is where the page is, so the link printed for other machines
// is it, scheme, path and all — not http://<its host>:<the page's port>,
// which nothing answers. One with a port names the host the container's
// ports are published on, and the page is there on its own port.
func TestNetworkLinksPublicURL(t *testing.T) {
	key := "/?k=0123456789abcdef"
	for _, c := range []struct{ env, want string }{
		{"https://starling.example.com", "https://starling.example.com/?k=0123456789abcdef"},
		{"https://starling.example.com/", "https://starling.example.com/?k=0123456789abcdef"},
		{"https://nas.example.com/starling/", "https://nas.example.com/starling/?k=0123456789abcdef"},
		{"http://192.168.1.20:3425", "http://192.168.1.20:3430/?k=0123456789abcdef"},
		{"nas.local:13425", "http://nas.local:3430/?k=0123456789abcdef"},
	} {
		t.Setenv("STARLING_PUBLIC_URL", c.env)
		if got := NetworkLinks("3430", key); !slices.Equal(got, []string{c.want}) {
			t.Errorf("STARLING_PUBLIC_URL=%q: %v, want %s", c.env, got, c.want)
		}
	}
}
