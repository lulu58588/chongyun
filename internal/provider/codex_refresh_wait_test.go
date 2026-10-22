package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// A Codex refresh OpenAI is answering is waited for though whoever asked
// gave up: the answer holds the rotated refresh token, and dropping it
// leaves the saved account with the spent one, refused on the next
// refresh (willz 报告: a saved account read as signed out while its
// requests still went through).
func TestCodexRefreshOutlivesItsCaller(t *testing.T) {
	signIn(t)
	codexSignIn(t, "work@example.com", "r-work")

	arrived, release := make(chan struct{}), make(chan struct{})
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		close(arrived)
		<-release // OpenAI has rotated the token; its answer is on the way
		json.NewEncoder(w).Encode(map[string]any{"access_token": fakeJWT(map[string]any{"exp": float64(time.Now().Add(time.Hour).Unix())}),
			"refresh_token": body["refresh_token"] + "-2"})
	}))
	defer fake.Close()
	defer CodexTokenURLForTest(fake.URL)()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, _, err := renewSavedLogin(ctx, "codex", "me@example.com", true)
		done <- err
	}()
	<-arrived
	cancel() // the request that asked for it was cancelled
	close(release)
	if err := <-done; err != nil {
		t.Fatalf("refresh: %v", err)
	}
	for _, l := range readLogins() {
		if l.User == "me@example.com" && !strings.Contains(string(l.Auth), `"r-2"`) {
			t.Fatalf("the rotated refresh token was dropped: %s", l.Auth)
		}
	}

	// a caller already gone sends nothing
	gone, stop := context.WithCancel(context.Background())
	stop()
	if _, _, err := renewSavedLogin(gone, "codex", "me@example.com", true); err == nil {
		t.Fatal("a caller already gone refreshed")
	}
}
