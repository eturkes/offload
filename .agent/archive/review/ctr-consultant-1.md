# consultant-1 — `offload ctr` contract (K1–K11) review

**Direction = right; contract needs revision before code.** Keep the static PID-1 init + control socket. It addresses wrapper SIGKILL where an EXIT trap does not. Highest-priority issues: build-cache input equivalence, C3/ctr status precedence, and the unproved X1/Y11 container-death barrier. All 16 seeded rows adjudicated; findings below bind to the draft at `.agent/contract.md:66–78`.

| id | check | verdict (sound / defect / gap / risk) | evidence (file:line, command + output, URL) | recommended change |
|----|-------|---------------------------------------|---------------------------------------------|--------------------|
| D0 | direction: init-in-container on a control socket vs scope placement / host reaper | sound | `.agent/spec.md:22,33`; MAIN's reported socket/PID-1 probes: wrapper SIGKILL → container gone in 0.09–0.18 s. Independent FD3 probe passed; full init spike not rerun. | Keep this mechanism. Scope placement alone depends on the manager; host traps alone miss SIGKILL. Prove teardown completion, not merely eventual death. |
| K1 | usage + flag parsing | gap | `.agent/contract.md:68,73`: K6 calls `offload __ctr-init CMD…`, not `offload ctr __ctr-init CMD…`. | Reserve `__…` in top-level internal dispatch explicitly; state that a CMD after `ctr --` remains verbatim. Define help precedence over invalid switch/options. |
| K2 | switch semantics + native exec path | gap | `.agent/contract.md:69` requires native exec; C3 (`:11`) globally gives wrapper signals precedence. | Scope C3 to `run`; define ctr's native/image/build/check statuses separately. Say no project/pool/Podman reads on native path. Specify PATH/ENOEXEC behavior rather than broad shell equivalence. |
| K3 | root, definition, binary resolution | gap | `.agent/contract.md:70` lacks Y1's damaged-`.git` guard (`:43`); binary is resolved before a potentially long build. `internal/mirror/mirror.go:35–55` already implements the guard. | Reuse root discovery deliberately; reject malformed repos rather than silently changing roots. Specify symlinked definition policy, cwd containment, and running-binary identity through the eventual bind mount. |
| K4 | content-hash tag | defect | `.agent/contract.md:71` collapses all executable modes into one kind yet promises every x-bit change changes the tag. COPY preserves permissions [S3]. | Define canonical framed bytes; hash relevant permission bits (files + directories), or normalize the context actually built. Define unsupported file types. Extend quiescence through hash/build, or snapshot. |
| K5 | build on miss + lock + signals | gap | `.agent/contract.md:72`; `podman image exists --help`: 0 = exists, 1 = absent. Build wait covers HUP/INT/TERM, not parent SIGKILL or surviving build descendants. | Other image-exists failures →125, never a cache miss. Scope lock to image store/user, define runtime-dir fallback, and keep it held until build descendants stop even if ctr dies. Specify cancellation + log cleanup. |
| K6 | run argv, mounts, HOME, ownership, streaming | defect | `.agent/contract.md:73`; colon mount probe →125, `incorrect volume format`; FD3/raw bytes/HOME probes passed [E1–E3]. Prototype image ENV still names `/venv` and `/uv-cache` (`prototype/container/Containerfile:29–30`). | Encode all supported paths using a mount representation that supports delimiters, or explicitly reject them. Assert private PID namespace + PID1 init. State stdin behavior and image responsibility for directing writable state under HOME. |
| K7 | exit status + no-record 125 | defect | `.agent/contract.md:11,74–75`: TERM→143 vs trapped CMD's unchanged 0. Raw Podman command exit125 independently observed [E2]. | Retain status record disambiguation. Fix C3 scope; define valid/partial records, init setup errors, and record + Podman-failure precedence. Reap Podman/finish teardown before returning status. |
| K8 | signals once, reaping, teardown, liveness | gap | `.agent/contract.md:75`; `internal/run/watch.bash:99–101` sends overlapping TERM; PID1's exit is itself the namespace KILL operation. | Promise one forward per delivered host notification, not one per external send. Specify startup queueing, FD ownership, record-before-exit ordering, and behavior when Podman dies while socket stays open. See XC1. |
| K9 | `--layers` | sound | `.agent/contract.md:76`; `.agent/spec.md:33` explicitly selects ordered RootFS layer digests, not image IDs. | Keep. Make ordering and zero layers valid explicitly; failed build/inspect must not emit a successful-looking partial list. Signal statuses inherit K5. |
| K10 | `--check` | gap | `.agent/contract.md:77` gives remote aggregate rc but not local-build/inspect failure, cancellation, malformed output, or plain-root behavior. Y3 excludes some layouts. | Validate pool before build; define local failure vs aggregate mismatch vs signal statuses. Compare complete ordered digest vectors with framed binary hash/status data. Inherit Y1/Y3/Y7 limits; state offline-host indefinite waiting. |
| K11 | requirements | gap | `.agent/contract.md:78`; Podman 4.9.3 `--preserve-fds` excludes remote-client mode [S1]; `.agent/gate:41` already builds static. | Require a local Linux rootless runtime, usable subuid + subgid mappings, and matching Linux binary/image architecture. Pin/assert PID namespace/init assumptions. 4.9.3 runtime behavior remains unconfirmed. |
| XC1 | K8 under X1, with + without systemd scope | risk | `internal/run/watch.bash:7–18,99–110`; topology probe: caller `session-201.scope`, container separate `libpod-….scope/container` [E4]. Y11 (`.agent/contract.md:53`) promises no live predecessor at successor sync. | Gate code on a nested-kill probe: lock/slot cannot become reusable before container emptiness. EOF gives eventual death, not this barrier. Also test duplicate TERM from session + scope. |
| XC2 | conformance to spec Decisions | sound | `.agent/spec.md:33`: project definition, per-machine switch, independent build, LAN-only, RootFS comparison. `:19` keeps ignored state per host. | Preserve this direction. Keep the layer check's claim narrow: layer + binary equality is not cache-state, kernel, UID, image-config, or workload-result equality. Adopted image ENV must match the new HOME layout. |
| XC3 | flag availability: 4.9.3 / 5.8.8 / 6.1.4 | sound | Every K5/K6 flag exists in 4.9.3 docs [S1–S2]; controller + pool `run/build --help` confirm flags; pool SSH rc0; controller runtime subset passed [E1–E2]. | No flag-version blocker found. Do not equate docs/help with an executed 4.9.3 or nixbox lifecycle test. No evidence here requires crun specifically for `--preserve-fds`. |
| XC4 | races / persistent HOME / disk / worktrees | gap | `.agent/contract.md:5` promises quiescence during upload only; K3–K6 read paths again later; K6 shares HOME across runs/images; Y4 rewrites remote worktree admin paths. | Declare races' outcomes + cache concurrency responsibility; test binary replacement, image disappearance, context changes, disk/permission failures, linked worktrees, and malformed status. Keep the list bounded to these acceptance cases. |

## Detail

### K4 — key equivalence must match build inputs

Two independent problems:

1. `0100 → 0010` remains “executable file (any x bit)”; bytes/path/kind stay equal, but K4 promises an x-bit change produces a new tag. This is a contradiction within K4, without relying on Podman behavior.
2. `0644 → 0600` leaves K4's hash unchanged although COPY preserves file permissions [S3]. A cached image can retain the old mode; a clean host can build different layers under the same tag. Hash modes or build from a normalized context. A definition using `COPY --chmod` can eliminate that dependency, but the general contract does not require it.

Specify lengths/type markers around path + content records; simple concatenation is ambiguous. A symlinked Containerfile whose referent is outside the hashed context needs rejection or referent hashing. FIFO/socket/device entries need a defined error rather than an attempted regular-file read. These are input-domain rules, not new features.

The current quiescence assumption ends at upload. Hashing context A and then letting Podman consume context B incorrectly names B with A's tag. Recommended minimal contract: context remains quiescent through hashing + build; `--check` holds the same precondition across all hosts. A snapshot is the alternative if live edits must be supported.

### K3 / K6 — paths, binary, writable state

- Preserve the running binary's identity until launch. Resolving its install path before a build does not prevent an atomic install replacing that path during the build. Either bind the opened/running inode or detect replacement before the mount and define the residual race. A symlink-resolved spelling alone is insufficient.
- `-v` has its own colon grammar, independent of shell argv quoting. Y2 sanitizes a workdir *name*, not K6's absolute bind paths. E3 demonstrates the delimiter failure. Tests need colon/comma/space path cases for whichever mount representation is selected.
- The HOME plan itself works: E1 confirms uid/gid 1000 and passwd home `/.offload/home`. It does not route the prototype's `/venv` and `/uv-cache` there. Image adoption must change those ENV values or provide corresponding writable mounts. This is not a request for automatic per-tool mounts.
- HOME is shared per root across concurrent invocations and image versions. That is a reasonable cache policy, not an isolation guarantee. Projects must use concurrency-safe tool caches; mutable venvs/tool state may need their own image-specific path or lock. No cache corruption was reproduced.
- For linked worktrees, both the worktree root and its external common directory must remain visible at the paths the gitfile names. Validate this on the already-rewritten remote layout (Y4), not only the controller's layout.

### K5 / K7 / K8 — lifecycle and statuses

Recommended split: `run` keeps C3; native ctr has direct-exec semantics; image ctr returns CMD's recorded status; build/wait/check cancellation has explicit signal statuses. Then a TERM handler returning 0 is not a contract contradiction.

Specify that status is exactly one complete valid record from init. EOF without one, an invalid/partial record, init setup failure, or Podman launch failure →125 with an offload diagnostic even when Podman emits none. A record racing Podman's exit must be drained rather than lost. A record followed by a Podman cleanup failure needs an explicit precedence ruling; “Podman failing means no record” is not generally true.

The host must wait for Podman completion after receiving a record. Init must send status **before** its final PID1 exit; when children ignore TERM, that exit performs the namespace-wide KILL, so there can be no “record afterward.” A normal child-exit path and a host-liveness-EOF path need distinct transitions.

Keep only the intended socket endpoint in each process; CMD must not inherit init's control fd. Podman dying must wake the host even if conmon/init still holds its socket peer. Queue or define signals received between host startup and CMD process-group creation. Standard signals may coalesce; exact cardinality of external sends is not a sound promise.

Build cancellation needs equivalent process-tree/lock-lifetime reasoning: a dead parent releases its flock while a surviving build could still be using the tag. HUP/INT/TERM coverage alone does not establish K5's “one build at a time” under parent SIGKILL. Actual build-descendant behavior is unconfirmed; no builds were allowed in this review.

### XC1 — eventual death is not the existing non-overlap guarantee

X1 waits for the session + recorded scope to empty (`internal/run/watch.bash:7–18`). E4 observes Podman placing its container in another scope. After X1's KILL takes ctr + Podman out of its own session/scope, the in-container init may not yet have run its EOF handler. The watcher can therefore observe emptiness first and release the workdir lock. That would violate Y11's “a successor never syncs into a workdir with a live predecessor.” This ordering is an **inference**, not a reproduced overlap.

The decisive probe is not “eventually no container”: hold a TERM-ignoring descendant writing a marker, kill the outer run, race a successor's lock acquisition, and assert no predecessor write after acquisition. Run with and without the systemd scope, locally and on the pool. If the race exists, add a container-death barrier observable by the lock holder, or explicitly narrow the guarantee with the owner's ruling. A socket can remain the primary liveness mechanism.

Separately, X1 uses both `pkill -TERM -s` and `systemctl … kill --signal=TERM` (`:99–100`), and EOF also sends TERM directly to the leader (`:101`). Those overlap. Init ignoring raw Podman-proxy signals does not deduplicate two legitimate notifications received by the host ctr and forwarded over its socket. Either the contract allows standard-signal notification semantics, or the teardown needs a defined one-shot transition.

### K10 — check outcomes and meaning

Recommended explicit outcomes: usage/config error →2 before builds; local setup/build/inspect failure →125; remote failures/mismatches →1 with ordered host rows; signals →the declared signal status. Specify whether local failure prevents remote work. `offload run --host` waits indefinitely for an offline host by L3; disclose that behavior rather than implying a bounded availability check.

Use a machine-framed response for binary hash + complete layer list + status, not unstructured stdout splitting. Build output stays on stderr; zero layers is a valid vector; equal counts alone are not equality. Non-git roots inherit plain sync (Y7), not “git-exact” sync. Unsupported remote layouts inherit Y3 and must be reported as failures.

Layer equality is the owner's chosen check, so retain it. It deliberately does not compare image configuration; e.g. ENV is not a filesystem layer. Reproducibility/pinned-input obligations remain with the project's definition. Cache contents under HOME are also outside this check.

## Evidence + coverage

### E1 — HOME / uid / passwd (controller)

`podman run --rm --pull=never --userns=keep-id --log-driver=none -e HOME=/.offload/home --passwd-entry '$USERNAME:*:$UID:$GID:$NAME:/.offload/home:/bin/sh' docker.io/library/debian:trixie-slim …`

Output: `uid=1000(eturkes) gid=1000(eturkes) groups=1000(eturkes)`; `HOME=/.offload/home`; passwd field `…:/.offload/home:/bin/sh`; rc0. No host-writable mounts used.

### E2 — preserved fd / raw streams / command125 (controller)

Same existing image, `-i --preserve-fds=1 --entrypoint /bin/sh -w /tmp`, fd3 supplied by a here-string. Output: stderr `fd3=ctr-fd3` + `stderr-live`; stdout bytes `61 00 62 0d 0a`; Podman rc125 (CMD explicitly exits125), `od` rc0. This verifies the flags/streaming path, not the proposed init's status protocol.

### E3 — bind delimiter (controller)

`podman run --rm --pull=never --log-driver=none -v /home/eturkes/Projects/offload:/tmp/ctr:colon:ro docker.io/library/debian:trixie-slim true`

Output: `Error: …: incorrect volume format, should be [host-dir:]ctr-dir[:option]`; rc125. No container workload started.

### E4 — cgroup topology (controller)

Read caller `/proc/self/cgroup`; run existing image with rootless keep-id and `--cgroupns=host` to expose, rather than hide, the container's host cgroup path.

Output: caller `0::/user.slice/user-1000.slice/session-201.scope`; container `0::/user.slice/user-1000.slice/user@1000.service/user.slice/libpod-3fd0176eec586ba6f586962c7e05f1591afc47af56e2620cf4d5b756ee50da9f.scope/container`; container SID1; rc0. This establishes separate cgroup placement, not a measured Y11 violation.

### Sources (verbatim excerpts)

- **S1** https://docs.podman.io/en/v4.9.3/markdown/podman-run.1.html — `--preserve-fds`: “Pass down to the process N additional file descriptors (in addition to 0, 1, 2). The total FDs are 3+N.” / “This option is not available with the remote Podman client”. `--log-driver`: “Currently available options are k8s-file, journald, none and passthrough”. `--passwd-entry`: “The variables $USERNAME, $UID, $GID, $NAME, $HOME are automatically replaced with their value at runtime.” `--rm`: “Automatically remove the container and any anonymous unnamed volume associated with the container when it exits.” `--sig-proxy`: “Proxy received signals to the container process. SIGCHLD, SIGURG, SIGSTOP, and SIGKILL are not proxied.” Every K6 flag has a heading in this versioned page.
- **S2** https://docs.podman.io/en/v4.9.3/markdown/podman-build.1.html — `--timestamp`: “Set the create timestamp to seconds since epoch to allow for deterministic builds”; “All files committed to the layers of the image is created with the timestamp.” Also documents `--file/-f` and `--tag/-t`. This normalizes timestamps, not all context metadata.
- **S3** https://docs.docker.com/reference/dockerfile/ — COPY/ADD local sources: “If the source is a file, the file and its metadata are copied to the destination. File permissions are preserved.” Directory sources: “the contents of the directory are copied, including filesystem metadata.” Buildah/Podman mode behavior was not independently rebuilt here; S3 supplies the documented Containerfile-language semantics, and K4's executable-bit contradiction is independently visible in the draft.

All three sources opened in owned Neo tabs; `webtext` source extraction rc0. WebSearch calls =0/6. Local help checked; pool `ssh -o BatchMode=yes offload@192.168.40.240 'podman run --help'` and `… 'podman build --help'` each rc0. Reported installed versions (6.1.4 / 5.8.8) came from MAIN; independently checked flag surfaces, not version commands.

Not run: builds, pulls, image removal, 4.9.3 runtime, nixbox runtime containers, full gate, lifecycle spike rerun, X1/Y11 nested race reproduction, disk-full/concurrent-cache/worktree tests. Only the four permitted controller container probes above executed; containers used `--rm`. Repo + remote host unchanged; this report alone written.

Validator positive control: interim file contained 8 pending-marker rows; `grep -c` printed8, rc0. Final validator printed0, rc1 (grep's normal no-match status); table count =16 adjudicated rows, expected16. Closing status also showed MAIN's concurrent implementation files; none were changed by this reviewer.
