package provider

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/lulu58588/chongyun/internal/steady"
)

// fakeCodexRefresh answers a refresh with a fresh access token and the
// refresh token rotated to "<old>-2", as OpenAI does.
func fakeCodexRefresh(t *testing.T) {
	t.Helper()
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		json.NewEncoder(w).Encode(map[string]any{
			"access_token":  fakeJWT(map[string]any{"exp": float64(time.Now().Add(time.Hour).Unix())}),
			"refresh_token": body["refresh_token"] + "-2"})
	}))
	t.Cleanup(fake.Close)
	old := codexTokenURL
	codexTokenURL = fake.URL
	t.Cleanup(func() { codexTokenURL = old })
}

// dueCodexAuth writes an auth.json, as Codex CLI does, whose access token
// is about to expire, so the next codexToken refreshes it.
func dueCodexAuth(t *testing.T) (path string, was []byte) {
	t.Helper()
	path = filepath.Join(t.TempDir(), ".codex", "auth.json")
	writeFile(t, path, map[string]any{
		"OPENAI_API_KEY": nil,
		"tokens": map[string]any{
			"access_token":  fakeJWT(map[string]any{"exp": float64(time.Now().Add(time.Minute).Unix())}),
			"refresh_token": "r-work",
			"account_id":    "acct-1",
		},
		"last_refresh": "2026-10-14T00:00:00Z",
	})
	was, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return path, was
}


// The same for Claude Code's own .credentials.json, written when magpie
// switches the sign-in Claude Code is on: a write whose flush failed is
// an error and leaves the file as it was.
func TestClaudeCredentialFlushesBeforeReplacing(t *testing.T) {
	isolate(t)
	defer func() { steady.Sync = (*os.File).Sync }()
	path := filepath.Join(t.TempDir(), ".claude", ".credentials.json")
	writeFile(t, path, map[string]any{"claudeAiOauth": map[string]any{"accessToken": "sk-ant-oat01-old", "refreshToken": "sk-ant-ort01-old"}})
	was, _ := os.ReadFile(path)
	c, ok := parseClaudeCredentials([]byte(`{"claudeAiOauth":{"accessToken":"sk-ant-oat01-new","refreshToken":"sk-ant-ort01-new","expiresAt":1999999999000}}`))
	if !ok {
		t.Fatal("credentials don't parse")
	}
	loc := claudeCredentialLocation{path: path}

	steady.Sync = func(*os.File) error { return errors.New("the disk went away") }
	if err := writeClaudeCredential(loc, c); err == nil {
		t.Error("a write whose flush failed reported as saved")
	}
	if b, _ := os.ReadFile(path); string(b) != string(was) {
		t.Fatalf("a flush that failed left the credentials %q, want them as they were", b)
	}

	flushed := false
	steady.Sync = func(f *os.File) error {
		if b, _ := os.ReadFile(path); string(b) != string(was) {
			t.Errorf("the credentials were changed before the new ones were flushed: %q", b)
		}
		flushed = true
		return f.Sync()
	}
	if err := writeClaudeCredential(loc, c); err != nil {
		t.Fatal(err)
	}
	if !flushed {
		t.Fatal("the credentials file was written without the new sign-in being flushed to the disk first")
	}
	if b, _ := os.ReadFile(path); !strings.Contains(string(b), "sk-ant-ort01-new") {
		t.Fatalf("written: %q", b)
	}
}
