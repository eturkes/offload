| id | verdict | key risk | evidence | test that exposes it |
|----|---------|----------|----------|----------------------|
| A | amend | Direction right; exact launch expands CMD arguments. Setup-failure status, scope-empty predicate + mandatory bus coverage also need amendment. | systemd v260 `src/run/run.c:78,2749–2761,2833–2865`; `src/core/scope.c:26–33`; `/home/eturkes/Projects/offload/test/e2e/harness_test.go:90–100`. | Literal `$HOME`/`${VAR}`/`$$` under a verified scope; escaped TERM-ignoring child; setup failure after successful probe; controller death during admission; absent vs reachable bus. |
| B | amend | FD/EOF premise is sound. Define socket-directory/quoting/length behavior, foreground-probe vs persistent-master lifetime, and real-SSH coverage. | OpenSSH 9.0 `ssh.c:1706–1739`, `mux.c:398–417,1881–1926,2265–2285`; 10.5 equivalent code verified; `/home/eturkes/Projects/offload/internal/probe/probe.go:62–78`. | Cold/warm mux SIGKILL, independent concurrent run, probe timeout, dead-master socket, whitespace/long runtime path; real no-op timing comparison. |
| C | wrong | Literal BOOT selects the wrong script path after `shift`; undefined `fail` can continue after setup failure; short input passes `head`; common filename has a publish→open race. | `/home/eturkes/Projects/offload/.scratch/consult-2/brief.md:55–62`; GNU coreutils `head.c:760–780`; shell probes + remote inode model below. | Distinct META/WORKDIR, failed bootstrap with stale script, valid short shell prefix at EOF, competing script versions paused before open; inspect remote argv. |

## A

**Direction right:** retain the external watcher + new session; add one transient scope for CMD and its descendants. The scope fixes `setsid` escape for descendants that remain in that scope. Keep the watcher and its lock fds outside it. Source supports direct exec (same launch PID), and the existing session/early-child checks remain necessary before scope admission.

### Amendments

1. **Blocking: add `--expand-environment=no`.** systemd v260 initializes `arg_expand_environment = true` (`src/run/run.c:78`), then its scope path calls `replace_env_argv(arg_cmdline, env, ...)` (`:2843–2860`) before `execvpe` (`:2863`). Installed systemd-run(1) states: “Expand environment variables in command arguments. If enabled (the default) … With --scope, this expansion is performed by systemd-run itself.” Shell quoting into `bash -lc` does not prevent this earlier argument rewrite. Existing X2 regression words already include `$HOME` (`/home/eturkes/Projects/offload/test/e2e/execution_test.go:130–142`); add `${VAR}`, `$$`, unset variables + a quoted remote shell program. Probe the **same required flag set** used for CMD, including `--collect` + disabled expansion; otherwise an older binary can pass the simplified probe and reject the real launch.

2. **“Same env” is not exact.** v260 obtains the new scope invocation ID and adds `INVOCATION_ID=…` (`src/run/run.c:2754–2761`), then merges it over the inherited environment (`:2833`). Preserve the intended X2 semantics explicitly: either restore the caller's value/presence in a small launch wrapper, or approve a narrow `INVOCATION_ID` exception. PID/cwd inheritance is compatible with the source; full environment equality is not. Local read-only check found the controller's `INVOCATION_ID` unset.

3. **Successful capability probe ≠ successful actual admission.** v260 can fail `StartTransientUnit`, job waiting or pre-exec setup before CMD (`src/run/run.c:2731–2756`), and `execvpe` failure returns an error (`:2863–2865`). The current watcher simply forwards its child status (`/home/eturkes/Projects/offload/internal/run/watch.bash:25–27,45–49`); the remote script then records that as CMD status (`/home/eturkes/Projects/offload/internal/run/remote.bash:60–69`). Define setup failure → 125, distinguish it from a real CMD's exit 1/125/255, and preserve 126/127 execution-error behavior. An admission/start marker is one approach; retrying CMD after an ambiguous launch is not. Also bound capability/setup queries so an unresponsive manager does not turn cancellation into an indefinite pre-launch wait.

4. **`systemctl is-active` nonzero is not proof of emptiness.** v260 scope state mapping is `[SCOPE_STOP_SIGTERM] = UNIT_DEACTIVATING`, `[SCOPE_STOP_SIGKILL] = UNIT_DEACTIVATING` (`src/core/scope.c:26–33`); these states can retain processes. A bus/query error also returns nonzero. systemctl(1) itself says its status-code mapping is imperfect and recommends specific states/substates. Require verified scope emptiness, not merely `! is-active`: capture the admitted cgroup identity and inspect recursive population (`cgroup.events` on cgroup v2), or otherwise distinguish populated/transitional/error states from confirmed gone. Keep the session/parent-child check until launch has settled. TERM/KILL must explicitly target the whole scope. S5/Y11 locks remain held until both containment sets are empty.

5. **Bus coverage must be explicit, not ambient.** `/home/eturkes/Projects/offload/test/e2e/harness_test.go:93` replaces `XDG_RUNTIME_DIR` with an isolated directory; it retains an inherited `DBUS_SESSION_BUS_ADDRESS` via `replaceEnv` (`:124–137`). Enabling linger + exporting only the real XDG directory before the suite does not fix this override. Export the real bus address for the scope suite, assert CMD's `/proc/self/cgroup` names its `offload-…scope`, and fail that required suite if scope coverage is unavailable. Force fallback separately. Live read-only probes:
   - `env -u DBUS_SESSION_BUS_ADDRESS systemctl --user show app.slice …` → `ActiveState=active`, rc 0: unsetting the address alone does **not** force fallback.
   - Unset both address + XDG runtime → diagnostic “$DBUS_SESSION_BUS_ADDRESS and $XDG_RUNTIME_DIR not defined”, rc 1.
   - CI `/home/eturkes/Projects/offload/.github/workflows/gate.yml` currently has no user-manager/bus setup. Exact proposed CI provisioning remains unconfirmed; it was inspected, not changed/run.

**Placement check passed (read-only nixbox):** systemd 260.5; app.slice ControlGroup `/user.slice/user-1000.slice/user@1000.service/app.slice`; trading `user-1001.slice` has `CPUWeight=idle`. This supports the sibling-slice claim. Controller oomd monitoring + actual scope placement were not independently exercised here; the brief's measurements remain unconfirmed by this review.

**Recommended predicate:** when the exact noninteractive scope capability succeeds, admit CMD into a uniquely named scope before executing it, with argv/cwd/signal/stdio semantics preserved and the environment exception decided. Capability unavailable → documented session-only fallback. Actual launch/setup failure → 125 without CMD execution. EOF or leader exit → TERM, then KILL after the existing grace, covering session + scope; release only after verified emptiness. Nested scope migration remains a separately recorded limitation, as the brief proposes; avoid claiming that arbitrary cgroup migration is contained.

**Required tests:** verified scope + verified fallback; existing X2 byte/argv/status tests on both; double-fork/setsid daemon after leader exit and controller SIGKILL; TERM-ignoring escaped descendant through the ten-second escalation while S5/Y11 locks stay held; EOF before setsid, during admission, after admission; probe success followed by launch failure; scope in deactivating state with a live escaped process; manager-query error distinct from absence. No transient scopes or manager mutations were created in this review, so PID/signal and whole-cgroup teardown assertions remain unconfirmed here.

## B

**Direction right:** host-scoped opt-in, default false. Source confirms the central EOF mechanism; no reason to reject ControlMaster on that basis.

### Verified mechanism

- OpenSSH 9.0 `mux.c:398–417` receives three fds from the client; `:1923–1926` sends stdin/stdout/stderr with `mm_send_fd`. The master's copy of the offload lifetime pipe is a **read end**, not an extra writer. Closing offload's write end still produces EOF.
- `ssh.c:1706–1739` forks the persistent master, reconnects the original foreground process as a mux client, then calls `stdfd_devnull(1, 1, …)` + `daemon(1, 1)`. Normal master fd 0/1/2 do not hold the initial probe's stdio open. The debug-stderr exception exists in source; do not claim unconditional redirection under `ssh -v`.
- A dead mux client also triggers `mux_master_control_cleanup_cb` (`mux.c:213–239`), which shuts down that session's channel IO. Other sessions remain separate.
- `-n` is per-client: `mux.c:1881–1882` replaces that client's stdin before fd passing. A probe created with `-n` does not force later clients to use null stdin. Keep it off the run connection and rsync transport. Current Go Probe already has nil `cmd.Stdin`, hence `/dev/null`; explicit `-n` is clarity, not a required EOF fix (`/home/eturkes/Projects/offload/internal/probe/probe.go:62–71`).
- Equivalent fd passing, detachment + stale-socket handling were read in OpenSSH 10.5, the nixbox version. Controller is 10.6p1. This verifies two source versions, **not every OpenSSH ≥9 release**.

Correct the wording “no offload pipe outlives the client”: passed session descriptors intentionally live in the master. The useful invariant is **no extra writer keeps the offload→watcher lifetime pipe alive**, and a timed-out probe's output readers return within R1's bound.

### Amendments

1. **Create and encode the socket directory before first use.** The draft path's parent is not automatically created by OpenSSH. Specify an owner-private directory, consistent path resolution for every SSH/rsync call, and quoting at both OpenSSH's `-o` parser and rsync's `-e` parser. Current `/home/eturkes/Projects/offload/internal/remote/ssh.go:29–30` uses `strings.Join`, which is insufficient once option values contain whitespace. Read-only positive/negative probes:
   - `ssh -G -o 'ControlPath=/tmp/offload runtime/%C' …` → `keyword controlpath extra arguments at end of line`, rc 255.
   - `-o 'ControlPath="/tmp/offload runtime/%C"'` → expected expanded path. Literal `%` in the directory also needs token-aware handling.

2. **Specify length/failure behavior.** `%C` compresses the destination, not the runtime prefix. OpenSSH 9.0 checks `sun_path` size (`mux.c:2265–2268`) and creates a temporary listener path with an additional dot + 16-character suffix (`:1271–1289`). Current `/run/user/1000` gives 67-byte final / 84-byte temporary paths, within Linux's limit. Longer XDG paths can fail. Prefer a clear preflight diagnostic (or explicitly documented non-mux fallback); do not let this silently mark an otherwise reachable pool down. Test the temporary-path bound, not only the final one.

3. **Dead-master stale sockets are handled, but a live stuck master is different.** `ECONNREFUSED` with ControlMaster=auto unlinks the stale socket (`mux.c:2282–2285`; 10.5 `:2413–2416`). A live master that accepts but does not answer is not the stale-socket case. R1's outer deadline still matters; ConnectTimeout is not a universal mux protocol deadline. Preserve the existing `Timeout - pipeGrace` + `WaitDelay` behavior (`/home/eturkes/Projects/offload/internal/probe/probe.go:63–78`).

4. **Clarify R1's process ownership.** Its probe client must be killed + reaped by the deadline; an intentional shared master may persist until 60 seconds **idle**, not 60 seconds after creation. Killing a shared master to enforce a probe deadline would kill unrelated runs. A cold probe killed during detachment must not pin captured output until the persist timeout. Real fault-injection coverage is needed; source inspection alone does not settle every scheduling window.

5. **Thread host settings through all call sites.** Pool parsing currently has only name/address/slots (`/home/eturkes/Projects/offload/internal/pool/pool.go:21–33`). Probe/status use only address; remoteRun and mirror functions also accept address alone (`/home/eturkes/Projects/offload/internal/cli/cli.go:228–276`, `/home/eturkes/Projects/offload/internal/mirror/mirror.go:278–313`). One resolved transport value must reach probe, upload, run, status recovery, download + --get. Preserve all X5 options on both masters and clients. `SSHArgs` forces `-T`; `RsyncShell` currently does not. Make no-PTY behavior explicit on rsync too if the contract relies on byte-exact IO regardless of SSH config.

**Required tests:** default/false means no tool-added mux options; true reaches every call; cold/warm actual OpenSSH run, controller SIGKILL and continued availability of a second concurrent run; probe timeout before/after detachment; stale socket after killing a test-owned master; real rc 125/255 status recovery + binary streaming; long/whitespace runtime directory; private test socket cleanup. Fake SSH only executes the command (`/home/eturkes/Projects/offload/test/e2e/testdata/ssh:58–62`): it cannot grade fd passing or persistence. Real mux lifecycle + no-op timing were **not run** in this read-only review. The claimed performance win remains unconfirmed pending a repeated cold/warm comparison with the same no-op workload; record wall-clock runs outside offload itself.

## C

**Direction right; literal draft wrong.** Script-file execution removes implementation text from the steady-state process argv while retaining the lifetime pipe. GNU `head -c N` is suitable for the first frame, with exact-length validation added.

### Confirmed defects

1. **Argument shift:** initial args = `N META WORKDIR …`; after `shift`, META is `$1` and WORKDIR is `$2`. Draft `exec bash "$2/run.bash" …` executes the wrong location. A shell-only probe printed `draft-script-path=offload/w/run.bash meta=offload/.meta/w`. Capture the script path before shifting or use a named-argument bootstrap; keep its mapping aligned with `/home/eturkes/Projects/offload/internal/run/remote.bash:37`.

2. **Failure path:** `fail` is defined inside the yet-unloaded remote script (`/home/eturkes/Projects/offload/internal/run/remote.bash:8–11`), not in BOOT. Literal `false || fail 125; printf …` produced `fail: command not found`, then continued, overall rc 0. Define BOOT's error exit independently and make every download/write/rename/open/exec failure terminate at 125; an existing prior script must never run after bootstrap failure. If `fail 125` was intended as pseudocode, the implementation contract still needs this explicit behavior.

3. **Short-input success:** `printf abc | head -c 10` returns rc 0. GNU coreutils v9.7 `head.c:775–780` breaks on EOF and returns true. Check file byte count equals N **before** publication/execution; return 125 and remove the partial file on mismatch. A short payload can be a valid shell prefix, so syntax checking alone is insufficient. Positive no-overread probe on an OS pipe: input `abcREMAINDER`, `head -c 3` → stdout `abc`, remaining pipe bytes `REMAINDER`, rc 0. Source `:765–769` bounds each read by the remaining count, supporting this part of the draft.

4. **Publish→open race:** atomic rename preserves an inode already opened, not the version a future `bash path` will open. The remote lock is currently acquired only after bash reads/starts the script (`/home/eturkes/Projects/offload/internal/run/remote.bash:36–40`). The controller normally serializes live runs, but S5 permits its lock to free early on controller death; a predecessor can still be in bootstrap while its successor proceeds. A disposable remote `/tmp` model opened old script fd 7, atomically replaced the common path, then ran both: `already-open-inode=old-script`; `open-after-publish=new-script`, rc 0. Use a per-run immutable filename (nonce), or acquire and retain the remote lock before publication + open. A pathname unique to the run is the simplest narrow fix. Define cleanup for setup failure and normal/EOF teardown.

5. **“One short line” needs a bounded claim.** CMD itself can contain arbitrarily long words/newlines. Acceptance should be: steady-state watcher argv contains the script pathname + necessary execution arguments, **not embedded watcher/remote source or global excludes content**. It must not promise a fixed total argv length or conceal user CMD. BOOT's short `bash -c` exists transiently before exec; state whether acceptance targets the steady-state process, as intended.

### Recommended predicate + tests

Receive exactly N script bytes into a run-owned path; validate completeness and terminate 125 on setup failure before executing any script/CMD. Execute that exact immutable file with preserved argument mapping. In git mode, consume + validate NONCE, EXCLUDES and exactly COUNT paths as a complete frame before any prune; update existing count/nonce checks for the added record. The remaining stdin is still the lifetime pipe; no helper retains a write end. Preserve X4, X6 + Y11. Clean run-owned partial/finished files under a defined teardown policy.

Tests: successful plain/git/nosync + distinct META/WORKDIR/subdir; truncation at every bootstrap/frame boundary, including a syntactically valid prefix; no CMD/prune/status after mismatch; pre-existing old script + induced write/rename/open failure; mixed-version publication barriers; large excludes absent from `/proc/<watcher>/cmdline`; binary stdout/stderr + status 125/255 + controller SIGKILL. Update fake SSH's run-failure classifier: `/home/eturkes/Projects/offload/test/e2e/testdata/ssh:46` currently recognizes `*offload_watch*` in argv, which disappears when the script leaves argv; otherwise `/home/eturkes/Projects/offload/test/e2e/close_test.go:14–30` no longer injects the intended transport failure. This is fixture routing, not permission to weaken those checks.

## Evidence + coverage

- Reviewed whole `/home/eturkes/Projects/offload/.agent/contract.md`; named execution/SSH/probe/pool sources, execution tests + fake-SSH harness; relevant mirror/CLI transport callers + current CI. Execution sources/contract were unchanged between initial `8742aa9` and observed HEAD `8e9b3673f671e47412098f225910652f14c4abcc`; that interval changed only `.agent/deferred.md`.
- Read-only host query rc 0: systemd 260.5, OpenSSH 10.5p1, reachable user bus + slice placement. No remote writes outside the disposable `/tmp/offload-consult-2.*` inode model. Trap removed it; follow-up `find` rc 0 showed no matching directories, with `/tmp` itself as the positive control (expected one `tmp`, observed one).
- No source/config/test edits, staging, commits, test-suite run, persistent mux master, scope creation, process kills or host configuration changes. Review deliverable is the sole repo write. Real mux/cgroup lifecycle + timings and CI bus provisioning remain unconfirmed as explicitly noted above. One WebSearch query used; source text read verbatim via webtext. Source queries were targeted excerpts, not audits of whole upstream repositories.

## Sources

- https://raw.githubusercontent.com/systemd/systemd/v260/src/run/run.c — expansion default + scope admission, environment merge and exec.
- https://raw.githubusercontent.com/systemd/systemd/v260/src/core/scope.c — active-state mapping + cgroup-empty notification.
- https://www.freedesktop.org/software/systemd/man/latest/systemd-run.html — scope environment/argument behavior; local installed systemd-run(1), systemd.scope(5), systemctl(1) also inspected.
- https://raw.githubusercontent.com/openssh/openssh-portable/V_9_0_P1/ssh.c — persistent master detachment.
- https://raw.githubusercontent.com/openssh/openssh-portable/V_9_0_P1/mux.c — fd passing, control cleanup, socket bounds + stale sockets.
- https://raw.githubusercontent.com/openssh/openssh-portable/V_10_5_P1/ssh.c — deployed-version detachment cross-check.
- https://raw.githubusercontent.com/openssh/openssh-portable/V_10_5_P1/mux.c — deployed-version fd/stale-socket cross-check.
- https://raw.githubusercontent.com/coreutils/coreutils/v9.7/src/head.c — bounded reads + short-EOF success.

consultant-2-DONE-1
