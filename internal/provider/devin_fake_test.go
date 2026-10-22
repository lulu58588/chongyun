package provider

// Fakes several tests share: a CLI stood in for by a shell script, and
// Devin's own report as `devin auth status` prints it.

import (
	"path/filepath"
	"runtime"
	"testing"

)

// devinSigned is the CLI's own report, as `devin auth status` prints it.
const devinSigned = `printf 'Logged in (via Devin).\n\nUser:\n  Email:             dev@example.com\n\nAccount:\n  Tier:              Devin Pro\n'`

// devinKeyRefused is what a real CLI prints for a token Devin's servers
// refused, a revoked or expired one: it exits 0, the report still begins
// `Logged in`, and there is no account in it.
const devinKeyRefused = `printf '%s\n' 'Logged in (via Devin).' '' 'User / team info:' '  Failed to fetch from server: Authentication required: failed to get primary API key; try logging out and logging in again: failed to validate Devin token: Invalid token (trace ID: 5f0c1d2e)'`

// devinUnreachable is the same report with Devin's servers unreachable,
// which must not drop the account.
const devinUnreachable = `printf '%s\n' 'Logged in (via Devin).' '' 'User / team info:' '  Failed to fetch from server: Connection failed: Connect HTTP error: connect: connection refused'`

// shellFakes skips a test whose fake CLI is a shell script, which Windows
// can't run.
func shellFakes(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the fake CLI is a shell script")
	}
}

// fakeDevin points DevinExecutable at a fake CLI of this test's, and puts
// the CLI back when it ends.
func fakeDevin(t *testing.T, exe string) {
	t.Helper()
	shellFakes(t)
	old := DevinExecutable
	DevinExecutable = func() string { return exe }
	t.Cleanup(func() { DevinExecutable = old })
}

// fakeDevinAt is fakeDevin for a CLI in a home of its own.
func fakeDevinAt(t *testing.T, home string) string {
	t.Helper()
	exe := filepath.Join(home, "devin")
	fakeDevin(t, exe)
	return exe
}
