package remote

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// K5: cancellation of the build command must deliver TERM, not immediate/later KILL alone.
func TestR7CommandCancellationDeliversTERM(t *testing.T) {
	dir := t.TempDir()
	ready, term := filepath.Join(dir, "ready"), filepath.Join(dir, "term")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := Command(ctx, "bash", "-c", `trap 'printf TERM > "$1"; exit 0' TERM; printf ready > "$2"; while :; do sleep 0.1; done`, "_", term, ready)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(20 * time.Second)
	for {
		if _, err := os.Stat(ready); err == nil {
			break
		}
		if time.Now().After(deadline) {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
			t.Fatal("command never became ready")
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	_ = cmd.Wait()
	got, err := os.ReadFile(term)
	if err != nil || string(got) != "TERM" {
		t.Fatalf("cancel did not deliver TERM: marker=%q err=%v", got, err)
	}
}
