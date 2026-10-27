package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/eturkes/offload/internal/mirror"
	"github.com/eturkes/offload/internal/pool"
	"github.com/eturkes/offload/internal/probe"
	"github.com/eturkes/offload/internal/slots"
)

type placementWriter func([]byte) (int, error)

func (w placementWriter) Write(b []byte) (int, error) { return w(b) }

func heldPlacementSlot(t *testing.T, d slots.Dir) *slots.Slot {
	t.Helper()
	s, ok, err := d.TryAcquire(" fixture")
	if err != nil || !ok {
		t.Fatalf("hold slot: ok=%v err=%v", ok, err)
	}
	t.Cleanup(s.Release)
	return s
}

func TestL3RetryAtTwoSeconds(t *testing.T) { // L3: the retry interval starts after the preceding round.
	synctest.Test(t, func(t *testing.T) {
		d := slots.Dir{Path: filepath.Join(t.TempDir(), "heavy"), Total: 1}
		held := heldPlacementSlot(t, d)
		p := placer{local: true, heavy: d, opts: options{argv: []string{"true"}}}
		p.stderr = placementWriter(func(b []byte) (int, error) {
			held.Release()
			return len(b), nil
		})
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		start := time.Now()
		pl, err := p.place(ctx)
		defer pl.release()
		if err != nil || pl.local == nil {
			t.Fatalf("retry did not take the released slot: %v", err)
		}
		if elapsed := time.Since(start); elapsed != 2*time.Second {
			t.Fatalf("next round after %s, want 2s", elapsed)
		}
	})
}

func TestL3WaitingLineOnceAcrossRounds(t *testing.T) { // L3: several unsuccessful rounds still emit one occupancy line.
	synctest.Test(t, func(t *testing.T) {
		d := slots.Dir{Path: filepath.Join(t.TempDir(), "heavy"), Total: 1}
		heldPlacementSlot(t, d)
		var stderr bytes.Buffer
		p := placer{local: true, heavy: d, stderr: &stderr, opts: options{argv: []string{"true"}}}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		pl, _ := p.place(ctx)
		defer pl.release()
		if got := stderr.String(); strings.Count(got, "offload: waiting") != 1 || !strings.Contains(got, "local 1/1 held") {
			t.Fatalf("waiting across three rounds = %q; want one line naming local 1/1 held", got)
		}
	})
}

func placementFixture(t *testing.T, names ...string) (*placer, string) {
	t.Helper()
	base := t.TempDir()
	bin, hosts := filepath.Join(base, "bin"), filepath.Join(base, "hosts")
	if err := os.MkdirAll(bin, 0o700); err != nil {
		t.Fatal(err)
	}
	ssh, err := os.ReadFile("../../test/e2e/testdata/ssh")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bin, "ssh"), ssh, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+":"+os.Getenv("PATH"))
	t.Setenv("OFFLOAD_FAKE_HOSTS", hosts)
	t.Setenv("XDG_RUNTIME_DIR", filepath.Join(base, "runtime"))
	p := &placer{
		opts: options{argv: []string{"true"}}, tree: mirror.Tree{Root: filepath.Join(base, "project"), Sub: "."}, stderr: io.Discard,
		downUntil: map[string]time.Time{}, downWhy: map[string]string{}, known: map[string]probe.Stats{},
	}
	for _, name := range names {
		p.hosts = append(p.hosts, pool.Host{Name: name, Address: "agent@" + name, Slots: 2})
		if err := os.MkdirAll(filepath.Join(hosts, name), 0o700); err != nil {
			t.Fatal(err)
		}
		placementProbe(t, hosts, name, "8 2 1048576 0\n")
	}
	p.targets = targets(p.hosts, io.Discard)
	return p, hosts
}

func placementProbe(t *testing.T, hosts, name, text string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(hosts, name, ".probe"), []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
}

func placementWorkdir(p *placer, name string) string {
	d := slots.Host(os.Getenv, name, 2)
	return filepath.Join(d.Path, "wd-"+filepath.Base(p.tree.Workdir())+".lock")
}

func requirePlacement(t *testing.T, p *placer, name string) placement {
	t.Helper()
	pl, ok, err := p.tryPool(t.Context())
	t.Cleanup(pl.release)
	if err != nil || !ok || pl.host.Name != name {
		t.Fatalf("placement=%q ok=%v err=%v, want %s", pl.host.Name, ok, err, name)
	}
	return pl
}

func TestL2RoundReleasesUnselectedSlots(t *testing.T) { // L2 + S3: losers are free while the selected run still holds its locks.
	p, _ := placementFixture(t, "alpha", "beta")
	requirePlacement(t, p, "alpha")
	d := slots.Host(os.Getenv, "beta", 2)
	if got := d.Holders(); len(got) != 0 {
		t.Fatalf("unselected host still holds a slot: %q", got)
	}
	b, err := os.ReadFile(filepath.Join(d.Path, "slot1.owner"))
	if err != nil || len(b) != 0 {
		t.Fatalf("unselected owner=%q err=%v, want empty", b, err)
	}
}

func TestL2LaterHostAfterBusyWorkdir(t *testing.T) { // L2 + S4: a busy workdir releases its slot and does not hide later hosts.
	p, _ := placementFixture(t, "alpha", "beta")
	f, held, err := slots.TryLock(os.Getenv("XDG_RUNTIME_DIR"), placementWorkdir(p, "alpha"))
	if err != nil || !held {
		t.Fatalf("hold workdir: held=%v err=%v", held, err)
	}
	defer func() { _ = f.Close() }()
	requirePlacement(t, p, "beta")
	if got := slots.Host(os.Getenv, "alpha", 2).Holders(); len(got) != 0 {
		t.Fatalf("busy workdir retained a candidate slot: %q", got)
	}
}

func TestL2LaterHostAfterDownHost(t *testing.T) { // L2: a failed probe or a cooldown excludes that host alone.
	for _, cooldown := range []bool{false, true} {
		t.Run(fmt.Sprintf("cooldown=%v", cooldown), func(t *testing.T) {
			p, hosts := placementFixture(t, "alpha", "beta")
			placementProbe(t, hosts, "alpha", "invalid\n")
			if cooldown {
				d := slots.Host(os.Getenv, "beta", 2)
				a, b := heldPlacementSlot(t, d), heldPlacementSlot(t, d)
				pl, ok, err := p.tryPool(t.Context())
				pl.release()
				if err != nil || ok {
					t.Fatalf("first round: ok=%v err=%v", ok, err)
				}
				a.Release()
				b.Release()
			}
			requirePlacement(t, p, "beta")
		})
	}
}

func TestL4RecoveredCPUHostEndsGPUWait(t *testing.T) { // L4: a successful retry supersedes the earlier down result.
	p, hosts := placementFixture(t, "alpha")
	p.opts.gpu = true
	placementProbe(t, hosts, "alpha", "invalid\n")
	synctest.Test(t, func(t *testing.T) {
		p.stderr = placementWriter(func(b []byte) (int, error) {
			placementProbe(t, hosts, "alpha", "8 2 1048576 0\n")
			time.Sleep(30 * time.Second)
			return len(b), nil
		})
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		pl, err := p.place(ctx)
		defer pl.release()
		if err == nil || !strings.Contains(err.Error(), "no pool host has an NVIDIA GPU") {
			t.Fatalf("all hosts now up without GPUs: err=%v, want no-GPU failure", err)
		}
	})
}

func TestS4PlacementKeepsWorkdirLockedThroughGC(t *testing.T) { // S4: the workdir remains excluded for the full placement lifetime.
	p, _ := placementFixture(t, "alpha")
	pl := requirePlacement(t, p, "alpha")
	for deadline := time.Now().Add(250 * time.Millisecond); time.Now().Before(deadline); {
		runtime.GC()
		f, held, err := slots.TryLock(os.Getenv("XDG_RUNTIME_DIR"), placementWorkdir(p, "alpha"))
		if err != nil {
			t.Fatal(err)
		}
		if held {
			_ = f.Close()
			t.Fatal("workdir lock became free while its placement was still live")
		}
		time.Sleep(time.Millisecond)
	}
	runtime.KeepAlive(pl)
}

func TestL3DownHostCooldownHasThirtySecondMinimum(t *testing.T) { // L3: no second probe before 30 seconds have elapsed.
	p, hosts := placementFixture(t, "alpha")
	placementProbe(t, hosts, "alpha", "invalid\n")
	synctest.Test(t, func(t *testing.T) {
		// A blocked diagnostic write shifts the 2-second rounds off whole-second boundaries.
		p.stderr = placementWriter(func(b []byte) (int, error) {
			time.Sleep(1500 * time.Millisecond)
			return len(b), nil
		})
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		pl, _ := p.place(ctx)
		defer pl.release()
		calls, err := filepath.Glob(filepath.Join(hosts, "alpha", ".ssh-call-*"))
		if err != nil || len(calls) != 1 {
			t.Fatalf("down host probed %d times before the 30s deadline; want one (err=%v)", len(calls), err)
		}
	})
}
