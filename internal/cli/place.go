package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/eturkes/offload/internal/mirror"
	"github.com/eturkes/offload/internal/pool"
	"github.com/eturkes/offload/internal/probe"
	"github.com/eturkes/offload/internal/remote"
	"github.com/eturkes/offload/internal/slots"
)

const (
	retryEvery = 2 * time.Second
	downFor    = 30 * time.Second
)

type placer struct {
	opts    options
	heavy   slots.Dir
	hosts   []pool.Host
	targets map[string]remote.Target
	tree    mirror.Tree
	stderr  io.Writer

	downUntil map[string]time.Time
	downWhy   map[string]string
	known     map[string]probe.Stats // last good probe per host (GPU inventory, L4)
}

type placement struct {
	local *slots.Slot
	host  pool.Host
	slot  *slots.Slot
	wd    *os.File
	stats probe.Stats
}

func (p placement) release() {
	p.local.Release()
	p.slot.Release()
	if p.wd != nil {
		_ = p.wd.Close()
	}
}

// place loops rounds (L2) until a slot is taken, ctx ends, or the request can never be met (L4).
// Pool first: the controller stays free for the owner's own work.
func (p *placer) place(ctx context.Context) (placement, error) {
	p.downUntil, p.downWhy, p.known = map[string]time.Time{}, map[string]string{}, map[string]probe.Stats{}
	waited := false
	for {
		if len(p.hosts) > 0 {
			pl, ok, err := p.tryPool(ctx)
			if err != nil || ok {
				return pl, err
			}
		}
		// A signal during the probe → no local slot (runCommand never releases one taken after it).
		if ctx.Err() != nil {
			return placement{}, ctx.Err()
		}
		if !p.opts.pool {
			s, ok, err := p.heavy.TryAcquire(ownerText(p.opts.argv))
			if err != nil || ok {
				return placement{local: s}, err
			}
		}
		if !waited {
			fmt.Fprintf(p.stderr, "offload: waiting for a free slot (%s)\n", p.occupancy())
			waited = true
		}
		select {
		case <-ctx.Done():
			return placement{}, ctx.Err()
		case <-time.After(retryEvery):
		}
	}
}

type candidate struct {
	host  pool.Host
	slot  *slots.Slot
	wd    *os.File
	stats probe.Stats
	err   error
}

// tryPool takes a slot + workdir lock on every eligible host that has both free, probes those
// hosts, keeps the least busy one and frees the rest.
func (p *placer) tryPool(ctx context.Context) (placement, bool, error) {
	now := time.Now()
	var cands []*candidate
	release := func(keep *candidate) {
		for _, c := range cands {
			if c != keep {
				c.slot.Release()
				_ = c.wd.Close()
			}
		}
	}
	for _, h := range p.hosts {
		if now.Before(p.downUntil[h.Name]) {
			continue
		}
		dir := slots.Host(os.Getenv, h.Name, h.Slots)
		s, ok, err := dir.TryAcquire(poolOwnerText(h.Name, p.tree.Root, p.opts.argv))
		if err != nil {
			release(nil)
			return placement{}, false, err
		}
		if !ok {
			continue
		}
		wd, held, err := slots.TryLock(dir.Base, filepath.Join(dir.Path, "wd-"+filepath.Base(p.tree.Workdir())+".lock"))
		if err != nil || !held {
			s.Release()
			if err != nil {
				release(nil)
				return placement{}, false, err
			}
			continue
		}
		cands = append(cands, &candidate{host: h, slot: s, wd: wd})
	}

	var wg sync.WaitGroup
	for _, c := range cands {
		wg.Go(func() { c.stats, c.err = probe.Probe(ctx, p.targets[c.host.Name]) })
	}
	wg.Wait()

	var best *candidate
	for _, c := range cands {
		if c.err != nil {
			if ctx.Err() == nil {
				p.downUntil[c.host.Name] = now.Add(downFor)
				p.downWhy[c.host.Name] = c.err.Error()
				delete(p.known, c.host.Name) // a down host's GPUs are unknown again (L4)
			}
			continue
		}
		delete(p.downWhy, c.host.Name)
		p.known[c.host.Name] = c.stats
		if p.opts.gpu && c.stats.GPUs == 0 {
			continue
		}
		if best == nil || c.stats.Busy() < best.stats.Busy() {
			best = c
		}
	}
	release(best)
	if best != nil {
		return placement{host: best.host, slot: best.slot, wd: best.wd, stats: best.stats}, true, nil
	}
	if p.opts.gpu && p.noGPUAnywhere() {
		return placement{}, false, errors.New("no pool host has an NVIDIA GPU")
	}
	return placement{}, false, nil
}

// noGPUAnywhere = every eligible host is up (its last probe answered) and none shows a GPU (L4).
func (p *placer) noGPUAnywhere() bool {
	for _, h := range p.hosts {
		if s, ok := p.known[h.Name]; !ok || s.GPUs > 0 || p.downWhy[h.Name] != "" {
			return false
		}
	}
	return true
}

func (p *placer) occupancy() string {
	var parts []string
	if !p.opts.pool {
		parts = append(parts, fmt.Sprintf("local %d/%d held", len(p.heavy.Holders()), p.heavy.Total))
	}
	for _, h := range p.hosts {
		if why, down := p.downWhy[h.Name]; down {
			parts = append(parts, fmt.Sprintf("%s down: %s", h.Name, why))
			continue
		}
		d := slots.Host(os.Getenv, h.Name, h.Slots)
		parts = append(parts, fmt.Sprintf("%s %d/%d held", h.Name, len(d.Holders()), d.Total))
	}
	return strings.Join(parts, "; ")
}

// status prints the local heavy slots, then one row per pool host (R2).
func status(stdout, stderr io.Writer) int {
	hosts, err := pool.Load(pool.Path(os.Getenv))
	if err != nil {
		fmt.Fprintf(stderr, "offload: %v\n", err)
		return ExitUsage
	}
	heavy, err := slots.Heavy(os.Getenv)
	if err != nil {
		fmt.Fprintf(stderr, "offload: %v\n", err)
		return ExitUsage
	}
	held := heavy.Holders()
	fmt.Fprintf(stdout, "local heavy slots: %d/%d held\n", len(held), heavy.Total)
	for _, line := range held {
		fmt.Fprintf(stdout, "  %s\n", line)
	}

	targets := targets(hosts, stderr)
	stats := make([]probe.Stats, len(hosts))
	errs := make([]error, len(hosts))
	var wg sync.WaitGroup
	for i, h := range hosts {
		wg.Go(func() { stats[i], errs[i] = probe.Probe(context.Background(), targets[h.Name]) })
	}
	wg.Wait()

	fmt.Fprintf(stdout, "\n%-14s %-6s %5s %5s %6s %5s %9s %4s\n", "HOST", "STATE", "SLOTS", "CPUS", "LOAD", "BUSY", "FREE MEM", "GPUS")
	for i, h := range hosts {
		d := slots.Host(os.Getenv, h.Name, h.Slots)
		used := fmt.Sprintf("%d/%d", len(d.Holders()), d.Total)
		if errs[i] != nil {
			fmt.Fprintf(stdout, "%-14s %-6s %5s  %s\n", h.Name, "down", used, errs[i])
			continue
		}
		s := stats[i]
		fmt.Fprintf(stdout, "%-14s %-6s %5s %5d %6.2f %4.0f%% %7.1fGB %4d\n", h.Name, "up", used, s.CPUs, s.Load,
			100*s.Busy(), float64(s.MemKB)/(1<<20), s.GPUs)
	}
	return 0
}
