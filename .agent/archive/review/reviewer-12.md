# reviewer-12 — survivor rulings, `internal/cli/ctr.go` (sweep at b7645d8, CI replay run 40534349336)

| id | site | func | mutator | change | replay | ruling |
|----|------|------|---------|--------|--------|--------|
| `72309683` | `ctr.go:37` | parseCtr | statement/remove | Changes: `fs.SetOutput(io.Discard)` → `_, _ = fs.SetOutput, io.Discard` | 49 covering tests: pass | new test: C4 — flag package leaks unprefixed diagnostics (`TestC4CtrFlagErrorsUseOffloadPrefix`; red ×2, green) |
| `0012c595` | `ctr.go:41` | parseCtr | statement/return | `- return o, err`; `- }`; `+ return ctrOptions{}, err`; `+ }` | 2 covering tests: pass | equivalent: `ctrCommand` returns on every `parseCtr` error before reading options (`ctr.go:55–62`). |
| `804fee7a` | `ctr.go:46` | parseCtr | statement/return | Changes: `return o, errors.New("--layers excludes --check")` → `return ctrOptions{}, errors.New("--layers excludes --check")` | 1 covering tests: pass | equivalent: `ctrCommand` returns on every `parseCtr` error before reading options (`ctr.go:55–62`). |
| `e62afff4` | `ctr.go:48` | parseCtr | statement/return | Changes: `return o, errors.New("--layers and --check take no command")` → `return ctrOptions{}, errors.New("--layers and --check take no command")` | 1 covering tests: pass | equivalent: `ctrCommand` returns on every `parseCtr` error before reading options (`ctr.go:55–62`). |
| `830dcf57` | `ctr.go:50` | parseCtr | statement/return | Changes: `return o, errors.New("give a command after `ctr`, e.g. `offload ctr -- make -j8`")` → `return ctrOptions{}, errors.New("give a command after `ctr`, e.g. `offload ctr -- make -j8`")` | 1 covering tests: pass | equivalent: `ctrCommand` returns on every `parseCtr` error before reading options (`ctr.go:55–62`). |
| `8169f7f2` | `ctr.go:80` | ctrCommand | expression/error-guard | `- if err != nil {`; `- fmt.Fprintf(stderr, "offload: %v\n", err)`; `- return ExitFailure`; `- }`; `+ if false {`; `+ fmt.Fprintf(stderr, "offload: %v\n", err)`; `+ return ExitFailure`; `+ }` | 34 covering tests: pass | contract-open: failed getwd falls through to project lookup; deleted cwd still yields 125 with git’s diagnostic (K3 fixes the resolved root/failure, not this diagnostic or preliminary retry). |
| `25348d26` | `ctr.go:81` | ctrCommand | statement/remove | `- fmt.Fprintf(stderr, "offload: %v\n", err)`; `- return ExitFailure`; `- }`; `+ _, _, _ = fmt.Fprintf, stderr, err`; `+ return ExitFailure`; `+ }` | no e2e test executes the line | contract-open: getcwd-failure diagnostic omitted; status 125 unchanged (K3 fixes setup failure status, not this diagnostic). |
| `75359f3c` | `ctr.go:81` | ctrCommand | branch/if | `- fmt.Fprintf(stderr, "offload: %v\n", err)`; `- return ExitFailure`; `- }`; `+ _, _, _ = fmt.Fprintf, stderr, err`; `+ `; `+ }` | no e2e test executes the line | contract-open: same as 8169f7f2; deleted cwd still yields 125 with git’s diagnostic (K3 leaves diagnostic wording/preliminary retry open). |
| `ee27fa2f` | `ctr.go:82` | ctrCommand | statement/return | `- fmt.Fprintf(stderr, "offload: %v\n", err)`; `- return ExitFailure`; `- }`; `+ fmt.Fprintf(stderr, "offload: %v\n", err)`; `+ return 0`; `+ }` | no e2e test executes the line | new test: K3 — unresolvable invocation cwd returns success without running CMD (`TestK3CtrUnlinkedCwdFails`; red ×2, green) |
| `7c884b3d` | `ctr.go:91` | execNative | statement/remove | `- fmt.Fprintln(stderr, msg)`; `- return rc`; `- }`; `+ _, _, _ = fmt.Fprintln, stderr, msg`; `+ return rc`; `+ }` | 2 covering tests: pass | contract-open: native lookup diagnostic omitted; 126/127 unchanged (K2/X2 leave diagnostic presence open). |
| `47b32f08` | `ctr.go:95` | execNative | statement/remove | `- err := syscall.Exec(path, argv, os.Environ()) //nolint:gosec // G204: the caller's own command, as given`; `- fmt.Fprintf(stderr, "offload: %s: %v\n", argv[0], err)`; `+ err := syscall.Exec(path, argv, os.Environ())`; `+ _, _, _, _ = fmt.Fprintf, stderr, argv, err` | 1 covering tests: pass | contract-open: native exec-refusal diagnostic omitted; 126 unchanged (K2 leaves diagnostic presence open). |
| `a31cfd70` | `ctr.go:103` | localLayers | expression/error-guard | `- if err == nil {`; `- p, err = ctr.Find(ctx, cwd)`; `- }`; `+ if true {`; `+ p, err = ctr.Find(ctx, cwd)`; `+ }` | 12 covering tests: pass | contract-open: --layers tries project lookup after failed getwd; deleted cwd still yields 125 with git’s diagnostic (K3/K9 leave diagnostic wording/preliminary retry open). |
| `59d50062` | `ctr.go:108` | localLayers | statement/remove | `- fmt.Fprintf(stderr, "offload: %v\n", err)`; `- return nil, ExitUsage`; `- }`; `+ _, _, _ = fmt.Fprintf, stderr, err`; `+ return nil, ExitUsage`; `+ }` | 1 covering tests: pass | new test: K3 — missing-definition diagnostic loses its path (`TestK3LayersMissingDefinitionNamesPath`; red ×2, green) |
| `ffefd9da` | `ctr.go:112` | localLayers | expression/error-guard | `- if err == nil {`; `- tag, err = ctr.Ensure(ctx, p, os.Getenv, stderr)`; `- }`; `+ if true {`; `+ tag, err = ctr.Ensure(ctx, p, os.Getenv, stderr)`; `+ }` | 11 covering tests: pass | contract-open: --layers can finish after its running binary is unlinked during Find; valid context + successful inspect still satisfy K9 (K6’s mounted-binary requirement applies to runs, not layer inspection). |
| `37f83a36` | `ctr.go:116` | localLayers | expression/error-guard | `- if err == nil {`; `- layers, err = ctr.Layers(ctx, tag)`; `- }`; `+ if true {`; `+ layers, err = ctr.Layers(ctx, tag)`; `+ }` | 11 covering tests: pass | new test: K5/K9 — failed image existence check becomes successful layer output (`TestK9LayersExistsFailureCannotBecomeSuccess`; red ×2, green) |
| `67f858df` | `ctr.go:120` | localLayers | conditional/negated | Changes: `if ctx.Err() == nil { // a signal's own status follows; its fallout is no news` → `if ctx.Err() != nil { // a signal's own status follows; its fallout is no news` | 4 covering tests: pass | new test: S6/K5 — rejected lock directory is not named (`TestS6LayersLockFailureNamesDirectory`; red ×2, green) |
| `e1dca9c8` | `ctr.go:120` | localLayers | expression/error-guard | `- if ctx.Err() == nil { // a signal's own status follows; its fallout is no news`; `- fmt.Fprintf(stderr, "offload: %v\n", err)`; `- }`; `+ if true {`; `+ fmt.Fprintf(stderr, "offload: %v\n", err)`; `+ }` | 4 covering tests: pass | contract-open: extra setup-error diagnostic after a signal (K8/K9 fix signal status, not silent stderr). |
| `1fb10618` | `ctr.go:121` | localLayers | statement/remove | `- if ctx.Err() == nil { // a signal's own status follows; its fallout is no news`; `- fmt.Fprintf(stderr, "offload: %v\n", err)`; `- }`; `+ if ctx.Err() == nil {`; `+ _, _, _ = fmt.Fprintf, stderr, err`; `+ }` | 3 covering tests: pass | new test: S6/K5 — rejected lock directory is not named (`TestS6LayersLockFailureNamesDirectory`; red ×2, green) |
| `d465d9c3` | `ctr.go:121` | localLayers | branch/if | `- if ctx.Err() == nil { // a signal's own status follows; its fallout is no news`; `- fmt.Fprintf(stderr, "offload: %v\n", err)`; `- }`; `+ if ctx.Err() == nil {`; `+ _, _, _ = fmt.Fprintf, stderr, err`; `+ }` | 3 covering tests: pass | new test: S6/K5 — rejected lock directory is not named (`TestS6LayersLockFailureNamesDirectory`; red ×2, green) |
| `533798d6` | `ctr.go:154` | ctrCheck | statement/remove | `- fmt.Fprintf(stderr, "offload: %v\n", err)`; `- return ExitUsage`; `- }`; `+ _, _, _ = fmt.Fprintf, stderr, err`; `+ return ExitUsage`; `+ }` | 1 covering tests: pass | contract-open: pool-file diagnostic omitted; rc 2 unchanged (K10 fixes missing/invalid pool status; P3 names paths for run/status, not ctr --check). |
| `3ac5cc11` | `ctr.go:159` | ctrCheck | expression/error-guard | `- if err == nil {`; `- sum, err = fileSHA256(self)`; `- }`; `+ if true {`; `+ sum, err = fileSHA256(self)`; `+ }` | 8 covering tests: pass | contract-open: executable-resolution error replaced by ReadFile failure on its empty/invalid result; 125 unchanged (K10 leaves this diagnostic wording open). |
| `15552755` | `ctr.go:162` | ctrCheck | expression/error-guard | `- if err != nil {`; `- fmt.Fprintf(stderr, "offload: %v\n", err)`; `- return ExitFailure`; `- }`; `+ if false {`; `+ fmt.Fprintf(stderr, "offload: %v\n", err)`; `+ return ExitFailure`; `+ }` | 8 covering tests: pass | new test: K10 — identical binary bytes reported as different after a local read failure (`TestK10UnreadableBinaryDoesNotInventDifference`; red ×2, green) |
| `6d59f5b7` | `ctr.go:163` | ctrCheck | statement/remove | `- fmt.Fprintf(stderr, "offload: %v\n", err)`; `- return ExitFailure`; `- }`; `+ _, _, _ = fmt.Fprintf, stderr, err`; `+ return ExitFailure`; `+ }` | no e2e test executes the line | contract-open: local binary-hash failure diagnostic omitted; 125 unchanged (K10 leaves this diagnostic presence open). |
| `eab3e0f9` | `ctr.go:163` | ctrCheck | branch/if | `- fmt.Fprintf(stderr, "offload: %v\n", err)`; `- return ExitFailure`; `- }`; `+ _, _, _ = fmt.Fprintf, stderr, err`; `+ `; `+ }` | no e2e test executes the line | new test: K10 — identical binary bytes reported as different after a local read failure (`TestK10UnreadableBinaryDoesNotInventDifference`; red ×2, green) |
| `5253b31b` | `ctr.go:164` | ctrCheck | statement/return | `- fmt.Fprintf(stderr, "offload: %v\n", err)`; `- return ExitFailure`; `- }`; `+ fmt.Fprintf(stderr, "offload: %v\n", err)`; `+ return 0`; `+ }` | no e2e test executes the line | new test: K10 — local binary read failure reports all-host success (`TestK10UnreadableBinaryDoesNotInventDifference`; red ×2, green) |
| `3062572d` | `ctr.go:198` | ctrCheck | statement/remove | `- lines[i] = fmt.Sprintf("%s: failed (%v)", h.Name, err)`; `- return`; `+ _, _, _, _, _ = lines, i, fmt.Sprintf, h.Name, err`; `+ return` | no e2e test executes the line | new test: K10 — failed self launch yields a blank host line (`TestK10CheckLaunchFailureNamesHost`; red ×2, green) |
| `3654fbf5` | `ctr.go:198` | ctrCheck | branch/case | `- lines[i] = fmt.Sprintf("%s: failed (%v)", h.Name, err)`; `- return`; `+ _, _, _, _, _ = lines, i, fmt.Sprintf, h.Name, err` | no e2e test executes the line | contract-open: startup-error reason changes; a zero-exit child with complete framed output can succeed despite an inherited stdout pipe surviving 30 s (K10 fixes framing/status/comparison, not post-exit pipe-drain policy). |
| `105cb513` | `ctr.go:203` | ctrCheck | branch/if | `- lines[i] = fmt.Sprintf("%s: failed (no offload-check line in its output)", h.Name)`; `- return`; `- }`; `+ _, _, _, _ = lines, i, fmt.Sprintf, h.Name`; `+ `; `+ }` | no e2e test executes the line | new test: K10 — no nonce-framed host result classified as different instead of failed (`TestK10CheckMissingFrameIsFailed`; red ×2, green) |
| `93f93f10` | `ctr.go:203` | ctrCheck | statement/remove | `- lines[i] = fmt.Sprintf("%s: failed (no offload-check line in its output)", h.Name)`; `- return`; `- }`; `+ _, _, _, _ = lines, i, fmt.Sprintf, h.Name`; `+ return`; `+ }` | no e2e test executes the line | new test: K10 — no nonce-framed host result yields a blank host line (`TestK10CheckMissingFrameIsFailed`; red ×2, green) |
| `4e889aee` | `ctr.go:220` | ctrCheck | statement/remove | `- // --host waits while the host is down (L4) ⇒ a down host is reported, not waited on.`; `+ `; `- // remote.Command: a signal reaches the host run as TERM (it tears down + logs its end, J2),`; `- // never as exec's default KILL.`; `+ `; `- rc = 0`; `+ _ = rc` | 5 covering tests: pass | equivalent: the `localLayers` rc must be 0 to pass the return guard (`ctr.go:166–172`); goroutines never write rc, so assigning 0 again has no effect. |
| `3ae4ecd2` | `ctr.go:246` | fileSHA256 | expression/error-guard | `- if err != nil {`; `- return "", err`; `- }`; `+ if false {`; `+ return "", err`; `+ }` | 8 covering tests: pass | new test: K10 — identical binary bytes reported as different after a local read failure (`TestK10UnreadableBinaryDoesNotInventDifference`; red ×2, green) |
| `ab9eef05` | `ctr.go:247` | fileSHA256 | statement/return | `- return "", err`; `- }`; `+ return "", nil`; `+ }` | no e2e test executes the line | new test: K10 — identical binary bytes reported as different after a local read failure (`TestK10UnreadableBinaryDoesNotInventDifference`; red ×2, green) |
| `b6abca99` | `ctr.go:247` | fileSHA256 | branch/if | `- return "", err`; `- }`; `+ _ = err`; `+ }` | no e2e test executes the line | new test: K10 — identical binary bytes reported as different after a local read failure (`TestK10UnreadableBinaryDoesNotInventDifference`; red ×2, green) |

## Notes

Scope = the 33 seeded ctr.go survivors at 14994a8. Go LSP unavailable (`No LSP server available for file type: .go`); references inspected with rg + enclosing source. Full gate not run by this reviewer; narrow tests below.

### 72309683

- Test = `test/e2e/reviewer12_ctr_mutation_test.go::TestC4CtrFlagErrorsUseOffloadPrefix`. Severity = minor (coverage gap; HEAD behavior passes).
- Exact agentic diff applied to 14994a8; command `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-12 && umask 022 && go test -count=1 -run '^TestC4CtrFlagErrorsUseOffloadPrefix$' ./test/e2e/)` → rc 1 twice; reverted command → rc 0.
```text
--- FAIL: TestC4CtrFlagErrorsUseOffloadPrefix (0.01s)
    reviewer12_ctr_mutation_test.go:16: unprefixed ctr diagnostic: "flag provided but not defined: -bogus\nUsage of ctr:… [remaining CLI help omitted]
```

### 59d50062

- Test = `test/e2e/reviewer12_ctr_mutation_test.go::TestK3LayersMissingDefinitionNamesPath`. Severity = minor (coverage gap; HEAD behavior passes).
- Exact agentic diff applied to 14994a8; command `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-12 && umask 022 && go test -count=1 -run '^TestK3LayersMissingDefinitionNamesPath$' ./test/e2e/)` → rc 1 twice; reverted command → rc 0.
```text
--- FAIL: TestK3LayersMissingDefinitionNamesPath (0.03s)
    reviewer12_ctr_mutation_test.go:25: "" does not contain "/tmp/TestK3LayersMissingDefinitionNamesPath469887219/001/project/.devcontainer/Containerfile"
```

### 67f858df

- Test = `test/e2e/reviewer12_ctr_mutation_test.go::TestS6LayersLockFailureNamesDirectory`. Severity = minor (coverage gap; HEAD behavior passes).
- Exact agentic diff applied to 14994a8; command `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-12 && umask 022 && go test -count=1 -run '^TestS6LayersLockFailureNamesDirectory$' ./test/e2e/)` → rc 1 twice; reverted command → rc 0.
```text
--- FAIL: TestS6LayersLockFailureNamesDirectory (0.04s)
    reviewer12_ctr_mutation_test.go:38: "" does not contain "/tmp/TestS6LayersLockFailureNamesDirectory3309127147/001/runtime/offload/ctr"
```

### 1fb10618

- Test = `test/e2e/reviewer12_ctr_mutation_test.go::TestS6LayersLockFailureNamesDirectory`. Severity = minor (coverage gap; HEAD behavior passes).
- Exact agentic diff applied to 14994a8; command `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-12 && umask 022 && go test -count=1 -run '^TestS6LayersLockFailureNamesDirectory$' ./test/e2e/)` → rc 1 twice; reverted command → rc 0.
```text
--- FAIL: TestS6LayersLockFailureNamesDirectory (0.01s)
    reviewer12_ctr_mutation_test.go:38: "" does not contain "/tmp/TestS6LayersLockFailureNamesDirectory4022721945/001/runtime/offload/ctr"
```

### d465d9c3

- Test = `test/e2e/reviewer12_ctr_mutation_test.go::TestS6LayersLockFailureNamesDirectory`. Severity = minor (coverage gap; HEAD behavior passes).
- Exact agentic diff applied to 14994a8; command `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-12 && umask 022 && go test -count=1 -run '^TestS6LayersLockFailureNamesDirectory$' ./test/e2e/)` → rc 1 twice; reverted command → rc 0.
```text
--- FAIL: TestS6LayersLockFailureNamesDirectory (0.01s)
    reviewer12_ctr_mutation_test.go:38: "" does not contain "/tmp/TestS6LayersLockFailureNamesDirectory3175089669/001/runtime/offload/ctr"
```

### ee27fa2f

- Test = `test/e2e/reviewer12_ctr_mutation_test.go::TestK3CtrUnlinkedCwdFails`. Severity = minor (coverage gap; HEAD behavior passes).
- Exact agentic diff applied to 14994a8; command `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-12 && umask 022 && go test -count=1 -run '^TestK3CtrUnlinkedCwdFails$' ./test/e2e/)` → rc 1 twice; reverted command → rc 0.
```text
--- FAIL: TestK3CtrUnlinkedCwdFails (0.01s)
    reviewer12_ctr_mutation_test.go:60: exit = 0, want 125; stdout="" stderr="offload: getwd: no such file or directory\n"
```

### 37f83a36

- Test = `test/e2e/reviewer12_ctr_mutation_test.go::TestK9LayersExistsFailureCannotBecomeSuccess`. Severity = minor (coverage gap; HEAD behavior passes).
- Exact agentic diff applied to 14994a8; command `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-12 && umask 022 && go test -count=1 -run '^TestK9LayersExistsFailureCannotBecomeSuccess$' ./test/e2e/)` → rc 1 twice; reverted command → rc 0.
```text
--- FAIL: TestK9LayersExistsFailureCannotBecomeSuccess (0.01s)
    reviewer12_ctr_mutation_test.go:48: exit = 0, want 125; stdout="sha256:a\n" stderr=""
```

### 15552755

- Test = `test/e2e/reviewer12_ctr_mutation_test.go::TestK10UnreadableBinaryDoesNotInventDifference`. Severity = minor (coverage gap; HEAD behavior passes).
- Exact agentic diff applied to 14994a8; command `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-12 && umask 022 && go test -count=1 -run '^TestK10UnreadableBinaryDoesNotInventDifference$' ./test/e2e/)` → rc 1 twice; reverted command → rc 0.
```text
--- FAIL: TestK10UnreadableBinaryDoesNotInventDifference (0.37s)
    reviewer12_ctr_mutation_test.go:81: unreadable identical binary: exit=1 stdout="alpha: same (1 layers); offload binary differs\n" stderr="offload: alpha slot 1/2 (8 CPUs, 25% busy) → ~/offload/project-e6b4db89\n"
```

### eab3e0f9

- Test = `test/e2e/reviewer12_ctr_mutation_test.go::TestK10UnreadableBinaryDoesNotInventDifference`. Severity = minor (coverage gap; HEAD behavior passes).
- Exact agentic diff applied to 14994a8; command `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-12 && umask 022 && go test -count=1 -run '^TestK10UnreadableBinaryDoesNotInventDifference$' ./test/e2e/)` → rc 1 twice; reverted command → rc 0.
```text
--- FAIL: TestK10UnreadableBinaryDoesNotInventDifference (0.37s)
    reviewer12_ctr_mutation_test.go:81: unreadable identical binary: exit=1 stdout="alpha: same (1 layers); offload binary differs\n" stderr="offload: alpha slot 1/2 (8 CPUs, 25% busy) → ~/offload/project-9a035858\n"
```

### 5253b31b

- Test = `test/e2e/reviewer12_ctr_mutation_test.go::TestK10UnreadableBinaryDoesNotInventDifference`. Severity = minor (coverage gap; HEAD behavior passes).
- Exact agentic diff applied to 14994a8; command `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-12 && umask 022 && go test -count=1 -run '^TestK10UnreadableBinaryDoesNotInventDifference$' ./test/e2e/)` → rc 1 twice; reverted command → rc 0.
```text
--- FAIL: TestK10UnreadableBinaryDoesNotInventDifference (0.02s)
    reviewer12_ctr_mutation_test.go:81: unreadable identical binary: exit=0 stdout="" stderr="offload: open /tmp/TestK10UnreadableBinaryDoesNotInventDifference2990326312/001/unreadable-offload: permission denied\n"
```

### 3ae4ecd2

- Test = `test/e2e/reviewer12_ctr_mutation_test.go::TestK10UnreadableBinaryDoesNotInventDifference`. Severity = minor (coverage gap; HEAD behavior passes).
- Exact agentic diff applied to 14994a8; command `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-12 && umask 022 && go test -count=1 -run '^TestK10UnreadableBinaryDoesNotInventDifference$' ./test/e2e/)` → rc 1 twice; reverted command → rc 0.
```text
--- FAIL: TestK10UnreadableBinaryDoesNotInventDifference (0.38s)
    reviewer12_ctr_mutation_test.go:81: unreadable identical binary: exit=1 stdout="alpha: same (1 layers); offload binary differs\n" stderr="offload: alpha slot 1/2 (8 CPUs, 25% busy) → ~/offload/project-d52796f3\n"
```

### ab9eef05

- Test = `test/e2e/reviewer12_ctr_mutation_test.go::TestK10UnreadableBinaryDoesNotInventDifference`. Severity = minor (coverage gap; HEAD behavior passes).
- Exact agentic diff applied to 14994a8; command `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-12 && umask 022 && go test -count=1 -run '^TestK10UnreadableBinaryDoesNotInventDifference$' ./test/e2e/)` → rc 1 twice; reverted command → rc 0.
```text
--- FAIL: TestK10UnreadableBinaryDoesNotInventDifference (0.35s)
    reviewer12_ctr_mutation_test.go:81: unreadable identical binary: exit=1 stdout="alpha: same (1 layers); offload binary differs\n" stderr="offload: alpha slot 1/2 (8 CPUs, 25% busy) → ~/offload/project-1e942a9d\n"
```

### b6abca99

- Test = `test/e2e/reviewer12_ctr_mutation_test.go::TestK10UnreadableBinaryDoesNotInventDifference`. Severity = minor (coverage gap; HEAD behavior passes).
- Exact agentic diff applied to 14994a8; command `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-12 && umask 022 && go test -count=1 -run '^TestK10UnreadableBinaryDoesNotInventDifference$' ./test/e2e/)` → rc 1 twice; reverted command → rc 0.
```text
--- FAIL: TestK10UnreadableBinaryDoesNotInventDifference (0.37s)
    reviewer12_ctr_mutation_test.go:81: unreadable identical binary: exit=1 stdout="alpha: same (1 layers); offload binary differs\n" stderr="offload: alpha slot 1/2 (8 CPUs, 25% busy) → ~/offload/project-3eb67bde\n"
```

### 105cb513

- Test = `test/e2e/reviewer12_ctr_mutation_test.go::TestK10CheckMissingFrameIsFailed`. Severity = minor (coverage gap; HEAD behavior passes).
- Exact agentic diff applied to 14994a8; command `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-12 && umask 022 && go test -count=1 -run '^TestK10CheckMissingFrameIsFailed$' ./test/e2e/)` → rc 1 twice; reverted command → rc 0.
```text
--- FAIL: TestK10CheckMissingFrameIsFailed (0.35s)
    reviewer12_ctr_mutation_test.go:95: missing host frame is not a failed host verdict: "alpha: different (here 1, there 0 layers); offload binary differs\n"
```

### 93f93f10

- Test = `test/e2e/reviewer12_ctr_mutation_test.go::TestK10CheckMissingFrameIsFailed`. Severity = minor (coverage gap; HEAD behavior passes).
- Exact agentic diff applied to 14994a8; command `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-12 && umask 022 && go test -count=1 -run '^TestK10CheckMissingFrameIsFailed$' ./test/e2e/)` → rc 1 twice; reverted command → rc 0.
```text
--- FAIL: TestK10CheckMissingFrameIsFailed (0.38s)
    reviewer12_ctr_mutation_test.go:95: missing host frame is not a failed host verdict: "\n"
```

### 3062572d

- Test = `test/e2e/reviewer12_ctr_mutation_test.go::TestK10CheckLaunchFailureNamesHost`. Severity = minor (coverage gap; HEAD behavior passes).
- Exact agentic diff applied to 14994a8; command `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-12 && umask 022 && go test -count=1 -run '^TestK10CheckLaunchFailureNamesHost$' ./test/e2e/)` → rc 1 twice; reverted command → rc 0.
```text
--- FAIL: TestK10CheckLaunchFailureNamesHost (0.05s)
    reviewer12_ctr_mutation_test.go:114: launch failure lost its host verdict: "\n"
```

### Probe HEAD

Temporary diagnostic test `Test(K3R12ProbeDeletedCwd|K9R12ProbeUnlinkedBinary|K10R12ProbeWaitDelay)`; `go test -v -count=1 -run '^Test(K3R12ProbeDeletedCwd|K9R12ProbeUnlinkedBinary|K10R12ProbeWaitDelay)$' ./test/e2e/` under worktree + umask 022 → rc 0. No failure assertion; observations only.
```text
    reviewer12_probe_test.go:19: args=["ctr" "--" "true"] code=125 stdout="" stderr="offload: getwd: no such file or directory\n"
    reviewer12_probe_test.go:19: args=["ctr" "--layers"] code=125 stdout="" stderr="offload: getwd: no such file or directory\n"
    reviewer12_probe_test.go:39: code=125 stdout="" stderr="offload: the offload binary was replaced while it ran; run the command again: lstat /tmp/TestK9R12ProbeUnlinkedBinary1111155007/001/vanishing-offload: no such file or directory\n"
    reviewer12_probe_test.go:62: code=1 stdout="alpha: failed (exec: WaitDelay expired before I/O complete)\n" stderr="offload: alpha slot 1/2 (8 CPUs, 25% busy) → ~/offload/project-688a7650\noffload: t25-480: no systemd scope; a process that starts its own session can outlive this run\n"
```

### Probe 8169f7f2

Temporary diagnostic test `TestK3R12ProbeDeletedCwd`; `go test -v -count=1 -run '^TestK3R12ProbeDeletedCwd$' ./test/e2e/` under worktree + umask 022 → rc 0. No failure assertion; observations only.
```text
    reviewer12_probe_test.go:19: args=["ctr" "--" "true"] code=125 stdout="" stderr="offload: git rev-parse --path-format=absolute --show-toplevel --git-dir --git-common-dir --show-prefix: fatal: Unable to read current working directory: No such file or directory\n"
    reviewer12_probe_test.go:19: args=["ctr" "--layers"] code=125 stdout="" stderr="offload: getwd: no such file or directory\n"
```

### Probe 75359f3c

Temporary diagnostic test `TestK3R12ProbeDeletedCwd`; `go test -v -count=1 -run '^TestK3R12ProbeDeletedCwd$' ./test/e2e/` under worktree + umask 022 → rc 0. No failure assertion; observations only.
```text
    reviewer12_probe_test.go:19: args=["ctr" "--" "true"] code=125 stdout="" stderr="offload: git rev-parse --path-format=absolute --show-toplevel --git-dir --git-common-dir --show-prefix: fatal: Unable to read current working directory: No such file or directory\n"
    reviewer12_probe_test.go:19: args=["ctr" "--layers"] code=125 stdout="" stderr="offload: getwd: no such file or directory\n"
```

### Probe a31cfd70

Temporary diagnostic test `TestK3R12ProbeDeletedCwd`; `go test -v -count=1 -run '^TestK3R12ProbeDeletedCwd$' ./test/e2e/` under worktree + umask 022 → rc 0. No failure assertion; observations only.
```text
    reviewer12_probe_test.go:19: args=["ctr" "--" "true"] code=125 stdout="" stderr="offload: getwd: no such file or directory\n"
    reviewer12_probe_test.go:19: args=["ctr" "--layers"] code=125 stdout="" stderr="offload: git rev-parse --path-format=absolute --show-toplevel --git-dir --git-common-dir --show-prefix: fatal: Unable to read current working directory: No such file or directory\n"
```

### Probe ffefd9da

Temporary diagnostic test `TestK9R12ProbeUnlinkedBinary`; `go test -v -count=1 -run '^TestK9R12ProbeUnlinkedBinary$' ./test/e2e/` under worktree + umask 022 → rc 0. No failure assertion; observations only.
```text
    reviewer12_probe_test.go:39: code=0 stdout="sha256:a\n" stderr=""
```

### Probe 3654fbf5

Temporary diagnostic test `TestK10R12ProbeWaitDelay`; `go test -v -count=1 -run '^TestK10R12ProbeWaitDelay$' ./test/e2e/` under worktree + umask 022 → rc 0. No failure assertion; observations only.
```text
    reviewer12_probe_test.go:62: code=0 stdout="alpha: same (1 layers)\n" stderr="offload: alpha slot 1/2 (8 CPUs, 25% busy) → ~/offload/project-79b410af\noffload: t25-480: no systemd scope; a process that starts its own session can outlive this run\n"
```

### 8169f7f2 / 75359f3c / a31cfd70

`ctr.Find` begins with `filepath.EvalSymlinks(cwd)` (`internal/ctr/ctr.go:60`), then `mirror.Locate` executes git. After the invocation cwd is removed, the mutation passes an empty cwd: Find reaches git, which also refuses the deleted cwd. K3/K9 do not select getwd's diagnostic over git's or fix the number of preliminary setup attempts. Probe = native `ctr -- bash -c 'rmdir -- "$PWD"; export OFFLOAD_CTR=1; exec "$@"' _ <binary> ctr -- true` (or `ctr --layers`); exact observations above. The committed `TestK3CtrUnlinkedCwdFails` also drives this setup error and kills the success-return mutation.

### ffefd9da

`ctr.Find` initializes Context before `Executable` (`internal/ctr/ctr.go:79–87`): removing the executable during git lookup gives a valid Context + error. Skipping the guard then allows Ensure/Layers to succeed. K9 requires a valid image and successful inspect, not a linked pathname for the controller binary; K6's mounted-binary check concerns image runs. Earlier Find errors leave Context empty (Tag then fails), or are UsageErrors returned before this guard. Probe: copied binary; PATH git wrapper removes that copy before execing real git; valid definition; fake podman says image exists + returns sha256:a. HEAD = 125; mutant = correct one-line layers, rc 0. Status divergence is real, not equivalent.

### 3654fbf5

A startup error emits no child stdout → parseCheck rejects the empty frame, keeping `failed (…)` + rc 1; only the reason changes. Context cancellation is checked after every worker joins, so the signal status wins either way. A separate observable case: remote login profile starts `setsid sh -c 'echo $$ > "$HOME/escaped.pid"; exec sleep 90' &`; a PATH systemd-run stub returns 1 (documented X1 no-scope limit), leaving that session alive with stdout open after the successful child exits. Valid framed output + same binary/layers reach the parent. `remote.Command` has `WaitDelay = 30 * time.Second` (`internal/remote/ssh.go:106`): HEAD reports failed on ErrWaitDelay; mutant accepts the complete same result. K10 does not fix this post-exit pipe-drain deadline or rejection policy. Probe observed both results; owned escaped pid killed at cleanup. Temporary probe tests removed; only the eight killing tests ship.

## Close

- Rulings = 33: new test 16; contract-open 12; equivalent 5; killed 0; n/a 0. Unknown = 0 (validator rc 1; positive control = 33 seeded table rows, rc 0).
- Tests = `test/e2e/reviewer12_ctr_mutation_test.go` (8). Branch = `wt/reviewer-12`; commit = `2ed8949839f8bcf65f0433b2b666e7b70aafa573`.
- Green: each new test after every mutant revert; all eight together with `go test -race -shuffle=on -count=1 -run '<eight exact names above>' ./test/e2e/` (3.541 s package); `golangci-lint fmt --diff test/e2e/reviewer12_ctr_mutation_test.go`; `golangci-lint run ./test/e2e/` (0 issues); `.agent/law` (new test tracked); `.agent/spec-lint`; `git diff --cached --check`.
- Full gate = not run (MAIN integration). Real podman = not used; fake podman models image exists/inspect responses, fake SSH hosts run the actual built binary + watcher. No product changes; temporary mutation/probe code reverted/removed. No unconfirmed executable test claims; unexercised recovery interleavings are not claimed equivalent.

reviewer-12-DONE-1
