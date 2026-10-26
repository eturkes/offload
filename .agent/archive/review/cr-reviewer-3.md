# reviewer-3 — code review @ 41fd3c0

## Verdict table

| row | check | verdict | finding ids | evidence (file:line, command) |
|-----|-------|---------|-------------|-------------------------------|
| U1 | m1 cli | pass | — | `internal/cli/cli.go:185–404`, `place.go:57–237`, `ctr.go:73–262`: transport resolution + command lookup reused; probe/check hosts parallel; placement vs upload lists intentionally revalidate across a slot wait. Full source read + `rg -n` callers. |
| U2 | m2 run | pass | — | `internal/run/run.go:54–147`, `watch.bash:23–118`, `remote.bash:16–82`: one supervisor + watcher; prune/download reuse `in_tree`/`git_list`; run body/hash computed once. |
| U3 | m3 mirror | pass | — | `internal/mirror/mirror.go:43–90,168–220,263–399`: common discovery, memoized ancestor checks, root-name derivation + rsync flag builders. Git file + admin uploads share a workdir lock; serialization intentional. | 
| U4 | m4 slots | pass | — | `internal/slots/slots.go:82–239`: directory checks reused across slots, workdir/build/log locks; blocking/nonblocking lock APIs differ deliberately (fd kept across polls). | 
| U5 | m5 pool | pass | — | `internal/pool/pool.go:16–110`: one parser + validation loop; raw/normalized Host forms encode absent-vs-explicit-zero, not redundant state. `rg -n 'pool.Load' internal tools` → CLI + installer consumers. | 
| U6 | m6 probe | pass | — | `internal/probe/probe.go:16–83`: one SSH probe + parser; capacity fields direct, busy ratio derived; caller-level host parallelism. | 
| U7 | m7 remote | pass | — | `internal/remote/ssh.go:20–108`: shared SSH option source, quoting + command lifecycle helper; no per-host duplicate. `rg -n 'remote.(Command\|Quote\|Join)' internal` callers traced. |
| U8 | m8 ctr | pass | — | `internal/ctr/ctr.go:57–195,205–406`, `init.go:28–135`, `tag.go:22–79`: shared discovery/lock/lookup; derived tags + per-root HOME; no zero-reference function in source-token scan. Tag's whole-context buffering noted in Register, not a proved workload bottleneck. | 
| U9 | m9 ledger | pass | — | `internal/ledger/ledger.go:100–313`: head embedded in tail; version once-cached, phase timers + log append centralized; slot lock machinery reused. |
| U10 | m10 shell | pass | — | `internal/shell/look.go:23–72`: one lookup implementation for native run/ctr/init; tiny status/description split has no evidenced performance issue. `rg -n 'shell.Look' internal` → two adapters. |
| U11 | m11 e2e harness + test helpers (duplicated helpers across `*_test.go`) | finding | R3-F1 | `internal/run/run_test.go:128–152` + `test/e2e/execution_test.go:134–151`: two copies of the scope-availability probe + fail/skip policy. Other helpers: sandbox + strict fake reuse; trivial cross-package `env`/`must` copies below blocking bar. |
| U12 | m12 tools | pass | — | `tools/heavy:16–89`, `tools/install:10–34`, `tools/pooladdrs/main.go:20–30`: common holder/trap helpers, build once then install both files, addresses reuse pool parser. Cross-language heavy parity required by Decision; not removable duplication. |
| U13 | m13 gate | pass | — | `.agent/gate:9–66`, `.agent/tool:14–91`, `.agent/pins:11–63`: timing/rc runner + fetch/cache/spec helpers reused; pins read tool metadata instead of a second version table. Independent steps serial; no measured critical-path gain established (short-gate Decision retained). All scoped config/workflows read. | 
| U14 | m14 machines | pass | — | `machines/nixbox/{host.nix,network.nix,configuration.nix,installer.nix,idle.nix}`, `check:24–88`: host facts/network module + compare helper shared; generated hardware file remains generated; no executable-code duplication past small setup prefixes. All tracked machine files read; no live mutations. |
| A1 | m1 cli | pass | — | `internal/cli/cli.go:179–195,296–359`, `ctr.go:165–246`: run-log wrapper covers exits; ctr check invokes ordinary run machinery; no host-specific scheduling branch. |
| A2 | m2 run | pass | — | `internal/run/run.go:54–147`, `watch.bash:62–118`: teardown owned by shared watcher, not callers; worktree/plain/nosync branches encode Y4/Y7/Y9. |
| A3 | m3 mirror | pass | — | `internal/mirror/mirror.go:43–90,321–332`: Locate shared with ctr; checksum policy applies to project transfers, documented fast-path exception for git objects/ignored gets. No project-name branch. | 
| A4 | m4 slots | pass | — | `internal/slots/slots.go:82–108,203–239`: S6 enforced in lock mechanism; heavy/pool directory constructors carry policy without forked acquisition. | 
| A5 | m5 pool | pass | — | `internal/pool/pool.go:50–110`: one P1/P2 mechanism; `tools/pooladdrs/main.go:21` consumes it instead of a second TOML parser. | 
| A6 | m6 probe | pass | — | `internal/probe/probe.go:62–83`: R1 deadline exception belongs here (not remote.Command's 30 s grace); CLI status/place/check share Probe. | 
| A7 | m7 remote | pass | — | `internal/remote/ssh.go:20–76`: transport options + multiplex fallback centralized; targets carry data, no known-host special cases. |
| A8 | m8 ctr | pass | — | `internal/ctr/ctr.go:298–341`, `init.go:28–85`: podman-specific session/scope + control-channel handling encodes K8; all real runs use this mechanism. Prior ledger R19–R22 honored. | 
| A9 | m9 ledger | pass | — | `internal/ledger/ledger.go:213–244,248–313`, `internal/cli/cli.go:155–179`: failure isolation + end emission live at wrappers; no exit-path-specific log call. |
| A10 | m10 shell | pass | — | `internal/shell/look.go:23–72`, `internal/cli/cli.go:360`, `internal/ctr/init.go:90`: bash-compatible resolution correctly lives in shared shell package, not wrapper-specific patches. |
| A11 | m11 e2e harness | pass | — | `test/e2e/harness_test.go:86–159,237–330`, `ctr_test.go:31–153`, `testdata/ssh:3–70`: process/environment/storage lifecycle shared; fake-host markers are explicit fixture mechanisms. Full-fake Podman dependency → R3-R1 (verification register, outside fixed lenses). |
| A12 | m12 tools | pass | — | `tools/install:20–34`, `tools/pooladdrs/main.go:20–30`: every pool spelling uses the same P1/P2 parser; installs operate on supplied/parsed addresses, no nixbox-specific branch. |
| A13 | m13 gate | pass | — | `.agent/gate:30–54`, `.agent/tool:45–64`, `.github/workflows/gate.yml:55`: CI uses common gate; generated-Nix exclusion + actionlint runner-label override express documented tool boundaries, not hidden project exceptions. | 
| A14 | m14 machines | pass | — | `machines/nixbox/configuration.nix:146–197,230–281`, `idle.nix:21–28`, `secureboot.nix:24–33`: host-level compatibility (nix-ld/envfs/CA/fonts), slice priority + fwupd quirk handled in their shared host mechanisms, not per-command wrappers. |

## Findings

### R3-F1 — duplicate scope-availability prerequisite

- Severity: **minor**. Row: U11. Subject: `41fd3c016caf6ed2e7cbace6f14e7f316d57ac32`.
- Sites: `internal/run/run_test.go:128–152` (`requireScope` + `scopeWorks`) and `test/e2e/execution_test.go:134–151` (`requireScope`).
- Law: `CLAUDE.md` Engineering → deduplicate; Cleanup → existing helpers reused, near-duplicate blocks merged.
- Defect: one test prerequisite has two independently editable implementations: find the same manager socket, execute the same probe, interpret the same required-mode variable, then emit the same failure/skip. This is gate-coverage policy, not independent behavioral oracles. A change to scope prerequisites must currently be carried in both packages; source inspection finds two probe implementations + two required/optional policies where one suffices.
- Fix shape: move the current test-only `requireScope`/`scopeWorks` mechanism into `internal/testutil.RequireScope(t)` (name illustrative); both packages call it. Delete their local copies. Keep this helper independent of production `ctr.scopeManager`/watcher code so a production regression cannot redefine the prerequisite. Preserve optional skip vs `OFFLOAD_REQUIRE_SCOPE` failure and current probe behavior.
- Text proof (both files unchanged from subject; `git show 41fd3c0:<file>` rematched):

```text
rg -Fn 'c := exec.Command("systemd-run", "--user", "--scope", "--quiet", "--collect", "--expand-environment=no", "--", "true")' <worktree>/internal/run/run_test.go <worktree>/test/e2e/execution_test.go
internal/run/run_test.go:146: c := exec.Command("systemd-run", "--user", "--scope", "--quiet", "--collect", "--expand-environment=no", "--", "true")
test/e2e/execution_test.go:140: c := exec.Command("systemd-run", "--user", "--scope", "--quiet", "--collect", "--expand-environment=no", "--", "true")
rc=0; matches=2

rg -Fn 't.Fatal("OFFLOAD_REQUIRE_SCOPE is set, but systemd-run --user --scope fails")' <same two files>
internal/run/run_test.go:134: t.Fatal("OFFLOAD_REQUIRE_SCOPE is set, but systemd-run --user --scope fails")
test/e2e/execution_test.go:148: t.Fatal("OFFLOAD_REQUIRE_SCOPE is set, but systemd-run --user --scope fails")
rc=0; matches=2
```

- Acceptance: one test-only implementation of that probe + fail/skip policy; both package call sites retained; scope-dependent tests retain their predicates. No performance claim or test-case removal proposed.

## Register

### R3-R1 — fully fake Podman tests require real Podman (proved; verification outside fixed lenses)

- Sites: `test/e2e/reviewer4_ctr_test.go:17–20` → `test/e2e/ctr_test.go:652–660`.
- `reviewFakePodman` installs a catch-all rejection, so no command can fall through to real Podman; nevertheless `fakePodman` unconditionally resolves real Podman before writing the fake. This rejects the fixture on a machine without Podman, including the existing `TestK7ReviewCommandStatusWithoutPodman`, `TestK6ReviewStreamsWithoutPodman`, `TestK8ReviewControlSignalWithoutPodman` and `TestK5ReviewBuildKeepsLockWithoutPodman` paths.
- Exact anchors: `fakePodman(t, s, cases+` at `reviewer4_ctr_test.go:19`; `podmanPath, err := exec.LookPath("podman")` at `ctr_test.go:654` (`:408` is a separate real-Podman test, not the helper).
- Red witness: `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-3/test/e2e/reviewer3_harness_test.go:10`, `TestR3FullFakePodmanNeedsNoInstalledPodman`; commit `da65a69e38b55305c7f912f99d9c92b80ae4150e` atop subject, implementation unchanged.
- Command actually run:

```sh
(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-3 && go test -count=1 -run '^TestR3FullFakePodmanNeedsNoInstalledPodman$' ./test/e2e/)
```

```text
--- FAIL: TestR3FullFakePodmanNeedsNoInstalledPodman (0.01s)
    reviewer3_harness_test.go:13: exec: "podman": executable file not found in $PATH
FAIL
FAIL github.com/eturkes/offload/test/e2e 7.491s
FAIL
rc=1
```

- Fix direction for verification owner: separate full-fake vs real-fallthrough construction; resolve Podman only for a real fallback. Preserve the exhaustive catch-all. No fix made.

### R3-R2 — unreferenced helper mode (below blocking bar)

`test/e2e/testdata/ctrhelper/main.go:41–42`: mode `id` has no literal caller in tracked e2e code. Positive control `rg -n 'case "id":' <worktree>/test/e2e/testdata/ctrhelper/main.go` → exactly one definition, rc 0; `rg -n '("ctrhelper",[[:space:]]*"id"|ctrhelper id)' <worktree>/test/e2e` → zero, rc 1. No dynamic use found while reading the helper + caller inventory. Remove the two-line branch if touched; not enough legibility/performance benefit alone to block.

### R3-R3 — whole-context tag buffering (unproven performance issue)

`internal/ctr/tag.go:29–76` keeps every file's content until the final sorted hashing loop; potential alternative = sort entry paths, then stream each length-framed file into the hash. No workload measurement establishes a material cost, and changing traversal/encoding carries tag-compatibility concerns. Left pass; no performance claim.

### R3-R4 — security-owner handoff (unassessed)

`machines/nixbox/deploy:8–9`, `tpm-bind:19`: bare SSH/rsync calls lack the explicit forwarding controls used in the other controller scripts; security assessment left to MAIN, no further exploration.

## Coverage + checks

- Fixed set: U1–U14 + A1–A14, 28/28 adjudicated; 27 pass, one finding (minor). Full executable source of the 14 modules read. U11: all named test-helper definitions inventoried across every tracked `*_test.go`, then helper bodies read; test-case coverage/verification outside these lenses, not a claim that every test assertion was audited.
- Subject code stayed read-only. Only worktree change = the committed red witness above. Deliverable = validator input; no other primary-tree writes.
- Reads include `.agent/{spec,contract}.md`, all `.claude/rules/*.md`, README, CLAUDE.md and all three prior `ledger-*.md` files; no accepted ruling contested.
- Function-name occurrence scan over all Go source under `cmd`, `internal`, `test/e2e`, `tools`: no non-test/non-main declaration with only its definition token. This is lexical evidence, **not** whole-program reachability proof. No deadcode tool run.
- LSP `documentSymbol` on `internal/cli/cli.go` → unavailable (`No LSP server available for file type: .go`); source reads + explicit `rg` used instead.
- Test run: the single red witness above (failed as intended). Full gate, full suites, mutation suite, live pool calls, live nixbox checks and deployment: **not run** (outside reviewer allowance). No claim of gate-green.
- Validator: seed positive control 28 unknown rows, rc 0; final fixed-row validator → `0` (grep rc 1 = zero matches). No search cap or truncated source output used for a finding.

reviewer-3-DONE-1
