package agent

import (
	"strings"
	"testing"
)

// A distro with no agent chongyun sets up, only the Copilot CLI signed in
// (#723, jia2), is kept: the CLI's folder makes no agent row.
func TestWSLCopilotHome(t *testing.T) {
	syncHome(t)
	root := t.TempDir()
	if !strings.Contains(wslProbeScript, `[ -d "$HOME/.copilot" ] && echo dir:.copilot; `) {
		t.Fatalf("the probe doesn't look for the Copilot CLI's folder:\n%s", wslProbeScript)
	}
	t.Cleanup(FakeWSL(
		map[string]string{"Ubuntu": "home:/home/jia\ndir:.copilot\n", "Debian": "home:/home/me\n"},
		map[string]string{"Ubuntu": root, "Debian": t.TempDir()}))

	if len(wslAgents()) != 0 {
		t.Fatalf("the Copilot CLI's folder made an agent row: %v", wslAgents())
	}
}
