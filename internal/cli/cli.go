// Package cli parses offload's command line, places a run (a pool host's slot, else a local heavy
// slot, else wait) and reports `status`.
package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/eturkes/offload/internal/ctr"
	"github.com/eturkes/offload/internal/mirror"
	"github.com/eturkes/offload/internal/pool"
	"github.com/eturkes/offload/internal/remote"
	"github.com/eturkes/offload/internal/run"
	"github.com/eturkes/offload/internal/slots"
)

// Exit statuses offload owns; every other status is the command's (C3).
const (
	ExitUsage   = 2
	ExitFailure = 125
)

const usage = `Usage:
  offload run [--local | --pool] [--host NAME] [--gpu] [--no-sync] [--get PATH]... [--] COMMAND [ARG...]
  offload ctr [--] COMMAND [ARG...]
  offload ctr --layers | --check
  offload status

run      Run COMMAND on the least-loaded pool host with a free slot, else in a free
         local heavy slot. When no slot is free, wait for one.
         Files that git ignores stay on each side, and each host keeps its copies
         between runs. Set up ignored tools in the same command, for example:
         offload run -- sh -c 'uv sync && uv run pytest'
ctr      Run COMMAND in the image of the project, which .devcontainer/Containerfile
         defines. This happens only on machines that set OFFLOAD_CTR=1. On other
         machines, COMMAND runs directly. Each machine builds the image when it
         first needs it and keeps it. To run in the image on a pool host:
         offload run -- offload ctr -- make test
status   Show the local heavy slots and each pool host's slots and load.

Run options:
  --local      Use the local heavy slots only.
  --pool       Use the pool hosts only.
  --host NAME  Use this pool host only.
  --gpu        Use pool hosts with an NVIDIA GPU only.
  --no-sync    Do not copy files to or from the host.
  --get PATH   After the run, also copy PATH back, even if git ignores it. Repeat for more paths.

Ctr options:
  --layers     Build the image if necessary, then show its layer digests.
  --check      Compare the image layers and the offload binary of this machine
               with each pool host. Waits for a free slot on each host.

The pool file is $OFFLOAD_CONFIG, else ~/.config/offload/pool.toml.
`

type options struct {
	local, pool, gpu, noSync bool
	host                     string
	gets                     []string
	argv                     []string
}

// Main runs offload with args (without the program name) and returns the exit status.
func Main(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return ExitUsage
	}
	switch args[0] {
	case "-h", "--help", "help":
		fmt.Fprint(stdout, usage)
		return 0
	case "status":
		if len(args) > 1 {
			if args[1] == "-h" || args[1] == "--help" {
				fmt.Fprint(stdout, usage)
				return 0
			}
			return usageError(stderr, "status takes no arguments")
		}
		return status(stdout, stderr)
	case "run":
		o, err := parseRun(args[1:])
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprint(stdout, usage)
			return 0
		}
		if err != nil {
			return usageError(stderr, err.Error())
		}
		return runCommand(o, stderr)
	case "ctr":
		return ctrCommand(args[1:], stdout, stderr)
	case "__ctr-init":
		return ctr.Init(args[1:], stderr)
	}
	return usageError(stderr, fmt.Sprintf("unknown command %q", args[0]))
}

func usageError(stderr io.Writer, msg string) int {
	fmt.Fprintf(stderr, "offload: %s\n\n%s", msg, usage)
	return ExitUsage
}

func parseRun(args []string) (options, error) {
	var o options
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.BoolVar(&o.local, "local", false, "")
	fs.BoolVar(&o.pool, "pool", false, "")
	fs.BoolVar(&o.gpu, "gpu", false, "")
	fs.BoolVar(&o.noSync, "no-sync", false, "")
	fs.StringVar(&o.host, "host", "", "")
	fs.Func("get", "", func(s string) error { o.gets = append(o.gets, s); return nil })
	if err := fs.Parse(args); err != nil {
		return o, err
	}
	o.argv = fs.Args()
	if len(o.argv) == 0 {
		return o, errors.New("give a command after `run`, e.g. `offload run -- make -j8`")
	}
	poolOnly := o.host != "" || o.gpu || o.noSync || len(o.gets) > 0
	if o.local && (o.pool || poolOnly) {
		return o, errors.New("--local excludes --pool, --host, --gpu, --no-sync and --get")
	}
	o.pool = o.pool || poolOnly
	return o, nil
}

// signals cancels the returned context on the first of sigs (default HUP/INT/TERM) and reports
// which one arrived.
func signals(sigs ...os.Signal) (context.Context, func() int) {
	if len(sigs) == 0 {
		sigs = []os.Signal{syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM}
	}
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, sigs...)
	var got syscall.Signal
	go func() {
		s := <-ch
		got, _ = s.(syscall.Signal)
		cancel()
	}()
	return ctx, func() int { return 128 + int(got) }
}

func runCommand(o options, stderr io.Writer) int {
	ctx, signalStatus := signals()
	// C3: a signal outranks the failure it caused (a killed git or rsync).
	fail := func(err error) int {
		if ctx.Err() != nil {
			return signalStatus()
		}
		fmt.Fprintf(stderr, "offload: %v\n", err)
		return ExitFailure
	}
	cwd, err := os.Getwd()
	if err != nil {
		return fail(err)
	}

	var hosts []pool.Host
	poolPath := pool.Path(os.Getenv)
	if !o.local {
		hosts, err = pool.Load(poolPath)
		switch {
		case errors.Is(err, pool.ErrNotExist) && !o.pool:
			fmt.Fprintf(stderr, "offload: %v; local slots only\n", err)
		case err != nil:
			fmt.Fprintf(stderr, "offload: %v\n", err)
			return ExitUsage
		}
	}
	if o.host != "" {
		hosts = filterHost(hosts, o.host)
		if len(hosts) == 0 {
			fmt.Fprintf(stderr, "offload: no host named %s in %s\n", o.host, poolPath)
			return ExitUsage
		}
	}

	t := mirror.Tree{Root: cwd, Sub: "."}
	if len(hosts) > 0 {
		if t, err = mirror.Inspect(ctx, cwd); err != nil {
			return fail(err)
		}
		if t.Git && t.LocalOnly == "" { // nested repos surface in the file list (Y3, before remote work)
			if _, err = mirror.UploadList(ctx, t.Root); errors.Is(err, mirror.ErrLocalOnly) {
				t.LocalOnly = err.Error()
			} else if err != nil {
				return fail(err)
			}
		}
		if t.LocalOnly != "" {
			if o.pool {
				fmt.Fprintf(stderr, "offload: %s runs locally only: %s\n", t.Root, t.LocalOnly)
				return ExitFailure
			}
			hosts = nil
		}
	}
	gets := make([]string, len(o.gets))
	for i, g := range o.gets {
		if gets[i], err = t.GetPath(g); err != nil {
			fmt.Fprintf(stderr, "offload: %v\n", err)
			return ExitUsage
		}
	}
	if !o.pool {
		if rc, ok := checkLocalCommand(o.argv[0], stderr); !ok {
			return rc
		}
	}

	heavy, err := slots.Heavy(os.Getenv)
	if err != nil {
		fmt.Fprintf(stderr, "offload: %v\n", err)
		return ExitUsage
	}
	p := placer{opts: o, heavy: heavy, hosts: hosts, targets: targets(hosts, stderr), tree: t, stderr: stderr}
	pl, err := p.place(ctx)
	switch {
	case ctx.Err() != nil:
		return signalStatus()
	case err != nil:
		return fail(err)
	}
	defer pl.release()

	if pl.local != nil {
		fmt.Fprintf(stderr, "offload: local slot %d/%d\n", pl.local.Index, pl.local.Total)
		rc, runErr := run.Local(ctx, o.argv, pl.local.File())
		if ctx.Err() != nil {
			return signalStatus()
		}
		if runErr != nil {
			return fail(runErr)
		}
		return rc
	}
	fmt.Fprintf(stderr, "offload: %s slot %d/%d (%d CPUs, %.0f%% busy) → ~/%s\n", pl.host.Name,
		pl.slot.Index, pl.slot.Total, pl.stats.CPUs, 100*pl.stats.Busy(), t.Workdir())
	rc, err := remoteRun(ctx, o, t, p.targets[pl.host.Name], gets, stderr)
	if ctx.Err() != nil {
		return signalStatus()
	}
	if err != nil {
		return fail(err)
	}
	return rc
}

// remoteRun = upload → run → download → --get (Y4–Y10).
func remoteRun(ctx context.Context, o options, t mirror.Tree, target remote.Target, gets []string, stderr io.Writer) (int, error) {
	r := run.Remote{Target: target, Workdir: t.Workdir(), Meta: t.Meta(), Sub: t.Sub, Argv: o.argv, Mode: "nosync"}
	switch {
	case o.noSync:
	case t.Git:
		r.Mode = "git"
		if t.Worktree != "" {
			r.Mode = "worktree:" + t.Worktree
		}
		list, err := mirror.UploadList(ctx, t.Root)
		if err != nil {
			return 0, err
		}
		if r.Excludes, err = mirror.GlobalExcludes(ctx, t.Root); err != nil {
			return 0, err
		}
		if err := mirror.UploadGit(ctx, t, target, list); err != nil {
			return 0, err
		}
		r.Paths = list
	default:
		r.Mode = "plain"
		if err := mirror.UploadPlain(ctx, t, target); err != nil {
			return 0, err
		}
	}
	rc, err := r.Run(ctx)
	if err != nil || ctx.Err() != nil {
		return rc, err
	}
	switch {
	case t.Git && !o.noSync:
		err = mirror.DownloadGit(ctx, t, target)
	case r.Mode == "plain":
		err = mirror.DownloadPlain(ctx, t, target)
	}
	for _, g := range gets {
		if err != nil {
			break
		}
		err = mirror.Get(ctx, t, target, g)
	}
	if err != nil {
		fmt.Fprintf(stderr, "offload: the command exited %d, but its outputs did not all come back\n", rc)
	}
	return rc, err
}

// checkLocalCommand mirrors the shell's 127/126 before any slot is taken (X2).
func checkLocalCommand(name string, stderr io.Writer) (int, bool) {
	path, err := exec.LookPath(name)
	if err == nil {
		return 0, true
	}
	if errors.Is(err, exec.ErrNotFound) || errors.Is(err, os.ErrNotExist) {
		fmt.Fprintf(stderr, "offload: %s: command not found\n", name)
		return 127, false
	}
	if path == "" {
		path = name
	}
	fmt.Fprintf(stderr, "offload: %s: %v\n", path, err)
	return 126, false
}

// targets resolves each host's transport once (X5): multiplexed hosts share one socket directory; when
// it cannot be set up they connect plainly, said once on stderr.
func targets(hosts []pool.Host, stderr io.Writer) map[string]remote.Target {
	out := make(map[string]remote.Target, len(hosts))
	var mux []string
	var resolved bool
	for _, h := range hosts {
		t := remote.Target{Address: h.Address}
		if h.Multiplex {
			if !resolved {
				var err error
				if mux, err = remote.Multiplex(os.Getenv); err != nil {
					fmt.Fprintf(stderr, "offload: connection multiplexing off: %v\n", err)
				}
				resolved = true
			}
			t.Opts = mux
		}
		out[h.Name] = t
	}
	return out
}

func filterHost(hosts []pool.Host, name string) []pool.Host {
	for _, h := range hosts {
		if h.Name == name {
			return []pool.Host{h}
		}
	}
	return nil
}

// ownerText = the slot owner suffix: " <word>…" shell-quoted (S3).
func ownerText(argv []string) string { return " " + remote.Join(argv) }

// poolOwnerText = " <host>: <root> <word>…" (S4).
func poolOwnerText(host, root string, argv []string) string {
	return " " + host + ": " + remote.Quote(root) + ownerText(argv)
}
