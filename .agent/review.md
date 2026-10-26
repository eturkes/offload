# Review ledger — MAINTAIN: code review of the implementation (HEAD 41fd3c0)

Subject = code, tests, gates, config + docs at 41fd3c0. Check set fixed + committed before the subject's first read (law read first: `CLAUDE.md`, `.claude/rules/*`, `.claude/agents/*`, `.agent/spec.md`, `.agent/deferred.md`, prior ledgers). Row = verdict (`pass` | `finding` + R-ids | `n/a` | `not run` + reason) + evidence; `open` rows block the close. Findings table below: severity + `file:line` + proof (red test | `rg -Fn` anchor bytes), fix commit, acceptance. Teammate claims = leads; MAIN's check decides. Security defects met on any lens → MAIN alone (security lens = the `Security review` body; prior rows `ledger-maintain-security-review.md`).

Modules (rows = lens × module): m1 cli = `cmd/offload`, `internal/cli` · m2 run = `internal/run` (go + bash) · m3 mirror · m4 slots · m5 pool · m6 probe · m7 remote · m8 ctr (`ctr.go`, `init.go`, `tag.go`) · m9 ledger · m10 shell · m11 e2e harness (`test/e2e` harness, `testdata/ssh`, `testdata/ctrhelper`) · m12 tools (`tools/heavy`, `tools/install`, `tools/pooladdrs`) · m13 gate (`.agent/{gate,tool,pins,plant,spec-lint}`, `.golangci.yml`, `statix.toml`, `.github/**`) · m14 machines (`machines/nixbox/**`).

## B — baseline: full gate + each runnable `Artifacts` entry by its run command (red step / failed run = finding)

| id | check | verdict | evidence / ruling | sev | state |
|----|-------|---------|-------------------|-----|-------|
| B1 | full gate `offload run -- .agent/gate` on 41fd3c0: rc 0, per-step rc + seconds, 10 slowest tests | pass | `offload run -- .agent/gate` on 41fd3c0 (nixbox slot 1/2): rc 0, 13 steps rc 0 (test 46.2 s of 53.2 s run; next fmt/lint/vuln ≤ 2.6 s), 475 tests, 10 slowest printed (max TestCancelEndsWholeSession 14.3 s); run log `run` 53.21 s | – | closed |
| B2 | fast gate `.agent/gate fast` (+ one `[pkg…]` form) | pass | `.agent/gate fast` rc 0 (controller, uncached: test 111.7 s under 6 reviewers' load); `.agent/gate fast ./internal/pool/` rc 0 (test 0.7 s) | – | closed |
| B3 | GitHub: latest `Gate`, `CodeQL`, `pins` runs on `main` green; open code-scanning + Dependabot alerts | pass | `gh run list` main: Gate + CodeQL success on 41fd3c0 (+ e071c9a, ebfb7b4); Pins: no `main` run yet (schedule Tue, last dispatch 6040e73 success); open alerts: code-scanning 0, Dependabot 0, secret-scanning 0; origin/main = 41fd3c0 | – | closed |
| B4 | `offload status` (controller slots + every pool host row) | pass | `offload status` rc 0: `local heavy slots: 0/2 held` + nixbox row up 1/2 8 CPUs load/busy/mem/GPUs | – | closed |
| B5 | `offload run -- <cmd>` default placement: pool first, stderr names placement, rc passed through (0 + nonzero) | pass | git fixture `/tmp/offload-cr/proj`: `offload: nixbox slot 1/2 (8 CPUs, 79% busy) → ~/offload/proj-5a649768`, `exit 3` → rc 3, `uname -n` → nixbox rc 0 | – | closed |
| B6 | `offload run --local -- <cmd>` (heavy slot, owner line visible to `heavy --status`) | pass | `--local`: `offload: local slot 1/2`, `heavy --status` + `offload status` during it list `slot 1: <pid> since 14:38:49: sh -c 'sleep 3; …'`; rc 0 | – | closed |
| B7 | `offload run --pool` + `--host nixbox` + unknown `--host` | pass | `--pool`, `--host nixbox` → nixbox rc 0; `--host nope` → `no host named nope in …/pool.toml` rc 2 | – | closed |
| B8 | `offload run --gpu` (no NVIDIA pool host) | pass | `--gpu` (nixbox 0 GPUs, up) → `offload: no pool host has an NVIDIA GPU` rc 125 (L4) | – | closed |
| B9 | `offload run --no-sync` | pass | `--no-sync -- sh -c 'pwd; ls'` → existing workdir, rc 0, run log upload/download null | – | closed |
| B10 | `offload run --get PATH` (ignored path fetched back) | pass | `--get out/` (ignored `out/`) → local `out/res` = `res`, rc 0 | – | closed |
| B11 | run from a subdir + from a linked git worktree + from a non-git dir | pass | subdir `sub` → remote cwd `…/proj-5a649768/sub`; linked worktree → `~/offload/proj-wt-a39c9572`, `A  w.txt` staged + branch `wt` seen; plain dir → rsync mirror, ignored `ign/` stayed, `made.txt` came back; usage forms (`--local --pool`, empty) rc 2, missing command rc 127 | – | closed |
| B12 | `offload ctr` fixture = a non-git scratch dir outside the repo (sync root = cwd) holding a minimal `.devcontainer/Containerfile`: `offload ctr -- <cmd>` with `OFFLOAD_CTR` set + unset, `--layers`, `--check`, and the consumer form `offload run -- offload ctr -- <cmd>` | pass | non-git `/tmp/offload-cr/img` (`FROM scratch` + `COPY hello /hello`, static `hello`): OFFLOAD_CTR=1 builds `localhost/offload-ctr:1ebf291404d18ece`, `./hello 7` → pid 7 HOME `/.offload/home` rc 7; `/hello 0` rc 0; =0 / unset → native pid, rc 5/6; =2 → rc 2; missing → 127; `--layers` → 1 digest; `--check` → `nixbox: same (1 layers)` rc 0 (nixbox built it); `offload run -- offload ctr -- ./hello 4` → rc 4 in nixbox's image; `--layers --check` rc 2; no definition → rc 2 | – | closed |
| B13 | run log: start + end records for B5–B11 runs in `~/.local/state/offload/runs.jsonl` (README `See past runs` recipe) | pass | runs.jsonl: start+end per placed run (14), end-only for never-placed rc 2/125/127 runs, nothing for usage errors (J1); README recipes (last 10, rc 125) rc 0 with expected fields | – | closed |
| B14 | `heavy` (`heavy --status`, `heavy <cmd>`) beside offload's local slots | pass | `heavy sh -c '…; exit 4'` rc 4; `heavy --status` rc 0; `heavy` → usage rc 2; HEAVY_SLOTS=x → rc 2; offload with HEAVY_SLOTS=0 → rc 2 (heavy's message) | – | closed |
| B15 | `tools/install` (at close when a shipped change landed) + `offload ctr --check` after it | unknown | | | open |
| B16 | `.agent/pins` | finding | `.agent/pins` rc 1: `mutago | 2.10.37 | 2.10.38 | behind` (other 12 pins current) → R1 | minor | open |
| B17 | `pool.example.toml` parses (`OFFLOAD_CONFIG=pool.example.toml offload status`) | pass | `OFFLOAD_CONFIG=pool.example.toml offload status` rc 0, nixbox row up | – | closed |
| B18 | `machines/nixbox/deploy` (+ its `check`) | pass | box `/etc/nixos/*.nix` = 41fd3c0 bytes (cmp ×8); `machines/nixbox/deploy dry-activate` rc 0, built config = `/run/current-system` (qg85bjd6…, nothing to activate); `machines/nixbox/check` → `check: ok` rc 0. A real `switch` runs only if a `machines/` change lands | – | closed |
| B19 | `machines/nixbox/{install,tpm-bind}`, installer ISO, firmware: driven where non-destructive (usage/arg paths, evaluation); destructive runs named `not run` | pass | `efi-entry` dry run on the box (`DRY_RUN=1`) rc 0: delete Boot0000 + create one `Linux Boot Manager`; installer ISO evaluates on nixbox with dummy `wifi.env` + `installer_host_key` (`nix-instantiate` → `…-nixos-minimal-26.05….iso.drv`, rc 0); usage paths: `install` (no disk) rc 1 `usage: install /dev/<disk>`, `tpm-bind --bogus` rc 2, `check --bogus` rc 2. Not run (destructive / boot-security state): `install /dev/nvme0n1` (wipes the disk), `tpm-bind` (re-seals LUKS), ISO build, firmware update | – | closed |

## C — Correctness (`Engineering` code diff lens): each function with its callers + callees; guards + anchors; classic pitfalls of the language (Go: error shadowing, nil maps, defer in loops, goroutine/ctx/fd leaks, races, integer conversion, close errors; bash: quoting, `set -e` traps, subshells, pipefail; Nix: option semantics); wrappers route to the wrapped instance; defaults intended; test setup ↔ teardown symmetric

| id | check | verdict | evidence / ruling | sev | state |
|----|-------|---------|-------------------|-----|-------|
| C1 | m1 cli | finding | reviewer-1 (C1): hybrid lookup before placement; `parseCheck` Scanner 64 KiB token, `sc.Err()` ignored; remote `cmd_rc` dropped under cancel; MAIN: `ctrCheck` children via `exec.CommandContext` = SIGKILL on a signal → R14, R18, R12, R25 | minor | open |
| C2 | m2 run | finding | reviewer-1: exported `SHELLOPTS=errexit` ends the watcher (`true` → rc 1); reviewer-2: prune sort/comm pipelines unchecked → R21, R15 | minor | open |
| C3 | m3 mirror | finding | reviewer-2: relative `core.excludesFile` read from the invocation cwd → R19, R15 | minor | open |
| C4 | m4 slots | finding | reviewer-1: `Holders` counts an open error as held; owner text read blocking (FIFO) + following links → R20 | minor | open |
| C5 | m5 pool | pass | reviewer-2 `.agent/archive/review/cr-reviewer-2.md`: P1–P3 decode/validation, callers fit; `./internal/pool` rc 0 | – | closed |
| C6 | m6 probe | pass | reviewer-2: 4-field parser, 7.5 s + 0.5 s WaitDelay, callers fan out/join; `./internal/probe` rc 0 | – | closed |
| C7 | m7 remote | pass | reviewer-1 `.agent/archive/review/cr-reviewer-1.md`: SSHOptions + `-T` on every path, mux resolved once, Quote/Join + process-group cancel fit; `./internal/remote` rc 0 | – | closed |
| C8 | m8 ctr | pass | reviewer-2: tag framing, build lock inheritance, argv, init PID-1/fd 3/reap/teardown; `./internal/ctr` rc 0 (real-podman K8 readiness reds under load → deferred row) | – | closed |
| C9 | m9 ledger | finding | reviewer-1 + reviewer-4 independently: `End` returns a second `final()` read → R11 | minor | open |
| C10 | m10 shell | finding | reviewer-1, -4, -6 independently: `filepath.SplitList("")` = no entries; bash searches cwd → R13 | minor | open |
| C11 | m11 e2e harness | finding | reviewer-1: fake ssh `*loadavg*|*nproc*` matches command words; reviewer-3: `fakePodman` resolves real podman before writing a full fake → R22, R23 | minor | open |
| C12 | m12 tools | finding | `tools/heavy` + `pooladdrs` pass (reviewer-2); MAIN: `tools/install` prints sums, never compares them; its ssh lacks `StrictHostKeyChecking=accept-new` (reviewer-6 register) → R27, R28 | minor | open |
| C13 | m13 gate | pass | reviewer-2: step rc captured per step, aggregate rc; `.agent/tool` digest refusal; pins rc precedence; plant revert | – | closed |
| C14 | m14 machines | finding | reviewer-2: `efi-entry` listing rc masked by sed; collector pipelines fail open (MAIN rules); reviewer-6: deploy runs `check` after `switch` alone; reviewer-3: deploy/tpm-bind ssh without forwarding/batch options → R16, R17, R26, R27 | minor | open |

## U — Cleanup: existing helpers reused; derivable state computed, never duplicated; near-duplicates merged; dead code removed; independent steps parallel; startup + hot paths lean; closures capture only what they use

| id | check | verdict | evidence / ruling | sev | state |
|----|-------|---------|-------------------|-----|-------|
| U1 | m1 cli | pass | reviewer-3 `.agent/archive/review/cr-reviewer-3.md` (rematched) | – | closed |
| U2 | m2 run | pass | reviewer-3 `.agent/archive/review/cr-reviewer-3.md` (rematched) | – | closed |
| U3 | m3 mirror | pass | reviewer-3 `.agent/archive/review/cr-reviewer-3.md` (rematched) | – | closed |
| U4 | m4 slots | pass | reviewer-3 `.agent/archive/review/cr-reviewer-3.md` (rematched) | – | closed |
| U5 | m5 pool | pass | reviewer-3 `.agent/archive/review/cr-reviewer-3.md` (rematched) | – | closed |
| U6 | m6 probe | pass | reviewer-3 `.agent/archive/review/cr-reviewer-3.md` (rematched) | – | closed |
| U7 | m7 remote | pass | reviewer-3 `.agent/archive/review/cr-reviewer-3.md` (rematched) | – | closed |
| U8 | m8 ctr | pass | reviewer-3 `.agent/archive/review/cr-reviewer-3.md` (rematched) | – | closed |
| U9 | m9 ledger | pass | reviewer-3 `.agent/archive/review/cr-reviewer-3.md` (rematched) | – | closed |
| U10 | m10 shell | pass | reviewer-3 `.agent/archive/review/cr-reviewer-3.md` (rematched) | – | closed |
| U11 | m11 e2e harness + test helpers (duplicated helpers across `*_test.go`) | finding | reviewer-3: `requireScope` probe + fail/skip policy twice (`internal/run/run_test.go:128`, `test/e2e/execution_test.go:134`); register: ctrhelper mode `id` without a caller → R24 | minor | open |
| U12 | m12 tools | pass | reviewer-3 `.agent/archive/review/cr-reviewer-3.md` (rematched) | – | closed |
| U13 | m13 gate | pass | reviewer-3 `.agent/archive/review/cr-reviewer-3.md` (rematched) | – | closed |
| U14 | m14 machines | pass | reviewer-3 `.agent/archive/review/cr-reviewer-3.md` (rematched) | – | closed |

## A — Altitude: a special case on shared infrastructure = a fix due in the mechanism

| id | check | verdict | evidence / ruling | sev | state |
|----|-------|---------|-------------------|-----|-------|
| A1 | m1 cli | pass | reviewer-3 `.agent/archive/review/cr-reviewer-3.md` | – | closed |
| A2 | m2 run | pass | reviewer-3 `.agent/archive/review/cr-reviewer-3.md` | – | closed |
| A3 | m3 mirror | pass | reviewer-3 `.agent/archive/review/cr-reviewer-3.md` | – | closed |
| A4 | m4 slots | pass | reviewer-3 `.agent/archive/review/cr-reviewer-3.md` | – | closed |
| A5 | m5 pool | pass | reviewer-3 `.agent/archive/review/cr-reviewer-3.md` | – | closed |
| A6 | m6 probe | pass | reviewer-3 `.agent/archive/review/cr-reviewer-3.md` | – | closed |
| A7 | m7 remote | pass | reviewer-3 `.agent/archive/review/cr-reviewer-3.md` | – | closed |
| A8 | m8 ctr | pass | reviewer-3 `.agent/archive/review/cr-reviewer-3.md` | – | closed |
| A9 | m9 ledger | pass | reviewer-3 `.agent/archive/review/cr-reviewer-3.md` | – | closed |
| A10 | m10 shell | pass | reviewer-3 `.agent/archive/review/cr-reviewer-3.md` | – | closed |
| A11 | m11 e2e harness | pass | reviewer-3 `.agent/archive/review/cr-reviewer-3.md` | – | closed |
| A12 | m12 tools | pass | reviewer-3 `.agent/archive/review/cr-reviewer-3.md` | – | closed |
| A13 | m13 gate | pass | reviewer-3 `.agent/archive/review/cr-reviewer-3.md` | – | closed |
| A14 | m14 machines | pass | reviewer-3 `.agent/archive/review/cr-reviewer-3.md` | – | closed |

## G — claim soundness + guarantee-vs-claim gaps (each claim vs observed behavior)

| id | check | verdict | evidence / ruling | sev | state |
|----|-------|---------|-------------------|-----|-------|
| G1 | `README.md` (every promise, example + recipe) | finding | reviewer-4 `.agent/archive/review/cr-reviewer-4.md`: log rc + native-ctr parity promises vs reds; `--no-sync` row omits `--get` → R11, R13, R29 | minor | open |
| G2 | `--help` / usage + user-facing error texts (`offload`, `offload run`, `offload ctr`, `heavy`) | finding | reviewer-4: usage omits the `XDG_CONFIG_HOME` tier; `--no-sync` line → R29 | minor | open |
| G3a | `.agent/contract.md` family C predicates vs behavior (each id) | finding | reviewer-4: C3 lists `worktree` among 125s while Y3/Y4 run linked worktrees → R29 | minor | open |
| G3b | `.agent/contract.md` family P predicates vs behavior (each id) | pass | reviewer-4: P1–P3 witnesses rc 0 | – | closed |
| G3c | `.agent/contract.md` family R predicates vs behavior (each id) | pass | reviewer-4: R1/R2 witnesses rc 0 | – | closed |
| G3d | `.agent/contract.md` family S predicates vs behavior (each id) | pass | reviewer-4: S1–S6 witnesses rc 0 | – | closed |
| G3e | `.agent/contract.md` family L predicates vs behavior (each id) | finding | reviewer-4 register R4-R1: L2 prose = probe → pick → acquire; code reserves slot + workdir lock first, then probes → R29 | minor | open |
| G3f | `.agent/contract.md` family Y predicates vs behavior (each id) | pass | reviewer-4: Y1–Y11 witnesses rc 0 | – | closed |
| G3g | `.agent/contract.md` family X predicates vs behavior (each id) | finding | reviewer-4: X2 fails on explicit empty PATH → R13 | minor | open |
| G3h | `.agent/contract.md` family J predicates vs behavior (each id) | finding | reviewer-4: J2 split status (red 7f9e791) → R11 | minor | open |
| G3i | `.agent/contract.md` family K predicates vs behavior (each id) | finding | reviewer-4: K2/K7 inherit X2 empty-PATH failure; real-podman batch reds under load → deferred row → R13 | minor | open |
| G4 | `.agent/spec.md` `Artifacts` (paths, run commands, live facts) | pass | reviewer-4: artifacts exist; installed offload + heavy sha256 equal both sides | – | closed |
| G5 | `machines/nixbox/README.md` runbook vs scripts + config | pass | reviewer-4: runbook vs scripts + live knobs agree | – | closed |
| G6 | `pool.example.toml` comments vs P-predicates | pass | reviewer-4: example parses, comments = P1/P2 | – | closed |
| G7 | facts stated in `.claude/rules/*` (timings, recipes, mechanisms) vs code | finding | reviewer-4: `code.md` `.transport-fail` = `(every call 255)`, fake ssh answers probes first → R29 | minor | open |
| G8 | code comments m1 cli | finding | reviewer-4: `ctr.go` execNative `as if typed` vs empty PATH → R13 | minor | open |
| G9 | code comments m2 run | pass | reviewer-4 | – | closed |
| G10 | code comments m3 mirror | pass | reviewer-4 | – | closed |
| G11 | code comments m4 slots | finding | reviewer-4: `slots.go:18` `the kernel frees it if offload dies first` vs S5 watcher inheritance → R29 | minor | open |
| G12 | code comments m5 pool + m6 probe | pass | reviewer-4 | – | closed |
| G13 | code comments m7 remote + m10 shell | finding | reviewer-4: `shell/look.go` package comment promises bash parity → R13 | minor | open |
| G14 | code comments m8 ctr | pass | reviewer-4 | – | closed |
| G15 | code comments m9 ledger | pass | reviewer-4 | – | closed |
| G16 | code comments m11 e2e harness + test comments citing ids | finding | reviewer-4: `ctr_test.go:465` `(its own group, offload's session)` vs K8 own session → R29 | minor | open |
| G17 | comments m12 tools | pass | reviewer-4 | – | closed |
| G18 | comments m13 gate | pass | reviewer-4 | – | closed |
| G19 | comments m14 machines (nix + scripts, `*.reviewed`) | pass | reviewer-4 | – | closed |

## V — verification integrity (`Engineering`): each test seen red on its unfixed revision (revision + command in a commit body) or mutation-credited (default operators, scope = files it imports/calls, every survivor ruled); kills via observable behavior; no test detection / fixture-value return / contract-unowned expected tables in the implementation; skipped, xfailed, deleted or demoted cases carry a `deferred.md` row + owner approval; tests cite contract ids (`code.md`)

| id | check | verdict | evidence / ruling | sev | state |
|----|-------|---------|-------------------|-----|-------|
| V1 | m1 cli (`internal/cli` + e2e `cli_test.go`) | finding | reviewer-5 `.agent/archive/review/cr-reviewer-5.md`: 4 cli unit tests without red or mutation record → R31, R8 | minor | open |
| V2 | m2 run (unit + e2e execution/close/failures/locks) | finding | reviewer-5: 9 run tests uncredited, 7 selected-plant-only; environment skips (`requireScope`) without a deferred row + approval → R32, R35, R37, R8 | minor | open |
| V3 | m3 mirror (unit + mutation + e2e sync) | finding | reviewer-5: 2 selected-plant-only (`TestY5GlobalExcludesReachTheHost`, `TestReviewer9ExcludesRecordIsNotAnUploadPath`) → R35, R8 | minor | open |
| V4 | m4 slots | finding | reviewer-5 E1: 85e2c60 MakeDirs survivors without an independent ruling → R36, R8 | minor | open |
| V5 | m5 pool | pass | reviewer-5: 6 tests, 5 mutation + 1 red | – | closed |
| V6 | m6 probe | pass | reviewer-5: 6 tests, 5 mutation + 1 red | – | closed |
| V7 | m7 remote | finding | reviewer-5: `TestTargetCarriesOptions`, `TestMultiplex`, `TestReviewer9MultiplexFallbackOnceAcrossHosts` selected-plant-only (mutago finds no source in `internal/remote`: `ssh.go` has no `ssh_test.go`) → R35, R8 | minor | open |
| V8 | m8 ctr (unit + real-podman e2e) | finding | reviewer-5: 80 tests credited; `requirePodman` local skip without a deferred row + approval → R37, R8 | minor | open |
| V9 | m9 ledger (unit + e2e `ledger_test.go`) | finding | reviewer-5: 13197df's 13 survivors ruled by the test author alone → R36 | minor | open |
| V10 | m10 shell | finding | reviewer-5: `TestLookMatchesBash` uncredited → R33, R8 | minor | open |
| V11 | m11 e2e harness self-tests (fake ssh X5 refusal, `-n`, markers) | finding | reviewer-5: harness self-tests credited; ids missing on review tests → R8 | minor | open |
| V12 | m12 tools (`pooladdrs` test; `heavy` + `install` coverage) | finding | reviewer-5: `TestEveryValidPoolSpellingYieldsItsAddress` uncredited → R34 | minor | open |
| V13 | m13 gate: every step fires on its `gate-checks.md` seed (replanted) | pass | replanted in a disposable worktree with each step's own gate command (`/tmp/offload-cr/src/plant.sh`): fmt 1, lint 1, tidy 1, test 1, build 1 (vet + race on the same file 0/0), vuln 3, secrets 1 (clean history 0), actionlint 1, zizmor 14, shell 1 (file listed by the gate's shebang scan), nix 1, statix 1 (`[10] Warning: Found empty pattern`), deadnix 1 (clean tree statix/deadnix 0/0), spec-lint 1, `.agent/tool` wrong sha256 → `tool: sha256 mismatch … refusing it` rc 2; `.agent/pins`: go-version 1.27.2 → `differs from go.mod` rc 1, `runs-on: ubuntu-22.04` → `behind` rc 1, unknown repo → `lookup failed` rc 2 (go-version + runs-on rc confounded by R1's live `behind` → rerun at R1's fix); extra: bare `//nolint` → nolintlint rc 1 | – | closed |
| V14 | m14 machines: `check` fires on every `machines.md` planted fact | pass | `check --dump` (46 facts) → `--facts`: live 0 `check: ok`; + `wrapper pkexec /bin/sh` 1, + `wrapper ping -` 1, + `wrapper doas ?` 1, − `wrapper umount -` 1, upgrades → `done <now−10 d>` 1 (240 h), + `failed <now−1 h>` 1, no upgrade 1, + `failed foo.service` 1, `sshd passwordauthentication yes` 1, + `listen tcp 0.0.0.0:8080 python3` 1, `fw temp-ports 1` 1, `fw input-policy accept` 1, `secret psk-files 1` 1, last success 71 h → 0; PSK scan positive control on the box: the pre-fix `…-lan` path → 1 file | – | closed |

## SF — spec fit: behavior vs `Intent` + `Decisions`

| id | check | verdict | evidence / ruling | sev | state |
|----|-------|---------|-------------------|-----|-------|
| SF1 | m1 cli: trigger + flags, placement order (pool → local heavy → wait both, one waiting message), load ranking, `--host/--gpu/--no-sync` ⇒ `--pool`, stderr placement | finding | reviewer-6 `.agent/archive/review/cr-reviewer-6.md` + reviewer-1: hybrid `offload run` → 127 for a host-only tool; `--pool` control rc 0 → R14 | minor | open |
| SF2 | m2 run: remote run sequence, rc passthrough, kill-on-signal incl. SIGKILL (watcher), per-(host, workdir) serialization, local runs never lock | pass | reviewer-6: upload→`bash -lc`→download, rc passthrough, watcher EOF, (host, workdir) lock, local unlocked | – | closed |
| SF3 | m3 mirror: sync root, workdir naming, git-exact up/down sets, deletion pass, download `--update` no deletes, `.git` never returns, non-git mirror, linked worktree layout | pass | reviewer-6: sync root, workdir naming, U/D sets, prune, `--update`, plain mirror, worktree layout | – | closed |
| SF4 | m4 slots: heavy interop (paths, flock, owner line, emptied at exit), pool slot locks, admission vs loadavg | finding | reviewer-6: Decision `<cmd %q>` vs S3 POSIX quoting (`remote.Join`) → R30 | minor | open |
| SF5 | m5 pool: TOML `[[host]]` (`name`, `address`, `slots` default 2), `OFFLOAD_CONFIG` | pass | reviewer-6 | – | closed |
| SF6 | m6 probe: load = rank only (1-min / nproc), GPU inventory | pass | reviewer-6 | – | closed |
| SF7 | m7 remote: BatchMode + accept-new, ssh_config alias addresses | pass | reviewer-6 | – | closed |
| SF8 | m8 ctr: K-predicates, switch `OFFLOAD_CTR`, `.devcontainer/Containerfile`, tag = content hash, host builds itself, `--check` = `RootFS.Layers` equality, LAN lanes only | finding | reviewer-6: K2/K7 lookup = R13 → R13 | minor | open |
| SF9 | m9 ledger: every `run` logged (start once a slot is held + end), fields per `Decisions`, log failure never touches the run | finding | Decision `Run log` (command + offload status): end record can disagree with the exit status (R11); remote `cmd_rc` null though the status came back (R12) → R11, R12 | minor | open |
| SF10 | m10 shell: lookup = bash's (126/127) | finding | reviewer-6 → R13 | minor | open |
| SF11 | m11 e2e: harness models the pool hosts `Decisions` describe (git + coreutils, ssh options) | finding | reviewer-6 pass on fixtures; reviewer-1: the fake answers probes for any command naming `nproc`/`loadavg` → R22 | minor | open |
| SF12 | m12 tools: install = one static build → controller + every pool host, sha256 per side; heavy separate script deployed alongside | pass | reviewer-6: one static build, heavy alongside, sums per side printed (comparison automation = R28 under C12) | – | closed |
| SF13 | m13 gate + CI: `Gate compute` + `CI gate` rulings, Gate CI on Ubuntu 26.04, pins weekly, Dependabot scope | pass | reviewer-6 | – | closed |
| SF14 | m14 machines: idle task slice, accounts, LAN-only nftables, key-only ssh, LUKS/TPM/Secure Boot, compute posture, carried tools, security posture, Wi-Fi secret path, `check` at every deploy | finding | reviewer-6: compute posture, idle slice, carried tools live = config; deploy check skipped off `switch`. MAIN security items: sshd key-only + root off, nftables LAN-only 22, PSK not in the closure, `wifi.env` 600 root (V14 live facts = `config.reviewed`, B18 `check: ok`); accounts `offload` wheel / `eturkes` unprivileged, LUKS TPM2 + PCR 15 measure, lanzaboote (config read) = Decisions → R26 | minor | open |
| SF15 | `Decisions` binding no code module (trading + Codexify, instruction tiers, nixbox role, gate image placement, pool run environment, run memory) vs machines config, README, rules: each fits or is ruled n/a by name | finding | reviewer-6 pass on Codexify, tiers, nixbox role, image placement, run memory; MAIN: 2 dead clauses in those Decisions → R10 | minor | open |

## M — `CLAUDE.md` + `.claude/rules/` conformance: each clause that binds a tracked file (tool-decidable → its gate check runs + fires on a planted bad input; none → finding, fixed by wiring the check)

| id | clause → binds | verdict | evidence / ruling | sev | state |
|----|----------------|---------|-------------------|-----|-------|
| M1 | Execution `Git`: `.gitignore` current (a full gate + runs leave no untracked, unignored file) | finding | `.gitignore:4-6` `__pycache__/`, `.venv/`, `*.pyc` = template-era Python entries (9f0ce71), no Python in the repo (`.agent/plant` runs python3 from stdin, no bytecode); `.ruff_cache/` (Oct 21, ruff 0.17.1) sits in the tree hidden by its own `*` .gitignore → R2 | minor | open |
| M2 | Execution gate command: full chain recorded in `.claude/rules/`, prints each step's wall-clock + the suite's slowest tests | pass | `.agent/gate` prints `== <step>: <cmd>`, the rc + seconds table, then `slowest tests` (10) — B1 output | – | closed |
| M3 | Execution fast gate recorded beside it (changed steps + tests, parallel + cached) | pass | `gate.md` 2 records `.agent/gate fast [pkg…]` = fmt + lint + cached tests + spec-lint; B2 both forms rc 0 | – | closed |
| M4 | Execution: a step dominating the full gate → a speed-up unit in `Tasks` or `deferred.md` | finding | test = 46.2 s of the 53.2 s gate (87 %; next step 2.6 s); `internal/run` 44.5 s with 0 of 27 tests `t.Parallel()`, e2e 41.9 s; no speed-up unit in `Tasks` or `deferred.md` → R3 | minor | open |
| M5 | Execution `CI gate` → keyed ruling (`gate.md`) current vs the repo's CI | pass | `gate.md` 3 keys the `CI gate` ruling (owner) + Gate CI dispatch procedure; Gate runs on `main` push + dispatch (B3). Its `≈ 1.5 min on nixbox` figure = the existing `deferred.md` row (3 of ≥ 5 records so far) | – | closed |
| M6 | Execution task close: no temporary/stale artifacts or dangling symlinks in the tree (candidate at seed: `.ruff_cache/` in the tree) | finding | `.ruff_cache/` = stale artifact (no ruff in the repo) → R2; dangling symlinks: `find . -xtype l` outside `.git`, `.tools`, `.scratch` → 0 (control: a planted dangling link in `.scratch` → 1) | minor | open |
| M7 | Execution UI design + charts | n/a | no UI, no charts or figures in the repo | – | closed |
| M8 | Authoring: agent-optimized text in every non-human-facing artifact (comments, contract, rules, internal docs, runbook) | pass | contract, rules, runbook, comments = telegraphic, symbol-forward; human-facing = README + usage alone | – | closed |
| M9 | Authoring: rules/facts omit provenance (dates, discovery events, origin stories) | pass | `rg` dates/discovery words over `.claude/rules`, spec, contract, runbook, README → 1 hit = an advisory id (GO-2021-0113); deferred.md dates = acceptance deadlines | – | closed |
| M10 | Authoring: future-facing text states actions positively; prohibitions name the thing | pass | rules: prohibitions name the thing (`never commit, never print`, `never copy to a pool host`); actions stated as `→` | – | closed |
| M11 | Authoring: no strict text caps in guidance or checks | pass | no length linter (`.golangci.yml`: no lll), spec-lint has no size caps, no word/line limits in rules | – | closed |
| M12 | Authoring: instruction tier routing (user `CLAUDE.md` / `CLAUDE.local.md` / agents / `CLAUDE.md` / rules / spec) + `portability.md` ruling | pass | `portability.md` keys the tier ruling; machine facts = versioned law here; no `CLAUDE.local.md` | – | closed |
| M13 | Authoring: `.claude/rules/` = live law + why, no origin stories; multi-step recipe → committed script; bare tier small; area-bound → `paths:` | pass | bare tier = `gate.md` (5) + `portability.md` (1); area law under `paths:` (code, gate-checks, machines); recipes → scripts (`.agent/plant`, `tpm-bind`, `check`, `deploy`); no origin stories | – | closed |
| M14 | Authoring: rule + agent frontmatter quoting + parse | finding | frontmatter of `code.md`, `gate-checks.md`, `machines.md` = `paths:` + double-quoted globs (by eye); no gate step parses or shapes it (gate = 13 steps, none reads `.claude/`) → R4 | minor | open |
| M15 | Authoring: `@path` imports only for files CC leaves unloaded; plain paths for auto-loaded ones | pass | `CLAUDE.md:1` `@.agent/spec.md` alone (spec = not auto-loaded); rules + agents named by plain path | – | closed |
| M16 | Authoring `Human-facing` (README, `--help`, error texts): ASD-STE100 register; filler words (`simply`, `robust`, `seamlessly`, `leverage`) absent | finding | README + usage read in STE register (short imperative sentences, condition first); filler words 0 today (`rg -i -w` over README, `cli.go`, `tools/heavy`), but no gate step checks them → R5 | minor | open |
| M17 | Engineering: comments spend on the why; code states the what | pass | comments carry constraints + measurements (e.g. `run.go` boot, `watch.bash` setsid/`&` notes, `performance.nix` fsync numbers); stale comments = G lens | – | closed |
| M18 | Engineering: deterministic checks own every tool-decidable rule; each purpose-built check ships its failing input + firing in a `paths:` rule | finding | tool-decidable clauses without a check: frontmatter (R4), filler words (R5), deferred acceptance (R6), test ids (R8), `__` names (R9); checked ones fire (V13, V14, M32, M37) | minor | open |
| M19 | Engineering: gate backing a durable claim reruns from committed state; scratch-local validators recorded + scheduled | pass | gates rerun from committed state (`.agent/gate`, `check --facts`); the 13197df mutation replay script lived in `.scratch` = credit evidence, not a gate (re-credit → V rows) | – | closed |
| M20 | Engineering: assurance tiers + contracts per unit (contract states tiers where units bind) | pass | contract tiers per section (CLI/Slots/Placement/Sync/Execution/Image runs = kernel; Pool file/Probe/Run log = data) | – | closed |
| M21 | Session flow: `spec.md` shape (5 sections, `Tasks` pointer last, `Phase` = state) via `spec-lint`; every line binds current/future work | finding | `.agent/spec-lint` rc 0 (shape); every-line-binds: `Decisions` holds 2 dead clauses — `Controller copy deleted after a checksum-verified copy (confirm immediately before)` (done at IMPLEMENT close) and `cement + cnl-ckc's cachyos-specific rules → the owner trims them … (\`.agent/deferred.md\`)` (row closed in 3f92e44) → R10 (Decisions edit = owner ruling) | minor | open |
| M22 | Session flow: `deferred.md` = one entry + acceptance check each | finding | 6/6 entries carry `Accept:` today; no check (spec-lint reads spec.md alone) → R6 | minor | open |
| M23 | Session flow: `spec-lint` in the gate command; archive holds finished work | pass | `spec-lint` = gate step (B1); `.agent/archive/` holds closed phases + ledgers | – | closed |
| M24 | Session flow: `CLAUDE.md` + `.claude/agents/` = current template copies | finding | `CLAUDE.md` + 6 agents = template bytes (cmp); `.agent/spec-lint` behind the template (`copy-template --check` rc 1: one `# shellcheck disable=SC2016` line, template commit 7e8ad36 today) → R7 | minor | open |
| M25 | `code.md` 1: contract = law; tests cite predicate ids | finding | reviewer-5: 21 tests cite no contract id; no check → R8 | minor | open |
| M26 | `code.md` 2: watch/remote/boot bash-only, GNU tools, no controller paths; every variable `__…`; `<&0` + `trap - INT QUIT` kept | finding | all variables in `watch.bash`, `remote.bash`, `local.bash`, boot are `__…` (read); the guard `TestReviewer9RemoteKeepsCallerVariables` exports a fixed list of 15 names ⇒ a new unprefixed variable outside that list stays green → R9 | minor | open |
| M27 | `code.md` 3: host deletion = `prune_one` + worktree `rm -f .git`; rsync `--delete` only inside `.git` / `<meta>/repo` | pass | `--delete` only in `UploadGit` (.git / `<meta>/repo`) + `UploadPlain` (Y7); host deletions = `prune_one`, worktree `rm -f .git`, meta files; U without `--delete` guarded by TestG5UploadTypeChangeKeepsIgnoredRemoteFile | – | closed |
| M28 | `code.md` 4: init PID-1 guard, handlers (no `signal.Ignore`), podman `Setsid` + scope + no `INVOCATION_ID`, nonblocking control end, `CGO_ENABLED=0` on every shipping build | pass | init refuses unless PID 1 (TestK8ReviewInitRefusesOutsidePID1), handlers not `signal.Ignore` (`init.go:39`), podman `Setsid` + scope + no `INVOCATION_ID` (`ctr.go` podmanCommand), nonblocking control end (`controlPair`); `CGO_ENABLED=0` in `tools/install`, gate build, e2e TestMain, CodeQL build | – | closed |
| M29 | `code.md` 5: shipped change → `tools/install`; `--check` reds a differing binary | unknown | | | open |
| M30 | `code.md` 6: real-podman e2e isolation (vfs storage, `FROM scratch` + ctrhelper static, prebuild fails never skips; CI `OFFLOAD_REQUIRE_PODMAN=1`) | pass | `ctr_test.go`: one vfs storage per package run (`CONTAINERS_STORAGE_CONF`), removed via `podman unshare rm -rf`; prebuild failure → Fatal under `OFFLOAD_REQUIRE_PODMAN`; `gate.yml` sets `OFFLOAD_REQUIRE_PODMAN: "1"` | – | closed |
| M31 | `code.md` 7: run log ownership (`runLogged`), J4, `signalStatus` ordering, e2e `XDG_STATE_HOME`, `runPair` validator | pass | `cli.go` runCommand → runLogged + one `End`; `signalStatus()` read after `ctx.Err()`; harness sets `XDG_STATE_HOME` (`harness_test.go:103`); `runPair` = exact key sets | – | closed |
| M32 | `code.md` 8: every ssh/rsync connection takes `remote.SSHOptions`; fake ssh refuses otherwise (rc 254), honors `-n` | pass | `testdata/ssh` refuses a call lacking any of the 8 X5 options (rc 254), honors `-n` | – | closed |
| M33 | `code.md` 9–10: fake-host marker semantics + mutation recipe (`.agent/plant`) match the harness + script | pass | markers in `testdata/ssh` = `.probe`, `.probe-hang`, `.transport-fail`, `.run-fail` (boot = `sha256sum`), `.upload-fail`/`.download-fail` (rc 23), `.ssh-call-<pid>` = `code.md` list; `.agent/plant` = the recipe's apply/run/`git checkout` revert | – | closed |
| M34 | `gate.md` 1–2: step list + fast gate = `.agent/gate`; CI runs the same command | pass | `gate.md` step list = `.agent/gate` (13 steps, same order); `gate.yml` runs `.agent/gate` | – | closed |
| M35 | `gate.md` 3: CI-gate ruling + Gate CI dispatch procedure + timing claims | pass | see M5; the `≈ 3 min` CI figure holds (Gate runs 148–203 s, deferred row) | – | closed |
| M36 | `gate.md` 4: tool pins (sha256 / sumdb / commit id), `go.mod` go line = CI setup-go + `GOTOOLCHAIN=local` | pass | release assets sha256 (`.agent/tool`), go-installed via sumdb, statix/deadnix pinned commit ids; `go.mod` `go 1.27.3` = setup-go `"1.27.3"` ×2 + `GOTOOLCHAIN: local` | – | closed |
| M37 | `gate.md` 5: suppressions name rule + reason (`nolintlint` fires on a bare one) | pass | `nolintlint` require-explanation + require-specific; bare `//nolint` seed → rc 1 (V13) | – | closed |
| M38 | `gate.md` 6: CodeQL + Dependabot scope + `pins` coverage of hand-moved pins | pass | CodeQL Go manual build; Dependabot gomod + github-actions weekly; `.agent/pins` covers `.agent/tool`, go line, `runs-on` (B16). Register: `machines/nixbox` fetchTarball pins (lanzaboote v1.2.0, nixos-unstable rev) sit outside `pins` → deferred row | – | closed |
| M39 | `gate-checks.md`: every recorded seed still fires (= V13) + every gate step has a seed | pass | V13: every recorded seed fires; every gate step (13) has one | – | closed |
| M40 | `machines.md` 1 + 4: `wifi.env` + `installer_host_key` gitignored, never committed; secrets referenced by string path, never a Nix path or inline | pass | `git check-ignore`: `machines/*/wifi.env`, `installer_host_key`(+`.pub`) ignored, `!! machines/nixbox/installer_host_key` untracked; system config names `/etc/nixos/wifi.env` by string (`network.nix`), PSK scan of the live closure 0 (V14, B18); installer.nix's Nix-path copies = the ISO's own store (runbook) | – | closed |
| M41 | `machines.md` 2–3 + 6–7: deploy/boot-chain/recovery-key/Codexify/reboot facts vs scripts + runbook | pass | `deploy` = rsync + `nixos-rebuild` + `check`; recovery key path `~/.local/share/offload/nixbox-recovery-key` in `install` + `tpm-bind`; Codexify + reboot facts = runbook; (check-at-every-deploy vs switch-only → SF14) | – | closed |
| M42 | `machines.md` 5: `check` planted-fact list (= V14) | pass | = V14 | – | closed |
| M43 | `portability.md`: ruling keyed + consistent with `CLAUDE.md` `Authoring` | pass | `portability.md` keys `Authoring` portability as inapplicable (owner) + names the replacement (versioned machine law) | – | closed |

## Findings

| id | finding | sev | `file:line` + proof | fix | acceptance | state |
|----|---------|-----|---------------------|-----|------------|-------|
| R1 | `mutago` pin 2.10.37 behind its latest release 2.10.38 (B16) | minor | `.agent/tool` `mutago) v=2.10.37` · `.agent/pins` rc 1 `behind` | b7fb72f | `.agent/pins` rc 0; Gate CI run 40278723558 (ci/cr-r1) success | closed |
| R2 | stale ignore entries + stale artifact (M1, M6) | minor | `.gitignore:4` `__pycache__/` (+ `.venv/`, `*.pyc`); `.ruff_cache/` in the tree |  | `.gitignore` = paths this repo makes; `ls -a` shows no `.ruff_cache`; gate leaves `git status --porcelain` empty | open |
| R3 | test step dominates the gate with no speed-up unit (M4) | minor | B1: test 46.2 s / 53.2 s; `internal/run` 0/27 parallel |  | a `deferred.md` row with an equal-strength acceptance check | open |
| R4 | rule frontmatter unchecked (M14) | minor | gate steps: none reads `.claude/rules` |  | gate step reds a rule with an unquoted glob / unknown key (seed recorded) | open |
| R5 | human-facing filler words unchecked (M16) | minor | gate steps: none reads README/usage for wording |  | gate step reds a planted `simply` in README (seed recorded) | open |
| R6 | deferred acceptance per entry unchecked (M22) | minor | `.agent/spec-lint` reads spec.md alone |  | gate step reds an entry without `Accept:` (seed recorded) | open |
| R7 | `.agent/spec-lint` behind the template (M24) | minor | `copy-template --check` rc 1 |  | `copy-template --check` rc 0 | open |
| R9 | `__` reservation guard covers a fixed name list (M26) | minor | `internal/run/reviewer9_test.go:104` `names := []string{"f", "d", …}` |  | guard derives names from the scripts; planted `x=1` in `remote.bash` → red | open |
| R10 | `Decisions` carries 2 dead clauses (M21) | minor | `rg -Fn 'Controller copy deleted' .agent/spec.md`; `rg -Fn 'owner trims them' .agent/spec.md` |  | owner ruling → clauses pruned, spec-lint rc 0 | open |
| R8 | 21 tests cite no contract id + no check (M25, V1–V11) | minor | reviewer-5 R5-F5 table (21 `func Test…` anchors) | | | open |
| R11 | end record + exit status can disagree (J2): `End` re-reads `final()` after the write | major | `internal/ledger/ledger.go:228` `rc, _ := final()`; red TestR4J2EndRecordAndExitShareFinalStatus (wt/reviewer-4 7f9e791) + TestReviewR1J2EndStatusFrozenWithRecord (wt/reviewer-1 3b56e10), rc 1 rerun by MAIN | | | open |
| R12 | remote `cmd_rc` dropped under cancellation though the status came back (J2) | minor | `internal/cli/cli.go:305` `if err != nil \|\| ctx.Err() != nil {`; red TestReviewR1J2CancellationKeepsReturnedCommandStatus rc 1 (`cmd_rc=<nil>`, want 17) | | | open |
| R13 | explicit empty PATH → 127 where bash runs the cwd command (X2, K2, K7) | minor | `internal/shell/look.go:28` `filepath.SplitList(os.Getenv("PATH"))`; reds TestReviewR1X2EmptyPathMatchesBash, TestR4X2EmptyPATHSearchesCurrentDirectory, TestSF10EmptyPathMatchesBash rc 1 | | | open |
| R14 | hybrid run refuses a command the pool host has and the controller lacks (Decision `Trigger`, L1) | major | `internal/cli/cli.go:240` `if !o.pool {`; reds TestSF1HybridRunsPoolOnlyCommand + TestReviewR1L2HybridCommandExistsOnlyOnHost rc 1 (127; `--pool` control rc 0) | | | open |
| R15 | remote prune inputs (sort/comm/excludes) unchecked → CMD runs on a wrong prune (Y5, Y10) | major | `internal/run/remote.bash:64` `tail -z -n +3 "$__meta/upload" \| LC_ALL=C sort -z`; red TestCRR2Y10PruneSortFailurePreventsCommand rc 1 (rc 0, ran=true) | | | open |
| R16 | `efi-entry` treats a failed listing as no stale entries | minor | `machines/nixbox/efi-entry:13` `stale=$("$efi" \| sed -n`; red TestCRR2EFIReadFailureStopsBeforeChanges rc 1 | | | open |
| R17 | `check` collectors fail open: a failing `systemctl --failed` or closure scan reads as clean | minor | `machines/nixbox/check:37` `systemctl --failed --no-legend --plain \| awk`; `:62` `xargs grep -rlF -f "$f" 2>/dev/null \| wc -l` | | | open |
| R18 | `--check` layer list truncated by a ≥ 64 KiB unframed line → false `same` (K10) | major | `internal/cli/ctr.go:228` `bufio.NewScanner(strings.NewReader(out))`; red TestReviewR1K10LongNoiseCannotTruncateLayers rc 1 | | | open |
| R19 | relative `core.excludesFile` read from the invocation cwd, not the root git used (Y5) | minor | `internal/mirror/mirror.go:233` `data, err := os.ReadFile(path)`; red TestCRR2Y5RelativeExcludesFromSubdirectory rc 1 | | | open |
| R20 | `status` counts a slot open error as held; owner text read blocks on a FIFO + follows links (R2, S15 class) | minor | `internal/slots/slots.go:155` `f, held, err := tryLock(p)`, `:160` `os.ReadFile(p + ".owner")`; red TestReviewR1R2StatusDoesNotTurnIOErrorIntoHolder rc 1 | | | open |
| R21 | exported `SHELLOPTS` (errexit…) changes the watcher's control flow (C3, X1) | minor | `internal/run/watch.bash:86` `pkill -TERM -s "$__p" 2>/dev/null`; red TestReviewR1X2ExportedErrexitPreservesStatus rc 1 (`true` → 1) | | | open |
| R22 | fake ssh serves the probe answer to any command naming `nproc`/`loadavg` | minor | `test/e2e/testdata/ssh:37` `*loadavg*\|*nproc*)`; red TestReviewR1HarnessProbeWordsStayCommandData rc 1 | | | open |
| R23 | full-fake podman helper needs a real podman on PATH | minor | `test/e2e/ctr_test.go:654` `podmanPath, err := exec.LookPath("podman")`; red TestR3FullFakePodmanNeedsNoInstalledPodman rc 1 | | | open |
| R24 | duplicate scope prerequisite helper + dead ctrhelper mode | minor | `rg -Fn 't.Fatal("OFFLOAD_REQUIRE_SCOPE is set' internal/run/run_test.go test/e2e/execution_test.go` → 2; `ctrhelper/main.go:41` `case "id":` (0 callers) | | | open |
| R25 | `ctr --check` SIGKILLs its child `offload run`s on a signal → start records without an end (J2, K10) | minor | `internal/cli/ctr.go:189` `exec.CommandContext(ctx, self, "run"` (default Cancel = Kill) | | | open |
| R26 | `deploy` runs `check` after `switch` alone (Decision: at every deploy) | minor | `machines/nixbox/deploy:10` `[ "${1:-switch}" != switch ] \|\| "$here/check"`; red TestSF14DeployChecksNonSwitch rc 1 (test/boot/build) | | | open |
| R27 | controller scripts' ssh options diverge: deploy + tpm-bind lack BatchMode/forwarding off, install lacks accept-new | minor | `machines/nixbox/deploy:9` `ssh "offload@$ip"`; `tpm-bind:19` `ssh "offload@$ip"`; `tools/install:29` no `StrictHostKeyChecking` | | | open |
| R28 | `tools/install` leaves the per-side sha256 comparison to the reader | minor | `tools/install:4` `Prints the sha256 each side reports per file; they must match.` | | | open |
| R29 | text vs behavior: usage pool tiers, C3 worktree clause, `--no-sync` text, `code.md` `.transport-fail`, 2 stale comments, L2 mechanism prose | minor | reviewer-4 R4-F1/F2/F5/F6/F7 + R4-R1 anchors (rematched at 41fd3c0) | | | open |
| R30 | Decision `heavy interop` says `<cmd %q>`; S3 + code use POSIX single quotes | minor | `rg -Fn '<cmd %q>' .agent/spec.md` (owner ruling) | | | open |
| R31 | cli unit tests lack verification credit | major | reviewer-5 R5-F1 (4 tests) | | | open |
| R32 | run tests lack verification credit | major | reviewer-5 R5-F2 (9 tests) | | | open |
| R33 | shell lookup test lacks verification credit | minor | reviewer-5 R5-F3 | | | open |
| R34 | pooladdrs test lacks verification credit | minor | reviewer-5 R5-F4 | | | open |
| R35 | 12 tests credited by selected plants alone (no default-operator sweep of their Go files) | minor | reviewer-5 R5-F6 | | | open |
| R36 | survivors ruled by the test author alone: 13197df (13, ledger), 85e2c60 (slots MakeDirs) | minor | reviewer-5 R5-F7 + E1 | | | open |
| R37 | environment-conditional skips (no user manager / no podman) carry no deferred row + owner approval | minor | reviewer-5 E2: `requireScope`, `requirePodman`, `locks_test.go:63` (owner ruling) | | | open |
