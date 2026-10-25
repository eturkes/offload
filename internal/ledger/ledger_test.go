package ledger

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"strings"
	"syscall"
	"testing"
	"time"
)

func env(vars map[string]string) func(string) string { return func(k string) string { return vars[k] } }

func put(base, path string, line []byte) error {
	return appendLine(base, path, func() ([]byte, error) { return line, nil })
}

func final(rc int, sig string) func() (int, string) { return func() (int, string) { return rc, sig } }

func lines(t *testing.T, path string) []map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var out []map[string]any
	for line := range strings.Lines(string(data)) {
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("%v: %q", err, line)
		}
		out = append(out, rec)
	}
	return out
}

func TestJ1StateBase(t *testing.T) {
	for _, tc := range []struct {
		xdg, home, want string
	}{
		{"/s", "/h", "/s"},
		{"", "/h", "/h/.local/state"},
		{"rel/state", "/h", "/h/.local/state"}, // XDG: a relative path is ignored
		{"", "", ""},
		{"", "rel", ""},
	} {
		got, err := stateBase(env(map[string]string{"XDG_STATE_HOME": tc.xdg, "HOME": tc.home}))
		if got != tc.want || (err != nil) != (tc.want == "") {
			t.Errorf("xdg=%q home=%q → %q, %v; want %q", tc.xdg, tc.home, got, err, tc.want)
		}
	}
}

func TestJ2WriterPassesBytesAndKeepsLines(t *testing.T) {
	var stderr bytes.Buffer
	r := Begin(env(map[string]string{"XDG_STATE_HOME": t.TempDir()}), "/w", []string{"x"}, Options{}, &stderr)
	w := r.Writer()
	for _, chunk := range []string{"offload: a\n\noffload: b", "c\n", "x"} {
		if n, err := w.Write([]byte(chunk)); n != len(chunk) || err != nil {
			t.Fatalf("Write(%q) = %d, %v", chunk, n, err)
		}
	}
	if got := stderr.String(); got != "offload: a\n\noffload: bc\nx" {
		t.Fatalf("stderr = %q", got)
	}
	r.End(final(0, ""))
	rec := lines(t, r.path)[0]
	if got, want := fmt.Sprint(rec["messages"]), fmt.Sprint([]any{"offload: a", "", "offload: bc", "x"}); got != want {
		t.Fatalf("messages = %s, want %s", got, want)
	}
}

type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, errors.New("stderr gone") }

func TestJ2WriterReturnsStderrError(t *testing.T) {
	r := Begin(env(map[string]string{"XDG_STATE_HOME": t.TempDir()}), "/w", []string{"x"}, Options{}, failWriter{})
	if n, err := r.Writer().Write([]byte("offload: a\n")); n != 0 || err == nil || err.Error() != "stderr gone" {
		t.Fatalf("Write = %d, %v; want stderr's own 0, error", n, err)
	}
}

func TestJ3J4Defaults(t *testing.T) {
	if maxSize != 64<<20 || lockTimeout != 2*time.Second {
		t.Fatalf("maxSize %d, lockTimeout %v; want 64 MiB, 2 s", maxSize, lockTimeout)
	}
}

func TestJ2SecondsKeepMilliseconds(t *testing.T) {
	for d, want := range map[time.Duration]float64{
		1234567 * time.Microsecond: 1.235, 999 * time.Millisecond: 0.999, 7200*time.Second + 400*time.Microsecond: 7200,
	} {
		if got := *seconds(d); got != want {
			t.Errorf("seconds(%v) = %v, want %v", d, got, want)
		}
	}
}

func TestJ3RotatesPastMaxSize(t *testing.T) {
	old := maxSize
	t.Cleanup(func() { maxSize = old })
	maxSize = 20
	dir := t.TempDir()
	path := filepath.Join(dir, "offload", "runs.jsonl")
	write := func(line string) {
		t.Helper()
		if err := put(dir, path, []byte(line+"\n")); err != nil {
			t.Fatal(err)
		}
	}
	files := func(log, rotated string) {
		t.Helper()
		for file, want := range map[string]string{path: log, path + ".1": rotated} {
			if got, err := os.ReadFile(file); err != nil || string(got) != want {
				t.Errorf("%s = %q, %v; want %q", filepath.Base(file), got, err, want)
			}
		}
	}
	write(`"seed"`)
	if err := os.Rename(path, path+".1"); err != nil {
		t.Fatal(err)
	}
	big := `"aaaaaaaaaaaaaaaaaaaaaaa"` // 26 bytes > maxSize
	write(big)                         // an empty file takes any record: no rotation, .1 kept
	files(big+"\n", `"seed"`+"\n")
	write(`"b"`) // 26 + 4 > 20 → rotate, replacing .1
	files(`"b"`+"\n", big+"\n")
	exact := `"ccccccccccccc"` // 4 + 16 = 20 = maxSize → append
	write(exact)
	files(`"b"`+"\n"+exact+"\n", big+"\n")
	write(`"d"`) // 20 + 4 > 20 → rotate
	files(`"d"`+"\n", `"b"`+"\n"+exact+"\n")
}

func TestJ3OneByteLogIsNonEmpty(t *testing.T) {
	old := maxSize
	t.Cleanup(func() { maxSize = old })
	maxSize = 20
	dir := t.TempDir()
	path := filepath.Join(dir, "offload", "runs.jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := put(dir, path, []byte(strings.Repeat("x", 19)+"\n")); err != nil { // 1 + 20 > 20
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(path + ".1"); string(got) != "\n" {
		t.Errorf(".1 = %q, want the 1-byte log", got)
	}
}

func TestJ4FailuresReportTheirCause(t *testing.T) {
	old := maxSize
	t.Cleanup(func() { maxSize = old })
	maxSize = 20
	for _, tc := range []struct {
		name, want string
		plant      func(path string) error
	}{
		{"group-writable dir (S6)", "lock directory", func(path string) error {
			return os.Chmod(filepath.Dir(path), 0o770)
		}},
		{"log is a directory", "is a directory", func(path string) error {
			return os.Mkdir(path, 0o700)
		}},
		{"rotation target is a non-empty directory", "rename", func(path string) error {
			if err := os.WriteFile(path, []byte("{}\n"), 0o600); err != nil {
				return err
			}
			return os.MkdirAll(filepath.Join(path+".1", "x"), 0o700)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "offload", "runs.jsonl")
			if err := os.Mkdir(filepath.Dir(path), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := tc.plant(path); err != nil {
				t.Fatal(err)
			}
			before, _ := os.ReadFile(path)
			err := put(dir, path, []byte(strings.Repeat("x", 30)+"\n"))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want one naming %q", err, tc.want)
			}
			if after, _ := os.ReadFile(path); !bytes.Equal(before, after) {
				t.Errorf("log changed after a failure: %q → %q", before, after)
			}
		})
	}
}

// TestJ4WriteErrorSurfaces runs itself under RLIMIT_FSIZE: the record's write fails with EFBIG.
func TestJ4WriteErrorSurfaces(t *testing.T) {
	if dir := os.Getenv("LEDGER_FSIZE_DIR"); dir != "" {
		signal.Ignore(syscall.SIGXFSZ)
		if err := syscall.Setrlimit(syscall.RLIMIT_FSIZE, &syscall.Rlimit{Cur: 8, Max: 8}); err != nil {
			fmt.Print("setrlimit: ", err)
			os.Exit(0)
		}
		fmt.Print(put(dir, filepath.Join(dir, "offload", "runs.jsonl"), []byte(strings.Repeat("x", 64)+"\n")))
		os.Exit(0)
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestJ4WriteErrorSurfaces$")
	cmd.Env = append(os.Environ(), "LEDGER_FSIZE_DIR="+t.TempDir())
	out, err := cmd.Output()
	if err != nil || !strings.Contains(string(out), "file too large") {
		t.Fatalf("child: %q, %v; want the write's EFBIG", out, err)
	}
}

func TestJ4OneFailedWriteNotesOnce(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "offload"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	Begin(env(map[string]string{"XDG_STATE_HOME": dir}), "/w", []string{"x"}, Options{}, &stderr).End(final(0, ""))
	if n := strings.Count(stderr.String(), "offload: run log: "); n != 1 {
		t.Errorf("stderr = %q, want one note for the one failed write", stderr.String())
	}
}

func TestJ4LockHeldElsewhereFailsWithinTimeout(t *testing.T) {
	old := lockTimeout
	t.Cleanup(func() { lockTimeout = old })
	lockTimeout = 300 * time.Millisecond
	dir := t.TempDir()
	path := filepath.Join(dir, "offload", "runs.jsonl")
	if err := put(dir, path, []byte("{}\n")); err != nil {
		t.Fatal(err)
	}
	lock, err := os.OpenFile(filepath.Join(dir, "offload", "runs.lock"), os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	r := Begin(env(map[string]string{"XDG_STATE_HOME": dir}), "/w", []string{"x"}, Options{}, &stderr)
	t0 := time.Now()
	r.Placed(Place{Kind: "local", Slot: 1, Slots: 2})
	r.End(final(0, ""))
	if d := time.Since(t0); d < 2*lockTimeout || d > 2*lockTimeout+time.Second {
		t.Errorf("two writes under a held lock took %v, want ≈ 2 × %v", d, lockTimeout)
	}
	if n := strings.Count(stderr.String(), "offload: run log: "); n != 1 || !strings.Contains(stderr.String(), "still locked") {
		t.Errorf("stderr = %q, want one lock note", stderr.String())
	}
	if got, _ := os.ReadFile(path); string(got) != "{}\n" {
		t.Errorf("log written past a held lock: %q", got)
	}
	_ = lock.Close()
	r.End(final(0, ""))
	if recs := lines(t, path); len(recs) != 2 || recs[1]["event"] != "end" {
		t.Errorf("lock released: records %v, want the seed + an end", recs)
	}
}

func TestJ2EndRecordSecondsAndStatus(t *testing.T) {
	dir := t.TempDir()
	r := Begin(env(map[string]string{"XDG_STATE_HOME": dir}), "/w", []string{"x"}, Options{Get: nil}, &bytes.Buffer{})
	stop := r.Time(Download)
	time.Sleep(20 * time.Millisecond)
	stop()
	r.CommandStatus(7)
	r.End(final(125, "TERM"))
	rec := lines(t, filepath.Join(dir, "offload", "runs.jsonl"))[0]
	if d, _ := rec["download"].(float64); d < 0.02 || d > 1 || rec["upload"] != nil || rec["run"] != nil {
		t.Errorf("phases = upload %v, run %v, download %v", rec["upload"], rec["run"], rec["download"])
	}
	if rec["cmd_rc"] != 7.0 || rec["rc"] != 125.0 || rec["signal"] != "TERM" {
		t.Errorf("cmd_rc %v rc %v signal %v", rec["cmd_rc"], rec["rc"], rec["signal"])
	}
	if msgs, ok := rec["messages"].([]any); !ok || len(msgs) != 0 {
		t.Errorf("messages = %#v, want []", rec["messages"])
	}
	if get, ok := rec["options"].(map[string]any)["get"].([]any); !ok || len(get) != 0 {
		t.Errorf("options.get = %#v, want []", rec["options"])
	}
}

func TestJ2VCSVersion(t *testing.T) {
	set := func(kv ...string) []debug.BuildSetting {
		var out []debug.BuildSetting
		for i := 0; i < len(kv); i += 2 {
			out = append(out, debug.BuildSetting{Key: kv[i], Value: kv[i+1]})
		}
		return out
	}
	for _, tc := range []struct {
		settings []debug.BuildSetting
		want     string // "" = nil
	}{
		{set("vcs", "git", "vcs.time", "x"), ""},
		{set("vcs.modified", "true"), ""},
		{set("vcs.revision", "abc", "vcs.modified", "false"), "abc"},
		{set("vcs.modified", "true", "vcs.revision", "abc"), "abc+dirty"},
		{set("vcs.revision", "abc"), "abc"},
	} {
		got := vcsVersion(tc.settings)
		if (got == nil) != (tc.want == "") || (got != nil && *got != tc.want) {
			t.Errorf("%v → %v, want %q", tc.settings, got, tc.want)
		}
	}
}

func TestJ1LockHeldAroundEveryOpen(t *testing.T) { // J1: the record is written under the flock
	oldMax, oldOpen := maxSize, openLog
	t.Cleanup(func() { maxSize, openLog = oldMax, oldOpen })
	maxSize = 20
	dir := t.TempDir()
	path := filepath.Join(dir, "offload", "runs.jsonl")
	opens := 0
	openLog = func(p string) (*os.File, error) {
		opens++
		probe, err := os.OpenFile(filepath.Join(dir, "offload", "runs.lock"), os.O_RDWR, 0)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = probe.Close() }()
		if err := syscall.Flock(int(probe.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); !errors.Is(err, syscall.EWOULDBLOCK) {
			t.Errorf("open %d: runs.lock free (flock = %v)", opens, err)
		}
		return oldOpen(p)
	}
	for _, line := range []string{`"aaaaaaaaaaaaaaa"`, `"b"`} { // the second rotates: two opens
		if err := put(dir, path, []byte(line+"\n")); err != nil {
			t.Fatal(err)
		}
	}
	if opens != 3 {
		t.Errorf("opens = %d, want 3 (append, rotate's close + reopen)", opens)
	}
}

// TestJ1FailedWriteLeavesNoPartialRecord: under RLIMIT_FSIZE a record's write stops short (also
// right after a rotation); the next record must still start its own line.
func TestJ1FailedWriteLeavesNoPartialRecord(t *testing.T) {
	if dir := os.Getenv("LEDGER_SHORT_DIR"); dir != "" {
		path := filepath.Join(dir, "offload", "runs.jsonl")
		if os.Getenv("LEDGER_SHORT_ROTATE") != "" {
			maxSize = 20 // the payload rotates the seed out first
		}
		signal.Ignore(syscall.SIGXFSZ)
		var limit syscall.Rlimit
		if err := syscall.Getrlimit(syscall.RLIMIT_FSIZE, &limit); err != nil {
			fmt.Print(err)
			os.Exit(0)
		}
		seed := []byte(`{"seed":true}` + "\n")
		short := limit
		short.Cur = uint64(len(seed)) + 8
		fmt.Println(put(dir, path, seed))
		fmt.Println(syscall.Setrlimit(syscall.RLIMIT_FSIZE, &short))
		fmt.Println(put(dir, path, []byte(`{"payload":"`+strings.Repeat("x", 100)+`"}`+"\n")))
		fmt.Println(syscall.Setrlimit(syscall.RLIMIT_FSIZE, &limit))
		fmt.Println(put(dir, path, []byte(`{"next":true}`+"\n")))
		os.Exit(0)
	}
	seed, next := `{"seed":true}`+"\n", `{"next":true}`+"\n"
	for _, tc := range []struct{ rotate, log, rotated string }{
		{"", seed + next, ""},
		{"1", next, seed},
	} {
		dir := t.TempDir()
		cmd := exec.Command(os.Args[0], "-test.run=^TestJ1FailedWriteLeavesNoPartialRecord$")
		cmd.Env = append(os.Environ(), "LEDGER_SHORT_DIR="+dir, "LEDGER_SHORT_ROTATE="+tc.rotate)
		out, err := cmd.Output()
		if want := "<nil>\n<nil>\n"; err != nil || !strings.HasPrefix(string(out), want) || !strings.Contains(string(out), "file too large") {
			t.Fatalf("rotate=%q child: %q, %v; want seed ok, EFBIG on the payload", tc.rotate, out, err)
		}
		path := filepath.Join(dir, "offload", "runs.jsonl")
		got, _ := os.ReadFile(path)
		rotated, _ := os.ReadFile(path + ".1")
		if string(got) != tc.log || string(rotated) != tc.rotated {
			t.Errorf("rotate=%q: log %q, .1 %q; want %q, %q (no partial payload)", tc.rotate, got, rotated, tc.log, tc.rotated)
		}
	}
}

func TestJ4EndReturnsFinalStatusWhenTheWriteFails(t *testing.T) {
	var stderr bytes.Buffer
	r := Begin(env(map[string]string{"HOME": "rel"}), "/w", []string{"x"}, Options{}, &stderr)
	if rc := r.End(final(143, "TERM")); rc != 143 {
		t.Fatalf("End = %d, want final's 143", rc)
	}
	if n := strings.Count(stderr.String(), "offload: run log: "); n != 1 {
		t.Errorf("stderr = %q, want one note", stderr.String())
	}
}
