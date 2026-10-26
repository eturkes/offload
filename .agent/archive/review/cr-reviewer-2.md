# reviewer-2 — code review @ 41fd3c0

## Verdict table

| row | check | verdict | finding ids | evidence (file:line, command) |
|-----|-------|---------|-------------|-------------------------------|
| C3 | m3 mirror | finding | R2-F1, R2-F2 | Y1–Y11: `internal/mirror/mirror.go:44–451`, caller `internal/cli/cli.go:211–240,300+`, callee `internal/run/remote.bash:12–85`: upload U/lstat/ignore rules; separate `.git` delete mirror + linked-admin filters; prune nonce/count + real-ancestor guards; download content compare + `--update`, no deletes; plain dir-merge mirror read. Targeted e2e regex `^(TestY4GitUploadDifferential\|TestY4LinkedWorktreeRunsRemotely\|TestY4GitMetadataMirrorDeletes\|TestY5Prune\|TestY6Download\|TestY7PlainMirror)$` → rc 1: upload seed-1 timed out (20 s); remaining selected tests passed. Isolated `TestY4GitUploadDifferential` rerun → rc 0, 12.754 s. Timeout cause unconfirmed. Both new reds → rc 1; details below. |
| C5 | m5 pool | pass | — | P1–P3: `internal/pool/pool.go:51–111` resolves nonempty env precedence; strict typed TOML, absent/zero slots distinct, unique names/addresses, errors preserve missing-file identity. Callers `internal/cli/cli.go:193–211`, `internal/cli/place.go:197`, `internal/cli/ctr.go:151`, `tools/pooladdrs/main.go:20` fit. `go test -count=1 ./internal/pool` → rc 0. |
| C6 | m6 probe | pass | — | R1/R2: `internal/probe/probe.go:17–92` script + exactly-four-field parser reject invalid counters/NaN/Inf; 7.5 s context + 0.5 s WaitDelay bounds/reaps ssh even with inherited output pipes. `internal/cli/place.go:135–140,214–233` fans out, joins, emits pool-file order; `internal/cli/ctr.go:180` shares probe. `go test -count=1 ./internal/probe` → rc 0 (8.212 s; deadline + local script included). |
| C8 | m8 ctr | pass | — | K1–K11: `internal/ctr/tag.go:21–80` deterministic length-framed tree hash; `ctr.go:57–197` root/definition checks, build lock inheritance + double-check, mount/podman argv; `ctr.go:204–410` setup signal cancellation, private session/scope, socket EOF + bounded record; `init.go:29–138` PID-1 guard, fd3 CLOEXEC, group signals, reaping/10 s teardown; caller `internal/cli/ctr.go:34–244` native switch, layers-only success output, nonce frames + ordered equality. `go test -count=1 ./internal/ctr` → rc 0 (3.596 s); K10 banner e2e passed. K8 e2e readiness/grace wall-time checks FAILED; isolated TERM-readiness rerun also failed (6 s, empty stdout/stderr). No proven production-contract violation from startup-time failures; coverage limit + register R2-R1 below. |
| C12 | m12 tools | pass | — | `tools/heavy:4–91` slot count normalization, ascending flock, exclusive owner text, fd closed in command, EXIT clears owner; differential/interoperability tests use this shipped script (`internal/slots/slots_test.go:18–181`), 4 targeted tests → rc 0. `tools/pooladdrs/main.go:20–30` reuses P parser; its package tests → rc 0. `tools/install:6–38` static one-build copy, atomic renames, all addresses parsed through pooladdrs, each host/file checksum printed; read only, never run. |
| C13 | m13 gate | pass | — | `.agent/gate:10–17,57–71`: child rc captured immediately + any nonzero aggregate → failure despite intentional set-u-only; extracted helper/aggregate witness positive=true→0, negative=exit23→23, aggregate rc1 (no full gate run). `.agent/tool:10–105` bootstrap rc/digest checks + exec; `--pins` → 11 rows, rc0. `.agent/pins:11–72` pipefail, lookup rc2 precedence, current/behind rc0/1; `.agent/plant:8–27` report-only mutant runner; `.agent/spec-lint` current spec rc0. All `.github/**`, `.golangci.yml`, `statix.toml` read: CI shell/pipe status, package/gate scope, static build, pins Go versions fit. Live network freshness, full gate + CI deliberately not run. |
| C14 | m14 machines | finding | R2-F3 | All `machines/nixbox/**` versioned files read; install/deploy/tpm-bind never run. `configuration.nix` list merges/imports, `idle.nix` user-slice drop-in, `performance.nix` boot/runtime settings reviewed against Decisions. Read-only SSH: deployed configuration/idle/performance sha256 = target; CPUWeight=idle, MemoryHigh=21474836480, daily upgrade 04:40, PL1=22000000, cap80 + Custom, relevant services Result=success. Initial unprivileged module digest query partly rc1 (permission); sudo readonly digest follow-up rc0. `check:23–104` collection + set comparison reviewed; six synthetic --facts cases → expected rc (fresh/71h 0; stale/failed-upgrade/missing-upgrade/failed-unit 1). No Nix evaluator installed locally; no fresh Nix eval/rebuild or complete live fact collection run. Collector reliability concern → security-boundary register R2-R2. EFI failure red below. |

## Findings

### R2-F1 — relative global excludes read from the wrong directory

- Severity: major. Target: `internal/mirror/mirror.go:226–237` @ 41fd3c0 (`path := strings.TrimSpace(string(out))`, `data, err := os.ReadFile(path)`).
- Y1/Y5: git runs with `-C <root>` and interprets relative `core.excludesFile` there; `GlobalExcludes` reads it relative to offload's invocation cwd instead. Running from a project subdirectory silently sends empty/different global excludes to the host. Remote prune can delete ignored host files, and output selection can return ignored files.
- Positive control: repository config `core.excludesFile=global-ignore`, file at root containing `*.cache\n`; `git -C <root> check-ignore saved.cache` outputs `saved.cache`. Invocation cwd = `<root>/sub`.
- Red test: `internal/mirror/cr_review_test.go:10`, `TestCRR2Y5RelativeExcludesFromSubdirectory`; committed on `wt/reviewer-2` as `2e94f4a` (production sources still 41fd3c0).
- Command (worktree): `go test -count=1 -run '^TestCRR2Y5RelativeExcludesFromSubdirectory$' ./internal/mirror` → rc 1.

```text
--- FAIL: TestCRR2Y5RelativeExcludesFromSubdirectory (0.02s)
    cr_review_test.go:22: Y5: excludes must match git -C root from any invocation subdirectory: got "", err <nil>; want "*.cache\n"
FAIL
FAIL github.com/eturkes/offload/internal/mirror 0.022s
```

- Acceptance: relative excludes resolve against the root used by git (retain absolute/tilde-expanded paths); test green.

### R2-F2 — failed prune preparation still executes CMD

- Severity: major. Target: `internal/run/remote.bash:64–65` @ 41fd3c0 (mirror's Y5 callee): `tail -z -n +3 "$__meta/upload" | LC_ALL=C sort -z >"$__meta/upload.sorted"` and `LC_ALL=C sort -z "$__meta/remote" | LC_ALL=C comm -z -23 - "$__meta/upload.sorted" >"$__meta/prune"`.
- Y5/Y10: the remote script uses `set -u` but neither checks these pipelines nor enables pipefail. A failing sort/comm (I/O/temp-space failure) leaves an invalid prune set yet CMD starts and may return success. Depending on which half fails, stale files stay or retained project files can be treated as absent from U.
- Red test: `internal/run/cr_review_test.go:10`, `TestCRR2Y10PruneSortFailurePreventsCommand`; commit `c2ca6f8` on `wt/reviewer-2`. A PATH-local sort stub emits a marker and exits 23; the test requires that marker before checking behavior.
- Command: `go test -count=1 -run '^TestCRR2Y10PruneSortFailurePreventsCommand$' ./internal/run` → rc 1 on 41fd3c0 production sources.

```text
--- FAIL: TestCRR2Y10PruneSortFailurePreventsCommand (1.83s)
    cr_review_test.go:26: Y10: failed prune preparation must stop before CMD: rc=0 ran=true stderr="injected sort failure\ninjected sort failure\n"
FAIL
FAIL github.com/eturkes/offload/internal/run 1.837s
```

- Acceptance: every failed input transformation/write used by pruning → 125 before any prune deletion or CMD; test green. Reported to reviewer-1 to avoid duplicate ownership of the callee.

### R2-F3 — EFI enumeration failure reported as successful repair

- Severity: minor. Target: `machines/nixbox/efi-entry:13` @ 41fd3c0: `stale=$("$efi" | sed -n 's/^Boot\([0-9A-F]\{4\}\)[* ] Linux Boot Manager\(\t.*\)\{0,1\}$/\1/p')`.
- Acceptance contract = script `efi-entry:2–4` + runbook: leave exactly one Linux Boot Manager entry on the new ESP, deleting old ones. `set -eu` under sh sees only sed's successful status. If efibootmgr cannot enumerate entries, the script silently treats the stale set as empty, creates another entry, and returns 0 without knowing/removing old entries. Install proceeds toward reboot despite incomplete boot-entry repair.
- Red test: `machines/nixbox/cr_review_test.go:12`, `TestCRR2EFIReadFailureStopsBeforeChanges`; commit `03324c3` on `wt/reviewer-2`. EFIBOOTMGR fixture only (no firmware access); positive control enumerates Boot0001 and observes delete-before-create. Failure fixture emits a marker and exits 23 only on enumeration, succeeding on mutation.
- Command: `go test -count=1 -run '^TestCRR2EFIReadFailureStopsBeforeChanges$' ./machines/nixbox` → rc 1 on 41fd3c0 production sources.

```text
--- FAIL: TestCRR2EFIReadFailureStopsBeforeChanges (0.09s)
    --- FAIL: TestCRR2EFIReadFailureStopsBeforeChanges/listing-fails (0.04s)
        cr_review_test.go:40: EFI discovery failed but install mutated entries: err=<nil> changes="-q -c -d /dev/fake-disk -p 1 -L Linux Boot Manager -l \\EFI\\systemd\\systemd-bootx64.efi\n" output="injected EFI listing failure\n"
FAIL
FAIL github.com/eturkes/offload/machines/nixbox 0.097s
```

- Acceptance: failed EFI enumeration returns nonzero before any boot-entry mutation; positive control + red test green.

## Register

- R2-R1 — unproven, verification surface outside this check set: `test/e2e/harness_test.go:293`, `test/e2e/ctr_test.go:450,530–532`; real-podman K8 signal readiness failed before signals were exercised (6 s, empty stdout/stderr), including isolated TERM rerun; leftover-process case exited with expected status6 but took23.414 s from launch (allowed9–20), so no proven 10 s teardown violation independent of startup delay; production cause unconfirmed, K8 real-signal behavior unverified in these runs.
- R2-R2 — unproven, security scope → MAIN: `machines/nixbox/check:31–66` uses sh pipelines/substitutions whose final command can mask collector failures (systemctl, nft, closure/PSK scan) as clean-looking facts; no further exploration under this brief.
- R2-R3 — tooling coverage: built-in LSP documentSymbol on `internal/pool/pool.go` returned `No LSP server available for file type: .go`; caller/callee coverage used explicit source search + reads instead.

## Reproduction + limits

- Worktree: `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-2`; red commits only (`2e94f4a`, `c2ca6f8`, `03324c3`), production source bytes remain 41fd3c0. Run commands above from this worktree; no full gate, CI, live pool run, install/deploy/reseal, or production edit performed.
- Selected source reads + package/targeted tests are the evidence; this is not a full-suite green claim. Retained prior accepted rulings (`ledger-implement-whole-product`, `ledger-pool-run-environment`, `ledger-maintain-security-review`); no accepted row reversed.
- Six synthetic check fixtures exercise upgrade/unit assessment, not live fact collection. Nix live option observations + matching deployed module digests confirm selected settings, not a complete module reevaluation. Full static Nix lint/eval reserved to MAIN's gate.

- Completion validation: positive-control unknown row count1/rc0; deliverable unknown count0/rc1 (grep no matches); adjudicated C rows7. `git diff --name-only 41fd3c0 HEAD` = the three red-test files alone; worktree clean; `gofmt -d` on all three = empty. Finding anchor bytes rematched against `git show 41fd3c0:<path>`.

reviewer-2-DONE-1
