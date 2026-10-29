# reviewer-7 — fix re-review

Subject = `wt/reviewer-7` @ `269d313b65bd5ac1eaf5d9bc598a6981f72f721d`; fixed set = the 28 seeded rows below, acceptance + Correctness on each fix diff/enclosing functions/callers. Security aspects → MAIN; no new whole-codebase review. Commands run from `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-7` unless stated. CI successes in acceptance cells = ledger claims, not independently rerun. No live nixbox / full gate / heavy mutation runs.

Complete = 28/28 rows: 21 pass · 4 findings (R7-F1–F4) · 3 MAIN security handoffs. Validator = 0 unknown cells (`grep -c` rc 1, expected for zero matches); four finding-detail sections. Review worktree clean, implementation bytes unchanged; red tests only in commits `5748fc5`, `bc350b3`, `02cb1a7`, `c60cad9`.

## Verdict table

| row | finding | fix | acceptance | verdict | evidence (file:line, command) |
|-----|---------|-----|------------|---------|-------------------------------|
| R2 | stale ignore entries + stale artifact (M1, M6) | 8e04230 | `.gitignore` = paths the repo makes; `.ruff_cache` gone; Gate CI 40284333001 success | pass | `.gitignore:1–8` + 8e04230 diff: Python-only entries removed; remaining paths = project outputs/local config/secrets. `test -e` for `.ruff_cache` in primary + review worktree → rc 1 both (positive file control: `.gitignore` read). |
| R3 | test step dominates the gate with no speed-up unit (M4) | 3fb8eec | `deferred.md` row `Gate speed` (accept: test step ≤ 30 s at equal strength or a floor measurement); Gate CI 40285054981 | pass | `.agent/deferred.md:9` contains the unit + equal-strength/floor acceptance; 3fb8eec diff read, `.agent/law` rc 0. Acceptance = scheduling the unit, not claiming its speedup achieved. |
| R4 | rule frontmatter unchecked (M14) | ec03608 | `.agent/law` frontmatter seed (unquoted glob) rc 1, clean rc 0; Gate CI 40284820501 | finding R7-F1 | Recorded unquoted-glob seed fires and clean fixture passes; malformed double-quoted YAML escapes pass the new regexp check (`.agent/law:25`). Red `TestR7LawRejectsUnparseableFrontmatter`. |
| R5 | human-facing filler words unchecked (M16) | ec03608 | `.agent/law` filler seed (`simply` in README) rc 1; Gate CI 40284820501 | pass | `.agent/law` clean rc 0; `TestR7LawSeeds/filler` verifies `simply` → rc 1 (combined seeds rc 0). `.agent/law:30` regexp + gate wiring read; no new behavior defect in this row's check. |
| R6 | deferred acceptance per entry unchecked (M22) | ec03608 | `.agent/law` deferred seed (entry without `Accept:`) rc 1; Gate CI 40284820501 | pass | `TestR7LawSeeds/deferred` verifies entry without `Accept:` → rc 1; clean rc 0. `.agent/law:32–36` flushes preceding + final entries, resets acceptance per entry, permits the existing `Accept (…):` form. |
| R7 | `.agent/spec-lint` behind the template (M24) | 4a7f959 | `copy-template --check` rc 0; Gate CI 40284341235 success | pass | `copy-template --check` rc 0 (`The template files are current.`); `.agent/spec-lint` rc 0. 4a7f959 = one suppression comment at `.agent/spec-lint:10`, no executable change. |
| R9 | `__` reservation guard covers a fixed name list (M26) | ec03608 | TestScriptVariablesAreReserved fires on 4 planted forms; Gate CI 40284820501 | finding R7-F3 | Baseline guard rc 0, recorded four forms covered by inspection; new names through `declare plain=1` + `plain+=suffix` pass. Red `TestR7ReservationGuardFindsDeclarationsAndAppends`, positive `plain=1` is rejected. `internal/run/run_test.go:679–713` omits declaration + append assignment grammar. |
| R8 | 21 tests cite no contract id + no check (M25, V1–V11) | ec03608 | ids on 23 tests; `.agent/law` test-ids seed rc 1; Gate CI 40284820501 | pass | ec03608's 23 comment additions read against surrounding tests; ids fit. `.agent/law` clean rc 0; `TestR7LawSeeds/test-ids` verifies untracked uncited test → rc 1. `.agent/law:38–53` reads current ids + tracked/untracked tests; new gate step reaches both full/fast paths. |
| R11 | end record + exit status can disagree (J2): `End` re-reads `final()` after the write | 2f01381 | TestR4J2EndRecordAndExitShareFinalStatus red → green; Gate CI 40280160938 success | pass | `go test -count=1 -run '^TestR4J2EndRecordAndExitShareFinalStatus$' ./internal/ledger/` rc 0. `internal/ledger/ledger.go:216–236`: frozen rc reused; pre-build failure still reads final status; write/appendLine + runCommand fit. |
| R12 | remote `cmd_rc` dropped under cancellation though the status came back (J2) | 2f01381 | TestReviewR1J2CancellationKeepsReturnedCommandStatus red → green; Gate CI 40280160938 success | pass | `go test -count=1 -run '^TestReviewR1J2CancellationKeepsReturnedCommandStatus$' ./test/e2e/` (combined R12/R14 regex) rc 0, both local/pool. `internal/cli/cli.go:311–320` retains nonambiguous status before cancellation return; `Remote.Run` + J2's 125/255 exception fit. |
| R13 | explicit empty PATH → 127 where bash runs the cwd command (X2, K2, K7) | 736bd7f | TestR4X2EmptyPATHSearchesCurrentDirectory red → green; binary `PATH= offload ctr -- zzc` rc 0; Gate CI 40280777077 success | pass | `go test -count=1 -run '^TestR4X2EmptyPATHSearchesCurrentDirectory$' ./internal/shell/` rc 0 (bash oracle). `internal/shell/look.go:28–34` restores one empty entry before the existing cwd branch; callers exec its path without a new PATH lookup. |
| R14 | hybrid run refuses a command the pool host has and the controller lacks (Decision `Trigger`, L1) | f4d1803 | TestReviewR1L2HybridCommandExistsOnlyOnHost + TestL1MissingLocalCommandUsesThePoolAlone red → green, guard plant killed; Gate CI 40280785143 success | pass | Both named tests in combined R12/R14 e2e regex rc 0. `internal/cli/cli.go:240–258` computes eligibility after Y3; `place.go:72,183` uses it for acquisition + waiting text; native ctr retains rc/message behavior. |
| R15 | remote prune inputs (sort/comm/excludes) unchecked → CMD runs on a wrong prune (Y5, Y10) | 70add2e | TestCRR2Y10PruneSortFailurePreventsCommand red → green; Gate CI 40282464839 success | pass | Named `./internal/run/` test rc 0. `internal/run/remote.bash:64–72,82–85` encloses list pipelines with pipefail + checked return; failure precedes prune/CMD. Checked down-list writes remove partial output; options remain subshell-local. |
| R16 | `efi-entry` treats a failed listing as no stale entries | 70add2e | TestCRR2EFIReadFailureStopsBeforeChanges red → green; Gate CI 40282464839 success | pass | Named `./machines/nixbox/` test rc 0, positive deletion/create control + failed-list zero changes. `machines/nixbox/efi-entry:15–17` standalone assignment exposes listing rc to `set -e`; parsing/run quoting unchanged. No live EFI mutation. |
| R17 | `check` collectors fail open: a failing `systemctl --failed` or closure scan reads as clean | 70add2e, 96fc907 | live probe: old collector fail-open rc 0, new rc 1 (systemctl, nix-store planted), facts unchanged, `check: ok`; EXIT trap (96fc907); Gate CI 40282464839 success | MAIN | Non-security control flow read: `machines/nixbox/check:27–35,77–79` launches Bash, enables pipefail, standalone `now` assignment, failed collection → rc 2; no defect found there. PSK scan/wrapper/security collectors unassessed; live acceptance not rerun (brief prohibits host access). |
| R18 | `--check` layer list truncated by a ≥ 64 KiB unframed line → false `same` (K10) | 7bb22d9 | TestReviewR1K10LongNoiseCannotTruncateLayers red → green; Gate CI 40282558813 success | pass | Named `./internal/cli/` test rc 0. `internal/cli/ctr.go:233–242` iterates the complete string, preserves ScanLines CR handling + ordered framing; ctrCheck compares the complete slice. |
| R19 | relative `core.excludesFile` read from the invocation cwd, not the root git used (Y5) | 6c7e953 | TestCRR2Y5RelativeExcludesFromSubdirectory red → green; Gate CI 40282699444 success | pass | Named `./internal/mirror/` test rc 0 with git check-ignore control. `internal/mirror/mirror.go:224–251` resolves nonempty relative paths against root, leaves absolute/empty paths and read errors intact; upload caller supplies sync root. |
| R20 | `status` counts a slot open error as held; owner text read blocks on a FIFO + follows links (R2, S15 class) | c88fa5c | status red + 2 new slots tests red → green (heavy parity 0400/0000); Gate CI 40282878239 success | pass | `go test -count=1 -run '^TestR2Holders' ./internal/slots/` rc 0 (both tests); status e2e in combined R20/R21 regex rc 0. `internal/slots/slots.go:149–190`: read-only lock probe + regular owner read; 0400/0000 parity preserved. Correctness-only; no security ruling. |
| R21 | exported `SHELLOPTS` (errexit…) changes the watcher's control flow (C3, X1) | 0f9c811 | 3 SHELLOPTS tests red → green (local, pool, scope + refused scope); Gate CI 40283202006 success | finding R7-F2 | Original e2e controls pass. New `TestR7X2ExportedShelloptsCannotSkipCommand/noexec` returns rc 0 without running CMD: inherited noexec prevents even the reset at `internal/run/watch.bash:12–15`. Same exported-option class remains. |
| R22 | fake ssh serves the probe answer to any command naming `nproc`/`loadavg` | 603e55b | TestReviewR1HarnessProbeWordsStayCommandData red → green; Gate CI 40283294678 success | pass | Named test in combined R22/R23 e2e regex rc 0. `test/e2e/testdata/ssh:35–47` recognizes the probe's leading script, not arbitrary substrings; actual CMD remains inside run.go's quoted boot command. |
| R23 | full-fake podman helper needs a real podman on PATH | 603e55b | TestR3FullFakePodmanNeedsNoInstalledPodman red → green; Gate CI 40283294678 success | pass | Named test in combined R22/R23 e2e regex rc 0, isolated empty PATH. `test/e2e/ctr_test.go:655–667` defers absence failure to fall-through; exhaustive fake runs without real podman (also R25's passing fake-podman test). |
| R24 | duplicate scope prerequisite helper + dead ctrhelper mode | 22a0dee | one `internal/testenv.RequireScope`; ctrhelper `id` gone; scope tests rc 0; Gate CI 40283456876 success | pass | `internal/testenv/scope.go:15–35` = relocated semantics; changed callers use it. `go test -v -count=1 -run '^(TestScopeHoldsCommand\|TestScopeKeepsCommandEnvironment)$' ./internal/run/` rc 0, both ran. ctrhelper id branch removed, no executable branch changed. |
| R25 | `ctr --check` SIGKILLs its child `offload run`s on a signal → start records without an end (J2, K10) | 42c18a8 | TestK10CheckSignalEndsHostRunsWithTheirEndRecords red → green; Gate CI 40282566022 success | pass | Named `./test/e2e/` test rc 0. `internal/cli/ctr.go:187–193` uses remote.Command; `internal/remote/ssh.go:102–107` TERM + 30 s grace + own process group preserves child teardown/log path. All goroutines joined before parent's signal return. |
| R26 | `deploy` runs `check` after `switch` alone (Decision: at every deploy) | 9eeba07 | TestDeployRunsCheckAfterEveryAction red → green; Gate CI 40282492093 success | pass | Named `./machines/nixbox/` test rc 0: switch/test/boot/build each reach stub check once. `machines/nixbox/deploy:12` unconditional after successful rebuild; failed rsync/rebuild still stops under set -e. No host contact. |
| R27 | controller scripts' ssh options diverge: deploy + tpm-bind lack BatchMode/forwarding off, install lacks accept-new | dfc3196 | every controller-script ssh carries BatchMode + forwarding off, install accept-new; Gate CI 40282500101 success | MAIN | Security options unassessed. Non-security diff read: `deploy:8–11` constant option splitting fits ssh/rsync -e; tpm-bind's pipeline/argv and install's stdin remain intact. Stub deploy/install acceptance tests rc 0; no SSH to a host. |
| R28 | `tools/install` leaves the per-side sha256 comparison to the reader | 30025c0 | tools/install_test.go red → green (lying host → rc 1); Gate CI 40283837639 success | finding R7-F4 | Original `TestInstallFailsWhenASideReportsAnotherSum` rc 0. New failed-digest test: sha256sum exits 23, masked by cut; source + installed digests empty, comparison accepts them and install exits 0. `tools/install:13,16`. |
| R29 | text vs behavior: usage pool tiers, C3 worktree clause, `--no-sync` text, `code.md` `.transport-fail`, 2 stale comments, L2 mechanism prose | ca26a8c | 7 anchors rewritten to the behavior; e2e C1/K1/P3 rc 0; Gate CI 40284292240 success | pass | ca26a8c diff read against pool.Path, remoteRun, tryPool, fake ssh, watcher fd + ctr session paths: corrected text fits. `go test -count=1 -run '^(TestC1FlagBoundary\|TestK1CtrUsage\|TestK1HelpWinsOverBadSwitch\|TestP3MissingPool)$' ./test/e2e/` rc 0. |
| R38 | ignored host copy + secret handling | 96fc907 | deletion + secret-copy fact/positive control + trap; rotation → owner | MAIN | Security handling unassessed. Non-security diff read: `machines/nixbox/check:57,70–71` EXIT cleanup and collector result follow existing Bash/fact-consumer plumbing; config.reviewed/readme/rule name the same new fact. Live deletion/scan/rotation unconfirmed here (host access forbidden). |

## Findings

### R7-F1 — R4 frontmatter check accepts invalid YAML

- Severity = minor. Location = `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-7/.agent/law:25` @ `269d313`.
- Contract = R4/M14: quoted globs **and a parse CC accepts**. The shape regexp accepts any nonquote bytes inside double quotes, including YAML's undefined escape `\q`; it never parses/validates string escapes. Rule seed `---\npaths:\n  - "internal/\q/**"\n---\n` passes with rc 0. A malformed rule can therefore close the gate without a usable `paths` selector — the same unchecked-frontmatter defect class remains.
- Red test = `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-7/tools/reviewer7_law_test.go`, `TestR7LawRejectsUnparseableFrontmatter`; fixture copies the target `.agent/law`, supplies every dependency, uses an isolated git repo. Positive controls: clean rc 0 + all four committed seeds rc 1 (`TestR7LawSeeds`, command rc 0).
- Command = `go test -count=1 -run '^TestR7LawRejectsUnparseableFrontmatter$' ./tools/`; rc 1 on target:

```text
--- FAIL: TestR7LawRejectsUnparseableFrontmatter (0.02s)
    reviewer7_law_test.go:36: invalid YAML escape \q: law rc=0, want 1; output=""
FAIL
FAIL github.com/eturkes/offload/tools 0.019s
```

- Independent YAML parse control: `printf '%s\n' 'paths:' '  - "internal/\q/**"' | /home/eturkes/Projects/offload/.tools/actionlint-1.7.12/actionlint -no-color -` → rc 1, `could not parse as YAML: found unknown escape character [syntax-check]` at line 2. Earlier Python oracle unavailable (`ModuleNotFoundError: No module named 'yaml'`); first actionlint attempt named its containing directory (rc 126), corrected to the binary above. No parser dependency installed.
- Red-test commit = `5748fc5`.

### R7-F2 — R21 inherited noexec still suppresses CMD

- Severity = major. Location = `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-7/internal/run/watch.bash:12–15` @ `269d313`; launch site `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-7/internal/run/run.go:56`.
- Contract = X2 + R21: inherited exported SHELLOPTS reaches CMD unchanged, while watcher scripts use default options. `SHELLOPTS=noexec` takes effect when Bash starts, so the in-script reset never executes. `offload run --local -- touch ran` returns **0 without starting touch**, leaving its marker absent. This is the same inherited-option control-flow defect class the fix addresses; resetting selected options after startup is insufficient for noexec.
- Red test = `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-7/test/e2e/reviewer7_shellopts_test.go`; baseline touch creates its marker, inherited onecmd case also passes. New noexec case fails. Existing `TestReviewR1X2ExportedErrexitPreservesStatus`, `TestX2ExportedShelloptsReachTheCommandAlone`, `TestExportedShelloptsReachTheCommandAlone` all pass; last command used `-v` and executed both scope + refused-scope (no skip).
- Command = `go test -count=1 -run '^TestR7X2ExportedShelloptsCannotSkipCommand$' ./test/e2e/`; rc 1:

```text
--- FAIL: TestR7X2ExportedShelloptsCannotSkipCommand (0.61s)
    --- FAIL: TestR7X2ExportedShelloptsCannotSkipCommand/noexec (0.20s)
        reviewer7_shellopts_test.go:25: SHELLOPTS=noexec: offload returned success without running touch: stat /tmp/TestR7X2ExportedShelloptsCannotSkipCommandnoexec815809098/001/project/ran: no such file or directory; stderr="offload: local slot 1/2\n"
FAIL
FAIL github.com/eturkes/offload/test/e2e 1.006s
```

Red-test commit for R7-F2 = `bc350b3`.

### R7-F3 — R9 reservation guard misses ordinary assignment forms

- Severity = minor. Location = `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-7/internal/run/run_test.go:679–713` @ `269d313`.
- Contract = X2/X8 + R9/M26: every script-owned variable uses `__…`; a new unreserved name must be detected, not just a fixed set of names/forms. The lexical scanner recognizes leading `NAME=` but not `NAME+=`, and its command switch omits `declare`. Both Bash forms assign variables; a pre-exported name becomes CMD's changed environment, exactly the reason for the reservation rule.
- Red test = `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-7/internal/run/reviewer7_guard_test.go`. A self-exec child appends one statement to `watchScript` **in memory** and invokes the actual existing guard. Positive control `plain=1` fails with `watch.bash sets "plain"`; `declare plain=1` + `plain+=suffix` both produce `PASS`, rc 0.
- Command = `go test -count=1 -run '^TestR7ReservationGuardFindsDeclarationsAndAppends$' ./internal/run/`; rc 1:

```text
--- FAIL: TestR7ReservationGuardFindsDeclarationsAndAppends (0.01s)
    --- FAIL: TestR7ReservationGuardFindsDeclarationsAndAppends/declare_plain=1 (0.00s)
        reviewer7_guard_test.go:29: reservation guard accepted "declare plain=1" or failed elsewhere: err=<nil> output="PASS\n"; want rejection naming plain
    --- FAIL: TestR7ReservationGuardFindsDeclarationsAndAppends/plain+=suffix (0.00s)
        reviewer7_guard_test.go:29: reservation guard accepted "plain+=suffix" or failed elsewhere: err=<nil> output="PASS\n"; want rejection naming plain
FAIL
FAIL github.com/eturkes/offload/internal/run 0.013s
```

Red-test commit for R7-F3 = `02cb1a7`.

### R7-F4 — R28 failed checksums compare equal

- Severity = minor. Location = `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-7/tools/install:13,16` @ `269d313`.
- Contract = R28: automatically compare each installed file's checksum with the built file's, refuse a mismatch. `sum` discards sha256sum's exit status behind `cut`; `report` accepts any equal strings, empty included. When checksum collection fails on both source + destination, installation reports empty digests as a successful match and exits 0. A failed check cannot establish byte equality.
- Red test = `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-7/tools/reviewer7_install_test.go`; temporary HOME + stub ssh, no host contact. Positive control runs real sha256sum through the same fixture and succeeds. Negative case makes sha256sum exit 23 with `injected checksum failure`, mirroring the script's local + remote pipelines.
- Command = `go test -count=1 -run '^TestR7InstallRejectsFailedChecksum$' ./tools/`; rc 1:

```text
--- FAIL: TestR7InstallRejectsFailedChecksum (7.89s)
    --- FAIL: TestR7InstallRejectsFailedChecksum/checksum-fails-1 (3.36s)
        reviewer7_install_test.go:47: install accepted missing checksums (rc 0): injected checksum failure
            injected checksum failure
            here offload:
            injected checksum failure
            injected checksum failure
            here heavy:
            injected checksum failure
            injected checksum failure
            fake-host offload:
            injected checksum failure
            injected checksum failure
            fake-host heavy:
FAIL
FAIL github.com/eturkes/offload/tools 7.925s
```

Red-test commit for R7-F4 = `c60cad9`.

## Register

- No off-contract defect asserted. Security adjudication remains MAIN's R17/R27/R38 work (table).
- Coverage limits: no full gate, full-suite mutation, live nixbox, live install/deploy, or CI rerun (brief restrictions). Acceptance CI results = ledger claims. LSP unavailable for `.go` (`documentSymbol` on `internal/shell/look.go` → no server); source/caller inspection + targeted tests used. Directive file absent at every between-row check.
- Runtime fixtures isolated: e2e fake hosts, stub ssh for deploy/install, temporary HOME/state; no live credentials read. No production source edits. Four red-test commits only on `wt/reviewer-7`; no pushes.
