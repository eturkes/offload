# reviewer-3 verdicts — closing diff fed6498..cac258b

Check set frozen before target reads: G1–G7 below; acceptance = README.md + .agent/contract.md + .agent/spec.md Decisions on fed6498..cac258b. Verdict ∈ {pass, finding + ids}; unknown = in progress only. Findings = severity + file:line + red test committed on wt/reviewer-3; a nondeterministic gap names its interleaving + test limitation. Outside-set defects → register only. No production edits; no real hosts or shared slots.

| id | check | verdict | evidence |
|----|-------|---------|----------|
| G1 | 'killing offload (any signal) ends the command' — every path incl. kill during upload/probe/download, local + remote | finding F1 | watch.bash:24–34 can finish session checks before child calls setsid, then wait forever on an un-killed command. TestG1CancellationBeforeSetsidKillsSession red: alive after 12 s. |
| G2 | 'a successor never syncs into a live predecessor' (Y11) — all phases, controller death, partition | pass | Every rsync uses same meta flock; remote run retains fd9 through watcher teardown; controller workdir lock spans all phases. Existing TestY11RemoteLock + TestY11LivePredecessorAfterControllerDeath passed. Real network partition unexercised; same remote-lock ownership remains through it. |
| G3 | 'tracked files always cross; ignored stay on each side' — every list/rsync path, plain mode, --get | finding F3, F5 | remote.bash:63 hides git_list failure behind pipeline success; returns 0 with empty download list. TestG3DownloadListFailureIsNotSuccess red: 0 vs 125. Ignored preservation also fails F5. |
| G4 | 'nothing deleted locally; newer local file kept' (Y6) — rsync flags vs claim | finding F4 | README.md:93 overclaims; --update loses a concurrent local edit older than remote output. TestG4ConcurrentLocalEditSurvives red. |
| G5 | 'remote ignored files never deleted' (Y5) — prune + rsync --delete scopes | finding F5 | mirror.go:267 upload recursively removes an ignored remote directory to install a local file. TestG5UploadTypeChangeKeepsIgnoredRemoteFile red; offload rc 0. |
| G6 | slot admission: local heavy slot held through teardown; pool slot over-admission bounds | finding F6 | README.md:48 claims hard concurrent host limit; controller death frees slot while prior command tears down. Different workdirs overlap despite slots=1. TestG6PoolLimitIncludesLiveOrphan red. Contract S5 explicitly permits early release; docs-only mismatch. |
| G7 | exit-status precedence holds when two failures coincide (signal + rsync fail, 125 vs command 125) | finding F7, F7b | cli.go:162 canceled Inspect → 125 instead of 143 (F7); run.go:82 status lookup ignores cancellation after command 125/255 (F7b), hangs after TERM. Both red. |

## F1 — major — EOF before setsid misses the session forever

- Mechanism: `internal/run/watch.bash:16–34`. `$!` names the launcher before its `setsid` call. On EOF, TERM + both pgrep checks may run while that PID still belongs to the watcher's old session; no matching session exists yet. The watcher then waits for the launcher. When it eventually calls setsid, CMD starts in a new session that is never checked again. Local watcher retains its heavy slot while hanging; remote uses the identical watcher.
- Red: `test/e2e/reviewer_guarantees_test.go` `TestG1CancellationBeforeSetsidKillsSession`. Scheduling shim delays only setsid execution, waits for both no-session checks after SIGTERM, then releases real setsid. `X1 startup race: command alive=true after 12s; watcher checked before setsid and never retried`. Test kills its own child afterward, allowing offload to return 143 and leaving no deliberate survivor.
- Acceptance: command/session startup and watcher cancellation handshake must exclude this pre-setsid window; killing or observing the launcher alone without tracking its future session is insufficient.

## F3 — major — output-list failure reported as successful run

- Mechanism: `internal/run/remote.bash:63–65`, exact bytes `git_list | while IFS= read -r -d '' f; do`. Pipeline has no pipefail/error guard; a failing git list leaves an empty `down` file and successful final loop. Script writes the command's success status; DownloadGit succeeds with no files.
- Red: `test/e2e/reviewer_guarantees_test.go` `TestG3DownloadListFailureIsNotSuccess`. Command modifies `tracked` and damages its remote `.git/config`; git-list error prints, but offload returns 0 and suppresses the changed output instead of reporting 125.
- Failure: `exit = 0, want 125` + `fatal: bad config line 2 in file .git/config`.
- Acceptance: failure to enumerate outputs is an offload failure, not an empty successful output set. Preserve command-status diagnostics while returning 125.

## F4 — minor — concurrent-edit claim exceeds --update

- Anchor: `README.md:93` exact bytes `A file that you changed here during the run is not replaced.`
- Mechanism: `internal/mirror/mirror.go:282` compares current mtimes only; no upload snapshot or concurrent-edit detection. Contract Y6 correctly limits protection to a newer local regular file and permits symlink/type replacement; README promises more.
- Red: `test/e2e/reviewer_guarantees_test.go` `TestG4ConcurrentLocalEditSurvives`. After upload, edit local tracked file; remote finishes with a later-mtime output. Result: `README concurrent-edit guarantee: local edit replaced with "remote-result"`.
- Acceptance: align README with Y6, or detect/preserve every concurrent local edit. Current test deliberately checks README's stronger promise; do not silently promote that promise into an implementation requirement.

## F5 — major — upload deletes ignored remote cache on type replacement

- Claim anchor: `README.md:90` exact bytes `Ignored files on the host are never deleted.`
- Mechanism: `internal/mirror/mirror.go:267` exact bytes `args := append(base(t.Meta()), "-r", "--mkpath", "--from0", "--files-from=-", "--delete",`. `--delete` + `-r` applies beyond `.git`: replacing an existing remote directory with a listed local file recursively deletes the directory, including ignored leaves. Leaf-only Y5 prune guards never see that deletion.
- Red: `test/e2e/reviewer_guarantees_test.go` `TestG5UploadTypeChangeKeepsIgnoredRemoteFile`. `.gitignore=cache/`; first command creates ignored remote `cache/precious`; controller adds tracked regular file `cache`; second upload removes `cache/precious`, command succeeds, offload rc 0.
- Failure: `Y5/README ignored remote file deleted by upload type change; exit=0`.
- Acceptance: incompatible upload type change must retain ignored remote descendants (refusing before destructive replacement is consistent with preserving them); test proves the upload path, not the guarded prune path.

## F6 — minor — host concurrency limit overclaims S5

- Claim: `README.md:48`, exact bytes `` `slots` is the number of runs the host takes at the same time. ``
- Mechanism: `internal/slots/slots.go:128–132` controller-only flock dies with offload; `internal/cli/place.go:111` admits another workdir. Remote watcher may still spend 10 s tearing down the prior command (or ~1 min waiting on sshd peer death). `.agent/contract.md:32` S5 expressly allows that early controller-lock release, protecting only same-workdir safety via Y11.
- Red: `test/e2e/reviewer_guarantees_test.go` `TestG6PoolLimitIncludesLiveOrphan`. Host slots=1; first command traps TERM during teardown; SIGKILL its controller; second root starts another command on the host while the first remains live.
- Failure: `README pool slots=1 admitted two live remote commands after controller death`.
- Acceptance: document the controller-lifetime admission bound + temporary overlap after controller death, or intentionally strengthen admission with remote host-slot ownership. Contract already allows overlap; do not silently turn this documentation defect into a design change.

## F7 — major — signal loses C3 precedence during inspection

- Mechanism: `internal/cli/cli.go:161–162`: `mirror.Inspect(ctx, cwd)` failure returns `fail(stderr, err)` before the later ctx/signal checks. Same pattern follows the pre-placement UploadList call at :168.
- Red: `test/e2e/reviewer_guarantees_test.go` `TestG7SignalDuringInspectionWins`. PATH wrapper pauses git rev-parse; deliver SIGTERM; context kills git; offload exits 125, not required 143.
- Failure: `exit = 125, want 143; stdout="" stderr="offload: git rev-parse --path-format=absolute --show-toplevel --git-dir --git-common-dir --show-prefix: \n"`.
- Acceptance: HUP/INT/TERM precedence applies to initialization/inspection failures as well as placement/upload/run/download.

## F7b — major — status-disambiguation SSH ignores cancellation

- Mechanism: `internal/run/run.go:82`, exact bytes `out, err := exec.Command("ssh", remote.SSHArgs(r.Address, "cat "+remote.Quote(r.Meta+"/status"))...).Output()`. After command 125/255, lookup uses neither ctx nor remote.Command; HUP/INT/TERM cannot close this request. The watcher has already finished, yet offload remains stuck on a stalled lookup instead of honoring X3.
- Red: `test/e2e/reviewer_guarantees_test.go` `TestG7SignalDuringStatusLookupCancels`. Fake SSH delays only the status lookup after remote command 125; SIGTERM offload; it remains blocked >12 s. Killing the test's lookup process releases it to the correct 143, proving the pending request owns the delay.
- Failure: `C3/X3 signal did not cancel status lookup after watcher finished; still blocked after 12s`.
- Acceptance: status lookup participates in cancellation; signal status wins and the lookup child is reaped.

## Reproduction + controls

Target production revision = `cac258b87885d725fbbf45d78afa10e1f9a1cd60`; test-only commit = `c76bc623338352b840db96f5c7a878cde009a039` on `wt/reviewer-3`. Production unchanged. All paths below resolve under `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-3` unless stated otherwise.

Verdicts = 7/7 adjudicated; 1 pass (G2), 6 finding rows; 7 distinct findings = 5 major behavior defects + 2 minor README overclaims. Validator positive control: exactly 7 G-id rows (rc 0); unknown-cell validator → `0` (rc 1 = no matches).

- Seven reds: `cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-3 && GOCACHE="$PWD/.scratch/go-build" go test ./test/e2e -run '^TestG[1-7]' -count=1 -timeout=90s` → rc 1; 25.849 s Go / 26.243 s wall; all seven named tests fail with the messages above.
- Existing controls: same cwd/env; `go test ./test/e2e -run '^Test(C3ExitPrecedence|X3SignalsSkipDownload|Y4GitUploadDifferential|Y5Prune|Y5PruneSymlinkAncestor|Y6Download|Y7PlainMirror|Y8Get|Y8GetKeepsNewerLocal|Y9NoSync|Y10UploadFailure|Y10DownloadFailureOverridesCommand|Y11RemoteLock|Y11LivePredecessorAfterControllerDeath)$' -count=1 -timeout=90s` → rc 0; 10.549 s Go / 10.759 s wall.
- Pinned formatter: `/home/eturkes/Projects/offload/.agent/tool golangci-lint fmt --diff test/e2e/reviewer_guarantees_test.go` → rc 0.
- Pinned lint: `/home/eturkes/Projects/offload/.agent/tool golangci-lint run ./test/e2e/...` → rc 0, `0 issues.`; private GOCACHE/GOLANGCI_LINT_CACHE.
- Limits: no real SSH hosts or network partitions; fake SSH + real rsync, isolated homes/slots. Go LSP unavailable (findReferences on internal/run/run.go). Full gate, whole test suite, race tests, mutation testing not run. Existing positive controls cover named cases, not every possible schedule. G2's partition reasoning is inference from lock ownership, not a live partition experiment.
- Register outside check set: none.
