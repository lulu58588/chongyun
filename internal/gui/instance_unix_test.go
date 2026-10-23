//go:build !windows

package gui

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// A second starling on the same config finds the first one's lock held, a
// copy elsewhere (starling-dev) has its own, and a lock whose holder ended,
// as when the Mac shuts down with starling running, is free again (Crispin
// 报告: two starlings after each boot).
func TestOneStarlingAtATime(t *testing.T) {
	dir := t.TempDir()
	if os.Getenv("STARLING_HOLD_LOCK") != "" {
		if held, err := holdInstance(os.Getenv("STARLING_HOLD_LOCK")); err != nil || !held {
			os.Exit(3)
		}
		os.Stdout.WriteString("held\n")
		select {}
	}
	lock := instanceLock(dir, "/Applications/Starling.app/Contents/MacOS/starling")
	if dev := instanceLock(dir, "/src/starling/starling-dev"); dev == lock {
		t.Fatal("starling-dev shares the app's lock")
	}
	// every installed copy shares the lock: another starling.app (one the
	// Mac reopens, one run translocated from Downloads) and Homebrew's of
	// any version (inaction 报告: two birds after a restart)
	for _, other := range []string{
		"/Users/u/Downloads/starling.app/Contents/MacOS/starling",
		"/private/var/folders/x/T/AppTranslocation/1F2E/d/starling.app/Contents/MacOS/starling",
		"/opt/homebrew/Cellar/starling/0.1.900/bin/starling",
	} {
		if instanceLock(dir, other) != lock {
			t.Errorf("%s has a lock of its own", other)
		}
	}
	if filepath.Dir(lock) != dir {
		t.Fatalf("lock %s not in the config dir", lock)
	}

	// the first starling: another process, killed as a shutdown would
	cmd := exec.Command(os.Args[0], "-test.run=^TestOneStarlingAtATime$")
	cmd.Env = append(os.Environ(), "STARLING_HOLD_LOCK="+lock)
	out, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 5)
	if _, err := out.Read(buf); err != nil || string(buf[:4]) != "held" {
		cmd.Process.Kill()
		t.Fatalf("the first starling didn't take the lock: %q %v", buf, err)
	}
	if held, err := holdInstance(lock); err != nil || held {
		cmd.Process.Kill()
		t.Fatalf("a second starling took the held lock: %v %v", held, err)
	}
	cmd.Process.Kill()
	cmd.Wait()
	if held, err := holdInstance(lock); err != nil || !held {
		t.Fatalf("the lock of a starling that was killed stayed held: %v %v", held, err)
	}
	heldInstance.Close()
	heldInstance = nil
}
