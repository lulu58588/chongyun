package provider

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/lulu58588/starling/internal/claudecode"
	"github.com/lulu58588/starling/internal/testenv"
)

// A provider switched off, or whose account was removed from starling, gets
// no request from starling in the background: its allowance isn't read, its
// sign-in isn't renewed or switched, its list isn't fetched (Riko 报告: with
// Claude switched off, Surge showed `claude auth status --json` every
// minute). Fake claude and devin CLIs on PATH log every run: starling asks
// no agent CLI who is signed in, so none of them is ever run — switched off
// or not.
func TestLeftAloneRunsNoCLI(t *testing.T) {
	shellFakes(t)
	findDevin := DevinExecutable

	bin, log := t.TempDir(), filepath.Join(t.TempDir(), "runs")
	for _, name := range []string{"claude", "devin"} {
		testenv.Program(t, filepath.Join(bin, name), "#!/bin/sh\necho \""+name+" $*\" >> "+log+"\n"+
			"echo '{\"loggedIn\":true,\"email\":\"me@example.com\",\"subscriptionType\":\"max\"}'\n")
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+"/usr/bin"+string(os.PathListSeparator)+"/bin")
	// the real lookups, which find them on PATH
	claudeExecutable, DevinExecutable = claudecode.Find, findDevin

	runs := func() []string {
		b, _ := os.ReadFile(log)
		return strings.Fields(strings.ReplaceAll(strings.TrimSpace(string(b)), " ", "_"))
	}
	loop := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		for range 3 {
			All()
			Hidden()
			fetchSubscriptionUsage(ctx)
			for _, agent := range []string{"claude", "devin"} {
				LoginUsage(ctx, agent)
			}
			RenewLogins(ctx, 0)
			SwitchWhenSpent(ctx, "claude")
			FetchNew(time.Second)
		}
	}

	save := func(off, hidden bool) {
		writeFile(t, Path(), map[string]any{"providers": []map[string]any{
			{"id": "claude", "off": off}, {"id": "devin", "hidden": hidden},
		}})
	}
	// switched off or removed: no agent CLI is run
	save(true, true)
	loop()
	if got := runs(); len(got) != 0 {
		t.Fatalf("switched off or removed, the CLIs were run: %v", got)
	}

	// switched on: still none. No agent's sign-in is read, so no agent's
	// CLI is ever asked who is signed in.
	save(false, false)
	loop()
	if got := runs(); len(got) != 0 {
		t.Fatalf("switched on, the CLIs were run: %v", got)
	}
}
