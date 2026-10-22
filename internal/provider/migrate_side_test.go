package provider

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/lulu58588/chongyun/internal/update"
)

// moveAndBack does what Move, and putBack after a failed one, do with a
// mover's accounts, but for the plugin in between: the accounts go out,
// the built-in's saved ones go aside, extra (sign-ins made through the
// plugin since) join them, and every one comes back. It gives the accounts that went out and the saved
// accounts before and after.
func moveAndBack(t *testing.T, id string, extra ...map[string]any) (out []Moving, before, after []savedLogin) {
	t.Helper()
	mv := movers[id]
	out, err := mv.out()
	if err != nil {
		t.Fatal(err)
	}
	ls := readLogins()
	before = slices.Clone(ls)
	var backup []savedLogin
	ls = slices.DeleteFunc(ls, func(l savedLogin) bool {
		if slices.Contains(mv.agents, l.Agent) {
			backup = append(backup, l)
			return true
		}
		return false
	})
	if err := writeLogins(ls); err != nil {
		t.Fatal(err)
	}
	// through the plugin, which keeps its sign-ins as JSON
	var auths []map[string]any
	for _, a := range out {
		if !a.Own {
			auths = append(auths, roundJSON(t, a.Auth))
		}
	}
	ls = readLogins()
	users := map[int]string{}
	for i, a := range out {
		if a.Own {
			continue
		}
		next, u, err := mv.back(ls, a.User, auths[0])
		auths = auths[1:]
		if err != nil {
			t.Fatalf("%s back: %v", a.User, err)
		}
		ls, users[i] = next, u
	}
	for _, a := range extra {
		next, _, err := mv.back(ls, "", roundJSON(t, a))
		if err != nil {
			t.Fatalf("back %v: %v", a, err)
		}
		ls = next
	}
	for _, b := range backup {
		if !slices.ContainsFunc(ls, func(l savedLogin) bool { return sameMoved(l, b) }) {
			ls = append(ls, b)
		}
	}
	for i, a := range out {
		if u := users[i]; u != "" && !strings.EqualFold(u, a.User) {
			t.Errorf("%s went back as %s", a.User, u)
		}
	}
	if err := writeLogins(ls); err != nil {
		t.Fatal(err)
	}
	return out, before, readLogins()
}

func roundJSON(t *testing.T, a map[string]any) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal([]byte(jsonText(a)), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func loginsOfAgent(ls []savedLogin, agent string) []savedLogin {
	return slices.DeleteFunc(slices.Clone(ls), func(l savedLogin) bool { return l.Agent != agent })
}


func TestMoveGrok(t *testing.T) {
	home := signIn(t)
	t.Setenv("GROK_HOME", filepath.Join(home, ".grok"))
	grokSignedIn(t, GrokHome(), "me@x.ai")
	extra, _ := newGrokHome()
	grokSignedIn(t, extra, "two@x.ai")
	if _, err := addGrokLogin(extra); err != nil {
		t.Fatal(err)
	}
	// a sign-in made through the plugin, in a home of the plugin's
	pluginHome := filepath.Join(home, ".cache", "opencode-grok-auth", "h1")
	grokSignedIn(t, pluginHome, "three@x.ai")

	out, before, after := moveAndBack(t, "grok",
		map[string]any{"type": "oauth", "refresh": GrokHome(), "access": "k", "expires": 0, "accountId": "me@x.ai"},
		map[string]any{"type": "oauth", "refresh": pluginHome, "access": "k-three@x.ai", "expires": 0, "accountId": "three@x.ai"})
	if len(out) != 2 || out[0].User != "me@x.ai" || !out[0].Own || out[0].Auth["refresh"] != GrokHome() ||
		out[1].User != "two@x.ai" || out[1].Own || out[1].Auth["refresh"] != extra || out[1].Auth["access"] != "k-two@x.ai" ||
		out[1].Auth["expires"].(int64) < time.Now().UnixMilli() {
		t.Fatalf("out %+v", out)
	}
	if _, err := os.Stat(filepath.Join(extra, "auth.json")); err != nil {
		t.Fatalf("the moved account's home is gone: %v", err)
	}
	ls := loginsOfAgent(after, "grok")
	if len(ls) != len(loginsOfAgent(before, "grok"))+1 {
		t.Fatalf("back %+v\nbefore %+v", ls, before)
	}
	got := map[string]string{}
	for _, l := range ls {
		got[l.User] = l.Home
	}
	if got["two@x.ai"] != extra || got["three@x.ai"] != pluginHome || got["me@x.ai"] != "" {
		t.Fatalf("homes %v", got)
	}
	if gl := grokLogins(); len(gl) != 3 {
		t.Fatalf("grok logins %+v", gl)
	}
}

func TestMoveCommandCode(t *testing.T) {
	home := claudeHome(t)
	writeFile(t, filepath.Join(home, ".commandcode", "auth.json"), cmdAuth{APIKey: "cc-own", UserID: "u1", UserName: "me"})
	writeLogins([]savedLogin{{Agent: CommandCodePlanID, User: "two", Plan: "Pro", On: true, Auth: []byte(jsonText(cmdAuth{APIKey: "cc-two", UserID: "u2", UserName: "two", KeyName: "laptop"}))}})

	out, before, after := moveAndBack(t, CommandCodePlanID,
		map[string]any{"type": "api", "key": "cc-stale", "metadata": map[string]any{"email": "me", "cli": true}},
		map[string]any{"type": "api", "key": "cc-three", "metadata": map[string]any{"email": "three", "userId": "u3"}})
	if len(out) != 2 || out[0].User != "me" || !out[0].Own || !out[0].First || out[0].Auth["key"] != "cc-own" ||
		out[1].User != "two" || out[1].Own || out[1].Auth["key"] != "cc-two" || out[1].Auth["metadata"].(map[string]any)["keyName"] != "laptop" ||
		out[0].Auth["metadata"].(map[string]any)["cli"] != true || out[1].Auth["metadata"].(map[string]any)["cli"] != nil {
		t.Fatalf("out %+v", out)
	}
	ls := loginsOfAgent(after, CommandCodePlanID)
	if len(ls) != len(loginsOfAgent(before, CommandCodePlanID))+1 {
		t.Fatalf("back %+v\nbefore %+v", ls, before)
	}
	for _, l := range ls {
		a, _ := cmdSaved(l)
		switch l.User {
		case "me":
			if !l.own() {
				t.Errorf("own %+v", l)
			}
		case "two":
			if a != (cmdAuth{APIKey: "cc-two", UserID: "u2", UserName: "two", KeyName: "laptop"}) || l.Plan != "Pro" {
				t.Errorf("two %+v %+v", l, a)
			}
		case "three":
			if a.APIKey != "cc-three" || a.UserID != "u3" || a.UserName != "three" {
				t.Errorf("three %+v", a)
			}
		default:
			t.Errorf("unexpected %+v", l)
		}
	}
	if l := cmdLogins(); len(l) != 3 {
		t.Fatalf("cmd logins %+v", l)
	}
	// 0.1.7 and before said every model took no pictures (the list says
	// nothing of them), which magpie takes over models.dev's answer
	if update.Newer("0.1.8", movers[CommandCodePlanID].min) {
		t.Fatalf("the move installs commandcode-auth %q, which marks every model text-only", movers[CommandCodePlanID].min)
	}
}

