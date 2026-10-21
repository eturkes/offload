# consultant-1 verdicts — offload contract review

**Direction = amend first; retain Go + hybrid placement + controller flocks + Git-selected sync.** The main blockers are destructive-sync boundaries and supervision/lock lifetime, not the architecture. No implementation reviewed; no real SSH connection made. Contract/spec references below = the versions read for this review; MAIN owns amendments.

| id | scope | verdict | one-line finding |
|----|-------|---------|------------------|
| V1 | C1–C5 CLI + exit codes | defect | No-match is both 125 and 2; child status vs SSH/wrapper failure needs a separate completion result. |
| V2 | P1–P3 pool file | gap | `.`/`..` pass the name grammar; empty/invalid-config and duplicate-machine policies are absent. |
| V3 | R1–R2 probe + status | gap | “Parseable” allows zero CPUs / NaN load; numeric domains, probe bound and percentage scaling need predicates. |
| V4 | S1–S6 slots + heavy interop | defect | flock interop passes; client-only locks release before a SIGKILLed client's command finishes cleanup. |
| V5 | L1–L4 placement | defect | GPU absence cannot be inferred from down/unprobed hosts; 2-second local retry vs blocking probes is undefined. |
| V6 | Y1–Y3 roots + worktrees | gap | External `.git` files pass the worktree test but cannot be mirrored as specified; identity/error cases need policy. |
| V7 | Y4–Y5 upload + prune | defect | Literal prune can delete outside workdir; directory records cross ignored data; ignored ignore-files break R's semantics. |
| V8 | Y6–Y10 download + flags | defect | Deleted tracked paths fail rsync; `--update` replaces newer symlinks; `--get` scope and no-sync precedence unresolved. |
| V9 | X1–X4 watcher + EOF | defect | Initial-process-group kill ≠ session kill; startup/result/EOF lifecycle and SSH-loss guarantees remain unspecified. |
| V10 | missing binding predicates | gap | SSH BatchMode + accept-new absent; required binaries and nonzero-command download behavior need explicit coverage. |
| V11 | testability + ambiguity | gap | Fix the boundary matrix before tests: deadlines, result precedence, manifest semantics, path types, source mutation. |
| V12 | build / amend / rethink | risk | Amend first, especially V4/V7/V9. No evidence requires abandoning the overall design. |

## V1

- Predicate: `/home/eturkes/Projects/offload/.agent/contract.md:9` C3 = “no host matches” → 125; `:38` L4 = unknown host / no eligible host → 2.
- Scenario: `offload run --host absent -- true` has two specified statuses. A command returning 255 and an SSH transport failure also share one raw SSH exit code, but C3 requires different handling.
- Evidence: direct C3/L4 conflict. Installed OpenSSH manual, command `gzip -cd /usr/share/man/man1/ssh.1.gz | rg -n -A 7 '^\.Sh EXIT STATUS'` → lines 1578–1581: “exits with the exit status of the remote command or with 255 if an error occurred.” Local status-channel illustration = `python3 -I /home/eturkes/Projects/offload/.scratch/consult-1/spike_watch.py`, section D: command rc 255; transport-like shim rc 255. The shim is illustrative, not an actual SSH failure test.
- Amendment: static usage/config/unknown-host errors → 2; runtime transport/supervisor/sync failure → 125; valid completion record → child status unchanged, including 2/125/255. A supervisor-owned control channel or completion file separates completion from ordinary stdout/stderr and transport status. Declare precedence: offload interruption > infrastructure/sync failure > completed command status. Specify missing executable 127 / non-executable 126, or choose another explicit convention; these are command failures rather than wrapper failures.
- CLI edge: C1 should state parsing stops at the first command word; `offload run echo --pool` passes `--pool` to echo. This is an uncovered test boundary, not a demonstrated implementation bug.

## V2

- Predicates: contract `:15–17` P1–P3; pool lock construction `:29` S4.
- Scenario: name `..` matches `[A-Za-z0-9._-]+` but resolves `.../offload/hosts/../slot1` outside the per-host directory. Two entries with different names but the same destination get different controller locks yet address the same remote workdir.
- Evidence: local path-normalization command, `python3 -I -c 'import posixpath; r="/run/user/1000/offload/hosts"; print([(n,posixpath.normpath(r+"/"+n+"/slot1")) for n in ("nixbox",".","..")])'` → `nixbox/.../slot1`; `.` → `/run/user/1000/offload/hosts/slot1`; `..` → `/run/user/1000/offload/slot1`. The latter two are not distinct host namespaces. Duplicate destinations follow directly from S4/S5's name-keyed paths; not experimentally tested.
- Amendment: names are nonempty single components excluding `.` and `..`; address is nonempty. Specify malformed/unreadable config versus missing config; empty env values; empty host list; whether explicit `--local` reads the pool at all. Recommended: explicit `--local` ignores pool config; default run validates any present pool file before admission; empty pool behaves as local-only for default run and an eligibility error for pool-only.
- Admission identity: reject exact duplicate destinations; document distinct physical machines as a pool invariant, or key locks by a verified remote identity. SSH aliases can name the same box, so string deduplication alone is not a full guarantee. MAIN must choose the intended domain rather than quietly claim machine-wide admission across aliases/controllers.
- Compatibility gap: “any ssh destination” includes `ssh://user@host:port` (installed ssh manual lines 88–90), whereas raw `address:workdir` construction is not generally an rsync-over-ssh destination. Ensure address parsing/transport supports the promised domain; URI/IPv6 runtime behavior unconfirmed here.

## V3

- Predicates: contract `:21–22` R1–R2, `:11` C5.
- Scenario: four syntactically parseable lines `0\nNaN\n-1\n-1\n` satisfy the present wording but cannot represent a usable host. A percentage displayed as `load/nproc` without ×100 is off by 100 when rendered with `%`.
- Evidence: R1 specifies only parseability; R2 says “busy % (= load/nproc).” No implementation exists, so these are predicate gaps, not observed Go parser bugs.
- Amendment: CPUs integer ≥1; load finite real ≥0; available memory + GPU count nonnegative integers within representable bounds; reject overflow, NaN/Inf, incomplete or extra output. Score = load/CPUs; displayed percent = 100×score with chosen rounding. Nonzero SSH/probe exit is down even if stdout looks valid. Specify end-to-end probe timeout (connect + remote command + read) and cancel/reap behavior. Snapshot owner text only for actually held locks; a stale text file is not a holder. Output host order/rounding should be deterministic for fixtures.

## V4

- Predicates: contract `:26–31` S1–S6 and `:55–57` X1/X3; binding spec `/home/eturkes/Projects/offload/.agent/spec.md:14–15,20`; real heavy `/home/eturkes/.local/bin/heavy:4–13,56–77`.
- **Core interop sound.** Command: `python3 -I /home/eturkes/Projects/offload/.scratch/consult-1/spike_slots.py` (rc 0; full output `spike_slots.log`). It sets a private XDG_RUNTIME_DIR; never touches real heavy slots.
  - heavy holds slot → Python flock(LOCK_EX|LOCK_NB) blocked `True`.
  - Python holds slot → heavy status `slot 1: spike-owner`; heavy execution prints `all 1 slots busy, waiting`; child did not start.
  - Heavy owner file after normal release = 0 bytes; waiting heavy TERM rc 143.
  - HEAVY_SLOTS `''` + `0002` → rc 0; `0` + `9223372036854775808` → rc 2, `heavy: HEAVY_SLOTS must be a positive shell integer.`; `x` → rc 2, `heavy: HEAVY_SLOTS must be a positive integer.` Empty means default, not malformed.
- **Lifetime defect:** S3 excludes command inheritance; S6 says a dead client releases its locks; X1 permits ten seconds of live command cleanup. Therefore a second client can acquire the same slot/workdir and upload/prune while the predecessor still writes it. This is worse than temporary CPU over-admission: the workdir serialization invariant fails.
- Evidence: `python3 -I /home/eturkes/Projects/offload/.scratch/consult-1/spike_watch.py`, B → `slot before client SIGKILL: busy ; after: acquired; child alive: True`. This proves the descriptor-lifetime mechanism, not an implemented watcher race; the contract's ten-second grace supplies that overlap window.
- Amendment: a supervisor holds the slot + workdir locks until command teardown is complete (including abnormal client death); the command itself inherits none. If remote completion cannot be established, retain/quarantine the workdir lease rather than starting a new destructive sync. S6 must name the cleanup owner rather than promise immediate kernel release by the dead client. This changes the contract's lock ownership, not the binding direction's controller-lock location.
- Other required mechanical invariants: keep lock files persistent (unlinking a locked inode permits a second lock on its replacement); clear owner text **before** unlocking, only after owning that lock; distinguish lock contention from open/flock I/O errors. S3's quote format should match heavy `%q` or explicitly permit equivalent shell quoting.

## V5

- Predicates: contract `:35–38` L1–L4; P2 `:16` has no GPU capability field.
- Scenario: A successfully probes with 0 GPUs; B is down (or saturated and therefore not probed under L2) but has a GPU. `--gpu` must not decide “none configured with GPU” from A alone. Unknown capability ≠ absent capability. All hosts busy likewise ≠ no eligible host.
- Evidence: L2 probes only free-slot hosts; L4 describes configured GPU absence that config cannot express. Direct contradiction between the information available and the proposed terminal decision.
- Amendment: distinguish configured/compatible, reachable, capability-known, and currently admissible. Unknown/down/busy hosts remain potential candidates; exit for GPU impossibility only on an explicitly defined complete successful capability inventory. Otherwise wait, preserving 30-second down-host backoff. Unknown `--host` is a static rc-2 error; all eligible hosts down waits.
- Timing gap: L3 says retry every 2 seconds while R1's probes can block for an unspecified longer timeout. Specify whether 2 seconds is sleep-after-round or a local-admission polling bound. Recommend local retries remain responsive while host probes run independently; every wait/probe/acquire is signal-cancellable. Tie-breaking by pool-file order is sufficient; FIFO fairness/starvation freedom is **not** presently promised and need not be invented.

## V6

- Predicates: contract `:42–45` Y1–Y4.
- Scenario: an ordinary repository created with `git init --separate-git-dir=...` has `.git` as an external pointer, but git-dir equals common-dir. Y3 therefore permits it; copying the `.git` file does not copy its metadata to the pool. A submodule checkout has the same class of pointer problem.
- Evidence: `python3 -I /home/eturkes/Projects/offload/.scratch/consult-1/spike_sync.py`, D (rc 0): gitdir and commondir both end in `/sync-cases/separate-metadata`; `.git is file True`; contents point to that controller path. The exact path exists only within this scratch spike.
- Amendment: define supported Git repository layouts explicitly. Recommended first implementation: only self-contained normal repos remotely; linked worktrees, external gitdirs, submodule-root checkouts, and object alternates requiring outside objects fall back local / fail pool-only with reason unless deliberately supported. Detect before remote mutation. This narrows currently implicit support, so MAIN must reconcile it with Decisions rather than silently treating all Git errors as plain mode.
- Identity policy: canonicalize root + gitdir/common-dir consistently; distinguish positively detected non-Git cwd from corrupt/unreadable/failed Git inspection. A Git error must not silently invoke the destructive plain-directory mirror. Roots with spaces/newlines remain valid; serialize paths byte-safely. Hash collision or multiple controllers is not experimentally investigated; single-controller scope should be explicit (V2/V11).

## V7

- Predicates: contract `:45–46` Y4–Y5; binding Git-exact/ignored-path promise spec `:18`.
- Reproduction command for all cases: `python3 -I /home/eturkes/Projects/offload/.scratch/consult-1/spike_sync.py` (rc 0); full commands + fixture construction are in that script; output in `spike_sync.log`. All affected files are scratch files. These are counterexamples to naive implementations **permitted by the current contract**, not claims about code that does not yet exist.

### V7.a — R−U can resolve outside workdir

- Scenario: index tracks `dir/file`; locally delete it and replace `dir` with symlink `../remote-only-data`. Target absent locally, present only on the remote; U keeps the symlink but omits `dir/file`. After uploading U + .git, cached Git enumeration still lists `dir/file`.
- Observed F: `local U: [b'dir'] remote R: [b'dir', b'dir/file'] R-U: [b'dir/file']`; outside file before `True`; literal `rm -f -- root/dir/file` rc 0; outside file after `False`.
- Amendment: deletion is root-confined and no-follow at **every ancestor**, not merely at the leaf; abort on a non-directory/symlink ancestor or a type conflict, before deleting anything from that prune plan. Root path itself must be verified. A lexical `..` check or quoting alone does not satisfy this predicate. Bind operations to validated directory handles or an equivalently race-safe mechanism; how to implement the boundary is MAIN's ruling.

### V7.b — Git's path list is not always a list of leaf files

- Scenario: untracked embedded repo `nested/` contains `nested/cache/valuable.bin`, ignored by the parent `.gitignore`; tracked submodule gitlink `module` likewise denotes a directory.
- Observed A: Git list = `b'.gitignore\x00nested/\x00'`; `git check-ignore nested/cache/valuable.bin` rc 0; `rsync -ar --from0 --files-from=manifest source/ dest/` rc 0, `ignored_payload_copied True`, `nested_dotgit_copied True`. Observed E: gitlink list = `b'module\x00'`.
- Consequence: recursive upload of such entries crosses ignored runtime/cache files; recursively deleting the single directory record removes ignored descendants. Nonrecursive upload instead omits needed source contents. None satisfies all current promises.
- Amendment: define lstat path types and gitlinks before transfer; reject unsupported nested repos/gitlinks preflight or expand through a deliberate per-repository ignore-aware policy. Prune only individually selected leaf files/symlinks; remove directories only when empty. Never treat a directory record as permission for recursive deletion of ignored descendants. The reject option requires a clear local fallback / pool-only error rule.

### V7.c — “same ignore sources” requires more than U + .git + global excludes

- Scenario: active local `.gitignore` contains `.gitignore\ncache/\n`; the file ignores itself, so U omits it. Global excludes and `.git/info/exclude` are empty. Existing remote `cache/keep` must remain remote-only.
- Observed G: local U = `b''`; remote R after uploading U + .git = `b'cache/keep\x00'`; local `git check-ignore cache/keep` rc 0; remote check rc 1. R−U now targets the cache because the active ignore file never crossed.
- Amendment: transmit an explicit complete ignore-source snapshot independently of whether each source is in U, or compute pruning against an equivalent controller-defined ignore evaluation. Include active per-directory `.gitignore`, `.git/info/exclude`, and resolved global excludes. Missing/inaccessible ignore sources, local Git enumeration errors, lstat errors other than ENOENT, manifest framing/count mismatch, or unsupported path types → fail before pruning. Count alone verifies record cardinality, not correct-root/complete-set semantics.
- Failure boundary: say “no worktree prune deletions” rather than “nothing deleted,” because `.git` mirroring can already have deleted metadata before the prune validation. Define nontransactional upload failure honestly; do not imply rollback that rsync does not provide.

## V8

- Predicates: contract `:47–51` Y6–Y10; spec `:16,18`.
- Evidence command: `python3 -I /home/eturkes/Projects/offload/.scratch/consult-1/spike_sync.py` (rc 0).
  - B: tracked `gone.txt` removed remotely remains in `git ls-files -co` as `b'gone.txt\x00'`; passing that D to rsync returns **23**, `link_stat .../gone.txt failed: No such file or directory (2)`. An ordinary command deleting a tracked file is therefore misreported as an offload failure.
  - C: destination symlink has newer lstat mtime; `rsync -a --update source/ dest/` rc 0 nevertheless changes `local-target` to `remote-target`. Thus the broad parenthetical “a file newer locally stays” is false for symlinks.
- Amendment: filter D to existing lstat entries just like U; distinguish absence from inspection failure. `--update` is an mtime-based policy for regular files, not general concurrent-edit protection; either state that exact limit or implement explicit local-type/mtime conflict refusal for symlinks and type changes. “Ignored remote files never return” must read **ignored untracked files**; tracked-but-ignored files belong to D by the binding tracked-file rule.
- `--get` scope gap: run from `repo/subdir`, request `--get out.bin`, command writes `repo/subdir/out.bin`. Y8 does not say root-relative or invocation-cwd-relative. Absolute paths, `../` escapes, `.git`, and symlink ancestors can also bypass the nominal worktree boundary unless constrained. Recommended: paths are invocation-cwd-relative; normalized source + destination stay within their respective sync roots; `.git` never downloads, including explicit get; validate before command start. Preserve file versus directory-contents/trailing-slash semantics explicitly.
- `--no-sync --get PATH`: Y8 says fetch after download; Y9 says no download. Prototype always performs explicit get (`offload.py:126–127`) even with automatic sync disabled. Recommended amendment: no-sync skips upload/prune/automatic download; explicit get still runs. MAIN should confirm this is the intended reading of Decisions rather than silently choosing the opposite.
- Y7 is deliberately rsync `.gitignore` dir-merge semantics, **not** Git's complete ignore language. Preserve that stated plain-mode distinction; a plain-mode symlink/ancestor boundary still needs the V7 safety predicate. Missing explicit-get source is a transfer failure; interruption skips both automatic and explicit downloads (X3), with precedence fixed under V1.

## V9

- Predicates: contract `:55–58` X1–X4; binding spec `:21`; prototype `/home/eturkes/Projects/offload/prototype/offload/offload.py:23–26` illustrates, but does not define, the new watcher.
- Command: `python3 -I /home/eturkes/Projects/offload/.scratch/consult-1/spike_watch.py` (rc 0; `spike_watch.log`).
  - A: leader `{pid: 2011418, pgid: 2011418, sid: 2011418}`; child `{pid: 2011419, pgid: 2011419, sid: 2011418}`. Child only calls setpgid, **not** setsid. Leader-group TERM kills leader (rc −15); child remains alive in original session. Leader-group KILL → ESRCH; child still alive. Therefore “kills session's process group” does not entail “ends the command's session.”
  - C: duplicate pipe write fd held in another process → after creator closes it, EOF `False`; after leaked writer exits, EOF `True`. This is a required descriptor-ownership invariant, not proof that the future implementation leaks descriptors.
  - E: subprocess signal result −15 maps to shell-compatible status 143.
- Amendment / decision: either contain and terminate **all** processes the promise covers (a session can contain several process groups), or explicitly limit offload eligibility/guarantee to commands staying in the initial group. The latter is a change to binding direction and needs MAIN's ruling. Merely sending TERM then KILL to `-leaderPID` cannot prove the present promise.
- Supervisor state machine must cover: EOF before spawn/session creation; setsid success/actual leader PID acknowledgment before kill target is used; command starting/exiting immediately; client signal during upload/probe/start/download; TERM-resistant descendants; normal child exit with live descendants; watcher disarm/reaping; command result independent of SSH rc; repeated signals; writer fds present only in the intended controller. Keep watchdog outside the command's killed group. Avoid a cleanup action aimed at a stale/reused PGID after normal completion. These are unverified implementation risks; no implementation exists.
- Remote transport limitation: TCP/SSH does not guarantee prompt remote EOF during a network partition. Client-side EOF cannot cross a broken link; client-only timeouts are not a remote-death detector. A bounded remote-stop guarantee needs a remote lease/heartbeat or a stated transport assumption; specify the bound and hold workdir locks accordingly (V4). No real SSH, sshd shutdown, network partition, remote shell, or remote setsid behavior was tested in this review; local pipes/processes settle only the Linux primitives.
- Required integration tests before shipping: fake SSH wrapper plus real-host run owned by MAIN; exact byte stdout/stderr streaming; command rc 0/2/125/255; launch failure; client INT/TERM/KILL at spawn/run/cleanup; child ignoring TERM; child changing group; normal completion leaves no helper; disconnect/timeout handling. The scope must match the amended guarantee, not silently exclude a failing case.

## V10

- Missing predicate: binding spec `/home/eturkes/Projects/offload/.agent/spec.md:22` requires SSH BatchMode + accept-new. Contract R1/X2 omit both. Prototype `offload.py:19` supplies them with a connection timeout.
- Amendment: every SSH consumer (probe, execution, upload, download, explicit get) shares BatchMode=yes, StrictHostKeyChecking=accept-new, and the chosen bounded connection settings; no interactive prompts. Validate capability errors before destructive work where possible.
- Dependency contract: local ssh/rsync/git plus supported OS/flock interface; remote bash/rsync/git/coreutils and whichever session/watch primitives the chosen supervisor actually uses. `setsid` is a util-linux command, not GNU coreutils; X1 should not imply coreutils alone supplies it. Static Go binary does not remove those external requirements. Local and remote command-not-found must be distinguished from missing offload infrastructure.
- Nonzero command download is present in Y6 (“any rc”) and should remain explicitly tested; do not accidentally make a Go `Run()` error skip outputs. Idle-task management is intentionally host configuration only (spec `:23`), not a missing offload-code predicate. No contract expansion into trading/deployment is warranted.

## V11

- Scope: mechanically decidable contract after the ambiguities above are removed. No need for open-ended prose claims like “safe” or “never crashes” without enumerating the accepted input classes.
- Unfixed test dimensions: probe/cleanup deadline and retry-clock meaning; placement tie order; byte/path framing including newline and non-UTF8 Git paths; directory/gitlink/symlink policy; child-vs-infrastructure exit precedence; explicit-get root and no-sync interaction; stale owner behavior; invalid versus absent config; platform assumptions; lock identity when same box appears under aliases.
- Source mutation: mirrored `.git` + worktree files is not an atomic Git/worktree snapshot. Excluding `*.lock` avoids shipping lock artifacts, not a concurrent Git operation's mixed state. The binding direction never promises a snapshot. Explicitly require a quiescent source during upload or detect/retry/refuse source changes; do not grade the implementation against an unowned snapshot guarantee. Actual concurrent-Git experiment **not run**.
- Suggested fixed acceptance set before tests/code: valid/malformed config table; bidirectional real-heavy flock spike; local-first/pool-fallback/wait/cancel placement table; sync matrix {tracked ignored, untracked ignored, deletion, global/per-dir/self-ignored ignore source, nested repo/gitlink, symlink ancestor/type change, corrupt/short manifest, upload failure}; run matrix {status transport distinction, stdin EOF, signal timing, TERM grace, group split, helper cleanup}; output transfer on success/nonzero plus interrupted no-transfer. Each new rejection/fallback must be a contract clause, not a test author's invention.

## V12

- Ruling: **amend first**. The selected stack/admission trigger and conventional flock interoperability fit the owner direction. No evidence favors a rewrite of the product concept or a different language.
- Before implementation: MAIN adjudicates V4 lock lifetime, V7 confined leaf-only prune + complete ignore-source handling, V9 session/process-group scope + remote-loss bound; then folds the smaller V1/V2/V3/V5/V6/V8/V10 clarifications into the contract. Decide unsupported Git layouts explicitly rather than ship partial fidelity under “git-exact.”
- Next step: turn the confirmed scratch counterexamples into the fixed acceptance matrix for the appropriate kernel units; tests must fail a deliberately naive implementation, then pass the actual one. Use real SSH only in MAIN's owned-host integration step. Re-review amendments against these findings once; do not restart an open-ended architecture review.

## Verification + limits

- Seed validator positive control: 12 seeded table rows; `grep -c '^| V.*| unknown' verdicts.md` → `12`, rc 0. Final: `grep -c '^| V[0-9]' /home/eturkes/Projects/offload/.scratch/consult-1/verdicts.md` → `12`, rc 0; `grep -c '^| V.*| unknown' /home/eturkes/Projects/offload/.scratch/consult-1/verdicts.md` → `0`, rc 1 (expected no matches). Structural validator checks exactly V1–V12, enum verdicts, and a detail heading for every non-sound row → `PASS 12`, rc 0. Cleanup positive control: current shell PID count = 1; `pgrep -af '[s]pike_watch.py'` → no matches, rc 1.
- Passed: isolated sync spike (7 cases), bidirectional real-heavy/private-flock spike + five HEAVY_SLOTS cases, isolated process-group/descriptor/lifetime/status spike. Logs + scripts are in this directory; all process children were explicitly terminated/reaped where owned.
- Not run: real SSH/pool, NixOS remote shell/session behavior, network partitions, actual Go implementation/tests (none reviewed), URI/IPv6 rsync transport, concurrent Git mutations, production/global gates. Claims in those areas are marked gaps/risks or deductions from the predicates, not verified implementation outcomes.
- Writes: this directory alone. Root repository, contract, spec, actual heavy, and real heavy runtime untouched. Scratch Git repositories exist only as experiment inputs; no root-repository Git mutation performed.
