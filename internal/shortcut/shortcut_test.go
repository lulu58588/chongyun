package shortcut

import "testing"

func TestTarget(t *testing.T) {
	for in, want := range map[string]string{
		`C:\Tools\starling-windows-amd64.exe`:          `C:\Tools\starling-windows-amd64.exe`,
		`C:\Tools\starling.exe.old`:                    `C:\Tools\starling.exe`,
		`C:\Tools\starling.exe.old-3`:                  `C:\Tools\starling.exe`,
		`C:\Tools\starling.exe.new`:                    `C:\Tools\starling.exe`,
		`C:\Tools\STARLING.EXE.OLD`:                    `C:\Tools\STARLING.EXE`,
		`C:\old\starling.exe`:                          `C:\old\starling.exe`,
		`C:\Tools\starling.exe.older`:                  `C:\Tools\starling.exe.older`,
		`/usr/local/bin/starling`:                      `/usr/local/bin/starling`,
		`C:\Users\me\Downloads\starling (1).exe.old-2`: `C:\Users\me\Downloads\starling (1).exe`,
	} {
		if got := Target(in); got != want {
			t.Errorf("Target(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestStale(t *testing.T) {
	exe := "/tools/starling.exe"
	for _, c := range []struct {
		target, dir string
		stale       bool
	}{
		{"", "", true},                              // no shortcut yet
		{"/tools/starling.exe", "/tools", false},    // already this one
		{"/TOOLS/Starling.EXE", "/Tools/", false},   // the same, in another case
		{"/downloads/starling.exe", "/tools", true}, // the exe was moved
		{"/tools/starling.exe", "/downloads", true}, // starts in another folder
		{"/tools/starling.exe", "", true},           // in no folder
	} {
		if got := Stale(c.target, c.dir, exe); got != c.stale {
			t.Errorf("Stale(%q, %q) = %v, want %v", c.target, c.dir, got, c.stale)
		}
	}
}
