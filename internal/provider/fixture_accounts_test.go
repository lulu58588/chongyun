package provider

// The tests set the agents' own files up the way they always did: the
// account an agent is signed in to is in them. starling never reads them —
// it signs the agents in itself and keeps what it signed in to in its own
// logins.json, marking there the account each agent is on.
//
// This shim is what the tests use in place of starling reading them: what
// the agents' own files hold is what starling's logins.json is given, the
// account marked as the one the agent is signed in to. It reads the test's
// own files, which are the tests' fixtures, and nothing else.

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"
)

func init() { rememberAccounts = rememberFixtureAccounts }

// rememberLogins brings the accounts the agents are signed in to into
// logins.json. The tests call it where they always did, to have the
// accounts looked at again.
func rememberLogins(force bool) { rememberFixtureAccounts(force) }

// seenEvery is how often an account still signed in has the time it was
// last seen written, when nothing else of it changed.
const seenEvery = 10 * time.Minute

// loginsLookDue says whether the accounts are due to be looked at: not
// within the half minute of the last look.
func loginsLookDue() bool {
	at := loginsSeenAt.Load()
	return at == 0 || time.Since(time.Unix(0, at)) >= 30*time.Second
}

// onlySeenMoved says next is was with nothing changed but the times
// accounts were last seen, each of them less than seenEvery before now.
func onlySeenMoved(was, next []savedLogin, now time.Time) bool {
	if len(was) != len(next) {
		return false
	}
	norm := func(ls []savedLogin) []savedLogin {
		out := slices.Clone(ls)
		for i := range out {
			out[i].ID = loginID(out[i])
		}
		sortLogins(out)
		return out
	}
	a, b := norm(was), norm(next)
	for i := range b {
		if a[i].ID != b[i].ID {
			return false
		}
		if !b[i].Seen.Equal(a[i].Seen) {
			if a[i].Seen.IsZero() || now.Sub(a[i].Seen) >= seenEvery || b[i].Seen.Before(a[i].Seen) {
				return false
			}
			b[i].Seen = a[i].Seen
		}
	}
	ja, err1 := json.Marshal(a)
	jb, err2 := json.Marshal(b)
	return err1 == nil && err2 == nil && bytes.Equal(ja, jb)
}

// rememberFixtureAccounts brings the accounts the agents in the test's own
// home are signed in to into logins.json: the account the agent is on is
// the one starling serves it, marked as such in logins.json.
func rememberFixtureAccounts(force bool) {
	if !force && !loginsLookDue() {
		return
	}
	loginsSeenAt.Store(time.Now().UnixNano())
	// not waited for: whoever holds the lock is looking at the accounts
	// too, and waiting for it here would deadlock
	if !loginsMu.TryLock() {
		return
	}
	defer loginsMu.Unlock()
	was := readLogins()
	ls := slices.Clone(was)
	now := time.Now().UTC().Truncate(time.Second)
	changed := false
	for _, l := range fixtureAgentAccounts() {
		l.Seen = now
		// the name the account is kept under, which upsertLogin settles on too
		l.User = codexName(ls, l)
		ls = upsertLogin(ls, l)
		ls = setOwnActive(ls, l.Agent, l.User)
		changed = true
	}
	if changed && !onlySeenMoved(was, ls, now) {
		_ = writeLogins(ls)
	}
}

// fixtureAgentAccounts are the accounts the agents in the test's own home
// are signed in to.
func fixtureAgentAccounts() []savedLogin {
	var out []savedLogin
	if l, ok := fixtureCodexAccount(); ok {
		out = append(out, l)
	}
	if l, ok := fixtureClaudeAccount(); ok {
		out = append(out, l)
	}
	out = append(out, fixtureCopilotAccounts()...)
	return out
}

// fixtureCodexAccount is the ChatGPT account the Codex CLI in the test's
// own home is signed in to.
func fixtureCodexAccount() (savedLogin, bool) {
	b, err := os.ReadFile(filepath.Join(codexHome(), "auth.json"))
	if err != nil {
		return savedLogin{}, false
	}
	var a codexAuth
	if json.Unmarshal(b, &a) != nil || a.Tokens.AccessToken == "" || a.AuthMode == "apikey" {
		return savedLogin{}, false
	}
	id := jwtClaims(a.Tokens.IDToken)
	user := codexUser(id)
	if user == "" {
		user = a.Tokens.AccountID
	}
	if user == "" {
		return savedLogin{}, false
	}
	l := savedLogin{Agent: "codex", User: user,
		Plan: claimString(id, "https://api.openai.com/auth", "chatgpt_plan_type"),
		Auth: json.RawMessage(bytes.TrimSpace(b)), Active: true}
	l.User = codexLiveName(user, l.Auth)
	return l, true
}

// fixtureClaudeAccount is the Claude account the Claude Code in the test's
// own home is signed in to.
func fixtureClaudeAccount() (savedLogin, bool) {
	home, err := os.UserHomeDir()
	if err != nil {
		return savedLogin{}, false
	}
	cred, err := os.ReadFile(filepath.Join(home, ".claude", ".credentials.json"))
	if err != nil {
		return savedLogin{}, false
	}
	c, ok := parseClaudeCredentials(cred)
	if !ok {
		return savedLogin{}, false
	}
	b, err := c.marshal()
	if err != nil {
		return savedLogin{}, false
	}
	var acct map[string]any
	if p, err := os.ReadFile(filepath.Join(home, ".claude.json")); err == nil {
		var pr struct {
			OAuthAccount map[string]any `json:"oauthAccount"`
		}
		json.Unmarshal(p, &pr)
		acct = pr.OAuthAccount
	}
	l, err := claudeLogin(c, acct)
	if err != nil {
		return savedLogin{}, false
	}
	l.Auth = json.RawMessage(b)
	l.Active = true
	return l, true
}

// fixtureCopilotAccounts are the Copilot accounts the editors in the test's
// own home are signed in to.
func fixtureCopilotAccounts() []savedLogin {
	cfg := os.Getenv("XDG_CONFIG_HOME")
	if cfg == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil
		}
		cfg = filepath.Join(home, ".config")
	}
	var out []savedLogin
	for _, name := range []string{"apps.json", "hosts.json"} {
		var apps map[string]copilotApp
		if !readJSON(filepath.Join(cfg, "github-copilot", name), &apps) {
			continue
		}
		var keys []string
		for k := range apps {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			app := apps[k]
			if app.Token == "" || app.User == "" {
				continue
			}
			host := ""
			if h, _, _ := strings.Cut(k, ":"); h != "github.com" {
				h, err := CopilotHost(h)
				if err != nil || h == "" {
					continue
				}
				host = h
			}
			auth, _ := json.Marshal(map[string]any{"oauth_token": app.Token, "host": host})
			out = append(out, savedLogin{Agent: "copilot", User: CopilotAccountName(app.User, host),
				Auth: auth, Active: true})
		}
	}
	return out
}
