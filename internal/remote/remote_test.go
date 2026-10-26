package remote

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// X2: CMD words survive `bash -lc` quoting.
func TestJoinRoundTripsThroughBash(t *testing.T) {
	words := []string{"plain", "", "a b", "it's", `"dq"`, "$HOME", "`x`", "a\nb", "-n", "*", "~", "\\", "é", "x;y|z&"}
	out, err := exec.Command("bash", "-c", "printf '%s\\0' "+Join(words)).Output()
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	start := 0
	for i, b := range out {
		if b == 0 {
			got = append(got, string(out[start:i]))
			start = i + 1
		}
	}
	if len(got) != len(words) {
		t.Fatalf("got %d words %q", len(got), got)
	}
	for i := range words {
		if got[i] != words[i] {
			t.Errorf("word %d: got %q want %q", i, got[i], words[i])
		}
	}
}

func TestTargetCarriesOptions(t *testing.T) { // X5: SSHOptions + the target's options + -T, everywhere
	tg := Target{Address: "u@h", Opts: []string{"-o", "ControlMaster=auto"}}
	args := tg.SSHArgs("true")
	want := append(append(append([]string{}, SSHOptions...), "-o", "ControlMaster=auto"), "-T", "--", "u@h", "true")
	if !slices.Equal(args, want) {
		t.Errorf("SSHArgs = %q", args)
	}
	if got := tg.RsyncShell(); got != "ssh "+strings.Join(append(append(append([]string{}, SSHOptions...), "-o", "ControlMaster=auto"), "-T"), " ") {
		t.Errorf("RsyncShell = %q", got)
	}
}

func TestMultiplex(t *testing.T) { // X5
	short, err := os.MkdirTemp("/tmp", "ox")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(short) })
	env := func(rt string) func(string) string {
		return func(k string) string {
			if k == "XDG_RUNTIME_DIR" {
				return rt
			}
			return ""
		}
	}
	opts, err := Multiplex(env(short))
	dir := filepath.Join(short, "offload", "ssh")
	want := []string{"-o", "ControlMaster=auto", "-o", "ControlPath=" + dir + "/%C", "-o", "ControlPersist=60"}
	if fi, statErr := os.Stat(dir); err != nil || !slices.Equal(opts, want) || statErr != nil || fi.Mode().Perm() != 0o700 {
		t.Fatalf("opts %q err %v dir %v %v", opts, err, fi, statErr)
	}
	at := short + "/" + strings.Repeat("r", maxControlDir-len(short+"/offload/ssh")-1) // dir exactly at the bound
	if len(filepath.Join(at, "offload", "ssh")) != maxControlDir {
		t.Fatalf("bound fixture is %d bytes", len(filepath.Join(at, "offload", "ssh")))
	}
	if _, err := Multiplex(env(at)); err != nil {
		t.Errorf("dir at the %d-byte bound refused: %v", maxControlDir, err)
	}
	for name, rt := range map[string]string{
		"unset": "", "space": short + "/a b", "percent": short + "/a%h", "one past the bound": at + "x",
	} {
		if opts, err := Multiplex(env(rt)); err == nil {
			t.Errorf("%s: opts %q, want refusal", name, opts)
		}
	}
}
