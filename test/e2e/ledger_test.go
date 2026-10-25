package e2e_test

import (
	"debug/buildinfo"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

var (
	startKeys = []string{"argv", "cwd", "event", "id", "options", "pid", "place", "root", "start", "v", "version", "waited"}
	endKeys   = slices.Sorted(slices.Values(append(slices.Clone(startKeys),
		"cmd_rc", "download", "end", "messages", "rc", "run", "signal", "upload")))
)

// wantVersion = J2's `version` of the binary under test, from its own build info (oracle).
func wantVersion(t *testing.T) any {
	t.Helper()
	info, err := buildinfo.ReadFile(binaryPath)
	must(t, err)
	var rev, dirty string
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.modified":
			if s.Value == "true" {
				dirty = "+dirty"
			}
		}
	}
	if rev == "" {
		return nil
	}
	return rev + dirty
}

// records = the sandbox's run log (J1): one JSON object per `\n`-terminated line.
func (s *sandbox) records(dir string) []map[string]any {
	s.t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "offload", "runs.jsonl"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	must(s.t, err)
	var recs []map[string]any
	for i, line := range strings.SplitAfter(string(data), "\n") {
		if line == "" {
			continue
		}
		if !strings.HasSuffix(line, "\n") {
			s.t.Fatalf("record %d unterminated: %q", i, line)
		}
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			s.t.Fatalf("record %d: %v: %q", i, err, line)
		}
		recs = append(recs, rec)
	}
	return recs
}

func (s *sandbox) stateDir() string { return filepath.Join(s.base, "state") }

func requireKeys(t *testing.T, rec map[string]any, want []string) {
	t.Helper()
	if got := slices.Sorted(maps.Keys(rec)); !slices.Equal(got, want) {
		t.Fatalf("%v record keys = %v, want %v", rec["event"], got, want)
	}
}

func requireField(t *testing.T, rec map[string]any, key string, want any) {
	t.Helper()
	if got := rec[key]; !reflect.DeepEqual(got, want) {
		t.Errorf("%v.%s = %#v, want %#v", rec["event"], key, got, want)
	}
}

func requireSeconds(t *testing.T, rec map[string]any, keys ...string) {
	t.Helper()
	for _, key := range keys {
		if v, ok := rec[key].(float64); !ok || v < 0 {
			t.Errorf("%v.%s = %#v, want seconds ≥ 0", rec["event"], key, rec[key])
		}
	}
}

func requireTime(t *testing.T, rec map[string]any, key string, after time.Time) time.Time {
	t.Helper()
	text, _ := rec[key].(string)
	at, err := time.Parse(time.RFC3339Nano, text)
	if err != nil || !strings.HasSuffix(text, "Z") || at.Before(after.Add(-time.Second)) || at.After(time.Now().Add(time.Second)) {
		t.Errorf("%v.%s = %q (err %v), want RFC 3339 UTC within the test", rec["event"], key, text, err)
	}
	return at
}

// runPair = the run's start + end records (J2): shared id, pid + every start field.
func (s *sandbox) runPair(since time.Time) (start, end map[string]any) {
	s.t.Helper()
	recs := s.records(s.stateDir())
	if len(recs) != 2 {
		s.t.Fatalf("records = %d, want start + end: %v", len(recs), recs)
	}
	start, end = recs[0], recs[1]
	requireKeys(s.t, start, startKeys)
	requireKeys(s.t, end, endKeys)
	requireField(s.t, start, "event", "start")
	requireField(s.t, end, "event", "end")
	for _, key := range startKeys {
		if key != "event" {
			requireField(s.t, end, key, start[key])
		}
	}
	requireField(s.t, start, "v", 1.0)
	if id, _ := start["id"].(string); id == "" {
		s.t.Errorf("id = %#v", start["id"])
	}
	if pid, _ := start["pid"].(float64); pid <= 1 {
		s.t.Errorf("pid = %#v", start["pid"])
	}
	requireField(s.t, start, "version", wantVersion(s.t))
	begun := requireTime(s.t, start, "start", since)
	if ended := requireTime(s.t, end, "end", since); ended.Before(begun) {
		s.t.Errorf("end %v before start %v", ended, begun)
	}
	requireSeconds(s.t, start, "waited")
	return start, end
}

func stderrLines(stderr string) []any {
	var out []any
	for line := range strings.Lines(stderr) {
		out = append(out, strings.TrimSuffix(line, "\n"))
	}
	return out
}

func noOptions() map[string]any {
	return map[string]any{"local": false, "pool": false, "host": "", "gpu": false, "no_sync": false, "get": []any{}}
}

func TestJ2LocalRunRecords(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, false)
	since := time.Now()
	r := s.run("run", "--local", "--", "sh", "-c", "echo out; echo err >&2; exit 3")
	r.requireCode(t, 3)
	start, end := s.runPair(since)
	requireField(t, start, "argv", []any{"sh", "-c", "echo out; echo err >&2; exit 3"})
	requireField(t, start, "cwd", s.root)
	requireField(t, start, "root", nil)
	opts := noOptions()
	opts["local"] = true
	requireField(t, start, "options", opts)
	requireField(t, start, "place", map[string]any{"kind": "local", "slot": 1.0, "slots": 2.0})
	requireSeconds(t, end, "run")
	requireField(t, end, "upload", nil)
	requireField(t, end, "download", nil)
	requireField(t, end, "cmd_rc", 3.0)
	requireField(t, end, "rc", 3.0)
	requireField(t, end, "signal", nil)
	requireField(t, end, "messages", []any{"offload: local slot 1/2"}) // the command's "err" never
}

func TestJ2PoolRunRecords(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, true)
	since := time.Now()
	r := s.run("run", "--get", "tracked", "--", "bash", "-c", "exit 0")
	r.requireCode(t, 0)
	start, end := s.runPair(since)
	requireField(t, start, "root", s.root)
	opts := noOptions()
	opts["pool"], opts["get"] = true, []any{"tracked"}
	requireField(t, start, "options", opts)
	requireField(t, start, "place", map[string]any{
		"kind": "pool", "slot": 1.0, "slots": 2.0, "host": "alpha", "cpus": 8.0, "busy": 25.0,
		"workdir": "offload/" + filepath.Base(s.workdir("alpha")),
	})
	requireSeconds(t, end, "upload", "run", "download")
	requireField(t, end, "cmd_rc", 0.0)
	requireField(t, end, "rc", 0.0)
	msgs, _ := end["messages"].([]any)
	if lines := stderrLines(r.stderr); len(msgs) == 0 || msgs[0] != lines[0] || !strings.HasPrefix(lines[0].(string), "offload: alpha slot 1/2 ") {
		t.Errorf("messages = %#v, stderr = %q", msgs, r.stderr)
	}
}

func TestJ2SyncFailures(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		marker              string
		cmdRC               any
		upload, run, dnload bool
	}{
		{".upload-fail", nil, true, false, false},
		{".download-fail", 2.0, true, true, true},
	} {
		t.Run(tc.marker, func(t *testing.T) {
			t.Parallel()
			s := newSandbox(t, true)
			s.write(filepath.Join(s.hosts, "alpha", tc.marker), "fail\n")
			r := s.run("run", "--pool", "--", "bash", "-c", "exit 2")
			r.requireCode(t, 125)
			_, end := s.runPair(time.Time{})
			for key, ran := range map[string]bool{"upload": tc.upload, "run": tc.run, "download": tc.dnload} {
				if ran {
					requireSeconds(t, end, key)
				} else {
					requireField(t, end, key, nil)
				}
			}
			requireField(t, end, "cmd_rc", tc.cmdRC)
			requireField(t, end, "rc", 125.0)
			msgs, _ := end["messages"].([]any)
			last, _ := msgs[len(msgs)-1].(string)
			if len(msgs) < 2 || !strings.HasPrefix(last, "offload: ") || !strings.Contains(r.stderr, last+"\n") {
				t.Errorf("messages = %#v, stderr = %q", msgs, r.stderr)
			}
		})
	}
}

func TestJ2SignalRecorded(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		args   []string
		placed string
		sig    syscall.Signal
		rc     float64
	}{
		{"TERM", []string{"--local"}, "local slot", syscall.SIGTERM, 143},
		{"INT", []string{"--pool"}, "alpha slot", syscall.SIGINT, 130},
		{"HUP", []string{"--pool"}, "alpha slot", syscall.SIGHUP, 129},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := newSandbox(t, true)
			p := s.start(append(append([]string{"run"}, tc.args...), "--", "sleep", "30")...)
			p.await("placement", func() bool { return strings.Contains(p.output().stderr, tc.placed) })
			p.await("start record", func() bool { return len(s.records(s.stateDir())) == 1 })
			p.signal(tc.sig)
			p.wait(20*time.Second).requireCode(t, int(tc.rc))
			_, end := s.runPair(time.Time{})
			requireField(t, end, "signal", tc.name)
			requireField(t, end, "rc", tc.rc)
		})
	}
}

func TestJ2UnplacedRunWritesEndAlone(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, true)
	s.write(s.config, "[[host]]\nname = 1\n")
	r := s.run("run", "--", "true")
	r.requireCode(t, 2)
	recs := s.records(s.stateDir())
	if len(recs) != 1 {
		t.Fatalf("records = %v, want one end record", recs)
	}
	end := recs[0]
	requireKeys(t, end, endKeys)
	requireField(t, end, "event", "end")
	for _, key := range []string{"place", "waited", "upload", "run", "download", "cmd_rc", "signal", "root"} {
		requireField(t, end, key, nil)
	}
	requireField(t, end, "rc", 2.0)
	requireField(t, end, "messages", stderrLines(r.stderr))
}

func TestJ1UsageErrorWritesNothing(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, false)
	s.run("run", "--local", "--pool", "--", "true").requireCode(t, 2)
	s.run("run").requireCode(t, 2)
	if recs := s.records(s.stateDir()); recs != nil {
		t.Errorf("usage errors logged: %v", recs)
	}
}

func TestJ5KilledOffloadLeavesStartAlone(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, false)
	p := s.start("run", "--local", "--", "sleep", "30")
	p.await("start record", func() bool { return len(s.records(s.stateDir())) == 1 })
	p.signal(syscall.SIGKILL)
	p.wait(10 * time.Second)
	recs := s.records(s.stateDir())
	if len(recs) != 1 {
		t.Fatalf("records = %v, want the start alone", recs)
	}
	requireKeys(t, recs[0], startKeys)
	requireField(t, recs[0], "event", "start")
}

func TestJ1DefaultPathAndModes(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, false)
	s.setenv("XDG_STATE_HOME", "")
	s.run("run", "--local", "--", "true").requireCode(t, 0)
	state := filepath.Join(s.base, "home", ".local", "state")
	if n := len(s.records(state)); n != 2 {
		t.Fatalf("records under ~/.local/state = %d, want 2", n)
	}
	assertAbsent(t, s.stateDir())
	for path, want := range map[string]fs.FileMode{
		filepath.Join(s.base, "home", ".local"): fs.ModeDir | 0o700, state: fs.ModeDir | 0o700,
		filepath.Join(state, "offload"): fs.ModeDir | 0o700, filepath.Join(state, "offload", "runs.jsonl"): 0o600,
		filepath.Join(state, "offload", "runs.lock"): 0o600,
	} {
		info, err := os.Lstat(path)
		if err != nil || info.Mode() != want {
			t.Errorf("%s: mode %v (err %v), want %v", path, info.Mode(), err, want)
		}
	}
}

func TestJ4LogFailureLeavesRunUnchanged(t *testing.T) {
	t.Parallel()
	for _, args := range [][]string{{"--local"}, {"--pool"}} {
		t.Run(args[0], func(t *testing.T) {
			t.Parallel()
			s := newSandbox(t, true)
			s.write(filepath.Join(s.stateDir(), "offload"), "a file where the log directory goes\n")
			r := s.run(append(append([]string{"run"}, args...), "--", "sh", "-c", "printf out; exit 4")...)
			r.requireCode(t, 4)
			if r.stdout != "out" {
				t.Errorf("stdout = %q", r.stdout)
			}
			if n := strings.Count(r.stderr, "offload: run log: "); n != 1 {
				t.Errorf("run log notes = %d, want 1 for both failed writes; stderr = %q", n, r.stderr)
			}
			placed := map[string]string{"--local": "offload: local slot 1/2\n", "--pool": "offload: alpha slot 1/2 "}[args[0]]
			requireContains(t, r.stderr, placed)
		})
	}
}

func TestJ2NoSyncTimesTheCommandAlone(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, true)
	s.run("run", "--no-sync", "--", "true").requireCode(t, 0)
	start, end := s.runPair(time.Time{})
	opts := noOptions()
	opts["pool"], opts["no_sync"] = true, true
	requireField(t, start, "options", opts)
	requireSeconds(t, end, "run")
	requireField(t, end, "upload", nil)
	requireField(t, end, "download", nil)
	requireField(t, end, "cmd_rc", 0.0)
}

func TestJ2BusyMatchesPlacementLine(t *testing.T) { // J2 busy = C5's printed %
	t.Parallel()
	for load, want := range map[float64]float64{1.0: 12, 1.004: 13} { // 12.5 % ties to even; 12.55 % rounds up
		s := newSandbox(t, true, host{name: "alpha", cpus: 8, load: load, slots: 2})
		r := s.run("run", "--pool", "--", "true")
		r.requireCode(t, 0)
		requireContains(t, r.stderr, "(8 CPUs, "+strconv.Itoa(int(want))+"% busy)")
		start, _ := s.runPair(time.Time{})
		if place, _ := start["place"].(map[string]any); place["busy"] != want {
			t.Errorf("load %v: place = %#v, want busy %v", load, place, want)
		}
	}
}

func TestJ1NoAbsoluteStateDirLogsNothing(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, false)
	s.setenv("XDG_STATE_HOME", "state")
	s.setenv("HOME", "home")
	r := s.run("run", "--local", "--", "true")
	r.requireCode(t, 0)
	if n := strings.Count(r.stderr, "offload: run log: "); n != 1 || !strings.Contains(r.stderr, "absolute path") {
		t.Errorf("stderr = %q, want one note naming the missing absolute path", r.stderr)
	}
	assertAbsent(t, filepath.Join(s.cwd, "offload"), filepath.Join(s.cwd, "state"), filepath.Join(s.cwd, "home"))
}

func TestJ4FIFOLogLeavesCommandRunnable(t *testing.T) {
	t.Parallel()
	for _, reader := range []bool{false, true} { // without one, open fails; with one, only the type check refuses
		t.Run(fmt.Sprint("reader=", reader), func(t *testing.T) {
			t.Parallel()
			s := newSandbox(t, false)
			fifo := filepath.Join(s.stateDir(), "offload", "runs.jsonl")
			must(t, os.MkdirAll(filepath.Dir(fifo), 0o700))
			must(t, syscall.Mkfifo(fifo, 0o600))
			if reader {
				r, err := os.OpenFile(fifo, os.O_RDONLY|syscall.O_NONBLOCK, 0)
				must(t, err)
				t.Cleanup(func() { _ = r.Close() })
			}
			r := s.start("run", "--local", "--", "sh", "-c", "printf out; exit 4").wait(5 * time.Second)
			r.requireCode(t, 4)
			if r.stdout != "out" || strings.Count(r.stderr, "offload: run log: ") != 1 {
				t.Errorf("stdout = %q, stderr = %q; want the command's output + one note", r.stdout, r.stderr)
			}
			if reader {
				requireContains(t, r.stderr, "runs.jsonl: not a regular file\n")
			}
		})
	}
}

func TestJ2SignalDuringEndWriteCounts(t *testing.T) { // C3 holds while the end record waits for its lock
	t.Parallel()
	s := newSandbox(t, false)
	release := filepath.Join(s.base, "release-command")
	p := s.start("run", "--local", "--", "bash", "-c", `printf ready; until [[ -e $1 ]]; do sleep 0.01; done; exit 7`, "_", release)
	p.await("command ready", func() bool { return p.output().stdout == "ready" })
	lock, err := os.OpenFile(filepath.Join(s.stateDir(), "offload", "runs.lock"), os.O_WRONLY, 0)
	must(t, err)
	must(t, syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB))
	s.write(release, "go\n")
	p.await("slot released (end write waiting)", func() bool {
		owner, err := os.ReadFile(filepath.Join(s.runtime, "heavy", "slot1.owner"))
		return err == nil && len(owner) == 0
	})
	p.signal(syscall.SIGTERM)
	time.Sleep(300 * time.Millisecond)
	must(t, lock.Close())
	p.wait(5*time.Second).requireCode(t, 143)
	_, end := s.runPair(time.Time{})
	requireField(t, end, "cmd_rc", 7.0)
	requireField(t, end, "rc", 143.0)
	requireField(t, end, "signal", "TERM")
}

func TestJ1ConcurrentRunsKeepEveryRecord(t *testing.T) {
	t.Parallel()
	s := newSandbox(t, false)
	const count = 12
	s.setenv("HEAVY_SLOTS", strconv.Itoa(count))
	var runs []*running
	for range count {
		runs = append(runs, s.start("run", "--local", "--", "true"))
	}
	for _, p := range runs {
		p.wait(10*time.Second).requireCode(t, 0)
	}
	recs := s.records(s.stateDir())
	events := map[string]string{}
	for _, rec := range recs {
		id, _ := rec["id"].(string)
		events[id] += rec["event"].(string) + ","
	}
	if len(recs) != 2*count || len(events) != count {
		t.Fatalf("records = %d over %d ids, want %d over %d", len(recs), len(events), 2*count, count)
	}
	for id, got := range events {
		if got != "start,end," {
			t.Errorf("id %s: events %s", id, got)
		}
	}
}

func TestJ2RemotePhasesPerMode(t *testing.T) { // + X6: a command's own 255 is its status
	t.Parallel()
	for _, mode := range []string{"plain", "git", "linked", "nosync", "nosync-get"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			s := newSandbox(t, mode != "plain")
			s.setenv("GIT_CEILING_DIRECTORIES", s.base)
			args := []string{"run", "--pool"}
			if mode == "linked" {
				linked := filepath.Join(s.base, "linked")
				s.git("worktree", "add", "-qb", "linked", linked)
				s.root, s.cwd = linked, linked
			}
			if strings.HasPrefix(mode, "nosync") {
				args = append(args, "--no-sync")
			}
			if mode == "nosync-get" {
				args = append(args, "--get", "fetch")
			}
			s.run(append(args, "--", "bash", "-c", "printf fetch > fetch; exit 255")...).requireCode(t, 255)
			_, end := s.runPair(time.Time{})
			requireField(t, end, "cmd_rc", 255.0)
			requireField(t, end, "rc", 255.0)
			requireSeconds(t, end, "run")
			if strings.HasPrefix(mode, "nosync") {
				requireField(t, end, "upload", nil)
			} else {
				requireSeconds(t, end, "upload")
			}
			if mode == "nosync" {
				requireField(t, end, "download", nil)
			} else {
				requireSeconds(t, end, "download")
			}
			if mode == "nosync-get" && s.read("fetch") != "fetch" {
				t.Error("--get output missing")
			}
		})
	}
}
