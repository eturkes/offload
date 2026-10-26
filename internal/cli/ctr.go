package cli

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"
	"sync"
	"syscall"

	"github.com/eturkes/offload/internal/ctr"
	"github.com/eturkes/offload/internal/pool"
	"github.com/eturkes/offload/internal/probe"
	"github.com/eturkes/offload/internal/remote"
)

// switchVar = the per-machine switch (K2): `1` runs ctr commands in the project's image.
const switchVar = "OFFLOAD_CTR"

type ctrOptions struct {
	layers, check bool
	argv          []string
}

func parseCtr(args []string) (ctrOptions, error) {
	var o ctrOptions
	fs := flag.NewFlagSet("ctr", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.BoolVar(&o.layers, "layers", false, "")
	fs.BoolVar(&o.check, "check", false, "")
	if err := fs.Parse(args); err != nil {
		return o, err
	}
	o.argv = fs.Args()
	switch {
	case o.layers && o.check:
		return o, errors.New("--layers excludes --check")
	case (o.layers || o.check) && len(o.argv) > 0:
		return o, errors.New("--layers and --check take no command")
	case !o.layers && !o.check && len(o.argv) == 0:
		return o, errors.New("give a command after `ctr`, e.g. `offload ctr -- make -j8`")
	}
	return o, nil
}

func ctrCommand(args []string, stdout, stderr io.Writer) int {
	o, err := parseCtr(args)
	if errors.Is(err, flag.ErrHelp) {
		fmt.Fprint(stdout, usage)
		return 0
	}
	if err != nil {
		return usageError(stderr, err.Error())
	}
	switch {
	case o.layers:
		return ctrLayers(stdout, stderr)
	case o.check:
		return ctrCheck(stdout, stderr)
	}
	switch v := os.Getenv(switchVar); v {
	case "", "0":
		return execNative(o.argv, stderr)
	case "1":
	default:
		fmt.Fprintf(stderr, "offload: %s=%q: use 1 (image) or 0 (native)\n", switchVar, v)
		return ExitUsage
	}
	sigs := ctr.Notify() // before any work: a signal during setup ends the run (K5)
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "offload: %v\n", err)
		return ExitFailure
	}
	return ctr.Run(cwd, o.argv, os.Getenv, sigs, stderr)
}

// execNative replaces offload with argv (K2): the switch is off, so the command runs as if typed.
func execNative(argv []string, stderr io.Writer) int {
	path, rc, msg := lookLocal(argv[0])
	if rc != 0 {
		fmt.Fprintln(stderr, msg)
		return rc
	}
	err := syscall.Exec(path, argv, os.Environ()) //nolint:gosec // G204: the caller's own command, as given
	fmt.Fprintf(stderr, "offload: %s: %v\n", argv[0], err)
	return 126
}

// localLayers builds the image on miss + returns its layer digests (K9).
func localLayers(ctx context.Context, stderr io.Writer) ([]string, int) {
	cwd, err := os.Getwd()
	var p ctr.Project
	if err == nil {
		p, err = ctr.Find(ctx, cwd)
	}
	var ue ctr.UsageError
	if errors.As(err, &ue) {
		fmt.Fprintf(stderr, "offload: %v\n", err)
		return nil, ExitUsage
	}
	var tag string
	if err == nil {
		tag, err = ctr.Ensure(ctx, p, os.Getenv, stderr)
	}
	var layers []string
	if err == nil {
		layers, err = ctr.Layers(ctx, tag)
	}
	if err != nil {
		if ctx.Err() == nil { // a signal's own status follows; its fallout is no news
			fmt.Fprintf(stderr, "offload: %v\n", err)
		}
		return nil, ExitFailure
	}
	return layers, 0
}

func ctrLayers(stdout, stderr io.Writer) int {
	ctx, signalStatus := signals(ctr.Forwarded...)
	layers, rc := localLayers(ctx, stderr)
	if ctx.Err() != nil {
		return signalStatus()
	}
	if rc != 0 {
		return rc
	}
	for _, l := range layers {
		fmt.Fprintln(stdout, l)
	}
	return 0
}

// remoteLayers = the command --check runs on a pool host: its installed offload's sha256, then its
// layer digests, each line framed with this run's nonce ($1) so nothing a login shell prints can
// pass for either (K10).
const remoteLayers = `set -eo pipefail; s=$(sha256sum "$(command -v offload)"); printf 'offload-check %s %s\n' "$1" "${s%% *}"
offload ctr --layers | sed "s/^/layer $1 /"`

// ctrCheck compares this host's image layers + offload binary with every pool host's (K10).
func ctrCheck(stdout, stderr io.Writer) int {
	ctx, signalStatus := signals(ctr.Forwarded...)
	hosts, err := pool.Load(pool.Path(os.Getenv))
	if err != nil {
		fmt.Fprintf(stderr, "offload: %v\n", err)
		return ExitUsage
	}
	self, err := ctr.Executable()
	var sum string
	if err == nil {
		sum, err = fileSHA256(self)
	}
	if err != nil {
		fmt.Fprintf(stderr, "offload: %v\n", err)
		return ExitFailure
	}
	here, rc := localLayers(ctx, stderr)
	if ctx.Err() != nil {
		return signalStatus()
	}
	if rc != 0 {
		return rc
	}

	lines := make([]string, len(hosts))
	same := make([]bool, len(hosts))
	tgs := targets(hosts, stderr)
	var wg sync.WaitGroup
	for i, h := range hosts {
		wg.Go(func() {
			// --host waits while the host is down (L4) ⇒ a down host is reported, not waited on.
			if _, err := probe.Probe(ctx, tgs[h.Name]); err != nil {
				lines[i] = fmt.Sprintf("%s: failed (down: %v)", h.Name, err)
				return
			}
			var out bytes.Buffer
			nonce := rand.Text()
			// remote.Command: a signal reaches the host run as TERM (it tears down + logs its end, J2),
			// never as exec's default KILL.
			cmd := remote.Command(ctx, self, "run", "--host", h.Name, "--", "bash", "-c", remoteLayers, "_", nonce)
			cmd.Stdout, cmd.Stderr = &out, stderr
			err := cmd.Run()
			var ee *exec.ExitError
			switch {
			case errors.As(err, &ee):
				lines[i] = fmt.Sprintf("%s: failed (rc %d)", h.Name, ee.ExitCode())
				return
			case err != nil:
				lines[i] = fmt.Sprintf("%s: failed (%v)", h.Name, err)
				return
			}
			theirSum, there, ok := parseCheck(out.String(), nonce)
			if !ok {
				lines[i] = fmt.Sprintf("%s: failed (no offload-check line in its output)", h.Name)
				return
			}
			if slices.Equal(here, there) {
				lines[i], same[i] = fmt.Sprintf("%s: same (%d layers)", h.Name, len(here)), theirSum == sum
			} else {
				lines[i] = fmt.Sprintf("%s: different (here %d, there %d layers)", h.Name, len(here), len(there))
			}
			if theirSum != sum {
				lines[i] += "; offload binary differs"
			}
		})
	}
	wg.Wait()
	if ctx.Err() != nil {
		return signalStatus()
	}
	rc = 0
	for i, l := range lines {
		fmt.Fprintln(stdout, l)
		if !same[i] {
			rc = 1
		}
	}
	return rc
}

// parseCheck reads remoteLayers' lines framed with nonce; ok = its offload-check line was there.
// Split, not bufio.Scanner: a login banner line past Scanner's 64 KiB token limit would end the scan
// early and drop the layers after it.
func parseCheck(out, nonce string) (sum string, layers []string, ok bool) {
	for _, line := range strings.Split(out, "\n") {
		if v, found := strings.CutPrefix(line, "offload-check "+nonce+" "); found {
			sum, ok = v, true
		} else if v, found := strings.CutPrefix(line, "layer "+nonce+" "); found {
			layers = append(layers, v)
		}
	}
	return sum, layers, ok
}

func fileSHA256(path string) (string, error) {
	data, err := os.ReadFile(path) //nolint:gosec // G304: this binary
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:]), nil
}
