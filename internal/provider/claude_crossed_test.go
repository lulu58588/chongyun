package provider

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
	"time"
)

// claudeThree saves three Claude accounts as chongyun signs in to each,
// last the Pro, and gives the helpers that change them.
//
// chongyun keeps the accounts it signed in to in its own logins.json,
// marking there the one the agent is on: the account signed in to last is
// that one, and a switch moves the mark to another. The helpers keep the
// mark where a switch left it.
func claudeThree(t *testing.T) (prof func(email, org string), creds func(tok, plan string)) {
	t.Helper()
	claudeHome(t)
	type account struct {
		email, org, plan, tok string
		renewed               bool // the sign-in Claude Code made afresh
	}
	var accounts []account
	// newest is the account signed in to last: the one the agent is on,
	// until a switch moves the mark to another
	newest, marked := "", ""
	save := func() {
		loginsMu.Lock()
		was := readLogins()
		loginsMu.Unlock()
		active := ""
		for _, l := range was {
			if l.Agent == "claude" && l.Active {
				active = claudeEmail(l.User)
			}
		}
		if active == "" || active == marked {
			active = newest
		}
		ls := slices.Clone(was)
		for _, a := range accounts {
			acct := map[string]any{"emailAddress": a.email,
				"organizationUuid": "org-" + a.email, "organizationName": a.org}
			blob, _ := json.Marshal(map[string]any{"claudeAiOauth": map[string]any{
				"accessToken": "tok-" + a.tok, "refreshToken": "r-" + a.tok,
				"expiresAt": time.Now().Add(time.Hour).UnixMilli(), "subscriptionType": a.plan,
				"scopes": []string{"user:inference", "user:profile"},
			}})
			c, ok := parseClaudeCredentials(blob)
			if !ok {
				t.Fatal("the Claude sign-in made for the test can't be read")
			}
			l, err := claudeLogin(c, acct)
			if err != nil {
				t.Fatal(err)
			}
			l.On = true
			l.Active = a.email == active
			if !a.renewed {
				// the sign-in the record holds, as it is: a record
				// crossed before keeps what it holds
				for _, w := range was {
					if w.Agent == "claude" && claudeEmail(w.User) == a.email {
						l.Auth = w.Auth
					}
				}
			}
			ls = upsertLogin(ls, l)
		}
		writeFile(t, loginsPath(), ls)
		forgetAccountCaches()
		marked = active
	}
	prof = func(email, org string) {
		for i := range accounts {
			if accounts[i].email == email {
				accounts[i].org = org
			}
		}
		save()
	}
	creds = func(tok, plan string) {
		for i := range accounts {
			if strings.HasPrefix(tok, accounts[i].email) {
				if accounts[i].tok != tok {
					accounts[i].renewed = true
				}
				accounts[i].plan, accounts[i].tok = plan, tok
			}
		}
		save()
	}
	for _, a := range [][3]string{{"team@x.com", "team", "Acme"}, {"max@x.com", "max", "max@x.com's Organization"}, {"pro@x.com", "pro", "pro@x.com's Organization"}} {
		accounts = append(accounts, account{email: a[0], org: a[2], plan: a[1], tok: a[0]})
		newest = a[0]
		prof(a[0], a[2])
		creds(a[0], a[1])
	}
	return prof, creds
}

func claudeEmail(user string) string {
	email, _, _ := strings.Cut(user, " · ")
	return email
}

// netfishx 报告: Pro, Max and Team accounts, Claude Code kept signed in
// to one. magpie switches Claude Code to the Team account, and a Claude
// Code started on the Pro a moment before writes the Pro's profile after:
// the sign-in is the Team's, the profile the Pro's. That isn't saved as
// the Pro's, whose runs (its own config directory) would then spend the
// Team and show the Team's plan and usage on the Pro's card.
func TestClaudeSwitchThenStaleProfileKeepsRecords(t *testing.T) {
	prof, _ := claudeThree(t)
	team := "team@x.com · Acme"
	if err := SwitchLogin("claude", team); err != nil {
		t.Fatal(err)
	}
	prof("pro@x.com", "pro@x.com's Organization")
	time.Sleep(10 * time.Millisecond)
	forgetAccountCaches()
	if err := SwitchLogin("claude", team); err != nil { // kept: switched back
		t.Fatal(err)
	}
	forgetAccountCaches()
	plans := map[string]string{"team@x.com": "team", "max@x.com": "max", "pro@x.com": "pro"}
	for _, l := range readLogins() {
		if l.Agent != "claude" {
			continue
		}
		c, _ := parseClaudeCredentials(l.Auth)
		if e := claudeEmail(l.User); c.OAuth.AccessToken != "tok-"+e || l.Plan != plans[e] {
			t.Errorf("%s holds %s, plan %q", l.User, c.OAuth.AccessToken, l.Plan)
		}
	}
	for _, l := range Logins("claude") {
		if l.Active {
			continue
		}
		dir, err := claudeSavedDir(l.User)
		if err != nil {
			t.Fatal(err)
		}
		if d, _ := readClaudeDir(dir); d.OAuth.AccessToken != "tok-"+claudeEmail(l.User) {
			t.Errorf("runs on %s use %s", l.User, d.OAuth.AccessToken)
		}
	}
}

// Records crossed before (the Pro's holding the Team's sign-in) don't stop
// the Team's own sign-in being saved: the Team's record holding it too, or
// Claude Code having renewed it since.
func TestClaudeCrossedRecordDoesntHoldTheOwnerBack(t *testing.T) {
	prof, creds := claudeThree(t)
	loginsMu.Lock()
	ls := readLogins()
	var teamAuth []byte
	for _, l := range ls {
		if strings.HasPrefix(l.User, "team@") {
			teamAuth = l.Auth
		}
	}
	for i := range ls {
		if strings.HasPrefix(ls[i].User, "pro@") {
			ls[i].Auth = teamAuth
		}
	}
	if err := writeLogins(ls); err != nil {
		t.Fatal(err)
	}
	loginsMu.Unlock()
	for _, tok := range []string{"team@x.com", "team@x.com-renewed"} {
		prof("team@x.com", "Acme")
		creds(tok, "team")
		forgetAccountCaches()
		for _, l := range readLogins() {
			if c, _ := parseClaudeCredentials(l.Auth); strings.HasPrefix(l.User, "team@") && c.OAuth.AccessToken != "tok-"+tok {
				t.Errorf("Team's record holds %s, Claude Code %s", c.OAuth.AccessToken, "tok-"+tok)
			}
		}
	}
}
