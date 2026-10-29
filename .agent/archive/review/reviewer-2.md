# reviewer-2 verdicts — closing diff fed6498..cac258b

Check set B1–B6 frozen before target reads. Target = committed fed6498..cac258b (worktree wt/reviewer-2); MAIN's in-flight spec excluded. Verdict ∈ {pass, finding + ids}; 6/6 rows adjudicated, 0 pending. Detail → finding-id sections. Real-host evidence = commit trace only; never live SSH. Tests/caches use private paths.

| id | check | verdict | evidence |
|----|-------|---------|----------|
| B1 | README.md claims vs actual behavior (run the binary where cheap) | finding F1, F3 | README local-edit guarantee exceeds `--update`; fresh `status` invents holders. Four isolated red tests run at cac258b; details below. |
| B2 | commit-body measurements + red-first claims (fed6498..HEAD) reproducible or traceable | finding F5 | Complete commit range read; e2e red/green logs, archived tester tag/report, heavy digest + 14-case differential, quoting test, gate logs, copy totals trace claims. Universal planted-input claim lacks build-step record. Real-host timings/deploy/trading checks not re-run: needs real host. |
| B3 | .claude/rules/gate-checks.md firing records vs actual check behavior | finding F5 | Re-ran fmt/lint/shell/actions/secrets/nix/spec seeds → rc1; zizmor → rc14. Vuln output + bad-digest seed traceable; tidy/test recipes inspected; build record missing. Nix AST equality not re-run: needs real host. |
| B4 | .agent/contract.md statements vs code | finding F2, F3 | P2 explicit zero slots accepted as default; R2 missing slot directories count as held. Source/contract cross-check complete; selected baseline coverage green (limits below). |
| B5 | .agent/spec.md Artifacts commands + Decisions statements vs reality | finding F4 | Installed artifact = static ELF, `--help` rc0; CLI/sync decisions match inspected code except `address = any ssh destination` exceeds P2/implementation. Trading/host state traceable to commit + general-1 report; not re-run: needs real host. |
| B6 | machines/nixbox README/idle.nix claims vs config | pass | `idle.nix`: uid 1001, linger, user-1001 slice, CPUWeight=idle, drop-in; configuration imports it. Measurements match f44db3a (1532/1789 solo, 1877 idle rival, 3847 normal rival). Live hardware/cgroup/deploy claims unconfirmed here; not re-run: needs real host. |

## F1 — major — README promises edit preservation that `--update` does not provide

- Target anchor: `README.md:93` = `A file that you changed here during the run is not replaced.` (`rg -Fn` matched).
- Why wrong: a controller file edited after upload is overwritten when the remote output has a newer mtime. Y6 explicitly promises only preservation of newer local regular files; symlinks/type changes also take the remote version. Published unconditional edit-preservation promise can lose work.
- Red: `test/e2e/claims_review_test.go:12`, `TestReadmeLocalEditDuringRunSurvivesDownload`. Starts real binary through fake SSH, waits until upload finished/CMD started, edits controller `tracked`, lets CMD replace remote `tracked` with a future-mtime result, then downloads.
- Command (cwd reviewer worktree, private GOCACHE/TMPDIR): `go test ./test/e2e -run '^TestReadmeLocalEditDuringRunSurvivesDownload$' -count=1`.
- Observed in combined red run, rc 1: `README local-edit guarantee: got "remote-result", want "local edit during run\\n"`.
- Acceptance: narrow README to Y6's actual mtime/type semantics; code can retain the accepted contract. Test = published-claim counterexample, not a request to replace Y6 with a stronger conflict-detection design.

## F2 — minor — explicit zero slots silently becomes two

- Target anchor: `.agent/contract.md:18` = `` `slots` (int ≥ 1, default 2) `` (`rg -Fn` matched).
- Implementation: `internal/pool/pool.go:92` accepts zero as the absent-field default; negative values alone reject.
- Why wrong: explicit `slots = 0` violates P2 but is silently admitted with two slots. A present invalid pool must fail before work (P2/C3).
- Red: `internal/pool/claims_review_test.go:8`, `TestP2ExplicitZeroSlotsRejected`; command `go test ./internal/pool -run '^TestP2ExplicitZeroSlotsRejected$' -count=1` → rc 1: `P2 explicit slots=0: err=<nil>, want slots validation error`.
- Acceptance: distinguish absent slots (default 2) from explicit zero (validation error); preserve positive values.

## F3 — minor — fresh `status` reports all nonexistent locks as held

- Target anchor: `.agent/contract.md:24` = `A slot counts as held by its lock alone` (`rg -Fn` matched). README status section promises slots in use.
- Implementation: `internal/slots/slots.go:113` treats every `tryLock` error as occupancy; nonexistent parent directory → ENOENT → appended empty holder.
- Red: `test/e2e/claims_review_test.go:24`, `TestR2FreshStatusHasNoHolders`; command `go test ./test/e2e -run '^TestR2FreshStatusHasNoHolders$' -count=1`.
- Observed in combined red run, rc 1: fresh private runtime shows `local heavy slots: 2/2 held`, blank `slot 1:`/`slot 2:` rows, and `alpha up 2/2` despite no locks.
- Acceptance: absent lock directories/files count as unheld; actual locked slots still count, stale owner text still ignored. Surface other I/O errors distinctly rather than inventing holders.

## F4 — minor — spec accepts every SSH destination; loader deliberately does not

- Target anchor: `.agent/spec.md:23` = `` `address` = any ssh destination `` (`rg -Fn` matched).
- Why wrong: raw IPv6 is an SSH destination but P2/README/loader require an SSH config alias. Spec Decisions outrank the contract; leaving the prototype's unrestricted wording makes the acceptance surfaces disagree.
- Positive control, no network connection: `ssh -G -F /dev/null offload@::1` → rc0, `user offload`, `hostname ::1`.
- Red: `internal/pool/claims_review_test.go:16`, `TestSpecAnySSHDestinationAcceptsIPv6`; command `go test ./internal/pool -run '^TestSpecAnySSHDestinationAcceptsIPv6$' -count=1`.
- Observed in combined pool run, rc1: `spec any-SSH-destination guarantee rejects raw IPv6: ... host alpha: address "offload@::1" must be [user@]host or an ssh_config alias`.
- Acceptance: align spec wording with P2/README (host or alias, alias for IPv6/ports). Counterexample tests the overbroad spec, not a request to broaden the supported address grammar.

## F5 — minor — universal gate-firing claim has no build-step evidence

- Target anchor: `.claude/rules/gate-checks.md:10` = `Each step fires on a planted bad input` (`rg -Fn` matched). be205eb body says `Every step fired on a planted seed`; `.agent/gate:35` includes the standalone static build step.
- Missing: build seed + command + observed failure; the recorded `t.Fatal` test seed is a test-only file and does not fail `go build ./cmd/offload`. All listed other seeds target different checks. No claim that build itself is broken.
- Negative evidence: positive control expected one fmt record → `rg -Fnc -- '- fmt:' .claude/rules/gate-checks.md` = `1`, rc0. `rg -Fn -- 'build' .claude/rules/gate-checks.md` → empty, rc1. Complete rule and commit range read; all named gate logs inspected for step/rc summaries; scratch seed inventory has no build fixture/trace.
- Acceptance: record/re-run a real compile-failure seed against the build step, restore clean state, record its rc in the firing rule. MAIN can supply an existing trace instead; until then universal coverage is unconfirmed for this one step.

## Evidence + limits

- All red observations = target cac258b plus review tests only, isolated fake hosts/config/runtime. Tests committed on wt/reviewer-2 at `2adfae2e6bb5e6ac6b227f43b34814500ac63e21` (only `internal/pool/claims_review_test.go` + `test/e2e/claims_review_test.go`). New tests use existing harness; Go 1.27.3. LSP `.go` unavailable.
- Baseline spot-checks passed: pool path/default/rejections/missing-file; probe parser; 14-word shell quoting; broken-repo refusal; upload-list/get/global-excludes; 14-value heavy differential; offload↔heavy status/flock; command lock-fd isolation. `go test` selected regex across internal pool/probe/remote/mirror/slots → rc0.
- Baseline real-binary fake-SSH spot-checks passed (22 functions): C1/C2/C3/C4/C5/P3/R2/L1/L2, Y1/Y2/Y3/Y4 differential/Y5 prune/Y6/Y7/Y8/Y9/Y10 upload, X2/X4/X6. `go test ./test/e2e -run '<the named functions>' -count=1` → rc0, 14.343s. Full gate, whole suite, mutation battery, live SSH not re-run.
- First baseline attempt put TMPDIR under the worktree; plain-directory fixtures discovered its enclosing git repo and failed. Those results were discarded as harness-environment artifacts; normal private Go temp directories outside git restored baseline green. Red F1/F2/F3/F4 subsequently reproduced there too.
- Gate seeds re-run with pinned tools, target unchanged: fmt file-mode rc1 (stdin-mode probe rc0 is a different interface, discarded); lint unchecked Remove rc1; shell rc1; actionlint rc1; Betterleaks synthetic committed token rc1 (redacted, validation=false); nixfmt pre-format files rc1 (six fail); spec-lint history line rc1; zizmor fixture rc14. Positive new-test fmt rc0; new-test package lint rc0. `go mod tidy` seed, vuln rc3, tool digest-mismatch rc2 not re-run; recipes/retained fixture/output inspected. Historical Nix AST comparison not re-run: needs real host.
- B2 trace: archive/tester-1 = 0b8eb3b; tester report retains 36 failing contract functions / 100 leaves, exact 282d82f base + command. e2e-1.log reproduces command-INT/nested-repo reds (and explicitly corrected NUL test error); e2e-2.log reproduces KILL-grace red; e2e-4.log = 132.550s green. gate-1/2 logs' cold/warm step totals fit 45s/4.6s; gate-10 = every step rc0. Heavy fixture SHA256 = cc6f4a503b4942fec9ace13820f3d9e963d74f6f9fd60c44e07f14cf7b41f71b. Trading copy trace = 104,913,754,525 total bytes / 18,424,452,359 sent / rsync rc0; verify log = rc0. General-1 Round 2 records paused deployment + retained activation blockers. Historical timing samples are not future performance bounds.
- Real-host claims unconfirmed by this reviewer: machine hardware, deployed cgroup/SSH/boot/account state, network/startup/CPU timings, live prune/session-lock experiments, trading gate census, AST equality. Where looked: full commit bodies, machine config, nixfmt before/after files, retained scratch reports/logs. Not re-run: needs real host. Unchanged boot/auth prose is outside this claim review's changed machine surface; no security audit implied.
- Initial scratch-file discovery output capped at 60/90 entries; subsequent checks used explicit known paths and full named files. No global assertion that unrelated scratch contains no further evidence. Historical intermediate watcher/pool/mirror reds without an immutable pre-fix revision are traceable reports, not independently reproduced here.

## Cross-review deduplication

- Independently reproduced overlaps: F2 = reviewer-1 F1 (P2 zero slots); F3 = reviewer-1 F2 (R2 missing dirs). Reviewer-1 reports its red tests committed at `df46bd66671ca4649221d512404c253173e3c56c`; one ruling + one retained regression per defect suffices. My README F1 + spec F4 counterexamples remain distinct.

## Register (outside blocking acceptance)

- `.agent/spec.md:22` says the watcher kills its process group; code uses `pkill -s` (session), matching the sentence's main guarantee and X1. Narrow explanatory wording cleanup; no weaker runtime behavior.
- Missing build record is the only evidence-gap finding; absence of a live rerun is not itself a finding where the brief explicitly accepts recorded real-host evidence.

