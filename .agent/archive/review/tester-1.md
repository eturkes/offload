# tester-1 coverage — diff-blind offload e2e

26 rows = red-ok; 1 row = partially not-testable (T-Y5 framing). All 36 contract test functions fail the committed CLI stub; the fake-SSH harness function passes. No implementation source read. No real SSH/network; every CLI process uses private HOME/config/runtime + fake hosts.

Status: `red-ok` = failing for expected stub behavior; `red-other` = another failure; `not-testable` = explicitly bounded missing public seam. `red-ok` records red evidence, not a green implementation verdict.

| id | predicate | test(s) | status |
|----|-----------|---------|--------|
| T-C1 | C1 flag parsing stops at `--`/first word | `TestC1FlagBoundary` | red-ok |
| T-C2 | C2 flag conflicts + implications | `TestC2ConflictsAndPoolImplications` | red-ok |
| T-C3 | C3 exit precedence (command 0/2/125/255, 127, signals 130/143) | `TestC3ExitPrecedence` | red-ok |
| T-C4 | C4 stdout = command output only; offload messages on stderr | `TestC4OutputChannels` | red-ok |
| T-C5 | C5 placement line (local + remote forms) | `TestC5PlacementLine` | red-ok |
| T-P3 | P3 missing pool file: run local-only note; --pool/status rc 2 | `TestP3MissingPool` | red-ok |
| T-R2 | R2 status output incl. a down host | `TestR2Status` | red-ok |
| T-L1 | L1 eligibility (local first, --pool, --local, worktree) | `TestL1Eligibility`, `TestY3UnsupportedGitLayouts` | red-ok |
| T-L2 | L2 placement picks a free slot / lowest busy, --host | `TestL2Placement` | red-ok |
| T-L3 | L3 waiting line + INT while waiting → 130 | `TestL3WaitingSignals` | red-ok |
| T-L4 | L4 --gpu with all hosts GPU-less → 125 | `TestL4NoGPU`, `TestL4DownOrBusyGPUHostsWait` | red-ok |
| T-Y1 | Y1 root = toplevel; subdir cwd runs in matching remote subdir | `TestY1RootAndSubdirectory`, `TestY1GitErrorIsNotPlainMode` | red-ok |
| T-Y2 | Y2 workdir naming | `TestY2WorkdirIdentity` | red-ok |
| T-Y3 | Y3 worktree / submodule / nested repo → local-only, --pool → 125 | `TestY3UnsupportedGitLayouts` | red-ok |
| T-Y4 | Y4 upload set (tracked-but-ignored crosses, `!negation` re-include, untracked ignored stays, self-ignored .gitignore crosses, .git mirrored, .git/*.lock not sent) — differential vs `git ls-files` | `TestY4GitUploadDifferential`, `TestY4GitMetadataMirrorDeletes` | red-ok |
| T-Y5 | Y5 prune (stale + locally-deleted tracked removed; remote ignored kept; symlink-ancestor never followed; dirs never removed; framing mismatch → 125 with zero prune deletions) | `TestY5Prune`, `TestY5PruneSymlinkAncestor` | not-testable |
| T-Y6 | Y6 download (new + modified outputs return, nonzero cmd still downloads, newer local file kept, deleted tracked path no error, .git + ignored untracked stay remote) | `TestY6Download` | red-ok |
| T-Y7 | Y7 plain (non-git) mode mirror + filter | `TestY7PlainMirror` | red-ok |
| T-Y8 | Y8 --get (ignored output fetched; dir/ contents; escape/absolute/.git refused rc 2; missing → 125) | `TestY8Get`, `TestY8GetKeepsNewerLocal` | red-ok |
| T-Y9 | Y9 --no-sync (no upload, --get still runs) | `TestY9NoSync`, `TestY9NoSyncCreatesWorkdir` | red-ok |
| T-Y10 | Y10 upload failure → 125, command never runs | `TestY10UploadFailure`, `TestY10DownloadFailureOverridesCommand` | red-ok |
| T-Y11 | Y11 successor waits for a live predecessor's remote lock | `TestY11RemoteLock`, `TestY11LivePredecessorAfterControllerDeath` | red-ok |
| T-X1 | X1 SIGKILL of offload ends the command's whole session (incl. a setpgid child), local + remote; TERM-ignoring child gets KILL | `TestX1ControllerDeathKillsSession`, `TestX1LeaderExitReapsBackground` | red-ok |
| T-X2 | X2 local argv no shell (`$HOME` literal stays literal); remote via bash -lc; missing command 127 | `TestX2ArgvAndEnvironment` | red-ok |
| T-X3 | X3 TERM to offload → 143, no download | `TestX3SignalsSkipDownload` | red-ok |
| T-X4 | X4 byte-exact streaming (binary stdout, stderr separate) | `TestX4ByteExactStreaming` | red-ok |
| T-X6 | X6 command exiting 255 remotely → 255, not 125 | `TestX6Remote255AndTransport` | red-ok |

## Checks

- Base = `d4d48284ca5d946c57fc1517e2379e7491079757`; branch `wt/tester-1`; test commit = `0b8eb3baeca8df043ae0e7bc9c871e6df3679ad8` (clean worktree).
- Worktree = `/home/eturkes/Projects/offload/.scratch/worktrees/tester-1`.
- `go vet ./test/e2e/` → rc 0.
- `/home/eturkes/Projects/offload/.agent/tool golangci-lint fmt ./test/e2e/...` → rc 0.
- `/home/eturkes/Projects/offload/.agent/tool golangci-lint run ./test/e2e/...` → rc 0, `0 issues.`
- `/home/eturkes/Projects/offload/.agent/tool shellcheck -S warning test/e2e/testdata/ssh` → rc 0.
- `go test ./test/e2e/ -count=1` → rc 1, expected red.
- `go test ./test/e2e/ -count=1 -json` → rc 1; 0.924s package time; 100 failing leaf cases, 16 passing leaf cases.
- `go test ./test/e2e/ -count=1 -race -run '^TestHarnessFakeSSH$'` → rc 0 (1.239s): argv/options + stdin through EOF; refused host rc255; real rsync copies binary bytes.
- LSP `.go` probe → no server available; vet/lint used. Full repository gate + mutation testing not run (red-suite task).
- Positive seed: 27 coverage rows, 27 unknown statuses. Final validator → count 0 (grep rc 1 = no matches); independent row-count control → 27 (rc 0).
- Passing baseline rejection cases are not credited as red-tested assurance; mutation evidence required when claiming them verified. No tests skipped, deleted, or xfailed.

## Shared failing stderr

`U` below abbreviates these exact bytes (including final newline):
```text
Usage:
  offload run [--local | --pool] [--host NAME] [--gpu] [--no-sync] [--get PATH]... -- COMMAND [ARG...]
  offload status
```

Per-row commands run from the worktree; evidence below is from the full red run with the same test functions. Every failing leaf diagnostic is retained; repeated `stderr="...Usage..."` is losslessly abbreviated as `stderr=U`.

## T-C1

Command: `go test ./test/e2e/ -count=1 -run '^(TestC1FlagBoundary)$'`.

```text
TestC1FlagBoundary/separator → FAIL
  cli_test.go:24: exit = 2, want 0; stdout="" stderr=U
TestC1FlagBoundary/first-word → FAIL
  cli_test.go:24: exit = 2, want 0; stdout="" stderr=U
TestC1FlagBoundary/run-h → FAIL
  cli_test.go:35: exit = 2, want 0; stdout="" stderr=U
TestC1FlagBoundary/run--help → FAIL
  cli_test.go:35: exit = 2, want 0; stdout="" stderr=U
TestC1FlagBoundary/status-h → FAIL
  cli_test.go:35: exit = 2, want 0; stdout="" stderr=U
TestC1FlagBoundary/status--help → FAIL
  cli_test.go:35: exit = 2, want 0; stdout="" stderr=U
```

## T-C2

Command: `go test ./test/e2e/ -count=1 -run '^(TestC2ConflictsAndPoolImplications)$'`.

```text
TestC2ConflictsAndPoolImplications/conflict---pool → PASS on stub; not red-credited
TestC2ConflictsAndPoolImplications/implies-pool---pool → FAIL
  cli_test.go:58: exit = 2, want 0; stdout="" stderr=U
TestC2ConflictsAndPoolImplications/conflict---host → PASS on stub; not red-credited
TestC2ConflictsAndPoolImplications/implies-pool---host → FAIL
  cli_test.go:58: exit = 2, want 0; stdout="" stderr=U
TestC2ConflictsAndPoolImplications/conflict---gpu → PASS on stub; not red-credited
TestC2ConflictsAndPoolImplications/implies-pool---gpu → FAIL
  cli_test.go:58: exit = 2, want 0; stdout="" stderr=U
TestC2ConflictsAndPoolImplications/conflict---no-sync → PASS on stub; not red-credited
TestC2ConflictsAndPoolImplications/implies-pool---no-sync → FAIL
  cli_test.go:58: exit = 2, want 0; stdout="" stderr=U
TestC2ConflictsAndPoolImplications/conflict---get → PASS on stub; not red-credited
TestC2ConflictsAndPoolImplications/implies-pool---get → FAIL
  cli_test.go:58: exit = 2, want 0; stdout="" stderr=U
```

## T-C3

Command: `go test ./test/e2e/ -count=1 -run '^(TestC3ExitPrecedence)$'`.

```text
TestC3ExitPrecedence/--local/0 → FAIL
  cli_test.go:72: exit = 2, want 0; stdout="" stderr=U
TestC3ExitPrecedence/--local/2 → FAIL
  cli_test.go:74: command did not execute: stdout="" stderr=U
TestC3ExitPrecedence/--local/42 → FAIL
  cli_test.go:72: exit = 2, want 42; stdout="" stderr=U
TestC3ExitPrecedence/--local/125 → FAIL
  cli_test.go:72: exit = 2, want 125; stdout="" stderr=U
TestC3ExitPrecedence/--local/255 → FAIL
  cli_test.go:72: exit = 2, want 255; stdout="" stderr=U
TestC3ExitPrecedence/--local/command-INT → FAIL
  cli_test.go:84: exit = 2, want 130; stdout="" stderr=U
TestC3ExitPrecedence/--local/command-TERM → FAIL
  cli_test.go:84: exit = 2, want 143; stdout="" stderr=U
TestC3ExitPrecedence/--local/command-HUP → FAIL
  cli_test.go:84: exit = 2, want 129; stdout="" stderr=U
TestC3ExitPrecedence/--local/not-found → FAIL
  cli_test.go:89: exit = 2, want 127; stdout="" stderr=U
TestC3ExitPrecedence/--local/not-executable → FAIL
  cli_test.go:97: exit = 2, want 126; stdout="" stderr=U
TestC3ExitPrecedence/--pool/0 → FAIL
  cli_test.go:72: exit = 2, want 0; stdout="" stderr=U
TestC3ExitPrecedence/--pool/2 → FAIL
  cli_test.go:74: command did not execute: stdout="" stderr=U
TestC3ExitPrecedence/--pool/42 → FAIL
  cli_test.go:72: exit = 2, want 42; stdout="" stderr=U
TestC3ExitPrecedence/--pool/125 → FAIL
  cli_test.go:72: exit = 2, want 125; stdout="" stderr=U
TestC3ExitPrecedence/--pool/255 → FAIL
  cli_test.go:72: exit = 2, want 255; stdout="" stderr=U
TestC3ExitPrecedence/--pool/command-INT → FAIL
  cli_test.go:84: exit = 2, want 130; stdout="" stderr=U
TestC3ExitPrecedence/--pool/command-TERM → FAIL
  cli_test.go:84: exit = 2, want 143; stdout="" stderr=U
TestC3ExitPrecedence/--pool/command-HUP → FAIL
  cli_test.go:84: exit = 2, want 129; stdout="" stderr=U
TestC3ExitPrecedence/--pool/not-found → FAIL
  cli_test.go:89: exit = 2, want 127; stdout="" stderr=U
TestC3ExitPrecedence/--pool/not-executable → FAIL
  cli_test.go:97: exit = 2, want 126; stdout="" stderr=U
TestC3ExitPrecedence/bad-pool-before-command → PASS on stub; not red-credited
TestC3ExitPrecedence/unknown-host-before-command → PASS on stub; not red-credited
```

## T-C4

Command: `go test ./test/e2e/ -count=1 -run '^(TestC4OutputChannels)$'`.

```text
TestC4OutputChannels/--local → FAIL
  cli_test.go:118: exit = 2, want 0; stdout="" stderr=U
TestC4OutputChannels/--pool → FAIL
  cli_test.go:118: exit = 2, want 0; stdout="" stderr=U
```

## T-C5

Command: `go test ./test/e2e/ -count=1 -run '^(TestC5PlacementLine)$'`.

```text
TestC5PlacementLine/--local → FAIL
  cli_test.go:141: exit = 2, want 0; stdout="" stderr=U
TestC5PlacementLine/--pool → FAIL
  cli_test.go:141: exit = 2, want 0; stdout="" stderr=U
```

## T-P3

Command: `go test ./test/e2e/ -count=1 -run '^(TestP3MissingPool)$'`.

```text
TestP3MissingPool/run → FAIL
  cli_test.go:168: exit = 2, want 0; stdout="" stderr=U
TestP3MissingPool/pool → FAIL
  cli_test.go:169: U does not contain "/tmp/TestP3MissingPoolpool2766504915/001/pool.toml"
TestP3MissingPool/status → FAIL
  cli_test.go:169: U does not contain "/tmp/TestP3MissingPoolstatus4002093960/001/pool.toml"
```

## T-R2

Literal contract reading: require pool order, down reason, held/total, CPUs, busy percentage, GPU presence, and held local owner; reject stale unlocked owner text. Status column formatting is unspecified, so the test does not impose a header/table template. Free-memory formatting and exact load-field rendering are not independently pinned.

Command: `go test ./test/e2e/ -count=1 -run '^(TestR2Status)$'`.

```text
TestR2Status → FAIL
  cli_test.go:191: exit = 2, want 0; stdout="" stderr=U
```

## T-L1

Command: `go test ./test/e2e/ -count=1 -run '^(TestL1Eligibility|TestY3UnsupportedGitLayouts)$'`.

```text
TestL1Eligibility/default → FAIL
  cli_test.go:225: exit = 2, want 0; stdout="" stderr=U
TestL1Eligibility/local → FAIL
  cli_test.go:225: exit = 2, want 0; stdout="" stderr=U
TestL1Eligibility/pool → FAIL
  cli_test.go:225: exit = 2, want 0; stdout="" stderr=U
TestL1Eligibility/local-full-falls-back → FAIL
  cli_test.go:240: exit = 2, want 0; stdout="" stderr=U
TestY3UnsupportedGitLayouts/worktree → FAIL
  sync_test.go:72: exit = 2, want 0; stdout="" stderr=U
TestY3UnsupportedGitLayouts/submodule → FAIL
  sync_test.go:72: exit = 2, want 0; stdout="" stderr=U
TestY3UnsupportedGitLayouts/nested → FAIL
  sync_test.go:72: exit = 2, want 0; stdout="" stderr=U
TestY3UnsupportedGitLayouts/external-git-dir → FAIL
  sync_test.go:72: exit = 2, want 0; stdout="" stderr=U
TestY3UnsupportedGitLayouts/alternates → FAIL
  sync_test.go:72: exit = 2, want 0; stdout="" stderr=U
```

## T-L2

Command: `go test ./test/e2e/ -count=1 -run '^(TestL2Placement)$'`.

```text
TestL2Placement/lowest-ratio → FAIL
  cli_test.go:272: exit = 2, want 0; stdout="" stderr=U
TestL2Placement/tie-pool-order → FAIL
  cli_test.go:272: exit = 2, want 0; stdout="" stderr=U
TestL2Placement/busy-best-host → FAIL
  cli_test.go:272: exit = 2, want 0; stdout="" stderr=U
TestL2Placement/explicit-host → FAIL
  cli_test.go:272: exit = 2, want 0; stdout="" stderr=U
```

## T-L3

Waiting signals + released-slot retry are covered. Exact 2s retry cadence and 30s down-host re-probe ceiling are not timed in this suite.

Command: `go test ./test/e2e/ -count=1 -run '^(TestL3WaitingSignals)$'`.

```text
TestL3WaitingSignals/INT → FAIL
  cli_test.go:293: offload exited before waiting message: exit=2 stdout="" stderr=U
TestL3WaitingSignals/TERM → FAIL
  cli_test.go:293: offload exited before waiting message: exit=2 stdout="" stderr=U
TestL3WaitingSignals/HUP → FAIL
  cli_test.go:293: offload exited before waiting message: exit=2 stdout="" stderr=U
TestL3WaitingSignals/release-retries → FAIL
  cli_test.go:309: offload exited before waiting message: exit=2 stdout="" stderr=U
```

## T-L4

Command: `go test ./test/e2e/ -count=1 -run '^(TestL4NoGPU|TestL4DownOrBusyGPUHostsWait)$'`.

```text
TestL4NoGPU → FAIL
  cli_test.go:322: exit = 2, want 125; stdout="" stderr=U
TestL4DownOrBusyGPUHostsWait/down → FAIL
  failures_test.go:117: offload exited before waiting with down/busy GPU candidate: exit=2 stdout="" stderr=U
TestL4DownOrBusyGPUHostsWait/busy → FAIL
  failures_test.go:117: offload exited before waiting with down/busy GPU candidate: exit=2 stdout="" stderr=U
```

## T-Y1

Command: `go test ./test/e2e/ -count=1 -run '^(TestY1RootAndSubdirectory|TestY1GitErrorIsNotPlainMode)$'`.

```text
TestY1RootAndSubdirectory → FAIL
  sync_test.go:23: exit = 2, want 0; stdout="" stderr=U
TestY1GitErrorIsNotPlainMode → FAIL
  failures_test.go:22: exit = 2, want 125; stdout="" stderr=U
```

## T-Y2

Command: `go test ./test/e2e/ -count=1 -run '^(TestY2WorkdirIdentity)$'`.

```text
TestY2WorkdirIdentity → FAIL
  sync_test.go:36: exit = 2, want 0; stdout="" stderr=U
```

## T-Y3

Command: `go test ./test/e2e/ -count=1 -run '^(TestY3UnsupportedGitLayouts)$'`.

```text
TestY3UnsupportedGitLayouts/worktree → FAIL
  sync_test.go:72: exit = 2, want 0; stdout="" stderr=U
TestY3UnsupportedGitLayouts/submodule → FAIL
  sync_test.go:72: exit = 2, want 0; stdout="" stderr=U
TestY3UnsupportedGitLayouts/nested → FAIL
  sync_test.go:72: exit = 2, want 0; stdout="" stderr=U
TestY3UnsupportedGitLayouts/external-git-dir → FAIL
  sync_test.go:72: exit = 2, want 0; stdout="" stderr=U
TestY3UnsupportedGitLayouts/alternates → FAIL
  sync_test.go:72: exit = 2, want 0; stdout="" stderr=U
```

## T-Y4

Differential oracle = real git ls-files -z -co --exclude-standard filtered by lstat + a separate walk for .gitignore files in nonignored directories. Three fixed generated seeds (1, 7, 91), 16 generated paths each; fixture also includes force-added ignored tracked paths, negations, nested/self-ignored .gitignore, info/global excludes, space/newline filenames, normal/dangling symlinks, missing tracked entries, and .git lock exclusions. Finite generated coverage, not exhaustive enumeration.

Command: `go test ./test/e2e/ -count=1 -run '^(TestY4GitUploadDifferential|TestY4GitMetadataMirrorDeletes)$'`.

```text
TestY4GitUploadDifferential/seed-1 → FAIL
  sync_test.go:181: exit = 2, want 0; stdout="" stderr=U
TestY4GitUploadDifferential/seed-7 → FAIL
  sync_test.go:181: exit = 2, want 0; stdout="" stderr=U
TestY4GitUploadDifferential/seed-91 → FAIL
  sync_test.go:181: exit = 2, want 0; stdout="" stderr=U
TestY4GitMetadataMirrorDeletes → FAIL
  sync_test.go:201: exit = 2, want 0; stdout="" stderr=U
```

## T-Y5

Partial row: prune cases are red-ok. Exact count/nonce mismatch injection = not-testable from the public contract alone: Y5 specifies semantic framing but no encoding, transport channel, or externally addressable frame path. No protocol guessed; MAIN must supply independent internal corruption tests or a specified fault seam. Zero-deletion-on-framing-error is unconfirmed. Existing tests cover stale untracked + locally deleted tracked removals, ignored cache retention, directory retention, and symlink-ancestor safety.

Command: `go test ./test/e2e/ -count=1 -run '^(TestY5Prune|TestY5PruneSymlinkAncestor)$'`.

```text
TestY5Prune → FAIL
  sync_test.go:213: exit = 2, want 0; stdout="" stderr=U
TestY5PruneSymlinkAncestor → FAIL
  sync_test.go:232: exit = 2, want 0; stdout="" stderr=U
```

## T-Y6

Command: `go test ./test/e2e/ -count=1 -run '^(TestY6Download)$'`.

```text
TestY6Download/command-0 → FAIL
  sync_test.go:258: exit = 2, want 0; stdout="" stderr=U
TestY6Download/command-2 → FAIL
  sync_test.go:260: modified/new outputs did not return
TestY6Download/command-125 → FAIL
  sync_test.go:258: exit = 2, want 125; stdout="" stderr=U
TestY6Download/command-255 → FAIL
  sync_test.go:258: exit = 2, want 255; stdout="" stderr=U
TestY6Download/symlink-and-type-change → FAIL
  sync_test.go:282: exit = 2, want 0; stdout="" stderr=U
```

## T-Y7

Command: `go test ./test/e2e/ -count=1 -run '^(TestY7PlainMirror)$'`.

```text
TestY7PlainMirror → FAIL
  sync_test.go:302: exit = 2, want 0; stdout="" stderr=U
```

## T-Y8

Literal trailing-slash reading: --get ignored/tree/ copies its contents into local ignored/tree/, not a second nested tree directory. Paths resolve from invocation cwd; normalization may reach another path inside root but must reject root escape and normalized .git paths.

Command: `go test ./test/e2e/ -count=1 -run '^(TestY8Get|TestY8GetKeepsNewerLocal)$'`.

```text
TestY8Get/ignored-and-directory → FAIL
  sync_test.go:325: exit = 2, want 0; stdout="" stderr=U
TestY8Get/reject-.._escape → PASS on stub; not red-credited
TestY8Get/reject-_absolute → PASS on stub; not red-credited
TestY8Get/reject-.git → PASS on stub; not red-credited
TestY8Get/reject-.git_config → PASS on stub; not red-credited
TestY8Get/reject-sub_.._.._escape → PASS on stub; not red-credited
TestY8Get/reject-sub_.._.git_HEAD → PASS on stub; not red-credited
TestY8Get/missing → FAIL
  sync_test.go:342: exit = 2, want 125; stdout="" stderr=U
TestY8GetKeepsNewerLocal → FAIL
  failures_test.go:34: exit = 2, want 0; stdout="" stderr=U
```

## T-Y9

Command: `go test ./test/e2e/ -count=1 -run '^(TestY9NoSync|TestY9NoSyncCreatesWorkdir)$'`.

```text
TestY9NoSync → FAIL
  sync_test.go:352: exit = 2, want 0; stdout="" stderr=U
TestY9NoSyncCreatesWorkdir → FAIL
  failures_test.go:44: exit = 2, want 0; stdout="" stderr=U
```

## T-Y10

Command: `go test ./test/e2e/ -count=1 -run '^(TestY10UploadFailure|TestY10DownloadFailureOverridesCommand)$'`.

```text
TestY10UploadFailure → FAIL
  sync_test.go:363: exit = 2, want 125; stdout="" stderr=U
TestY10DownloadFailureOverridesCommand/0 → FAIL
  failures_test.go:56: exit = 2, want 125; stdout="" stderr=U
TestY10DownloadFailureOverridesCommand/2 → FAIL
  failures_test.go:56: exit = 2, want 125; stdout="" stderr=U
TestY10DownloadFailureOverridesCommand/255 → FAIL
  failures_test.go:56: exit = 2, want 125; stdout="" stderr=U
```

## T-Y11

Both externally held specified meta/lock and real live predecessor are covered. Live case kills first controller, holds predecessor in TERM cleanup behind a barrier, starts successor, checks no upload or CMD while predecessor lives, then releases barrier and requires successor completion. The 300s timeout boundary is not exercised to keep this e2e suite bounded.

Command: `go test ./test/e2e/ -count=1 -run '^(TestY11RemoteLock|TestY11LivePredecessorAfterControllerDeath)$'`.

```text
TestY11RemoteLock → FAIL
  sync_test.go:378: offload exited before remote connection: exit=2 stdout="" stderr=U
TestY11LivePredecessorAfterControllerDeath → FAIL
  failures_test.go:75: offload exited before predecessor pid: exit=2 stdout="" stderr=U
```

## T-X1

Strict gone = kill(pid, 0) returns ESRCH, including reaping; zombies count as remaining. New-session assertion precedes session-wide inspection/cleanup, so a broken setsid implementation cannot target the test runner session. One local TERM-ignoring child permits 13s observation for the specified 10s escalation plus scheduling/reaping.

Command: `go test ./test/e2e/ -count=1 -run '^(TestX1ControllerDeathKillsSession|TestX1LeaderExitReapsBackground)$'`.

```text
TestX1ControllerDeathKillsSession/local → FAIL
  execution_test.go:74: offload exited before session child: exit=2 stdout="" stderr=U
TestX1ControllerDeathKillsSession/remote → FAIL
  execution_test.go:74: offload exited before session child: exit=2 stdout="" stderr=U
TestX1ControllerDeathKillsSession/local-kill-escalation → FAIL
  execution_test.go:74: offload exited before session child: exit=2 stdout="" stderr=U
TestX1LeaderExitReapsBackground/--local → FAIL
  execution_test.go:124: exit = 2, want 7; stdout="" stderr=U
TestX1LeaderExitReapsBackground/--pool → FAIL
  execution_test.go:124: exit = 2, want 7; stdout="" stderr=U
```

## T-X2

Command: `go test ./test/e2e/ -count=1 -run '^(TestX2ArgvAndEnvironment)$'`.

```text
TestX2ArgvAndEnvironment/--local → FAIL
  execution_test.go:137: fork/exec /tmp/offload-e2e-build-1199544396/offload: invalid argument
TestX2ArgvAndEnvironment/--pool → FAIL
  execution_test.go:137: fork/exec /tmp/offload-e2e-build-1199544396/offload: invalid argument
TestX2ArgvAndEnvironment/remote-login → FAIL
  execution_test.go:154: exit = 2, want 0; stdout="" stderr=U
```

## T-X3

Command: `go test ./test/e2e/ -count=1 -run '^(TestX3SignalsSkipDownload)$'`.

```text
TestX3SignalsSkipDownload/TERM → FAIL
  execution_test.go:175: offload exited before remote output and command pid: exit=2 stdout="" stderr=U
TestX3SignalsSkipDownload/INT → FAIL
  execution_test.go:175: offload exited before remote output and command pid: exit=2 stdout="" stderr=U
TestX3SignalsSkipDownload/HUP → FAIL
  execution_test.go:175: offload exited before remote output and command pid: exit=2 stdout="" stderr=U
```

## T-X4

Command: `go test ./test/e2e/ -count=1 -run '^(TestX4ByteExactStreaming)$'`.

```text
TestX4ByteExactStreaming/--local → FAIL
  execution_test.go:197: offload exited before live stdout/stderr before command can exit: exit=2 stdout="" stderr=U
TestX4ByteExactStreaming/--pool → FAIL
  execution_test.go:197: offload exited before live stdout/stderr before command can exit: exit=2 stdout="" stderr=U
```

## T-X6

Command rc255 + download is covered. Transport rc255 injection currently occurs at the first post-probe SSH phase, so rc125 proves remote-setup/transport failure precedence, not specifically the run-status recovery branch. Matching/stale/missing nonce readback behavior remains unconfirmed beyond the successful command255 end-to-end case.

Command: `go test ./test/e2e/ -count=1 -run '^(TestX6Remote255AndTransport)$'`.

```text
TestX6Remote255AndTransport/command-255 → FAIL
  execution_test.go:216: exit = 2, want 255; stdout="" stderr=U
TestX6Remote255AndTransport/transport-255 → FAIL
  execution_test.go:225: exit = 2, want 125; stdout="" stderr=U
```
