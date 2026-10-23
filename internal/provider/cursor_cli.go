package provider

// Cursor's CLI, and where its sign-in is not read from: the built-in is
// retired (retiredBuiltins) and served by its plugin, so what is left here
// is the CLI's own path on this machine.

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/lulu58588/starling/internal/proc"
)

// CursorExecutable finds the cursor-agent CLI; a var so tests can fake it.
var CursorExecutable = func() string {
	for _, name := range []string{"cursor-agent", "agent"} {
		if p, err := proc.LookPath(name); err == nil && (name == "cursor-agent" || isCursorAgent(p)) {
			return p
		}
	}
	home, _ := os.UserHomeDir()
	for _, p := range append([]string{filepath.Join(home, ".local", "bin", "cursor-agent")}, proc.SystemDirs("/usr/local/bin/cursor-agent", "/opt/homebrew/bin/cursor-agent")...) {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return ""
}

// isCursorAgent tells Cursor's `agent` from any other program by that name.
func isCursorAgent(path string) bool {
	real, err := filepath.EvalSymlinks(path)
	return err == nil && strings.Contains(real, "cursor-agent")
}
