package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lulu58588/starling/internal/edit"
	"github.com/lulu58588/starling/internal/gateway"
	"github.com/lulu58588/starling/internal/settings"
)

// setPort is Settings' gateway port set to p, as the Settings page saves it.
func setPort(t *testing.T, p int) {
	t.Helper()
	s := settings.Load()
	s.Port = p
	if err := settings.Save(s); err != nil {
		t.Fatal(err)
	}
}

// filesWith are the files under home holding s.
func filesWith(home, s string) []string {
	var out []string
	filepath.WalkDir(home, func(p string, e os.DirEntry, err error) error {
		if err != nil || e.IsDir() {
			return nil
		}
		if b, _ := os.ReadFile(p); strings.Contains(string(b), s) {
			out = append(out, p)
		}
		return nil
	})
	return out
}

// The gateway's port set in Settings (Magic_zero 报告) moves every
// agent starling connected: each one's config is written with the gateway's
// URL on the new port, none keeps the old one, and each is still connected
// with nothing off.
func TestPortMovesEveryAgent(t *testing.T) {
	t.Setenv("STARLING_ADDR", "")
	home, _ := codexHome(t, "", "")
	managed := claudeManaged
	claudeManaged = func() string { return filepath.Join(home, "managed-settings.json") }
	t.Cleanup(func() { claudeManaged = managed })
	almaApp := startAlma(t)
	os.MkdirAll(filepath.Join(home, ".hanako", "agents", "hana"), 0o755)
	os.WriteFile(filepath.Join(home, ".hanako", "agents", "hana", "config.yaml"), []byte("agent:\n  name: Hana\n"), 0o644)
	// AstrBot is connected once it has written its settings
	os.MkdirAll(filepath.Join(home, ".astrbot", "data"), 0o755)
	os.WriteFile(filepath.Join(home, ".astrbot", "data", "cmd_config.json"), []byte(astrbotConfig), 0o644)
	setPort(t, 3591)
	if gateway.URL() != "http://127.0.0.1:3591" {
		t.Fatalf("Settings' port isn't the gateway's: %s", gateway.URL())
	}
	var wired []*Agent
	for _, a := range All() {
		// as TestDriftUnwiredEveryAgent: agy's gateway is in the command
		// that starts it, and a WSL agent's files are beyond the sandbox
		if a.Check == nil || a.Launch != nil || a.WSL != "" {
			continue
		}
		if a.Native != nil {
			if err := a.Connect(); err != nil {
				t.Fatal(err)
			}
			wired = append(wired, a)
			continue
		}
		var key, want string
		for _, g := range a.Fields {
			for _, o := range g.Options(a.Values()) {
				if want == "" && (o.Ref == "fake/m1" || o.Value == starlingID) {
					key, want = g.Key, o.Value
				}
			}
		}
		if want == "" {
			t.Fatalf("%s: no field takes starling's models", a.ID)
		}
		if err := a.Apply(key, want); err != nil {
			t.Fatalf("%s: %v", a.ID, err)
		}
		wired = append(wired, a)
	}
	if len(filesWith(home, "127.0.0.1:3591")) == 0 {
		t.Fatal("no agent's config has the gateway's URL")
	}
	on := OnGateway()
	for _, a := range wired {
		if !containsAgent(on, a.ID) {
			t.Errorf("%s is connected but not on the gateway: %+v", a.ID, a.Drift())
		}
	}

	setPort(t, 3592)
	moved, err := Rewire(on)
	if err != nil {
		t.Fatal(err)
	}
	// OpenChamber reads OpenCode's config, which moves with OpenCode
	if len(moved) < len(wired)-1 {
		t.Errorf("moved %d of %d: %v", len(moved), len(wired), moved)
	}
	if left := filesWith(home, "127.0.0.1:3591"); len(left) > 0 {
		t.Errorf("still on the old port: %v", left)
	}
	if almaApp.repoint("127.0.0.1:3591", "127.0.0.1:3591") > 0 {
		t.Error("Alma still has the old port")
	}
	for _, a := range wired {
		if !a.Wired() {
			t.Errorf("%s isn't connected after the move", a.ID)
		}
		if d := a.Drift(); d != nil {
			t.Errorf("%s is off after the move: %+v", a.ID, d)
		}
	}
}

func containsAgent(as []*Agent, id string) bool {
	for _, a := range as {
		if a.ID == id {
			return true
		}
	}
	return false
}

// 星鸣不读别家的登录态，此用例随之移除。

// An agent taken off starling before the port moved is left as it is: it is
// the user's to set again, not the move's.
func TestPortLeavesAnUnwiredAgent(t *testing.T) {
	t.Setenv("STARLING_ADDR", "")
	home, read := codexHome(t, "", "")
	setPort(t, 3591)
	cx := codex(home)
	if err := cx.Apply("model", "fake/m1"); err != nil {
		t.Fatal(err)
	}
	cfg := filepath.Join(home, ".codex", "config.toml")
	os.WriteFile(cfg, []byte(strings.ReplaceAll(read(), "127.0.0.1:3591", "127.0.0.1:9")), 0o644)
	on := OnGateway()
	setPort(t, 3592)
	if _, err := Rewire(on); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(read(), "127.0.0.1:3592") {
		t.Fatalf("an agent taken off starling was moved:\n%s", read())
	}
}

// A starling STARLING_ADDR puts beside the installed one tells the installed
// one's wiring by the port the installed one's Settings have, not 3425.
func TestDriftInstalledStarlingOnItsPort(t *testing.T) {
	home, _ := codexHome(t, "", "")
	setPort(t, 3591)
	t.Setenv("STARLING_ADDR", "127.0.0.1:3426")
	cx := codex(home)
	if err := cx.Apply("model", "fake/m1"); err != nil {
		t.Fatal(err)
	}
	cfg := filepath.Join(home, ".codex", "config.toml")
	os.WriteFile(cfg, []byte("openai_base_url = \"http://127.0.0.1:3591"+gateway.CodexPath+"\"\nmodel = \"fake/m1\"\n"), 0o644)
	if d := cx.Drift(); d != nil {
		t.Fatalf("the installed starling's wiring taken for drift: %+v", d)
	}
}

// Moved, an agent still puts back what it had before starling when switched
// off: its own endpoint and token, not starling's at the old port, which
// setting it again there would otherwise have noted as its own.
func TestPortKeepsWhatAnAgentHadBefore(t *testing.T) {
	t.Setenv("STARLING_ADDR", "")
	home, _ := codexHome(t, "", "")
	managed := claudeManaged
	claudeManaged = func() string { return filepath.Join(home, "managed-settings.json") }
	t.Cleanup(func() { claudeManaged = managed })
	path := filepath.Join(home, ".claude", "settings.json")
	os.MkdirAll(filepath.Dir(path), 0o755)
	os.WriteFile(path, []byte(`{"env":{"ANTHROPIC_BASE_URL":"https://relay.example","ANTHROPIC_AUTH_TOKEN":"own-token"}}`), 0o644)
	setPort(t, 3591)
	cc := claude(home)
	var want string
	for _, o := range cc.Field("model").Options(cc.Values()) {
		if o.Ref == "fake/m1" {
			want = o.Value
		}
	}
	if err := cc.Apply("model", want); err != nil {
		t.Fatal(err)
	}
	on := OnGateway()
	setPort(t, 3592)
	if _, err := Rewire(on); err != nil {
		t.Fatal(err)
	}
	if u, _ := edit.GetJSON(path, "env.ANTHROPIC_BASE_URL"); u != "http://127.0.0.1:3592" {
		t.Fatalf("Claude Code not moved: %q", u)
	}
	if err := cc.Disconnect(); err != nil {
		t.Fatal(err)
	}
	u, _ := edit.GetJSON(path, "env.ANTHROPIC_BASE_URL")
	tok, _ := edit.GetJSON(path, "env.ANTHROPIC_AUTH_TOKEN")
	if u != "https://relay.example" || tok != "own-token" {
		t.Fatalf("switched off, Claude Code got %q %q back, not its own", u, tok)
	}
}
