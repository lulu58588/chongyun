package provider

import (
	"context"
	"os"
	"testing"
	"time"
)

// Claude Code wired to magpie, its accounts served by magpie: the user logs
// out of Claude Code (it says to, over magpie's ANTHROPIC_AUTH_TOKEN), and
// the accounts saved in magpie are still served — the one it was signed in
// to first, in a config directory of its own, the others on behind it — and
// none of them is dropped from the list (StringKe 报告: all six went).
func TestClaudeAccountsOutliveLogout(t *testing.T) {
	home := claudeHome(t)
	t.Setenv("CLAUDE_CONFIG_DIR", "")
	noAnthropic(t)
	cred := claudeCodeSignIn(t, home, time.Now().Add(time.Hour))
	far := time.Now().Add(24 * time.Hour).UnixMilli()
	oauth := func(tok string) map[string]any {
		return map[string]any{"claudeAiOauth": map[string]any{"accessToken": tok, "refreshToken": "r-" + tok, "expiresAt": far, "subscriptionType": "max"}}
	}
	seen := time.Now().Add(-time.Hour).UTC().Truncate(time.Second)
	writeFile(t, loginsPath(), []map[string]any{
		{"agent": "claude", "user": "on@example.com", "plan": "max", "on": true, "seen": seen, "auth": oauth("tok-on")},
		{"agent": "claude", "user": "off@example.com", "plan": "pro", "seen": seen, "auth": oauth("tok-off")},
		{"agent": "claude", "user": "two@example.com", "plan": "max", "on": true, "seen": seen.Add(-time.Hour), "auth": oauth("tok-two")},
	})
	loginsMu.Lock()
	loginsSeenAt.Store(0)
	loginsMu.Unlock()
	if p, ok := find(All(), "claude"); !ok || p.Account.User != "on@example.com" {
		t.Fatalf("signed in: %v %+v", ok, p.Account)
	}

	// /logout: Claude Code's sign-in is gone. chongyun doesn't read it, so
	// none of its accounts changes: they are sign-ins of its own.
	os.Remove(cred)
	loginsMu.Lock()
	loginsSeenAt.Store(0)
	loginsMu.Unlock()

	p, ok := find(All(), "claude")
	if !ok || p.Account == nil || p.Account.User != "on@example.com" {
		t.Fatalf("logged out: %v %+v", ok, p.Account)
	}
	dir, own, err := p.Account.Token(context.Background())
	if err != nil || !own || dir != claudeAccountDir("on@example.com") {
		t.Fatalf("dir %q %v %v", dir, own, err)
	}
	if c, ok := readClaudeDir(dir); !ok || c.OAuth.RefreshToken != "r-tok-on" {
		t.Fatalf("dir credentials %+v", c.OAuth)
	}
	var also []string
	for _, q := range p.AlsoOn() {
		also = append(also, q.Account.User)
	}
	if len(also) != 1 || also[0] != "two@example.com" {
		t.Fatalf("also on: %v", also)
	}
	ls := Logins("claude")
	if len(ls) != 3 {
		t.Fatalf("logins: %+v", ls)
	}
	for _, l := range ls {
		if l.Active != (l.User == "on@example.com") ||
			l.On != (l.User == "on@example.com" || l.User == "two@example.com") || l.Lapsed != "" {
			t.Fatalf("login %+v", l)
		}
	}
	if u := InUseLogin("claude"); u != "on@example.com" {
		t.Fatalf("in use %q", u)
	}

	// the account in use can be paused behind another on
	if err := SetLoginOn("claude", "on@example.com", false); err != nil {
		t.Fatal(err)
	}
	if p, _ = find(All(), "claude"); !p.OwnPaused() {
		t.Fatal("not paused")
	}
	if u := InUseLogin("claude"); u != "two@example.com" {
		t.Fatalf("in use %q", u)
	}
}

// A switch in magpie takes Claude Code off one account and puts it on
// another, saving the first's sign-in as it was: nothing revoked it, and it
// isn't taken for logged out.
func TestClaudeSwitchIsNoLogout(t *testing.T) {
	claudeHome(t)
	t.Setenv("CLAUDE_CONFIG_DIR", "")
	noAnthropic(t)
	far := time.Now().Add(24 * time.Hour).UnixMilli()
	writeFile(t, loginsPath(), []map[string]any{
		{"agent": "claude", "user": "first@example.com", "plan": "max", "on": true, "seen": time.Now().UTC(),
			"profile": map[string]any{"emailAddress": "first@example.com"},
			"auth":    map[string]any{"claudeAiOauth": map[string]any{"accessToken": "tok-first", "refreshToken": "r-first", "expiresAt": far}}},
		{"agent": "claude", "user": "other@example.com", "plan": "max", "seen": time.Now().Add(-time.Hour).UTC(),
			"profile": map[string]any{"emailAddress": "other@example.com"},
			"auth":    map[string]any{"claudeAiOauth": map[string]any{"accessToken": "tok-other", "refreshToken": "r-other", "expiresAt": far}}},
	})
	ForgetAccounts()
	if u := InUseLogin("claude"); u != "first@example.com" {
		t.Fatalf("in use %q", u)
	}
	if err := SwitchLogin("claude", "other@example.com"); err != nil {
		t.Fatal(err)
	}
	ForgetAccounts()
	for _, l := range Logins("claude") {
		if l.Lapsed != "" || l.Active != (l.User == "other@example.com") {
			t.Fatalf("login %+v", l)
		}
	}
}

// 崇云不读别家的登录态，此用例随之移除。
//
// The case asserted that an account signed in to in chongyun while Claude
// Code is logged out stays chongyun's: Claude Code isn't signed in to it
// again, which would only bring back its warning over chongyun's token.
// What told it that Claude Code is logged out was Claude Code's own
// credentials, which chongyun no longer reads; and the account chongyun
// signs in to is the one the agent is on (addLogin is always using), so
// Claude Code is signed in to it as well.

// Every saved Claude account's sign-in gone while Claude Code is signed out
// (LeeHom 报告: two Claude Pro accounts vanished from Providers, and
// were taken for a ban): the one Claude Code held went with its sign-in,
// and the other had gone the same way before. Claude Code stays on the
// Providers list, signed out, each account on it saying to sign in again,
// and its token says so too; none is served as if it could be used.
func TestClaudeAllSignedOutStaysListed(t *testing.T) {
	claudeHome(t)
	t.Setenv("CLAUDE_CONFIG_DIR", "")
	noAnthropic(t)
	far := time.Now().Add(24 * time.Hour).UnixMilli()
	seen := time.Now().Add(-time.Hour).UTC().Truncate(time.Second)
	writeFile(t, loginsPath(), []map[string]any{
		{"agent": "claude", "user": "one@example.com", "plan": "max", "on": true, "seen": seen, "lapsed": claudeLogoutLapse,
			"auth": map[string]any{"claudeAiOauth": map[string]any{"accessToken": "tok-one", "refreshToken": "r-one", "expiresAt": far, "subscriptionType": "max"}}},
		{"agent": "claude", "user": "two@example.com", "plan": "pro", "on": true, "seen": seen.Add(-time.Hour), "lapsed": claudeLogoutLapse,
			"auth": map[string]any{"claudeAiOauth": map[string]any{"accessToken": "tok-two", "refreshToken": "r-two", "expiresAt": far, "subscriptionType": "pro"}}},
	})
	loginsMu.Lock()
	loginsSeenAt.Store(0)
	loginsMu.Unlock()
	p, ok := find(All(), "claude")
	if !ok || p.Account == nil {
		t.Fatalf("Claude left the Providers list with every account signed out: %v", ok)
	}
	if p.Account.User != "one@example.com" {
		t.Fatalf("listed as %q; want the one seen last", p.Account.User)
	}
	if _, _, err := p.Account.Token(context.Background()); err == nil || err.Error() != claudeLogoutLapse {
		t.Fatalf("token of a signed-out account: %v", err)
	}
	ls := Logins("claude")
	if len(ls) != 2 {
		t.Fatalf("logins: %+v", ls)
	}
	listedOn := false
	for _, l := range ls {
		if l.Lapsed != claudeLogoutLapse {
			t.Fatalf("not signed out: %+v", l)
		}
		listedOn = listedOn || l.On && l.User == "one@example.com"
	}
	// the row reads as signed out: the account it names is the one on
	if !listedOn {
		t.Fatalf("the listed account isn't on: %+v", ls)
	}
	if u := InUseLogin("claude"); u != "one@example.com" && u != "" {
		t.Fatalf("in use %q", u)
	}
	// what magpie providers, the TUI and the app's row say of it
	if !p.SignedOut() {
		t.Fatal("listed on a signed-out account, the provider doesn't say it is signed out")
	}
}
