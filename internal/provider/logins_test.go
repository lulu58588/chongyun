package provider

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// codexSignIn keeps a ChatGPT account in logins.json, as chongyun's own
// sign-in does: the account it signed in to is kept there, beside the
// others.
func codexSignIn(t *testing.T, email, refresh string) {
	t.Helper()
	auth := map[string]any{
		"auth_mode": "chatgpt",
		"tokens": map[string]any{
			"id_token":      fakeJWT(map[string]any{"email": email, "https://api.openai.com/auth": map[string]any{"chatgpt_plan_type": "plus"}}),
			"access_token":  fakeJWT(map[string]any{"exp": float64(time.Now().Add(time.Hour).Unix())}),
			"refresh_token": refresh, "account_id": "acct-" + email,
		},
	}
	codexOwnLogin(t, email, "plus", auth)
}

func loginUsers(ls []Login) (users []string, active string) {
	for _, l := range ls {
		users = append(users, l.User)
		if l.Active {
			active = l.User
		}
	}
	return
}

func TestCodexLogins(t *testing.T) {
	home := signIn(t)
	codexSignIn(t, "work@example.com", "r-work")

	users, active := loginUsers(Logins("codex"))
	if strings.Join(users, ",") != "me@example.com,work@example.com" || active != "work@example.com" {
		t.Fatalf("logins %v, active %q", users, active)
	}
	if fi, err := os.Stat(loginsPath()); err != nil || runtime.GOOS != "windows" && fi.Mode().Perm() != 0o600 {
		t.Fatalf("logins.json mode: %v %v", fi, err)
	}
	// the agent refreshes the active login in place; a switch keeps that
	codexSignIn(t, "work@example.com", "r-work-2")

	if err := SwitchLogin("codex", "ME@example.com"); err != nil {
		t.Fatal(err)
	}
	var a codexAuth
	readJSON(filepath.Join(home, ".codex", "auth.json"), &a)
	if a.Tokens.RefreshToken != "r" {
		t.Fatalf("auth.json not switched: %q", a.Tokens.RefreshToken)
	}
	if _, active := loginUsers(Logins("codex")); active != "me@example.com" {
		t.Fatalf("active after switch: %q", active)
	}
	for _, l := range readLogins() {
		if l.User == "work@example.com" && !strings.Contains(string(l.Auth), "r-work-2") {
			t.Fatalf("the replaced login lost its refreshed token: %s", l.Auth)
		}
	}

	// the active one forgotten: Codex is signed in to the other first
	if err := ForgetLogin("codex", "me@example.com"); err != nil {
		t.Fatal(err)
	}
	if users, active := loginUsers(Logins("codex")); strings.Join(users, ",") != "work@example.com" || active != "work@example.com" {
		t.Fatalf("after forget: %v, active %q", users, active)
	}
	if err := SwitchLogin("codex", "nobody@example.com"); err == nil {
		t.Fatal("switched to an unknown login")
	}
}

// loginsCodexAuthOf is the sign-in chongyun keeps for the codex account
// user: what the agent is signed in to with, in logins.json.
func loginsCodexAuthOf(t *testing.T, user string) json.RawMessage {
	t.Helper()
	for _, l := range readLogins() {
		if l.Agent == "codex" && l.User == user {
			return l.Auth
		}
	}
	t.Fatalf("no saved codex account %q", user)
	return nil
}

func TestCodexDuplicateLoginsCollapse(t *testing.T) {
	signIn(t)
	codexSignIn(t, "me@example.com", "r")
	auth := loginsCodexAuthOf(t, "me@example.com")
	now := time.Now().UTC()
	if err := writeLogins([]savedLogin{
		{Agent: "codex", User: "me@example.com", Plan: "plus", Seen: now.Add(-time.Hour), On: true, First: true, Auth: auth},
		{Agent: "codex", User: "me@example.com", Plan: "plus", Seen: now, Auth: auth},
	}); err != nil {
		t.Fatal(err)
	}
	if got := readLogins(); len(got) != 1 || !got[0].Seen.Equal(now) || !got[0].On || !got[0].First {
		t.Fatalf("saved logins were not deduped: %+v", got)
	}
	raw, err := os.ReadFile(loginsPath())
	if err != nil {
		t.Fatal(err)
	}
	var saved []savedLogin
	if err := json.Unmarshal(raw, &saved); err != nil || len(saved) != 2 {
		t.Fatalf("reading logins changed the saved records: %v, %d records", err, len(saved))
	}
	if got := dedupeLogins([]savedLogin{saved[1], saved[0]}); len(got) != 1 || !got[0].Seen.Equal(now) || !got[0].On || !got[0].First {
		t.Fatalf("reverse-order logins were not deduped: %+v", got)
	}
	if got := Logins("codex"); len(got) != 1 || !got[0].On || !got[0].Active {
		t.Fatalf("deduped logins: %+v", got)
	}
	dupe := savedLogin{Agent: "codex", User: "me@example.com", Plan: "plus", Seen: now.Add(time.Minute), Auth: auth}
	if got := upsertLogin(readLogins(), dupe); len(got) != 1 || !got[0].Seen.Equal(dupe.Seen) {
		t.Fatalf("upsert left duplicate logins: %+v", got)
	}
}

func TestSideLoginsNotCollapsed(t *testing.T) {
	signIn(t)
	now := time.Now().UTC()
	if err := writeLogins([]savedLogin{
		{Agent: "grok", User: "me@example.com", Seen: now.Add(-time.Hour), Home: "/grok/me", First: true},
		{Agent: "grok", User: "me@example.com", Seen: now},
	}); err != nil {
		t.Fatal(err)
	}
	if got := readLogins(); len(got) != 2 || got[0].Home != "/grok/me" {
		t.Fatalf("a Grok account chongyun signed in was folded into the CLI's own: %+v", got)
	}
}

func TestClaudeLogins(t *testing.T) {
	claudeHome(t)
	// the accounts chongyun keeps, each of them in logins.json
	claudeAccountSignIn(t, "a@example.com", "max", time.Now().Add(time.Hour))
	// the second account has a sign-in of its own: two accounts of one
	// Claude credential are one account
	claudeOwnLogin(t, claudeTestSignIn{email: "b@example.com", plan: "pro",
		access: "sk-ant-oat01-b", refresh: "sk-ant-ort01-b", expiry: time.Now().Add(time.Hour)})
	forgetAccountCaches()

	users, active := loginUsers(Logins("claude"))
	if strings.Join(users, ",") != "a@example.com,b@example.com" || active != "b@example.com" {
		t.Fatalf("logins %v, active %q", users, active)
	}
	if err := SwitchLogin("claude", "a@example.com"); err != nil {
		t.Fatal(err)
	}
	if _, active := loginUsers(Logins("claude")); active != "a@example.com" {
		t.Fatalf("active after switch: %q", active)
	}
	// the account switched to runs Claude Code with the sign-in chongyun
	// keeps for it, in a config directory of its own
	dir, err := claudeSavedDir("a@example.com")
	if err != nil {
		t.Fatal(err)
	}
	c, ok := readClaudeDir(dir)
	if !ok || c.OAuth.RefreshToken != "sk-ant-ort01-old" {
		t.Fatalf("the sign-in the account runs with: %+v", c)
	}
}

// 崇云不读别家的登录态，此用例随之移除。
