# reviewer-1 — code review @ 640f24f

## Verdict table

| row | check | verdict | finding ids | evidence (file:line, command) |
|-----|-------|---------|-------------|-------------------------------|
| C1 | m1 cli | finding | R1-F2, R1-F5, R1-F8 | `cli.go:169–355`: all run returns flow through `runCommand` → `End`; C3 signal precedence checked after inspection/placement/run/sync. `place.go:57–173`: pool-first scoring, candidate release, one waiting line, 2s retry, 30s down cache + forgotten GPU inventory read. `ctr.go:228–237`: truncated parser red. Original cli `go test -race -count=1 ./internal/cli/` → rc 0; new tests → rc 1. E2e evidence/limits below. |
| C2 | m2 run | finding | R1-F3 | `run.go:96–149`: X3 pipe-close/wait + X6 nonce recovery checked; `watch.bash:56–115`: launch, EOF/TERM/HUP, session/scope teardown + fd 3 retention; `remote.bash:43–91`: flock, framed prune, status + down-list sequence. All watcher/remote/boot variable bindings use `__…` except intentional XDG_RUNTIME_DIR/INVOCATION_ID restoration; local.bash only delegates. Imported errexit bypasses status/teardown. `go test -race -count=1 ./internal/run/` → rc 0 (120.619s); new errexit e2e → rc 1. |
| C4 | m4 slots | finding | R1-F6 | `slots.go:113–145,178–239`: ascending nonblocking lock, busy vs I/O error on acquisition, owner write/clear before close, idempotent release, CLOEXEC + local watcher transfer, wait lock closes on cancellation. `slots.go:148–162`: status conflates an I/O error with held lock. `go test -race -count=1 ./internal/slots/` → rc 0; new R2 e2e → rc 1. S5 controller-SIGKILL scenario exercised in serial e2e sample. |
| C7 | m7 remote | pass | — | `remote/ssh.go:19–46,59–110`: shared SSH options/`-T` reach probe, rsync, run, recovery + ctr-check; mux resolved once per host set, short-path/failure fallback; Quote/Join and process-group cancellation callers fit. `go test -race -count=1 ./internal/remote/` → rc 0. X3/X6 status recovery standalone e2e → rc 0; full mux-path e2e not rerun (source + existing test read). |
| C9 | m9 ledger | finding | R1-F4 | `ledger.go:103–308`: J1 path fallback, serialized whole-line write + rollback, J3 rotation, J4 once-only advisory error, J5 start/end topology read; `cli.go:169–182,264–334`: placement/timing/status integration + end wrapper. `ledger.go:215–229`: return rereads status after record was fixed. Original ledger `go test -race -count=1 -run '^TestJ' ./internal/ledger/` → rc 0; new J2 seam test → rc 1. J e2e tests pass in serial sample (sample overall fails an unrelated X3 readiness bound). |
| C10 | m10 shell | finding | R1-F1 | `internal/shell/look.go:28`: explicit empty PATH becomes zero search entries; Bash searches cwd. `go test -count=1 -run '^TestReviewR1X2EmptyPathMatchesBash$' ./internal/shell/` → rc 1. Callers: `cli.checkLocalCommand` → run + native ctr, `ctr.Init`. |
| C11 | m11 e2e harness | finding | R1-F7 | `harness_test.go:25–66,97–151,244–315`: static temporary builds, sandbox env/log dirs, process/files cleanup, wait/read channels reviewed; `ctrhelper/main.go` every mode read, readiness follows signal setup. `testdata/ssh:35–42`: probe marker dispatch can steal command data. `TestHarnessFakeSSH` passes in serial e2e sample; new probe-word e2e → rc 1. Real-podman helper paths not rerun in this scope. |

## Findings

### R1-F1 — empty PATH prevents a valid current-directory command

- Severity: minor. Anchor: `internal/shell/look.go:28` @ 640f24f (`for _, dir := range filepath.SplitList(os.Getenv("PATH")) {`). Contract: X2 / K2 Bash-compatible lookup.
- `filepath.SplitList("")` returns no entries; Bash treats explicit `PATH=''` as a current-directory search. Consequently native `ctr` and `run --local` reject a runnable command with 127. Nonempty PATHs containing an empty element are handled; the zero-byte whole value is not.
- Red test: `internal/shell/cr_reviewer1_test.go`, `TestReviewR1X2EmptyPathMatchesBash`; absolute `/bin/bash` oracle resolves the same script with the same empty PATH (rc 0), then `Look` returns 127.
- Command (cwd = `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-1`): `go test -count=1 -run '^TestReviewR1X2EmptyPathMatchesBash$' ./internal/shell/` → rc 1 on target code.

```text
--- FAIL: TestReviewR1X2EmptyPathMatchesBash (0.00s)
    cr_reviewer1_test.go:25: bash resolved the current-directory command (rc 0); Look("offload-review-current-command") = "", 127, command not found
```

### R1-F2 — hybrid placement rejects host-only commands before probing

- Severity: major. Anchor: `internal/cli/cli.go:240–243` @ 640f24f (`if _, rc, ok := checkLocalCommand(o.argv[0], stderr); !ok {`). Contract: L1/L2 pool-first placement + X2 remote login PATH.
- Every hybrid request checks the controller's PATH before `p.place`. A command installed only on an available pool host is rejected (127), although the same request with `--pool` succeeds. The local preflight must constrain local fallback, not invalidate eligible remote execution.
- Red test: `test/e2e/cr_reviewer1_test.go`, `TestReviewR1L2HybridCommandExistsOnlyOnHost`; fake host `.bash_profile` prepends its `.local/bin`; `--pool` control succeeds with `host-only\n`; otherwise identical hybrid run fails.
- Command (review worktree cwd): `go test -count=1 -run '^TestReviewR1L2HybridCommandExistsOnlyOnHost$' ./test/e2e/` (also run in a two-test regex) → rc 1 on target code.

```text
--- FAIL: TestReviewR1L2HybridCommandExistsOnlyOnHost (6.14s)
    cr_reviewer1_test.go:27: exit = 127, want 0; stdout="" stderr="offload: offload-review-host-only-command: command not found\n"
```

### R1-F3 — exported errexit terminates the watcher before normal completion

- Severity: major. Anchor: `internal/run/watch.bash:56–58,86,99` @ 640f24f (`pkill -TERM -s "$__p" 2>/dev/null`; unguarded `wait -n` likewise). Contract: C3 command status + X1 teardown + X2 inherited environment.
- Bash imports `SHELLOPTS` before executing the embedded watcher. With exported `SHELLOPTS=errexit`, normal nonzero control operations (`wait`, `pkill` when the session is empty) terminate the watcher instead of running its explicit status/teardown path. Even `true` exits 1. X2 reserves `__…` names, not SHELLOPTS. Preserve the command environment while isolating watcher control flow from imported shell options.
- Red test: `test/e2e/cr_reviewer1_test.go`, `TestReviewR1X2ExportedErrexitPreservesStatus`; environment `SHELLOPTS=errexit`, command `true`, `--local` → 1 instead of 0.
- Command (review worktree cwd): `go test -count=1 -run '^TestReviewR1X2ExportedErrexitPreservesStatus$' ./test/e2e/` (also run in a two-test regex) → rc 1 on target code.

```text
--- FAIL: TestReviewR1X2ExportedErrexitPreservesStatus (1.62s)
    cr_reviewer1_test.go:37: exit = 1, want 0; stdout="" stderr="offload: local slot 1/2\n"
```

### R1-F4 — end-record status and returned status can disagree

- Severity: minor. Anchor: `internal/ledger/ledger.go:215–229` @ 640f24f, especially `rc, _ := final()` after `r.write`. Contract: J2 fixes both record + exit status once the end write owns its lock.
- `final()` is read once when encoding the end record, again after file open/write/close. A signal published between those reads produces recorded `rc=0, signal=null` while offload returns 143. Existing `TestJ2SignalDuringEndWriteCounts` covers the wait *before* lock acquisition, not this later interval.
- Red test: `internal/ledger/cr_reviewer1_test.go`, `TestReviewR1J2EndStatusFrozenWithRecord`; existing `openLog` seam publishes the equivalent final-status change after record encoding, before the file operation.
- Command (review worktree cwd): `go test -count=1 -run '^TestReviewR1J2EndStatusFrozenWithRecord$' ./internal/ledger/` → rc 1 on target code.

```text
--- FAIL: TestReviewR1J2EndStatusFrozenWithRecord (0.00s)
    cr_reviewer1_test.go:24: end record rc=0 signal=<nil>, End returned 143; J2 requires one fixed status after the lock
```

### R1-F5 — long unrelated output silently truncates the layer list

- Severity: major. Anchor: `internal/cli/ctr.go:228–237` @ 640f24f (`sc := bufio.NewScanner(strings.NewReader(out))`, `return sum, layers, ok`). Contract: K10 complete ordered layer-list comparison; unrelated output never counts.
- Scanner's default token bound is 64 KiB; `sc.Err()` is ignored. A valid header, then a long unframed output line, then differing framed layers returns `ok=true` and an empty layer list. An empty local image with the same binary is falsely reported `same`. This is new evidence beyond `ledger-pool-run-environment.md` R10's short-banner repair; no prior accepted limit excludes long banners.
- Red test: `internal/cli/cr_reviewer1_test.go`, `TestReviewR1K10LongNoiseCannotTruncateLayers`; header + 65536-byte unrelated line + one framed differing layer.
- Command (review worktree cwd): `go test -count=1 -run '^TestReviewR1K10LongNoiseCannotTruncateLayers$' ./internal/cli/` → rc 1 on target code.

```text
--- FAIL: TestReviewR1K10LongNoiseCannotTruncateLayers (0.00s)
    cr_reviewer1_test.go:15: parseCheck accepted=true sum="same-binary" layers=[]; long unframed noise hid the differing layer (an empty local image now compares same)
```

### R1-F6 — status labels an I/O error as an occupied slot

- Severity: minor. Anchor: `internal/slots/slots.go:155–161` @ 640f24f (`f, held, err := tryLock(p)`, then owner append for every result except successful acquisition). Contract: R2 a slot counts as held by its lock alone; S2 contention differs from I/O failure.
- An existing slot file that refuses write-open produces an error, not evidence of a flock holder. `Holders` nevertheless appends it and `status` exits 0 with `1/2 held`. This misreports occupancy rather than surfacing the inaccessible state.
- Red test: `test/e2e/cr_reviewer1_test.go`, `TestReviewR1R2StatusDoesNotTurnIOErrorIntoHolder`; slot mode 0400; positive controls prove write-open fails and read-open/nonblocking exclusive flock succeeds, then releases it. Accepts a nonzero diagnostic or truthful no-holder status; rejects successful false occupancy.
- Command (review worktree cwd): `go test -count=1 -run '^TestReviewR1R2StatusDoesNotTurnIOErrorIntoHolder$' ./test/e2e/` (also run paired with harness red) → rc 1 on target code.

```text
--- FAIL: TestReviewR1R2StatusDoesNotTurnIOErrorIntoHolder (0.13s)
    cr_reviewer1_test.go:58: unlocked file's I/O error reported as a held slot: "local heavy slots: 1/2 held\n  slot 1: \n\nHOST           STATE  SLOTS  CPUS   LOAD  BUSY  FREE MEM GPUS\nalpha          up       0/2     8   2.00   25%     1.0GB    0\n"
```

### R1-F7 — fake ssh interprets command words as a probe

- Severity: minor. Anchor: `test/e2e/testdata/ssh:35–42` @ 640f24f, especially `*loadavg*|*nproc*)`. Contract/check: C11 wrapper routes to the requested operation; `.claude/rules/code.md` fake-host `.probe` = probe answer, not command replacement.
- Substring dispatch examines the entire remote command, including its quoted user argv. Any run containing `nproc` or `loadavg` bypasses execution and returns the probe fixture. A no-sync run of `printf %s literal-nproc` succeeds with `8 2.000000 1048576 0\n`, not its requested output. This is a harness defect, not a claim that real OpenSSH does this.
- Red test: `test/e2e/cr_reviewer1_test.go`, `TestReviewR1HarnessProbeWordsStayCommandData`.
- Command (review worktree cwd): `go test -count=1 -run '^TestReviewR1HarnessProbeWordsStayCommandData$' ./test/e2e/` (also run paired with R2 red) → rc 1 on target code.

```text
--- FAIL: TestReviewR1HarnessProbeWordsStayCommandData (0.14s)
    cr_reviewer1_test.go:68: fake ssh returned probe fields instead of executing printf: stdout="8 2.000000 1048576 0\n"
```

### R1-F8 — remote cancellation discards a returned command status

- Severity: minor. Anchor: `internal/cli/cli.go:304–307` @ 640f24f (`if err != nil || ctx.Err() != nil {` precedes `runlog.CommandStatus(rc)`). Contract: J2 `cmd_rc` = command status; null only when it never ran or its status never came back. X3 still owns the outer 143.
- `Remote.Run` waits for the remote watcher and returns its ordinary status (17 here), even under cancellation. `remoteRun` returns early solely because ctx is canceled and drops that known status. Local execution records it correctly. Logging the returned command status must remain independent of the signal's precedence over offload's own status.
- Red test: `test/e2e/cr_reviewer1_test.go`, `TestReviewR1J2CancellationKeepsReturnedCommandStatus`; local + remote TERM-trapping commands exit 17, offload exits 143. Local control records 17. Remote status file positively confirms `<nonce> 17`; remote end record records null.
- Command (review worktree cwd): `go test -count=1 -run '^TestReviewR1J2CancellationKeepsReturnedCommandStatus$' ./test/e2e/` → rc 1 on target code.

```text
--- FAIL: TestReviewR1J2CancellationKeepsReturnedCommandStatus (6.09s)
    --- FAIL: TestReviewR1J2CancellationKeepsReturnedCommandStatus/--pool (4.39s)
        cr_reviewer1_test.go:89: --pool returned command status was discarded: cmd_rc=<nil>, rc=143, signal=TERM; want cmd_rc=17
```

## Test artifact + verification

- Red-test commit: `3b56e10ce7980988d0b35300111cd7083b3eadcd` on `wt/reviewer-1`; parent = target `640f24f293d7c0f3da20b0761bd4b521d3be2e13`. Implementation unchanged. Four new files: `internal/{cli,ledger,shell}/cr_reviewer1_test.go`, `test/e2e/cr_reviewer1_test.go`.
- All commands below run from `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-1`. New tests finalized + rerun before their test-only commit:

| command | result |
|---|---|
| `go test -count=1 -run '^TestReviewR1' ./internal/cli/` | rc 1; R1-F5 fires |
| `go test -count=1 -run '^TestReviewR1' ./internal/ledger/` | rc 1; R1-F4 fires |
| `go test -count=1 -run '^TestReviewR1' ./internal/shell/` | rc 1; R1-F1 fires |
| `go test -count=1 -run '^TestReviewR1' ./test/e2e/` | rc 1, 18.786s; R1-F2/F3/F6/F7/F8 all fire; local cancellation + host-only pool controls pass |
| `go test -race -count=1 ./internal/slots/` | rc 0, 2.838s |
| `go test -race -count=1 ./internal/remote/` | rc 0, 1.090s |
| `go test -race -count=1 -run '^TestJ' ./internal/ledger/` | rc 0, 5.001s; original tests only |
| `go test -race -count=1 ./internal/cli/` | rc 0, 1.048s; before the new red test was added |
| `go test -race -count=1 ./internal/run/` | rc 0, 120.619s |
| `go test -count=1 -run '^TestReviewX3SignalDuringStatusRecovery$' ./test/e2e/` | rc 0, 8.961s; isolated rerun |

Original targeted e2e sample command:

```sh
go test -count=1 -run '^(TestX1ControllerDeathKillsSession|TestX3SignalsSkipDownload|TestX6Remote255AndTransport|TestReviewX3SignalDuringStatusRecovery|TestReviewC3SignalDuringInspection|TestL3DownHostProbedAtMostEvery30s|TestS4SameTreeRunsSerializeAtController|TestHarnessFakeSSH|TestJ[125].*)$' ./test/e2e/
```

- First run: rc 1, 49.248s. Failures: status-recovery readiness; X3 TERM/INT/HUP readiness; X1 remote + local-kill-escalation readiness; J2 `.download-fail` runtime; S4 serialized run runtime; J2 phases plain/nosync-get/git runtime. Bounds = 6s readiness / 20s runtime. Representative: `no status recovery within 6s; stdout="" stderr="offload: alpha slot 1/2 (8 CPUs, 25% busy) → ~/offload/project-98421819\n"`.
- Same command with `-parallel=1`: rc 1, 124.898s; only `TestReviewX3SignalDuringStatusRecovery` fails (`no status recovery within 6s`, otherwise only placement line). All other selected original checks pass. Isolated rerun of that test passes, above. No timing assertion was edited; these failed runs remain failed evidence, not declared green.
- Coverage limits: no full gate, live-pool execution, full e2e suite, mutation run or real-podman helper-path execution; per brief. Scope-specific case execution versus skip is unconfirmed: the nonverbose run-package output does not enumerate skipped cases. LSP on `internal/cli/cli.go` returned `No LSP server available for file type: .go`; call graph checked with source + path-bounded `rg`.

## Register

- R1-R1, unproven, security lead only: `internal/slots/slots.go:160` reads an owner with blocking `os.ReadFile` after lock-open errors; a FIFO owner may stall status. No exploration; security owner must rule.
- R1-R2, unconfirmed: original e2e readiness/runtime failures above have no established cause. Serial execution removed all but one; isolated rerun removed the last. Shared runtime contention is an inference, not a finding or exoneration.
- Source/law/config edits: none. Red tests only in the isolated worktree; this deliverable is MAIN's validator input. No accepted prior ruling reversed.

reviewer-1-DONE-1
