# Mutation credit — MAINTAIN code review

Ruler = MAIN (not the author of the tests credited; survivors of MAIN-authored tests → reviewer at close). Sweeps at 508e0dc (disposable worktree + stub `place_test.go`, `ctr_test.go`, `ssh_test.go`; recipe `.claude/rules/code.md`), unit tests per package; escapes replayed through e2e (`patch` the agentic diff, run the regex, revert). Rows = every survivor after the replay.

## shell

59 mutants, 46 killed by unit tests, 13 escaped; e2e replay (C3/K1/K2/K7/X2/L1/R1L2) killed 4.

| id | site | mutator | change | ruling |
|----|------|---------|--------|--------|
| `2bd88240` | `look.go:71` | branch/case | Changes: `return errors.New("is a directory")` → `_ = errors.New` | contract-open: error wording (X2 fixes 126/127, not the message) |
| `3bf0e7da` | `look.go:69` | branch/case | Changes: `return err` → `_ = err` | contract-open: error wording |
| `ece4a28a` | `look.go:34` | branch/if | Removes an if-block body so the condition becomes a no-op, keeping a zero-value return when the branch was ter | equivalent: filepath.Join("", name) = name, resolved against the cwd as "." |
| `23562572` | `look.go:41` | expression/remove | Changes: `if fi, err := os.Stat(p); err == nil && !fi.IsDir() && lose == "" { //nolint:gosec // G703: the call | contract-open: which non-executable PATH match the message names (status 126 either way) |
| `ee4eaa2e` | `look.go:34` | statement/remove | Removes a statement entirely, dropping its side effect or return value | equivalent: as ece4a28a |
| `50d16e75` | `look.go:49` | statement/return | Replaces a return value with its zero value | equivalent: callers ignore err for NotFound (lookLocal + init print their own "command not found") |
| `24bcd5d6` | `look.go:25` | statement/return | Replaces a return value with its zero value | contract-open: error wording (status from check() unchanged) |
| `c05c975d` | `look.go:69` | statement/return | Changes: `return err` → `return nil` | contract-open: error wording |
| `2c763b81` | `look.go:71` | statement/return | Changes: `return errors.New("is a directory")` → `return nil` | contract-open: error wording |

## remote

48, 31 killed, 17 escaped; replay (X2–X6/C3/C5/L2/Y4/Y6/Reviewer9/harness) killed 2.

| id | site | mutator | change | ruling |
|----|------|---------|--------|--------|
| `5a20c144` | `ssh.go:106` | arithmetic/base | Changes: `cmd.WaitDelay = 30 * time.Second` → `cmd.WaitDelay = 30 / time.Second` | contract-open: post-cancel kill delay (no predicate; X3 waits for the watcher) |
| `b0941db1` | `ssh.go:68` | branch/if | Removes an if-block body so the condition becomes a no-op, keeping a zero-value return when the branch was ter | equivalent: CreateTemp in the missing dir then refuses (X5 note kept) |
| `54e9ec23` | `ssh.go:83` | branch/if | Removes an if-block body so the condition becomes a no-op, keeping a zero-value return when the branch was ter | equivalent: quoting a safe word is still exact shell |
| `abb7c21b` | `ssh.go:104` | composite/field-clear | Changes: `cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}` → `cmd.SysProcAttr = &syscall.SysProcAttr{}` | contract-open: own process group (C3/X3 outcomes unchanged) |
| `01b0f6cb` | `ssh.go:67` | expression/error-guard | Removes an error-guard block so the function continues even on error | equivalent: as b0941db1 |
| `25475cd5` | `ssh.go:73` | expression/errorf-wrap | Changes: `return nil, fmt.Errorf("socket directory %s refuses new files: %w", dir, err)` → `return nil, fmt.Er | contract-open: error wrapping/wording |
| `bf5b8ec8` | `ssh.go:54` | numbers/decrementer | Changes: `const maxControlDir = 49` → `const maxControlDir = 48` | killed: TestMultiplex (bound = X5's literal 49, fc031d9) |
| `7cf80936` | `ssh.go:54` | numbers/incrementer | Changes: `const maxControlDir = 49` → `const maxControlDir = 50` | killed: TestMultiplex (fc031d9) |
| `8851a05c` | `ssh.go:106` | numbers/decrementer | Changes: `cmd.WaitDelay = 30 * time.Second` → `cmd.WaitDelay = 29 * time.Second` | contract-open: as 5a20c144 |
| `ae40b356` | `ssh.go:106` | numbers/incrementer | Changes: `cmd.WaitDelay = 30 * time.Second` → `cmd.WaitDelay = 31 * time.Second` | contract-open: as 5a20c144 |
| `db61c2a6` | `ssh.go:104` | statement/remove | Removes a statement entirely, dropping its side effect or return value | contract-open: as abb7c21b |
| `4d318e50` | `ssh.go:106` | statement/remove | Removes a statement entirely, dropping its side effect or return value | contract-open: as 5a20c144 |
| `7724e72c` | `ssh.go:105` | statement/remove | Removes a statement entirely, dropping its side effect or return value | contract-open: TERM vs KILL on cancel (C3/X3 outcomes unchanged) |
| `14da4c1d` | `ssh.go:68` | statement/return | Replaces a return value with its zero value | killed: TestMultiplex uncreatable dir (fc031d9) |
| `7aa26568` | `ssh.go:105` | statement/return | Replaces a return value with its zero value | contract-open: as 7724e72c |

## cli

813, 68 killed, 740 escaped; replay scope = Main/usageError/parseRun/parseCheck/targets (65): 52 killed; the other ~675 escapes (e2e-credited code) → `.agent/deferred.md`.

| id | site | mutator | change | ruling |
|----|------|---------|--------|--------|
| `802566c8` | `cli.go:106` | branch/case | Changes: `return ctr.Init(args[1:], stderr)` → `_, _, _ = ctr.Init, args, stderr` | killed: TestCtrInitIsDispatched (fc031d9) |
| `4d46e9ab` | `cli.go:86` | branch/if | Removes an if-block body so the condition becomes a no-op, keeping a zero-value return when the branch was ter | killed: TestUsageErrorsOfStatusCommandsAndFlags (status -h) |
| `73ad10cd` | `cli.go:106` | numbers/decrementer | Changes: `return ctr.Init(args[1:], stderr)` → `return ctr.Init(args[0:], stderr)` | killed: TestK7StatusPassesThrough (.agent/plant) |
| `65f5e129` | `cli.go:85` | numbers/incrementer | Changes: `if len(args) > 1 {` → `if len(args) > 2 {` | killed: TestUsageErrorsOfStatusCommandsAndFlags (status extra) |
| `e0d93f38` | `cli.go:106` | numbers/incrementer | Changes: `return ctr.Init(args[1:], stderr)` → `return ctr.Init(args[2:], stderr)` | killed: TestK7StatusPassesThrough (.agent/plant) |
| `fb800c96` | `cli.go:108` | numbers/incrementer | Changes: `return usageError(stderr, fmt.Sprintf("unknown command %q", args[0]))` → `return usageError(stderr,  | killed: TestUsageErrorsOfStatusCommandsAndFlags (unknown command) |
| `e839ea84` | `cli.go:119` | statement/remove | Changes: `fs.SetOutput(io.Discard)` → `_, _ = fs.SetOutput, io.Discard` | killed: e2e TestC4UnknownFlagReasonIsOffloads |
| `8805d817` | `cli.go:108` | statement/return | Changes: `return usageError(stderr, fmt.Sprintf("unknown command %q", args[0]))` → `return 0` | killed: TestUsageErrorsOfStatusCommandsAndFlags (unknown command rc) |
| `896f956b` | `cli.go:90` | statement/return | Replaces a return value with its zero value | killed: TestUsageErrorsOfStatusCommandsAndFlags (status extra rc) |
| `3db37409` | `cli.go:127` | statement/return | Replaces a return value with its zero value | equivalent: callers read the error alone |
| `59f2f318` | `cli.go:131` | statement/return | Replaces a return value with its zero value | equivalent: as 3db37409 |
| `9f816eb1` | `cli.go:135` | statement/return | Replaces a return value with its zero value | equivalent: as 3db37409 |
| `e47bc289` | `ctr.go:235` | statement/remove | Removes a statement entirely, dropping its side effect or return value | code removed (parseCheck CR trim, no CR source), fc031d9 |

## ledger

195, 153 killed, 42 escaped; replay (TestJ*) killed 27 (13197df's 13 author-ruled sites included, re-ruled here).

| id | site | mutator | change | ruling |
|----|------|---------|--------|--------|
| `ac657a13` | `ledger.go:131` | branch/if | Removes an if-block body so the condition becomes a no-op, keeping a zero-value return when the branch was ter | equivalent: ReadBuildInfo is ok in every module build |
| `3743e003` | `ledger.go:270` | branch/if | Removes an if-block body so the condition becomes a no-op, keeping a zero-value return when the branch was ter | equivalent: record() = json.Marshal of fixed types, cannot fail |
| `12d5ad7f` | `ledger.go:290` | branch/if | Removes an if-block body so the condition becomes a no-op, keeping a zero-value return when the branch was ter | contract-open: fd close on an error path (process exits; no predicate) |
| `e5b26095` | `ledger.go:298` | branch/if | Removes an if-block body so the condition becomes a no-op, keeping a zero-value return when the branch was ter | contract-open: a close error after a whole write (not covered by J1/J4) |
| `ac6b0499` | `ledger.go:289` | conditional/negated | Changes: `if f != nil {` → `if f == nil {` | contract-open: as 12d5ad7f |
| `697daf0d` | `ledger.go:269` | expression/error-guard | Removes an error-guard block so the function continues even on error | equivalent: as 3743e003 |
| `cdff9057` | `ledger.go:278` | expression/error-guard | Removes an error-guard block so the function continues even on error | equivalent: Stat on an open fd cannot fail |
| `ac2ed2c5` | `ledger.go:313` | expression/error-guard | Removes an error-guard block so the function continues even on error | equivalent: as cdff9057 |
| `fe499bf8` | `ledger.go:281` | expression/remove | Removes an expression statement entirely, dropping its side effect | equivalent: as cdff9057 |
| `37da9104` | `ledger.go:311` | expression/remove | Removes an expression statement entirely, dropping its side effect | equivalent: as cdff9057 |
| `408ea4ed` | `ledger.go:220` | numbers/decrementer | Changes: `rc, fixed := 0, false` → `rc, fixed := (-1), false` | equivalent: initial rc overwritten before any use |
| `f5f829bb` | `ledger.go:220` | numbers/incrementer | Changes: `rc, fixed := 0, false` → `rc, fixed := 1, false` | equivalent: as 408ea4ed |
| `c0a0d043` | `ledger.go:298` | statement/remove | Removes a statement entirely, dropping its side effect or return value | contract-open: as e5b26095 |
| `bc73f804` | `ledger.go:245` | statement/return | Replaces a return value with its zero value | equivalent: as 3743e003 |
| `b3332c10` | `ledger.go:270` | statement/return | Replaces a return value with its zero value | equivalent: as 3743e003 |

## run (run.go)

85 mutants, 12 killed by unit tests, 73 escaped; replay (X1–X8/C3/C5/Y5/Y6/Y10/Y11/J2/L3/S4/G1/G3/G7/Reviewer9/K8UnderOffloadRun) killed 38.

| id | site | mutator | change | ruling |
|----|------|---------|--------|--------|
| `2a3be66a` | `run.go:144` | arithmetic/base | Changes: `return 128 + int(ws.Signal()), nil` → `return 128 - int(ws.Signal()), nil` | killed: TestSuperviseReportsAKilledProcessAs128PlusN (exact diff confirmed) |
| `2265a963` | `run.go:103` | branch/if | Removes an if-block body so the condition becomes a no-op, keeping a zero-value return when the branch was ter | equivalent/contract-open: Run then always reads the status record (one more ssh, X6 names it only for 255) or the signal path ends in the same C3 status |
| `048e38b5` | `run.go:119` | branch/if | Removes an if-block body so the condition becomes a no-op, keeping a zero-value return when the branch was ter | equivalent: StdinPipe/Start of bash/ssh (X7 requirements) + the Linux WaitStatus assertion cannot fail; a value beside an error is never read |
| `75c9b30f` | `run.go:122` | branch/if | Removes an if-block body so the condition becomes a no-op, keeping a zero-value return when the branch was ter | equivalent: StdinPipe/Start of bash/ssh (X7 requirements) + the Linux WaitStatus assertion cannot fail; a value beside an error is never read |
| `7ac953cd` | `run.go:141` | branch/if | Removes an if-block body so the condition becomes a no-op, keeping a zero-value return when the branch was ter | equivalent: StdinPipe/Start of bash/ssh (X7 requirements) + the Linux WaitStatus assertion cannot fail; a value beside an error is never read |
| `82a4e23c` | `run.go:144` | branch/if | Removes an if-block body so the condition becomes a no-op, keeping a zero-value return when the branch was ter | killed: TestSuperviseReportsAKilledProcessAs128PlusN (exact diff confirmed) |
| `5b6b7e2a` | `run.go:125` | concurrency/goroutine-remove | Converts a goroutine launch to a regular blocking call, removing concurrency | equivalent: preamble write synchronous/empty or the closer goroutine left blocked — the remote reads its preamble, Wait closes stdin, the process exits |
| `fae7762e` | `run.go:126` | expression/comparison | Changes: `if len(preamble) > 0 {` → `if len(preamble) >= 0 {` | equivalent: preamble write synchronous/empty or the closer goroutine left blocked — the remote reads its preamble, Wait closes stdin, the process exits |
| `47a8b90a` | `run.go:118` | expression/error-guard | Removes an error-guard block so the function continues even on error | equivalent: StdinPipe/Start of bash/ssh (X7 requirements) + the Linux WaitStatus assertion cannot fail; a value beside an error is never read |
| `e75f07b9` | `run.go:121` | expression/error-guard | Removes an error-guard block so the function continues even on error | equivalent: StdinPipe/Start of bash/ssh (X7 requirements) + the Linux WaitStatus assertion cannot fail; a value beside an error is never read |
| `ae80ad7e` | `run.go:141` | expression/errorf-wrap | Changes: `return 0, fmt.Errorf("%s: %w", cmd.Path, err)` → `return 0, fmt.Errorf("%s: %v", cmd.Path, err)` | equivalent: StdinPipe/Start of bash/ssh (X7 requirements) + the Linux WaitStatus assertion cannot fail; a value beside an error is never read |
| `9b7c5689` | `run.go:102` | expression/logical | Changes: `if err != nil \|\| ctx.Err() != nil \|\| (rc != 125 && rc != 255) {` → `if (err != nil \|\| ctx.Err( | equivalent/contract-open: Run then always reads the status record (one more ssh, X6 names it only for 255) or the signal path ends in the same C3 status |
| `5ab8c60d` | `run.go:102` | expression/logical | Changes: `if err != nil \|\| ctx.Err() != nil \|\| (rc != 125 && rc != 255) {` → `if err != nil && ctx.Err() ! | equivalent/contract-open: Run then always reads the status record (one more ssh, X6 names it only for 255) or the signal path ends in the same C3 status |
| `74fd5b8c` | `run.go:108` | expression/logical | Changes: `if err == nil && strings.TrimSpace(string(out)) == nonce+" "+strconv.Itoa(rc) {` → `if err == nil \| | killed: e2e TestX6StaleStatusRecordIsNotTheCommands (exact diff confirmed) |
| `5b4bdaca` | `run.go:102` | expression/remove | Removes an expression statement entirely, dropping its side effect | equivalent/contract-open: Run then always reads the status record (one more ssh, X6 names it only for 255) or the signal path ends in the same C3 status |
| `2e0f6ff9` | `run.go:102` | expression/remove | Changes: `if err != nil \|\| ctx.Err() != nil \|\| (rc != 125 && rc != 255) {` → `if err != nil \|\| ctx.Err() | equivalent/contract-open: Run then always reads the status record (one more ssh, X6 names it only for 255) or the signal path ends in the same C3 status |
| `ba760e89` | `run.go:102` | expression/remove | Removes an expression statement entirely, dropping its side effect | equivalent/contract-open: Run then always reads the status record (one more ssh, X6 names it only for 255) or the signal path ends in the same C3 status |
| `2c9cad4d` | `run.go:102` | expression/remove | Changes: `if err != nil \|\| ctx.Err() != nil \|\| (rc != 125 && rc != 255) {` → `if err != nil \|\| false \|\ | equivalent/contract-open: Run then always reads the status record (one more ssh, X6 names it only for 255) or the signal path ends in the same C3 status |
| `3324917e` | `run.go:108` | expression/remove | Removes an expression statement entirely, dropping its side effect | equivalent: a failed lookup prints nothing ⇒ no nonce match ⇒ ErrNotStarted either way |
| `484cd8aa` | `run.go:119` | numbers/decrementer | Changes: `return 0, err` → `return (-1), err` | equivalent: StdinPipe/Start of bash/ssh (X7 requirements) + the Linux WaitStatus assertion cannot fail; a value beside an error is never read |
| `7a1e2ed7` | `run.go:122` | numbers/decrementer | Changes: `return 0, err` → `return (-1), err` | equivalent: StdinPipe/Start of bash/ssh (X7 requirements) + the Linux WaitStatus assertion cannot fail; a value beside an error is never read |
| `d1b48af1` | `run.go:126` | numbers/decrementer | Changes: `if len(preamble) > 0 {` → `if len(preamble) > (-1) {` | equivalent: preamble write synchronous/empty or the closer goroutine left blocked — the remote reads its preamble, Wait closes stdin, the process exits |
| `78544a03` | `run.go:141` | numbers/decrementer | Changes: `return 0, fmt.Errorf("%s: %w", cmd.Path, err)` → `return (-1), fmt.Errorf("%s: %w", cmd.Path, err)` | equivalent: StdinPipe/Start of bash/ssh (X7 requirements) + the Linux WaitStatus assertion cannot fail; a value beside an error is never read |
| `c2020c66` | `run.go:144` | numbers/decrementer | Changes: `return 128 + int(ws.Signal()), nil` → `return 127 + int(ws.Signal()), nil` | killed: TestSuperviseReportsAKilledProcessAs128PlusN (exact diff confirmed) |
| `b976f6dd` | `run.go:119` | numbers/incrementer | Changes: `return 0, err` → `return 1, err` | equivalent: StdinPipe/Start of bash/ssh (X7 requirements) + the Linux WaitStatus assertion cannot fail; a value beside an error is never read |
| `7d71b0dd` | `run.go:122` | numbers/incrementer | Changes: `return 0, err` → `return 1, err` | equivalent: StdinPipe/Start of bash/ssh (X7 requirements) + the Linux WaitStatus assertion cannot fail; a value beside an error is never read |
| `6195fb11` | `run.go:126` | numbers/incrementer | Changes: `if len(preamble) > 0 {` → `if len(preamble) > 1 {` | equivalent: preamble write synchronous/empty or the closer goroutine left blocked — the remote reads its preamble, Wait closes stdin, the process exits |
| `a6d9d6f4` | `run.go:141` | numbers/incrementer | Changes: `return 0, fmt.Errorf("%s: %w", cmd.Path, err)` → `return 1, fmt.Errorf("%s: %w", cmd.Path, err)` | equivalent: StdinPipe/Start of bash/ssh (X7 requirements) + the Linux WaitStatus assertion cannot fail; a value beside an error is never read |
| `3dc29e36` | `run.go:144` | numbers/incrementer | Changes: `return 128 + int(ws.Signal()), nil` → `return 129 + int(ws.Signal()), nil` | killed: TestSuperviseReportsAKilledProcessAs128PlusN (exact diff confirmed) |
| `1550745e` | `run.go:133` | select/case-remove | Removes a case from a select statement, reducing channel handling paths | equivalent: preamble write synchronous/empty or the closer goroutine left blocked — the remote reads its preamble, Wait closes stdin, the process exits |
| `a575e466` | `run.go:138` | statement/remove | Removes a statement entirely, dropping its side effect or return value | equivalent: preamble write synchronous/empty or the closer goroutine left blocked — the remote reads its preamble, Wait closes stdin, the process exits |
| `69eb783c` | `run.go:103` | statement/return | Replaces a return value with its zero value | equivalent/contract-open: Run then always reads the status record (one more ssh, X6 names it only for 255) or the signal path ends in the same C3 status |
| `b05447bb` | `run.go:119` | statement/return | Replaces a return value with its zero value | equivalent: StdinPipe/Start of bash/ssh (X7 requirements) + the Linux WaitStatus assertion cannot fail; a value beside an error is never read |
| `39e78138` | `run.go:122` | statement/return | Replaces a return value with its zero value | equivalent: StdinPipe/Start of bash/ssh (X7 requirements) + the Linux WaitStatus assertion cannot fail; a value beside an error is never read |
| `0bc2b582` | `run.go:144` | statement/return | Replaces a return value with its zero value | killed: TestSuperviseReportsAKilledProcessAs128PlusN (exact diff confirmed) |

## slots

221 mutants, 158 killed by unit tests, 63 escaped; replay (S1–S6/R2/L3/K5/J1/ReviewR2/S4/R1R2/P2) killed 16. A full-e2e re-replay under load 24–38 killed 2 more through unrelated K8/X3 timing failures = noise, not credited (survived rows below).

| id | site | mutator | change | ruling |
|----|------|---------|--------|--------|
| `7826ce37` | `slots.go:253` | arithmetic/base | Changes: `case <-time.After(200 * time.Millisecond):` → `case <-time.After(200 / time.Millisecond):` | contract-open: the 200 ms poll interval (K5/J1 name none) |
| `696e3513` | `slots.go:169` | arithmetic/bitwise | Changes: `f, err := os.OpenFile(path, os.O_RDONLY\|syscall.O_NOFOLLOW\|syscall.O_NONBLOCK, 0) //nolint:gosec / | killed: TestR2HoldersNeverFollowOrWaitOnSlotFiles (exact diff confirmed) |
| `ff9e495e` | `slots.go:169` | arithmetic/bitwise | Changes: `f, err := os.OpenFile(path, os.O_RDONLY\|syscall.O_NOFOLLOW\|syscall.O_NONBLOCK, 0) //nolint:gosec / | killed: TestR2HoldersNeverFollowOrWaitOnSlotFiles (exact diff confirmed) |
| `a9f9b31f` | `slots.go:90` | branch/if | Removes an if-block body so the condition becomes a no-op, keeping a zero-value return when the branch was ter | equivalent (rc): an ignored Mkdir/Lstat error ends in the next Mkdir/Lstat/open failing (125, other wording); Lstat right after a Mkdir of our own 0700 dir cannot fail |
| `f88d2609` | `slots.go:96` | branch/if | Removes an if-block body so the condition becomes a no-op, keeping a zero-value return when the branch was ter | equivalent (rc): an ignored Mkdir/Lstat error ends in the next Mkdir/Lstat/open failing (125, other wording); Lstat right after a Mkdir of our own 0700 dir cannot fail |
| `e48abd4f` | `slots.go:100` | branch/if | Removes an if-block body so the condition becomes a no-op, keeping a zero-value return when the branch was ter | equivalent (rc): an ignored Mkdir/Lstat error ends in the next Mkdir/Lstat/open failing (125, other wording); Lstat right after a Mkdir of our own 0700 dir cannot fail |
| `1a2a3b4c` | `slots.go:171` | branch/if | Removes an if-block body so the condition becomes a no-op, keeping a zero-value return when the branch was ter | equivalent: a nil file Fds/Stats as invalid ⇒ the same false/"" result; a FIFO owner read nonblocking returns "" |
| `7a6722a7` | `slots.go:182` | branch/if | Removes an if-block body so the condition becomes a no-op, keeping a zero-value return when the branch was ter | equivalent: a nil file Fds/Stats as invalid ⇒ the same false/"" result; a FIFO owner read nonblocking returns "" |
| `ec672191` | `slots.go:186` | branch/if | Removes an if-block body so the condition becomes a no-op, keeping a zero-value return when the branch was ter | equivalent: a nil file Fds/Stats as invalid ⇒ the same false/"" result; a FIFO owner read nonblocking returns "" |
| `f5f8e867` | `slots.go:198` | branch/if | Removes an if-block body so the condition becomes a no-op, keeping a zero-value return when the branch was ter | equivalent: a nil file after a failed open fails at the next call (WriteString / flock EBADF) ⇒ the same error result |
| `962c86d7` | `slots.go:202` | branch/if | Removes an if-block body so the condition becomes a no-op, keeping a zero-value return when the branch was ter | contract-open: a close error after a whole owner write (not injectable on the runtime fs) |
| `95c2eff4` | `slots.go:214` | branch/if | Removes an if-block body so the condition becomes a no-op, keeping a zero-value return when the branch was ter | equivalent: a nil file after a failed open fails at the next call (WriteString / flock EBADF) ⇒ the same error result |
| `8b6134b3` | `slots.go:238` | branch/if | Removes an if-block body so the condition becomes a no-op, keeping a zero-value return when the branch was ter | equivalent: a nil file after a failed open fails at the next call (WriteString / flock EBADF) ⇒ the same error result |
| `625e2519` | `slots.go:246` | branch/if | Removes an if-block body so the condition becomes a no-op, keeping a zero-value return when the branch was ter | equivalent (reachable domain): flock(LOCK_NB) on a freshly opened local runtime-dir fd fails only with EWOULDBLOCK (S2's I/O-error branch unreachable there) |
| `0955e49f` | `slots.go:262` | branch/if | Removes an if-block body so the condition becomes a no-op, keeping a zero-value return when the branch was ter | equivalent: TryLock's one caller (placer) has just run MakeDirs on the same hosts/<name> dir via TryAcquire |
| `e2b2929f` | `slots.go:71` | composite/field-clear | Changes: `return Dir{Base: base, Path: filepath.Join(base, "heavy"), Total: n}, nil` → `return Dir{Path: filep | equivalent: an empty Base = the Path's parent = the runtime dir (TryAcquire) |
| `e2e10fdf` | `slots.go:89` | expression/error-guard | Removes an error-guard block so the function continues even on error | equivalent (rc): an ignored Mkdir/Lstat error ends in the next Mkdir/Lstat/open failing (125, other wording); Lstat right after a Mkdir of our own 0700 dir cannot fail |
| `490172ac` | `slots.go:99` | expression/error-guard | Removes an error-guard block so the function continues even on error | equivalent (rc): an ignored Mkdir/Lstat error ends in the next Mkdir/Lstat/open failing (125, other wording); Lstat right after a Mkdir of our own 0700 dir cannot fail |
| `e562d372` | `slots.go:197` | expression/error-guard | Removes an error-guard block so the function continues even on error | equivalent: a nil file after a failed open fails at the next call (WriteString / flock EBADF) ⇒ the same error result |
| `fac63cd4` | `slots.go:213` | expression/error-guard | Removes an error-guard block so the function continues even on error | equivalent: a nil file after a failed open fails at the next call (WriteString / flock EBADF) ⇒ the same error result |
| `2e8cdf42` | `slots.go:237` | expression/error-guard | Removes an error-guard block so the function continues even on error | equivalent: a nil file after a failed open fails at the next call (WriteString / flock EBADF) ⇒ the same error result |
| `8f3b8f92` | `slots.go:261` | expression/error-guard | Removes an error-guard block so the function continues even on error | equivalent: TryLock's one caller (placer) has just run MakeDirs on the same hosts/<name> dir via TryAcquire |
| `52e8ea29` | `slots.go:221` | expression/errorf-wrap | Changes: `return nil, false, fmt.Errorf("flock %s: %w", path, err)` → `return nil, false, fmt.Errorf("flock %s | contract-open: error wrapping/wording |
| `8d5026ab` | `slots.go:247` | expression/errorf-wrap | Changes: `return nil, fmt.Errorf("flock %s: %w", path, err)` → `return nil, fmt.Errorf("flock %s: %v", path, e | contract-open: error wrapping/wording |
| `11b554c5` | `slots.go:103` | expression/logical | Changes: `if !info.IsDir() \|\| !ok \|\| int(st.Uid) != os.Geteuid() \|\| info.Mode().Perm()&0o022 != 0 {` → ` | equivalent (rc): a non-directory component passes the IsDir term alone only when ours + 0600 — then the next Mkdir/open fails ENOTDIR (125); a symlink fails the 0o022 term (mode 0777) |
| `d3323d98` | `slots.go:185` | expression/logical | Changes: `if info, err := f.Stat(); err != nil \|\| !info.Mode().IsRegular() {` → `if info, err := f.Stat(); e | equivalent: a nil file Fds/Stats as invalid ⇒ the same false/"" result; a FIFO owner read nonblocking returns "" |
| `d9731b99` | `slots.go:58` | expression/remove | Changes: `if err != nil \|\| n == 0 {` → `if err != nil \|\| false {` | equivalent: a count beside an error is never read; an all-zero HEAVY_SLOTS already fails Atoi("") |
| `4b600c66` | `slots.go:103` | expression/remove | Removes an expression statement entirely, dropping its side effect | equivalent (rc): a non-directory component passes the IsDir term alone only when ours + 0600 — then the next Mkdir/open fails ENOTDIR (125); a symlink fails the 0o022 term (mode 0777) |
| `e324146d` | `slots.go:54` | numbers/decrementer | Changes: `return 0, errors.New("HEAVY_SLOTS must be a positive integer")` → `return (-1), errors.New("HEAVY_SL | equivalent: a count beside an error is never read; an all-zero HEAVY_SLOTS already fails Atoi("") |
| `fced9218` | `slots.go:58` | numbers/decrementer | Changes: `if err != nil \|\| n == 0 {` → `if err != nil \|\| n == (-1) {` | equivalent: a count beside an error is never read; an all-zero HEAVY_SLOTS already fails Atoi("") |
| `c2dbd97c` | `slots.go:59` | numbers/decrementer | Changes: `return 0, errors.New("HEAVY_SLOTS must be a positive shell integer")` → `return (-1), errors.New("HE | equivalent: a count beside an error is never read; an all-zero HEAVY_SLOTS already fails Atoi("") |
| `6a18c1bf` | `slots.go:253` | numbers/decrementer | Changes: `case <-time.After(200 * time.Millisecond):` → `case <-time.After(199 * time.Millisecond):` | contract-open: the 200 ms poll interval (K5/J1 name none) |
| `79dd2415` | `slots.go:54` | numbers/incrementer | Changes: `return 0, errors.New("HEAVY_SLOTS must be a positive integer")` → `return 1, errors.New("HEAVY_SLOTS | equivalent: a count beside an error is never read; an all-zero HEAVY_SLOTS already fails Atoi("") |
| `f0805927` | `slots.go:59` | numbers/incrementer | Changes: `return 0, errors.New("HEAVY_SLOTS must be a positive shell integer")` → `return 1, errors.New("HEAVY | equivalent: a count beside an error is never read; an all-zero HEAVY_SLOTS already fails Atoi("") |
| `2742d1ff` | `slots.go:169` | numbers/incrementer | Changes: `f, err := os.OpenFile(path, os.O_RDONLY\|syscall.O_NOFOLLOW\|syscall.O_NONBLOCK, 0) //nolint:gosec / | equivalent: the perm argument of a non-creating open is unused |
| `1e44dd8b` | `slots.go:180` | numbers/incrementer | Changes: `f, err := os.OpenFile(path, os.O_RDONLY\|syscall.O_NOFOLLOW\|syscall.O_NONBLOCK, 0) //nolint:gosec / | equivalent: the perm argument of a non-creating open is unused |
| `fc377d6a` | `slots.go:196` | numbers/incrementer | Changes: `f, err := os.OpenFile(path, os.O_WRONLY\|os.O_CREATE\|os.O_TRUNC\|syscall.O_NOFOLLOW\|syscall.O_NONB | contract-open: the mode of a slot/owner file (S2/S3 name none; heavy's `>` gives 0644) |
| `94660610` | `slots.go:212` | numbers/incrementer | Changes: `f, err = os.OpenFile(path, os.O_WRONLY\|os.O_CREATE\|syscall.O_NOFOLLOW\|syscall.O_NONBLOCK, 0o600)  | contract-open: the mode of a slot/owner file (S2/S3 name none; heavy's `>` gives 0644) |
| `936c532c` | `slots.go:253` | numbers/incrementer | Changes: `case <-time.After(200 * time.Millisecond):` → `case <-time.After(201 * time.Millisecond):` | contract-open: the 200 ms poll interval (K5/J1 name none) |
| `adfa7180` | `slots.go:202` | statement/remove | Removes a statement entirely, dropping its side effect or return value | contract-open: a close error after a whole owner write (not injectable on the runtime fs) |
| `9345741f` | `slots.go:96` | statement/return | Replaces a return value with its zero value | equivalent (rc): an ignored Mkdir/Lstat error ends in the next Mkdir/Lstat/open failing (125, other wording); Lstat right after a Mkdir of our own 0700 dir cannot fail |
| `6e8a9706` | `slots.go:100` | statement/return | Replaces a return value with its zero value | equivalent (rc): an ignored Mkdir/Lstat error ends in the next Mkdir/Lstat/open failing (125, other wording); Lstat right after a Mkdir of our own 0700 dir cannot fail |
| `ac5f6797` | `slots.go:221` | statement/return | Replaces a return value with its zero value | equivalent (reachable domain): flock(LOCK_NB) on a freshly opened local runtime-dir fd fails only with EWOULDBLOCK (S2's I/O-error branch unreachable there) |
| `29dabddc` | `slots.go:227` | statement/return | Replaces a return value with its zero value | killed: e2e TestX1ControllerDeathKillsSession (S5; exact diff confirmed, control passes) |
| `2232af85` | `slots.go:247` | statement/return | Replaces a return value with its zero value | equivalent (reachable domain): flock(LOCK_NB) on a freshly opened local runtime-dir fd fails only with EWOULDBLOCK (S2's I/O-error branch unreachable there) |
| `96ffcc03` | `slots.go:252` | statement/return | Replaces a return value with its zero value | killed: TestWaitLockEndsWithItsContext (exact diff confirmed) |
| `45b9e900` | `slots.go:262` | statement/return | Replaces a return value with its zero value | equivalent: TryLock's one caller (placer) has just run MakeDirs on the same hosts/<name> dir via TryAcquire |

## pooladdrs

13 mutants (main.go); before c44837d: 3 killed, 10 escaped; after the new run-error + main exit/message tests: 12 killed (mutago rerun).

| id | site | mutator | change | ruling |
|----|------|---------|--------|--------|
| `a1666792` | `main.go:16` | numbers/incrementer | `os.Exit(1)` → `os.Exit(2)` | contract-open: install needs a nonzero status (`go run` reports exit 1 for any) |

