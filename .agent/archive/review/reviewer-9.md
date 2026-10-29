Target = e5f45a5..547e86b; checkout `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-9`. R1–R7 fixed before target read. **7/7 adjudicated: 2 pass rows, 5 finding rows; 5 distinct findings (1 major, 4 minor).** Tracked target unchanged; review-only tests/fixtures + logs remain in worktree.

| id | check | verdict | finding ids |
|----|-------|---------|-------------|
| R1 | X5 multiplex: every connection to a `multiplex = true` host (probe, upload ×2, run, status lookup, download, --get) carries the 3 options; other hosts none; unusable socket dir → plain + one note; -T on every ssh incl. rsync; no shared-state or quoting hazard | finding | F3 |
| R2 | X8 boot: fails closed (125, nothing run, no temp left) on short / corrupt / unwritable copies; argument mapping after `shift 2`; content-addressed name; a file already at the name never runs instead; `head -c` stdin accounting leaves the frame + EOF pipe intact | finding | F2 |
| R3 | Y5/Y6 framing: NONCE, EXCLUDES, COUNT paths; count+2 check before any prune; excludes reach `git_list` both before CMD (prune) and after (download list) | finding | F1 |
| R4 | X2: every variable in remote.bash + boot + watch.bash is `__…`; no plain-named assignment or local left that could shadow a caller export | pass | — |
| R5 | X1/X3 TERM/HUP trap: teardown paths unchanged otherwise (leader exit, EOF, relaunch, KILL escalation, pre-setsid); no infinite loop or lost status; X6 status record still written | pass | — |
| R6 | Claims: contract P2, X1, X2, X5, X7, X8, Y5 + README + pool.example.toml match the code exactly (no overclaim, no stale text) | finding | F1, F2, F3, F4 |
| R7 | Verification integrity: each new/changed test fails on the defect it names; no vacuous pass (skips, timing, oracle bugs, environment-dependent passes) | finding | F5 |

## F1 — major — NUL in global excludes shifts the upload frame and deletes uploaded files

- Target: `internal/run/run.go:87–89`; consumer `internal/run/remote.bash:47–57`.
- `GlobalExcludes` returns file bytes without validation. `command` puts them directly into a NUL-delimited record. A NUL inside those bytes becomes another record; `head` still sees the requested count+2 and the expected nonce, but stops before the last real upload path. Prune then deletes that freshly uploaded path as R−U. This violates Y5; the checked count is not a check of the intended records.
- Accepted input is not hypothetical Git rejection: excludes bytes `*.tmp\x00\n` pass a positive control (`git check-ignore out.tmp` → exactly `out.tmp\n`). Git still applies the pattern, while offload's framing fails.
- Red test = worktree `test/e2e/reviewer9_test.go`, `TestReviewer9GlobalExcludesNULPreservesUpload`.

```sh
(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-9 && env -u TMPDIR go test -count=1 -v ./test/e2e -run '^TestReviewer9GlobalExcludesNULPreservesUpload$')
```

Target rc 1 (`review9-adversarial-e2e.log`):

```text
Y5 valid upload lost: rc=1 stdout="" stderr="offload: alpha slot 1/2 (8 CPUs, 25% busy) → ~/offload/project-f2f4a37d\ncat: tracked: No such file or directory\n" remote tracked exists=false
--- FAIL: TestReviewer9GlobalExcludesNULPreservesUpload (0.28s)
```

Acceptance: encode excludes losslessly, or reject a NUL before any prune; a malformed excludes record never silently changes U. Pin ordinary + NUL-bearing inputs and preserved uploaded paths.

## F2 — minor — boot treats an existing script-name directory as an install destination

- Target: `internal/run/run.go:43,49` (`mv -f "$__f.$$" "$__f"`; `exec bash "$__f" "$@"`).
- GNU `mv` without `-T` treats a directory or symlink-to-directory at the destination as a directory to move into. It succeeds, so boot bypasses its 125 cleanup path. Bash then opens the directory, exits 126, and leaves the temporary copy inside it. No CMD ran, but `Remote.Run` treats this as an ordinary command status (only 125/255 trigger setup-record recovery). X8's failed-install status/cleanup contract is broken.
- Red test = worktree `internal/run/reviewer9_test.go`, `TestReviewer9BootDestinationDirectoryFailsClosed` (directory + symlink-to-directory).

```sh
(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-9 && env -u TMPDIR go test -count=1 -v ./internal/run -run '^TestReviewer9BootDestinationDirectoryFailsClosed$')
```

Target rc 1, both cases (`review9-adversarial-unit.log`):

```text
X8 setup failure: rc=126 (want 125), CMD ran=false, misplaced temp count=1, stderr="offload/.meta/w/run-348a181e0d21478b.bash: offload/.meta/w/run-348a181e0d21478b.bash: Is a directory\n"
```

Acceptance: rename targets the exact script pathname, never a directory's contents; refusal returns125, starts no CMD, leaves no temporary copy. Preserve correct replacement of ordinary stale files.

## F3 — minor — existing unwritable socket directory disables a reachable host instead of falling back

- Target: `internal/remote/ssh.go:62–65`.
- `MkdirAll(dir,0700)` succeeds when dir already exists, even when the user cannot create a socket there. `Multiplex` returns its options, not an error; `targets` emits no fallback note. OpenSSH then fails the probe with `Permission denied`, so offload marks the reachable host down and a pool run waits instead of connecting plainly. R1's unusable-dir behavior + README's fallback claim are false for this case.
- Red test = worktree `internal/remote/reviewer9_test.go`, `TestReviewer9UnwritableSocketDirectoryRefused`; fixture first proves that creating a file in the 0500 directory fails.

```sh
(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-9 && env -u TMPDIR go test -count=1 -v ./internal/remote -run '^TestReviewer9UnwritableSocketDirectoryRefused$')
```

Target rc 1 (`review9-mux-unwritable-red.log`):

```text
X5 unusable socket dir accepted: opts=["-o" "ControlMaster=auto" "-o" "ControlPath=/tmp/r9mux-3281287774/offload/ssh/%C" "-o" "ControlPersist=60"] err=<nil>; want plain fallback
```

- Real-host confirmation, freshly built target `review9-offload`, scratch `review9-pool.toml`, socket dir only under reviewer worktree. `XDG_RUNTIME_DIR=/proc/self/cwd/.r9` supplies a short absolute alias into that worktree. Existing `.r9/offload/ssh` mode0500 → `offload status` rc0 but host `down`:

```text
reviewer-9     down     0/2  unix_listener: cannot bind to path /proc/self/cwd/.r9/offload/ssh/b956391f443d8cd522401e5ca3da52f8aad598c2.m4wWRoKEbMgpuscQ: Permission denied
```

- Same host, direct plain SSH immediately afterward → stdout `reachable`, rc0. Log = `review9-mux-unwritable.log`. Runtime mode restored to0700. No successful mux master started by this probe; no remote project writes.
- Acceptance: verify existing socket directory usable before enabling options, then preserve the one-note/plain fallback across every host/call. Pin unwritable existing dir in addition to absent/bad/long paths.

## F4 — minor — README omits the accepted nested-command exception

- Target: `README.md:136`; exact bytes matched by `rg -Fn`:

```text
An offload run that the command starts stops together with its own command.
```

- X1 at `.agent/contract.md:57` explicitly limits that new guarantee. Distinctive matched bytes:

```text
a TERM-ignoring inner command can outlive it, as the outer KILL may end the inner watcher first
```

- The README sentence promises an unconditional nested-command stop while the contract accepts this survivor. This review preserves that design ruling; the defect is the new human-facing overclaim, not a request to redesign nested teardown.
- Acceptance: qualify the README statement with the existing TERM-ignoring exception.

## F5 — minor — known excludes-as-path mutant is observable, but its regression is absent

- Target: `test/e2e/sync_test.go:197–214`, especially line201 (`s.write(excludes, "*.tmp\n")`); mechanism `internal/run/remote.bash:55`.
- Independent ruling of commit32ecee1's noted `tail +3 → +2` survivor: **specified behavior, not equivalent**. An excludes file containing exactly `!stale` does not ignore a path literally named `!stale`; treating that whole record as an upload path wrongly protects the remote-only file from Y5 prune. Current `+3` code is correct. This finding is the missing regression required by Verification integrity for a known, contract-relevant survivor.
- Review-only regression = worktree `test/e2e/reviewer9_test.go`, `TestReviewer9ExcludesRecordIsNotAnUploadPath`. Target → pass. The same test with the one-line `+2` overlay → red; existing `TestRemotePrune` + `TestY5GlobalExcludesReachTheHost` stay green under that defect.

```sh
(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-9 && env -u TMPDIR GOFLAGS="-overlay=$PWD/.review-plants/tail-plus-two/overlay.json" go test -count=1 -v ./internal/run ./test/e2e -run '^(TestRemotePrune|TestY5GlobalExcludesReachTheHost|TestReviewer9ExcludesRecordIsNotAnUploadPath)$')
```

Result rc1 (`review9-tail-survivor.log`):

```text
--- PASS: TestRemotePrune (0.15s)
Y5 excludes record became upload path: rc=1 stale exists=true stderr="offload: alpha slot 1/2 (8 CPUs, 25% busy) → ~/offload/project-37146e91\n"
--- FAIL: TestReviewer9ExcludesRecordIsNotAnUploadPath (0.55s)
--- PASS: TestY5GlobalExcludesReachTheHost (0.56s)
```

- This is a test-sensitivity finding: failure is on the explicit defect overlay, not a claimed failure of unmodified production code. Only these existing relevant checks were rerun against this survivor; the commit's full-suite-survival claim was not independently rerun.
- Acceptance: commit a regression for this record/path distinction, prove it green on target + red on `+2`, and record the survivor as killed rather than equivalent. The supplied review test is one executable counterexample.

## Evidence + limits

- Read whole touched execution/transport files, CLI callers, harness; contract/README/example; full target diff. LSP Go request → no server; rg + enclosing-function reads instead.
- Targeted unit checks (`review9-target-unit.log`) → rc0: transport options/bounds/modes, pool parsing, all Boot/Remote tests, watcher TERM-before-setsid, status/open-stdin/exported-variable checks.
- Initial e2e run put TMPDIR inside the worktree: X2/X4 non-git sandboxes inherited the parent linked-worktree root and Y3 refused before execution. Invalid review setup, not target regression. Rerun with `env -u TMPDIR` → rc0 (`review9-target-e2e-rerun.log`), including new mux/boot/nested/excludes + X2/X4/X6 checks.
- Teardown selection (`review9-teardown.log`) → rc0: leader exit, cancellation, scope escaped daemon, refused scope fallback, ten-second escalation, admission EOF, local/remote controller death, all HUP/INT/TERM CLI statuses + no download, Y11 successor serialization, pre-setsid cancellation, signal during status lookup. `OFFLOAD_REQUIRE_SCOPE=1`; no silent scope skip.
- R4: all internal locals/assignments reserved `__…`; intentional XDG_RUNTIME_DIR/INVOCATION_ID restore remains the already-ruled scope mechanism; IFS/LC_ALL assignments command-scoped. Local existing exported-name tests + review remote exported-name probe pass.
- 18 isolated overlay plants → all killed by named changed tests. Command = `(cd <worktree> && offload run --local -- python -I "$PWD/review9-plants.py")`; summary `.review-plants/summary.json`, full per-plant output retained, root log `review9-plants.log`. Includes mux option/bound/mode/percent/tty/note/parser/type; boot hash/cleanup/error/shared-name; frame count/excludes record/empty; watcher trap/default-who; preboot revision58a169e vs process-title test. These are selected plants, not exhaustive mutation.
- R5 direct watcher probes: `OFFLOAD_REQUIRE_SCOPE=1 go test -count=1 -v ./internal/run ./test/e2e -run '^(TestReviewer9WatcherSignalsRecordStatus|TestReviewer9ExcludesRecordIsNotAnUploadPath)$'` → rc0 (`review9-signal-and-tail-baseline.log`). TERM + HUP to the remote watcher kill its running command, return143, and write the exact nonce/status record. Tail distinction passes on unchanged target. The trap's post-function lifetime was considered; no new accepted-contract failure confirmed, so no speculative finding.
- Real OpenSSH controller/host path: scratch-config fresh target binary `status` cold + warm → host up, same owned master PID3496365. Kill that observed owned master, leave its socket, repeat status → host up through new PID3497570; close it through its private ControlPath. Logs `review9-mux-live.log`, `review9-mux-stale.log`.
- Real mux lifetime: `python -I /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-9/review9-mux-lifecycle.py` → rc0 (`review9-mux-lifecycle.log`). Shared target watcher sent directly over SSH; cold master + stdin EOF → watcher143 + CMD gone; warm master + client SIGKILL → that CMD gone while an independent concurrent client prints `B-done`, rc0; master3499868 remains alive until explicit private-socket close. This exercises actual SSH fd/session behavior, not real-host boot/sync; the latter was exercised through the local/fake-host tests above. Remote writes limited to the watcher's own transient `/tmp` marker; no pool config or host project edits.
- ShellCheck0.11.0 `-S warning` over target watch.bash + remote.bash → rc0. No Go LSP available. Selected plants = 18/18 killed; the extra known tail survivor now has the independent F5 ruling + red oracle, not a claim of exhaustive mutation.
- Final target reds rerun together → rc1 with exactly the three implementation failures F1/F2/F3 (`review9-final-reds.log`). Review tests permit safe alternatives: NUL refusal125 before prune; exact verified replacement of the destination symlink; making a socket directory usable before opting in. No fix choice is imposed by the red tests.
- Unconfirmed/not run: cold probe deadline during persistent-master detachment; real-host full boot/upload/download/get sequence; timing/performance numbers from commit bodies; all OpenSSH releases; exhaustive scheduler/load exploration; full gate/race suite/CI/pin-release validation. This was a bounded changed-unit review, not those full gates. The old scope unit's accepted design rulings were not reopened.

## Register + cleanup

- No off-check-set defect registered. F5 = explicit verification gap, not a defect in the current `+3` implementation.
- All review-owned SSH clients reaped and masters closed/killed. Final `ps` over recorded PIDs3496365/3497570/3499868/3499871 → rc1, no output; positive control = current shell expected1, observed1, rc0. Other sessions' processes left untouched.
- Removed review binary + disposable `.r9`/`.review-tmp` directories. Red tests, overlays, reproducible plant/lifecycle scripts, scratch pool file + logs retained as validator/replay inputs. Worktree deliberately retained for MAIN. No commits, staging, primary-tree source edits, instruction edits, account/host configuration changes or persistent remote project writes.
- Final `git diff --exit-code` in worktree → rc0; HEAD=547e86b. Main target remains read-only.

reviewer-9-DONE-1
