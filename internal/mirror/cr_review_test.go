package mirror

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestCRR2Y5RelativeExcludesFromSubdirectory(t *testing.T) {
	root := repo(t)
	write(t, filepath.Join(root, "global-ignore"), "*.cache\n")
	write(t, filepath.Join(root, "saved.cache"), "ignored\n")
	write(t, filepath.Join(root, "sub", "file"), "tracked\n")
	run(t, root, "git", "config", "core.excludesFile", "global-ignore")
	if got := run(t, root, "git", "check-ignore", "saved.cache"); strings.TrimSpace(got) != "saved.cache" {
		t.Fatalf("positive control: git must ignore saved.cache, got %q", got)
	}
	t.Chdir(filepath.Join(root, "sub"))
	got, err := GlobalExcludes(context.Background(), root)
	if err != nil || got != "*.cache\n" {
		t.Fatalf("Y5: excludes must match git -C root from any invocation subdirectory: got %q, err %v; want %q", got, err, "*.cache\n")
	}
}
