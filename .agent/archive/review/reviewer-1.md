# reviewer-1 — correctness/spec — c32846f..47a8417

Fixed A1–A8 before target reads. Target = `47a8417b2bf241239d17824f45d285600ebece90`; tests-only worktree = `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-1`. **8/8 adjudicated; 11 findings = 8 major + 3 minor.** Production unchanged. Tests committed on `wt/reviewer-1`: `df46bd66671ca4649221d512404c253173e3c56c`; worktree clean.

| id | check | verdict | evidence |
|----|-------|---------|----------|
| A1 | C1–C5 CLI + exit precedence | finding F3 | SIGINT during Inspect →125, not130; flags/output/placement adjacent controls pass |
| A2 | P1–P3 pool + R1–R2 probe/status | finding F1 F2 F11 | explicit slots=0 accepted; missing lock dirs count held; probe exceeds8s |
| A3 | S1–S5 slots + heavy interop/lifetime | finding F8 | remote owner omits host; S1/S2/S3 selected controls pass; watcher lifetime also F10 |
| A4 | L1–L4 placement, release, GPU, wait/signals | finding F7 | stale GPU inventory turns a currently-down host into definitive no-GPU; L1/L2/L3 controls pass |
| A5 | Y1–Y11 U/R/D, prune, rsync/lock, get/nosync/plain | finding F4 F5 F9 | down-list/prune failures swallowed; directory get nests basename; selected framing/symlink/lock controls pass |
| A6 | X1–X7 watcher, EOF, status/nonce, SSH options | finding F6 F10 | status SSH ignores cancel; death before setsid leaves live orphan; options/call sites read |
| A7 | spec Decisions implemented/deferred | finding F4 F5 F10 | promised output return/session death violated; Go/hybrid/slots/root/git/serialization decisions traced; machine-only Decisions excluded, below |
| A8 | Go/bash pitfalls: fd/goroutine, errors, quoting, signals, writes | finding F4 F6 F9 F10 F11 | unchecked shell failures + unbounded subprocess waits + session-start race; all cmd/internal production functions + callers/callees read |

## Reproduction

- E2E reds: `go -C /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-1 test -count=1 ./test/e2e -run '^TestReview' -timeout 60s`.
- Probe red: `go -C /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-1 test -count=1 ./internal/probe -run '^TestReviewR1ProbeDeadlineWithInheritedOutput$' -timeout 20s`.
- Each named test below observed red on the target; rc1. Added files: `test/e2e/reviewer1_test.go`, `internal/probe/reviewer1_test.go`. Tests use private fake SSH, HOME/runtime/config dirs; no real-host SSH.

## F1 — minor — explicit zero pool slots defaults to two

`internal/pool/pool.go:90–93`: only negative slots rejected; absent slots + explicit `slots = 0` share Go's zero value. P2 requires int≥1; C3 invalid pool →2 before work. Zero instead admits a command.

Red `TestReviewP2ExplicitZeroSlots`: `exit = 0, want 2; stdout="" stderr="offload: local slot 1/2\n"`.

## F2 — minor — fresh status falsely reports every slot held

`internal/slots/slots.go:108–120`: `Holders` opens slot paths without creating missing parent dirs; every open error falls through to a held-slot row. R2 says held iff locked. Fresh controller/host dirs report full while no lock exists.

Red `TestReviewR2FreshStatusHasNoHolders`: local `local heavy slots: 2/2 held`; alpha `up       2/2`; expected both `0/2`.

## F3 — major — signal precedence lost during repo inspection

`internal/cli/cli.go:161–162`: cancellation kills `git rev-parse`; early error returns `fail` before checking the signal. C3 requires SIGINT →130 regardless of an accompanying offload error. UploadList preflight has the same early-return pattern.

Red `TestReviewC3SignalDuringInspection`: fake git marks entry then sleeps; SIGINT → `exit = 125, want 130; … git rev-parse …`.

## F4 — major — remote output-list failure reports success

`internal/run/remote.bash:62–68`: unchecked `git_list | while …` without pipefail produces empty `down`; success status still recorded; empty rsync succeeds. Y6/Y10 require inventory + failed download →125. A command producing output then corrupting `.git/index` returns success while result stays remote.

Red `TestReviewY6DownloadListFailure`: `exit = 0, want 125; … fatal: .git/index: index file smaller than expected`. Positive control: remote result exists before status assertion.

## F5 — major — --get directory duplicates its basename

`internal/mirror/mirror.go:295–298`: source `…/output` + destination `<root>/output` feeds rsync's copy-inside-existing-directory semantics. Y8 says PATH copied remote→local; repeat retrieval creates `<root>/output/output/result` instead of updating `<root>/output/result`.

Red `TestReviewY8GetExistingDirectory`: `--get output must return output/result at its original path; nested=true`; forbidden `output/output` exists. Existing trailing-slash case remains distinct.

## F6 — major — status-recovery SSH cannot be interrupted

`internal/run/run.go:82`: raw `exec.Command(...).Output()` uses neither ctx nor bounded cancellation. After125/255 a stalled status read ignores HUP/INT/TERM. X3/C3 require teardown + signal exit; CMD has already ended.

Red `TestReviewX3SignalDuringStatusRecovery`: fake SSH stalls only `cat …/status`; SIGINT → `offload timed out after 3s`, expected130. Cleanup kills offload + its own process group.

## F7 — major — stale GPU inventory ignores a host's current down state

`internal/cli/place.go:139–170`: probe failure marks down without invalidating `known`; `noGPUAnywhere` tests historical successes alone. Round1: alpha up/noGPU, beta busy →wait. Round2: alpha down, beta up/noGPU →125. L4 says a down eligible host keeps waiting.

Red `TestReviewL4PreviouslyUpHostGoesDown`: `down alpha must keep waiting; stale no-GPU inventory exited 125: "offload: waiting …\noffload: no pool host has an NVIDIA GPU\n"`.

## F8 — minor — remote slot owner omits host

`internal/cli/cli.go:310–312`, caller `internal/cli/place.go:111`: formatter accepts root+argv alone and emits ` <root>: <CMD>`. S4 requires ` <host>: <root> <CMD>`. Host attribution disappears from owner text.

Red `TestReviewS4PoolOwnerNamesHost`: `owner must name host then root; got "<pid> since <time>: /tmp/…/project: sleep 120\n"`.

## F9 — major — failed prune still launches command

`internal/run/remote.bash:27–28,55`: failed `rm -f` ignored by unchecked prune loop; command starts on stale input. Y10 requires failed prune →125 before CMD.

Red `TestReviewY10FailedPrunePreventsCommand`: `exit = 0, want 125; … rm: injected prune failure`. Positive control checks injected diagnostic; stale remote path seeded by a successful preceding run.

## F10 — major — controller death before setsid leaves a live command

`internal/run/watch.bash:14–35`: EOF can win while child is scheduled before `setsid()`. Both session checks find no session; watcher then `wait`s for the child without another kill. Child subsequently creates its session and runs forever, keeping the local heavy slot too. X1 + spec's any-signal death guarantee fail.

Red `TestReviewX1DeathBeforeSessionCreation`: delays setsid until watcher completes both post-EOF session checks (pgrep wrapper preserves actual results), kills controller, releases child. `session process <pid> remains after 13s (kill(0)=<nil>)`. This models a legal scheduling pause before setsid, not altered session semantics. All marked PIDs cleaned.

## F11 — major — probe deadline does not bound inherited output pipes

`internal/probe/probe.go:58–68`: CommandContext kills SSH at8s, but `Output` waits indefinitely for stdout/stderr EOF from descendants (`WaitDelay` unset). An SSH helper retaining those descriptors blocks probe/status/placement past R1's end-to-end bound.

Red `TestReviewR1ProbeDeadlineWithInheritedOutput`: fake SSH forks a sleeper with inherited output; `Probe still blocked after 9s; killed ssh's descendant holds stdout/stderr open`. Cleanup kills marked child + joins Probe.

## Passing controls / limits

- `go -C /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-1 test -count=1 ./test/e2e -run '^(TestC1FlagBoundary|TestP3MissingPool|TestL1Eligibility|TestL2Placement|TestL3WaitingSignals|TestY5PruneSymlinkAncestor|TestY11LivePredecessorAfterControllerDeath|TestX6Remote255AndTransport)$' -timeout 60s` →rc0 (6.511s).
- `go -C /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-1 test -count=1 ./internal/... -run '^(TestPathPrecedence|TestLoadDefaultsSlots|TestLoadRejects|TestParse|TestHeavySlotsAgreesWithHeavy|TestSlotHeldByOffloadIsBusyToHeavy|TestLowestFreeSlotFirst|TestCommandNeverInheritsLock|TestUploadList|TestInspect|TestGetPath|TestJoinRoundTripsThroughBash|TestLockReachesWatcherNotCommand|TestRemoteFramingMismatchPrunesNothing)$' -timeout 60s` →rc0. Six packages ran matching tests; cli had no match (not counted).
- Full gate/race suite/mutation suite not run by reviewer; selected controls + red reproductions only. Production read coverage = all cmd/internal Go + embedded bash, full enclosing functions, caller/callee paths, adjacent tests.
- Go LSP unavailable (`documentSymbol` on real `.go` → no server); direct reads/search instead. Toolchain go1.27.3.
- A7 machine-only Decisions (idle trading/Codexify deployment, controller deletion, real SSH firewall/accounts, encrypted boot, compute posture) = outside cmd/internal target; **unconfirmed** here. Read `.agent/spec.md`; no deployment/live SSH attempted. Static X5 client options inspected; sshd/live timeouts unconfirmed.
- Added-test formatting (`golangci-lint fmt --diff` on both files) + lint (`golangci-lint run ./test/e2e ./internal/probe`) →rc0, `0 issues`. Initial lint found one unchecked channel error; fixed before the final pass.
- Final replay `go -C /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-1 test -count=1 ./test/e2e ./internal/probe -run '^TestReview' -timeout 60s` →rc1, all11 red (e2e19.909s; probe9.006s). `git diff --cached --check` →rc0; commit contains only the two test files.
- Validator: eight expected A-rows present; unresolved-verdict count →0 (grep rc1). No unresolved check rows.

## Register (outside acceptance / scope)

- `signals()` leaves Notify subscription + blocked goroutine if `cli.Main` is repeatedly embedded; shipped CLI calls it once then exits. No merge finding against current executable contract.
- One read-only real `heavy --status` used for scheduler visibility despite brief's stricter private-runtime boundary; no real-slot acquisition/mutation. Every test used sandbox runtime/config/HOME.
