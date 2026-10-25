// Package ledger appends offload's run log: a start record once a run holds its slot, an end
// record at exit (contract J).
package ledger

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"syscall"
	"time"

	"github.com/eturkes/offload/internal/slots"
)

var (
	maxSize     int64 = 64 << 20 // J3
	lockTimeout       = 2 * time.Second
)

// Options = the run's flags after C2's implications.
type Options struct {
	Local  bool     `json:"local"`
	Pool   bool     `json:"pool"`
	Host   string   `json:"host"`
	GPU    bool     `json:"gpu"`
	NoSync bool     `json:"no_sync"`
	Get    []string `json:"get"`
}

// Place = the slot a run holds; Host…Workdir only for Kind "pool".
type Place struct {
	Kind    string `json:"kind"`
	Slot    int    `json:"slot"`
	Slots   int    `json:"slots"`
	Host    string `json:"host,omitempty"`
	CPUs    int    `json:"cpus,omitempty"`
	Busy    *int   `json:"busy,omitempty"`
	Workdir string `json:"workdir,omitempty"`
}

type head struct {
	V       int      `json:"v"`
	ID      string   `json:"id"`
	Event   string   `json:"event"`
	Version *string  `json:"version"`
	PID     int      `json:"pid"`
	Start   string   `json:"start"`
	Cwd     string   `json:"cwd"`
	Argv    []string `json:"argv"`
	Options Options  `json:"options"`
	Root    *string  `json:"root"`
	Place   *Place   `json:"place"`
	Waited  *float64 `json:"waited"`
}

type tail struct {
	head
	End      string   `json:"end"`
	Upload   *float64 `json:"upload"`
	Run      *float64 `json:"run"`
	Download *float64 `json:"download"`
	CmdRC    *int     `json:"cmd_rc"`
	RC       int      `json:"rc"`
	Signal   *string  `json:"signal"`
	Messages []string `json:"messages"`
}

// Phase = a timed part of a run.
type Phase int

// Phases → the end record's `upload`, `run`, `download` (J2).
const (
	Upload Phase = iota
	Command
	Download
)

// Run = one `offload run`'s log, used from one goroutine. Its Writer passes offload's stderr
// through and keeps each line.
type Run struct {
	base, path string
	pathErr    error
	t0         time.Time
	stderr     io.Writer
	rec        tail
	partial    []byte
	noted      bool
}

// Begin starts the log of a run that passed usage parsing (J1).
func Begin(getenv func(string) string, cwd string, argv []string, opts Options, stderr io.Writer) *Run {
	r := &Run{t0: time.Now(), stderr: stderr}
	r.base, r.pathErr = stateBase(getenv)
	r.path = filepath.Join(r.base, "offload", "runs.jsonl")
	if opts.Get == nil {
		opts.Get = []string{}
	}
	r.rec = tail{
		head: head{
			V: 1, ID: rand.Text(), Version: version(), PID: os.Getpid(), Start: stamp(r.t0),
			Cwd: cwd, Argv: argv, Options: opts,
		},
		Messages: []string{},
	}
	return r
}

// stateBase = $XDG_STATE_HOME, else ~/.local/state; empty or relative values count as unset (XDG).
func stateBase(getenv func(string) string) (string, error) {
	if d := getenv("XDG_STATE_HOME"); filepath.IsAbs(d) {
		return d, nil
	}
	if home := getenv("HOME"); filepath.IsAbs(home) {
		return filepath.Join(home, ".local", "state"), nil
	}
	return "", errors.New("neither XDG_STATE_HOME nor HOME is an absolute path")
}

var version = sync.OnceValue(func() *string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return nil
	}
	return vcsVersion(info.Settings)
})

// vcsVersion = J2's `version`: the VCS revision, `+dirty` for a modified tree; nil without one.
func vcsVersion(settings []debug.BuildSetting) *string {
	var rev, dirty string
	for _, s := range settings {
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
	rev += dirty
	return &rev
}

func stamp(t time.Time) string { return t.UTC().Format("2006-01-02T15:04:05.000Z07:00") }

func seconds(d time.Duration) *float64 {
	s := math.Round(d.Seconds()*1000) / 1000
	return &s
}

// Writer = stderr unchanged (C4, X4), each complete line also kept for the end record.
func (r *Run) Writer() io.Writer { return tee{r} }

type tee struct{ r *Run }

func (t tee) Write(p []byte) (int, error) {
	n, err := t.r.stderr.Write(p)
	t.r.partial = append(t.r.partial, p[:n]...)
	for {
		i := bytes.IndexByte(t.r.partial, '\n')
		if i < 0 {
			break
		}
		t.r.rec.Messages = append(t.r.rec.Messages, string(t.r.partial[:i]))
		t.r.partial = t.r.partial[i+1:]
	}
	return n, err
}

// Root records the sync root the run inspected.
func (r *Run) Root(root string) { r.rec.Root = &root }

// Placed records the held slot and appends the start record.
func (r *Run) Placed(p Place) {
	r.rec.Place, r.rec.Waited = &p, seconds(time.Since(r.t0))
	h := r.rec.head
	h.Event = "start"
	r.write(func() any { return h })
}

// Time starts timing phase; the returned func ends it.
func (r *Run) Time(phase Phase) func() {
	t := time.Now()
	return func() {
		d := seconds(time.Since(t))
		switch phase {
		case Upload:
			r.rec.Upload = d
		case Command:
			r.rec.Run = d
		case Download:
			r.rec.Download = d
		}
	}
}

// CommandStatus records the status the command itself exited with.
func (r *Run) CommandStatus(rc int) { r.rec.CmdRC = &rc }

// End appends the end record and returns offload's exit status. final (rc; signal = "" unless HUP,
// INT or TERM ended offload) builds the record once its lock is held and gives the status after the
// write, so a signal arriving during the wait counts in both (C3).
func (r *Run) End(final func() (rc int, signal string)) int {
	if len(r.partial) > 0 {
		r.rec.Messages, r.partial = append(r.rec.Messages, string(r.partial)), nil
	}
	r.write(func() any {
		rc, sig := final()
		rec := r.rec
		rec.Event, rec.End, rec.RC = "end", stamp(time.Now()), rc
		if sig != "" {
			rec.Signal = &sig
		}
		return rec
	})
	rc, _ := final()
	return rc
}

// write appends the record build returns, called once the log's lock is held. A failure leaves the
// run alone and says so once per run (J4).
func (r *Run) write(build func() any) {
	err := r.pathErr
	if err == nil {
		err = appendLine(r.base, r.path, func() ([]byte, error) {
			line, merr := json.Marshal(build())
			return append(line, '\n'), merr
		})
	}
	if err != nil && !r.noted {
		r.noted = true
		fmt.Fprintf(r.Writer(), "offload: run log: %v\n", err)
	}
}

// appendLine writes record()'s line whole at the end of path under runs.lock, rotating past maxSize
// (J1, J3); a failed write is cut back off, since a partial line would prefix the next record.
func appendLine(base, path string, record func() ([]byte, error)) error {
	ctx, cancel := context.WithTimeout(context.Background(), lockTimeout)
	defer cancel()
	lockPath := filepath.Join(filepath.Dir(path), "runs.lock")
	lock, err := slots.WaitLock(ctx, base, lockPath)
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%s: still locked after %s", lockPath, lockTimeout)
	}
	if err != nil {
		return err
	}
	defer func() { _ = lock.Close() }()
	line, err := record()
	if err != nil {
		return err
	}
	f, err := openLog(path)
	if err != nil {
		return err
	}
	var size int64
	info, err := f.Stat()
	if err == nil {
		size = info.Size()
	}
	if err == nil && size > 0 && size+int64(len(line)) > maxSize {
		_ = f.Close()
		f, size = nil, 0
		if err = os.Rename(path, path+".1"); err == nil { //nolint:gosec // G703: the state dir's own log path
			f, err = openLog(path)
		}
	}
	if err != nil {
		if f != nil {
			_ = f.Close()
		}
		return err
	}
	if _, err = f.Write(line); err != nil {
		_ = f.Truncate(size)
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	return err
}

// openLog opens the log for appending, a seam for tests that check the lock is held around it.
// O_NONBLOCK + the regular-file check: a FIFO at the path fails at once instead of blocking the run
// (J4), with or without a reader.
var openLog = func(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0o600) //nolint:gosec // G304: the state dir's own log path
	if err != nil {
		return nil, err
	}
	if info, err := f.Stat(); err != nil || !info.Mode().IsRegular() {
		_ = f.Close()
		if err == nil {
			err = fmt.Errorf("%s: not a regular file", path)
		}
		return nil, err
	}
	return f, nil
}
