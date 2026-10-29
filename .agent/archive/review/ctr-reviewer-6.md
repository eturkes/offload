# reviewer-6 — post-review-fix re-review

Target = `81a86d5..fa6936e`; execution revision = `fa6936eb99039adc837dd6b1e2cfe3e06bfc72a8`. Check set Q1–Q5 fixed before target read. Primary tree + remote hosts read-only; tests ran in a disposable git-archive copy. MAIN edited primary files after the initial findings message; this verdict remains against **fa6936e**, not those later edits.

| id | check (acceptance) | verdict (pass / finding) | finding ids | evidence |
|----|--------------------|---------------------------|-------------|----------|
| Q1 | R21: `podmanCommand`/`scopeManager` — X1 isolation; plain fallback; failed-probe fallback; environment boundary | finding | F1, F2 | X1 kill, systemd-unit/conmon, slot barrier, terminal stdin, image-owned environment + failed-probe fallback passed. New probe is outside the cancellation context; scoped podman receives a newly generated invocation id. |
| Q2 | R20: INVOCATION_ID dropped only for podman; regression red before the fix, green now | finding | F2 | R20 behavior independently red on `7a4a2d0` + green on `fa6936e`; inherited id removed without mutating parent environment. Literal final-podman-env claim fails after `systemd-run` regenerates the variable. |
| Q3 | `run(find, …)` seam preserves K3–K8: signal-context lookup, UsageError → 2, record/125 paths | pass | — | Wrapper still calls `Find(ctx, cwd)` inside `run`'s signal context. Existing boundary/fault tests + added lookup-context/usage test passed; CLI lookup cancellation passed for run/layers/check. |
| Q4 | R19 + e2e hygiene: storage version key, bus address, cleanup wait + removeStorage; checks/skip policy preserved | pass | — | Diff keeps existing assertions + failure/skip distinctions; adds pre-TempDir container wait, not relaxed assertions. Targeted real-podman tests passed with both REQUIRE switches set; test storage removed. Actual 4.9.3/5.8.8 execution not repeated; limits below. |
| Q5 | K8 + code.md describe session, scope + INVOCATION_ID implementation accurately | finding | F2 | Session + scope descriptions match; `without INVOCATION_ID` does not describe the scoped path's actual environment. |

## Findings

### F1 — major — new scope probe ignores setup cancellation

- Target: `internal/ctr/ctr.go:264,324–326` (primary path `/home/eturkes/Projects/offload/internal/ctr/ctr.go`).
- Wrong: `run` creates the K8 cancellation context, but `podmanCommand`/`scopeManager` do not accept it. `scopeManager` calls plain `exec.Command(...).Run()`. A blocked user-manager probe therefore survives a TERM delivered to `offload ctr`; the signal goroutine cancels `ctx`, yet the setup thread cannot reach its `got != 0` check until the probe returns. This reintroduces an uncancellable setup subprocess. Contract K3 places setup under K8 signals; K5 requires a setup signal to end the run with `128+n`.
- Red on untouched target production bytes: fake `systemd-run` signals readiness, then waits; send TERM only to offload, as an ordinary caller does. This is a process-lifecycle test, not an assertion about a private helper's shape. No real Podman/store needed.
- Command (inside target copy): `TMPDIR="$scratch/tmp" go test -count=1 -v -run '^TestReviewer6Q1SignalCancelsScopeProbe$' ./test/e2e`; rc **1**.

```text
=== RUN   TestReviewer6Q1SignalCancelsScopeProbe
    reviewer6_test.go:34: SIGTERM during scope setup did not cancel systemd-run probe within 2s
--- FAIL: TestReviewer6Q1SignalCancelsScopeProbe (2.03s)
FAIL
FAIL github.com/eturkes/offload/test/e2e 2.313s
```

- Repair acceptance: run the scope probe under the existing signal context; repeat this test → rc 143 promptly, probe reaped, no container launch. Keep the successful-probe + failing-probe fallback behavior.

### F2 — minor — scope launcher regenerates the variable claimed absent

- Targets: `internal/ctr/ctr.go:301–308`; `.agent/contract.md:75`; `.claude/rules/code.md:12` (under `/home/eturkes/Projects/offload/`).
- Wrong claim, **not an isolation failure**: removing the inherited `INVOCATION_ID` from `exec.Cmd.Env` does not leave the eventual scoped podman without that variable. `systemd-run --scope` creates its own id before it execs podman. Podman then sees the new scope's id, so conmon can remain with podman in that separate scope; this is still outside X1. The direct/plain branch does leave it unset.
- Exact target text (`rg -Fn` matches): contract = ``with `INVOCATION_ID` dropped from its environment``; rule = ``without `INVOCATION_ID` (with it, podman keeps conmon``. Anchors verified in the archived fa6936e files; later primary edits had already changed these bytes.
- Red on target: the actual `podmanCommand` launches a PATH fixture through the real user manager, with the same `Setsid` as Run. Fixture prints its environment + cgroup. Command: `TMPDIR="$scratch/tmp" go test -count=1 -v -run '^TestReviewer6Q1RealScopePodmanEnvironment$' ./internal/ctr`; rc **1** (also observed in the combined `^TestReviewer6Q[13]` run).

```text
podman environment + cgroup: INVOCATION_ID=20ba2aed0eb4487cab5abb4513658eaf
    0::/user.slice/user-1000.slice/user@1000.service/app.slice/run-p997648-i9385818.scope
podman inherited a systemd-generated INVOCATION_ID: "INVOCATION_ID=20ba2aed0eb4487cab5abb4513658eaf\n0::/user.slice/user-1000.slice/user@1000.service/app.slice/run-p997648-i9385818.scope\n"
--- FAIL: TestReviewer6Q1RealScopePodmanEnvironment (0.03s)
```

- Repair acceptance: describe the inherited-vs-generated distinction in K8 + the rule (recommended: no behavior change); retain plain-path id removal + X1 isolation. Alternatively make final absence an intentional tested behavior. Image CMD's environment stays correct: an added real-container test with image-owned `INVOCATION_ID=image-owned` + `XDG_RUNTIME_DIR=/image/runtime` passed despite a host id and runtime directory.

## Passing evidence

All target tests used `go test -count=1 -v -run '<regex>'`; only the named packages/tests ran. Local toolchain = Go 1.27.3; Podman 6.1.4; systemd 262. Production `internal/ctr/ctr.go` in the test copy matched `git show fa6936e:internal/ctr/ctr.go` by SHA256 `948252f8e0b3057bb0194410f1247bdbd1567860fdb69b3d9b2b3aecf2f7819d`.

- Unit regex `^(TestK3ReviewFind|TestK3ReviewFindFailures|TestK7ReviewRunBoundary|TestK5ReviewSignalAfterEnsure|TestK7ReviewSocketpairFailure)$`, `./internal/ctr` → rc 0, 0.080 s. Boundary includes record, missing record (podman success/failure), Ensure/Home/mount-path/binary-replacement/launch errors.
- Real-container regex `^(TestK8UnderSystemdUnitConmonStaysOutOfX1Scope|TestK8UnderOffloadRunKillEndsContainer|TestK8SlotFreesAfterContainerLastWrite|TestK6MountsHomeAndOwnership|TestK6StreamsAndStdinByteExact|TestK7StatusPassesThrough|TestK7NoStatusRecordIs125)$`, `./test/e2e`, `OFFLOAD_REQUIRE_PODMAN=1 OFFLOAD_REQUIRE_SCOPE=1` → rc 0, 12.550 s. No skips. Slot freed **161.097314 ms after** the container's last write; teardown tests retained their 9 s lower bound.
- Added `TestReviewer6Q1ProbeFailureFallsBack` → pass: a socket exists but fake systemd-run exits 1; direct podman selected, id unset there, original runtime preserved, parent id unchanged.
- Added `TestReviewer6Q3LookupContextAndUsage` → pass: injected UsageError returns 2; injected lookup waits for supplied context cancellation and TERM returns 143.
- Regex `^(TestReviewer6Q1ContainerEnvIsolation|TestReviewA3SignalCancelsProjectLookup|TestReviewG6TerminalStdinCompletes)$`, `./test/e2e`, both REQUIRE switches → rc 0, 5.308 s. No skips. Lookup cancellation passed run/layers/check subtests; PTY Ctrl-C returned CMD's 5 with one interrupt.
- R20 red-green independently confirmed. Pre-fix = archive `7a4a2d0`, only `test/e2e/ctr_test.go` overlaid from `d8503a0` to supply the regression test. Command: same TMPDIR + both REQUIRE switches, `go test -count=1 -v -run '^TestK8UnderSystemdUnitConmonStaysOutOfX1Scope$' ./test/e2e` → rc 1, 17.149 s:

```text
ctr_test.go:839: container record left after X1's KILL (conmon killed before --rm): "1f6716b64611743694e8690e89079cd78602fbbc0ee03641ef5b6bbd30db36c4 __ctr-init ctrhelper ignore k8unit-1792694696890588445 .../project/ready\n"
--- FAIL: TestK8UnderSystemdUnitConmonStaysOutOfX1Scope (16.63s)
```

Same test on target → pass, 11.90 s, included in the real-container command above.

## Reproduce the two reds

Create a fresh `/tmp/reviewer-6-*` directory, `repo/` + `tmp/` beneath it, then `git -C /home/eturkes/Projects/offload archive fa6936e | tar -x -C "$scratch/repo"`. Add the following tests in that copy alone; package helpers are already committed. Run the commands above from the copy. Remove the copy after (subuid-owned remnants → `podman unshare rm -rf` using that test store's config).

`test/e2e/reviewer6_test.go`:

```go
package e2e_test

import (
    "net"
    "os"
    "path/filepath"
    "syscall"
    "testing"
    "time"
)

func TestReviewer6Q1SignalCancelsScopeProbe(t *testing.T) {
    s := newSandbox(t, false)
    s.setenv("OFFLOAD_CTR", "1")
    s.setenv("XDG_CACHE_HOME", filepath.Join(s.base, "cache"))
    s.write(".devcontainer/Containerfile", "FROM scratch\n")
    must(t, os.MkdirAll(filepath.Join(s.runtime, "systemd"), 0700))
    listener, err := net.Listen("unix", filepath.Join(s.runtime, "systemd", "private"))
    must(t, err)
    defer listener.Close()
    ready := filepath.Join(s.base, "probe-ready")
    s.setenv("REVIEW_PROBE_READY", ready)
    s.write(filepath.Join(s.base, "bin", "podman"), "#!/bin/sh\n[ \"$1 $2\" = 'image exists' ] && exit 0\nexit 99\n")
    must(t, os.Chmod(filepath.Join(s.base, "bin", "podman"), 0755))
    s.write(filepath.Join(s.base, "bin", "systemd-run"), "#!/bin/sh\necho $$ > \"$REVIEW_PROBE_READY\"\nexec sleep 30\n")
    must(t, os.Chmod(filepath.Join(s.base, "bin", "systemd-run"), 0755))
    p := s.start("ctr", "--", "true")
    p.await("scope probe", func() bool { return exists(ready) })
    p.signal(syscall.SIGTERM)
    select {
    case <-p.done:
        p.wait(time.Second).requireCode(t, 143)
    case <-time.After(2 * time.Second):
        t.Fatal("SIGTERM during scope setup did not cancel systemd-run probe within 2s")
    }
}
```

`internal/ctr/reviewer6_test.go`:

```go
package ctr

import (
    "os"
    "path/filepath"
    "strings"
    "syscall"
    "testing"
)

func TestReviewer6Q1RealScopePodmanEnvironment(t *testing.T) {
    if scopeManager(os.Getenv) == "" {
        t.Fatal("real user scope unavailable for this acceptance test")
    }
    bin := t.TempDir()
    must(t, os.WriteFile(filepath.Join(bin, "podman"), []byte("#!/bin/sh\nprintf 'INVOCATION_ID=%s\\n' \"${INVOCATION_ID-unset}\"\ncat /proc/self/cgroup\n"), 0755))
    t.Setenv("PATH", bin+":"+os.Getenv("PATH"))
    t.Setenv("INVOCATION_ID", "0123456789abcdef0123456789abcdef")
    cmd := podmanCommand(os.Getenv, []string{"run"})
    cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
    out, err := cmd.Output()
    must(t, err)
    t.Logf("podman environment + cgroup: %s", out)
    if !strings.HasPrefix(string(out), "INVOCATION_ID=unset\n") {
        t.Fatalf("podman inherited a systemd-generated INVOCATION_ID: %q", out)
    }
}
```

## Register

None. F1/F2 bind to the seeded scope-manager/environment/claim checks; no wider review opened.

## Coverage + cleanup

- Read target production hunk with enclosing Run + callees; read e2e harness changes + lifecycle helpers, contract K3–K8, rules, ledger R19–R21. Q3's seam assessed by production diff + focused execution, not a new audit of every harvested test's mutation credit.
- Unconfirmed here: behavior on actual Podman 4.9.3/5.8.8, CI runs cited in the ledger, nixbox-specific repetition. Looked at version-selection code + ledger evidence; ran only local 6.1.4. No remote host accessed.
- No-user-manager branch: code inspection + failed-probe fallback fixture; an actually absent `/run/user/<uid>/systemd/private` was not manufactured on this live host.
- Not run: full gate, race lane, mutation suite, remote checks, install. LSP lookup attempted on `internal/ctr/ctr.go` → no Go server configured; code navigation used bounded text search.
- Cleanup complete: removed `/tmp/reviewer-6-zABswg`, rc 0; explicit absence check rc 0 (existence positive control rc 0 before removal). Pre-fix harness left only owner-owned store lock files; ordinary removal sufficed. Production tree never written by reviewer-6.
- Validator passed: positive-control count 1 / rc 0; deliverable pending-cell count 0 / grep rc 1 (zero matches). Five of five seeded rows adjudicated.
