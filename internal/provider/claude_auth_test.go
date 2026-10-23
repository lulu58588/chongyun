package provider

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/lulu58588/starling/internal/testenv"
)

func claudeAuthFixture(t *testing.T, lapse string) (Provider, savedLogin) {
	t.Helper()
	claudeHome(t)
	noAnthropic(t)
	t.Setenv("CLAUDE_CONFIG_DIR", "")
	// Claude Code's own sign-in is not read any more: the accounts are the
	// ones starling keeps, each of them in logins.json.
	main := savedLogin{Agent: "claude", User: "main@example.com", On: true, Plan: "max", Seen: time.Now(),
		Auth: mustJSONRaw(t, map[string]any{"claudeAiOauth": map[string]any{
			"accessToken": "own-access", "refreshToken": "own-refresh",
			"expiresAt": time.Now().Add(time.Hour).UnixMilli(), "subscriptionType": "max",
		}})}
	side := savedLogin{Agent: "claude", User: "side@example.com", On: true, Plan: "max", Seen: time.Now().Add(-time.Hour), Lapsed: lapse,
		Auth: mustJSONRaw(t, map[string]any{"claudeAiOauth": map[string]any{
			"accessToken": "old-access", "refreshToken": "old-refresh",
			"expiresAt": time.Now().Add(time.Hour).UnixMilli(), "subscriptionType": "max",
		}})}
	writeFile(t, loginsPath(), []savedLogin{main, side})
	forgetAccountCaches()
	p, ok := find(All(), "claude")
	if !ok {
		t.Fatal("Claude account missing")
	}
	return p, side
}

// sideToken is the saved account's turn at a request: the directory Claude
// Code runs on it in, or why it can't.
func sideToken(t *testing.T, p Provider, user string) (string, error) {
	t.Helper()
	for _, q := range p.AlsoOn() {
		if q.Account.User == user {
			dir, _, err := q.Account.Token(context.Background())
			return dir, err
		}
	}
	t.Fatalf("%s is not among the accounts on", user)
	return "", nil
}

func claudeLapseOf(user string) string {
	for _, l := range Logins("claude") {
		if strings.EqualFold(l.User, user) {
			return l.Lapsed
		}
	}
	return ""
}

func TestClaudeLegacyLapseIsNotRunOrBlamedOnLogout(t *testing.T) {
	p, side := claudeAuthFixture(t, legacyClaudeLogoutLapse)
	if _, err := sideToken(t, p, side.User); err == nil || !strings.Contains(err.Error(), "sign in again") {
		t.Fatalf("lapsed login run: %v", err)
	}
	if _, err := os.Stat(claudeAccountDir(side.User)); !os.IsNotExist(err) {
		t.Fatalf("a directory was made for the lapsed login: %v", err)
	}
	if got := claudeLapseOf(side.User); got != claudeLogoutLapse {
		t.Fatalf("lapse shown as %q", got)
	}
}

// A refusal holds for the credential it was made on: Claude Code refreshing
// the account's sign-in in its directory afterwards, or a run started on
// the refused credential reporting late, leaves the account usable.
func TestClaudeRefusalHoldsForItsCredentialOnly(t *testing.T) {
	p, side := claudeAuthFixture(t, "")
	refused := ClaudeLoginVersion(side.User)
	NoteClaudeSignInFailure(side.User, refused, "Failed to authenticate: OAuth session expired and could not be refreshed")
	if _, err := sideToken(t, p, side.User); err == nil || !strings.Contains(err.Error(), "sign in again") {
		t.Fatalf("refused login run: %v", err)
	}
	if err := SwitchLogin("claude", side.User); err == nil {
		t.Fatal("Claude Code signed in to a refused login")
	}
	writeFile(t, filepath.Join(claudeAccountDir(side.User), ".credentials.json"), map[string]any{"claudeAiOauth": map[string]any{
		"accessToken": "new-access", "refreshToken": "new-refresh",
		"expiresAt": time.Now().Add(8 * time.Hour).UnixMilli(), "subscriptionType": "max",
	}})
	if _, err := sideToken(t, p, side.User); err != nil {
		t.Fatalf("the sign-in Claude Code refreshed to is not used: %v", err)
	}
	NoteClaudeSignInFailure(side.User, refused, "Failed to authenticate: OAuth token revoked.")
	if _, err := sideToken(t, p, side.User); err != nil || claudeLapseOf(side.User) != "" {
		t.Fatalf("an old run's refusal holds for the new sign-in: %v", err)
	}
}

// Claude Code empties a sign-in Anthropic refused to refresh: the copy
// starling kept of it is the refused one, never put back for Claude Code.
func TestClaudeEmptiedSignInIsNotPutBack(t *testing.T) {
	p, side := claudeAuthFixture(t, "")
	dir, err := sideToken(t, p, side.User)
	if err != nil {
		t.Fatal(err)
	}
	creds := filepath.Join(dir, ".credentials.json")
	writeFile(t, creds, map[string]any{"claudeAiOauth": map[string]any{"accessToken": "", "refreshToken": "", "expiresAt": 0}})
	if err := SwitchLogin("claude", side.User); err == nil {
		t.Fatal("Claude Code signed in to the refused copy")
	}
	if _, err := sideToken(t, p, side.User); err == nil {
		t.Fatal("ran on a sign-in Claude Code emptied")
	}
	if b, _ := os.ReadFile(creds); strings.Contains(string(b), "old-refresh") {
		t.Fatal("the refused copy was put back")
	}
	if claudeLapseOf(side.User) != claudeAuthLapse {
		t.Fatalf("lapse: %q", claudeLapseOf(side.User))
	}
}

func TestClaudeClearedKeychainDoesNotFallBackToStaleFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("a shell script stands in for the keychain")
	}
	p, side := claudeAuthFixture(t, "")
	dir, err := sideToken(t, p, side.User)
	if err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	cleared := filepath.Join(bin, "cleared.json")
	writeFile(t, cleared, map[string]any{"claudeAiOauth": map[string]any{
		"accessToken": "", "refreshToken": "", "expiresAt": 0, "subscriptionType": "max",
	}})
	testenv.Program(t, filepath.Join(bin, "security"), "#!/bin/sh\ncat '"+cleared+"'\n")
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	claudeKeychain = true
	if _, err := sideToken(t, p, side.User); err == nil {
		t.Fatal("cleared keychain item passed over for the file beside it")
	}
	if b, _ := os.ReadFile(filepath.Join(dir, ".credentials.json")); !strings.Contains(string(b), "old-refresh") {
		t.Fatal("the file was rewritten")
	}
}

// Claude Code's own sign-in is not read: the accounts starling serves are
// the ones it keeps, whatever Claude Code happens to be signed in to.
func TestClaudeCodeOwnSignInIsNotRead(t *testing.T) {
	p, side := claudeAuthFixture(t, "")
	users := map[string]bool{}
	for _, l := range Logins("claude") {
		users[l.User] = true
	}
	if !users["main@example.com"] || !users[side.User] {
		t.Fatalf("the accounts starling keeps: %v", users)
	}
	// Claude Code's own sign-in is none of them, and its going away changes
	// nothing: the accounts are starling's own
	cred := filepath.Join(os.Getenv("HOME"), ".claude", ".credentials.json")
	os.Remove(cred)
	forgetAccountCaches()
	if _, _, err := p.Account.Token(context.Background()); err != nil {
		t.Fatalf("Claude Code's own sign-in is needed: %v", err)
	}
	if p.Account.User != "main@example.com" {
		t.Fatalf("the account served: %q", p.Account.User)
	}
}

func TestClaudeSignInRequiredDoesNotBlameTemporaryFailures(t *testing.T) {
	for _, tc := range []struct {
		message string
		want    bool
	}{
		{"Failed to authenticate: OAuth session expired and could not be refreshed", true},
		{"Failed to authenticate: OAuth token revoked. Please log in again or contact your administrator.", true},
		{"OAuth access token has been revoked.", true},
		{"Failed to refresh OAuth token: another Claude Code process is refreshing it or exited mid-refresh", false},
		{"OAuth token refresh failed (HTTP 503)", false},
		{"Failed to authenticate: network connection timed out", false},
		{"You've hit your session limit", false},
	} {
		if got := ClaudeSignInRequired(tc.message); got != tc.want {
			t.Errorf("%q: got %v, want %v", tc.message, got, tc.want)
		}
	}
}

func TestClaudeProbeReportsTheAccountAndMarksItsFailedLogin(t *testing.T) {
	p, side := claudeAuthFixture(t, "")
	old := claudeCLIProbe
	t.Cleanup(func() { claudeCLIProbe = old })
	claudeCLIProbe = func(context.Context, string, string) error {
		return errors.New("Failed to authenticate: OAuth session expired and could not be refreshed")
	}
	var saved Provider
	for _, q := range p.AlsoOn() {
		saved = q
	}
	result := saved.testClaude(context.Background(), "claude-sonnet-5")
	if result.OK || result.Account != side.User || !strings.Contains(result.Error, "OAuth session expired") {
		t.Fatalf("failed test lost the account or original error: %+v", result)
	}
	if _, err := sideToken(t, p, side.User); err == nil {
		t.Fatal("the login that failed its test is still run")
	}
}
