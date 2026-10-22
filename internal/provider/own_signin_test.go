package provider

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// The sign-ins the tests make are chongyun's own: they go into logins.json,
// where chongyun keeps what it signed in to. No other agent's credential
// file is read any more, so a test that wants an account signed in puts it
// where chongyun puts it.

// ownLogin keeps one account in logins.json, the way chongyun's own sign-in
// does, beside the ones already there: it is the account the agent is signed
// in to now, so the ones before it are signed out of it.
//
// Whether the account is also in use beside it (on) is the user's to say:
// the accounts chongyun finds the agents signed in to are not in use beside
// the one in use, and the ones the user turns on are.
func ownLogin(t *testing.T, l savedLogin) {
	t.Helper()
	if l.Seen.IsZero() {
		l.Seen = time.Now().UTC()
	}
	loginsMu.Lock()
	defer loginsMu.Unlock()
	ls := readLogins()
	// the name the account is kept under, which upsertLogin settles on too
	l.User = codexName(ls, l)
	ls = setOwnActive(upsertLogin(ls, l), l.Agent, l.User)
	if err := writeLogins(ls); err != nil {
		t.Fatal(err)
	}
}

// ownFirstLogin keeps one account in logins.json as the one served first
// for its agent, the way chongyun's own sign-in leaves it when it is the
// account the agent is on.
func ownFirstLogin(t *testing.T, l savedLogin) {
	t.Helper()
	l.On = true
	l.First = true
	if l.Seen.IsZero() {
		l.Seen = time.Now().UTC()
	}
	loginsMu.Lock()
	defer loginsMu.Unlock()
	ls := readLogins()
	l.User = codexName(ls, l)
	ls = setOwnActive(upsertLogin(ls, l), l.Agent, l.User)
	if err := writeLogins(ls); err != nil {
		t.Fatal(err)
	}
}

// codexAuthBlob is a ChatGPT sign-in as auth.json keeps it.
func codexAuthBlob(email, plan, accountID string, expiry time.Time) map[string]any {
	return map[string]any{
		"auth_mode": "chatgpt",
		"tokens": map[string]any{
			"id_token":      fakeJWT(map[string]any{"email": email, "https://api.openai.com/auth": map[string]any{"chatgpt_plan_type": plan, "chatgpt_account_id": accountID}}),
			"access_token":  fakeJWT(map[string]any{"exp": float64(expiry.Unix())}),
			"refresh_token": "r", "account_id": accountID,
		},
	}
}

// codexOwnLogin keeps a ChatGPT account in logins.json, as chongyun's own
// sign-in does: the account signed in to is the one in use, the one before
// it is out of it.
func codexOwnLogin(t *testing.T, user, plan string, auth map[string]any) {
	t.Helper()
	b, _ := json.Marshal(auth)
	ownLogin(t, savedLogin{Agent: "codex", User: user, Plan: plan, Auth: b, Seen: time.Now().UTC()})
}

// copilotOwnLogin keeps a Copilot account in logins.json, as chongyun's own
// sign-in does.
func copilotOwnLogin(t *testing.T, user, plan, token string) {
	t.Helper()
	ownLogin(t, savedLogin{Agent: "copilot", User: user, Plan: plan,
		Auth: []byte(`{"oauth_token":"` + token + `"}`)})
}

// claudeTestSignIn is a Claude sign-in a test makes.
type claudeTestSignIn struct {
	email, plan, access, refresh string
	expiry                       time.Time
	org                          string
}

// claudeOwnLogin keeps a Claude account in logins.json, as chongyun's own
// sign-in does: its sign-in, and who it is.
func claudeOwnLogin(t *testing.T, in claudeTestSignIn) savedLogin {
	t.Helper()
	blob, _ := json.Marshal(map[string]any{"claudeAiOauth": map[string]any{
		"accessToken": in.access, "refreshToken": in.refresh,
		"expiresAt": in.expiry.UnixMilli(), "subscriptionType": in.plan,
		"scopes": []string{"user:inference", "user:profile"},
	}})
	c, ok := parseClaudeCredentials(blob)
	if !ok {
		t.Fatal("the Claude sign-in made for the test can't be read")
	}
	acct := map[string]any{"emailAddress": in.email}
	if in.org != "" {
		acct["organizationName"] = in.org
	}
	l, err := claudeLogin(c, acct)
	if err != nil {
		t.Fatal(err)
	}
	ownLogin(t, l)
	return l
}

// claudeAccountSignIn keeps a Claude account in logins.json, as chongyun's
// own sign-in does.
func claudeAccountSignIn(t *testing.T, email, plan string, expiry time.Time) savedLogin {
	t.Helper()
	return claudeOwnLogin(t, claudeTestSignIn{email: email, plan: plan,
		access: "sk-ant-oat01-old", refresh: "sk-ant-ort01-old", expiry: expiry})
}

// claudeRemember keeps the account Claude Code is signed in to in
// logins.json, the way chongyun kept it when it read Claude Code's own
// sign-in. The tests that want an account signed in set Claude Code's own
// files up and call this: chongyun's own accounts are in logins.json.
func claudeRemember(t *testing.T, home string) savedLogin {
	t.Helper()
	cred, err := os.ReadFile(filepath.Join(home, ".claude", ".credentials.json"))
	if err != nil {
		t.Fatal(err)
	}
	c, ok := parseClaudeCredentials(cred)
	if !ok {
		t.Fatal("the Claude sign-in Claude Code keeps can't be read")
	}
	b, err := os.ReadFile(filepath.Join(home, ".claude.json"))
	if err != nil {
		t.Fatal(err)
	}
	var p struct {
		OAuthAccount map[string]any `json:"oauthAccount"`
	}
	json.Unmarshal(b, &p)
	l, err := claudeLogin(c, p.OAuthAccount)
	if err != nil {
		t.Fatal(err)
	}
	ownLogin(t, l)
	return l
}

// claudeCodeSignIn writes the sign-in Claude Code keeps in a temp home.
// chongyun doesn't read it any more; the tests that need Claude Code to
// have one write it.
func claudeCodeSignIn(t *testing.T, home string, expiry time.Time) string {
	t.Helper()
	path := filepath.Join(home, ".claude", ".credentials.json")
	writeFile(t, path, map[string]any{
		"claudeAiOauth": map[string]any{
			"accessToken": "sk-ant-oat01-old", "refreshToken": "sk-ant-ort01-old",
			"expiresAt": expiry.UnixMilli(), "subscriptionType": "max",
			"scopes": []string{"user:inference", "user:profile"},
		},
		// what magpie must not clobber when it writes a refreshed token back
		"mcpOAuth": map[string]any{"plugin:slack:slack|x": map[string]any{"serverName": "Slack"}},
	})
	return path
}
