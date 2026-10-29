# reviewer-5 — verification integrity @ 640f24f

## Verdict table

| row | check | verdict | finding ids | evidence |
|---|---|---|---|---|
| V1 | m1 cli (`internal/cli` + e2e `cli_test.go`) | finding | R5-F1, R5-F5 | 16 tests: 12 red, 4 uncredited. Complete function inventory below; text anchors = each declaration. |
| V2 | m2 run (unit + e2e execution/close/failures/locks) | finding | R5-F2, R5-F5, R5-F6 | 64 tests: 5 mutation, 43 red, 16 uncredited (7 selected-plant only). Complete function inventory below. |
| V3 | m3 mirror (unit + mutation + e2e sync) | finding | R5-F5, R5-F6 | 50 tests: 19 mutation, 29 red, 2 uncredited (selected-plant only). Complete function inventory below. |
| V4 | m4 slots | finding | R5-F5 | 17 tests: 14 mutation, 3 red. Later MakeDirs survivor attribution unconfirmed (E1). Complete function inventory below. |
| V5 | m5 pool | pass | — | 6 tests: 5 mutation, 1 red. Complete function inventory below; text anchors = each declaration. |
| V6 | m6 probe | pass | — | 6 tests: 5 mutation, 1 red. Complete function inventory below; text anchors = each declaration. |
| V7 | m7 remote | finding | R5-F6, R5-F5 | 5 tests: 1 mutation, 1 red, 3 uncredited. Complete function inventory below; text anchors = each declaration. |
| V8 | m8 ctr (unit + real-podman e2e) | finding | R5-F5 | 80 tests: 39 mutation, 41 red. Complete function inventory below; text anchors = each declaration. |
| V9 | m9 ledger (unit + e2e `ledger_test.go`) | finding | R5-F7 | 32 tests: 11 red, 21 mutation-credit records incomplete: 13 survivors ruled by the test author alone. No repeat of already recorded kills requested. |
| V10 | m10 shell | finding | R5-F3, R5-F5 | 1 tests: 1 uncredited. Complete function inventory below; text anchors = each declaration. |
| V11 | m11 e2e harness self-tests (fake ssh X5 refusal, `-n`, markers) | finding | R5-F5 | 2 tests: 2 red. Complete function inventory below; text anchors = each declaration. |
| V12 | m12 tools (`pooladdrs` test; `heavy` + `install` coverage) | finding | R5-F4 | 1 tests: 1 uncredited. Complete function inventory below; text anchors = each declaration. |

## Scope + evidence protocol

- Fixed before target read: seeded V1–V12; credit, test-detection/fixture-special-case scan, skips/deletions/demotions, contract citations. Source target stayed 640f24f; primary implementation excluded.
- Inventory: 282 `func Test` declarations = 280 test functions + 2 TestMain hooks. `rg -n '^func Test' <worktree>/internal <worktree>/test/e2e <worktree>/tools`; every test assigned once below. Cross-cutting archived-review tests assigned by behavior; e2e execution/close/failures/locks stay V2, except the fake-SSH self-test (V11).
- Per-test history: `git -C <worktree> log --all -S<TestName> --format='%h %s%n%b' -- <test-file>`; additionally all commit bodies and committed review/mutation records. All 282 traces inspected by record family; full-history searches close the uncredited list. No history pagination or search cap.
- R = recorded red on an unfixed revision; M = complete mutation-credit record; uncredited includes selected-plant-only evidence or a campaign whose survivors lack independent rulings. Hand plants on embedded shell retain their value, but do not substitute for default operators over the Go files the test calls. Accepted IMPLEMENT R24/R26 and pool-environment R17 rulings retained. R/M credit historical evidence, not a rerun here. Some older bodies say “pre-fix” without a command/SHA; those record-quality limits remain explicit, not repaired by invented execution.
- Test-only production branch scan: whole implementations read/scanned; `OFFLOAD_FAKE_HOSTS`, `OFFLOAD_REVIEW_*`, `SLOTS_TEST_*`, `-test.run`, testing.Testing occur only in test/harness code (0 production matches). `run(find, …)` and `ledger.openLog` inject dependencies, not test detection. Production constants map contract choices; no fixture expected-output table or test-result shortcut found.
- Runtime checks in this review: none. No mutago, full gate, CI, real-pool run or live-host write. Findings below are missing-evidence/text conformance findings, not newly demonstrated live behavior failures. LSP on internal/cli/cli.go → no Go server; source/rg/history used.
- Paths in tables = repo-relative @640f24f; resolve against `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-5`. Deliverable = validator input, not target source. Scratch evidence is read-derived only.

## Findings

### R5-F1 — major — CLI unit tests lack verification credit

Engineering verification-integrity law covers each test, not merely a nearby e2e test of similar behavior. The fc3c575 red command selects `./test/e2e/`; it cannot credit internal/cli units. K10’s banner fix has a live/e2e red witness, but no recorded red/default sweep for the new parseCheck unit. This leaves usage, option boundaries and nonce parsing asserted without the required sensitivity evidence.

| target test @640f24f | guarded implementation | evidence gap |
|---|---|---|
| `internal/cli/cli_test.go:10` `TestHelpPrintsUsageOnStdout` | `internal/cli/cli.go` | no test-specific red record or applicable mutation record found; introduced be205eb |
| `internal/cli/cli_test.go:20` `TestNoArgumentsIsAUsageError` | `internal/cli/cli.go` | no test-specific red record or applicable mutation record found; introduced be205eb |
| `internal/cli/cli_test.go:30` `TestParseRun` | `internal/cli/cli.go` | no test-specific red record or applicable mutation record found; introduced fc3c575 |
| `internal/cli/cli_test.go:58` `TestParseCheckTakesOnlyNonceFramedLines` | `internal/cli/ctr.go` | no test-specific red record or applicable mutation record found; introduced 81a86d5 |

Text proof: exact distinctive anchor for each row = `func <TestName>(t *testing.T) {`; `rg -Fn` matches the named file at the listed line. Empty-history claims use positive controls: fc3c575's explicit 36-function **e2e** red record and a18ba8c's explicit pool/slots/mirror/probe sweeps both exist; neither names the above missing scope. Full `git log --all` and each `-S` trace were searched, not only the introducing body.

Acceptance: Run the stack default mutation tool over internal/cli sources (include cli.go, place.go, ctr.go despite test filenames), replay survivors through unit/e2e checks, independently rule every survivor, and record scope/command/results. Preserve existing checks.

### R5-F2 — major — run white-box tests lack verification credit

C3, X1/X2/X3, S3/S5 and Y5/Y6 guard status, teardown, inherited descriptors/environment and remote tree changes. Their green result is not credit without a recorded red or mutation kill. fc3c575 explicitly credits the **e2e** suite plus selected unit reds; that does not credit all run_test.go functions. bd8c3fe explicitly calls the two reviewer9 additions “green, pin behavior”; its four recorded plants concern NUL/mv/socket/tail, not these two assertions. reviewer-9 reports their baseline passes, not their failing counterexamples.

| target test @640f24f | guarded implementation | evidence gap |
|---|---|---|
| `internal/run/run_test.go:38` `TestLocalStatus` | `internal/run/run.go + watch.bash + local.bash` | no test-specific red record or applicable mutation record found; introduced fc3c575 |
| `internal/run/run_test.go:70` `TestCancelEndsWholeSession` | `internal/run/run.go + watch.bash` | no test-specific red record or applicable mutation record found; introduced fc3c575 |
| `internal/run/run_test.go:103` `TestStragglersEndWithLeader` | `internal/run/run.go + watch.bash` | no test-specific red record or applicable mutation record found; introduced fc3c575 |
| `internal/run/run_test.go:112` `TestLockReachesWatcherNotCommand` | `internal/run/run.go + watch.bash` | no test-specific red record or applicable mutation record found; introduced fc3c575 |
| `internal/run/run_test.go:534` `TestRemotePrune` | `internal/run/remote.bash + run.go` | no test-specific red record or applicable mutation record found; introduced fc3c575 |
| `internal/run/run_test.go:601` `TestRemoteRefusesMissingGit` | `internal/run/remote.bash` | no test-specific red record or applicable mutation record found; introduced fc3c575 |
| `internal/run/run_test.go:610` `TestRemoteDownList` | `internal/run/remote.bash` | no test-specific red record or applicable mutation record found; introduced fc3c575 |
| `internal/run/reviewer9_test.go:58` `TestReviewer9WatcherSignalsRecordStatus` | `internal/run/watch.bash + remote.bash + run.go` | no test-specific red record or applicable mutation record found; introduced bd8c3fe |
| `internal/run/reviewer9_test.go:102` `TestReviewer9RemoteKeepsCallerVariables` | `internal/run/remote.bash + run.go` | no test-specific red record or applicable mutation record found; introduced bd8c3fe |

Text proof: exact distinctive anchor for each row = `func <TestName>(t *testing.T) {`; `rg -Fn` matches the named file at the listed line. Empty-history claims use positive controls: fc3c575's explicit 36-function **e2e** red record and a18ba8c's explicit pool/slots/mirror/probe sweeps both exist; neither names the above missing scope. Full `git log --all` and each `-S` trace were searched, not only the introducing body.

Acceptance: Default mutation sweep over internal/run/run.go plus recorded targeted shell plants for the exact uncredited watcher/remote predicates; demonstrate each test’s observable kill, independently rule survivors, record commands + target revision.

### R5-F3 — minor — shell lookup unit lacks verification credit

X2/K2/K7 depend on this shared lookup. 81a86d5 records e2e PATH-nonexecutable reds, not a red execution of the newly introduced shell unit; no shell mutation sweep is recorded. TestLookMatchesBash uses a literal expected table, not execution of Bash as an independent oracle (the table itself is reasonable and not alleged wrong).

| target test @640f24f | guarded implementation | evidence gap |
|---|---|---|
| `internal/shell/look_test.go:9` `TestLookMatchesBash` | `internal/shell/look.go` | no test-specific red record or applicable mutation record found; introduced 81a86d5 |

Text proof: exact distinctive anchor for each row = `func <TestName>(t *testing.T) {`; `rg -Fn` matches the named file at the listed line. Empty-history claims use positive controls: fc3c575's explicit 36-function **e2e** red record and a18ba8c's explicit pool/slots/mirror/probe sweeps both exist; neither names the above missing scope. Full `git log --all` and each `-S` trace were searched, not only the introducing body.

Acceptance: Mutate internal/shell/look.go with default operators, run its unit + affected e2e tests, record/rule all survivors. Add its contract citations (F5).

### R5-F4 — minor — pooladdrs unit lacks verification credit

The old sed parser’s three-host fixture was demonstrated red (81a86d5 / pool-environment R6). The new Go test calls only new `run`, not the old install parser, and its own red/default mutation record is absent. The manual fixture is useful evidence, but does not prove this test detects regressions in the replacement helper. tools/install live deployments and heavy differential coverage remain credited separately.

| target test @640f24f | guarded implementation | evidence gap |
|---|---|---|
| `tools/pooladdrs/main_test.go:10` `TestEveryValidPoolSpellingYieldsItsAddress` | `tools/pooladdrs/main.go + tools/install` | no test-specific red record or applicable mutation record found; introduced 81a86d5 |

Text proof: exact distinctive anchor for each row = `func <TestName>(t *testing.T) {`; `rg -Fn` matches the named file at the listed line. Empty-history claims use positive controls: fc3c575's explicit 36-function **e2e** red record and a18ba8c's explicit pool/slots/mirror/probe sweeps both exist; neither names the above missing scope. Full `git log --all` and each `-S` trace were searched, not only the introducing body.

Acceptance: Mutate tools/pooladdrs/main.go with default operators, retain the TOML fixture, record a test kill + independently ruled survivors. Do not change grading expectations.

### R5-F6 — minor — selected plants do not supply the required default-operator sweep

These tests never have a recorded unfixed-revision red. Their selected hand plants are useful sensitivity evidence, but Engineering requires the stack’s default mutation operators over the files the test imports/calls. The changed Go transport/target and run mechanisms have no recorded default sweep. Embedded Bash needs targeted shell plants **in addition to**, not instead of, the Go-file sweep. IMPLEMENT R26’s accepted no-source fallback and pool-environment R17’s accepted combined campaign remain credited at their recorded scope; those rulings do not cover these later changes.

| target test @640f24f | guarded implementation | selected-plant record; missing sweep |
|---|---|---|
| `internal/remote/remote_test.go:36` `TestTargetCarriesOptions` | `internal/remote/ssh.go` | 2dbc582; no default sweep of changed transport |
| `internal/remote/remote_test.go:48` `TestMultiplex` | `internal/remote/ssh.go` | 2dbc582 + reviewer-9; no default sweep of Multiplex |
| `test/e2e/reviewer9_test.go:44` `TestReviewer9MultiplexFallbackOnceAcrossHosts` | `internal/cli/cli.go` targets + `internal/remote/ssh.go` | reviewer-9/bd8c3fe; no default sweep of targets |
| `internal/run/run_test.go:233` `TestRefusedScopeRunsCommandWithoutScope` | `internal/run/run.go + watch.bash + remote.bash` | e5f45a5 + reviewer-8 K7 fallback plant; selected-plant only |
| `internal/run/run_test.go:302` `TestWatcherTermBeforeSetsidRunsNothing` | `internal/run/run.go + watch.bash + remote.bash` | 128af08; trap/default-who plants; selected-plant only |
| `internal/run/run_test.go:491` `TestBootRefusesShortOrCorruptScript` | `internal/run/run.go + watch.bash + remote.bash` | 7807b19 + reviewer-9 R7 hash/cleanup/error plants; selected-plant only |
| `internal/run/run_test.go:510` `TestBootReplacesWrongScriptFile` | `internal/run/run.go + watch.bash + remote.bash` | 7807b19 + reviewer-9 R7 shared-name plant; selected-plant only |
| `internal/run/run_test.go:520` `TestBootFailureNeverRunsAStaleScript` | `internal/run/run.go + watch.bash + remote.bash` | 7807b19 + reviewer-9 R7 boot error plant; selected-plant only |
| `internal/run/run_test.go:574` `TestRemoteFramingMismatchPrunesNothing` | `internal/run/run.go + watch.bash + remote.bash` | fc3c575; reviewer-4 framing-guard red replay; selected-plant only |
| `test/e2e/execution_test.go:186` `TestX1RefusedScopeFallsBackQuietly` | `internal/run/run.go + watch.bash + remote.bash` | e5f45a5; reviewer-8 K7 stderr-suppression plant; selected-plant only |
| `test/e2e/reviewer9_test.go:29` `TestReviewer9ExcludesRecordIsNotAnUploadPath` | `internal/run/run.go + remote.bash; internal/mirror/mirror.go; internal/cli/cli.go` | bd8c3fe; reviewer-9 F5 independent +2 survivor ruling/red; selected-plant only |
| `test/e2e/sync_test.go:202` `TestY5GlobalExcludesReachTheHost` | `internal/run/run.go + remote.bash; internal/mirror/mirror.go; internal/cli/cli.go` | 7807b19 excludes-empty/read-wrong-record plants; reviewer-9 R7; selected-plant only |

Text proof: exact distinctive anchor for every row = `func <TestName>(t *testing.T) {`; `rg -Fn` matches its file/line. 2dbc582 calls the two unit checks “new API” and records six selected plants. e5f45a5, 128af08, 7807b19 and reviewer-8/-9 record the listed shell plants, not a default sweep of their Go callers. Full commit-history searches found default sweeps for pool/slots/mirror/probe/ctr/ledger, not the missing transport/run/CLI scope.

Acceptance: default-operator sweeps over `internal/{remote,run,cli}` and the changed mirror paths reached by these tests; configure source discovery so differing test filenames do not omit files. Replay survivors through the named tests, independently rule each survivor, record commands/scope/results. Preserve targeted shell plants. Existing kills need no redundant rerun merely to rename their evidence.

### R5-F5 — minor — contract IDs absent from 21 test names/comments

`.claude/rules/code.md:9` requires “tests cite ids in names or comments.” These functions have no C/P/R/S/L/Y/X/J/K predicate in either their name or their own comments. Review-finding labels G1/G3/G5/G7/V6 are not contract IDs; an ID only inside an error string is not a name/comment citation. Traceability matters when deciding which survivor violates specified behavior. tools/pooladdrs is outside this rule’s path scope and is not counted.

| text anchor @640f24f | exact distinctive bytes (`rg -Fn`) |
|---|---|
| `internal/cli/cli_test.go:10` | `func TestHelpPrintsUsageOnStdout(t *testing.T) {` |
| `internal/cli/cli_test.go:20` | `func TestNoArgumentsIsAUsageError(t *testing.T) {` |
| `internal/mirror/mirror_test.go:214` | `func TestGlobalExcludes(t *testing.T) {` |
| `internal/remote/remote_test.go:12` | `func TestJoinRoundTripsThroughBash(t *testing.T) {` |
| `internal/remote/reviewer9_test.go:9` | `func TestReviewer9UnwritableSocketDirectoryRefused(t *testing.T) {` |
| `internal/run/reviewer9_test.go:14` | `func TestReviewer9BootDestinationDirectoryFailsClosed(t *testing.T) {` |
| `internal/run/reviewer9_test.go:58` | `func TestReviewer9WatcherSignalsRecordStatus(t *testing.T) {` |
| `internal/run/reviewer9_test.go:102` | `func TestReviewer9RemoteKeepsCallerVariables(t *testing.T) {` |
| `internal/shell/look_test.go:9` | `func TestLookMatchesBash(t *testing.T) {` |
| `internal/slots/slots_test.go:215` | `func TestSlotFilesNeverFollowSymlinks(t *testing.T) {` |
| `internal/slots/slots_test.go:236` | `func TestSlotFilesNeverBlockOnPlantedFIFO(t *testing.T) {` |
| `test/e2e/review_ctr_test.go:21` | `func TestReviewG6TerminalStdinCompletes(t *testing.T) {` |
| `test/e2e/reviewer4_probe_test.go:16` | `func TestReviewV6ProbeFailureIsNotAbsence(t *testing.T) {` |
| `test/e2e/reviewer9_test.go:9` | `func TestReviewer9GlobalExcludesNULPreservesUpload(t *testing.T) {` |
| `test/e2e/reviewer9_test.go:29` | `func TestReviewer9ExcludesRecordIsNotAnUploadPath(t *testing.T) {` |
| `test/e2e/reviewer9_test.go:44` | `func TestReviewer9MultiplexFallbackOnceAcrossHosts(t *testing.T) {` |
| `test/e2e/reviewer_guarantees_test.go:13` | `func TestG1CancellationBeforeSetsidKillsSession(t *testing.T) {` |
| `test/e2e/reviewer_guarantees_test.go:67` | `func TestG3DownloadListFailureIsNotSuccess(t *testing.T) {` |
| `test/e2e/reviewer_guarantees_test.go:74` | `func TestG5UploadTypeChangeKeepsIgnoredRemoteFile(t *testing.T) {` |
| `test/e2e/reviewer_guarantees_test.go:91` | `func TestG7SignalDuringStatusLookupCancels(t *testing.T) {` |
| `test/e2e/reviewer_guarantees_test.go:115` | `func TestG7SignalDuringInspectionWins(t *testing.T) {` |

Proof = each exact declaration matched once; surrounding function comments read. Positive control: `TestParseRun` carries `// C1 + C2`; `TestS6LockDirsRefusePlantedComponents` cites S6 by name. Acceptance: add the correct contract IDs to these function names/comments, without weakening assertions; harness-only assertion may cite the production predicate it protects (X5/K8).

### R5-F7 — minor — ledger survivors ruled by the test author alone

Targets: `internal/ledger/ledger.go:131,239,263,264,272,275,283,284,292,305,307` @640f24f; tests listed below. Engineering requires a contract ruling by someone other than the test author for **every** survivor. 192f8bc records 186 mutants (146 unit kills + 27 e2e replay kills), but its 13 survivors were ruled equivalent by the test author alone. MAIN confirms the unit’s reviewer-1 reviewed the closing diff, not those survivors. Thus this is now a finding, not the prior unconfirmed E1 entry. The existing mutation execution/kill evidence remains valid; the missing operation is independent adjudication.

Record command: `git -C <worktree> show -s --format='%B' 192f8bc`. Exact excerpt: `13 survivors, ruled equivalent: ReadBuildInfo !ok (module builds always carry it, :131), record()/json.Marshal error ignored (cannot fail for these types, :239 :263 :264×2), Stat on an open fd failing (:272 :275 :305 :307), fd close skipped on error paths (:283 :284; process exits), close error dropped (:292×2; not injectable).` The same body assigns reviewer-1 only the four closing-diff findings and their re-review. `git diff --quiet 192f8bc 640f24f -- internal/ledger/ledger.go` → rc 0: sites unchanged.

| survivor | recorded site @640f24f | author’s equivalence reason | exact source anchor (`rg -Fn`; enclosing line where needed) |
|---|---|---|---|
| L1 | `internal/ledger/ledger.go:131` | ReadBuildInfo !ok; module builds always carry it | `info, ok := debug.ReadBuildInfo()` (:129) |
| L2 | `internal/ledger/ledger.go:239` | json.Marshal error cannot occur for these types | `return append(line, '\n'), merr` (:239) |
| L3 | `internal/ledger/ledger.go:263` | record error cannot occur for these types | `line, err := record()` (:262) |
| L4 | `internal/ledger/ledger.go:264` (first mutant) | record error cannot occur for these types | `line, err := record()` (:262) |
| L5 | `internal/ledger/ledger.go:264` (second mutant) | record error cannot occur for these types | `line, err := record()` (:262) |
| L6 | `internal/ledger/ledger.go:272` | Stat on open fd cannot fail | `info, err := f.Stat()` (:271) |
| L7 | `internal/ledger/ledger.go:275` | Stat on open fd cannot fail | `if err == nil && size > 0 && size+int64(len(line)) > maxSize {` (:275) |
| L8 | `internal/ledger/ledger.go:305` | Stat on open fd cannot fail | `!info.Mode().IsRegular() {` (:305) |
| L9 | `internal/ledger/ledger.go:307` | Stat on open fd cannot fail | `!info.Mode().IsRegular() {` (:305) |
| L10 | `internal/ledger/ledger.go:283` | skipped close on error; process exits | `if f != nil {` (:283) |
| L11 | `internal/ledger/ledger.go:284` | skipped close on error; process exits | `if f != nil {` (:283) |
| L12 | `internal/ledger/ledger.go:292` (first mutant) | dropped close error not injectable | `if cerr := f.Close(); err == nil {` (:291) |
| L13 | `internal/ledger/ledger.go:292` (second mutant) | dropped close error not injectable | `if cerr := f.Close(); err == nil {` (:291) |

The body does not preserve each mutant operator or exact replacement bytes; L4/L5 and L12/L13 are distinct survivors sharing sites, not fabricated distinct replacements. That detail is unconfirmed; recover it from the campaign artifacts or regenerate only what independent adjudication needs. “Not injectable” alone does not establish contract equivalence. No ruling here asserts these mutations are wrong or equivalent.

Affected never-red tests (all guard `internal/ledger/ledger.go`; e2e also calls `internal/cli/cli.go`):

| test @640f24f | credit status |
|---|---|
| `internal/ledger/ledger_test.go:44` `TestJ1StateBase` | campaign/kill evidence recorded; survivor ruling incomplete |
| `internal/ledger/ledger_test.go:61` `TestJ2WriterPassesBytesAndKeepsLines` | campaign/kill evidence recorded; survivor ruling incomplete |
| `internal/ledger/ledger_test.go:84` `TestJ2WriterReturnsStderrError` | campaign/kill evidence recorded; survivor ruling incomplete |
| `internal/ledger/ledger_test.go:91` `TestJ3J4Defaults` | campaign/kill evidence recorded; survivor ruling incomplete |
| `internal/ledger/ledger_test.go:97` `TestJ2SecondsKeepMilliseconds` | campaign/kill evidence recorded; survivor ruling incomplete |
| `internal/ledger/ledger_test.go:107` `TestJ3RotatesPastMaxSize` | campaign/kill evidence recorded; survivor ruling incomplete |
| `internal/ledger/ledger_test.go:143` `TestJ3OneByteLogIsNonEmpty` | campaign/kill evidence recorded; survivor ruling incomplete |
| `internal/ledger/ledger_test.go:163` `TestJ4FailuresReportTheirCause` | campaign/kill evidence recorded; survivor ruling incomplete |
| `internal/ledger/ledger_test.go:206` `TestJ4WriteErrorSurfaces` | campaign/kill evidence recorded; survivor ruling incomplete |
| `internal/ledger/ledger_test.go:224` `TestJ4OneFailedWriteNotesOnce` | campaign/kill evidence recorded; survivor ruling incomplete |
| `internal/ledger/ledger_test.go:236` `TestJ4LockHeldElsewhereFailsWithinTimeout` | campaign/kill evidence recorded; survivor ruling incomplete |
| `internal/ledger/ledger_test.go:273` `TestJ2EndRecordSecondsAndStatus` | campaign/kill evidence recorded; survivor ruling incomplete |
| `internal/ledger/ledger_test.go:296` `TestJ2VCSVersion` | campaign/kill evidence recorded; survivor ruling incomplete |
| `internal/ledger/ledger_test.go:321` `TestJ1LockHeldAroundEveryOpen` | campaign/kill evidence recorded; survivor ruling incomplete |
| `internal/ledger/ledger_test.go:395` `TestJ4EndReturnsFinalStatusWhenTheWriteFails` | campaign/kill evidence recorded; survivor ruling incomplete |
| `test/e2e/ledger_test.go:284` `TestJ1UsageErrorWritesNothing` | campaign/kill evidence recorded; survivor ruling incomplete |
| `test/e2e/ledger_test.go:352` `TestJ2NoSyncTimesTheCommandAlone` | campaign/kill evidence recorded; survivor ruling incomplete |
| `test/e2e/ledger_test.go:366` `TestJ2BusyMatchesPlacementLine` | campaign/kill evidence recorded; survivor ruling incomplete |
| `test/e2e/ledger_test.go:380` `TestJ1NoAbsoluteStateDirLogsNothing` | campaign/kill evidence recorded; survivor ruling incomplete |
| `test/e2e/ledger_test.go:443` `TestJ1ConcurrentRunsKeepEveryRecord` | campaign/kill evidence recorded; survivor ruling incomplete |
| `test/e2e/ledger_test.go:471` `TestJ2RemotePhasesPerMode` | campaign/kill evidence recorded; survivor ruling incomplete |

Acceptance: MAIN independently adjudicates L1–L13 against J1–J4, records a reason and the actual mutation for each, and adds a killing test only where specified behavior is lost. Reuse the 173 recorded kills; rerun only missing/reconstructed evidence. A completed independent ruling restores these tests’ mutation credit without rewriting the old commit.

## Complete credit inventory

### V1 — m1 cli (`internal/cli` + e2e `cli_test.go`)

| test / hook @640f24f | classification | record |
|---|---|---|
| `internal/cli/cli_test.go:10` `TestHelpPrintsUsageOnStdout` | uncredited | R5-F1 |
| `internal/cli/cli_test.go:20` `TestNoArgumentsIsAUsageError` | uncredited | R5-F1 |
| `internal/cli/cli_test.go:30` `TestParseRun` | uncredited | R5-F1 |
| `internal/cli/cli_test.go:58` `TestParseCheckTakesOnlyNonceFramedLines` | uncredited | R5-F1 |
| `test/e2e/cli_test.go:14` `TestC1FlagBoundary` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/cli_test.go:45` `TestC2ConflictsAndPoolImplications` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/cli_test.go:68` `TestC3ExitPrecedence` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/cli_test.go:116` `TestC4OutputChannels` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/cli_test.go:138` `TestC5PlacementLine` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/cli_test.go:159` `TestP3MissingPool` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/cli_test.go:189` `TestR2Status` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/cli_test.go:220` `TestL1Eligibility` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/cli_test.go:268` `TestL2Placement` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/cli_test.go:302` `TestL2SignalDuringPoolProbeTakesNoLocalSlot` | red | 061dc2f, base b91c323; pool probe unreachable |
| `test/e2e/cli_test.go:322` `TestL3WaitingSignals` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/cli_test.go:362` `TestL4NoGPU` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |

### V2 — m2 run (unit + e2e execution/close/failures/locks)

| test / hook @640f24f | classification | record |
|---|---|---|
| `internal/run/reviewer9_test.go:14` `TestReviewer9BootDestinationDirectoryFailsClosed` | red | bd8c3fe; reviewer-9 F2 target 547e86b |
| `internal/run/reviewer9_test.go:58` `TestReviewer9WatcherSignalsRecordStatus` | uncredited | R5-F2 |
| `internal/run/reviewer9_test.go:102` `TestReviewer9RemoteKeepsCallerVariables` | uncredited | R5-F2 |
| `internal/run/run_test.go:38` `TestLocalStatus` | uncredited | R5-F2 |
| `internal/run/run_test.go:47` `TestCommandOutlivesOpenStdin` | red | fc3c575; reviewer-4 red replay (cat stdin) |
| `internal/run/run_test.go:54` `TestLocalArgvIsNotShellParsed` | red | e5f45a5; reviewer-8 expansion red |
| `internal/run/run_test.go:70` `TestCancelEndsWholeSession` | uncredited | R5-F2 |
| `internal/run/run_test.go:103` `TestStragglersEndWithLeader` | uncredited | R5-F2 |
| `internal/run/run_test.go:112` `TestLockReachesWatcherNotCommand` | uncredited | R5-F2 |
| `internal/run/run_test.go:200` `TestScopeEndsEscapedDaemonWithLeader` | red | e5f45a5; reviewer-8 K7 base replay |
| `internal/run/run_test.go:211` `TestScopeEndsEscapedDaemonOnCancel` | red | e5f45a5; reviewer-8 K7 base replay |
| `internal/run/run_test.go:233` `TestRefusedScopeRunsCommandWithoutScope` | uncredited | R5-F6 — selected-plant only: e5f45a5 + reviewer-8 K7 fallback plant |
| `internal/run/run_test.go:244` `TestScopeHoldsCommand` | red | e5f45a5; reviewer-8 K7 base replay |
| `internal/run/run_test.go:254` `TestScopeKeepsCommandEnvironment` | red | e5f45a5; reviewer-8 K7 base replay |
| `internal/run/run_test.go:280` `TestWatcherKeepsCallerVariables` | red | e5f45a5; reviewer-8 F1 |
| `internal/run/run_test.go:302` `TestWatcherTermBeforeSetsidRunsNothing` | uncredited | R5-F6 — selected-plant only: 128af08; trap/default-who plants |
| `internal/run/run_test.go:340` `TestScopeMarkLivesInRuntimeDirUntilTeardown` | red | 6e06a39; base bd8c3fe |
| `internal/run/run_test.go:356` `TestScopeKillsTermIgnoringDaemonAfterGrace` | red | e5f45a5; reviewer-8 K7 base replay |
| `internal/run/run_test.go:380` `TestEOFDuringScopeAdmissionRunsNothing` | red | e5f45a5; reviewer-8 K7 base replay |
| `internal/run/run_test.go:491` `TestBootRefusesShortOrCorruptScript` | uncredited | R5-F6 — selected-plant only: 7807b19 + reviewer-9 R7 hash/cleanup/error plants |
| `internal/run/run_test.go:510` `TestBootReplacesWrongScriptFile` | uncredited | R5-F6 — selected-plant only: 7807b19 + reviewer-9 R7 shared-name plant |
| `internal/run/run_test.go:520` `TestBootFailureNeverRunsAStaleScript` | uncredited | R5-F6 — selected-plant only: 7807b19 + reviewer-9 R7 boot error plant |
| `internal/run/run_test.go:534` `TestRemotePrune` | uncredited | R5-F2 |
| `internal/run/run_test.go:574` `TestRemoteFramingMismatchPrunesNothing` | uncredited | R5-F6 — selected-plant only: fc3c575; reviewer-4 framing-guard red replay |
| `internal/run/run_test.go:601` `TestRemoteRefusesMissingGit` | uncredited | R5-F2 |
| `internal/run/run_test.go:610` `TestRemoteDownList` | uncredited | R5-F2 |
| `internal/run/run_test.go:637` `TestRemoteLockSerializes` | red | fc3c575; reviewer-4 stdin red replay; R21 oracle repaired |
| `test/e2e/close_test.go:15` `TestX6TransportFailureDuringRunIs125` | mutation | 69861da targeted survivors + IMPLEMENT R18–R20 |
| `test/e2e/close_test.go:28` `TestX6TransportFailureDuringNoSyncRunIs125` | mutation | 69861da ErrNotStarted→nil killed |
| `test/e2e/close_test.go:35` `TestR2StatusShowsProbeNumbers` | mutation | 69861da targeted survivors + IMPLEMENT R18–R20 |
| `test/e2e/close_test.go:53` `TestL3DownHostProbedAtMostEvery30s` | mutation | 69861da targeted survivors + IMPLEMENT R18–R20 |
| `test/e2e/close_test.go:71` `TestS4SameTreeRunsSerializeAtController` | mutation | 69861da targeted survivors + IMPLEMENT R18–R20 |
| `test/e2e/close_test.go:86` `TestY1HomeDirectoryNeverGoesToAHost` | red | 0a42d5f red-before home-sync statement |
| `test/e2e/close_test.go:113` `TestX5MultiplexOptIn` | red | 2dbc582, base e5f45a5; unknown multiplex key |
| `test/e2e/execution_test.go:58` `TestX1ControllerDeathKillsSession` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/execution_test.go:114` `TestX1LeaderExitReapsBackground` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/execution_test.go:153` `TestX1ScopeEndsDoubleForkedDaemon` | red | e5f45a5; reviewer-8 K7 base watcher replay |
| `test/e2e/execution_test.go:186` `TestX1RefusedScopeFallsBackQuietly` | uncredited | R5-F6 — selected-plant only: e5f45a5; reviewer-8 K7 stderr-suppression plant |
| `test/e2e/execution_test.go:205` `TestX1NestedRunEndsWithOuterRun` | red | 128af08, base 7807b19; nested command alive |
| `test/e2e/execution_test.go:254` `TestX8RemoteProcessShowsScriptPath` | red | 7807b19, base 2dbc582; embedded process title |
| `test/e2e/execution_test.go:273` `TestX2ArgvAndEnvironment` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/execution_test.go:305` `TestX3SignalsSkipDownload` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/execution_test.go:334` `TestX4ByteExactStreaming` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/execution_test.go:358` `TestX6Remote255AndTransport` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/failures_test.go:14` `TestY1GitErrorIsNotPlainMode` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/failures_test.go:28` `TestY8GetKeepsNewerLocal` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/failures_test.go:42` `TestY9NoSyncCreatesWorkdir` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/failures_test.go:53` `TestY10DownloadFailureOverridesCommand` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/failures_test.go:70` `TestY11LivePredecessorAfterControllerDeath` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/failures_test.go:113` `TestL4DownOrBusyGPUHostsWait` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/locks_test.go:20` `TestS6PlantedHeavyDirIs125` | red | a12f9eb, base a910536; TestS6 command |
| `test/e2e/locks_test.go:30` `TestS6PlantedHostDirIs125` | red | a12f9eb, base a910536; TestS6 command |
| `test/e2e/locks_test.go:40` `TestS6PlantedCtrLockDirIs125` | red | a12f9eb, base a910536; TestS6 command |
| `test/e2e/locks_test.go:57` `TestS6ForeignOwnedHeavyDirIs125` | red | a12f9eb, base a910536; TestS6 command |
| `test/e2e/reviewer1_test.go:14` `TestReviewP2ExplicitZeroSlots` | red | 3c38235, base 46aa85e; archived reviewer-1/-3/-4 reds |
| `test/e2e/reviewer1_test.go:23` `TestReviewR2FreshStatusHasNoHolders` | red | 3c38235, base 46aa85e; archived reviewer-1/-3/-4 reds |
| `test/e2e/reviewer1_test.go:38` `TestReviewC3SignalDuringInspection` | red | 3c38235, base 46aa85e; archived reviewer-1/-3/-4 reds |
| `test/e2e/reviewer1_test.go:77` `TestReviewL4PreviouslyUpHostGoesDown` | red | 3c38235, base 46aa85e; archived reviewer-1/-3/-4 reds |
| `test/e2e/reviewer1_test.go:97` `TestReviewS4PoolOwnerNamesHost` | red | 3c38235, base 46aa85e; archived reviewer-1/-3/-4 reds |
| `test/e2e/reviewer1_test.go:129` `TestReviewX1DeathBeforeSessionCreation` | red | 3c38235, base 46aa85e; archived reviewer-1/-3/-4 reds |
| `test/e2e/reviewer1_test.go:176` `TestReviewX3SignalDuringStatusRecovery` | red | 3c38235, base 46aa85e; archived reviewer-1/-3/-4 reds |
| `test/e2e/reviewer_guarantees_test.go:13` `TestG1CancellationBeforeSetsidKillsSession` | red | 3c38235, base 46aa85e; archived reviewer-1/-3/-4 reds |
| `test/e2e/reviewer_guarantees_test.go:91` `TestG7SignalDuringStatusLookupCancels` | red | 3c38235, base 46aa85e; archived reviewer-1/-3/-4 reds |
| `test/e2e/reviewer_guarantees_test.go:115` `TestG7SignalDuringInspectionWins` | red | 3c38235, base 46aa85e; archived reviewer-1/-3/-4 reds |

### V3 — m3 mirror (unit + mutation + e2e sync)

| test / hook @640f24f | classification | record |
|---|---|---|
| `internal/mirror/mirror_test.go:48` `TestUploadList` | mutation | a18ba8c / IMPLEMENT R26, mirror sweep + reviewer-6 |
| `internal/mirror/mirror_test.go:102` `TestUploadListRefusesNestedRepo` | mutation | a18ba8c / IMPLEMENT R26, mirror sweep + reviewer-6 |
| `internal/mirror/mirror_test.go:118` `TestInspect` | mutation | a18ba8c / IMPLEMENT R26, mirror sweep + reviewer-6 |
| `internal/mirror/mirror_test.go:176` `TestInspectRefusesBrokenRepo` | red | 7b96bf0; reviewer-4 damaged-repo replay |
| `internal/mirror/mirror_test.go:186` `TestGetPath` | mutation | a18ba8c / IMPLEMENT R26, mirror sweep + reviewer-6 |
| `internal/mirror/mirror_test.go:201` `TestWorkdirName` | mutation | a18ba8c / IMPLEMENT R26, mirror sweep + reviewer-6 |
| `internal/mirror/mirror_test.go:214` `TestGlobalExcludes` | mutation | a18ba8c / IMPLEMENT R26, mirror sweep + reviewer-6; bd8c3fe NUL plant |
| `internal/mirror/mirror_test.go:237` `TestRootsThatCopyEverythingStayLocal` | red | 0a42d5f red-before statement |
| `internal/mirror/mutation_test.go:31` `TestMirrorY1GitFailurePlainTree` | mutation | a18ba8c / IMPLEMENT R26, mirror sweep + reviewer-6 |
| `internal/mirror/mutation_test.go:38` `TestMirrorY1RevParseFraming` | mutation | a18ba8c / IMPLEMENT R26, mirror sweep + reviewer-6 |
| `internal/mirror/mutation_test.go:45` `TestMirrorY1DamagedAncestor` | mutation | a18ba8c / IMPLEMENT R26, mirror sweep + reviewer-6 |
| `internal/mirror/mutation_test.go:57` `TestMirrorY1IndexFailure` | mutation | a18ba8c / IMPLEMENT R26, mirror sweep + reviewer-6 |
| `internal/mirror/mutation_test.go:65` `TestMirrorY3SplitGitDirs` | mutation | a18ba8c / IMPLEMENT R26, mirror sweep + reviewer-6 |
| `internal/mirror/mutation_test.go:93` `TestMirrorY3NestedReplacesTracked` | mutation | a18ba8c / IMPLEMENT R26, mirror sweep + reviewer-6 |
| `internal/mirror/mutation_test.go:109` `TestMirrorY4ListErrors` | mutation | a18ba8c / IMPLEMENT R26, mirror sweep + reviewer-6 |
| `internal/mirror/mutation_test.go:121` `TestMirrorY4LstatError` | mutation | a18ba8c / IMPLEMENT R26, mirror sweep + reviewer-6 |
| `internal/mirror/mutation_test.go:143` `TestMirrorY4IgnoredSymlink` | mutation | a18ba8c / IMPLEMENT R26, mirror sweep + reviewer-6 |
| `internal/mirror/mutation_test.go:161` `TestMirrorY4AfterSymlinkAncestor` | mutation | a18ba8c / IMPLEMENT R26, mirror sweep + reviewer-6 |
| `internal/mirror/mutation_test.go:183` `TestMirrorY5ExcludesErrors` | mutation | a18ba8c / IMPLEMENT R26, mirror sweep + reviewer-6 |
| `internal/mirror/mutation_test.go:200` `TestMirrorY3UnreadableMetadata` | mutation | a18ba8c / IMPLEMENT R26, mirror sweep + reviewer-6 |
| `test/e2e/mirror_mutation_test.go:11` `TestMirrorY10FileUploadFailure` | mutation | a18ba8c / IMPLEMENT R26, mirror mi18/mi28/mi66 |
| `test/e2e/review_test.go:8` `TestReviewY10OutputListFailureOverridesCommand` | red | 3c38235, base 46aa85e; archived reviewer-1/-3/-4 reds |
| `test/e2e/reviewer1_test.go:54` `TestReviewY6DownloadListFailure` | red | 3c38235, base 46aa85e; archived reviewer-1/-3/-4 reds |
| `test/e2e/reviewer1_test.go:64` `TestReviewY8GetExistingDirectory` | red | 3c38235, base 46aa85e; archived reviewer-1/-3/-4 reds |
| `test/e2e/reviewer1_test.go:111` `TestReviewY10FailedPrunePreventsCommand` | red | 3c38235, base 46aa85e; archived reviewer-1/-3/-4 reds |
| `test/e2e/reviewer9_test.go:9` `TestReviewer9GlobalExcludesNULPreservesUpload` | red | bd8c3fe; reviewer-9 F1 target 547e86b |
| `test/e2e/reviewer9_test.go:29` `TestReviewer9ExcludesRecordIsNotAnUploadPath` | uncredited | R5-F6 — selected-plant only: bd8c3fe; reviewer-9 F5 independent +2 survivor ruling/red |
| `test/e2e/reviewer_guarantees_test.go:67` `TestG3DownloadListFailureIsNotSuccess` | red | 3c38235, base 46aa85e; archived reviewer-1/-3/-4 reds |
| `test/e2e/reviewer_guarantees_test.go:74` `TestG5UploadTypeChangeKeepsIgnoredRemoteFile` | red | 3c38235, base 46aa85e; archived reviewer-1/-3/-4 reds |
| `test/e2e/sync_test.go:18` `TestY1RootAndSubdirectory` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/sync_test.go:33` `TestY2WorkdirIdentity` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/sync_test.go:51` `TestY3UnsupportedGitLayouts` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/sync_test.go:142` `TestY4GitUploadDifferential` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/sync_test.go:202` `TestY5GlobalExcludesReachTheHost` | uncredited | R5-F6 — selected-plant only: 7807b19 excludes-empty/read-wrong-record plants; reviewer-9 R7 |
| `test/e2e/sync_test.go:220` `TestY4SameSecondRewriteCrosses` | red | 2a85b31, base 501ec85; same-second upload |
| `test/e2e/sync_test.go:247` `TestY4SameStatRewriteCrosses` | red | ad9ee05, base 7dabc80; SameStatRewrite selector |
| `test/e2e/sync_test.go:260` `TestY6SameStatRewriteReturns` | red | ad9ee05, base 7dabc80; SameStatRewrite selector |
| `test/e2e/sync_test.go:272` `TestY7SameStatRewriteCrosses` | red | ad9ee05, base 7dabc80; SameStatRewrite selector |
| `test/e2e/sync_test.go:285` `TestY7SameStatRewriteReturns` | red | 5c11b58; mirror.go at 7dabc80 |
| `test/e2e/sync_test.go:302` `TestY4LinkedWorktreeRunsRemotely` | red | 7bb4c90, base 2a85b31; linked worktree refused |
| `test/e2e/sync_test.go:349` `TestY4WorktreeRefusesAHostGitDirectory` | red | 7bb4c90, base 2a85b31; linked worktree refused |
| `test/e2e/sync_test.go:362` `TestY4GitMetadataMirrorDeletes` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/sync_test.go:372` `TestY5Prune` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/sync_test.go:393` `TestY5PruneSymlinkAncestor` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/sync_test.go:412` `TestY6Download` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/sync_test.go:464` `TestY7PlainMirror` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/sync_test.go:487` `TestY8Get` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/sync_test.go:517` `TestY9NoSync` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/sync_test.go:530` `TestY10UploadFailure` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |
| `test/e2e/sync_test.go:540` `TestY11RemoteLock` | red | fc3c575, base 282d82f; archive/tester-1 36-function replay |

### V4 — m4 slots

| test / hook @640f24f | classification | record |
|---|---|---|
| `internal/slots/reviewer7_test.go:12` `TestS1HeavyRejectsInvalidCounts` | mutation | a18ba8c / IMPLEMENT R26, slots sweep + reviewer-7 |
| `internal/slots/reviewer7_test.go:20` `TestS1RuntimeDirFallback` | mutation | a18ba8c / IMPLEMENT R26, slots sweep + reviewer-7 |
| `internal/slots/reviewer7_test.go:26` `TestS3ReleasePreservesSuccessorOwner` | mutation | a18ba8c / IMPLEMENT R26, slots sweep + reviewer-7 |
| `internal/slots/reviewer7_test.go:44` `TestR2HoldersScanOnlyConfiguredRange` | mutation | a18ba8c / IMPLEMENT R26, slots sweep + reviewer-7 |
| `internal/slots/reviewer7_test.go:67` `TestS4WorkdirLockCreatesParents` | mutation | a18ba8c / IMPLEMENT R26, slots sweep + reviewer-7 |
| `internal/slots/reviewer7_test.go:85` `TestS3OwnerWriteFailureReleasesSlot` | mutation | a18ba8c / IMPLEMENT R26, slots sweep + reviewer-7 |
| `internal/slots/slots_test.go:30` `TestHeavySlotsAgreesWithHeavy` | mutation | a18ba8c / IMPLEMENT R26, slots sweep + reviewer-7 |
| `internal/slots/slots_test.go:69` `TestSlotHeldByHeavyIsBusy` | mutation | a18ba8c / IMPLEMENT R26, slots sweep + reviewer-7 |
| `internal/slots/slots_test.go:99` `TestSlotHeldByOffloadIsBusyToHeavy` | mutation | a18ba8c / IMPLEMENT R26, slots sweep + reviewer-7 |
| `internal/slots/slots_test.go:135` `TestLowestFreeSlotFirst` | mutation | a18ba8c / IMPLEMENT R26, slots sweep + reviewer-7 |
| `internal/slots/slots_test.go:158` `TestCommandNeverInheritsLock` | mutation | a18ba8c / IMPLEMENT R26, slots sweep + reviewer-7 |
| `internal/slots/slots_test.go:183` `TestMain` | hook | setup, not a test |
| `internal/slots/slots_test.go:194` `TestKilledHolderFreesSlot` | mutation | a18ba8c / IMPLEMENT R26, slots sweep + reviewer-7 |
| `internal/slots/slots_test.go:215` `TestSlotFilesNeverFollowSymlinks` | red | 46aa85e; IMPLEMENT S6 |
| `internal/slots/slots_test.go:236` `TestSlotFilesNeverBlockOnPlantedFIFO` | red | 81a86d5; pool-environment S15, base 57903b5 |
| `internal/slots/slots_test.go:274` `TestS6LockDirsRefusePlantedComponents` | red | a12f9eb, base a910536 |
| `internal/slots/slots_test.go:311` `TestS6LockDirsAcceptOwnPrivateTree` | mutation | a12f9eb slots sweep + plants; independent-ruling record unconfirmed (register E1) |
| `internal/slots/slots_test.go:367` `TestS6MakeDirsBoundsAndBase` | mutation | a12f9eb slots sweep + plants; independent-ruling record unconfirmed (register E1) |

### V5 — m5 pool

| test / hook @640f24f | classification | record |
|---|---|---|
| `internal/pool/pool_test.go:13` `TestPathPrecedence` | mutation | a18ba8c / IMPLEMENT R26, pool sweep + reviewer-7 |
| `internal/pool/pool_test.go:38` `TestLoadDefaultsSlots` | mutation | a18ba8c / IMPLEMENT R26, pool sweep + reviewer-7 |
| `internal/pool/pool_test.go:48` `TestLoadMultiplex` | red | 2dbc582, base e5f45a5 |
| `internal/pool/pool_test.go:55` `TestLoadRejects` | mutation | a18ba8c / IMPLEMENT R26, pool sweep + reviewer-7 |
| `internal/pool/pool_test.go:84` `TestLoadMissingFile` | mutation | a18ba8c / IMPLEMENT R26, pool sweep + reviewer-7 |
| `internal/pool/reviewer7_test.go:5` `TestP2OneSlotAndDotName` | mutation | a18ba8c / IMPLEMENT R26, pool sweep + reviewer-7 |

### V6 — m6 probe

| test / hook @640f24f | classification | record |
|---|---|---|
| `internal/probe/probe_test.go:8` `TestParse` | mutation | a18ba8c / IMPLEMENT R26, probe sweep + reviewer-7 |
| `internal/probe/probe_test.go:23` `TestScriptOnThisMachine` | mutation | a18ba8c / IMPLEMENT R26, probe sweep + reviewer-7 |
| `internal/probe/reviewer1_test.go:16` `TestReviewR1ProbeDeadlineWithInheritedOutput` | red | 3c38235, base 46aa85e; 69861da bound plant |
| `internal/probe/reviewer7_test.go:8` `TestR1ParseBoundaries` | mutation | a18ba8c / IMPLEMENT R26, probe sweep + reviewer-7 |
| `internal/probe/reviewer7_test.go:22` `TestR1MalformedCounters` | mutation | a18ba8c / IMPLEMENT R26, probe sweep + reviewer-7 |
| `internal/probe/reviewer7_test.go:37` `TestR1MalformedReasonNamesRejectedValue` | mutation | a18ba8c / IMPLEMENT R26, probe sweep + reviewer-7 |

### V7 — m7 remote

| test / hook @640f24f | classification | record |
|---|---|---|
| `internal/remote/remote_test.go:12` `TestJoinRoundTripsThroughBash` | mutation | a18ba8c / IMPLEMENT R26 re1–re3 (accepted no-source fallback) |
| `internal/remote/remote_test.go:36` `TestTargetCarriesOptions` | uncredited | R5-F6 (selected plants exist; default sweep absent) |
| `internal/remote/remote_test.go:48` `TestMultiplex` | uncredited | R5-F6 (selected plants exist; default sweep absent) |
| `internal/remote/reviewer9_test.go:9` `TestReviewer9UnwritableSocketDirectoryRefused` | red | bd8c3fe; reviewer-9 F3 target 547e86b |
| `test/e2e/reviewer9_test.go:44` `TestReviewer9MultiplexFallbackOnceAcrossHosts` | uncredited | R5-F6 (selected plants exist; default sweep absent) |

### V8 — m8 ctr (unit + real-podman e2e)

| test / hook @640f24f | classification | record |
|---|---|---|
| `internal/ctr/ctr_test.go:54` `TestK4TagEqualForEqualTrees` | mutation | 7a4a2d0 / pool-environment R17, ctr sweep; ctr-reviewer-4 |
| `internal/ctr/ctr_test.go:64` `TestK4TagChangesWithEveryInput` | mutation | 7a4a2d0 / pool-environment R17, ctr sweep; ctr-reviewer-4 |
| `internal/ctr/ctr_test.go:99` `TestK4TagRefusesOtherFileTypes` | mutation | 7a4a2d0 / pool-environment R17, ctr sweep; ctr-reviewer-4 |
| `internal/ctr/ctr_test.go:107` `TestK4TagRefusesSymlinkedRoot` | red | 81a86d5; pool-environment R1 |
| `internal/ctr/ctr_test.go:116` `TestK6RunArgs` | red | 81a86d5; pool-environment R3 |
| `internal/ctr/ctr_test.go:137` `TestK6Home` | mutation | 7a4a2d0 / pool-environment R17, ctr sweep; ctr-reviewer-4 |
| `internal/ctr/ctr_test.go:152` `TestK3MountableRefusesColons` | mutation | 7a4a2d0 / pool-environment R17, ctr sweep; ctr-reviewer-4 |
| `internal/ctr/ctr_test.go:161` `TestK6RunningBinaryIdentity` | mutation | 7a4a2d0 / pool-environment R17, ctr sweep; ctr-reviewer-4 |
| `internal/ctr/ctr_test.go:176` `TestK7ReadRecord` | mutation | 7a4a2d0 / pool-environment R17, ctr sweep; ctr-reviewer-4 |
| `internal/ctr/ctr_test.go:201` `TestK7MissingRecordWithOpenPeerReturns` | red | 81a86d5; pool-environment R2 |
| `internal/ctr/reviewer4_faults_test.go:41` `TestK6ReviewUnavailableSelfIdentity` | mutation | 7a4a2d0 / pool-environment R17, ctr sweep; ctr-reviewer-4 |
| `internal/ctr/reviewer4_faults_test.go:62` `TestK3ReviewFindFailures` | mutation | 7a4a2d0 / pool-environment R17, ctr sweep; ctr-reviewer-4 |
| `internal/ctr/reviewer4_faults_test.go:92` `TestK7ReviewSocketpairFailure` | mutation | 7a4a2d0 / pool-environment R17, ctr sweep; ctr-reviewer-4 |
| `internal/ctr/reviewer4_test.go:55` `TestK3ReviewFind` | mutation | 7a4a2d0 / pool-environment R17, ctr sweep; ctr-reviewer-4 |
| `internal/ctr/reviewer4_test.go:78` `TestK6ReviewExternalGitDir` | mutation | 7a4a2d0 / pool-environment R17, ctr sweep; ctr-reviewer-4 |
| `internal/ctr/reviewer4_test.go:107` `TestK5ReviewEnsure` | mutation | 7a4a2d0 / pool-environment R17, ctr sweep; ctr-reviewer-4 |
| `internal/ctr/reviewer4_test.go:172` `TestK5ReviewWaitCancellation` | mutation | 7a4a2d0 / pool-environment R17, ctr sweep; ctr-reviewer-4 |
| `internal/ctr/reviewer4_test.go:225` `TestK6ReviewHomeFallback` | mutation | 7a4a2d0 / pool-environment R17, ctr sweep; ctr-reviewer-4 |
| `internal/ctr/reviewer4_test.go:239` `TestK9ReviewLayers` | mutation | 7a4a2d0 / pool-environment R17, ctr sweep; ctr-reviewer-4 |
| `internal/ctr/reviewer4_test.go:281` `TestK7ReviewRunBoundary` | mutation | 7a4a2d0 / pool-environment R17, ctr sweep; ctr-reviewer-4 |
| `internal/ctr/reviewer4_test.go:353` `TestK5ReviewSignalAfterEnsure` | mutation | 7a4a2d0 / pool-environment R17, ctr sweep; ctr-reviewer-4 |
| `internal/ctr/reviewer4_test.go:383` `TestK8ReviewNotify` | mutation | 7a4a2d0 / pool-environment R17, ctr sweep; ctr-reviewer-4 |
| `internal/ctr/reviewer4_test.go:396` `TestK4ReviewRootModeAndUnreadable` | mutation | 7a4a2d0 / pool-environment R17, ctr sweep; ctr-reviewer-4 |
| `internal/ctr/reviewer4_test.go:424` `TestK4ReviewSymlinkModesAreIgnored` | mutation | 7a4a2d0 / pool-environment R17, ctr sweep; ctr-reviewer-4 |
| `internal/ctr/reviewer4_test.go:470` `TestK4ReviewEveryPermissionBit` | mutation | 7a4a2d0 / pool-environment R17, ctr sweep; ctr-reviewer-4 |
| `test/e2e/ctr_test.go:160` `TestK1CtrUsage` | red | 0d14378, base 5c11b58; 29 K-functions replay (ctr-reviewer-4 V3) |
| `test/e2e/ctr_test.go:174` `TestK2SwitchOffExecsInPlace` | red | 0d14378, base 5c11b58; 29 K-functions replay (ctr-reviewer-4 V3) |
| `test/e2e/ctr_test.go:196` `TestK2BadSwitchValue` | red | 0d14378, base 5c11b58; 29 K-functions replay (ctr-reviewer-4 V3) |
| `test/e2e/ctr_test.go:205` `TestK3MissingDefinitionIsUsageError` | red | 0d14378, base 5c11b58; 29 K-functions replay (ctr-reviewer-4 V3) |
| `test/e2e/ctr_test.go:216` `TestK5BuildsOnceThenReusesTag` | red | 0d14378, base 5c11b58; 29 K-functions replay (ctr-reviewer-4 V3) |
| `test/e2e/ctr_test.go:238` `TestK5ConcurrentFirstRunsBuildOnce` | red | 0d14378, base 5c11b58; 29 K-functions replay (ctr-reviewer-4 V3) |
| `test/e2e/ctr_test.go:253` `TestK5SignalWhileWaitingForBuildLock` | red | 0d14378, base 5c11b58; 29 K-functions replay (ctr-reviewer-4 V3) |
| `test/e2e/ctr_test.go:321` `TestK6MountsHomeAndOwnership` | red | 0d14378, base 5c11b58; 29 K-functions replay (ctr-reviewer-4 V3) |
| `test/e2e/ctr_test.go:357` `TestK6WorktreeCommonDirMounted` | red | 0d14378, base 5c11b58; 29 K-functions replay (ctr-reviewer-4 V3) |
| `test/e2e/ctr_test.go:371` `TestK6StreamsAndStdinByteExact` | red | 0d14378, base 5c11b58; 29 K-functions replay (ctr-reviewer-4 V3) |
| `test/e2e/ctr_test.go:388` `TestK7StatusPassesThrough` | red | 0d14378, base 5c11b58; 29 K-functions replay (ctr-reviewer-4 V3) |
| `test/e2e/ctr_test.go:405` `TestK7NoStatusRecordIs125` | red | 0d14378, base 5c11b58; 29 K-functions replay (ctr-reviewer-4 V3) |
| `test/e2e/ctr_test.go:441` `TestK8SignalReachesCommandOnce` | red | 0d14378, base 5c11b58; 29 K-functions replay (ctr-reviewer-4 V3) |
| `test/e2e/ctr_test.go:461` `TestK8GroupWideSignalStillOnce` | red | 0d14378, base 5c11b58; 29 K-functions replay (ctr-reviewer-4 V3) |
| `test/e2e/ctr_test.go:497` `TestK8SigkillEndsContainer` | red | 0d14378, base 5c11b58; 29 K-functions replay (ctr-reviewer-4 V3) |
| `test/e2e/ctr_test.go:523` `TestK8LeftoverProcessesEndAfterCommand` | red | 0d14378, base 5c11b58; 29 K-functions replay (ctr-reviewer-4 V3) |
| `test/e2e/ctr_test.go:536` `TestK8UnderOffloadRunKillEndsContainer` | red | 0d14378, base 5c11b58; 29 K-functions replay (ctr-reviewer-4 V3) |
| `test/e2e/ctr_test.go:551` `TestK9LayersMatchPodman` | red | 0d14378, base 5c11b58; 29 K-functions replay (ctr-reviewer-4 V3) |
| `test/e2e/ctr_test.go:563` `TestK10CheckComparesHosts` | red | 0d14378, base 5c11b58; 29 K-functions replay (ctr-reviewer-4 V3) |
| `test/e2e/ctr_test.go:588` `TestK10CheckReportsDownHost` | red | 0d14378, base 5c11b58; 29 K-functions replay (ctr-reviewer-4 V3) |
| `test/e2e/ctr_test.go:603` `TestK1HelpWinsOverBadSwitch` | red | 0d14378, base 5c11b58; 29 K-functions replay (ctr-reviewer-4 V3) |
| `test/e2e/ctr_test.go:611` `TestK3DefinitionMustBeRegularFile` | red | 0d14378, base 5c11b58; 29 K-functions replay (ctr-reviewer-4 V3) |
| `test/e2e/ctr_test.go:623` `TestK3DamagedRepoIs125` | red | 0d14378, base 5c11b58; 29 K-functions replay (ctr-reviewer-4 V3) |
| `test/e2e/ctr_test.go:634` `TestK3ColonInPathIs125` | red | 0d14378, base 5c11b58; 29 K-functions replay (ctr-reviewer-4 V3) |
| `test/e2e/ctr_test.go:662` `TestK5ImageExistsFailureIs125` | red | 0d14378, base 5c11b58; 29 K-functions replay (ctr-reviewer-4 V3) |
| `test/e2e/ctr_test.go:675` `TestK5BuildKeepsLockPastOffloadDeath` | red | 0d14378, base 5c11b58; 29 K-functions replay (ctr-reviewer-4 V3) |
| `test/e2e/ctr_test.go:707` `TestK8SlotFreesAfterContainerLastWrite` | red | 0d14378, base 5c11b58; 29 K-functions replay (ctr-reviewer-4 V3) |
| `test/e2e/ctr_test.go:737` `TestK10CheckIgnoresLoginBannerAndReportsMissingOffload` | red | 0d14378, base 5c11b58; 29 K-functions replay (ctr-reviewer-4 V3) |
| `test/e2e/ctr_test.go:771` `TestK6RemoteWorktreeLayout` | red | 0d14378, base 5c11b58; 29 K-functions replay (ctr-reviewer-4 V3) |
| `test/e2e/ctr_test.go:788` `TestK3DefinitionDirMustBeRealDirectory` | red | 81a86d5 / pool-environment R1/R2/R4/R7/R11; archived reviewer reds |
| `test/e2e/ctr_test.go:803` `TestK2PathNonExecutableIs126` | red | 81a86d5 / pool-environment R1/R2/R4/R7/R11; archived reviewer reds |
| `test/e2e/ctr_test.go:815` `TestK7ImagePathNonExecutableIs126` | red | 81a86d5 / pool-environment R1/R2/R4/R7/R11; archived reviewer reds |
| `test/e2e/ctr_test.go:825` `TestK8UnderSystemdUnitConmonStaysOutOfX1Scope` | red | d8503a0 / pool-environment R20; independent replay in R22 |
| `test/e2e/review_ctr_test.go:21` `TestReviewG6TerminalStdinCompletes` | red | 81a86d5 / pool-environment R1/R2/R4/R7/R11; archived reviewer reds |
| `test/e2e/review_ctr_test.go:68` `TestReviewG6PodmanDeathIs125` | red | 81a86d5 / pool-environment R1/R2/R4/R7/R11; archived reviewer reds |
| `test/e2e/review_ctr_test.go:95` `TestReviewA3SignalCancelsProjectLookup` | red | 81a86d5 / pool-environment R1/R2/R4/R7/R11; archived reviewer reds |
| `test/e2e/review_ctr_test.go:127` `TestReviewQ1SignalCancelsScopeProbe` | red | 1d93774; ctr-reviewer-6 F1 target fa6936e |
| `test/e2e/reviewer4_ctr_test.go:22` `TestK3ReviewFindErrorsBeforePodman` | mutation | 7a4a2d0 / pool-environment R17; ctr-reviewer-4 default sweep + Init/CLI plants |
| `test/e2e/reviewer4_ctr_test.go:35` `TestK1ReviewNativeWordsAndENOEXEC` | mutation | 7a4a2d0 / pool-environment R17; ctr-reviewer-4 default sweep + Init/CLI plants |
| `test/e2e/reviewer4_ctr_test.go:51` `TestK7ReviewCommandStatusWithoutPodman` | mutation | 7a4a2d0 / pool-environment R17; ctr-reviewer-4 default sweep + Init/CLI plants |
| `test/e2e/reviewer4_ctr_test.go:62` `TestK6ReviewStreamsWithoutPodman` | mutation | 7a4a2d0 / pool-environment R17; ctr-reviewer-4 default sweep + Init/CLI plants |
| `test/e2e/reviewer4_ctr_test.go:81` `TestK8ReviewControlSignalWithoutPodman` | mutation | 7a4a2d0 / pool-environment R17; ctr-reviewer-4 default sweep + Init/CLI plants |
| `test/e2e/reviewer4_ctr_test.go:99` `TestK5ReviewCancelActiveBuild` | mutation | 7a4a2d0 / pool-environment R17; ctr-reviewer-4 default sweep + Init/CLI plants |
| `test/e2e/reviewer4_ctr_test.go:121` `TestK5ReviewBuildKeepsLockWithoutPodman` | mutation | 7a4a2d0 / pool-environment R17; ctr-reviewer-4 default sweep + Init/CLI plants |
| `test/e2e/reviewer4_ctr_test.go:160` `TestK8ReviewInitRefusesOutsidePID1` | mutation | 7a4a2d0 / pool-environment R17; ctr-reviewer-4 default sweep + Init/CLI plants |
| `test/e2e/reviewer4_ctr_test.go:169` `TestK9ReviewLayersFailureAndSwitch` | mutation | 7a4a2d0 / pool-environment R17; ctr-reviewer-4 default sweep + Init/CLI plants |
| `test/e2e/reviewer4_ctr_test.go:190` `TestK10ReviewOrderedLayersAndHosts` | mutation | 7a4a2d0 / pool-environment R17; ctr-reviewer-4 default sweep + Init/CLI plants |
| `test/e2e/reviewer4_ctr_test.go:225` `TestK10ReviewHostsStartInParallel` | mutation | 7a4a2d0 / pool-environment R17; ctr-reviewer-4 default sweep + Init/CLI plants |
| `test/e2e/reviewer4_ctr_test.go:257` `TestK10ReviewLocalFailuresBeforeRemote` | mutation | 7a4a2d0 / pool-environment R17; ctr-reviewer-4 default sweep + Init/CLI plants |
| `test/e2e/reviewer4_ctr_test.go:281` `TestK7ReviewImageENOEXECAndFD3` | mutation | 7a4a2d0 / pool-environment R17; ctr-reviewer-4 default sweep + Init/CLI plants |
| `test/e2e/reviewer4_ctr_test.go:290` `TestK8ReviewAllSignals` | mutation | 7a4a2d0 / pool-environment R17; ctr-reviewer-4 default sweep + Init/CLI plants |
| `test/e2e/reviewer4_ctr_test.go:308` `TestK8ReviewLeftoverTERM` | mutation | 7a4a2d0 / pool-environment R17; ctr-reviewer-4 default sweep + Init/CLI plants |
| `test/e2e/reviewer4_ctr_test.go:318` `TestK8ReviewWholeCommandGroup` | mutation | 7a4a2d0 / pool-environment R17; ctr-reviewer-4 default sweep + Init/CLI plants |
| `test/e2e/reviewer4_ctr_test.go:350` `TestK8ReviewLivenessEndsContainer` | mutation | 7a4a2d0 / pool-environment R17; ctr-reviewer-4 default sweep + Init/CLI plants |
| `test/e2e/reviewer4_ctr_test.go:375` `TestK8ReviewPodmanGoneFirst` | red | ctr-reviewer-4 VF3 target 57903b5; harvested 7a4a2d0 |

### V9 — m9 ledger (unit + e2e `ledger_test.go`)

| test / hook @640f24f | classification | record |
|---|---|---|
| `internal/ledger/ledger_test.go:44` `TestJ1StateBase` | uncredited | R5-F7 — 192f8bc kills recorded; 13 survivors ruled by test author alone |
| `internal/ledger/ledger_test.go:61` `TestJ2WriterPassesBytesAndKeepsLines` | uncredited | R5-F7 — 192f8bc kills recorded; 13 survivors ruled by test author alone |
| `internal/ledger/ledger_test.go:84` `TestJ2WriterReturnsStderrError` | uncredited | R5-F7 — 192f8bc kills recorded; 13 survivors ruled by test author alone |
| `internal/ledger/ledger_test.go:91` `TestJ3J4Defaults` | uncredited | R5-F7 — 192f8bc kills recorded; 13 survivors ruled by test author alone |
| `internal/ledger/ledger_test.go:97` `TestJ2SecondsKeepMilliseconds` | uncredited | R5-F7 — 192f8bc kills recorded; 13 survivors ruled by test author alone |
| `internal/ledger/ledger_test.go:107` `TestJ3RotatesPastMaxSize` | uncredited | R5-F7 — 192f8bc kills recorded; 13 survivors ruled by test author alone |
| `internal/ledger/ledger_test.go:143` `TestJ3OneByteLogIsNonEmpty` | uncredited | R5-F7 — 192f8bc kills recorded; 13 survivors ruled by test author alone |
| `internal/ledger/ledger_test.go:163` `TestJ4FailuresReportTheirCause` | uncredited | R5-F7 — 192f8bc kills recorded; 13 survivors ruled by test author alone |
| `internal/ledger/ledger_test.go:206` `TestJ4WriteErrorSurfaces` | uncredited | R5-F7 — 192f8bc kills recorded; 13 survivors ruled by test author alone |
| `internal/ledger/ledger_test.go:224` `TestJ4OneFailedWriteNotesOnce` | uncredited | R5-F7 — 192f8bc kills recorded; 13 survivors ruled by test author alone |
| `internal/ledger/ledger_test.go:236` `TestJ4LockHeldElsewhereFailsWithinTimeout` | uncredited | R5-F7 — 192f8bc kills recorded; 13 survivors ruled by test author alone |
| `internal/ledger/ledger_test.go:273` `TestJ2EndRecordSecondsAndStatus` | uncredited | R5-F7 — 192f8bc kills recorded; 13 survivors ruled by test author alone |
| `internal/ledger/ledger_test.go:296` `TestJ2VCSVersion` | uncredited | R5-F7 — 192f8bc kills recorded; 13 survivors ruled by test author alone |
| `internal/ledger/ledger_test.go:321` `TestJ1LockHeldAroundEveryOpen` | uncredited | R5-F7 — 192f8bc kills recorded; 13 survivors ruled by test author alone |
| `internal/ledger/ledger_test.go:352` `TestJ1FailedWriteLeavesNoPartialRecord` | red | 192f8bc reviewer short-write regression + no-truncate plant |
| `internal/ledger/ledger_test.go:395` `TestJ4EndReturnsFinalStatusWhenTheWriteFails` | uncredited | R5-F7 — 192f8bc kills recorded; 13 survivors ruled by test author alone |
| `test/e2e/ledger_test.go:154` `TestJ2LocalRunRecords` | red | 192f8bc, base ee0a79b; 8 J tests records=0 |
| `test/e2e/ledger_test.go:177` `TestJ2PoolRunRecords` | red | 192f8bc, base ee0a79b; 8 J tests records=0 |
| `test/e2e/ledger_test.go:201` `TestJ2SyncFailures` | red | 192f8bc, base ee0a79b; 8 J tests records=0 |
| `test/e2e/ledger_test.go:236` `TestJ2SignalRecorded` | red | 192f8bc, base ee0a79b; 8 J tests records=0 |
| `test/e2e/ledger_test.go:264` `TestJ2UnplacedRunWritesEndAlone` | red | 192f8bc, base ee0a79b; 8 J tests records=0 |
| `test/e2e/ledger_test.go:284` `TestJ1UsageErrorWritesNothing` | uncredited | R5-F7 — 192f8bc kills recorded; 13 survivors ruled by test author alone |
| `test/e2e/ledger_test.go:294` `TestJ5KilledOffloadLeavesStartAlone` | red | 192f8bc, base ee0a79b; 8 J tests records=0 |
| `test/e2e/ledger_test.go:309` `TestJ1DefaultPathAndModes` | red | 192f8bc, base ee0a79b; 8 J tests records=0 |
| `test/e2e/ledger_test.go:331` `TestJ4LogFailureLeavesRunUnchanged` | red | 192f8bc, base ee0a79b; 8 J tests records=0 |
| `test/e2e/ledger_test.go:352` `TestJ2NoSyncTimesTheCommandAlone` | uncredited | R5-F7 — 192f8bc kills recorded; 13 survivors ruled by test author alone |
| `test/e2e/ledger_test.go:366` `TestJ2BusyMatchesPlacementLine` | uncredited | R5-F7 — 192f8bc kills recorded; 13 survivors ruled by test author alone |
| `test/e2e/ledger_test.go:380` `TestJ1NoAbsoluteStateDirLogsNothing` | uncredited | R5-F7 — 192f8bc kills recorded; 13 survivors ruled by test author alone |
| `test/e2e/ledger_test.go:393` `TestJ4FIFOLogLeavesCommandRunnable` | red | 192f8bc reviewer FIFO/end-lock regression + recorded plants |
| `test/e2e/ledger_test.go:419` `TestJ2SignalDuringEndWriteCounts` | red | 192f8bc reviewer FIFO/end-lock regression + recorded plants |
| `test/e2e/ledger_test.go:443` `TestJ1ConcurrentRunsKeepEveryRecord` | uncredited | R5-F7 — 192f8bc kills recorded; 13 survivors ruled by test author alone |
| `test/e2e/ledger_test.go:471` `TestJ2RemotePhasesPerMode` | uncredited | R5-F7 — 192f8bc kills recorded; 13 survivors ruled by test author alone |

### V10 — m10 shell

| test / hook @640f24f | classification | record |
|---|---|---|
| `internal/shell/look_test.go:9` `TestLookMatchesBash` | uncredited | R5-F3 |

### V11 — m11 e2e harness self-tests (fake ssh X5 refusal, `-n`, markers)

| test / hook @640f24f | classification | record |
|---|---|---|
| `test/e2e/execution_test.go:384` `TestHarnessFakeSSH` | red | 3c38235 / IMPLEMENT R17; reviewer-4 F01 -n red; a5528fc stronger X5 checks |
| `test/e2e/harness_test.go:26` `TestMain` | hook | setup, not a test |
| `test/e2e/reviewer4_probe_test.go:16` `TestReviewV6ProbeFailureIsNotAbsence` | red | 7a4a2d0; ctr-reviewer-4 VF1 red on 57903b5 |

### V12 — m12 tools (`pooladdrs` test; `heavy` + `install` coverage)

| test / hook @640f24f | classification | record |
|---|---|---|
| `tools/pooladdrs/main_test.go:10` `TestEveryValidPoolSpellingYieldsItsAddress` | uncredited | R5-F4 |

## Skip/deletion/demotion adjudication

| site / case | adjudication | evidence |
|---|---|---|
| `test/e2e/ctr_test.go:64` requirePodman | pass, accepted ruling retained | pool-environment V4/R16 + code.md: absent podman may skip locally; OFFLOAD_REQUIRE_PODMAN makes absence fatal in CI; prebuild failure always fatal |
| `internal/run/run_test.go:136` + `test/e2e/execution_test.go:150` requireScope | pass of behavior; policy evidence limit E2 | reviewer-8 K8 accepted the explicit absent-manager handling and proved OFFLOAD_REQUIRE_SCOPE fatal; e5f45a5/b5b8a404 record conditional skip; no deferred/owner-waiver text located |
| `test/e2e/locks_test.go:63` foreign-owned fixture | pass of behavior; policy evidence limit E2 | explicit missing-podman skip; OFFLOAD_REQUIRE_PODMAN guard present and fatal; a12f9eb records red on pre-fix implementation; no deferred/owner-waiver text located |
| retired README counterexamples | pass, existing rulings retained | IMPLEMENT R13–R15: claimed behavior narrowed, counterexamples retired with those claims |
| two dropped ctr review checks | pass, existing rulings retained | pool-environment R17: pre-fix socket construction superseded by controlPair test; help-behind-bad-flag expectation retired with K1 R5 |
| TestX1RefusedScopeFallsBackQuietly conditional requireScope | pass of rationale; policy evidence limit E2 | b5b8a404: a refusal requires an actual manager; unsupported environment is not evidence against fallback behavior |
| other skip/delete/xfail/demotion paths | none found in target test tree | four Skip call sites total; no testing.Short, xfail, build-tag disable path found; source-history test removals checked against the above accepted records |

## Register / evidence limits

- **E1 — unconfirmed slots survivor adjudication only**: a12f9eb’s MakeDirs survivors have no named independent adjudicator (“Teammates: none”). Searched full git bodies, mutation-implement.md, ledger-*.md and committed reviewer reports; an independent ruling remains unconfirmed. The ledger portion is no longer here: MAIN confirms author-only rulings, now R5-F7. This register entry is not a claim that the slots survivors violate the contract.
- **E2 — unconfirmed skip authorization**: target deferred.md has six rows, none for unavailable systemd/podman tests; history records the technical reason/CI-required guards, but owner approval for these local skips was not found. Existing accepted review rulings remain in force; no attempt to turn their already-ruled environmental limitation into a new behavior finding. MAIN should retain an explicit waiver or deferred acceptance record under the current stricter clause.
- **E3 — accepted historical scope**: IMPLEMENT R26 and ctr R17 retain their accepted aggregate/default-sweep + hand-fallback rulings. Later selected-only plants are now R5-F6, not credited as a default sweep. No historical plants were rerun here.
- **E4 — tools coverage boundary**: heavy is the S1 differential oracle in internal/slots (tests in V4, original byte-copy hash + 4ea3d6a deployment witnesses); tools/install has live build/deploy/checksum witnesses be96752/4ea3d6a and the old-parser fixture red at 81a86d5. No dedicated automated install transaction test exists. This inventory is not a new requirement to mock real deployments; F4 concerns only the shipped Go test’s missing sensitivity credit.
- Historical reds with incomplete pre-fix SHA/command wording are not silently upgraded to fresh replays. This report identifies the actual record containing the claim. No baseline/full-suite health verdict is made.

## Validation

- Inventory validator → rc 0: expected 282 declarations, 282 exact `rg -Fn` matches at recorded lines; expected 282 unique inventory rows, got 280 tests + 2 hooks. Verdicts = 12 (10 finding, 2 pass); finding sections = 7 (2 major, 5 minor); citation anchors = 21/21 matched. Incomplete credit = 48 tests: 15 lack a red/mutation record, 12 have selected plants without the required default sweep, 21 ledger tests have recorded kills but author-only survivor rulings. Follow-up validation → rc 0: all 282 test declarations rematched; all 13 survivor anchors matched; selected-plant-only table = 12 tests; classification totals = 144 red + 88 mutation + 48 uncredited + 2 hooks; assigned completion marker remains final.
- Unknown-row validator: expected 0 → output `0`, grep rc 1 (no matches). Positive control: expected 1 seeded unresolved-verdict cell → `1`, rc 0. An initial final check matched this note’s literal sample cell; the note was reworded and the validator rerun. Historical-credit positive controls: fc3c575’s 36-function e2e record expected 1 → 1; a18ba8c’s named package sweep record expected 1 → 1.
- Production test-detection scan → no matches, rg rc 1; positive control OFFLOAD_FAKE_HOSTS in the fake-SSH fixture expected 1 file → 1, rc 0. Test-removal history scan: 0 committed test-function names removed and absent at target; retired review-only counterexamples remain governed by their accepted ledger rulings.
- `git diff --quiet 640f24f --` → rc 0; target `git status --porcelain` empty. No production, test, instruction, gate or configuration edit. No commits. Temporary inventory/history/validator files removed from this reviewer’s worktree; the complete inventory + proof commands remain in this deliverable.

reviewer-5-DONE-1
