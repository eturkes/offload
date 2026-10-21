// Package probe measures a pool host's capacity + load in one ssh round trip.
package probe

import (
	"context"
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/eturkes/offload/internal/remote"
)

// Script prints nproc, 1-min loadavg, MemAvailable kB and the NVIDIA GPU count, one per line.
const Script = `nproc; cut -d' ' -f1 /proc/loadavg; awk '/^MemAvailable:/{print $2}' /proc/meminfo; ` +
	`{ nvidia-smi -L 2>/dev/null || true; } | grep -c '^GPU' || true`

// Timeout bounds one probe end to end; a slower host counts as down. pipeGrace of it goes to
// waiting out descendants of a killed ssh that still hold its pipes.
const (
	Timeout   = 8 * time.Second
	pipeGrace = 500 * time.Millisecond
)

// Stats = one host's measurements.
type Stats struct {
	CPUs  int
	Load  float64
	MemKB int64
	GPUs  int
}

// Busy = load per CPU, the placement score (lower = better).
func (s Stats) Busy() float64 { return s.Load / float64(s.CPUs) }

// Parse reads Script's output.
func Parse(out string) (Stats, error) {
	f := strings.Fields(out)
	if len(f) != 4 {
		return Stats{}, fmt.Errorf("probe printed %d fields, want 4", len(f))
	}
	var s Stats
	var err error
	if s.CPUs, err = strconv.Atoi(f[0]); err != nil || s.CPUs < 1 {
		return Stats{}, fmt.Errorf("probe: bad CPU count %q", f[0])
	}
	if s.Load, err = strconv.ParseFloat(f[1], 64); err != nil || !(s.Load >= 0) || math.IsInf(s.Load, 1) {
		return Stats{}, fmt.Errorf("probe: bad load %q", f[1])
	}
	if s.MemKB, err = strconv.ParseInt(f[2], 10, 64); err != nil || s.MemKB < 0 {
		return Stats{}, fmt.Errorf("probe: bad MemAvailable %q", f[2])
	}
	if s.GPUs, err = strconv.Atoi(f[3]); err != nil || s.GPUs < 0 {
		return Stats{}, fmt.Errorf("probe: bad GPU count %q", f[3])
	}
	return s, nil
}

// Probe measures the host at t. The error carries ssh's last stderr line.
func Probe(ctx context.Context, t remote.Target) (Stats, error) {
	ctx, cancel := context.WithTimeout(ctx, Timeout-pipeGrace)
	defer cancel()
	//nolint:gosec // G204: address = trusted pool file entry, after "--"; Script is a constant
	cmd := exec.CommandContext(ctx, "ssh", t.SSHArgs(Script)...)
	// Descendants of a killed ssh can keep its pipes open; WaitDelay caps the wait past the deadline.
	cmd.WaitDelay = pipeGrace
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if ctx.Err() != nil {
		return Stats{}, fmt.Errorf("no answer within %s", Timeout)
	}
	if err != nil {
		lines := strings.Split(strings.TrimSpace(stderr.String()), "\n")
		if msg := lines[len(lines)-1]; msg != "" {
			return Stats{}, fmt.Errorf("%s", msg)
		}
		return Stats{}, fmt.Errorf("ssh: %w", err)
	}
	return Parse(string(out))
}
