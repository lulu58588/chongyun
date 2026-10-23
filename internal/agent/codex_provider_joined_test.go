package agent

import (
	"strings"
	"testing"
)

// isymbo (#2267): Codex joined beside its ChatGPT sign-in, a starling model
// picked, then starling API chosen. The config had starling as Codex's
// provider, but the joined mark stayed, so the Agents row still said
// Codex's own models stay in /model beside starling's: the ChatGPT state
// the user had just left, which is the one the Codex app blocks sending
// in while the account is out. Joined is false while starling is Codex's
// provider, and true again once it is back beside the sign-in.
func TestCodexProviderIsNotJoined(t *testing.T) {
	home, read := codexHome(t, `{"tokens":{"access_token":"x","id_token":"x.e30.x"}}`, `model = "gpt-5.5"`+"\n")
	// starling never reads the agent's own sign-in: a Codex signed in to
	// ChatGPT is one starling wired beside that sign-in.
	codexBeside(t, home, "beside")
	cx := codex(home)
	if ok, err := cx.Join(); err != nil || !ok || !cx.Joined() {
		t.Fatalf("join: %v %v, joined %v", ok, err, cx.Joined())
	}
	if err := cx.Fields[0].Set("fake/m1"); err != nil {
		t.Fatal(err)
	}
	if !cx.Joined() || !cx.Beside() {
		t.Fatalf("a starling model beside the sign-in: joined %v, beside %v", cx.Joined(), cx.Beside())
	}
	if err := cx.Field("login").Set("api"); err != nil {
		t.Fatal(err)
	}
	if cfg := read(); !strings.Contains(cfg, `model_provider = "starling"`) {
		t.Fatalf("api:\n%s", cfg)
	}
	if cx.Joined() || cx.Beside() {
		t.Fatalf("starling API: joined %v, beside %v", cx.Joined(), cx.Beside())
	}
	if !cx.Wired() {
		t.Fatal("starling API: not wired")
	}

	// the ChatGPT account held by the Codex app: starling its provider too
	if err := cx.Field("login").Set(""); err != nil {
		t.Fatal(err)
	}
	if !cx.Joined() {
		t.Fatal("back beside the sign-in: not joined")
	}
	codexUsedUp = func() bool { return true }
	t.Cleanup(func() { codexUsedUp = func() bool { return false } })
	if err := cx.Sync(); err != nil {
		t.Fatal(err)
	}
	if cfg := read(); !strings.Contains(cfg, `model_provider = "starling"`) || cx.Joined() {
		t.Fatalf("held: joined %v\n%s", cx.Joined(), cfg)
	}
	codexUsedUp = func() bool { return false }
	if err := cx.Sync(); err != nil {
		t.Fatal(err)
	}
	if cfg := read(); strings.Contains(cfg, "model_provider =") || !cx.Joined() {
		t.Fatalf("room again: joined %v\n%s", cx.Joined(), cfg)
	}
}
