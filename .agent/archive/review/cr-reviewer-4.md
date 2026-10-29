# reviewer-4 — code review @ 640f24f

Subject root: `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-4`; paths below are relative to it. Production bytes = 640f24f throughout. Test-only commits: `7f9e791` (J2), `148ea90` (X2/K2). Check set = the 27 seeded rows, fixed before source reads. `pass` = no proven blocking claim defect within the stated evidence/limits, not exhaustive runtime proof.

## Verdict table

| row | check | verdict | finding ids | evidence (file:line, command) |
|-----|-------|---------|-------------|-------------------------------|
| G1 | `README.md` (every promise, example + recipe) | finding | R4-F3, R4-F4, R4-F5 | README log rc + native-ctr parity promises meet proved behavioral counterexamples; no-sync text omits Y9 exception. Other recipes traced to implementations, sync/exit witnesses executed; destructive/shared recipes inspected only. |
| G2 | `--help` / usage + user-facing error texts (`offload`, `offload run`, `offload ctr`, `heavy`) | finding | R4-F1 | `internal/cli/cli.go:63` omits XDG lookup tier. Built `.scratch/offload`; help/usage/missing-pool/conflict/bad-switch forms + native ctr smoke run (`.scratch/usage.log`); CLI package + C1–C5 e2e rc 0. |
| G3a | `.agent/contract.md` family C predicates vs behavior (each id) | finding | R4-F2 | C1/C2/C4/C5: parser + CLI output/placement tests; C3 status matrix + signal tests. `go test -count=1 -run '^(TestC[1-5].*\|TestP3MissingPool\|TestR2.*\|TestL[1-4].*\|TestY4LinkedWorktreeRunsRemotely\|TestY9NoSync.*)$' ./test/e2e` rc 0; C3 worktree clause contradicts passing Y4 witness. |
| G3b | `.agent/contract.md` family P predicates vs behavior (each id) | pass | | P1–P3: `internal/pool/pool.go:43–111`, parser tests (`go test -count=1 ./internal/pool`, rc 0); P2/P3 run integration in `test/e2e/cli_test.go:159,220`; isolated missing-pool CLI → rc 2 + path. |
| G3c | `.agent/contract.md` family R predicates vs behavior (each id) | pass | | R1: `internal/probe/probe.go:17–96`, parser boundaries + inherited-output deadline test; R2: `internal/cli/place.go:196–236`, e2e `TestR2Status`, `TestR2StatusShowsProbeNumbers`. Probe package + selected R2 e2e rc 0. |
| G3d | `.agent/contract.md` family S predicates vs behavior (each id) | pass |  | S1–S6: `internal/slots/slots.go:35–251`, heavy differential, ascending locks, owner/fd/teardown + S6 tests; slots package rc 0. S4 serialization e2e + S5 watcher-lock tests ran rc 0. |
| G3e | `.agent/contract.md` family L predicates vs behavior (each id) | pass |  | L1–L4: `internal/cli/place.go:57–181`; selected L e2e rc 0, including ratio/order, fallback, waiting/signals, down/busy/no-GPU. L2 reservation-order prose is mechanism drift (register R4-R1), not a demonstrated placement-result defect. |
| G3f | `.agent/contract.md` family Y predicates vs behavior (each id) | pass |  | Y1–Y11: `internal/mirror/mirror.go`, `internal/run/remote.bash`, `internal/cli/cli.go:295–356`; mirror package + Y-prefixed e2e + run prune/framing/lock tests rc 0. Linked-worktree + checksum + ignored-path + no-sync/get behaviors exercised. |
| G3g | `.agent/contract.md` family X predicates vs behavior (each id) | finding | R4-F4 | X2 fails explicit empty PATH. X1/X3/X8: run package rc 0 (93.7 s), including scope admission/teardown + boot witnesses; X2–X6/X8 e2e rc 0 (54.9 s batch); X5 options checked in fake ssh + remote package. X7 = environment requirements, not automatic enforcement. |
| G3h | `.agent/contract.md` family J predicates vs behavior (each id) | finding | R4-F3 | J2 finalization does not freeze the returned status with the serialized one (`internal/ledger/ledger.go:215–229`), red `TestR4J2EndRecordAndExitShareFinalStatus` @ 640f24f, test commit `7f9e791`. J1/J3/J4/J5 mapping + remaining witnesses below. |
| G3i | `.agent/contract.md` family K predicates vs behavior (each id) | finding | R4-F4 | K2/K7 inherit X2 empty-PATH failure. K1/K3–K10 implementations + tests mapped below; original ctr package rc 0. Parallel real K/X1 batch rc 1; narrow serial K7/K8/K10/X1 rerun rc 0 (7 events, zero skips); R4-R6 records limits. K11 requirements structurally present, compatibility floor not rerun on another host. |
| G4 | `.agent/spec.md` `Artifacts` (paths, run commands, live facts) | pass |  | `.agent/spec.md:7–8`: named artifacts exist; build + CLI usage exercised. Read-only controller/pool sha256: offload `27f42327…`, heavy `cc6f4a50…` equal; live pool nixbox multiplex=true. Deploy/reinstall/re-seal recipes inspected, not executed. |
| G5 | `machines/nixbox/README.md` runbook vs scripts + config | pass |  | `machines/nixbox/README.md:1–70` against all machine scripts + Nix imports; live read-only CPU=8, memory≈62.6 GiB, OFFLOAD_CTR=1, node26, swipl10.0.2, PL1=22000000, BIOS1.43.0, CPUWeight=idle/MemoryHigh=20GiB. Historical timings/firmware events unconfirmed; limits below. |
| G6 | `pool.example.toml` comments vs P-predicates | pass | | `pool.example.toml:1–8` matches P1/P2 + X5; the parser accepts address/slots/multiplex (`internal/pool/pool_test.go:38,48`, rc 0). |
| G7 | facts stated in `.claude/rules/*` (timings, recipes, mechanisms) vs code | finding | R4-F6 | All five `.claude/rules/*.md` read vs code/workflows; fake marker direct probe→0 vs non-probe→255 contradicts “every call”. Gate shape + --pins metadata + clean/red spec-lint checked. Historical timing/firing claims not all replayed (coverage limits). |
| G8 | code comments m1 cli | finding | R4-F4 | `internal/cli/ctr.go:87` “as if typed” native exec claim fails empty PATH via shared shell.Look. Other comments in cli/place/ctr match implemented sequencing; usage-text findings tracked under G2. |
| G9 | code comments m2 run | pass |  | Comments in `internal/run/{run.go,watch.bash,local.bash,remote.bash}` match boot framing, status recovery, session/scope teardown + fd ownership. `go test -count=1 ./internal/run` rc 0; quiet output did not collect individual skip events (coverage limit). |
| G10 | code comments m3 mirror | pass |  | Comments in `internal/mirror/mirror.go` match root/layout checks, upload/prune/download/get and checksum flags. Mirror package + Y e2e rc 0; old ~0.5s sync measurement not rebenchmarked. |
| G11 | code comments m4 slots | finding | R4-F7 | `internal/slots/slots.go:18` comment says death frees lock, unlike S5 watcher inheritance. Other slot comments match MakeDirs, acquire/release/wait code; slots package rc 0. |
| G12 | code comments m5 pool + m6 probe | pass | | `internal/pool/pool.go:1–49`, `internal/probe/probe.go:1–70` comments match decode, validation, parsing + timeout; package tests both rc 0. R1 timeout's descendant bound closes pipes, not descendants; the stated killed/reaped subject is ssh. |
| G13 | code comments m7 remote + m10 shell | finding | R4-F4 | `internal/shell/look.go:1–3,20–22` promises bash parity contradicted by empty-PATH red. `internal/remote/ssh.go` comments match connection argv, timeout, mux paths + quoting; original remote/shell packages rc 0. |
| G14 | code comments m8 ctr | pass |  | `internal/ctr/{ctr,tag,init}.go` comments checked against image/tag, socket, scope and PID1 code. Original ctr package rc 0; package comment summarizes K8 and inherits its documented stopped-init/uninterruptible-I/O limit. |
| G15 | code comments m9 ledger | pass |  | `internal/ledger/ledger.go` comments describe implemented append/rotation/timing and two final-status reads. J2 behavioral mismatch is ruled under G3h/R4-F3; the End comment itself narrows its equality statement to signals during lock waiting. Original J tests rc 0. |
| G16 | code comments m11 e2e harness + test comments citing ids | finding | R4-F7 | `test/e2e/ctr_test.go:465` still calls podman part of offload session. Harness/fake-host + contract-id test comments otherwise matched to assertions; claimed test skip conditions + static builds checked. No mutation credit claimed for this review. |
| G17 | comments m12 tools | pass |  | `tools/heavy`, `tools/install`, `tools/pooladdrs/main.go` comments match argv/slots/install paths/parser. Help/native smoke + slot differential executed; install not executed (writes installed/shared state). |
| G18 | comments m13 gate | pass |  | `.agent/{gate,tool,pins,plant,spec-lint}` comments traced to code. Full/fast gate steps + timings/summary wiring match; tool --pins + spec-lint smoke below. Gate/tool downloads/pin network lookups not rerun. |
| G19 | comments m14 machines (nix + scripts, `*.reviewed`) | pass |  | All `machines/nixbox` Nix/script/`*.reviewed` comments read against local implementation; read-only host facts agree on non-security configuration. Historical performance + firmware observations unconfirmed; security claims handed to MAIN by scope. |

## Findings

### R4-F1 — help skips the XDG_CONFIG_HOME pool tier

- Severity: minor. Target: `internal/cli/cli.go:63` @ 640f24f.
- Wrong against P1: when OFFLOAD_CONFIG is unset, a nonempty XDG_CONFIG_HOME takes precedence over `~/.config`; the shared usage text says the latter is next. This is the displayed help of run, ctr + status, not merely a code comment.
- Exact anchor: `The pool file is $OFFLOAD_CONFIG, else ~/.config/offload/pool.toml.`
- Proof: `rg -Fn 'The pool file is $OFFLOAD_CONFIG, else ~/.config/offload/pool.toml.' internal/cli/cli.go` → line 63. `internal/pool/pool.go:46–51` implements the missing tier; `go test -count=1 ./internal/pool` → `ok .../internal/pool` (rc 0), including `TestPathPrecedence`. `.scratch/offload --help` prints that anchor (rc 0).
- Acceptance: help names all three tiers in their actual precedence.

### R4-F2 — C3 still rejects supported linked worktrees

- Severity: minor. Target: `.agent/contract.md:11` @ 640f24f.
- Wrong against Decision “A linked git worktree runs remotely too” + Y3/Y4: C3 lists a worktree under `--pool` among offload failures (125), although supported linked worktrees run successfully. This is a stale contract clause, not a requested behavioral restriction.
- Exact anchor: `worktree or unsupported repo layout`.
- Proof: `rg -Fn 'worktree or unsupported repo layout' .agent/contract.md` → line 11. `go test -count=1 -run '^TestY4LinkedWorktreeRunsRemotely$' ./test/e2e` is the narrow witness; the selected C/P/R/L/Y suite containing it ran rc 0. The test invokes pool mode and checks the linked worktree branch, staged changes, refs + git admin links (`test/e2e/sync_test.go:302–347`).
- Acceptance: C3 names only unsupported layouts as rc 125, consistent with Y3/Y4.

### R4-F3 — late signal splits run-log rc from returned rc

- Severity: major. Target: `internal/ledger/ledger.go:215–229` @ 640f24f; especially the second `final()` at line 228.
- Wrong against J2: rc + signal must be fixed together once the end write holds its lock. End serializes the first `final()` result, then calls `final()` again after the file write. A signal delivered in between changes the returned status but not the stored record. The log can claim success for a run whose caller receives 143, defeating its field-assessment purpose.
- Red test: `internal/ledger/cr_reviewer4_test.go`, `TestR4J2EndRecordAndExitShareFinalStatus`; committed `7f9e791` on `wt/reviewer-4`, production files unchanged from 640f24f.
- Command (worktree root): `go test -count=1 -run '^TestR4J2EndRecordAndExitShareFinalStatus$' ./internal/ledger` → rc 1. The no-signal control passes; the existing openLog seam changes the status callback after serialization, inside the actual write interval.

```text
--- FAIL: TestR4J2EndRecordAndExitShareFinalStatus (0.00s)
    --- FAIL: TestR4J2EndRecordAndExitShareFinalStatus/signal-after-record-built (0.00s)
        cr_reviewer4_test.go:35: J2 split final status: end record rc=0 signal=<nil>, End returned 143
FAIL
FAIL github.com/eturkes/offload/internal/ledger 0.005s
```

- Acceptance: successful end writes return the same frozen rc represented by the record; failed writes still return the correct final status (J4). Preserve the existing signal-during-lock-wait witness.

### R4-F4 — empty PATH fails bash-compatible lookup

- Severity: major. Target: `internal/shell/look.go:28` @ 640f24f.
- Wrong against X2 + K2/K7: a name without `/` must use bash-compatible PATH lookup. An empty PATH denotes the current directory to bash; `filepath.SplitList("")` yields no entries, so Look reports 127 even for an executable in cwd. Native `offload ctr` therefore does not behave “as if typed”; the same helper is used by local run preflight + container init.
- Red test: `internal/shell/cr_reviewer4_test.go`, `TestR4X2EmptyPATHSearchesCurrentDirectory`; committed `148ea90` on `wt/reviewer-4`, production unchanged from 640f24f. Bash is an executed differential oracle, not an assumed expectation.
- Command: `go test -count=1 -run '^TestR4X2EmptyPATHSearchesCurrentDirectory$' ./internal/shell` → rc 1.

```text
--- FAIL: TestR4X2EmptyPATHSearchesCurrentDirectory (0.04s)
    cr_reviewer4_test.go:29: X2/K2: bash ran the current-directory command with empty PATH, Look returned path="" status=127 err=command not found
FAIL
FAIL github.com/eturkes/offload/internal/shell 0.055s
```

- Real binary witness: from an isolated directory with executable `r4cmd`, `PATH='' /bin/bash --noprofile --norc -c 'r4cmd'` → `regular-command`, rc 0; `PATH='' OFFLOAD_CTR=0 <worktree>/.scratch/offload ctr -- r4cmd` → `offload: r4cmd: command not found`, rc 127.
- Acceptance: explicitly empty PATH searches cwd, consistent with empty components inside a nonempty PATH; the oracle + real native ctr witness pass. Keep unset-PATH semantics separately adjudicated rather than conflating them with explicit empty.

### R4-F5 — no-sync help promises more than Y9

- Severity: minor. Targets: `README.md:94`, `internal/cli/cli.go:55` @ 640f24f.
- Wrong against Y9: `--no-sync` suppresses upload, prune + automatic download, but explicit `--get` still copies files back. Both user-facing surfaces instead promise no copies in either direction, without that exception.
- Exact anchor in both targets: `Do not copy files to or from the host.`
- Proof: `rg -Fn 'Do not copy files to or from the host.' README.md internal/cli/cli.go` → lines 94,55. `TestY9NoSync` (`test/e2e/sync_test.go:517–528`) invokes `--no-sync --get explicit`, requires the fetched content and absence of automatic output; it passed in both selected e2e batches. `internal/cli/cli.go:311–318` implements that exception.
- Acceptance: both descriptions name automatic sync and explicitly preserve `--get`.

### R4-F6 — fake transport-failure marker documentation includes probes incorrectly

- Severity: minor. Target: `.claude/rules/code.md:17` @ 640f24f.
- Wrong against the stated harness recipe: `.transport-fail` is documented as failing every call, but the fake ssh answers `.probe` / `.probe-hang` before it checks `.transport-fail`. An agent following this instruction to model a down host actually models an up host whose subsequent transfer/run fails.
- Exact anchor: `(every call 255)`.
- Proof: `rg -Fn '.transport-fail' .claude/rules/code.md` → line 17; `test/e2e/testdata/ssh:39–55` returns the probe first. `test/e2e/execution_test.go:373` names the fixture `fail after probe`; `TestX6Remote255AndTransport` passed in the selected X suite. Direct fake-ssh control: probe under marker → rc 0 + four fields, ordinary command under same marker → rc 255 (see execution record).
- Acceptance: describe `.transport-fail` as failing non-probe calls (or change the harness + tests under an approved behavior change); preserve the distinction from `.run-fail`.

### R4-F7 — lifecycle comments retain superseded lock/session semantics

- Severity: minor. Targets: `internal/slots/slots.go:18`, `test/e2e/ctr_test.go:465` @ 640f24f.
- Wrong against S5 + K8: offload's local slot remains held by the watcher after offload dies; podman runs in its own session, not merely another process group of offload's session. The current comments say the opposite. The test sends TERM to both PIDs directly; it does not establish shared-session delivery.
- Exact anchors:
  - `Release frees it; the kernel frees it if offload dies first.`
  - `offload + podman (its own group, offload's session), as X1's`
- Proof: `rg -Fn 'Release frees it; the kernel frees it if offload dies first.' internal/slots/slots.go` → 18; `rg -Fn 'offload + podman (its own group, offload' test/e2e/ctr_test.go` → 465. Mechanisms: `internal/run/run.go:55–62` passes the slot fd to the watcher; `internal/ctr/ctr.go:268–273` sets `Setsid: true`. `TestLockReachesWatcherNotCommand` passed with the run package. Prior accepted ledger `ledger-pool-run-environment.md:R11–R12` already establishes the own-session design; no ruling is contested.
- Acceptance: slot comment states last-holder/descendant lifetime; test comment describes deliberate direct delivery to both processes rather than session membership.

## Register

- R4-R1 — mechanism drift, no demonstrated placement-result defect: `.agent/contract.md:39` L2 describes probe → choose → acquire (lost race → retry); `internal/cli/place.go:113–167` reserves each candidate slot + workdir before probing, then releases losers. The existing lowest-ratio/order/wait tests pass. MAIN should decide whether to align the mechanism prose; no new business behavior requirement inferred here.
- R4-R2 — unconfirmed historical measurements: `.claude/rules/gate.md:5` gate wall times; `internal/mirror/mirror.go:329–332` checksum cost; `machines/nixbox/README.md:37,51,67–70` thermal/firmware/idle-task measurements; corresponding Nix comments. Read code + prior ledgers; did not rerun gate benchmarks, sustained thermal loads, firmware flashes, trading workloads or old-version package failures. Live knobs/versions checked separately.
- R4-R3 — security scope held: machine trust/secret/boot claims and `*.reviewed` judgments read for structural consistency alone; no independent security assessment, exploit experiments, credential access or security ruling overrides. MAIN owns that lens.
- R4-R4 — coverage limits: recipes that install, deploy, reinstall, re-seal, mutate shared podman storage or change live services were not executed. Remote work = one read-only fact/hash query; CLI execution used fake hosts or isolated local fixtures. No full gate, vulnerability scanners, online pin freshness checks, mutation sweep or CI dispatch run. Go LSP probe returned “No LSP server available for file type: .go”; symbol tracing used source search + tests.
- R4-R5 — requirements coverage: X7 + K11 are prerequisites, not app-enforced version gates. They have integration/build coverage, not dedicated tests proving every minimum-version combination. No new unsupported minimum-version claim is made.
- R4-R6 — unproven runtime failures: real isolated K/X1 batch returned rc 1 (53 pass events, 17 fail events including subtests, zero skips). Failures included 6 s readiness timeouts, 5 s fake-host rendezvous timeout, Podman `invalid internal status, unable to create a new pause process: open /tmp/offload-e2e-podman-345871988/run/vfs-containers/.../userdata/conmon.pid: no such file or directory`, and a 28.9 s teardown versus the test's 20 s bound. Podman=6.1.4. Narrow `-parallel=1` rerun of K7 status, K8 signals, K10 parallel hosts + X1 nested run passed (7/7 events; zero skips). Fanout/contention is an inference, not established cause; other failed cases were not rerun. No additional product defect assigned from the batch. MAIN notified; no shared `podman system migrate`, restart or global cleanup performed. Owned failed-store residue removed after helper/storage-linked process searches returned rc 1. Full outputs: `.scratch/claims-ctr-x1.json`, `.scratch/claims-ctr-serial.json`.

## Predicate witness map

Representative existing tests per id; all implementations read. This maps id coverage, not exhaustive subclause coverage. Executed sets + limits follow; tests named outside them were read rather than rerun.

- C1 `TestC1FlagBoundary`; C2 `TestC2ConflictsAndPoolImplications`; C3 `TestC3ExitPrecedence`, waiting/setup signal witnesses; C4 `TestC4OutputChannels`; C5 `TestC5PlacementLine`.
- P1 `TestPathPrecedence`; P2 `TestLoadDefaultsSlots`, `TestLoadMultiplex`, `TestLoadRejects`, `TestL1Eligibility/local`; P3 `TestLoadMissingFile`, `TestP3MissingPool`.
- R1 `TestParse`, `TestR1ParseBoundaries`, `TestReviewR1ProbeDeadlineWithInheritedOutput`; R2 `TestR2Status`, `TestR2StatusShowsProbeNumbers`, slot-holder tests.
- S1 `TestHeavySlotsAgreesWithHeavy`; S2 heavy↔offload + lowest-slot + killed-holder tests; S3 `TestCommandNeverInheritsLock`, `TestLockReachesWatcherNotCommand`; S4 `TestS4SameTreeRunsSerializeAtController`, `TestReviewS4PoolOwnerNamesHost`; S5 watcher-lock + `TestX1ControllerDeathKillsSession/local-kill-escalation`; S6 `TestS6LockDirsRefusePlantedComponents`, `TestS6LockDirsAcceptOwnPrivateTree` + e2e lock-dir cases.
- L1 `TestL1Eligibility`; L2 `TestL2Placement`, `TestL2SignalDuringPoolProbeTakesNoLocalSlot`; L3 `TestL3WaitingSignals`, `TestL3DownHostProbedAtMostEvery30s`; L4 `TestL4NoGPU`, `TestL4DownOrBusyGPUHostsWait`.
- Y1 `TestInspect`, `TestY1RootAndSubdirectory`, damaged-repo + home-root tests; Y2 `TestWorkdirName`, `TestY2WorkdirIdentity`; Y3 `TestY3UnsupportedGitLayouts`; Y4 `TestUploadList`, `TestY4GitUploadDifferential`, same-second/same-stat + linked-worktree/metadata tests; Y5 `TestRemotePrune`, framing mismatch + `TestY5Prune`, global-excludes + symlink ancestor tests; Y6 `TestY6Download`, same-stat + output-list failures; Y7 `TestY7PlainMirror`, same-stat tests; Y8 `TestGetPath`, `TestY8Get`, newer-local/existing-directory tests; Y9 `TestY9NoSync`, create-workdir case; Y10 upload/download/prune failure tests; Y11 `TestRemoteLockSerializes`, `TestY11RemoteLock`, live-predecessor case.
- X1 session death/leader exit/scope refusal/admission/daemon + nested-run tests; X2 `TestLocalArgvIsNotShellParsed`, `TestX2ArgvAndEnvironment`, `TestLookMatchesBash` + R4-F4 red; X3 `TestX3SignalsSkipDownload`, cancel/signal tests; X4 `TestX4ByteExactStreaming`; X5 `TestTargetCarriesOptions`, `TestMultiplex`, `TestX5MultiplexOptIn`, fake-ssh option guard; X6 `TestX6Remote255AndTransport`, run-only transport failures; X7 prerequisite integration (R4-R5); X8 boot integrity/install failure tests + `TestX8RemoteProcessShowsScriptPath`.
- J1 state path/modes/record shape, concurrent writes, failed-write rollback + FIFO tests; J2 local/pool/phase/status/signal/rounding/version tests + R4-F3 red; J3 rotation/default/max-size tests; J4 log failure transparency, lock timeout, write errors + final-status fallback; J5 `TestJ5KilledOffloadLeavesStartAlone` (per-machine nested logging traced, not separately exercised).
- K1 `TestK1CtrUsage`, `TestK1HelpWinsOverBadSwitch`; K2 switch-off exec identity + invalid-switch/nonexecutable/ENOEXEC cases + R4-F4 red; K3 Find/missing/symlink/damaged/colon tests; K4 tag equal/change/type/mode/unreadable tests; K5 build reuse/concurrent lock/cancel/death tests; K6 argv/mount/home/ownership/env/worktree/streams + binary identity tests; K7 readRecord/missing peer/status/lookup/FD3 tests; K8 control signals/PID1 guard/liveness/stragglers/systemd scope + signal-probe tests; K9 Layers/podman differential/failure tests; K10 framed parsing/order/binary/host failures/parallel host tests; K11 static TestMain + real rootless podman execution (R4-R5).

## Execution record

- Build: `CGO_ENABLED=0 go build -o .scratch/offload ./cmd/offload` → rc 0. Built before the red-test commits; production revision 640f24f.
- Direct CLI: no args; top/run/status/ctr help; missing command; local/pool conflict; missing pool; unknown verb; ctr layers/check conflict; bad switch; native ctr literal argv; heavy no args/-h/--help/--status. Outputs + rc in `.scratch/usage.log`; all expected. No real-pool run launched.
- Packages, separately: `go test -count=1 ./internal/{pool,cli,probe,slots,mirror,remote,shell,ctr,run}` (one package per invocation) → each rc 0 before its respective red test was added; ledger originals `go test -count=1 -run '^TestJ' ./internal/ledger` → rc 0. Run package 93.712 s; other per-package outputs in `.scratch/claims-packages.log` (pool/cli/probe output in transcript). Quiet package output does not report individual skip counts.
- CLI e2e subset: `go test -count=1 -run '^(TestC[1-5].*|TestP3MissingPool|TestR2.*|TestL[1-4].*|TestY4LinkedWorktreeRunsRemotely|TestY9NoSync.*)$' ./test/e2e` → rc 0, 43.996 s (`.scratch/claims-cli.log`).
- Sync/log/exec subset: `go test -count=1 -run '^(TestJ.*|TestY.*|TestX[2-8].*|TestS4SameTreeRunsSerializeAtController)$' ./test/e2e` → rc 0, 54.943 s (`.scratch/claims-sync-log.log`).
- Real K/X1 batch: `OFFLOAD_REQUIRE_PODMAN=1 OFFLOAD_REQUIRE_SCOPE=1 go test -json -count=1 -run '^(TestK.*|TestX1.*)$' ./test/e2e` → rc 1, 145.095 s; R4-R6 above.
- Narrow rerun: `OFFLOAD_REQUIRE_PODMAN=1 OFFLOAD_REQUIRE_SCOPE=1 go test -json -count=1 -parallel=1 -run '^(TestK7StatusPassesThrough|TestK8SignalReachesCommandOnce|TestK10ReviewHostsStartInParallel|TestX1NestedRunEndsWithOuterRun)$' ./test/e2e` → rc 0, 7 pass events, zero skips. Does not erase the batch failure or verify the remaining failed cases.
- Text recipes: `.agent/tool --pins` → 11 metadata rows, rc 0; `.agent/spec-lint` → rc 0, same spec + planted Phase history → rc 1 naming history. Fake transport marker: required X5 options, `.probe`=`8 2 1048576 0`, `.transport-fail` present; fake ssh `nproc` → four fields/rc 0, `true` → simulated reset/rc 255.
- Live non-mutating check: controller/pool installed offload + heavy SHA256 pairs equal; pool config multiplex=true; host CPU/memory/tool versions/power/BIOS/idle-slice settings queried, rc 0. No secret files read.
- Deliberate reds: R4-F3 + R4-F4 each rc 1 on unchanged production; commands/output/commits in their findings. No production fix attempted.
- Completion: 27/27 rows adjudicated, zero unknown; 7 findings = 2 major behavioral + 5 minor text. Production unchanged; temporary CLI/fake-host fixtures removed, evidence logs retained under the worktree's `.scratch/` for MAIN. Failed-batch fixture roots observed in its log no longer exist.

reviewer-4-DONE-1
