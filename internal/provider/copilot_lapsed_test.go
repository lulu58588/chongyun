package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// lapsedCopilot is a machine with the editors signed in to octocat (gho_x,
// taken) and magpie to hubot (gho_kept), whose trade for a session token
// answers what status says.
func lapsedCopilot(t *testing.T) *atomic.Int32 {
	t.Helper()
	signIn(t)
	var status atomic.Int32
	status.Store(200)
	exchange := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "token gho_kept" && status.Load() != 200 {
			w.WriteHeader(int(status.Load()))
			w.Write([]byte(`{"message":"Bad credentials","documentation_url":"https://docs.github.com/rest","status":"401"}`))
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"token": "tid", "expires_at": time.Now().Add(time.Hour).Unix()})
	}))
	t.Cleanup(exchange.Close)
	old := CopilotTokenURL
	CopilotTokenURL = exchange.URL
	t.Cleanup(func() { CopilotTokenURL = old })
	resetCopilotState()
	t.Cleanup(resetCopilotState)
	copilotLoginList()
	if err := addCopilotLogin("hubot", "Pro", "gho_kept", ""); err != nil {
		t.Fatal(err)
	}
	return &status
}

func resetCopilotState() {
	copilotMu.Lock()
	copilotSessions = map[string]copilotSession{}
	copilotMu.Unlock()
	copilotRefusals.Lock()
	copilotRefusals.at, copilotRefusals.status = map[string]time.Time{}, map[string]int{}
	copilotRefusals.Unlock()
	copilotLapsedMu.Lock()
	copilotLapsedSeen = map[string]bool{}
	copilotLapsedMu.Unlock()
}

func copilotLapsedOf(t *testing.T, user string) string {
	t.Helper()
	for _, l := range Logins("copilot") {
		if strings.EqualFold(l.User, user) {
			return l.Lapsed
		}
	}
	t.Fatalf("no copilot account %s in %+v", user, Logins("copilot"))
	return ""
}

// copilotTrade sends a request on user's account the way the gateway
// does, and says what it got.
func copilotTrade(t *testing.T, user string) error {
	t.Helper()
	for _, c := range copilotLogins(copilotConfigDir()) {
		if strings.EqualFold(c.User, user) {
			copilotMu.Lock()
			copilotSessions = map[string]copilotSession{}
			copilotMu.Unlock()
			p := copilotProvider(c.app, c.Plan)
			req, _ := http.NewRequest("POST", "https://api.githubcopilot.com/chat/completions", nil)
			return p.Sign(context.Background(), req, Chat, []byte(`{"messages":[{"role":"user","content":"hi"}]}`))
		}
	}
	t.Fatalf("no copilot account %s", user)
	return nil
}

// echo_ts 报告: GitHub revoked a Copilot account's token, and the
// account still read as signed in while every request failed. Refused
// with 401 (or 403), it is signed out on its own row; signed in again in
// place it is in use as it was, its mark gone.
func TestCopilotRefusedTokenSignsTheAccountOut(t *testing.T) {
	for _, code := range []int{401, 403} {
		t.Run(http.StatusText(code), func(t *testing.T) {
			status := lapsedCopilot(t)
			if err := copilotTrade(t, "hubot"); err != nil {
				t.Fatalf("taken: %v", err)
			}
			if l := copilotLapsedOf(t, "hubot"); l != "" {
				t.Fatalf("lapsed while taken: %q", l)
			}
			status.Store(int32(code))
			if err := copilotTrade(t, "hubot"); err == nil || !strings.Contains(err.Error(), "signed out") {
				t.Fatalf("refused: %v", err)
			}
			if l := copilotLapsedOf(t, "hubot"); l != copilotLapsedText {
				t.Fatalf("refused %d, the account reads %q", code, l)
			}
			if l := copilotLapsedOf(t, "octocat"); l != "" {
				t.Fatalf("the editors' octocat, still taken, reads %q", l)
			}
			// marked in logins.json: it reads so after a restart as well
			resetCopilotState()
			if l := copilotLapsedOf(t, "hubot"); l != copilotLapsedText {
				t.Fatalf("after a restart the account reads %q", l)
			}
			var before Login
			for _, l := range Logins("copilot") {
				if l.User == "hubot" {
					before = l
				}
			}
			// signed in again from its row: the new token, in use as before
			if err := addCopilotLogin("hubot", "Pro", "gho_new", ""); err != nil {
				t.Fatal(err)
			}
			if l := copilotLapsedOf(t, "hubot"); l != "" {
				t.Fatalf("signed in again, the account reads %q", l)
			}
			// nothing of the old token's mark is kept in logins.json
			for _, l := range readLogins() {
				if l.Agent == "copilot" && l.User == "hubot" && (l.Lapsed != "" || l.Refused != "") {
					t.Fatalf("signed in again, saved %+v", l)
				}
			}
			n := 0
			for _, l := range Logins("copilot") {
				if l.User == "hubot" {
					n++
					if l.On != before.On || l.Active != before.Active || l.Plan != before.Plan {
						t.Fatalf("signed in again: %+v, was %+v", l, before)
					}
				}
			}
			if n != 1 {
				t.Fatalf("hubot listed %d times after signing in again", n)
			}
			if err := copilotTrade(t, "hubot"); err != nil {
				t.Fatalf("the new sign-in: %v", err)
			}
		})
	}
}

// Only GitHub's refusal of the token signs an account out: a 5xx, a 429 or
// a network error leave it as it was.
func TestCopilotTransientErrorsDontSignOut(t *testing.T) {
	status := lapsedCopilot(t)
	for _, code := range []int32{500, 502, 503, 429} {
		status.Store(code)
		if err := copilotTrade(t, "hubot"); err == nil {
			t.Fatalf("%d: no error", code)
		}
		if l := copilotLapsedOf(t, "hubot"); l != "" {
			t.Fatalf("%d signed the account out: %q", code, l)
		}
	}
	old := CopilotTokenURL
	dead := httptest.NewServer(http.NotFoundHandler())
	CopilotTokenURL = dead.URL
	dead.Close()
	if err := copilotTrade(t, "hubot"); err == nil {
		t.Fatal("unreachable: no error")
	}
	CopilotTokenURL = old
	if l := copilotLapsedOf(t, "hubot"); l != "" {
		t.Fatalf("a network error signed the account out: %q", l)
	}
}

// Marked signed out, an account GitHub takes again (a request served on
// the same token) is in use again without a sign-in.
func TestCopilotServedAgainClearsTheMark(t *testing.T) {
	status := lapsedCopilot(t)
	status.Store(401)
	copilotTrade(t, "hubot")
	if copilotLapsedOf(t, "hubot") == "" {
		t.Fatal("not marked")
	}
	resetCopilotState()
	status.Store(200)
	if err := copilotTrade(t, "hubot"); err != nil {
		t.Fatal(err)
	}
	if l := copilotLapsedOf(t, "hubot"); l != "" {
		t.Fatalf("served again, the account reads %q", l)
	}
}



// The account's allowance read is GitHub's answer too: 401 there, nothing
// in its place, signs it out and says so on its row; a 503 doesn't; read
// again, it is signed in again.
func TestCopilotUsageRefusedSignsOut(t *testing.T) {
	lapsedCopilot(t)
	var status atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "token gho_kept" && status.Load() != 200 {
			w.WriteHeader(int(status.Load()))
			return
		}
		w.Write([]byte(`{"copilot_plan":"individual","quota_snapshots":{"chat":{"unlimited":true},"completions":{"unlimited":true},"premium_interactions":{"has_quota":true,"entitlement":300,"quota_remaining":300}}}`))
	}))
	defer srv.Close()
	old := CopilotUserURL
	CopilotUserURL = srv.URL
	defer func() { CopilotUserURL = old }()

	status.Store(503)
	if q := loginQuota(context.Background(), Login{Agent: "copilot", User: "hubot"}); q.Error == "" || q.Error == copilotLapsedText {
		t.Fatalf("503: %+v", q)
	}
	if l := copilotLapsedOf(t, "hubot"); l != "" {
		t.Fatalf("503 signed the account out: %q", l)
	}
	status.Store(401)
	if q := loginQuota(context.Background(), Login{Agent: "copilot", User: "hubot"}); q.Error != copilotLapsedText {
		t.Fatalf("401: %+v", q)
	}
	if l := copilotLapsedOf(t, "hubot"); l != copilotLapsedText {
		t.Fatalf("401, the account reads %q", l)
	}
	if l := copilotLapsedOf(t, "octocat"); l != "" {
		t.Fatalf("the editors' octocat reads %q", l)
	}
	status.Store(200)
	if q := loginQuota(context.Background(), Login{Agent: "copilot", User: "hubot"}); q.Error != "" {
		t.Fatalf("200: %+v", q)
	}
	if l := copilotLapsedOf(t, "hubot"); l != "" {
		t.Fatalf("read again, the account reads %q", l)
	}
}

