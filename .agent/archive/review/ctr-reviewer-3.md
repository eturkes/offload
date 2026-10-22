# reviewer-3 — lens G: guarantee vs claim gaps

Target = `ef515f2..2c2744d` (implementation revision `2c2744dba7cc99ca5e1d360cadc940242995ec90`). Check set G1–G9 frozen before target read. Verdict: 9/9 adjudicated; 3 pass, 6 finding rows; 6 distinct major findings. All reproduction writes = reviewer-3 worktree; no remote access or installer execution. Primary `internal/slots/slots.go` changed during review; target evidence uses the pinned worktree, not that edit.

| id | check | verdict (pass / finding) | finding ids | evidence (file:line, command + output) |
|----|-------|---------------------------|-------------|----------------------------------------|
| G1 | K8 'each delivered signal reaches CMD's group once' across X1 paths (session TERM, scope TERM, leader TERM), terminal Ctrl-C, direct kill of podman | pass | — | `init.go:38–41,46–73`, `ctr.go:193–210,245–260`: socket-only forwarding + separate podman group; real PTY Ctrl-C logs exactly one INTERRUPT. Terminal completion is G-F6/G6, not a duplicate signal. Actual scope-wide signal injection not rerun; X1 paths read. |
| G2 | K8 liveness: offload ctr SIGKILLed at any phase (setup, build, lock wait, podman starting, CMD running, teardown) → no surviving container/process, no leaked lock | finding | G-F2 | K5 explicitly permits an orphan build to retain its lock until the build ends. K8's run-container guarantee fails with an unscheduled init: real-Podman G9 red. |
| G3 | K5 one build per tag per host (concurrent runs, SIGKILL during build, stale lock files, different storage confs) | pass | — | `ctr.go:101–131`, `slots.go:165–192`: tag lock + recheck + inherited fd; real concurrent miss builds exactly once (`TestReviewG3G5Controls`, pass). Lock files remain, flock ownership does not; storage config does not change the lock key. Build-death fd inheritance reviewed, not rerun with a real long build. |
| G4 | K4 equal trees → equal tags on every host (rsync-synced perms, umask, git checkout), any build-relevant change → new tag | finding | G-F3 | `tag.go:25–30`: symlinked context accepted by K3 hashes no entries. Containerfile edit leaves `localhost/offload-ctr:e3b0c44298fc1c14`; unit red rc 1. |
| G5 | K6 HOME persistence/sharing + ownership + worktree visibility (controller layout and the pool host's Y4 worktree layout) | pass | — | `ctr.go:165–190`: root-keyed HOME, keep-id, common-dir mount. `TestReviewG3G5Controls` passes: HOME write/read, host uid ownership, controller worktree common-dir HEAD, fake-pool Y4 gitfile + mounted common-dir HEAD. Existing K6 Home/RunArgs/RunningBinaryIdentity unit tests pass. |
| G6 | K7 status precedence + 125 cases (record vs podman rc, partial record, podman killed, init refused) | finding | G-F1, G-F6 | `ctr.go:235,304`: blocking socket defeats timeout; unit + real-Podman reds. Podman SIGKILL → parent remains running >3 s rather than 125. TTY stdin + separate podman pgid → completed CMD never returns its status; redirecting only podman's stdin to `/dev/null` passes. |
| G7 | K10 --check: equality claim (framing vs login-shell output, binary identity, down/busy hosts, plain vs git roots, Y3 layouts) | finding | G-F4 | `cli/ctr.go:229–240`: login banner contributes layers before header; unequal 1/0 image lists become `alpha: same (1 layers)`, rc 0. |
| G8 | tools/install 'same binary everywhere' (addresses parsing, aliases, failures mid-way, PATH on hosts) | finding | G-F5 | `tools/install:17`: valid literal-string/quoted-key TOML addresses omitted. Extractor-only test uses actual source; standard double-quote positive control passes. |
| G9 | README stop claims ('all processes stop at once and the container is removed'; TERM forwarding; 10 s) and the Y11/S5 barrier claim in K8 | finding | G-F2 | `.agent/contract.md:75`, `watch.bash:99–111`: stopped init → command wrote 248.633038 ms after heavy slot freed. No uninterruptible I/O. |

## Findings

### G-F1 — major — missing-record timeout cannot fire

- Target: `internal/ctr/ctr.go:235–240,262–274,303–304`.
- K7/K8 require `podman gone first → no record → 125`, then parent exit closes liveness. `Socketpair(AF_UNIX, SOCK_STREAM|SOCK_CLOEXEC, 0)` creates blocking fds; `os.NewFile` does not enable Go's poller for these. `SetReadDeadline` fails, ignored. `readRecord` then blocks while init still owns the peer; parent never closes liveness, so CMD can continue indefinitely.
- Anchor bytes: `syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM|syscall.SOCK_CLOEXEC, 0)`; `_ = f.SetReadDeadline(time.Now().Add(time.Second))`.
- Reds from `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-3`:
  - `TMPDIR="$PWD/.tmp" go test ./internal/ctr -run '^TestReviewG6ReadRecordDeadline$' -count=1 -v` → rc 1.
  - `TMPDIR="$PWD/.tmp" OFFLOAD_REQUIRE_PODMAN=1 go test ./test/e2e -run '^TestReviewG(6|9)' -count=1 -v` → rc 1, including actual Podman SIGKILL.

```text
readRecord still blocked after 1.5 s with a live peer and no record; K7 requires 125 after podman dies
--- FAIL: TestReviewG6ReadRecordDeadline (1.51s)
offload ctr still running 3 s after podman SIGKILL; expected 125 + container teardown
--- FAIL: TestReviewG6PodmanDeathReturns125 (7.02s)
```

### G-F2 — major — a delay is not a container teardown barrier

- Target: `.agent/contract.md:75`, `README.md:224`, `internal/ctr/init.go:46–59`, `internal/run/watch.bash:99–111` (unchanged watcher is relied on by the new K8 claim).
- K8 assumes EOF handling finishes before the watcher's next 100 ms emptiness check. No acknowledgement or container-lifetime lock connects those events. Stop PID 1 with SIGSTOP (same invoking uid; no root), leave CMD writing, SIGKILL the outer `offload run`: watcher kills offload ctr/podman, releases its slot, but PID 1 cannot process EOF and CMD continues. A delay in scheduling PID 1 is the same ordering hazard; neither mechanism enforces a bound.
- Exact claim anchor: `the container dies before X1's watcher can observe its session empty`.
- Red: `TestReviewG9StoppedInitKeepsSlotUntilLastWrite` in `test/e2e/reviewer3_test.go`; command above → rc 1. Test resumes PID 1 + force-removes only its isolated-storage container during cleanup.

```text
container wrote 248.633038ms after heavy slot freed while init was SIGSTOPed; no uninterruptible I/O involved
--- FAIL: TestReviewG9StoppedInitKeepsSlotUntilLastWrite (16.04s)
```

- Reproduced = local S5 last-write barrier. Remote Y11 uses the same watcher mechanism; remote failure is inferred, not exercised (brief prohibits real-host access). Normal unsuspended-init behavior does not establish the universal ordering claim.

### G-F3 — major — accepted symlinked build context hashes the empty tree

- Target: `internal/ctr/ctr.go:76–79`, `internal/ctr/tag.go:25–30`.
- K3 checks only that `Containerfile` itself is regular. `.devcontainer` may be a symlink to a directory. `filepath.WalkDir` does not descend a symlink root, and `path == dir` skips even that entry ⇒ SHA256(empty), shared across every such project/definition. Editing Containerfile reuses the old image; unrelated definitions share its tag.
- Anchor bytes: `p.Context = filepath.Dir(p.Definition)`; `if path == dir {`.
- Red: `TMPDIR="$PWD/.tmp" go test ./internal/ctr -run '^TestReviewG4' -count=1 -v` → rc 1 on 2c2744d. Positive control checks `Lstat(link/Containerfile).Mode().IsRegular()`.

```text
Containerfile content changed, but symlinked context kept tag localhost/offload-ctr:e3b0c44298fc1c14
--- FAIL: TestReviewG4SymlinkedContextChangesTag (0.00s)
```

### G-F4 — major — login output can turn unequal layer lists into green

- Target: `internal/cli/ctr.go:147–150,201–214,229–240`.
- Parser accepts every `layer ` line, even before `offload-check `. Thus a login banner can supply a missing layer. Probe: same installed offload bytes both sides; fake Podman reports one local layer and zero remote layers. Without banner → rc 1/different (positive control). Add `.bash_profile` output `layer sha256:local-layer` → rc 0/same. The comparison no longer means complete ordered lists are equal.
- Anchor bytes: `each line framed so a login shell's own output cannot pass for either`; `} else if v, found := strings.CutPrefix(sc.Text(), "layer "); found {`.
- Red: `TMPDIR="$PWD/.tmp" go test ./test/e2e -run '^TestReviewG7' -count=1 -v` → rc 1. Fake-host harness only; no network.

```text
unequal image lists became green solely from login banner: rc=0 stdout="alpha: same (1 layers)\n"
--- FAIL: TestReviewG7LoginBannerCannotSupplyLayers (4.12s)
```

### G-F5 — major — install silently omits hosts in valid pool files

- Target: `tools/install:14–18`; shipped promise `README.md:39`.
- P2's TOML parser accepts literal strings and quoted keys; installer accepts only unquoted `address` with a double-quoted value. E.g. `address = 'agent@alpha'` or `"address" = "agent@alpha"` yields no destination, while the script completes its loop successfully and has already replaced the controller binary. Hosts retain old bytes or lack offload entirely.
- Anchor bytes: `set -- $(sed -n`; README: `on each host of the pool file.`
- Red: `TMPDIR="$PWD/.tmp" go test ./internal/pool -run '^TestReviewG8' -count=1 -v` → rc 1. Test extracts/runs only the sed expression from target installer bytes; never invokes installer/build/copy/ssh. `pool.Load` accepts each fixture with exactly one expected host; standard double-quoted positive control passes.

```text
--- PASS: TestReviewG8InstallParsesEveryPoolAddress/double-quoted-positive-control (0.00s)
pool.Load has address "agent@alpha"; installer extraction rc=0 yields [] (want 1 address)
--- FAIL: TestReviewG8InstallParsesEveryPoolAddress/literal-string (0.01s)
--- FAIL: TestReviewG8InstallParsesEveryPoolAddress/quoted-key (0.01s)
```

### G-F6 — major — terminal stdin leaves ctr hanging after CMD exits

- Target: `internal/ctr/ctr.go:243–246,262` (K6 stdin + K7 status).
- `cmd.Stdin = os.Stdin` gives podman the controlling terminal, but `Setpgid: true` moves it outside the foreground process group. Actual PTY (`script`) → Ctrl-C reaches CMD once; CMD completes its 3 s trap program with status 5; podman + offload remain running beyond 6 s. Differential positive control changes only podman's stdin to `/dev/null` (same terminal/offload/CMD/signal) → returns 5. Thus this is not the trap program or the PTY harness waiting on its input pipe.
- Observed podman state = `RNl`, not stopped. Inferred mechanism = background-terminal input repeatedly triggers SIGTTIN while podman's signal proxy catches/forwards it; no syscall trace was taken. Reproduction establishes the terminal-stdin hang independently of that inference.
- Anchor bytes: `cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, stderr`; `cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}`.
- Red: `TMPDIR="$PWD/.tmp" go test ./test/e2e -run '^TestReviewG1TerminalCtrlC$' -count=1 -v` → rc 1; final combined rerun also red.

```text
terminal ctr did not finish after CMD exited; log="ready\nINTERRUPT\n" terminal="^C" children= 610918  610872  610918  610872 RNl  podman
--- FAIL: TestReviewG1TerminalCtrlC/terminal-stdin (10.12s)
--- PASS: TestReviewG1TerminalCtrlC/detached-stdin-positive-control (8.06s)
```

## Register

### R-G1 — fixture setup assumes TMPDIR is outside every git worktree

- Outside G1–G9; not a merge-blocking product finding. `test/e2e/ctr_test.go:37,46–51` creates its shared image fixture under `os.MkdirTemp("", …)` without making it a git root. With TMPDIR inside a worktree, `ctr.Find` correctly discovers that enclosing repository, not the intended plain fixture; image prebuild fails before the named check runs. Normal `/tmp` is the intended e2e placement; the project-local TMPDIR rule applies to mutago.
- Witness against the pinned target: from reviewer-3 worktree, `mkdir -p .tmp; TMPDIR="$PWD/.tmp" go test ./test/e2e -run '^TestK5ConcurrentFirstRunsBuildOnce$' -count=1 -v` → rc 1.

```text
ctr_test.go:199: offload ctr --layers (shared test image): exit status 2: offload: no image definition at /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-3/.devcontainer/Containerfile (a regular file, not a symlink)
--- FAIL: TestK5ConcurrentFirstRunsBuildOnce (0.01s)
```

- Acceptance if hardened: run the same named test with TMPDIR both inside and outside a git tree; fixture root remains its own project in both cases. Current review's test-only git-root fixture already avoids this precondition.

## Coverage / limits

- Target source read: ctr lifecycle/hash/init, CLI ctr/check, install, README, relevant mirror/slots/CLI/Nix integration diff, X1 watcher + fake-host/real-Podman test harness.
- Full gate not run (brief: targeted tests only). LSP attempted on `internal/ctr/init.go`; no Go server configured.
- Initial real-Podman attempts failed shared-fixture setup before exercising the behavior: placing TMPDIR inside a git worktree changes the plain fixture's root. Final test-only helper uses a git-root fixture + private vfs storage inside reviewer-3's worktree; target source unchanged.
- Native paths, rootless Podman 6.1.4, Go 1.27.3 tested locally. Actual nixbox, Podman minimum 4.9, real installer distribution not run.
- Final combined reproduction: from reviewer-3 worktree, `mkdir -p .tmp; TMPDIR="$PWD/.tmp" OFFLOAD_REQUIRE_PODMAN=1 go test ./internal/ctr ./internal/pool ./test/e2e -run '^TestReviewG' -count=1 -v` → rc 1. All six defects red again; G3/G5 controls pass; G8 standard-TOML + G6 detached-terminal-stdin positive controls pass. Latest barrier failure = a write 251.073872 ms after slot release.
- Existing controls: `TMPDIR="$PWD/.tmp" go test ./internal/ctr -run '^TestK(4|6|7)' -count=1 -v` → rc 0 (7 tests). New passing controls are review probes, not claimed mutation credit; mutation tool not run.
- Final validator: 9 table rows (rc 0); seed-marker count = 0 (rc 1 = no matches). Exact target anchor bytes rematched; implementation files remain unchanged.
- Red probes committed on `wt/reviewer-3`: `46a1634d5c0f70380ac2af6c9c8401e93facaa76`. Worktree clean. Temporary logs/storage removed; no process with the worktree's temporary-path marker remained (`pgrep` rc 1).
