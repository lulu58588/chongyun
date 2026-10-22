package gui

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lulu58588/chongyun/internal/provider"
	"github.com/lulu58588/chongyun/internal/settings"
)

// chongyunLogins writes chongyun's own logins.json: the accounts it holds,
// each marked as the one its agent is on. chongyun signs the agents in
// itself and keeps what it signed in to there; it never reads another tool's
// own file to find out which account that is, so a test that means "the
// agent is signed in to this account" writes it here.
func chongyunLogins(t *testing.T, entries ...map[string]any) {
	t.Helper()
	path := filepath.Join(filepath.Dir(provider.Path()), "logins.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(entries)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatal(err)
	}
	provider.ForgetAccounts()
	t.Cleanup(provider.ForgetAccounts)
}

// #1865: Settings › Images lists a ChatGPT account's one GPT Image, marked
// as the vendor's pick, and a gpt-image-2.5 saved when it was listed shows
// as that entry while the saved value is kept as it is.
func TestImageGenCodexOneEntry(t *testing.T) {
	h := t.TempDir()
	t.Setenv("HOME", h)
	t.Setenv("USERPROFILE", h)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(h, ".config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(h, ".cache"))
	t.Setenv("CODEX_HOME", "")
	claims := func(m map[string]any) string {
		b, _ := json.Marshal(m)
		return "h." + base64.RawURLEncoding.EncodeToString(b) + ".s"
	}
	auth, _ := json.Marshal(map[string]any{"auth_mode": "chatgpt", "tokens": map[string]any{
		"id_token":      claims(map[string]any{"email": "me@example.com"}),
		"access_token":  claims(map[string]any{"exp": time.Now().Add(time.Hour).Unix()}),
		"refresh_token": "r", "account_id": "acct-1"}})
	os.MkdirAll(filepath.Join(h, ".codex"), 0o755)
	if err := os.WriteFile(filepath.Join(h, ".codex", "auth.json"), auth, 0o600); err != nil {
		t.Fatal(err)
	}
	// chongyun keeps the account it signed the agent in to in its own
	// logins.json, marked as the one the agent is on: it never reads the
	// agent's own file to find out which account that is.
	chongyunLogins(t, map[string]any{"agent": "codex", "user": "me@example.com", "active": true,
		"seen": time.Now(), "auth": json.RawMessage(auth)})
	provider.ForgetAccounts()
	t.Cleanup(provider.ForgetAccounts)
	st := settings.Load()
	st.ImageGen = "codex/gpt-image-2.5"
	if err := settings.Save(st); err != nil {
		t.Fatal(err)
	}
	s := settingsState()
	var codex []modelRef
	for _, m := range s.ImageGenModels {
		if m.Provider == "codex" {
			codex = append(codex, m)
		}
	}
	if len(codex) != 1 || codex[0].ID != "codex/gpt-image-2" || codex[0].Name != "GPT Image" || !codex[0].VendorPicks {
		t.Fatalf("codex image models %+v", codex)
	}
	if s.ImageGen != "codex/gpt-image-2.5" || s.ImageGenShown != "codex/gpt-image-2" || s.ImageGenMissing != "" {
		t.Fatalf("imageGen %q shown %q missing %q", s.ImageGen, s.ImageGenShown, s.ImageGenMissing)
	}
	st.ImageGen = "codex/gpt-image-2"
	if err := settings.Save(st); err != nil {
		t.Fatal(err)
	}
	if s := settingsState(); s.ImageGenShown != "" {
		t.Fatalf("a listed pick is shown as %q", s.ImageGenShown)
	}
}
