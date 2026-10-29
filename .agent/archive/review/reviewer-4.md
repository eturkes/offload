# reviewer-4 — verification integrity

Target = `fed6498..cac258b87885d725fbbf45d78afa10e1f9a1cd60`; private branch `wt/reviewer-4`. V1–V7 fixed before target inspection. Acceptance = `.agent/contract.md` + `.claude/rules/gate*.md` + verification-integrity law. Production plants reverted after each experiment. No real SSH, real pool config, or real heavy slots.

| id | check | verdict | evidence |
|----|-------|---------|----------|
| V1 | each test asserts what its name/comment claims (no tautologies, no assert-free paths) | finding | F02, F03, F07: X6 negative path stops during upload; R2 metrics can become zero; remote-lock unit test times its own fixture |
| V2 | e2e fake ssh fidelity vs real OpenSSH (argv joining, stdin EOF, exit codes, env) — gaps that let a real-ssh bug pass | finding | F01: shim discards `-n`; production `SSHArgs += -n` survives full suite despite OpenSSH `stdinnull yes` |
| V3 | skipped / xfailed / conditionally disabled cases: none, or each justified | pass | no Skip/Short/build tags/Parallel; only `SLOTS_TEST_HOLD` helper subprocess branch (`internal/slots/slots_test.go:181–198`) is intentional + exercised |
| V4 | red-first evidence in commit bodies plausible + each named test can actually fail on the described bug | pass | archive/tester-1 = exactly 36 failing functions / 100 leaves; stdin, INT, nested preflight, framing, broken repo, finite load, address validation regressions all reproduced red; delayed-pgrep old grace loop red/current loop green |
| V5 | gate covers every package + script; nothing excluded silently (go test ./..., shellcheck list, nixfmt list) | pass | `.agent/gate:33–44`: race/shuffle/count=1 `./...`; 9 packages, all 10 tracked sh/bash + embedded scripts, all 7 Nix files; Go pins agree; test-only gosec exception explicit |
| V6 | flakiness risk: timing bounds vs CI runners (2-core), parallel tests sharing dirs/ports | pass | private TempDirs/no shared ports/no t.Parallel; clean `taskset -c 0,1 go test -race -shuffle=on -count=2 ./...` rc 0 (e2e 60.397 s); bounded sample, not a CI-host proof |
| V7 | contract predicates without any test (list them) | finding | F04, F05, F06: R1/L3/S4/X5 survivors; explicit zero slots accepted; failed output listing silently preserves command status |

## Experiment protocol

`W=/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-4`; each mutant starts at `cac258b`, changes one production mechanism, runs the same command, restores original bytes:

```sh
cd "$W"
env GOCACHE="$W/.scratch/review/cache" XDG_RUNTIME_DIR="$W/.scratch/review/runtime" \
  heavy go test -race -shuffle=on -count=1 ./...
```

Baseline rc 0; e2e 39.711 s. Mutants M1–M7 each rc 0, all 8 tested packages `ok`, entry-point package `[no test files]`. Logs + executable patch driver: `$W/.scratch/review/{mutations.py,mutations.log,M*.log}`. This is a seven-case targeted mutation audit, not exhaustive default-operator mutation coverage.

## F01 — major — fake SSH suppresses an stdin-breaking regression

- Anchor: `test/e2e/testdata/ssh:8`, exact bytes `-o*|-l*|-p*|-i*|-F*|-J*|-S*|-T|-t|-tt|-n|-q|-v|-vv|-vvv|-x|-C) shift ;;`.
- Why: `-n` means remote stdin `/dev/null`; the shim instead forwards caller stdin unchanged. OpenSSH non-network control: `/usr/bin/ssh -G -F /dev/null review.invalid` → `stdinnull no`; same command with `-n` → `stdinnull yes` (both rc 0). EOF is the watcher lifetime protocol (X1), so this semantic mismatch masks command-killing regressions.
- M1: `internal/remote/ssh.go:25`, `return append(args, "-T", "--", address, script)` → `return append(args, "-n", "-T", "--", address, script)`. Full-suite command above rc 0; `ok github.com/eturkes/offload/test/e2e 29.373s`.
- Target-revision red test = `test/e2e/review_test.go:11`, `TestReviewFakeSSHNullStdin`, committed in `691bb76`. Command: `go test -count=1 ./test/e2e -run '^TestReviewFakeSSHNullStdin$'`; included in the combined red run below (rc 1). Output: `ssh -n must read /dev/null, but fake ssh forwarded stdin: "must-not-reach-remote-command"`.
- Acceptance: fake option semantics or assertions must make M1 red; preserve ordinary byte-exact stdin-through-EOF coverage.

## F02 — major — X6 transport test never reaches X6

- Anchors: `test/e2e/execution_test.go:221–227` creates `.transport-fail`, then invokes `run --pool`; `test/e2e/testdata/ssh:32–35` fails every post-probe connection. First post-probe connection = upload rsync. Therefore this test only proves upload failure returns 125, not run-script transport/status disambiguation.
- M2: `internal/run/run.go:86`, `return rc, ErrNotStarted` → `return rc, nil`. Full-suite command rc 0; `ok github.com/eturkes/offload/test/e2e 27.868s`. A transport 255 during execution can now be treated as command 255 with no valid status record.
- Acceptance: exercise execution-stage SSH failure after a successful probe, e.g. `--no-sync`; cover absent/stale/mismatched nonce records. M2 must fail through CLI status, not inspection of implementation text.

## F03 — minor — status assertions accept missing metrics

- Anchor: `test/e2e/cli_test.go:206`, patterns `{1/2, CPUs=8, busy=25, word 1}`. The GPU check `\\b1\\b` already matches slot count `1/2`; LOAD and FREE MEM have no assertions.
- M3: `internal/cli/place.go:226–227`, replace printed `s.Load`, `float64(s.MemKB)/(1<<20)`, `s.GPUs` with typed zero constants, leaving CPUs/busy/slots intact. Full-suite command rc 0; `ok github.com/eturkes/offload/test/e2e 28.168s`.
- Acceptance: parse/compare independently distinguishable status columns against fixture values; replacing any R2 metric with zero must fail.

## F04 — major — timing/concurrency/connection predicates lack regression protection

No effective test assertion for these contract clauses; each mutant independently survives the full-suite command:

| mutant | anchor / exact change | violated predicate | rc / e2e output |
|---|---|---|---|
| M4 | `internal/probe/probe.go:21`: `Timeout = 8 * time.Second` → `80 * time.Second` | R1 ≤8 s deadline; no timeout/reap test, only Parse + local Script tests | 0; `ok .../test/e2e 29.806s` |
| M5 | `internal/remote/ssh.go:14–20`: `SSHOptions = []string{}` | X5 shared connection options; `.ssh-call-*` fixtures only count calls, never assert options | 0; `ok .../test/e2e 31.743s` |
| M6 | `internal/cli/place.go:22`: `downFor = 30 * time.Second` → `0 * time.Second` | L3 down-host probe throttle ≥30 s | 0; `ok .../test/e2e 32.103s` |
| M7 | `internal/cli/place.go:119`: workdir lock basename appends `fmt.Sprint(os.Getpid())` | S4 same-root runs must share one controller workdir lock; per-PID lock admits concurrent phases | 0; `ok .../test/e2e 34.094s` |

Additional clauses with no direct assertion after complete test-source inspection: R1 probes overlap + transport killed/reaped at deadline; S2 I/O failure vs contention; S3 CLI-generated pool owner text; S4 busy-workdir releases slot; L2 lost acquisition race; L3 retry lower bound; Y4 non-ENOENT lstat failure; Y5 git-list failure before prune + global-excludes behavior on remote stale paths; Y11 300 s lock timeout; X1 default QUIT disposition. These are coverage inventory, not separate demonstrated defects. Standard dependency behavior need not be duplicated in tests; the demonstrated survivors concern offload-owned guarantees.

Acceptance: independently staged fake transports and concurrent command markers exercise the four demonstrated predicates, killing M4–M7; remaining inventory receives an explicit contract-based ruling.

## F05 — minor — explicit zero slots is untested and accepted

- Anchor: `internal/pool/pool.go:90–93`, exact bytes `case h.Slots == 0:` → `h.Slots = DefaultSlots`. Missing key and explicitly supplied zero collapse into the same Go value. P2 requires an integer ≥1 when supplied; only absence defaults to 2.
- Red test = `internal/pool/review_test.go:5`, `TestReviewP2ExplicitZeroSlotsIsInvalid`, commit `691bb76`. Command: `go test -count=1 ./internal/pool -run '^TestReviewP2ExplicitZeroSlotsIsInvalid$'`; included in combined red run below, rc 1.
- Output: `P2: explicit slots=0 accepted as [{Name:alpha Address:alpha Slots:2}]; only an absent slots key defaults to 2`.
- Acceptance: distinguish absence from explicit zero; reject zero before any work; default-absent and positive-slot tests remain green.

## F06 — major — failed output listing reports command success / loses outputs

- Anchor: `internal/run/remote.bash:63–68`, exact bytes `git_list | while IFS= read -r -d '' f; do`; the script lacks pipefail/error propagation for this listing. If `git ls-files` fails after CMD, the empty downstream loop succeeds, an empty download list is written, and command status is recorded as if output collection succeeded. Y6/Y10 require listing/download failure →125.
- Red test = `test/e2e/review_test.go:23`, `TestReviewY10OutputListFailureOverridesCommand`, commit `691bb76`. CMD writes an output, corrupts only remote `.git/index`, then exits 0/42/255. All three preserve that status despite `fatal: .git/index: index file smaller than expected`.
- Command: `go test -count=1 ./test/e2e -run '^TestReviewY10OutputListFailureOverridesCommand$'`; included in combined red run below, rc 1.
- Output (0 branch): `exit = 0, want 125; stdout="command-ran" stderr="offload: alpha slot 1/2 (...)\nfatal: .git/index: index file smaller than expected\n"`. Other branches: `exit = 42, want 125`; `exit = 255, want 125`.
- Acceptance: output-list generation failure becomes an offload failure without pretending a complete download; preserve ordinary command rc semantics after successful collection.

## F07 — minor — lock-serialization unit test measures its own holder

- Anchor: `internal/run/run_test.go:304–305`, exact bytes `_ = holder.Wait()` immediately before `time.Since(t0) < 1500*time.Millisecond`.
- Why: elapsed time includes waiting for the fixture's independent two-second `sleep`. The run script may return immediately without acquiring a lock; waiting for the fixture then supplies the required elapsed time anyway.
- M8: `internal/run/remote.bash:40`, replace `flock -w 300 9` with `true`. Command: `env GOCACHE="$W/.scratch/review/cache" go test -count=1 ./internal/run -run '^TestRemoteLockSerializes$'`. Output: `ok github.com/eturkes/offload/internal/run 2.003s`; rc 0.
- Scope: this proves the named unit test is ineffective, not that every Y11 test is ineffective; e2e predecessor/lock tests provide independent coverage.
- Acceptance: use lock-acquired/released handshakes and assert the run cannot finish while the holder owns the lock; M8 must fail this test without counting holder cleanup time.

## Red-first replay evidence (V4)

Each replacement is isolated + reverted. Commands use the private GOCACHE/runtime environment above; actual command selectors + complete output are in `red_replays.py` / `R*.log`.

| rollback | command after `go test -count=1` | rc / distinctive output |
|---|---|---|
| `cat <&0` → `cat` | `./internal/run -run '^Test(CommandOutlivesOpenStdin\|RemoteLockSerializes)$'` | 1; `rc 143 err <nil>; want 4`; remote-lock case `rc 143` |
| remove `trap - INT QUIT` | `./test/e2e -run '^TestC3ExitPrecedence$/(--local\|--pool)/command-INT$'` | 1; both places `exit = 0, want 130` |
| disable nested preflight | `./test/e2e -run '^TestY3UnsupportedGitLayouts$/nested$'` | 1; `remote work began before layout rejection` |
| disable framing guard | `./internal/run -run '^TestRemoteFramingMismatchPrunesNothing$'` | 1; all four cases `rc 143 stderr ""` |
| disable damaged `.git` check | `./internal/mirror -run '^TestInspectRefusesBrokenRepo$'` | 1; `broken repo inspected as {... Git:false ...}` |
| remove finite-load guard | `./internal/probe -run '^TestParse$'` | 1; `Parse("8\\nNaN\\n1\\n0\\n") accepted`; `+Inf` likewise |
| disable address/duplicate-address checks | `./internal/pool -run '^TestLoadRejects$'` | 1; missing/option/URI/port/space/user-option/duplicate address `err = <nil>` |

Grace replay: separate temporary snapshot of `cac258b`; fixture PATH wraps `pgrep` with 40 ms delay (models process-list latency, no protocol change). `heavy go test -count=1 ./test/e2e -run '^TestX1ControllerDeathKillsSession$/local-kill-escalation$'`: current deadline rc 0; old `100 × (sleep .1 + pgrep)` count loop rc 1, `session process ... remains after 13s`, `session survived controller SIGKILL: [...]`. Initial driver run failed its own expected-anchor-count assertion before tests; corrected 2→1, then both checks ran. This recreates the reported slow-process-query failure, not the exact historical host load.

## Checks / limits

- Three new red functions (five failing leaf cases) committed on `wt/reviewer-4`, `691bb7622d5e07ea17ad55d3028a7fd71c2aa3ba`; only `internal/pool/review_test.go` + `test/e2e/review_test.go` changed. Actual combined command: `go test -count=1 ./internal/pool ./test/e2e -run '^TestReview'` (private env above), rc 1. Individual commands in F01/F05/F06 are focused reproductions; combined run supplied the observed output.
- New tests: pinned `golangci-lint fmt --diff` rc 0; `golangci-lint run ./internal/pool ./test/e2e` rc 0 (`0 issues.`). They intentionally keep the target red; no fix applied.
- All tracked Go test files, fake SSH fixture, relevant production functions, contract, gate/tool/config/workflows, every rule, and all closing-diff commit bodies read.
- Archived red suite reproduced from tag `archive/tester-1` (`0b8eb3b`): `go test -json -count=1 ./test/e2e` rc 1; 36 failing top-level functions, 100 failing leaves, 121 total failure rows. Sample: `exit = 2, want 0; stdout="" stderr="Usage:..."`.
- Skip query rc 1 = no `t.Skip`, `testing.Short`, build constraints, `t.Parallel` in internal/test; positive control helper env branch expected 1, got 1, rc 0.
- Full gate not run by reviewer: package/script selection inspected; full test step ran in nine invocations (baseline + seven isolated mutants + clean two-CPU `-count=2`). Actual gate tool integrity/deployment/network claims remain outside these local executions.
- Real-host commit claims unconfirmed under explicit no-real-SSH brief. No live CI runner exercised. LSP unavailable for `.go`; source/caller review used instead.
- Branch contains only the two committed test files; production bytes restored, tracked tree clean. Disposable caches/snapshot copies/runtime dirs removed; `.scratch/review/` retains experiment drivers + logs as harvest evidence. No instruction/spec/config edits.
- Final validator: expected 0 unknown rows → got 0 (`grep -c` rc 1 = no matches); positive control expected 1 → got 1, rc 0. 7/7 rows adjudicated: 4 pass, 3 finding; 7 findings (4 major, 3 minor).
