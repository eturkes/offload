# reviewer-6 — code review @ 41fd3c0

Result: 15/15 rows adjudicated; 10 pass, 5 finding rows; 4 unique findings (2 major behavior, 1 minor behavior, 1 minor Decision-text). Red tests committed `779fa9a` + `cb3f5d7` + `d4f708c`; production target unchanged.

Scope = fixed SF1–SF15 below; subject 41fd3c016caf6ed2e7cbace6f14e7f316d57ac32 in `.scratch/worktrees/reviewer-6`. Decisions outrank contract predicates. Security posture carve-outs = MAIN, not adjudicated here. Prior accepted rulings in all three `ledger-*.md` read; any reversal requires new evidence.

Validator seed: expected 15 unknown rows → 15 (rc 0); pass-only control expected 0 → 0 (grep rc 1). Evidence paths below are subject-root-relative unless stated otherwise. No full gate or real-pool run permitted; targeted tests + read-only live queries alone.

## Verdict table

| row | check | verdict | finding ids | evidence (file:line, command) |
|-----|-------|---------|-------------|-------------------------------|
| SF1 | m1 cli: trigger + flags, placement order (pool → local heavy → wait both, one waiting message), load ranking, `--host/--gpu/--no-sync` ⇒ `--pool`, stderr placement | finding | R6-F1 | Decision `User/Trigger`: hybrid pool-first **finding** (local lookup rejects host-only command before placement); pool→heavy→wait, one waiting message **pass** `place.go:59–88`; load ranking **pass** `place.go:146`; implied flags **pass** `cli.go:137–141`; placement stderr **pass** `cli.go:268,285`. Targeted C2/C5/L1/L2/L3/L4 e2e rc 0; red `TestSF1HybridRunsPoolOnlyCommand` rc 1. |
| SF2 | m2 run: remote run sequence, rc passthrough, kill-on-signal incl. SIGKILL (watcher), per-(host, workdir) serialization, local runs never lock | pass | — | Decisions `Remote run`: upload→login-shell CMD→download/get **pass** `cli.go:301–337`, `run/remote.bash:76–91`; command rc including 125/255 **pass** `run.go:109–125`; `Killing offload`: watcher pipe EOF→session TERM/KILL **pass** `watch.bash:65–122`, `run.go:130–163`; `Remote runs…serialize`: controller host/workdir lock **pass** `place.go:116–126`, remote phase lock `remote.bash:39–41`; local bypasses workdir lock **pass** `place.go:74`. Targeted X1ControllerDeath/X3Signals/X4Streaming/X6Status/S4Serialization e2e rc 0. |
| SF3 | m3 mirror: sync root, workdir naming, git-exact up/down sets, deletion pass, download `--update` no deletes, `.git` never returns, non-git mirror, linked worktree layout | pass | — | Decisions `Sync root`, `Git-exact sync`, `linked git worktree`: root/subdir **pass** `mirror.go:44–94`; basename+SHA1 workdir **pass** `:261–270`; U + tracked/ignored rules **pass** `:178–219,341–367`; prune with local global excludes **pass** `run/remote.bash:28–73`; D/--update/no deletes/.git exclusion **pass** `remote.bash:82–87`, `mirror.go:380–398`; plain mirror **pass** `:369–378`; linked common/admin layout **pass** `:358–365`, `remote.bash:45–51`. Targeted sync e2e batch had timeouts in linked-worktree + download; both individual reruns passed (11.721 s, 24.014 s package times); other selected cases passed. See verification limits. |
| SF4 | m4 slots: heavy interop (paths, flock, owner line, emptied at exit), pool slot locks, admission vs loadavg | finding | R6-F3 | Decisions `heavy interop`, `Admission`: runtime/fallback path + HEAVY_SLOTS/default2 **pass** `slots.go:37–70`; ascending exclusive nonblocking flock **pass** `:121–137,189–208`; owner pid/time + emptied at exit **pass** `:133,147–155`; `%q` wording **finding** against Decision (documented POSIX shell quoting, not Bash printf %q), `cli.go:406`, `remote/ssh.go:81–95`; host lock path **pass** `slots.go:74–77`; load never admits **pass** `place.go:106–151`. Seven targeted heavy differential/interoperability/fd/death tests rc 0. |
| SF5 | m5 pool: TOML `[[host]]` (`name`, `address`, `slots` default 2), `OFFLOAD_CONFIG` | pass | — | Decisions `Pool`, `Admission` + Artifacts pool override: [[host]] fields **pass** `pool.go:21–38`; absent slots→2, explicit positive counts **pass** `:18,96–117`; alias/address syntax **pass** `:49,108–109`; OFFLOAD_CONFIG first **pass** `:53–61`. `go test -count=1 ./internal/pool/` rc 0. |
| SF6 | m6 probe: load = rank only (1-min / nproc), GPU inventory | pass | — | Decisions `Admission`, `User/Trigger --gpu`: one-minute load + nproc **pass** `probe.go:17–18`; score = load/nproc **pass** `:37`; load not an admission limit **pass** `cli/place.go:106–151` (slot/workdir first, finite load merely compared); NVIDIA inventory/filter **pass** `probe.go:18,56–58`, `place.go:143–145`. Probe Parse/ScriptOnThisMachine/boundary/counter tests rc 0; L2/L4 controls rc 0. |
| SF7 | m7 remote: BatchMode + accept-new, ssh_config alias addresses | pass | — | Decision `Pool/SSH`: BatchMode=yes + accept-new **pass** `remote/ssh.go:21–23`; alias/address passed unchanged to ssh/rsync **pass** `:38–46`, `pool/pool.go:49` (ports/IPv6 delegated to alias; prior accepted address narrowing preserved). TestTargetCarriesOptions/JoinRoundTripsThroughBash/Multiplex rc 0. Security assessment beyond the named transport settings = MAIN. |
| SF8 | m8 ctr: K-predicates, switch `OFFLOAD_CTR`, `.devcontainer/Containerfile`, tag = content hash, host builds itself, `--check` = `RootFS.Layers` equality, LAN lanes only | finding | R6-F2 | Decision `Gate image placement`→K1–K11: CLI flags/switch/definition **pass** `cli/ctr.go:34–85`, `ctr/ctr.go:57–91`; K2/K7 Bash lookup **finding**, shared R6-F2; K4 content identity **pass** `ctr/tag.go:21–83`; K5 per-host build + zero timestamp **pass** `ctr.go:112–146`; K6 mounts/env/HOME/static-init **pass** `:175–205`, tools/install:10; K7 status + K8 liveness/teardown **pass** `ctr/init.go:28–130`, `ctr.go:217–317`; K9 layers **pass** `:165–172`; K10 ordered layer + binary equality **pass** `cli/ctr.go:148–229`; K11 rootless runtime documented + machines provisioned (SF14). LAN placement only, no registry distribution **pass**, CI native (SF13). Targeted tag/ensure/argv/layers unit tests rc 0. Real-podman rerun not performed. |
| SF9 | m9 ledger: every `run` logged (start once a slot is held + end), fields per `Decisions`, log failure never touches the run | pass | — | Decision `Run log`: per-machine path **pass** `ledger.go:108–132`; start only once placed + end centralized on all parsed-run exits **pass** `cli.go:164–180,267,281`, `ledger.go:188–197,222–239`; placement/wait/phase/cmd+offload rc/signal/messages **pass** `ledger.go:31–78,169–180,199–239`; failed log leaves run/status intact **pass** `:245–258`; SIGKILL caveat = J5, not a promise of impossible end writes. Targeted local/pool/unplaced/nosync/failure e2e rc 0. |
| SF10 | m10 shell: lookup = bash's (126/127) | finding | R6-F2 | Decision `Gate image placement`→K2/K7, X2 local lookup; native command identity **finding** for empty PATH: `shell/look.go:28` uses `filepath.SplitList("")`→no candidates, but bash searches cwd. Other declared path/slash/nonexec cases **pass** by code `:25–69` + existing TestLookMatchesBash read. Red `TestSF10EmptyPathMatchesBash` rc 1; bash oracle rc 0 vs Look 127. |
| SF11 | m11 e2e: harness models the pool hosts `Decisions` describe (git + coreutils, ssh options) | pass | — | Decisions `Pool/SSH`, `Git-exact sync`, `Remote run`: fake hosts have separate HOME/workdirs + login-shell execution **pass** `test/e2e/testdata/ssh:39–41,67–72`; real git/rsync/GNU tools execute the actual embedded remote scripts **pass**, sync controls exercise them; BatchMode + accept-new mandatory on every fake call **pass** `:23–31`; -n consumes /dev/null **pass** `:9,71`. `TestHarnessFakeSSH` rc 0. Limits: dirs on this host ≠ a separate OS/network; kernel/systemd/PATH shared; SSH alias resolution + real SSH transport not emulated. |
| SF12 | m12 tools: install = one static build → controller + every pool host, sha256 per side; heavy separate script deployed alongside | pass | — | Decisions `Stack`, `heavy interop`, `Gate image placement` + Artifacts install: one CGO_ENABLED=0 build **pass** `tools/install:10`; same build + separate heavy copied locally and to all configured hosts by default **pass** `:11–35`, `tools/pooladdrs/main.go:20–29`; atomic rename + each side's sha256 output **pass** `install:15–18,28–35` (comparison is operator responsibility, not automated claim). `go test -count=1 ./tools/pooladdrs/` rc 0. Install not executed (shared machine mutation outside review). |
| SF13 | m13 gate + CI: `Gate compute` + `CI gate` rulings, Gate CI on Ubuntu 26.04, pins weekly, Dependabot scope | pass | — | Decision `Gate compute`: offload exception + local-pool full gate **pass** `.claude/rules/gate.md:7`; second-environment Gate CI/native LAN-image split **pass** `.github/workflows/gate.yml:5–11,24,47–53` (main push + dispatch, ubuntu-26.04, direct .agent/gate); chain/tool versions + static build **pass** `.agent/gate:31–59`, workflow `:36–39`; Pins weekly **pass** `pins.yml:4–6`, tool/go/runner checks `.agent/pins:16–73`; Dependabot exactly gomod/actions weekly **pass** `.github/dependabot.yml:4–20`; CodeQL separate Go/manual build **pass** `codeql.yml:45–56`. General CI fan-out owner policy = n/a to this repo by explicit exception; current remote CI state/full gate not run here (MAIN). |
| SF14 | m14 machines: idle task slice, accounts, LAN-only nftables, key-only ssh, LUKS/TPM/Secure Boot, compute posture, carried tools, security posture, Wi-Fi secret path, `check` at every deploy | finding | R6-F4 | Decisions `Idle task`: user sibling slice CPUWeight=idle + MemoryHigh20G **pass** `idle.nix:24–30`, live cpu.idle=1/memory.high=21474836480; trading guards/disabled units/Luna hold **pass**, read-only facts below. `Compute posture`: governor/profile/PL1=22W/thermald off/earlyoom, THP+sysctls+zram+kyber+NOFILE+/tmp, lid+80% Custom battery **pass** `configuration.nix:84–135`, `performance.nix:5–48`, live values match; nix-ld/managed uv/cert **pass** `configuration.nix:153–193`, live bundle cmp rc0. `Carried tools`: envfs, managers/auto+stable, node26/system24, clang/shellcheck/swipl10.0.2/uv/bwrap, Chromium libs, git identity, Noto/UTC **pass** `configuration.nix:197–313`, live versions/defaults below. `Pool run environment`: podman+subids/OFFLOAD_CTR **pass** `:251–256`, live 5.8.8/subids/1. `check every deploy` **finding against deploy**: only switch invokes it (`deploy:10`; README:14); fake-deploy `TestSF14DeployChecksNonSwitch` rc1 (test/boot/build skip check; switch control passes). Accounts/sudo, LAN nftables/sshd, disk/TPM/Secure Boot, Wi-Fi secrets, remaining security posture = **MAIN**. |
| SF15 | `Decisions` binding no code module (trading + Codexify, instruction tiers, nixbox role, gate image placement, pool run environment, run memory) vs machines config, README, rules: each fits or is ruled n/a by name | pass | — | `Trading/Codexify`: pool trading path + home workspace/catalogue `.`/alias trading/never/AGENTS routing **pass** `machines/nixbox/README.md:53–58`, live confirmed; controller trading path absent (positive control offload path rc0, trading test-e rc1); key-copy/security facts = **MAIN**, historical confirmation/checksum event = **n/a** (not reconstructable by code review). `Instruction tiers` **pass** `.claude/rules/portability.md:3` carries explicit offload exception; other projects' guidance/owner trimming = **n/a**, external owner work. `nixbox role` **pass** as policy compatibility (persistent workdirs, idle trading, no CI/size/time admission rule); capacity numbers + past latencies = **n/a** historical observations, not rebenchmarked. `Gate image placement` **pass** README image workflow + SF8 mechanism (R6-F2 tracked there); per-project adoption/pinned inputs/Aeon/other repos' CI choices = **n/a**, consumer-owned. `Pool run environment` **pass** configuration.nix:251–256 + controller/pool podman/switch checks, no two-sided Nix project pin. `Run memory` **pass** README:170–181 + no memory limit in run/ctr, earlyoom active; agent-chosen limits + controller oomd policy external (not run here). |

## Findings

### R6-F1 — hybrid pool-first rejects a command available only on the pool

Severity: **major**. Target: `internal/cli/cli.go:240–243` @ 41fd3c0 (`if !o.pool {` → `checkLocalCommand` before `p.place`). Decision: `User/Trigger` promises an available pool slot before any local fallback; `Remote run` resolves CMD in the host's login shell. A host-only tool is valid on an available host, yet the default hybrid command returns 127 from the controller without contacting the host. Explicit `--pool` succeeds with the same CMD + pool. Prior accepted R7 in `ledger-pool-run-environment.md` establishes 126/127 for local execution, not this extra eligibility condition on remote execution; no accepted ruling reversed.

Red witness: `test/e2e/specfit_review_test.go:11` `TestSF1HybridRunsPoolOnlyCommand`; commit `779fa9a` on `wt/reviewer-6` (test alone; production = subject). Fake host login profile adds a tool absent on the controller. Explicit `--pool` is the positive control (rc 0 + exact `pool-only\n`), then hybrid must produce that output from alpha.

Command (worktree cwd): `go test -count=1 -run '^TestSF1HybridRunsPoolOnlyCommand$' ./test/e2e/`.

```text
--- FAIL: TestSF1HybridRunsPoolOnlyCommand (0.73s)
    specfit_review_test.go:27: exit = 127, want 0; stdout="" stderr="offload: reviewer6-pool-only: command not found\n"
FAIL
FAIL github.com/eturkes/offload/test/e2e 3.554s
red-test rc=1
```

Acceptance: local command lookup gates only a local execution path; the red hybrid test passes while local 126/127 + pre-slot validation remain correct.

### R6-F2 — empty PATH is not Bash-compatible

Severity: **major**. Target: `internal/shell/look.go:28` @ 41fd3c0, `for _, dir := range filepath.SplitList(os.Getenv("PATH")) {`. SF10 + K2/K7/X2 require Bash lookup. An explicitly empty PATH means cwd to Bash; Go's `filepath.SplitList("")` returns zero entries, so Look incorrectly returns 127 even when cwd contains an executable. This affects native `offload ctr` and image-init lookup, not only the helper's internal result. Prior accepted `ledger-pool-run-environment.md` R7 requires Bash-compatible 126/127 and has no exception for empty PATH.

Red witness: `internal/shell/specfit_review_test.go:10` `TestSF10EmptyPathMatchesBash`, commit `cb3f5d7` on `wt/reviewer-6`. Creates an executable in cwd, sets PATH to `""`, runs Bash through its previously resolved absolute path as the positive control (rc 0), then calls Look with the same env/cwd.

Command (worktree cwd): `go test -count=1 -run '^TestSF10EmptyPathMatchesBash$' ./internal/shell/`.

```text
--- FAIL: TestSF10EmptyPathMatchesBash (0.02s)
    specfit_review_test.go:27: empty PATH: bash rc=0, Look="" rc=127 err=command not found
FAIL
FAIL github.com/eturkes/offload/internal/shell 0.026s
SF10 red rc=1
```

Acceptance: empty PATH searches cwd; native + init callers keep their Bash-compatible lookup/126/127 behavior.

### R6-F3 — Decision names Bash `%q`; documented owner text uses POSIX shell quoting

Severity: **minor**. Text target: `.agent/spec.md:14` @ 41fd3c0, exact anchor `<cmd %q>`. The Decision describes the owner suffix as Bash's `%q` rendering, but `.agent/contract.md` S3 says `shell-quoted`, `internal/cli/cli.go:406` deliberately routes it through `remote.Join`, and `internal/remote/ssh.go:81–85` implements POSIX single-quote quoting. Space-bearing words become `'a b'`, not Bash `%q`'s `a\ b`; newline-bearing words can span physical lines, unlike `%q`. The held-slot/status interoperability passes; this finding is the stale exact-format wording, not a demand to redesign locking. Owner ruling required by the brief before changing a Decision.

Text proof (all matched, rc 0):

```sh
rg -Fn '<cmd %q>' .agent/spec.md
# 14:- `heavy` interop: … owner line `<pid> since HH:MM:SS: <cmd %q>` …
rg -Fn 'func ownerText(argv []string) string { return " " + remote.Join(argv) }' internal/cli/cli.go
# 406:func ownerText(argv []string) string { return " " + remote.Join(argv) }
rg -Fn '// Quote = s as one POSIX shell word.' internal/remote/ssh.go
# 80:// Quote = s as one POSIX shell word.
```

Acceptance: owner chooses (a) Decision says shell-quoted argv (matches S3 + current implementation), or (b) `%q` byte-format really binds, then implementation/tests must encode it. No behavioral red claimed for this text finding.

### R6-F4 — deploy skips check for successful non-switch actions

Severity: **minor**. Behavior target: `machines/nixbox/deploy:10` @ 41fd3c0, exact anchor `[ "${1:-switch}" != switch ] || "$here/check"`. Decision `.agent/spec.md:26` says ``runs at every `deploy` ``. The script accepts a nixos-rebuild action but calls `check` only for `switch`; `test`, `boot` and `build` all complete without it. README:14 documents the same narrower behavior, but Decisions outrank it. MAIN directed this finding against deploy and owns the fix-side ruling; R6-F3 alone remains a Decision-text finding.

Red witness: `test/e2e/specfit_review_test.go` `TestSF14DeployChecksNonSwitch`, commit `d4f708c` on `wt/reviewer-6`. The unchanged subject deploy is copied into a temp tree with fake ssh/rsync/check, fixture host.nix + inert wifi.env. Switch is the passing positive control; every action's log proves the transfer + requested rebuild mock ran, then requires one check. No real host contact or deploy/rebuild occurs.

Command (worktree cwd): `go test -count=1 -run '^TestSF14DeployChecksNonSwitch$' ./test/e2e/`.

```text
--- FAIL: TestSF14DeployChecksNonSwitch (0.43s)
    --- FAIL: TestSF14DeployChecksNonSwitch/test (0.05s)
        specfit_review_test.go:66: deploy test succeeded without check: "rsync\nssh:offload@192.0.2.1 sudo nixos-rebuild test\n"
    --- FAIL: TestSF14DeployChecksNonSwitch/boot (0.05s)
        specfit_review_test.go:66: deploy boot succeeded without check: "rsync\nssh:offload@192.0.2.1 sudo nixos-rebuild boot\n"
    --- FAIL: TestSF14DeployChecksNonSwitch/build (0.29s)
        specfit_review_test.go:66: deploy build succeeded without check: "rsync\nssh:offload@192.0.2.1 sudo nixos-rebuild build\n"
FAIL
FAIL github.com/eturkes/offload/test/e2e 2.280s
SF14 red rc=1
```

Text anchors (all matched, rc0):

```sh
rg -Fn 'runs at every `deploy`' .agent/spec.md
# 26
rg -Fn '[ "${1:-switch}" != switch ] || "$here/check"' machines/nixbox/deploy
# 10
rg -Fn 'runs it after every switch' machines/nixbox/README.md
# 14
```

Acceptance: every successful deploy action covered by the Decision invokes check; MAIN rules the implementation side. The unchanged switch path must keep its check invocation.

## Live evidence — SF14/SF15 (read-only)

Sources: subject config + SSH read-only queries. Logs in reviewer worktree `.scratch/nixbox-{compute,layout,layout-followup,trading-guards,conditions}.txt`; commands used `ssh -o BatchMode=yes -o ConnectTimeout=5 -T offload@192.168.40.240 'bash -s'`, no offload real-pool run. No service/config/resource mutation. `sudo` used only to read root-owned Nix files and query `eturkes`'s existing files/user manager.

- Subject vs deployed `/etc/nixos` SHA-256 equal: `configuration.nix` = `df3af4fec008dfe75c37925d18865159a626e7f74135b9c75003ebab19ce6326`; `idle.nix` = `e9b7381e43e7341e7c938416f76986e26183639f23e01b52896e29b75d5c3118`; `performance.nix` = `1af58a494ece905a8aaa14bc11ffdc731fae13ee4195f2d2aaf33860f841ed06`.
- `systemctl show user-1001.slice -p CPUWeight -p MemoryHigh` → idle / 21474836480; cgroup files cpu.idle=1 + memory.high=21474836480. CPU sysfs: governor/profile performance, PL1 22000000; battery 75/80 + `[Custom]`; THP `[always]` / `[defer+madvise]`; NVMe `[kyber]`; cmdline includes mitigations=off, nowatchdog, transparent_hugepage=always, zswap.enabled=0. `cpu-power-limit`/`battery-charge-limit` success; earlyoom running; thermald inactive.
- `sysctl` matches every performance.nix value: swappiness150, vfs_cache_pressure50, dirty_bytes268435456/background67108864/writeback_centisecs1500, page-cluster0, nmi_watchdog0, split_lock_mitigate0, kptr_restrict2, netdev_max_backlog4096. zram zstd/125.2G/priority100; `/tmp` tmpfs size32827588k; NOFILE soft2048/hard2097152. `systemd-analyze cat-config systemd/logind.conf` → all three lid settings ignore; four sleep targets masked.
- UV_PYTHON_PREFERENCE=only-managed, OFFLOAD_CTR=1; `/etc/ssl/cert.pem` byte-equal system CA bundle (`cmp -s`, rc0); `/etc/localtime` resolves UTC; fc-match sans/serif/mono/emoji → Noto Sans CJK JP / Noto Serif CJK JP / Noto Sans Mono CJK JP / Noto Color Emoji.
- Tools: elan4.2.1, rustup1.29.2/default stable, Go1.26.8/GOTOOLCHAIN=auto (project pin auto-fetches), offload-user node26.12.1/system node24.21.0, clang21.1.8, shellcheck0.11.0, SWI-Prolog10.0.2, uv0.13.4, bwrap0.12.0, podman5.8.8. Envfs mounts active; nix-ld2.0.6 loader; git user.name=offload; offload subuid/subgid100000:65536. Controller podman6.1.4, OFFLOAD_CTR=1. These prove provisioned prerequisites, not every consumer gate or fresh container launch.
- Codexify `projects list --query trading` → exactly 1/1 match, home, selector `.`, alias trading, access root `/home/eturkes`; config selected `worktrees.mode=never`; `/home/eturkes/AGENTS.md:9,11` routes to Projects/trading. Trading git directory exists (rc0), controller trading directory absent (test-e rc1; offload dir positive control rc0).
- Trading service-file listing: operation units disabled (two research ones static), Codexify enabled/running. All enumerated operation names carry `95-trading-pause.conf` + `96-operational-scope.conf` conditions; optional uninstalled bsc-luna has guards too. `rg -n '^ConditionPathExists=' ~/.config/systemd/user` showed global-hold-absent + per-unit permit conditions; Luna adds luna-usage-paused-absent. Luna hold file exists (rc0); global trading hold absent, as current trading `AGENTS.md:13` + `docs/operations.md:8` prescribe. `docs/operations.md:83,94–95` says v1 source-stopped, owner-approved reviewed v2/controller/permit required, no v2/permit exists. This is a read of operational authority + installed guards, not a start attempt or a full independent audit of trading's code.

## Verification limits

- Go LSP requested on `internal/cli/place.go` → no server for `.go`; navigation by file reads + bounded searches.
- Sync selected batch rc 1: `TestY4LinkedWorktreeRunsRemotely` timed out after 20 s after printing correct branch/HEAD/index output; `TestY6Download/command-0` timed out after 20 s. Individual reruns rc 0 (11.721 s / 24.014 s package runtime). Cause unconfirmed; concurrent workload is only a hypothesis, not an established explanation. No target change between attempts.

- Live-query failures corrected: `go --version` is not Go's syntax → explicit `go version` succeeded; two root-owned Nix hash reads denied → sudo read-only hashes succeeded; guessed config `.projects` returned null → documented Codexify catalogue CLI proved the entry; wildcard `systemctl cat` yielded no unit text → explicit unit show + all installed condition files used instead. No empty result from those failed approaches is used as absence proof.
- Bounds: one live host (nixbox), CPU governor spot-check cpu0, no thermal/stress rerun, no exhaustive consumer-tool invocation, no ad-hoc trading run; source-path discovery capped at 30 lines and service-file listing at 80 (28 actual), followed by targeted authority + uncapped guard reads. Historical peaks/thermal limit/benchmark/CI capacity numbers not independently reconfirmed. Other repos' adoption/guidance + migration's past approval/checksum not audited.
- Not run: full gate, race lane, mutation sweep, real-podman e2e, live install/deploy/rebuild, real-pool offload jobs, live CI dispatch. Deploy ran only in R6-F4's isolated test with all remote effects mocked. These are outside this reviewer's write/resource scope, not implicitly green.
- Executed controls (worktree cwd, each rc0 except documented sync batch): `go test -count=1 -run '^Test(C2ConflictsAndPoolImplications|C5PlacementLine|L1Eligibility|L2Placement|L3WaitingSignals|L4NoGPU)$' ./test/e2e/`; `go test -count=1 -run '^Test(X1ControllerDeathKillsSession|X3SignalsSkipDownload|X4ByteExactStreaming|X6Remote255AndTransport|S4SameTreeRunsSerializeAtController)$' ./test/e2e/`; sync batch `^Test(Y1RootAndSubdirectory|Y2WorkdirIdentity|Y4GitUploadDifferential|Y5GlobalExcludesReachTheHost|Y4LinkedWorktreeRunsRemotely|Y4GitMetadataMirrorDeletes|Y5Prune|Y6Download|Y7PlainMirror|Y8Get|Y9NoSync)$`→rc1, then individual `^TestY4LinkedWorktreeRunsRemotely$` + `^TestY6Download$`→rc0; slots `^Test(HeavySlotsAgreesWithHeavy|SlotHeldByHeavyIsBusy|SlotHeldByOffloadIsBusyToHeavy|LowestFreeSlotFirst|CommandNeverInheritsLock|KilledHolderFreesSlot|S1RuntimeDirFallback)$`; whole `./internal/pool/`; probe `^Test(Parse|ScriptOnThisMachine|R1ParseBoundaries|R1MalformedCounters)$`; remote `^Test(TargetCarriesOptions|JoinRoundTripsThroughBash|Multiplex)$`; ctr `^Test(K4TagEqualForEqualTrees|K4TagChangesWithEveryInput|K5ReviewEnsure|K6RunArgs|K9ReviewLayers)$`; e2e log `go test -parallel=2 -count=1 -run '^Test(J2LocalRunRecords|J2PoolRunRecords|J2UnplacedRunWritesEndAlone|J4LogFailureLeavesRunUnchanged|J2NoSyncTimesTheCommandAlone)$' ./test/e2e/`; `^TestHarnessFakeSSH$` e2e; whole `./tools/pooladdrs/`. Every command omitted above uses `go test -count=1 -run '<regex>' ./internal/<named-package>/` or the explicitly named e2e package. Red tests/commands/results are in findings.

## Register

- MAIN note / unproven scope question: `tools/install:28–29` does not pass StrictHostKeyChecking=accept-new (core transport does); installer applicability of the shared SSH Decision left to MAIN, no security assessment performed.

Completion validator: 15 distinct SF rows, 0 unknown; all finding sections complete. Unknown-cell grep → 0 (grep rc1, expected for no matches). Text anchors rematched against subject; worktree diff = the two committed red test files alone. No Intent, CLAUDE.md, `.claude/agents/` or machine configuration edits.

reviewer-6-DONE-1
