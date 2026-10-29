# reviewer-1 — A: correctness / spec

Target = `5c11b58..57903b5` (`57903b5dafc3c4fb98b358bc2546af4285990c83`). Check set fixed before source reads. **12/12 adjudicated: 3 pass; 9 finding rows; 6 distinct findings (4 major, 2 minor).** Red tests committed on `wt/reviewer-1`: `20aee9512f7594a92c10cf452d4667daf7b9e48c` + `d291d01d30b224b8b49f6ec382d6eaf8a3afcfef` (six added files; original source/tests unchanged; clean worktree). A2's initial pass reversed on new PATH-classification evidence from reviewer-2, independently reproduced below.

| id | check | verdict | finding ids | evidence |
|----|-------|---------|-------------|----------|
| A1 | K1 parsing + top-level dispatch (internal/cli/cli.go, internal/cli/ctr.go) | finding | F-A4 | `internal/cli/ctr.go:39`: parse stops at earlier flag error; help loses. Dispatch, stop-at-CMD/`--`, option exclusion otherwise fit. |
| A2 | K2 switch values + native exec path (127/126, env/argv/cwd/pid kept, nothing else read) | finding | F-A6 | `internal/cli/cli.go:317` via `internal/cli/ctr.go:89`: nonexecutable PATH entry misclassified 127, not 126. Other native/switch controls pass; project/pool/podman not read. |
| A3 | K3 root via mirror.Locate, damaged repo, symlink resolution, definition regular-file rule, ':' refusal, binary resolution | pass | — | `internal/ctr/ctr.go:53–86,284–298`; root/cwd resolve before lookup; lstat refuses a definition symlink; mounts checked; inode rechecked before launch. K3 definition controls + mountable/inode unit tests pass. Context-hashing defect belongs to A4. |
| A4 | K4 Tag: entry set, order, framing, kinds, mode bits, symlinks, unsupported types (internal/ctr/tag.go) | finding | F-A1 | `internal/ctr/tag.go:25–30`: accepted symlink context hashes zero entries. Other entry kinds, byte ordering, 07777 mode bits + framing fit; existing K4 controls pass. |
| A5 | K5 Ensure: image-exists rc mapping, lock path + O_NOFOLLOW, re-check, build argv, lock inherited by build, signal → 128+n | finding | F-A3 | `internal/cli/ctr.go:100`: setup uses uncancelable git context. Ensure itself preserves exists 0/1/error, tag lock, second check, ExtraFiles lock + TERM cancellation. |
| A6 | K6 RunArgs + Run: mounts (root, common dir, HOME, binary ro), env, stdin, -w, --pid=private, inode check | pass | — | `internal/ctr/ctr.go:159–183,221–263`; RunArgs/Home/running tests pass under race detector; private-vfs real-podman repro starts helper + writes readiness through root mount. Remaining mount/env invariants = source/callee review. |
| A7 | K7 status record path: readRecord, podman failure, missing/partial record → 125, record wins over podman rc | finding | F-A2, F-A6 | `internal/ctr/ctr.go:235,304–308`: missing-record hang. `internal/ctr/init.go:91` misclassifies nonexecutable PATH command as 127. Existing complete/partial closed-peer tests pass but miss live peer. |
| A8 | K8 init.go: PID-1 guard, CLOEXEC fd 3, start (LookPath 127/126), forwarding to CMD's group, reaping, teardown TERM + 10 s, record-before-exit, liveness EOF; host forwarding in ctr.Run (pre-start cancel vs post-start socket) | finding | F-A2, F-A6 | `internal/ctr/init.go:91`: PATH permission classification wrong; guard/handlers/ForkExec/Wait4/teardown otherwise fit. Host missing-record hang prevents liveness closure when podman dies first; real-podman red reproduces. |
| A9 | K9 --layers + K10 --check: outcomes (2/125/1/0/128+n), probe-first down hosts, framed parse, full-vector compare, binary sha | finding | F-A3 | `internal/cli/ctr.go:100,113–232`: setup cancellation missing in both modes. Otherwise local-first work, parallel probe-before-run, ordered output, full `slices.Equal`, binary SHA + status decisions fit. Live remote behavior not run. |
| A10 | refactors keep behavior: mirror.Locate/Inspect (Y1, Y3), mirror.Name (Y2), slots.WaitLock vs TryLock semantics | pass | — | `internal/mirror/mirror.go:39–88,259–270`, `internal/slots/slots.go:165–195`; moved guards preserved; callers enumerated; TryLock unchanged. Selected mirror + slot regressions pass under race detector. |
| A11 | tools/install + machines/nixbox/configuration.nix changes do what their comments + README say | finding | F-A5 | `tools/install:17`: parser skips valid literal-string TOML addresses. Static Nix delta correctly enables local-bin PATH + machine-wide OFFLOAD_CTR; deployment not rerun. |
| A12 | Go pitfalls: data races (shared vars across goroutines), goroutine/fd leaks, signal.Notify semantics, exec.Cmd ExtraFiles/CLOEXEC, Wait4(-1) vs other waiters | finding | F-A2 | Control-fd deadline assumption is wrong. Other inspected invariants: Run shared ctl/got under mutex; Go ≥1.22 per-iteration captures; WaitGroup orders result reads; init is sole Wait4 waiter; native child only inherits 0/1/2; persistent signal registrations/goroutines end with one-shot CLI process. Selected race controls pass, not an exhaustive concurrency proof. |

## Findings

### F-A1 — major — accepted symlink context shares the empty-tree tag

- Target: `internal/ctr/tag.go:25–30`; caller `internal/ctr/ctr.go:81`.
- K3 accepts a regular `.devcontainer/Containerfile` below a symlinked `.devcontainer` directory. Find resolves root/cwd, not context. WalkDir never follows its root symlink; skipping `path == dir` hashes zero entries. All such definitions share `localhost/offload-ctr:e3b0c44298fc1c14`; a changed definition can reuse a stale/wrong image, violating K4.
- Red: `internal/ctr/reviewer1_test.go`, `TestReviewA4SymlinkContextChangesTag` (Find accepts fixture before definition changes).
- Command: red battery below → rc 1 on target revision.
- Failing output: `K4: changed Containerfile below accepted symlink context kept tag localhost/offload-ctr:e3b0c44298fc1c14`.
- Acceptance: accepted context directory hashes its actual entries; same regular tree through a symlink = same tag; changed file = changed tag.

### F-A2 — major — podman death leaves offload + container alive

- Target: `internal/ctr/ctr.go:235,304–308`.
- Blocking Socketpair + os.NewFile → fd not registered with Go's poller. readRecord ignores unsupported SetReadDeadline and blocks in Read. If podman dies first while init/conmon holds the peer, offload cannot return 125 or close liveness; container keeps running. K7 + K8 require the opposite.
- Red unit: `internal/ctr/reviewer1_test.go`, `TestReviewA7MissingRecordWithOpenPeer` → `K7/K8: missing status blocked past its 1 s deadline while peer remained open`.
- Red real-podman: `test/e2e/reviewer1_ctr_test.go`, `TestReviewA8PodmanDeathReturns125`; private vfs storage below worktree temp directory. Waits for CMD readiness; kills sole podman child; offload remains live at 3 s; forced offload SIGKILL then ends container. Storage cleaned.
- Command: red battery below → rc 1 on target revision.
- Failing output: `K8: podman died first, but offload stayed blocked and its container kept running`.
- Acceptance: podman dies without record → bounded read wait, offload 125 + liveness closes + container ends. Complete buffered record still wins over podman rc.

### F-A3 — major — signals cannot cancel project lookup

- Target: `internal/cli/ctr.go:77–82,96–100,114`.
- findProject passes context.Background to ctr.Find/mirror.Locate/git, discarding cancellation for --layers/--check; run mode has not started the signal-consuming Run goroutine yet. TERM during a stalled git lookup leaves git/offload running. Releasing that lookup by terminating fake git makes run mode return 125, not required 143. --layers/--check eventually return 143 but cannot stop setup.
- Red: `test/e2e/reviewer1_ctr_test.go`, `TestReviewA5SignalCancelsProjectLookup/{run,layers,check}`. Fake git records its pid before blocking; test signals offload, then terminates uncanceled git for cleanup/status proof.
- Command: red battery below → rc 1 on target revision.
- Failing output (all 3): `K5/K9/K10: SIGTERM during project lookup did not cancel git`; run additionally: `exit = 125, want 143; stdout="" stderr="offload: git rev-parse --path-format=absolute --show-toplevel --git-dir --git-common-dir --show-prefix: \n"`.
- Acceptance: each mode cancels in-flight lookup on each K8 signal, returns 128+n, starts no subsequent build/container; preserve post-start forwarding semantics.

### F-A4 — minor — earlier flag errors outrank help

- Target: `internal/cli/ctr.go:39–40,55–61`.
- K1 says help wins over any other problem. flag.Parse returns on the first malformed/unsupported flag, before a later help flag inside the same flag prefix.
- Red: `internal/cli/reviewer1_test.go`, `TestReviewA1HelpWinsOverBadFlag`.
- Command: red battery below → rc 1 on target revision.
- Failing outputs: `K1: ["ctr" "--bogus" "--help"] returned 2 instead of help's 0`; `K1: ["ctr" "--layers=invalid" "-h"] returned 2 instead of help's 0`.
- Acceptance: help in the parsed flag prefix wins; help after CMD or `--` remains CMD argv.

### F-A5 — major — installer silently omits valid pool hosts

- Target: `tools/install:17`; claim `README.md:39`, `tools/install:2–4`.
- Default install promises every pool-file address. Its sed expression accepts double-quoted values only. P2/Go pool.Load accepts literal TOML strings: `address = 'agent@alpha'`. Installer parser emits no address; the loop skips the host, leaving its offload missing/stale while the controller changes.
- Red: `internal/pool/reviewer1_test.go`, `TestReviewA11InstallerReadsValidPoolAddresses`. Executes only the extracted read-only sed expression from the target file; never runs tools/install, go build, install, SSH, or a remote action.
- Positive controls: pool.Load accepts exactly one `agent@alpha` host for both strings; double-quoted installer parser subtest passes (exactly one line). Literal-string parser exits 0 with zero bytes.
- Command: red battery below → rc 1 on target revision. Separate `go test -v -count=1 -run '^TestReviewA11InstallerReadsValidPoolAddresses$' ./internal/pool` confirms control pass + literal-string failure.
- Failing output: `installer parser emitted "", want exactly one agent@alpha address accepted by pool.Load`.
- Acceptance: default host list uses the same TOML semantics/validation as pool.Load; every accepted address enumerated exactly once. Actual install/deployment remains untested by this review.

### F-A6 — minor — nonexecutable PATH commands reported as missing

- Target: `internal/cli/cli.go:317–319` (new native caller `internal/cli/ctr.go:89`); `internal/ctr/init.go:91–93`.
- exec.LookPath returns ErrNotFound when PATH contains only an existing nonexecutable match. Both native and init paths map that result straight to 127. K2 + K7 require 126 for not-executable commands; the existing tests use a slash-containing command and miss PATH lookup.
- Reds: `internal/cli/reviewer1_path_test.go` (`TestReviewA2NonExecutableOnPath`), `internal/ctr/reviewer1_path_test.go` (`TestReviewA7NonExecutableOnPath`). Positive fixture = an existing regular file, 0600/0644, in PATH; neither test executes it.
- Command: `GOCACHE="$PWD/.scratch/reviewer1-go-cache" go test -count=1 -run '^TestReviewA[27]NonExecutableOnPath$' ./internal/cli ./internal/ctr` → rc 1 on target source.
- Failing outputs: `K2: existing nonexecutable PATH entry returned 127, want 126`; `K7: existing nonexecutable PATH entry returned pid=0 status=127, want 0/126`; both diagnostics say `command not found`.
- Acceptance: native + image lookup distinguish absent command (127) from a nonexecutable PATH candidate with no usable match (126); executable later PATH candidates still win. The run path's preexisting shared helper is also affected, but no separate existing-run finding is asserted.

## Commands + coverage

All commands run from `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-1`; target source unchanged. Test files are the only worktree source additions.

**Red witnesses (8 tests, 6 defects):** original six-test battery → rc 1; two PATH follow-up tests → rc 1 (F-A6 command above). Combined reproduction command:

```sh
mkdir -p .tmp .cache/go-build
TMPDIR="$PWD/.tmp" GOCACHE="$PWD/.cache/go-build" \
  go test -count=1 -run '^TestReviewA' \
  ./internal/cli ./internal/ctr ./internal/pool ./test/e2e
```

**Passing controls:**

- Race-enabled existing ctr tests: `TestK[3467]` → ctr package rc 0; entry hashing/modes/FIFO, RunArgs/Home, mountability, inode identity, complete/partial closed-peer records.
- Race-enabled existing slots tests: `TestSlotHeldBy*`, `TestLowestFreeSlotFirst`, `TestCommandNeverInheritsLock`, `TestSlotFilesNeverFollowSymlinks`, `TestS4WorkdirLockCreatesParents` → slots package rc 0.
- Race-enabled existing cli tests: `TestHelpPrintsUsageOnStdout`, `TestNoArgumentsIsAUsageError`, `TestParseRun` → cli package rc 0.
- `TMPDIR="$PWD/.tmp" GOCACHE="$PWD/.cache/go-build" GIT_CEILING_DIRECTORIES="$PWD/.tmp" go test -race -count=1 -run '^(TestInspectRefusesBrokenRepo|TestWorkdirName|TestMirrorY1.*|TestMirrorY3SplitGitDirs)$' ./internal/mirror` → rc 0.
- `TMPDIR="$PWD/.tmp" GOCACHE="$PWD/.cache/go-build" go test -count=1 -run '^(TestK[12][A-Z].*|TestK3MissingDefinitionIsUsageError|TestK3DefinitionMustBeRegularFile)$' ./test/e2e` → rc 0 (CLI/native exec + definition controls).

**Limits / failed attempts:**

- Full gate + full suite not run (brief: targeted tests only); no SSH, nixbox access, install, deployment, or shared image/container removal.
- First new e2e filename collided with baseline `test/e2e/reviewer1_test.go`; restored from target, additions moved to `reviewer1_ctr_test.go`, red battery rerun with original tests intact (`git diff --exit-code 57903b5 --` → rc 0 before staging additions).
- One private-storage cleanup raced conmon's asynchronous `--rm` and failed (`Directory not empty`). Fixture cleanup now synchronously removes containers in its private store before deleting storage; final targeted killed-client red still fails for F-A2 alone (9.845 s including builds), cleanup succeeds. Private `.tmp` + `.cache` removed at close; no matching review-temp process observed (`pgrep` rc 1).
- Initial mirror controls under worktree-local TMPDIR failed because damaged fixtures fell back to enclosing worktree; rerun with Git ceiling passed. No target implementation change.
- Existing real-podman shared prebuild fails before behavior when TMPDIR lies below this worktree: it locates the enclosing repository, not its plain fixture. One broad `TestK1` regex also selected K10 tests and hit that same fixture failure; corrected CLI-only selection passed. Custom real-podman F-A2 test uses an explicit git fixture + private storage and reproduces independently. Existing K5/K6/K8/K9/K10 e2e controls therefore not credited as passed.
- LSP call on actual `.go` file → no Go language server; call sites searched directly. E2e TestMain builds static non-race binaries; race credit above applies to unit controls, not all container subprocess code.
- Read-and-conform surfaces = current primary `.agent/spec.md`, `.agent/contract.md`, `.claude/rules/{code,gate-checks,gate,machines}.md`. Primary source, instructions + remote state remain unchanged.

## Register

None outside fixed check set.
