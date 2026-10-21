Target = 76d1d7c + `.scratch/review-8/row10.diff`, reconstructed at `.scratch/worktrees/reviewer-8`; primary working tree excluded. K1–K8 fixed before target read. Complete: 8/8 rows adjudicated; 4 pass, 4 finding rows; 3 distinct findings. Target tracked diff remains byte-identical to the brief's diff (`cmp` rc 0).

| id | check | verdict | finding ids |
|----|-------|---------|-------------|
| K1 | X1 teardown: session + scope TERM/KILL, both empty; leader exit + EOF, pre-setsid/admission windows | pass | — |
| K2 | Fallback: unstarted scope → exactly one plain CMD; manager noise suppressed, CMD stderr retained | pass | — |
| K3 | X2: argv/env/cwd/stdin/signal/fd/status semantics through scope | finding | F1 |
| K4 | S5/Y11: fd 3 / fd 9 stay with watcher through scope teardown, never CMD/scope | pass | — |
| K5 | Shell correctness: quoting, nounset, dynamic scope, PID/mark/wait lifecycle | finding | F1 |
| K6 | Contract X1/X7 + README claims match implementation | finding | F2 |
| K7 | Added/changed tests detect named defects; no vacuous pass | finding | F3 |
| K8 | CI ubuntu-24.04 user manager + OFFLOAD_REQUIRE_SCOPE cannot mask absence | pass | — |

## F1 — major — new shell locals overwrite inherited CMD environment

- Target: `internal/run/watch.bash:24–25,40–41,53–59`.
- X2 = env inherited. Bash locals inherit a pre-existing variable's export attribute. The new `unit`, `mark`, `rt`, `hx`, `vx`, `hi`, `vi`, `d` locals therefore replace caller-provided exported values in CMD. The scope wrapper's `read -r cg` also replaces an exported `cg`. Fallback has the same newly introduced local-variable corruption. No request/contract exception permits it.
- Base 76d1d7c preserves the same selected variables; target changes them. With an ordinary environment, full `env -0` baseline/target comparison found no delta except intentionally distinct probe TMPDIRs; this defect is specifically the exported-name collision.
- Red test added only in reviewer worktree: `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-8/internal/run/reviewer8_test.go`.

Command:

```sh
(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-8 && OFFLOAD_REQUIRE_SCOPE=1 go test -count=1 ./internal/run -run '^TestReviewer8InheritedLowercaseEnvironment$')
```

Target result = rc 1; both subtests fail. Distinctive output (`review8-env-red.log`):

```text
--- FAIL: TestReviewer8InheritedLowercaseEnvironment (0.76s)
X2 inherited env changed: rc=0 err=<nil> read=<nil> got="offload-3189574-958410947.scope\n/tmp/tmp.JPniXJXC4m\n/run/user/1000\n1\n/run/user/1000\n\n\n/run/user/1000\n0::/user.slice/user-1000.slice/user@1000.service/app.slice/offload-3189574-958410947.scope\n" want="callers-unit\ncallers-mark\ncallers-rt\ncallers-hx\ncallers-vx\ncallers-hi\ncallers-vi\ncallers-d\ncallers-cg\n"
X2 inherited env changed: rc=0 err=<nil> read=<nil> got="\n\n\n1\n/run/user/1000\n\n\n/run/user/1000\ncallers-cg\n" want="callers-unit\ncallers-mark\ncallers-rt\ncallers-hx\ncallers-vx\ncallers-hi\ncallers-vi\ncallers-d\ncallers-cg\n"
FAIL
```

Independent baseline/target probe = `python -I /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-8/review8-probe.py exported-env` (base watcher saved from `git show 76d1d7c:internal/run/watch.bash`). Initial eight-variable probe: baseline returns every `callers-…` value; target returns internals; assertion rc 1. Probe now also includes `cg`.

Acceptance: preserve inherited exported values/presence for these names on both paths; rerun the red test. Renaming internals alone reduces collision probability, not X2's input domain.

## F2 — minor — README promises scope containment whenever a user manager exists

- Target: `README.md:23,124–126`; target contract: `.agent/contract.md:57,63`.
- Exact bytes matched by `rg -Fn`:
  - `When a systemd user manager runs for your user, offload puts the command in its own systemd scope.` (`README.md:124`)
  - `Then a process that the command started also stops when it starts its own session, for example a daemon.` (`README.md:125`)
- These conditions are insufficient. X7 requires systemd ≥254 + cgroup v2; X1 explicitly allows refused/failed scope admission to fall back silently. The new refused-scope tests exercise exactly that fallback with a live manager. X1 also explicitly excludes a process adopted by another systemd unit, including a nested offload scope; README does not qualify its daemon-stop guarantee.
- Acceptance: align human-facing claims with X1/X7: compatible manager + successful admission; otherwise session-only teardown; adoption by another unit remains outside scope containment. No change requested to the accepted fallback/nested-run design rulings.

## F3 — minor — fallback test rejects a correctly inherited caller scope

- Target: `internal/run/run_test.go:220–221`.
- Exact bytes: `strings.Contains(string(cg), "offload-")` (rg match at line 220).
- Fallback creates no new scope; it does not remove CMD from its caller's pre-existing scope. If this test runs inside `offload run --local -- …` (the prescribed heavy-suite runner after the new binary is installed), CMD correctly inherits the outer `offload-…scope`. The new assertion reports a defect anyway. This is a test oracle error, not the accepted nested-run teardown limitation.
- Reproduced with an actual transient caller scope; no modified watcher. Command:

```sh
(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-8 && systemd-run --user --scope --quiet --collect --expand-environment=no --unit="offload-reviewer8-$BASHPID.scope" -- go test -count=1 ./internal/run -run '^TestRefusedScopeRunsCommandWithoutScope$')
```

Result = rc 1 (`review8-nested-scope-red.log`):

```text
--- FAIL: TestRefusedScopeRunsCommandWithoutScope (0.22s)
    run_test.go:221: rc 6 err <nil> runs "run\n" cgroup "0::/user.slice/user-1000.slice/user@1000.service/app.slice/offload-reviewer8-3205746.scope\n"
FAIL
```

The direct unwrapped target test passed in the initial regression run. `rc 6`, exactly one `run`, and the unchanged outer scope show correct fallback in the red case. A separate actual target-binary wrapper attempt waited on occupied local slots; it was stopped before execution (owned PID 3203013, TERM), so that attempt is not claimed as executed coverage.

Acceptance: compare CMD's fallback cgroup with the caller's cgroup (or otherwise distinguish a newly created scope from the inherited one); both unwrapped and offload-wrapped checks pass while a mutant that starts a new scope still fails.

## Passing-row evidence + limits

- K1/K2/K3/K4: `OFFLOAD_REQUIRE_SCOPE=1 go test -count=1 ./internal/run ./test/e2e -run 'Test(LocalStatus|LocalArgvIsNotShellParsed|LockReachesWatcherNotCommand|Scope|RefusedScope|EOFDuringScope|X1Scope|X1RefusedScope|X2ArgvAndEnvironment|X4ByteExactStreaming)'` in reviewer worktree → rc 0 (`internal/run` 18.674 s; `test/e2e` 9.075 s). Includes local + fake-SSH remote leader exit/controller SIGKILL, escaped daemon, TERM-ignoring escaped daemon, scope identity, env presence, refusal + stderr, argv, byte-exact streams, admission cancellation. No skip expected with mandatory-scope setting; log = `review8-baseline.log`.
- K3: `python -I …/review8-probe.py matrix` → rc 0: base + target statuses 126/127/255 and INT/QUIT/TERM 130/131/143; ordinary env byte comparison otherwise equal. Initial SIGQUIT probe used Bash's own intentionally ignored QUIT disposition and was invalid; corrected to Perl, all default-signal checks pass.
- K1/K5: `python -I …/review8-probe.py early-eof` → rc 0. A mktemp barrier delays launch; positive control confirms live EOF reader; close controller pipe, prove reader exited, release marker allocation. Target under `bash -u` returns 143, stdout `started`, never `finished`. Suspected lost-EOF race rejected by this evidence. This is one controlled ordering, not exhaustive scheduler exploration.
- K1/K4: `python -I …/review8-probe.py locks` → rc 0. Four cases = fd3/fd9 × leader-exit/EOF; a verified scoped, own-session, TERM-ignoring daemon survives into grace. `flock -n` rc 1 during teardown, rc 0 afterward; daemon gone. Watcher statuses = 5 (leader exit) / 143 (EOF). CMD explicitly checks fds 3–9 absent. Tests use real local systemd 262.1 + cgroup v2; remote lock semantics exercised through the shared watcher with fd9, not a live nixbox run.
- K5/K8 mechanical checks: target `watch.bash` shellcheck 0.11.0 `-S warning` → rc 0; target `gate.yml` actionlint 1.7.12 → rc 0.
- K7 sensitivity: source stays unchanged; review-only `internal/run/reviewer8_overlay_test.go` sets the embedded watcher string from `REVIEW8_WATCHER` in test memory. Against `review8-base-watch.bash`, the six new internal scope tests all fail on their named behavior: leaked daemon (leader exit + cancel), missing scope, missing scoped environment, no ten-second escaped-daemon grace, CMD running during delayed admission. Command: `OFFLOAD_REQUIRE_SCOPE=1 REVIEW8_WATCHER=<worktree>/review8-base-watch.bash go test -count=1 ./internal/run -run '^(TestScopeEndsEscapedDaemonWithLeader|TestScopeEndsEscapedDaemonOnCancel|TestScopeHoldsCommand|TestScopeKeepsCommandEnvironment|TestScopeKillsTermIgnoringDaemonAfterGrace|TestEOFDuringScopeAdmissionRunsNothing|TestReviewer8InheritedLowercaseEnvironment)$'` → rc 1, `review8-base-red.log`. F1's reviewer test passes separately against that base (rc 0).
- K7 planted defects, separate source copies selected through the same test-memory hook: remove `--expand-environment=no` → `TestLocalArgvIsNotShellParsed` rc 1; remove both env-restoration lines → `TestScopeKeepsCommandEnvironment` rc 1; disable fallback branch → `TestRefusedScopeRunsCommandWithoutScope` rc 1. Files/logs = `review8-mutant-{expand,env,fallback}.{bash,log}`. These are targeted mutation samples, not a full mutation run; no survivor in this selected set.
- K7 e2e sensitivity: Go overlay JSONs each map exactly one target path to a separate fixture, leaving disk target untouched; GOFLAGS propagates into e2e's CLI build. `OFFLOAD_REQUIRE_SCOPE=1 GOFLAGS=-overlay=<worktree>/review8-overlay-noise.json go test -count=1 ./test/e2e -run '^TestX1RefusedScopeFallsBackQuietly$'` → rc 1; local + remote both show `systemd-run: refused` beside `own-stderr` after removing only launcher's stderr suppression. `GOFLAGS=-overlay=<worktree>/review8-overlay-base.json` with `-run '^TestX1ScopeEndsDoubleForkedDaemon$'` → rc 1; all four local/remote × leader-exit/controller-kill cases report a remaining daemon. Logs = `review8-mutant-noise.log`, `review8-e2e-base-red.log`.
- K8 negative control: prepend `review8-refused-bin` to PATH and set `OFFLOAD_REQUIRE_SCOPE=1`; `go test -count=1 ./internal/run ./test/e2e -run '^(TestScopeHoldsCommand|TestX1ScopeEndsDoubleForkedDaemon)$'` → rc 1. Both fail with `OFFLOAD_REQUIRE_SCOPE is set, but systemd-run --user --scope fails` (`review8-required-scope-red.log`); no silent skip.
- K8 judgment: `loginctl enable-linger` failure is fail-closed under GitHub's Bash step; bounded socket wait then hard failure. A stale/inaccessible socket cannot green the scope tests because `OFFLOAD_REQUIRE_SCOPE=1` upgrades failed scope admission to fatal. Actual GitHub ubuntu-24.04 VM provisioning remains **unconfirmed**: workflow + test code inspected and local systemd path exercised; no push/workflow execution performed. Full gate, race suite, real-host runs not run by this review.
- LSP coverage: real `run.go` reference request returned `No LSP server available for file type: .go`; caller/callee search performed with rg + file reads instead.

## Register + cleanup

- No additional off-check-set defect registered. Previously tentative caller-scope concern was reproduced and adjudicated as F3 (test oracle, not nested-run teardown design).
- Test-owned leaked-daemon mutation PIDs + queued runner PID absent on final `ps` (rc 1, no listed PID); positive control `pgrep -c -x bash` expected ≥1, observed 32. Probe temp dirs self-remove; temporary review CLI binary removed. Red tests, immutable baseline/mutant fixtures, overlays + logs remain in reviewer worktree for harvest/replay; worktree deliberately retained per brief. No commits, staging, primary-tree target edits, host config changes or remote writes.

reviewer-8-DONE-1
