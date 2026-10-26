# Review ledger — MAINTAIN: code review of the implementation (HEAD 41fd3c0)

Subject = code, tests, gates, config + docs at 41fd3c0. Check set fixed + committed before the subject's first read (law read first: `CLAUDE.md`, `.claude/rules/*`, `.claude/agents/*`, `.agent/spec.md`, `.agent/deferred.md`, prior ledgers). Row = verdict (`pass` | `finding` + R-ids | `n/a` | `not run` + reason) + evidence; `open` rows block the close. Findings table below: severity + `file:line` + proof (red test | `rg -Fn` anchor bytes), fix commit, acceptance. Teammate claims = leads; MAIN's check decides. Security defects met on any lens → MAIN alone (security lens = the `Security review` body; prior rows `ledger-maintain-security-review.md`).

Modules (rows = lens × module): m1 cli = `cmd/offload`, `internal/cli` · m2 run = `internal/run` (go + bash) · m3 mirror · m4 slots · m5 pool · m6 probe · m7 remote · m8 ctr (`ctr.go`, `init.go`, `tag.go`) · m9 ledger · m10 shell · m11 e2e harness (`test/e2e` harness, `testdata/ssh`, `testdata/ctrhelper`) · m12 tools (`tools/heavy`, `tools/install`, `tools/pooladdrs`) · m13 gate (`.agent/{gate,tool,pins,plant,spec-lint}`, `.golangci.yml`, `statix.toml`, `.github/**`) · m14 machines (`machines/nixbox/**`).

## B — baseline: full gate + each runnable `Artifacts` entry by its run command (red step / failed run = finding)

| id | check | verdict | evidence / ruling | sev | state |
|----|-------|---------|-------------------|-----|-------|
| B1 | full gate `offload run -- .agent/gate` on 41fd3c0: rc 0, per-step rc + seconds, 10 slowest tests | unknown | | | open |
| B2 | fast gate `.agent/gate fast` (+ one `[pkg…]` form) | unknown | | | open |
| B3 | GitHub: latest `Gate`, `CodeQL`, `pins` runs on `main` green; open code-scanning + Dependabot alerts | unknown | | | open |
| B4 | `offload status` (controller slots + every pool host row) | unknown | | | open |
| B5 | `offload run -- <cmd>` default placement: pool first, stderr names placement, rc passed through (0 + nonzero) | unknown | | | open |
| B6 | `offload run --local -- <cmd>` (heavy slot, owner line visible to `heavy --status`) | unknown | | | open |
| B7 | `offload run --pool` + `--host nixbox` + unknown `--host` | unknown | | | open |
| B8 | `offload run --gpu` (no NVIDIA pool host) | unknown | | | open |
| B9 | `offload run --no-sync` | unknown | | | open |
| B10 | `offload run --get PATH` (ignored path fetched back) | unknown | | | open |
| B11 | run from a subdir + from a linked git worktree + from a non-git dir | unknown | | | open |
| B12 | `offload ctr` fixture = a non-git scratch dir outside the repo (sync root = cwd) holding a minimal `.devcontainer/Containerfile`: `offload ctr -- <cmd>` with `OFFLOAD_CTR` set + unset, `--layers`, `--check`, and the consumer form `offload run -- offload ctr -- <cmd>` | unknown | | | open |
| B13 | run log: start + end records for B5–B11 runs in `~/.local/state/offload/runs.jsonl` (README `See past runs` recipe) | unknown | | | open |
| B14 | `heavy` (`heavy --status`, `heavy <cmd>`) beside offload's local slots | unknown | | | open |
| B15 | `tools/install` (at close when a shipped change landed) + `offload ctr --check` after it | unknown | | | open |
| B16 | `.agent/pins` | unknown | | | open |
| B17 | `pool.example.toml` parses (`OFFLOAD_CONFIG=pool.example.toml offload status`) | unknown | | | open |
| B18 | `machines/nixbox/deploy` (+ its `check`) | unknown | | | open |
| B19 | `machines/nixbox/{install,tpm-bind}`, installer ISO, firmware: driven where non-destructive (usage/arg paths, evaluation); destructive runs named `not run` | unknown | | | open |

## C — Correctness (`Engineering` code diff lens): each function with its callers + callees; guards + anchors; classic pitfalls of the language (Go: error shadowing, nil maps, defer in loops, goroutine/ctx/fd leaks, races, integer conversion, close errors; bash: quoting, `set -e` traps, subshells, pipefail; Nix: option semantics); wrappers route to the wrapped instance; defaults intended; test setup ↔ teardown symmetric

| id | check | verdict | evidence / ruling | sev | state |
|----|-------|---------|-------------------|-----|-------|
| C1 | m1 cli | unknown | | | open |
| C2 | m2 run | unknown | | | open |
| C3 | m3 mirror | unknown | | | open |
| C4 | m4 slots | unknown | | | open |
| C5 | m5 pool | unknown | | | open |
| C6 | m6 probe | unknown | | | open |
| C7 | m7 remote | unknown | | | open |
| C8 | m8 ctr | unknown | | | open |
| C9 | m9 ledger | unknown | | | open |
| C10 | m10 shell | unknown | | | open |
| C11 | m11 e2e harness | unknown | | | open |
| C12 | m12 tools | unknown | | | open |
| C13 | m13 gate | unknown | | | open |
| C14 | m14 machines | unknown | | | open |

## U — Cleanup: existing helpers reused; derivable state computed, never duplicated; near-duplicates merged; dead code removed; independent steps parallel; startup + hot paths lean; closures capture only what they use

| id | check | verdict | evidence / ruling | sev | state |
|----|-------|---------|-------------------|-----|-------|
| U1 | m1 cli | unknown | | | open |
| U2 | m2 run | unknown | | | open |
| U3 | m3 mirror | unknown | | | open |
| U4 | m4 slots | unknown | | | open |
| U5 | m5 pool | unknown | | | open |
| U6 | m6 probe | unknown | | | open |
| U7 | m7 remote | unknown | | | open |
| U8 | m8 ctr | unknown | | | open |
| U9 | m9 ledger | unknown | | | open |
| U10 | m10 shell | unknown | | | open |
| U11 | m11 e2e harness + test helpers (duplicated helpers across `*_test.go`) | unknown | | | open |
| U12 | m12 tools | unknown | | | open |
| U13 | m13 gate | unknown | | | open |
| U14 | m14 machines | unknown | | | open |

## A — Altitude: a special case on shared infrastructure = a fix due in the mechanism

| id | check | verdict | evidence / ruling | sev | state |
|----|-------|---------|-------------------|-----|-------|
| A1 | m1 cli | unknown | | | open |
| A2 | m2 run | unknown | | | open |
| A3 | m3 mirror | unknown | | | open |
| A4 | m4 slots | unknown | | | open |
| A5 | m5 pool | unknown | | | open |
| A6 | m6 probe | unknown | | | open |
| A7 | m7 remote | unknown | | | open |
| A8 | m8 ctr | unknown | | | open |
| A9 | m9 ledger | unknown | | | open |
| A10 | m10 shell | unknown | | | open |
| A11 | m11 e2e harness | unknown | | | open |
| A12 | m12 tools | unknown | | | open |
| A13 | m13 gate | unknown | | | open |
| A14 | m14 machines | unknown | | | open |

## G — claim soundness + guarantee-vs-claim gaps (each claim vs observed behavior)

| id | check | verdict | evidence / ruling | sev | state |
|----|-------|---------|-------------------|-----|-------|
| G1 | `README.md` (every promise, example + recipe) | unknown | | | open |
| G2 | `--help` / usage + user-facing error texts (`offload`, `offload run`, `offload ctr`, `heavy`) | unknown | | | open |
| G3a | `.agent/contract.md` family C predicates vs behavior (each id) | unknown | | | open |
| G3b | `.agent/contract.md` family P predicates vs behavior (each id) | unknown | | | open |
| G3c | `.agent/contract.md` family R predicates vs behavior (each id) | unknown | | | open |
| G3d | `.agent/contract.md` family S predicates vs behavior (each id) | unknown | | | open |
| G3e | `.agent/contract.md` family L predicates vs behavior (each id) | unknown | | | open |
| G3f | `.agent/contract.md` family Y predicates vs behavior (each id) | unknown | | | open |
| G3g | `.agent/contract.md` family X predicates vs behavior (each id) | unknown | | | open |
| G3h | `.agent/contract.md` family J predicates vs behavior (each id) | unknown | | | open |
| G3i | `.agent/contract.md` family K predicates vs behavior (each id) | unknown | | | open |
| G4 | `.agent/spec.md` `Artifacts` (paths, run commands, live facts) | unknown | | | open |
| G5 | `machines/nixbox/README.md` runbook vs scripts + config | unknown | | | open |
| G6 | `pool.example.toml` comments vs P-predicates | unknown | | | open |
| G7 | facts stated in `.claude/rules/*` (timings, recipes, mechanisms) vs code | unknown | | | open |
| G8 | code comments m1 cli | unknown | | | open |
| G9 | code comments m2 run | unknown | | | open |
| G10 | code comments m3 mirror | unknown | | | open |
| G11 | code comments m4 slots | unknown | | | open |
| G12 | code comments m5 pool + m6 probe | unknown | | | open |
| G13 | code comments m7 remote + m10 shell | unknown | | | open |
| G14 | code comments m8 ctr | unknown | | | open |
| G15 | code comments m9 ledger | unknown | | | open |
| G16 | code comments m11 e2e harness + test comments citing ids | unknown | | | open |
| G17 | comments m12 tools | unknown | | | open |
| G18 | comments m13 gate | unknown | | | open |
| G19 | comments m14 machines (nix + scripts, `*.reviewed`) | unknown | | | open |

## V — verification integrity (`Engineering`): each test seen red on its unfixed revision (revision + command in a commit body) or mutation-credited (default operators, scope = files it imports/calls, every survivor ruled); kills via observable behavior; no test detection / fixture-value return / contract-unowned expected tables in the implementation; skipped, xfailed, deleted or demoted cases carry a `deferred.md` row + owner approval; tests cite contract ids (`code.md`)

| id | check | verdict | evidence / ruling | sev | state |
|----|-------|---------|-------------------|-----|-------|
| V1 | m1 cli (`internal/cli` + e2e `cli_test.go`) | unknown | | | open |
| V2 | m2 run (unit + e2e execution/close/failures/locks) | unknown | | | open |
| V3 | m3 mirror (unit + mutation + e2e sync) | unknown | | | open |
| V4 | m4 slots | unknown | | | open |
| V5 | m5 pool | unknown | | | open |
| V6 | m6 probe | unknown | | | open |
| V7 | m7 remote | unknown | | | open |
| V8 | m8 ctr (unit + real-podman e2e) | unknown | | | open |
| V9 | m9 ledger (unit + e2e `ledger_test.go`) | unknown | | | open |
| V10 | m10 shell | unknown | | | open |
| V11 | m11 e2e harness self-tests (fake ssh X5 refusal, `-n`, markers) | unknown | | | open |
| V12 | m12 tools (`pooladdrs` test; `heavy` + `install` coverage) | unknown | | | open |
| V13 | m13 gate: every step fires on its `gate-checks.md` seed (replanted) | unknown | | | open |
| V14 | m14 machines: `check` fires on every `machines.md` planted fact | unknown | | | open |

## SF — spec fit: behavior vs `Intent` + `Decisions`

| id | check | verdict | evidence / ruling | sev | state |
|----|-------|---------|-------------------|-----|-------|
| SF1 | m1 cli: trigger + flags, placement order (pool → local heavy → wait both, one waiting message), load ranking, `--host/--gpu/--no-sync` ⇒ `--pool`, stderr placement | unknown | | | open |
| SF2 | m2 run: remote run sequence, rc passthrough, kill-on-signal incl. SIGKILL (watcher), per-(host, workdir) serialization, local runs never lock | unknown | | | open |
| SF3 | m3 mirror: sync root, workdir naming, git-exact up/down sets, deletion pass, download `--update` no deletes, `.git` never returns, non-git mirror, linked worktree layout | unknown | | | open |
| SF4 | m4 slots: heavy interop (paths, flock, owner line, emptied at exit), pool slot locks, admission vs loadavg | unknown | | | open |
| SF5 | m5 pool: TOML `[[host]]` (`name`, `address`, `slots` default 2), `OFFLOAD_CONFIG` | unknown | | | open |
| SF6 | m6 probe: load = rank only (1-min / nproc), GPU inventory | unknown | | | open |
| SF7 | m7 remote: BatchMode + accept-new, ssh_config alias addresses | unknown | | | open |
| SF8 | m8 ctr: K-predicates, switch `OFFLOAD_CTR`, `.devcontainer/Containerfile`, tag = content hash, host builds itself, `--check` = `RootFS.Layers` equality, LAN lanes only | unknown | | | open |
| SF9 | m9 ledger: every `run` logged (start once a slot is held + end), fields per `Decisions`, log failure never touches the run | unknown | | | open |
| SF10 | m10 shell: lookup = bash's (126/127) | unknown | | | open |
| SF11 | m11 e2e: harness models the pool hosts `Decisions` describe (git + coreutils, ssh options) | unknown | | | open |
| SF12 | m12 tools: install = one static build → controller + every pool host, sha256 per side; heavy separate script deployed alongside | unknown | | | open |
| SF13 | m13 gate + CI: `Gate compute` + `CI gate` rulings, Gate CI on Ubuntu 26.04, pins weekly, Dependabot scope | unknown | | | open |
| SF14 | m14 machines: idle task slice, accounts, LAN-only nftables, key-only ssh, LUKS/TPM/Secure Boot, compute posture, carried tools, security posture, Wi-Fi secret path, `check` at every deploy | unknown | | | open |
| SF15 | `Decisions` binding no code module (trading + Codexify, instruction tiers, nixbox role, gate image placement, pool run environment, run memory) vs machines config, README, rules: each fits or is ruled n/a by name | unknown | | | open |

## M — `CLAUDE.md` + `.claude/rules/` conformance: each clause that binds a tracked file (tool-decidable → its gate check runs + fires on a planted bad input; none → finding, fixed by wiring the check)

| id | clause → binds | verdict | evidence / ruling | sev | state |
|----|----------------|---------|-------------------|-----|-------|
| M1 | Execution `Git`: `.gitignore` current (a full gate + runs leave no untracked, unignored file) | unknown | | | open |
| M2 | Execution gate command: full chain recorded in `.claude/rules/`, prints each step's wall-clock + the suite's slowest tests | unknown | | | open |
| M3 | Execution fast gate recorded beside it (changed steps + tests, parallel + cached) | unknown | | | open |
| M4 | Execution: a step dominating the full gate → a speed-up unit in `Tasks` or `deferred.md` | unknown | | | open |
| M5 | Execution `CI gate` → keyed ruling (`gate.md`) current vs the repo's CI | unknown | | | open |
| M6 | Execution task close: no temporary/stale artifacts or dangling symlinks in the tree (candidate at seed: `.ruff_cache/` in the tree) | unknown | | | open |
| M7 | Execution UI design + charts | unknown | | | open |
| M8 | Authoring: agent-optimized text in every non-human-facing artifact (comments, contract, rules, internal docs, runbook) | unknown | | | open |
| M9 | Authoring: rules/facts omit provenance (dates, discovery events, origin stories) | unknown | | | open |
| M10 | Authoring: future-facing text states actions positively; prohibitions name the thing | unknown | | | open |
| M11 | Authoring: no strict text caps in guidance or checks | unknown | | | open |
| M12 | Authoring: instruction tier routing (user `CLAUDE.md` / `CLAUDE.local.md` / agents / `CLAUDE.md` / rules / spec) + `portability.md` ruling | unknown | | | open |
| M13 | Authoring: `.claude/rules/` = live law + why, no origin stories; multi-step recipe → committed script; bare tier small; area-bound → `paths:` | unknown | | | open |
| M14 | Authoring: rule + agent frontmatter quoting + parse | unknown | | | open |
| M15 | Authoring: `@path` imports only for files CC leaves unloaded; plain paths for auto-loaded ones | unknown | | | open |
| M16 | Authoring `Human-facing` (README, `--help`, error texts): ASD-STE100 register; filler words (`simply`, `robust`, `seamlessly`, `leverage`) absent | unknown | | | open |
| M17 | Engineering: comments spend on the why; code states the what | unknown | | | open |
| M18 | Engineering: deterministic checks own every tool-decidable rule; each purpose-built check ships its failing input + firing in a `paths:` rule | unknown | | | open |
| M19 | Engineering: gate backing a durable claim reruns from committed state; scratch-local validators recorded + scheduled | unknown | | | open |
| M20 | Engineering: assurance tiers + contracts per unit (contract states tiers where units bind) | unknown | | | open |
| M21 | Session flow: `spec.md` shape (5 sections, `Tasks` pointer last, `Phase` = state) via `spec-lint`; every line binds current/future work | unknown | | | open |
| M22 | Session flow: `deferred.md` = one entry + acceptance check each | unknown | | | open |
| M23 | Session flow: `spec-lint` in the gate command; archive holds finished work | unknown | | | open |
| M24 | Session flow: `CLAUDE.md` + `.claude/agents/` = current template copies | unknown | | | open |
| M25 | `code.md` 1: contract = law; tests cite predicate ids | unknown | | | open |
| M26 | `code.md` 2: watch/remote/boot bash-only, GNU tools, no controller paths; every variable `__…`; `<&0` + `trap - INT QUIT` kept | unknown | | | open |
| M27 | `code.md` 3: host deletion = `prune_one` + worktree `rm -f .git`; rsync `--delete` only inside `.git` / `<meta>/repo` | unknown | | | open |
| M28 | `code.md` 4: init PID-1 guard, handlers (no `signal.Ignore`), podman `Setsid` + scope + no `INVOCATION_ID`, nonblocking control end, `CGO_ENABLED=0` on every shipping build | unknown | | | open |
| M29 | `code.md` 5: shipped change → `tools/install`; `--check` reds a differing binary | unknown | | | open |
| M30 | `code.md` 6: real-podman e2e isolation (vfs storage, `FROM scratch` + ctrhelper static, prebuild fails never skips; CI `OFFLOAD_REQUIRE_PODMAN=1`) | unknown | | | open |
| M31 | `code.md` 7: run log ownership (`runLogged`), J4, `signalStatus` ordering, e2e `XDG_STATE_HOME`, `runPair` validator | unknown | | | open |
| M32 | `code.md` 8: every ssh/rsync connection takes `remote.SSHOptions`; fake ssh refuses otherwise (rc 254), honors `-n` | unknown | | | open |
| M33 | `code.md` 9–10: fake-host marker semantics + mutation recipe (`.agent/plant`) match the harness + script | unknown | | | open |
| M34 | `gate.md` 1–2: step list + fast gate = `.agent/gate`; CI runs the same command | unknown | | | open |
| M35 | `gate.md` 3: CI-gate ruling + Gate CI dispatch procedure + timing claims | unknown | | | open |
| M36 | `gate.md` 4: tool pins (sha256 / sumdb / commit id), `go.mod` go line = CI setup-go + `GOTOOLCHAIN=local` | unknown | | | open |
| M37 | `gate.md` 5: suppressions name rule + reason (`nolintlint` fires on a bare one) | unknown | | | open |
| M38 | `gate.md` 6: CodeQL + Dependabot scope + `pins` coverage of hand-moved pins | unknown | | | open |
| M39 | `gate-checks.md`: every recorded seed still fires (= V13) + every gate step has a seed | unknown | | | open |
| M40 | `machines.md` 1 + 4: `wifi.env` + `installer_host_key` gitignored, never committed; secrets referenced by string path, never a Nix path or inline | unknown | | | open |
| M41 | `machines.md` 2–3 + 6–7: deploy/boot-chain/recovery-key/Codexify/reboot facts vs scripts + runbook | unknown | | | open |
| M42 | `machines.md` 5: `check` planted-fact list (= V14) | unknown | | | open |
| M43 | `portability.md`: ruling keyed + consistent with `CLAUDE.md` `Authoring` | unknown | | | open |

## Findings

| id | finding | sev | `file:line` + proof | fix | acceptance | state |
|----|---------|-----|---------------------|-----|------------|-------|
