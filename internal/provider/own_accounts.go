package provider

// The accounts starling signed in to itself.
//
// Its own OAuth flow signs in to Claude Code and to ChatGPT (Codex), and
// keeps what it signed in to in its own logins.json. Those accounts are
// the ones served here: no other agent's sign-in is read, neither its
// credential file nor its keychain item, and no agent CLI is asked who it
// is signed in to.

import (
	"encoding/json"
	"strings"
)


// rememberAccounts brings the accounts the agents are signed in to into
// logins.json.
//
// starling signs the agents in itself, and keeps what it signed in to in its
// own logins.json, marking there the account each agent is on. So there is
// nothing to bring in: no other agent's sign-in is read, neither its
// credential file nor its keychain item, and no agent CLI is asked who it is
// signed in to.
//
// The tests set the agents' own files up the way they always did, and a
// shim there reads those for them (fixture_accounts_test.go).
var rememberAccounts = func(force bool) {}

// loginUsable says whether a saved account can still be used: its sign-in
// is there and the vendor hasn't refused it.
func loginUsable(l savedLogin) bool {
	if l.Agent == "claude" {
		return claudeSignedOut(l) == ""
	}
	return l.Lapsed == ""
}

// ownActive is the account starling signed the agent in to: the one the
// agent runs on, and the one the gateway goes to first. It is read from
// starling's own logins.json, where starling marks it when it signs the
// agent in, never from the agent's own store.
//
// A logins.json written before that mark was kept has none: the account
// turned on in it (on) is taken for the one signed in to, so the accounts
// starling already had keep working.
//
// ok is false when the agent has none.
func ownActive(agent string) (savedLogin, bool) {
	var out savedLogin
	best := 0
	for _, l := range readLogins() {
		if l.Agent != agent {
			continue
		}
		rank := 0
		switch {
		case l.Active && loginUsable(l):
			rank = 3
		case l.Active:
			rank = 2
		case l.On && loginUsable(l):
			rank = 1
		}
		if rank > best || rank == best && rank > 0 && l.Seen.After(out.Seen) {
			out, best = l, rank
		}
	}
	return out, best > 0
}

// setOwnActive marks the account starling signed the agent in to, and takes
// the mark off the others of that agent. The list is the one to write. An
// account not in the list is left as it is: the one before it keeps the
// mark.
func setOwnActive(ls []savedLogin, agent, user string) []savedLogin {
	found := false
	for i := range ls {
		if ls[i].Agent == agent && strings.EqualFold(ls[i].User, user) {
			found = true
		}
	}
	if !found {
		return ls
	}
	for i := range ls {
		if ls[i].Agent == agent {
			ls[i].Active = strings.EqualFold(ls[i].User, user)
		}
	}
	return ls
}

// CodexAPIKeySignedIn says the ChatGPT account starling serves first is
// signed in with an OpenAI API key rather than a ChatGPT account
// (auth.json's auth_mode): its sign-in is a key's, and spends no account a
// gateway key's list governs.
func CodexAPIKeySignedIn() bool {
	l, ok := ownActive("codex")
	if !ok {
		return false
	}
	var a codexAuth
	return json.Unmarshal(l.Auth, &a) == nil && a.AuthMode == "apikey"
}

// CodexChatGPTSignedIn says the ChatGPT account starling signed Codex in
// to is signed in with a ChatGPT account, not with an API key.
//
// It is starling's own account, kept in its own logins.json, that is
// asked: Codex's own auth.json is never read, so a ChatGPT account signed
// in to in Codex itself, outside starling, is not told here.
func CodexChatGPTSignedIn() bool {
	l, ok := ownActive("codex")
	if !ok {
		return false
	}
	var a codexAuth
	if json.Unmarshal(l.Auth, &a) != nil {
		return false
	}
	if a.AuthMode == "apikey" {
		return false
	}
	return a.Tokens.AccessToken != ""
}

// CodexSignedInAccount is the ChatGPT account id (chatgpt-account-id) of
// the account starling serves first; "" when it has none.
func CodexSignedInAccount() string {
	l, ok := ownActive("codex")
	if !ok {
		return ""
	}
	id, _ := codexSessionIdentity(l.Auth)
	return id.AccountID
}
