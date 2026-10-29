# reviewer-11 — survivor rulings, `internal/cli/place.go` (sweep at b7645d8, CI replay run 40534349336)

| id | site | func | mutator | change | replay | ruling |
|----|------|------|---------|--------|--------|--------|
| `150cbd04` | `place.go:22` | (pkg) | numbers/decrementer | Changes: `retryEvery = 2 * time.Second` → `retryEvery = 1 * time.Second` | full suite: pass | new test: L3 — next round starts after 1 s, not 2 s; `TestL3RetryAtTwoSeconds` |
| `69b4ce86` | `place.go:22` | (pkg) | arithmetic/base | Changes: `retryEvery = 2 * time.Second` → `retryEvery = 2 / time.Second` | full suite: pass | new test: L3 — next round starts immediately, not after 2 s; `TestL3RetryAtTwoSeconds` |
| `a91ef094` | `place.go:22` | (pkg) | numbers/incrementer | Changes: `retryEvery = 2 * time.Second` → `retryEvery = 3 * time.Second` | full suite: pass | new test: L3 — next round starts after 3 s, not 2 s; `TestL3RetryAtTwoSeconds` |
| `236ae769` | `place.go:23` | (pkg) | numbers/decrementer | Changes: `downFor    = 30 * time.Second` → `downFor    = 29 * time.Second` | full suite: pass | new test: L3 — a down host is probed again at 29.5 s; `TestL3DownHostCooldownHasThirtySecondMinimum` |
| `ff9eb11d` | `place.go:23` | (pkg) | numbers/incrementer | Changes: `downFor    = 30 * time.Second` → `downFor    = 31 * time.Second` | full suite: pass | contract-open: 31 s cooldown still meets L3’s “at most every 30 s”; no latest recovery deadline |
| `a84dcbb0` | `place.go:50` | placement.release | statement/remove | Changes: `p.slot.Release()` → `_ = p.slot.Release` | 89 covering tests: pass | new test: S3/S4 — normal completion leaves pool slot owner text behind; `TestS3PoolReleaseClearsOwner` |
| `11bdf2f2` | `place.go:51` | placement.release | conditional/negated | Changes: `if p.wd != nil {` → `if p.wd == nil {` | 89 covering tests: pass | contract-open: workdir fd survives until GC/process exit; X3/S4 fix exclusion + eventual release, not release before the end-log write; nil File.Close returns ErrInvalid |
| `fafc0e4e` | `place.go:52` | placement.release | branch/if | `- _ = p.wd.Close()`; `- }`; `+ _ = p.wd.Close`; `+ }` | 73 covering tests: pass | contract-open: workdir fd survives until GC/process exit; X3/S4 leave the post-command/pre-exit release instant open |
| `16498274` | `place.go:62` | placer.place | numbers/decrementer | Changes: `if len(p.hosts) > 0 {` → `if len(p.hosts) > (-1) {` | 98 covering tests: pass | equivalent: no hosts → tryPool returns empty,false,nil; --gpu cannot reach an empty eligible host list (cli.go runLogged) |
| `92b8b879` | `place.go:62` | placer.place | expression/comparison | Changes: `if len(p.hosts) > 0 {` → `if len(p.hosts) >= 0 {` | 98 covering tests: pass | equivalent: as 16498274; empty-host tryPool adds no effects or placement |
| `6bcd43c0` | `place.go:70` | placer.place | statement/return | `- return placement{}, ctx.Err()`; `- }`; `+ return placement{}, nil`; `+ }` | 1 covering tests: pass | equivalent: runLogged checks ctx.Err before the returned err or placement (cli.go:260–266) |
| `2474637f` | `place.go:80` | placer.place | statement/remove | `- fmt.Fprintf(p.stderr, "offload: waiting for a free slot (%s)\n", p.occupancy())`; `- waited = true`; `- }`; `+ fmt.Fprintf(p.stderr, "offload: waiting for a free slot (%s)\n", p.occupancy())`; `+ _ = waited`; `+ }` | 6 covering tests: pass | new test: L3 — three waiting lines instead of one; `TestL3WaitingLineOnceAcrossRounds` |
| `a806646b` | `place.go:80` | placer.place | conditional/bool-literal | Changes: `waited = true` → `waited = false` | 6 covering tests: pass | new test: L3 — three waiting lines instead of one; `TestL3WaitingLineOnceAcrossRounds` |
| `6a9244a3` | `place.go:83` | placer.place | select/case-remove | `- case <-ctx.Done():`; `- return placement{}, ctx.Err()`; `- case <-time.After(retryEvery):`; `- }`; `+ `; `+ case <-time.After(retryEvery):`; `+ }` | full suite: pass | contract-open: cancellation waits for the ≤2 s retry timer, then the unchanged ctx.Err guard exits; L3 specifies status/locks, no tighter latency |
| `090c7982` | `place.go:84` | placer.place | statement/return | Changes: `return placement{}, ctx.Err()` → `return placement{}, nil` | 5 covering tests: pass | equivalent: runLogged checks ctx.Err before returned err/placement; the cancelled context still fixes C3’s signal status |
| `f22ca4f6` | `place.go:106` | placer.tryPool | statement/remove | `- c.slot.Release()`; `- _ = c.wd.Close()`; `- }`; `+ _ = c.slot.Release`; `+ _ = c.wd.Close()`; `+ }` | 7 covering tests: pass | new test: L2/S3 — the unselected host retains its slot + owner; `TestL2RoundReleasesUnselectedSlots` |
| `3865b1b4` | `place.go:113` | placer.tryPool | loop/break | `- if now.Before(p.downUntil[h.Name]) {`; `- continue`; `- }`; `- dir := slots.Host(os.Getenv, h.Name, h.Slots)`; `- s, ok, err := dir.TryAcquire(poolOwnerText(h.Name, p.tree.Root, p.opts.argv))`; `+ if now.Before(p.downUntil[h.Name]) {`; `+ break`; `+ }`; `+ dir := slots.Host(os.Getenv, h.Name, h.S | 1 covering tests: pass | new test: L2 — a cooling-down first host hides a later eligible host; `TestL2LaterHostAfterDownHost` |
| `ea3948b7` | `place.go:118` | placer.tryPool | statement/remove | `- release(nil)`; `- return placement{}, false, err`; `- }`; `+ _ = release`; `+ return placement{}, false, err`; `+ }` | 1 covering tests: pass | new test: S3/S4 — a later slot-open error leaves the earlier candidate owner uncleared; `TestS4AcquisitionErrorClearsEarlierCandidates` |
| `7629ba80` | `place.go:125` | placer.tryPool | expression/remove | `- if err != nil \|\| !held {`; `+ if false \|\|`; `+ !held {` | 78 covering tests: pass | equivalent: TryLock returns held=false on every error (slots.go:260–265 → tryLock:209–225); !held already enters the error branch |
| `a3e8fd0c` | `place.go:126` | placer.tryPool | statement/remove | `- s.Release()`; `- if err != nil {`; `- release(nil)`; `- return placement{}, false, err`; `- }`; `- continue`; `- }`; `+ _ = s.Release`; `+ if err != nil {`; `+ release(nil)`; `+ return placement{}, false, err`; `+ }`; `+ continue`; `+ }` | 2 covering tests: pass | new test: S4 — a busy workdir retains its candidate slot; `TestL2LaterHostAfterBusyWorkdir` |
| `3322aa31` | `place.go:127` | placer.tryPool | conditional/negated | Changes: `if err != nil {` → `if err == nil {` | 2 covering tests: pass | new test: L2/S4 — a busy first workdir aborts the round before the later free host; `TestL2LaterHostAfterBusyWorkdir` |
| `3bd2407a` | `place.go:127` | placer.tryPool | expression/error-guard | `- if err != nil {`; `- release(nil)`; `- return placement{}, false, err`; `- }`; `+ if false {`; `+ release(nil)`; `+ return placement{}, false, err`; `+ }` | 2 covering tests: pass | new test: S2/C3 — a workdir-open error is swallowed and the command runs; `TestS4AcquisitionErrorClearsEarlierCandidates` |
| `19abdc33` | `place.go:128` | placer.tryPool | statement/remove | `- release(nil)`; `- return placement{}, false, err`; `- }`; `+ _ = release`; `+ return placement{}, false, err`; `+ }` | no e2e test executes the line | new test: S3/S4 — a later workdir-open error leaves the earlier candidate owner uncleared; `TestS4AcquisitionErrorClearsEarlierCandidates` |
| `3afd948b` | `place.go:128` | placer.tryPool | branch/if | `- release(nil)`; `- return placement{}, false, err`; `- }`; `+ _, _ = release, err`; `+ `; `+ }` | no e2e test executes the line | new test: S2/C3 — a workdir-open error is swallowed and the command runs; `TestS4AcquisitionErrorClearsEarlierCandidates` |
| `209c0e7d` | `place.go:129` | placer.tryPool | statement/return | `- release(nil)`; `- return placement{}, false, err`; `- }`; `+ release(nil)`; `+ return placement{}, false, nil`; `+ }` | no e2e test executes the line | new test: S2/C3 — a workdir-open error becomes local execution instead of 125; `TestS4AcquisitionErrorClearsEarlierCandidates` |
| `5db71380` | `place.go:131` | placer.tryPool | loop/break | `- if now.Before(p.downUntil[h.Name]) {`; `- continue`; `- }`; `- dir := slots.Host(os.Getenv, h.Name, h.Slots)`; `- s, ok, err := dir.TryAcquire(poolOwnerText(h.Name, p.tree.Root, p.opts.argv))`; `+ if now.Before(p.downUntil[h.Name]) {`; `+ continue`; `+ }`; `+ dir := slots.Host(os.Getenv, h.Name,  | 2 covering tests: pass | new test: L2/S4 — a busy first workdir hides a later free host; `TestL2LaterHostAfterBusyWorkdir` |
| `f58073c7` | `place.go:145` | placer.tryPool | expression/error-guard | `- if ctx.Err() == nil {`; `- p.downUntil[c.host.Name] = now.Add(downFor)`; `- p.downWhy[c.host.Name] = c.err.Error()`; `- delete(p.known, c.host.Name) // a down host's GPUs are unknown again (L4)`; `- }`; `+ if true {`; `+ p.downUntil[c.host.Name] = now.Add(downFor)`; `+ p.downWhy[c.host.Name] = c. | 5 covering tests: pass | equivalent: maps are changed only after cancellation, then place/runLogged exit on ctx.Err without reading them again |
| `b57b381e` | `place.go:147` | placer.tryPool | statement/remove | `- p.downUntil[c.host.Name] = now.Add(downFor)`; `- p.downWhy[c.host.Name] = c.err.Error()`; `- delete(p.known, c.host.Name) // a down host's GPUs are unknown again (L4)`; `- }`; `+ p.downUntil[c.host.Name] = now.Add(downFor)`; `+ _, _, _ = p.downWhy, c.host.Name, c.err.Error`; `+ delete(p.known, c. | 4 covering tests: pass | new test: R1/L3 — a down host is displayed as occupancy instead of down with its reason; `TestL3WaitingNamesEveryPlace` |
| `70181c0d` | `place.go:148` | placer.tryPool | statement/remove | `- p.downUntil[c.host.Name] = now.Add(downFor)`; `- p.downWhy[c.host.Name] = c.err.Error()`; `- delete(p.known, c.host.Name) // a down host's GPUs are unknown again (L4)`; `- }`; `+ p.downUntil[c.host.Name] = now.Add(downFor)`; `+ p.downWhy[c.host.Name] = c.err.Error()`; `+ _, _ = p.known, c.host.Na | 4 covering tests: pass | equivalent: downWhy remains nonempty, so noGPUAnywhere still rejects that host; a successful retry replaces known and clears downWhy |
| `cd10da7f` | `place.go:150` | placer.tryPool | loop/break | `- if c.err != nil {`; `- if ctx.Err() == nil {`; `- p.downUntil[c.host.Name] = now.Add(downFor)`; `- p.downWhy[c.host.Name] = c.err.Error()`; `- delete(p.known, c.host.Name) // a down host's GPUs are unknown again (L4)`; `- }`; `- continue`; `- }`; `- delete(p.downWhy, c.host.Name)`; `- p.known[c.h | 5 covering tests: pass | new test: L2 — the first failed probe hides a later successful candidate; `TestL2LaterHostAfterDownHost` |
| `36fcaaea` | `place.go:152` | placer.tryPool | statement/remove | `- if c.err != nil {`; `- if ctx.Err() == nil {`; `- p.downUntil[c.host.Name] = now.Add(downFor)`; `- p.downWhy[c.host.Name] = c.err.Error()`; `- delete(p.known, c.host.Name) // a down host's GPUs are unknown again (L4)`; `- }`; `- continue`; `- }`; `- delete(p.downWhy, c.host.Name)`; `- p.known[c.h | 76 covering tests: pass | new test: L4 — a recovered CPU-only host keeps the request waiting instead of failing; `TestL4RecoveredCPUHostEndsGPUWait` |
| `1016409a` | `place.go:163` | placer.tryPool | composite/field-clear | Changes: `return placement{host: best.host, slot: best.slot, wd: best.wd, stats: best.stats}, true, nil` → `return placement{host: best.host, slot: best.slot, stats: best.stats}, true, nil` | 73 covering tests: pass | new test: S4 — GC frees the workdir lock while the placement is still live; `TestS4PlacementKeepsWorkdirLockedThroughGC` |
| `7f7f476f` | `place.go:165` | placer.tryPool | expression/remove | `- if p.opts.gpu && p.noGPUAnywhere() {`; `+ if true &&`; `+ p.noGPUAnywhere() {` | 9 covering tests: pass | equivalent: without --gpu every successful probe returns a placement immediately; an ongoing non-GPU wait never has all hosts in known |
| `801c4c06` | `place.go:174` | placer.noGPUAnywhere | expression/remove | Changes: `if s, ok := p.known[h.Name]; !ok \|\| s.GPUs > 0 \|\| p.downWhy[h.Name] != "" {` → `if s, ok := p.known[h.Name]; !ok \|\| s.GPUs > 0 \|\| false {` | 3 covering tests: pass | equivalent: every non-cancelled failed probe deletes known[h]; a successful probe clears downWhy[h], so !ok already covers every down host |
| `876bd4fb` | `place.go:174` | placer.noGPUAnywhere | numbers/incrementer | Changes: `if s, ok := p.known[h.Name]; !ok \|\| s.GPUs > 0 \|\| p.downWhy[h.Name] != "" {` → `if s, ok := p.known[h.Name]; !ok \|\| s.GPUs > 1 \|\| p.downWhy[h.Name] != "" {` | 3 covering tests: pass | equivalent: any positive-GPU probe wins or accompanies another winner and returns before noGPUAnywhere; known GPU=1 is unreachable in an ongoing wait |
| `0d100f50` | `place.go:184` | placer.occupancy | statement/remove | `- parts = append(parts, fmt.Sprintf("local %d/%d held", len(p.heavy.Holders()), p.heavy.Total))`; `- }`; `+ _, _, _, _, _ = parts, parts, fmt.Sprintf, p.heavy.Holders, p.heavy.Total`; `+ }` | 1 covering tests: pass | new test: L3 — the waiting line omits local occupancy; `TestL3WaitingLineOnceAcrossRounds` |
| `7cd66ee2` | `place.go:184` | placer.occupancy | branch/if | `- parts = append(parts, fmt.Sprintf("local %d/%d held", len(p.heavy.Holders()), p.heavy.Total))`; `- }`; `+ _, _, _, _, _ = parts, parts, fmt.Sprintf, p.heavy.Holders, p.heavy.Total`; `+ }` | 1 covering tests: pass | new test: L3 — the waiting line omits local occupancy; `TestL3WaitingLineOnceAcrossRounds` |
| `67232af2` | `place.go:188` | placer.occupancy | branch/if | `- parts = append(parts, fmt.Sprintf("%s down: %s", h.Name, why))`; `- continue`; `- }`; `+ _, _, _, _, _ = parts, parts, fmt.Sprintf, h.Name, why`; `+ `; `+ }` | 2 covering tests: pass | new test: R1/L3 — the waiting line loses a down host’s state/reason; `TestL3WaitingNamesEveryPlace` |
| `ae331179` | `place.go:188` | placer.occupancy | statement/remove | `- parts = append(parts, fmt.Sprintf("%s down: %s", h.Name, why))`; `- continue`; `- }`; `+ _, _, _, _, _ = parts, parts, fmt.Sprintf, h.Name, why`; `+ continue`; `+ }` | 2 covering tests: pass | new test: L3 — the waiting line omits the down host entirely; `TestL3WaitingNamesEveryPlace` |
| `07423c5b` | `place.go:189` | placer.occupancy | loop/break | `- if why, down := p.downWhy[h.Name]; down {`; `- parts = append(parts, fmt.Sprintf("%s down: %s", h.Name, why))`; `- continue`; `- }`; `- d := slots.Host(os.Getenv, h.Name, h.Slots)`; `- parts = append(parts, fmt.Sprintf("%s %d/%d held", h.Name, len(d.Holders()), d.Total))`; `+ if why, down := p.do | 2 covering tests: pass | new test: L3 — a down first host suppresses occupancy of later hosts; `TestL3WaitingNamesEveryPlace` |
| `be934352` | `place.go:205` | status | expression/error-guard | `- if err != nil {`; `- fmt.Fprintf(stderr, "offload: %v\n", err)`; `- return ExitUsage`; `- }`; `+ if false {`; `+ fmt.Fprintf(stderr, "offload: %v\n", err)`; `+ return ExitUsage`; `+ }` | 5 covering tests: pass | new test: S1/C3 — invalid HEAVY_SLOTS produces status success instead of a usage error; `TestS1StatusRejectsInvalidHeavySlots` |
| `369df7a7` | `place.go:206` | status | branch/if | `- fmt.Fprintf(stderr, "offload: %v\n", err)`; `- return ExitUsage`; `- }`; `+ _, _, _ = fmt.Fprintf, stderr, err`; `+ `; `+ }` | no e2e test executes the line | new test: S1/C3 — invalid HEAVY_SLOTS produces status success instead of a usage error; `TestS1StatusRejectsInvalidHeavySlots` |
| `bf895ab1` | `place.go:206` | status | statement/remove | `- fmt.Fprintf(stderr, "offload: %v\n", err)`; `- return ExitUsage`; `- }`; `+ _, _, _ = fmt.Fprintf, stderr, err`; `+ return ExitUsage`; `+ }` | no e2e test executes the line | new test: S1/C4 — invalid HEAVY_SLOTS exits without its diagnostic; `TestS1StatusRejectsInvalidHeavySlots` |
| `11773c4f` | `place.go:207` | status | statement/return | `- fmt.Fprintf(stderr, "offload: %v\n", err)`; `- return ExitUsage`; `- }`; `+ fmt.Fprintf(stderr, "offload: %v\n", err)`; `+ return 0`; `+ }` | no e2e test executes the line | new test: S1/C3 — invalid HEAVY_SLOTS returns 0 instead of 2; `TestS1StatusRejectsInvalidHeavySlots` |
| `9db15d80` | `place.go:224` | status | statement/remove | Changes: `fmt.Fprintf(stdout, "\n%-14s %-6s %5s %5s %6s %5s %9s %4s\n", "HOST", "STATE", "SLOTS", "CPUS", "LOAD", "BUSY", "FREE MEM", "GPUS")` → `_, _ = fmt.Fprintf, stdout` | 5 covering tests: pass | contract-open: R2 fixes host rows + values/order, not a header row or its labels |
| `241a4e88` | `place.go:230` | status | loop/break | `- d := slots.Host(os.Getenv, h.Name, h.Slots)`; `- used := fmt.Sprintf("%d/%d", len(d.Holders()), d.Total)`; `- if errs[i] != nil {`; `- fmt.Fprintf(stdout, "%-14s %-6s %5s  %s\n", h.Name, "down", used, errs[i])`; `- continue`; `- }`; `- s := stats[i]`; `- fmt.Fprintf(stdout, "%-14s %-6s %5s %5d %6 | 1 covering tests: pass | new test: R2 — a down first host suppresses later status rows; `TestR2StatusContinuesAfterDownHost` |
| `37b44696` | `place.go:234` | status | numbers/incrementer | Changes: `100*s.Busy(), float64(s.MemKB)/(1<<20), s.GPUs)` → `101*s.Busy(), float64(s.MemKB)/(1<<20), s.GPUs)` | 5 covering tests: pass | new test: R2 — load 12 / 8 CPUs prints 152% busy instead of 150%; `TestR2StatusContinuesAfterDownHost` |
| `f48dfbe3` | `place.go:234` | status | numbers/decrementer | Changes: `100*s.Busy(), float64(s.MemKB)/(1<<20), s.GPUs)` → `99*s.Busy(), float64(s.MemKB)/(1<<20), s.GPUs)` | 5 covering tests: pass | new test: R2 — load 12 / 8 CPUs prints 148% busy instead of 150%; `TestR2StatusContinuesAfterDownHost` |

## Notes

Read target = 14994a8b6d9831e26e1dc060d6ea3d937110f24e. Source-only equivalence/open rulings above apply to reachable CLI executions, not arbitrary forged placer states. Go LSP call returned “No LSP server available for file type: .go”; source reads + targeted executions supply evidence.

### 150cbd04

Test `TestL3RetryAtTwoSeconds`; predicate L3. Exact JSON diff applied to `internal/cli/place.go` at 14994a8, then reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestL3RetryAtTwoSeconds$ ./internal/cli)`.

Run 1: rc 1.
```text
--- FAIL: TestL3RetryAtTwoSeconds (0.00s)
    place_test.go:54: next round after 1s, want 2s
FAIL
FAIL	github.com/eturkes/offload/internal/cli	0.007s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestL3RetryAtTwoSeconds (0.00s)
    place_test.go:54: next round after 1s, want 2s
FAIL
FAIL	github.com/eturkes/offload/internal/cli	0.007s
FAIL
```

Green: targeted unmutated unit-test regex (all eight added tests) rc 0 before replay; final clean rerun below.

### 69b4ce86

Test `TestL3RetryAtTwoSeconds`; predicate L3. Exact JSON diff applied to `internal/cli/place.go` at 14994a8, then reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestL3RetryAtTwoSeconds$ ./internal/cli)`.

Run 1: rc 1.
```text
--- FAIL: TestL3RetryAtTwoSeconds (0.00s)
    place_test.go:54: next round after 0s, want 2s
FAIL
FAIL	github.com/eturkes/offload/internal/cli	0.007s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestL3RetryAtTwoSeconds (0.00s)
    place_test.go:54: next round after 0s, want 2s
FAIL
FAIL	github.com/eturkes/offload/internal/cli	0.003s
FAIL
```

Green: targeted unmutated unit-test regex (all eight added tests) rc 0 before replay; final clean rerun below.

### a91ef094

Test `TestL3RetryAtTwoSeconds`; predicate L3. Exact JSON diff applied to `internal/cli/place.go` at 14994a8, then reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestL3RetryAtTwoSeconds$ ./internal/cli)`.

Run 1: rc 1.
```text
--- FAIL: TestL3RetryAtTwoSeconds (0.00s)
    place_test.go:54: next round after 3s, want 2s
FAIL
FAIL	github.com/eturkes/offload/internal/cli	0.004s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestL3RetryAtTwoSeconds (0.00s)
    place_test.go:54: next round after 3s, want 2s
FAIL
FAIL	github.com/eturkes/offload/internal/cli	0.003s
FAIL
```

Green: targeted unmutated unit-test regex (all eight added tests) rc 0 before replay; final clean rerun below.

### 236ae769

Test `TestL3DownHostCooldownHasThirtySecondMinimum` drives `placer.place` through real rounds using `testing/synctest`; a delayed diagnostic writer represents an in-domain blocked stderr pipe. Fake ssh runs use real processes/files; the fake clock advances only between completed probes. Exact diff from `agentic.json` applied at 14994a8 + reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestL3DownHostCooldownHasThirtySecondMinimum$ ./internal/cli)`.

Run 1: rc 1.
```text
--- FAIL: TestL3DownHostCooldownHasThirtySecondMinimum (0.03s)
    place_test.go:229: down host probed 2 times before the 30s deadline; want one (err=<nil>)
FAIL
FAIL	github.com/eturkes/offload/internal/cli	0.032s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestL3DownHostCooldownHasThirtySecondMinimum (0.02s)
    place_test.go:229: down host probed 2 times before the 30s deadline; want one (err=<nil>)
FAIL
FAIL	github.com/eturkes/offload/internal/cli	0.031s
FAIL
```

Green: unmutated `go test -race -count=1 -timeout=60s -run '^Test(L3DownHostCooldownHasThirtySecondMinimum|L4RecoveredCPUHostEndsGPUWait)$' ./internal/cli` rc 0; final clean rerun below.

### 2474637f

Test `TestL3WaitingLineOnceAcrossRounds`; predicate L3. Exact JSON diff applied to `internal/cli/place.go` at 14994a8, then reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestL3WaitingLineOnceAcrossRounds$ ./internal/cli)`.

Run 1: rc 1.
```text
--- FAIL: TestL3WaitingLineOnceAcrossRounds (0.00s)
    place_test.go:70: waiting across three rounds = "offload: waiting for a free slot (local 1/1 held)\noffload: waiting for a free slot (local 1/1 held)\noffload: waiting for a free slot (local 1/1 held)\n"; want one line naming local 1/1 held
FAIL
FAIL	github.com/eturkes/offload/internal/cli	0.002s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestL3WaitingLineOnceAcrossRounds (0.00s)
    place_test.go:70: waiting across three rounds = "offload: waiting for a free slot (local 1/1 held)\noffload: waiting for a free slot (local 1/1 held)\noffload: waiting for a free slot (local 1/1 held)\n"; want one line naming local 1/1 held
FAIL
FAIL	github.com/eturkes/offload/internal/cli	0.002s
FAIL
```

Green: targeted unmutated unit-test regex (all eight added tests) rc 0 before replay; final clean rerun below.

### a806646b

Test `TestL3WaitingLineOnceAcrossRounds`; predicate L3. Exact JSON diff applied to `internal/cli/place.go` at 14994a8, then reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestL3WaitingLineOnceAcrossRounds$ ./internal/cli)`.

Run 1: rc 1.
```text
--- FAIL: TestL3WaitingLineOnceAcrossRounds (0.00s)
    place_test.go:70: waiting across three rounds = "offload: waiting for a free slot (local 1/1 held)\noffload: waiting for a free slot (local 1/1 held)\noffload: waiting for a free slot (local 1/1 held)\n"; want one line naming local 1/1 held
FAIL
FAIL	github.com/eturkes/offload/internal/cli	0.003s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestL3WaitingLineOnceAcrossRounds (0.00s)
    place_test.go:70: waiting across three rounds = "offload: waiting for a free slot (local 1/1 held)\noffload: waiting for a free slot (local 1/1 held)\noffload: waiting for a free slot (local 1/1 held)\n"; want one line naming local 1/1 held
FAIL
FAIL	github.com/eturkes/offload/internal/cli	0.002s
FAIL
```

Green: targeted unmutated unit-test regex (all eight added tests) rc 0 before replay; final clean rerun below.

### f22ca4f6

Test `TestL2RoundReleasesUnselectedSlots`; predicate L2/S3. Exact JSON diff applied to `internal/cli/place.go` at 14994a8, then reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestL2RoundReleasesUnselectedSlots$ ./internal/cli)`.

Run 1: rc 1.
```text
--- FAIL: TestL2RoundReleasesUnselectedSlots (0.00s)
    place_test.go:134: unselected host still holds a slot: ["slot 1: 1788387 since 01:34:35: beta: /tmp/TestL2RoundReleasesUnselectedSlots377826907/001/project true"]
FAIL
FAIL	github.com/eturkes/offload/internal/cli	0.005s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestL2RoundReleasesUnselectedSlots (0.00s)
    place_test.go:134: unselected host still holds a slot: ["slot 1: 1788506 since 01:34:36: beta: /tmp/TestL2RoundReleasesUnselectedSlots368164966/001/project true"]
FAIL
FAIL	github.com/eturkes/offload/internal/cli	0.006s
FAIL
```

Green: targeted unmutated unit-test regex (all eight added tests) rc 0 before replay; final clean rerun below.

### 3865b1b4

Test `TestL2LaterHostAfterDownHost`; predicate L2. Exact JSON diff applied to `internal/cli/place.go` at 14994a8, then reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestL2LaterHostAfterDownHost$ ./internal/cli)`.

Run 1: rc 1.
```text
--- FAIL: TestL2LaterHostAfterDownHost (0.01s)
    --- FAIL: TestL2LaterHostAfterDownHost/cooldown=true (0.00s)
        place_test.go:171: placement="" ok=false err=<nil>, want beta
FAIL
FAIL	github.com/eturkes/offload/internal/cli	0.010s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestL2LaterHostAfterDownHost (0.01s)
    --- FAIL: TestL2LaterHostAfterDownHost/cooldown=true (0.00s)
        place_test.go:171: placement="" ok=false err=<nil>, want beta
FAIL
FAIL	github.com/eturkes/offload/internal/cli	0.010s
FAIL
```

Green: targeted unmutated unit-test regex (all eight added tests) rc 0 before replay; final clean rerun below.

### a3e8fd0c

Test `TestL2LaterHostAfterBusyWorkdir`; predicate S4. Exact JSON diff applied to `internal/cli/place.go` at 14994a8, then reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestL2LaterHostAfterBusyWorkdir$ ./internal/cli)`.

Run 1: rc 1.
```text
--- FAIL: TestL2LaterHostAfterBusyWorkdir (0.00s)
    place_test.go:151: busy workdir retained a candidate slot: ["slot 1: 1788909 since 01:34:37: alpha: /tmp/TestL2LaterHostAfterBusyWorkdir3717362097/001/project true"]
FAIL
FAIL	github.com/eturkes/offload/internal/cli	0.005s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestL2LaterHostAfterBusyWorkdir (0.00s)
    place_test.go:151: busy workdir retained a candidate slot: ["slot 1: 1789033 since 01:34:37: alpha: /tmp/TestL2LaterHostAfterBusyWorkdir4201349316/001/project true"]
FAIL
FAIL	github.com/eturkes/offload/internal/cli	0.006s
FAIL
```

Green: targeted unmutated unit-test regex (all eight added tests) rc 0 before replay; final clean rerun below.

### 3322aa31

Test `TestL2LaterHostAfterBusyWorkdir`; predicate L2/S4. Exact JSON diff applied to `internal/cli/place.go` at 14994a8, then reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestL2LaterHostAfterBusyWorkdir$ ./internal/cli)`.

Run 1: rc 1.
```text
--- FAIL: TestL2LaterHostAfterBusyWorkdir (0.00s)
    place_test.go:149: placement="" ok=false err=<nil>, want beta
FAIL
FAIL	github.com/eturkes/offload/internal/cli	0.002s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestL2LaterHostAfterBusyWorkdir (0.00s)
    place_test.go:149: placement="" ok=false err=<nil>, want beta
FAIL
FAIL	github.com/eturkes/offload/internal/cli	0.002s
FAIL
```

Green: targeted unmutated unit-test regex (all eight added tests) rc 0 before replay; final clean rerun below.

### 5db71380

Test `TestL2LaterHostAfterBusyWorkdir`; predicate L2/S4. Exact JSON diff applied to `internal/cli/place.go` at 14994a8, then reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestL2LaterHostAfterBusyWorkdir$ ./internal/cli)`.

Run 1: rc 1.
```text
--- FAIL: TestL2LaterHostAfterBusyWorkdir (0.00s)
    place_test.go:149: placement="" ok=false err=<nil>, want beta
FAIL
FAIL	github.com/eturkes/offload/internal/cli	0.002s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestL2LaterHostAfterBusyWorkdir (0.00s)
    place_test.go:149: placement="" ok=false err=<nil>, want beta
FAIL
FAIL	github.com/eturkes/offload/internal/cli	0.002s
FAIL
```

Green: targeted unmutated unit-test regex (all eight added tests) rc 0 before replay; final clean rerun below.

### cd10da7f

Test `TestL2LaterHostAfterDownHost`; predicate L2. Exact JSON diff applied to `internal/cli/place.go` at 14994a8, then reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestL2LaterHostAfterDownHost$ ./internal/cli)`.

Run 1: rc 1.
```text
--- FAIL: TestL2LaterHostAfterDownHost (0.01s)
    --- FAIL: TestL2LaterHostAfterDownHost/cooldown=false (0.00s)
        place_test.go:171: placement="" ok=false err=<nil>, want beta
FAIL
FAIL	github.com/eturkes/offload/internal/cli	0.013s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestL2LaterHostAfterDownHost (0.01s)
    --- FAIL: TestL2LaterHostAfterDownHost/cooldown=false (0.00s)
        place_test.go:171: placement="" ok=false err=<nil>, want beta
FAIL
FAIL	github.com/eturkes/offload/internal/cli	0.012s
FAIL
```

Green: targeted unmutated unit-test regex (all eight added tests) rc 0 before replay; final clean rerun below.

### 36fcaaea

Test `TestL4RecoveredCPUHostEndsGPUWait` drives `placer.place` through real rounds using `testing/synctest`; a delayed diagnostic writer represents an in-domain blocked stderr pipe. Fake ssh runs use real processes/files; the fake clock advances only between completed probes. Exact diff from `agentic.json` applied at 14994a8 + reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestL4RecoveredCPUHostEndsGPUWait$ ./internal/cli)`.

Run 1: rc 1.
```text
--- FAIL: TestL4RecoveredCPUHostEndsGPUWait (0.17s)
    place_test.go:191: all hosts now up without GPUs: err=context deadline exceeded, want no-GPU failure
FAIL
FAIL	github.com/eturkes/offload/internal/cli	0.174s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestL4RecoveredCPUHostEndsGPUWait (0.16s)
    place_test.go:191: all hosts now up without GPUs: err=context deadline exceeded, want no-GPU failure
FAIL
FAIL	github.com/eturkes/offload/internal/cli	0.160s
FAIL
```

Green: unmutated `go test -race -count=1 -timeout=60s -run '^Test(L3DownHostCooldownHasThirtySecondMinimum|L4RecoveredCPUHostEndsGPUWait)$' ./internal/cli` rc 0; final clean rerun below.

### 1016409a

Test `TestS4PlacementKeepsWorkdirLockedThroughGC`; predicate S4. Exact JSON diff applied to `internal/cli/place.go` at 14994a8, then reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestS4PlacementKeepsWorkdirLockedThroughGC$ ./internal/cli)`.

Run 1: rc 1.
```text
--- FAIL: TestS4PlacementKeepsWorkdirLockedThroughGC (0.00s)
    place_test.go:206: workdir lock became free while its placement was still live
FAIL
FAIL	github.com/eturkes/offload/internal/cli	0.006s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestS4PlacementKeepsWorkdirLockedThroughGC (0.01s)
    place_test.go:206: workdir lock became free while its placement was still live
FAIL
FAIL	github.com/eturkes/offload/internal/cli	0.012s
FAIL
```

Green: targeted unmutated unit-test regex (all eight added tests) rc 0 before replay; final clean rerun below.

### 0d100f50

Test `TestL3WaitingLineOnceAcrossRounds`; predicate L3. Exact JSON diff applied to `internal/cli/place.go` at 14994a8, then reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestL3WaitingLineOnceAcrossRounds$ ./internal/cli)`.

Run 1: rc 1.
```text
--- FAIL: TestL3WaitingLineOnceAcrossRounds (0.00s)
    place_test.go:70: waiting across three rounds = "offload: waiting for a free slot ()\n"; want one line naming local 1/1 held
FAIL
FAIL	github.com/eturkes/offload/internal/cli	0.002s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestL3WaitingLineOnceAcrossRounds (0.00s)
    place_test.go:70: waiting across three rounds = "offload: waiting for a free slot ()\n"; want one line naming local 1/1 held
FAIL
FAIL	github.com/eturkes/offload/internal/cli	0.002s
FAIL
```

Green: targeted unmutated unit-test regex (all eight added tests) rc 0 before replay; final clean rerun below.

### 7cd66ee2

Test `TestL3WaitingLineOnceAcrossRounds`; predicate L3. Exact JSON diff applied to `internal/cli/place.go` at 14994a8, then reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestL3WaitingLineOnceAcrossRounds$ ./internal/cli)`.

Run 1: rc 1.
```text
--- FAIL: TestL3WaitingLineOnceAcrossRounds (0.00s)
    place_test.go:70: waiting across three rounds = "offload: waiting for a free slot ()\n"; want one line naming local 1/1 held
FAIL
FAIL	github.com/eturkes/offload/internal/cli	0.002s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestL3WaitingLineOnceAcrossRounds (0.00s)
    place_test.go:70: waiting across three rounds = "offload: waiting for a free slot ()\n"; want one line naming local 1/1 held
FAIL
FAIL	github.com/eturkes/offload/internal/cli	0.002s
FAIL
```

Green: targeted unmutated unit-test regex (all eight added tests) rc 0 before replay; final clean rerun below.

### a84dcbb0

Test `TestS3PoolReleaseClearsOwner`; predicate S3/S4. Exact JSON diff applied to `internal/cli/place.go` at 14994a8, then reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestS3PoolReleaseClearsOwner$ ./test/e2e)`.

Run 1: rc 1.
```text
--- FAIL: TestS3PoolReleaseClearsOwner (0.16s)
    place_survivor_test.go:18: completed pool run left owner text: "1798540 since 01:35:28: alpha: /tmp/TestS3PoolReleaseClearsOwner1995909123/001/project true\n"
FAIL
FAIL	github.com/eturkes/offload/test/e2e	0.754s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestS3PoolReleaseClearsOwner (0.18s)
    place_survivor_test.go:18: completed pool run left owner text: "1798747 since 01:35:29: alpha: /tmp/TestS3PoolReleaseClearsOwner1555164852/001/project true\n"
FAIL
FAIL	github.com/eturkes/offload/test/e2e	0.707s
FAIL
```

Green: targeted unmutated e2e regex (all five added tests) rc 0 before replay; final clean rerun below.

### ea3948b7

Test `TestS4AcquisitionErrorClearsEarlierCandidates`; predicate S3/S4. Exact JSON diff applied to `internal/cli/place.go` at 14994a8, then reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestS4AcquisitionErrorClearsEarlierCandidates$ ./test/e2e)`.

Run 1: rc 1.
```text
--- FAIL: TestS4AcquisitionErrorClearsEarlierCandidates (0.01s)
    --- FAIL: TestS4AcquisitionErrorClearsEarlierCandidates/slot (0.00s)
        place_survivor_test.go:45: acquisition failure left alpha owner text: "1798963 since 01:35:30: alpha: /tmp/TestS4AcquisitionErrorClearsEarlierCandidatesslot1008188195/001/project touch /tmp/TestS4AcquisitionErrorClearsEarlierCandidatesslot1008188195/001/must-not-run\n"
FAIL
FAIL	github.com/eturkes/offload/test/e2e	0.566s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestS4AcquisitionErrorClearsEarlierCandidates (0.01s)
    --- FAIL: TestS4AcquisitionErrorClearsEarlierCandidates/slot (0.00s)
        place_survivor_test.go:45: acquisition failure left alpha owner text: "1799133 since 01:35:31: alpha: /tmp/TestS4AcquisitionErrorClearsEarlierCandidatesslot1259708502/001/project touch /tmp/TestS4AcquisitionErrorClearsEarlierCandidatesslot1259708502/001/must-not-run\n"
FAIL
FAIL	github.com/eturkes/offload/test/e2e	0.434s
FAIL
```

Green: targeted unmutated e2e regex (all five added tests) rc 0 before replay; final clean rerun below.

### 3bd2407a

Test `TestS4AcquisitionErrorClearsEarlierCandidates`; predicate S2/C3. Exact JSON diff applied to `internal/cli/place.go` at 14994a8, then reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestS4AcquisitionErrorClearsEarlierCandidates$ ./test/e2e)`.

Run 1: rc 1.
```text
--- FAIL: TestS4AcquisitionErrorClearsEarlierCandidates (0.22s)
    --- FAIL: TestS4AcquisitionErrorClearsEarlierCandidates/workdir (0.21s)
        place_survivor_test.go:36: exit = 0, want 125; stdout="" stderr="offload: alpha slot 1/2 (8 CPUs, 0% busy) → ~/offload/project-5a03879b\n"
FAIL
FAIL	github.com/eturkes/offload/test/e2e	0.708s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestS4AcquisitionErrorClearsEarlierCandidates (0.23s)
    --- FAIL: TestS4AcquisitionErrorClearsEarlierCandidates/workdir (0.23s)
        place_survivor_test.go:36: exit = 0, want 125; stdout="" stderr="offload: alpha slot 1/2 (8 CPUs, 0% busy) → ~/offload/project-257f094e\n"
FAIL
FAIL	github.com/eturkes/offload/test/e2e	0.639s
FAIL
```

Green: targeted unmutated e2e regex (all five added tests) rc 0 before replay; final clean rerun below.

### 19abdc33

Test `TestS4AcquisitionErrorClearsEarlierCandidates`; predicate S3/S4. Exact JSON diff applied to `internal/cli/place.go` at 14994a8, then reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestS4AcquisitionErrorClearsEarlierCandidates$ ./test/e2e)`.

Run 1: rc 1.
```text
--- FAIL: TestS4AcquisitionErrorClearsEarlierCandidates (0.01s)
    --- FAIL: TestS4AcquisitionErrorClearsEarlierCandidates/workdir (0.00s)
        place_survivor_test.go:45: acquisition failure left alpha owner text: "1799802 since 01:35:34: alpha: /tmp/TestS4AcquisitionErrorClearsEarlierCandidatesworkdir3826370715/001/project touch /tmp/TestS4AcquisitionErrorClearsEarlierCandidatesworkdir3826370715/001/must-not-run\n"
FAIL
FAIL	github.com/eturkes/offload/test/e2e	0.594s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestS4AcquisitionErrorClearsEarlierCandidates (0.01s)
    --- FAIL: TestS4AcquisitionErrorClearsEarlierCandidates/workdir (0.00s)
        place_survivor_test.go:45: acquisition failure left alpha owner text: "1799967 since 01:35:34: alpha: /tmp/TestS4AcquisitionErrorClearsEarlierCandidatesworkdir190160481/001/project touch /tmp/TestS4AcquisitionErrorClearsEarlierCandidatesworkdir190160481/001/must-not-run\n"
FAIL
FAIL	github.com/eturkes/offload/test/e2e	0.444s
FAIL
```

Green: targeted unmutated e2e regex (all five added tests) rc 0 before replay; final clean rerun below.

### 3afd948b

Test `TestS4AcquisitionErrorClearsEarlierCandidates`; predicate S2/C3. Exact JSON diff applied to `internal/cli/place.go` at 14994a8, then reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestS4AcquisitionErrorClearsEarlierCandidates$ ./test/e2e)`.

Run 1: rc 1.
```text
--- FAIL: TestS4AcquisitionErrorClearsEarlierCandidates (0.22s)
    --- FAIL: TestS4AcquisitionErrorClearsEarlierCandidates/workdir (0.21s)
        place_survivor_test.go:36: exit = 0, want 125; stdout="" stderr="offload: alpha slot 1/2 (8 CPUs, 0% busy) → ~/offload/project-f1c704af\n"
FAIL
FAIL	github.com/eturkes/offload/test/e2e	0.711s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestS4AcquisitionErrorClearsEarlierCandidates (0.21s)
    --- FAIL: TestS4AcquisitionErrorClearsEarlierCandidates/workdir (0.21s)
        place_survivor_test.go:36: exit = 0, want 125; stdout="" stderr="offload: alpha slot 1/2 (8 CPUs, 0% busy) → ~/offload/project-d259dc8f\n"
FAIL
FAIL	github.com/eturkes/offload/test/e2e	0.620s
FAIL
```

Green: targeted unmutated e2e regex (all five added tests) rc 0 before replay; final clean rerun below.

### 209c0e7d

Test `TestS4AcquisitionErrorClearsEarlierCandidates`; predicate S2/C3. Exact JSON diff applied to `internal/cli/place.go` at 14994a8, then reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestS4AcquisitionErrorClearsEarlierCandidates$ ./test/e2e)`.

Run 1: rc 1.
```text
--- FAIL: TestS4AcquisitionErrorClearsEarlierCandidates (0.09s)
    --- FAIL: TestS4AcquisitionErrorClearsEarlierCandidates/workdir (0.09s)
        place_survivor_test.go:36: exit = 0, want 125; stdout="" stderr="offload: local slot 1/2\n"
FAIL
FAIL	github.com/eturkes/offload/test/e2e	0.609s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestS4AcquisitionErrorClearsEarlierCandidates (0.12s)
    --- FAIL: TestS4AcquisitionErrorClearsEarlierCandidates/workdir (0.12s)
        place_survivor_test.go:36: exit = 0, want 125; stdout="" stderr="offload: local slot 1/2\n"
FAIL
FAIL	github.com/eturkes/offload/test/e2e	0.604s
FAIL
```

Green: targeted unmutated e2e regex (all five added tests) rc 0 before replay; final clean rerun below.

### b57b381e

Test `TestL3WaitingNamesEveryPlace`; predicate R1/L3. Exact JSON diff applied to `internal/cli/place.go` at 14994a8, then reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestL3WaitingNamesEveryPlace$ ./test/e2e)`.

Run 1: rc 1.
```text
--- FAIL: TestL3WaitingNamesEveryPlace (0.02s)
    place_survivor_test.go:64: "offload: waiting for a free slot (local 2/2 held; alpha 0/2 held; beta 2/2 held)\n" does not contain "alpha down:"
    place_survivor_test.go:64: "offload: waiting for a free slot (local 2/2 held; alpha 0/2 held; beta 2/2 held)\n" does not contain "Connection refused"
FAIL
FAIL	github.com/eturkes/offload/test/e2e	0.660s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestL3WaitingNamesEveryPlace (0.05s)
    place_survivor_test.go:64: "offload: waiting for a free slot (local 2/2 held; alpha 0/2 held; beta 2/2 held)\n" does not contain "alpha down:"
    place_survivor_test.go:64: "offload: waiting for a free slot (local 2/2 held; alpha 0/2 held; beta 2/2 held)\n" does not contain "Connection refused"
FAIL
FAIL	github.com/eturkes/offload/test/e2e	1.894s
FAIL
```

Green: targeted unmutated e2e regex (all five added tests) rc 0 before replay; final clean rerun below.

### 67232af2

Test `TestL3WaitingNamesEveryPlace`; predicate R1/L3. Exact JSON diff applied to `internal/cli/place.go` at 14994a8, then reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestL3WaitingNamesEveryPlace$ ./test/e2e)`.

Run 1: rc 1.
```text
--- FAIL: TestL3WaitingNamesEveryPlace (0.05s)
    place_survivor_test.go:64: "offload: waiting for a free slot (local 2/2 held; alpha 0/2 held; beta 2/2 held)\n" does not contain "alpha down:"
    place_survivor_test.go:64: "offload: waiting for a free slot (local 2/2 held; alpha 0/2 held; beta 2/2 held)\n" does not contain "Connection refused"
FAIL
FAIL	github.com/eturkes/offload/test/e2e	2.197s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestL3WaitingNamesEveryPlace (0.05s)
    place_survivor_test.go:64: "offload: waiting for a free slot (local 2/2 held; alpha 0/2 held; beta 2/2 held)\n" does not contain "alpha down:"
    place_survivor_test.go:64: "offload: waiting for a free slot (local 2/2 held; alpha 0/2 held; beta 2/2 held)\n" does not contain "Connection refused"
FAIL
FAIL	github.com/eturkes/offload/test/e2e	1.719s
FAIL
```

Green: targeted unmutated e2e regex (all five added tests) rc 0 before replay; final clean rerun below.

### ae331179

Test `TestL3WaitingNamesEveryPlace`; predicate L3. Exact JSON diff applied to `internal/cli/place.go` at 14994a8, then reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestL3WaitingNamesEveryPlace$ ./test/e2e)`.

Run 1: rc 1.
```text
--- FAIL: TestL3WaitingNamesEveryPlace (0.03s)
    place_survivor_test.go:64: "offload: waiting for a free slot (local 2/2 held; beta 2/2 held)\n" does not contain "alpha down:"
    place_survivor_test.go:64: "offload: waiting for a free slot (local 2/2 held; beta 2/2 held)\n" does not contain "Connection refused"
FAIL
FAIL	github.com/eturkes/offload/test/e2e	2.046s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestL3WaitingNamesEveryPlace (0.02s)
    place_survivor_test.go:64: "offload: waiting for a free slot (local 2/2 held; beta 2/2 held)\n" does not contain "alpha down:"
    place_survivor_test.go:64: "offload: waiting for a free slot (local 2/2 held; beta 2/2 held)\n" does not contain "Connection refused"
FAIL
FAIL	github.com/eturkes/offload/test/e2e	1.439s
FAIL
```

Green: targeted unmutated e2e regex (all five added tests) rc 0 before replay; final clean rerun below.

### 07423c5b

Test `TestL3WaitingNamesEveryPlace`; predicate L3. Exact JSON diff applied to `internal/cli/place.go` at 14994a8, then reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestL3WaitingNamesEveryPlace$ ./test/e2e)`.

Run 1: rc 1.
```text
--- FAIL: TestL3WaitingNamesEveryPlace (0.02s)
    place_survivor_test.go:64: "offload: waiting for a free slot (local 2/2 held; alpha down: ssh: connect to host alpha port 22: Connection refused)\n" does not contain "beta 2/2 held"
FAIL
FAIL	github.com/eturkes/offload/test/e2e	1.014s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestL3WaitingNamesEveryPlace (0.02s)
    place_survivor_test.go:64: "offload: waiting for a free slot (local 2/2 held; alpha down: ssh: connect to host alpha port 22: Connection refused)\n" does not contain "beta 2/2 held"
FAIL
FAIL	github.com/eturkes/offload/test/e2e	0.811s
FAIL
```

Green: targeted unmutated e2e regex (all five added tests) rc 0 before replay; final clean rerun below.

### be934352

Test `TestS1StatusRejectsInvalidHeavySlots`; predicate S1/C3. Exact JSON diff applied to `internal/cli/place.go` at 14994a8, then reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestS1StatusRejectsInvalidHeavySlots$ ./test/e2e)`.

Run 1: rc 1.
```text
--- FAIL: TestS1StatusRejectsInvalidHeavySlots (0.00s)
    place_survivor_test.go:89: exit = 0, want 2; stdout="local heavy slots: 0/0 held\n\nHOST           STATE  SLOTS  CPUS   LOAD  BUSY  FREE MEM GPUS\nalpha          up       0/2     8   2.00   25%     1.0GB    0\n" stderr=""
FAIL
FAIL	github.com/eturkes/offload/test/e2e	0.784s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestS1StatusRejectsInvalidHeavySlots (0.00s)
    place_survivor_test.go:89: exit = 0, want 2; stdout="local heavy slots: 0/0 held\n\nHOST           STATE  SLOTS  CPUS   LOAD  BUSY  FREE MEM GPUS\nalpha          up       0/2     8   2.00   25%     1.0GB    0\n" stderr=""
FAIL
FAIL	github.com/eturkes/offload/test/e2e	0.467s
FAIL
```

Green: targeted unmutated e2e regex (all five added tests) rc 0 before replay; final clean rerun below.

### 369df7a7

Test `TestS1StatusRejectsInvalidHeavySlots`; predicate S1/C3. Exact JSON diff applied to `internal/cli/place.go` at 14994a8, then reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestS1StatusRejectsInvalidHeavySlots$ ./test/e2e)`.

Run 1: rc 1.
```text
--- FAIL: TestS1StatusRejectsInvalidHeavySlots (0.00s)
    place_survivor_test.go:89: exit = 0, want 2; stdout="local heavy slots: 0/0 held\n\nHOST           STATE  SLOTS  CPUS   LOAD  BUSY  FREE MEM GPUS\nalpha          up       0/2     8   2.00   25%     1.0GB    0\n" stderr=""
FAIL
FAIL	github.com/eturkes/offload/test/e2e	0.473s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestS1StatusRejectsInvalidHeavySlots (0.00s)
    place_survivor_test.go:89: exit = 0, want 2; stdout="local heavy slots: 0/0 held\n\nHOST           STATE  SLOTS  CPUS   LOAD  BUSY  FREE MEM GPUS\nalpha          up       0/2     8   2.00   25%     1.0GB    0\n" stderr=""
FAIL
FAIL	github.com/eturkes/offload/test/e2e	0.358s
FAIL
```

Green: targeted unmutated e2e regex (all five added tests) rc 0 before replay; final clean rerun below.

### bf895ab1

Test `TestS1StatusRejectsInvalidHeavySlots`; predicate S1/C4. Exact JSON diff applied to `internal/cli/place.go` at 14994a8, then reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestS1StatusRejectsInvalidHeavySlots$ ./test/e2e)`.

Run 1: rc 1.
```text
--- FAIL: TestS1StatusRejectsInvalidHeavySlots (0.00s)
    place_survivor_test.go:90: "" does not contain "offload: "
    place_survivor_test.go:90: "" does not contain "HEAVY_SLOTS"
FAIL
FAIL	github.com/eturkes/offload/test/e2e	0.424s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestS1StatusRejectsInvalidHeavySlots (0.00s)
    place_survivor_test.go:90: "" does not contain "offload: "
    place_survivor_test.go:90: "" does not contain "HEAVY_SLOTS"
FAIL
FAIL	github.com/eturkes/offload/test/e2e	0.402s
FAIL
```

Green: targeted unmutated e2e regex (all five added tests) rc 0 before replay; final clean rerun below.

### 11773c4f

Test `TestS1StatusRejectsInvalidHeavySlots`; predicate S1/C3. Exact JSON diff applied to `internal/cli/place.go` at 14994a8, then reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestS1StatusRejectsInvalidHeavySlots$ ./test/e2e)`.

Run 1: rc 1.
```text
--- FAIL: TestS1StatusRejectsInvalidHeavySlots (0.00s)
    place_survivor_test.go:89: exit = 0, want 2; stdout="" stderr="offload: HEAVY_SLOTS must be a positive integer\n"
FAIL
FAIL	github.com/eturkes/offload/test/e2e	0.426s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestS1StatusRejectsInvalidHeavySlots (0.00s)
    place_survivor_test.go:89: exit = 0, want 2; stdout="" stderr="offload: HEAVY_SLOTS must be a positive integer\n"
FAIL
FAIL	github.com/eturkes/offload/test/e2e	0.366s
FAIL
```

Green: targeted unmutated e2e regex (all five added tests) rc 0 before replay; final clean rerun below.

### 241a4e88

Test `TestR2StatusContinuesAfterDownHost`; predicate R2. Exact JSON diff applied to `internal/cli/place.go` at 14994a8, then reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestR2StatusContinuesAfterDownHost$ ./test/e2e)`.

Run 1: rc 1.
```text
--- FAIL: TestR2StatusContinuesAfterDownHost (0.01s)
    place_survivor_test.go:80: status rows=["alpha down 0/2 ssh: connect to host alpha port 22: Connection refused"], want down alpha followed by beta at 150% busy
FAIL
FAIL	github.com/eturkes/offload/test/e2e	0.512s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestR2StatusContinuesAfterDownHost (0.01s)
    place_survivor_test.go:80: status rows=["alpha down 0/2 ssh: connect to host alpha port 22: Connection refused"], want down alpha followed by beta at 150% busy
FAIL
FAIL	github.com/eturkes/offload/test/e2e	0.462s
FAIL
```

Green: targeted unmutated e2e regex (all five added tests) rc 0 before replay; final clean rerun below.

### 37b44696

Test `TestR2StatusContinuesAfterDownHost`; predicate R2. Exact JSON diff applied to `internal/cli/place.go` at 14994a8, then reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestR2StatusContinuesAfterDownHost$ ./test/e2e)`.

Run 1: rc 1.
```text
--- FAIL: TestR2StatusContinuesAfterDownHost (0.01s)
    place_survivor_test.go:80: status rows=["alpha down 0/2 ssh: connect to host alpha port 22: Connection refused" "beta up 0/3 8 12.00 152% 1.0GB 1"], want down alpha followed by beta at 150% busy
FAIL
FAIL	github.com/eturkes/offload/test/e2e	0.567s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestR2StatusContinuesAfterDownHost (0.01s)
    place_survivor_test.go:80: status rows=["alpha down 0/2 ssh: connect to host alpha port 22: Connection refused" "beta up 0/3 8 12.00 152% 1.0GB 1"], want down alpha followed by beta at 150% busy
FAIL
FAIL	github.com/eturkes/offload/test/e2e	0.432s
FAIL
```

Green: targeted unmutated e2e regex (all five added tests) rc 0 before replay; final clean rerun below.

### f48dfbe3

Test `TestR2StatusContinuesAfterDownHost`; predicate R2. Exact JSON diff applied to `internal/cli/place.go` at 14994a8, then reverted. Red command (twice): `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -count=1 -timeout=60s -run ^TestR2StatusContinuesAfterDownHost$ ./test/e2e)`.

Run 1: rc 1.
```text
--- FAIL: TestR2StatusContinuesAfterDownHost (0.01s)
    place_survivor_test.go:80: status rows=["alpha down 0/2 ssh: connect to host alpha port 22: Connection refused" "beta up 0/3 8 12.00 148% 1.0GB 1"], want down alpha followed by beta at 150% busy
FAIL
FAIL	github.com/eturkes/offload/test/e2e	0.560s
FAIL
```

Run 2: rc 1.
```text
--- FAIL: TestR2StatusContinuesAfterDownHost (0.01s)
    place_survivor_test.go:80: status rows=["alpha down 0/2 ssh: connect to host alpha port 22: Connection refused" "beta up 0/3 8 12.00 148% 1.0GB 1"], want down alpha followed by beta at 150% busy
FAIL
FAIL	github.com/eturkes/offload/test/e2e	0.461s
FAIL
```

Green: targeted unmutated e2e regex (all five added tests) rc 0 before replay; final clean rerun below.

## Equivalence/open evidence

- `16498274`, `92b8b879`, `7f7f476f`, `876bd4fb`: `internal/cli/place.go:141–178` returns any selected candidate before evaluating no-GPU failure. Without `--gpu`, every successful candidate is eligible. With `--gpu`, every positive inventory is eligible. Thus continued waits never carry positive GPU inventory, and a continued non-GPU wait never knows every host up. `internal/cli/cli.go:193–231` rejects absent/invalid/unknown-host pool-only configurations and remote-ineligible trees rather than passing an empty host set in GPU mode.
- `6bcd43c0`, `090c7982`: `internal/cli/cli.go:262` exact anchor `case ctx.Err() != nil:` precedes the returned-error branch and `defer pl.release()`; cancellation fixes the exit independently of the mutated error result.
- `7629ba80`: `internal/slots/slots.go:260–265`, `TryLock` forwards `tryLock`; every error return in `tryLock:209–225` is `return nil, false, …`. Removing the explicit error operand cannot change entry into the `!held` branch.
- `f58073c7`: cancelled probes reach `place.go:69–71` → `cli.go:262–263`; altered maps are never read again.
- `70181c0d`, `801c4c06`: `place.go:145–153` installs `downWhy` + deletes `known` on failure, clears `downWhy` + replaces `known` on success. The surviving test in `noGPUAnywhere:174` already covers each down host; cancellation ends this placer.
- `ff9eb11d`: contract `L3` exact anchor `a down host is re-probed at most every 30 s` constrains maximum frequency, not latest retry. 31 s meets it.
- `11bdf2f2`, `fafc0e4e`: `cli.go:267` defers release until `runLogged` returns; `runCommand:174–183` then writes the end record before process exit. The variants can retain a workdir fd during that post-command interval. S4/X3 require exclusion through the run and release on exit, not a tighter pre-exit close deadline. Verified with `go doc -src os.File.Close`: a nil receiver returns `ErrInvalid` rather than panicking.
- `6a9244a3`: the timer still wakes in ≤2 s; the next round’s cancelled probes release candidates, then `place.go:69–71` exits. L3 fixes 130/143/129 + nothing held, not a sub-retry response bound.
- `9db15d80`: R2 fixes local slots + host fields/order, not the table header. Every data row is unchanged.

## Final verification

- Final clean tests: `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-11 && umask 022 && go test -race -count=1 -shuffle=on -timeout=90s ./internal/cli)` → rc 0, package 1.414 s.
- Final clean CLI cases: same flags + `-run '^Test(S3PoolReleaseClearsOwner|S4AcquisitionErrorClearsEarlierCandidates|L3WaitingNamesEveryPlace|R2StatusContinuesAfterDownHost|S1StatusRejectsInvalidHeavySlots)$' ./test/e2e` → rc 0, package 3.057 s. TestMain builds/drives the real CLI; fake SSH hosts remain isolated in test temporary directories.
- Format: `golangci-lint fmt --diff internal/cli/place_test.go test/e2e/place_survivor_test.go` → rc 0.
- Lint: `golangci-lint run ./internal/cli/... ./test/e2e/...` → rc 0, `0 issues.`
- `.agent/law` + `.agent/spec-lint` → rc 0.
- All 33 new-test rows = exact default-operator diff, two assertion failures, unmutated green. For `236ae769` + `36fcaaea`, final tests drive the complete placement loop; their notes supersede the initial direct-round experiments.
- Coverage limit: full gate + full e2e suite not run (brief requires narrow cases while other reviewers share the controller); no tests skipped in the executed selections. LSP unavailable as recorded above. No unconfirmed behavior claim remains.
- Only added test files committed; product source restored after every mutation. Notify-set edits = none.

Commit on `wt/reviewer-11`: `cc8ba573604751e211465e8b3127db5e9b4062c8`.

Counts: new test 33; equivalent 10; contract-open 5; killed 0; n/a 0; unknown 0.

reviewer-11-DONE-1
