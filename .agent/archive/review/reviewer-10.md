# reviewer-10 — survivor rulings, `internal/cli/cli.go` (sweep at 09e9969, CI replay run 40534349336)

| id | site | func | mutator | change | replay | ruling |
|----|------|------|---------|--------|--------|--------|
| `73ad10cd` | `cli.go:106` | Main | numbers/decrementer | Changes: `return ctr.Init(args[1:], stderr)` → `return ctr.Init(args[0:], stderr)` | 1 covering tests: pass | killed: TestK7StatusPassesThrough — argv[0] becomes __ctr-init; rc 127 instead of 0 (red twice; baseline green). |
| `e0d93f38` | `cli.go:106` | Main | numbers/incrementer | Changes: `return ctr.Init(args[1:], stderr)` → `return ctr.Init(args[2:], stderr)` | 1 covering tests: pass | killed: TestK7StatusPassesThrough — argv[0] becomes exit; rc 127 instead of 0 (red twice; baseline green). |
| `3db37409` | `cli.go:127` | parseRun | statement/return | `- return o, err`; `- }`; `+ return options{}, err`; `+ }` | 2 covering tests: pass | equivalent: Main ignores parseRun options whenever err != nil (cli.go:95–101). |
| `59f2f318` | `cli.go:131` | parseRun | statement/return | `- return o, errors.New("give a command after `run`, e.g. `offload run -- make -j8`")`; `- }`; `+ return options{}, errors.New("give a command after `run`, e.g. `offload run -- make -j8`")`; `+ }` | 1 covering tests: pass | equivalent: Main ignores parseRun options whenever err != nil (cli.go:95–101). |
| `9f816eb1` | `cli.go:135` | parseRun | statement/return | `- return o, errors.New("--local excludes --pool, --host, --gpu, --no-sync and --get")`; `- }`; `+ return options{}, errors.New("--local excludes --pool, --host, --gpu, --no-sync and --get")`; `+ }` | 2 covering tests: pass | equivalent: Main ignores parseRun options whenever err != nil (cli.go:95–101). |
| `1a092486` | `cli.go:166` | runCommand | composite/field-clear | `- Local: o.local, Pool: o.pool, Host: o.host, GPU: o.gpu, NoSync: o.noSync, Get: o.gets,`; `- }, stderr)`; `+ Local:	o.local, Pool: o.pool, Host: o.host, NoSync: o.noSync, Get: o.gets,`; `+ }, stderr)` | 104 covering tests: pass | new test: J2 — gpu=true disappears from logged options; `TestJ2HostAndGPUOptionsRecorded` (red ×2, baseline green). |
| `f645377f` | `cli.go:166` | runCommand | composite/field-clear | `- Local: o.local, Pool: o.pool, Host: o.host, GPU: o.gpu, NoSync: o.noSync, Get: o.gets,`; `- }, stderr)`; `+ Local:	o.local, Pool: o.pool, GPU: o.gpu, NoSync: o.noSync, Get: o.gets,`; `+ }, stderr)` | 104 covering tests: pass | new test: J2 — host=alpha disappears from logged options; `TestJ2HostAndGPUOptionsRecorded` (red ×2, baseline green). |
| `d3b544cf` | `cli.go:183` | runLogged | expression/error-guard | `- if ctx.Err() != nil {`; `- return signalStatus()`; `- }`; `+ if false {`; `+ return signalStatus()`; `+ }` | 22 covering tests: pass | contract-open: cancellation may add an error diagnostic; C3/J2 status still fixed by runCommand End callback (cli.go:170–175); C4 fixes channel/prefix, not this message. |
| `22915288` | `cli.go:184` | runLogged | branch/if | `- return signalStatus()`; `- }`; `+ _ = signalStatus`; `+ }` | 2 covering tests: pass | contract-open: cancellation may add an error diagnostic; C3/J2 status still fixed by runCommand End callback (cli.go:170–175); C4 fixes channel/prefix, not this message. |
| `fbf1fd69` | `cli.go:184` | runLogged | statement/return | `- return signalStatus()`; `- }`; `+ return 0`; `+ }` | 2 covering tests: pass | equivalent: only canceled contexts return here; runCommand End overrides this intermediate rc with the signal status (cli.go:170–175). |
| `47db9207` | `cli.go:189` | runLogged | expression/error-guard | `- if err != nil {`; `- return fail(err)`; `- }`; `+ if false {`; `+ return fail(err)`; `+ }` | 104 covering tests: pass | new test: C3 — removed cwd becomes a successful local run; `TestC3RemovedCwdFailsBeforeCommand` (red ×2, baseline green). |
| `2f05fb0f` | `cli.go:190` | runLogged | statement/return | `- return fail(err)`; `- }`; `+ return 0`; `+ }` | no e2e test executes the line | new test: C3 — removed cwd returns zero; `TestC3RemovedCwdFailsBeforeCommand` (red ×2, baseline green). |
| `e4012e60` | `cli.go:190` | runLogged | branch/if | `- return fail(err)`; `- }`; `+ _, _ = fail, err`; `+ }` | no e2e test executes the line | new test: C3 — removed cwd becomes a successful local run; `TestC3RemovedCwdFailsBeforeCommand` (red ×2, baseline green). |
| `83dec61a` | `cli.go:208` | runLogged | statement/remove | `- fmt.Fprintf(stderr, "offload: no host named %s in %s\n", o.host, poolPath)`; `- return ExitUsage`; `- }`; `+ _, _, _, _ = fmt.Fprintf, stderr, o.host, poolPath`; `+ return ExitUsage`; `+ }` | 1 covering tests: pass | contract-open: unknown-host diagnostic presence (C3 fixes rc 2 before work; C4 fixes existing messages, not a required unknown-host message). |
| `18039d5e` | `cli.go:213` | runLogged | composite/field-clear | Changes: `t := mirror.Tree{Root: cwd, Sub: "."}` → `t := mirror.Tree{Root: cwd}` | 102 covering tests: pass | equivalent: default Tree survives only without hosts; remote work never uses it, --get implies pool with hosts, local execution ignores Sub (cli.go:213–234,267). |
| `99818e7e` | `cli.go:213` | runLogged | composite/field-clear | Changes: `t := mirror.Tree{Root: cwd, Sub: "."}` → `t := mirror.Tree{Sub: "."}` | 102 covering tests: pass | equivalent: default Tree survives only without hosts; local placement/execution ignores Root, and J2 root stays null until Inspect (cli.go:213–218,267). |
| `7352283d` | `cli.go:219` | runLogged | expression/remove | Changes: `if t.Git && t.LocalOnly == "" { // nested repos surface in the file list (Y3, before remote work)` → `if t.Git && true { // nested repos surface in the file list (Y3, before remote work)` | 83 covering tests: pass | new test: Y3 — local-only repository rejected by irrelevant upload-list failure; `TestY3LocalOnlyTreeSkipsUploadFileChecks` (red ×2, baseline green). |
| `c876ff51` | `cli.go:222` | runLogged | expression/error-guard | `- } else if err != nil {`; `- return fail(err)`; `- }`; `+ } else if false {`; `+ return fail(err)`; `+ }` | 60 covering tests: pass | new test: Y4 — non-ENOENT upload-list failure falls back to successful local command; `TestY4UnreadableUploadEntryIsFailure` (red ×2, baseline green). |
| `747be994` | `cli.go:223` | runLogged | branch/if | `- return fail(err)`; `- }`; `+ _, _ = fail, err`; `+ }` | no e2e test executes the line | new test: Y4 — non-ENOENT upload-list failure falls back to successful local command; `TestY4UnreadableUploadEntryIsFailure` (red ×2, baseline green). |
| `f70dc17a` | `cli.go:223` | runLogged | statement/return | `- return fail(err)`; `- }`; `+ return 0`; `+ }` | no e2e test executes the line | new test: Y4 — non-ENOENT upload-list failure returns zero; `TestY4UnreadableUploadEntryIsFailure` (red ×2, baseline green). |
| `7285aa62` | `cli.go:237` | runLogged | statement/remove | `- fmt.Fprintf(stderr, "offload: %v\n", err)`; `- return ExitUsage`; `- }`; `+ _, _, _ = fmt.Fprintf, stderr, err`; `+ return ExitUsage`; `+ }` | 1 covering tests: pass | contract-open: invalid --get diagnostic presence (Y8 fixes refusal + rc 2, not a required diagnostic; C4 fixes only existing messages). |
| `b6f9dabe` | `cli.go:245` | runLogged | expression/remove | `- if o.local \|\| len(hosts) == 0 {`; `+ if false \|\|`; `+ len(hosts) == 0 {` | 4 covering tests: pass | equivalent: o.local implies hosts empty: pool.Load is inside !o.local and no later assignment adds hosts (cli.go:193–209). |
| `2ddb8b92` | `cli.go:255` | runLogged | expression/error-guard | `- if err != nil {`; `- fmt.Fprintf(stderr, "offload: %v\n", err)`; `- return ExitUsage`; `- }`; `+ if false {`; `+ fmt.Fprintf(stderr, "offload: %v\n", err)`; `+ return ExitUsage`; `+ }` | 98 covering tests: pass | new test: S1 — invalid HEAVY_SLOTS accepted for pool run; `TestS1InvalidHeavySlotsRejectBeforePlacement` (red ×2, baseline green). |
| `346b2157` | `cli.go:256` | runLogged | branch/if | `- fmt.Fprintf(stderr, "offload: %v\n", err)`; `- return ExitUsage`; `- }`; `+ _, _, _ = fmt.Fprintf, stderr, err`; `+ `; `+ }` | no e2e test executes the line | new test: S1 — invalid HEAVY_SLOTS accepted for pool run; `TestS1InvalidHeavySlotsRejectBeforePlacement` (red ×2, baseline green). |
| `e240a70b` | `cli.go:256` | runLogged | statement/remove | `- fmt.Fprintf(stderr, "offload: %v\n", err)`; `- return ExitUsage`; `- }`; `+ _, _, _ = fmt.Fprintf, stderr, err`; `+ return ExitUsage`; `+ }` | no e2e test executes the line | contract-open: invalid HEAVY_SLOTS diagnostic presence (S1 fixes acceptance/rejection; C4 fixes only existing messages). |
| `1356d186` | `cli.go:257` | runLogged | statement/return | `- fmt.Fprintf(stderr, "offload: %v\n", err)`; `- return ExitUsage`; `- }`; `+ fmt.Fprintf(stderr, "offload: %v\n", err)`; `+ return 0`; `+ }` | no e2e test executes the line | new test: S1 — invalid HEAVY_SLOTS returns zero; `TestS1InvalidHeavySlotsRejectBeforePlacement` (red ×2, baseline green). |
| `942294c3` | `cli.go:263` | runLogged | statement/return | Changes: `return signalStatus()` → `return 0` | 6 covering tests: pass | equivalent: canceled context reaches runCommand End, which fixes signal status independently of runLogged rc (cli.go:170–175). |
| `d04f40ae` | `cli.go:275` | runLogged | expression/error-guard | `- if runErr == nil {`; `- runlog.CommandStatus(rc)`; `- }`; `+ if true {`; `+ runlog.CommandStatus(rc)`; `+ }` | 29 covering tests: pass | new test: J2 — watcher startup error logs fabricated cmd_rc=0; `TestC3LocalWatcherStartFailureIsNotCommandSuccess` (red ×2, baseline green). |
| `ef66706a` | `cli.go:278` | runLogged | expression/error-guard | `- if ctx.Err() != nil {`; `- return signalStatus()`; `- }`; `+ if false {`; `+ return signalStatus()`; `+ }` | 29 covering tests: pass | equivalent: remaining local branch returns rc or fail(runErr); runCommand End overrides either under cancellation (cli.go:170–175,281–284). |
| `31e899ef` | `cli.go:279` | runLogged | branch/if | `- return signalStatus()`; `- }`; `+ _ = signalStatus`; `+ }` | 5 covering tests: pass | equivalent: remaining local branch returns rc or fail(runErr); runCommand End overrides either under cancellation (cli.go:170–175,281–284). |
| `9d065a5b` | `cli.go:279` | runLogged | statement/return | `- return signalStatus()`; `- }`; `+ return 0`; `+ }` | 5 covering tests: pass | equivalent: canceled context reaches runCommand End, which overrides intermediate zero with signal status (cli.go:170–175). |
| `97d7d9a4` | `cli.go:281` | runLogged | expression/error-guard | `- if runErr != nil {`; `- return fail(runErr)`; `- }`; `+ if false {`; `+ return fail(runErr)`; `+ }` | 24 covering tests: pass | new test: C3 — missing watcher returns zero instead of offload failure; `TestC3LocalWatcherStartFailureIsNotCommandSuccess` (red ×2, baseline green). |
| `455035b4` | `cli.go:282` | runLogged | statement/return | `- return fail(runErr)`; `- }`; `+ return 0`; `+ }` | no e2e test executes the line | new test: C3 — missing watcher returns zero instead of offload failure; `TestC3LocalWatcherStartFailureIsNotCommandSuccess` (red ×2, baseline green). |
| `50586361` | `cli.go:282` | runLogged | branch/if | `- return fail(runErr)`; `- }`; `+ _, _ = fail, runErr`; `+ }` | no e2e test executes the line | new test: C3 — missing watcher returns zero instead of offload failure; `TestC3LocalWatcherStartFailureIsNotCommandSuccess` (red ×2, baseline green). |
| `b6aa04bc` | `cli.go:294` | runLogged | expression/error-guard | `- if ctx.Err() != nil {`; `- return signalStatus()`; `- }`; `+ if false {`; `+ return signalStatus()`; `+ }` | 73 covering tests: pass | equivalent: remaining remote branch returns rc or fail(err); runCommand End overrides either under cancellation (cli.go:170–175,297–300). |
| `46550e86` | `cli.go:295` | runLogged | branch/if | `- return signalStatus()`; `- }`; `+ _ = signalStatus`; `+ }` | 7 covering tests: pass | equivalent: remaining remote branch returns rc or fail(err); runCommand End overrides either under cancellation (cli.go:170–175,297–300). |
| `a7614ce7` | `cli.go:295` | runLogged | statement/return | `- return signalStatus()`; `- }`; `+ return 0`; `+ }` | 7 covering tests: pass | equivalent: canceled context reaches runCommand End, which overrides intermediate zero with signal status (cli.go:170–175). |
| `11258585` | `cli.go:307` | remoteRun | numbers/incrementer | Changes: `return 0, err` → `return 1, err` | 7 covering tests: pass | equivalent: upload failure returns nonnil err; sole caller discards rc on err, and no CommandStatus call ran (cli.go:293–300,306–308). |
| `b0710ce4` | `cli.go:307` | remoteRun | numbers/decrementer | Changes: `return 0, err` → `return (-1), err` | 7 covering tests: pass | equivalent: upload failure returns nonnil err; sole caller discards rc on err, and no CommandStatus call ran (cli.go:293–300,306–308). |
| `0c6a2fe0` | `cli.go:314` | remoteRun | numbers/incrementer | Changes: `if err == nil && (ctx.Err() == nil \|\| (rc != ExitFailure && rc != 255)) {` → `if err == nil && (ctx.Err() == nil \|\| (rc != ExitFailure && rc != 256)) {` | 70 covering tests: pass | new test: J2 — canceled remote 255 is recorded as a known command status; `TestJ2SignalClassifiesRemoteStatus` (red ×2, baseline green). |
| `8382e91e` | `cli.go:314` | remoteRun | expression/remove | `- if err == nil && (ctx.Err() == nil \|\| (rc != ExitFailure && rc != 255)) {`; `+ if true &&`; `+ (ctx.Err() == nil \|\| (rc != ExitFailure && rc != 255)) {` | 70 covering tests: pass | new test: J2 — failed remote setup is recorded as command status 255; `TestJ2RemoteSetupFailureHasNoCommandStatus` (red ×2, baseline green). |
| `8ce0aab5` | `cli.go:314` | remoteRun | expression/remove | Changes: `if err == nil && (ctx.Err() == nil \|\| (rc != ExitFailure && rc != 255)) {` → `if err == nil && true {` | 70 covering tests: pass | new test: J2 — canceled remote 125/255 are recorded as known command statuses; `TestJ2SignalClassifiesRemoteStatus` (red ×2, baseline green). |
| `90bd0cd6` | `cli.go:314` | remoteRun | numbers/decrementer | Changes: `if err == nil && (ctx.Err() == nil \|\| (rc != ExitFailure && rc != 255)) {` → `if err == nil && (ctx.Err() == nil \|\| (rc != ExitFailure && rc != 254)) {` | 70 covering tests: pass | new test: J2 — canceled 254 becomes unknown and canceled 255 becomes known; `TestJ2SignalClassifiesRemoteStatus` (red ×2, baseline green). |
| `c8e19e0c` | `cli.go:314` | remoteRun | expression/remove | Changes: `if err == nil && (ctx.Err() == nil \|\| (rc != ExitFailure && rc != 255)) {` → `if err == nil && (ctx.Err() == nil \|\| (rc != ExitFailure && true)) {` | 70 covering tests: pass | new test: J2 — canceled remote 255 is recorded as a known command status; `TestJ2SignalClassifiesRemoteStatus` (red ×2, baseline green). |
| `c9974722` | `cli.go:314` | remoteRun | expression/logical | Changes: `if err == nil && (ctx.Err() == nil \|\| (rc != ExitFailure && rc != 255)) {` → `if err == nil \|\| (ctx.Err() == nil \|\| (rc != ExitFailure && rc != 255)) {` | 70 covering tests: pass | new test: J2 — failed remote setup is recorded as command status 255; `TestJ2RemoteSetupFailureHasNoCommandStatus` (red ×2, baseline green). |
| `e8764565` | `cli.go:314` | remoteRun | expression/logical | Changes: `if err == nil && (ctx.Err() == nil \|\| (rc != ExitFailure && rc != 255)) {` → `if err == nil && (ctx.Err() == nil \|\| (rc != ExitFailure \|\| rc != 255)) {` | 70 covering tests: pass | new test: J2 — canceled remote 125/255 are recorded as known command statuses; `TestJ2SignalClassifiesRemoteStatus` (red ×2, baseline green). |
| `fcf62b80` | `cli.go:314` | remoteRun | expression/remove | `- if err == nil && (ctx.Err() == nil \|\| (rc != ExitFailure && rc != 255)) {`; `+ if err == nil && (ctx.Err() == nil \|\| (true &&`; `+ rc != 255)) {` | 70 covering tests: pass | new test: J2 — canceled remote 125 is recorded as a known command status; `TestJ2SignalClassifiesRemoteStatus` (red ×2, baseline green). |
| `f87d9cad` | `cli.go:317` | remoteRun | expression/remove | Changes: `if err != nil \|\| ctx.Err() != nil {` → `if err != nil \|\| false {` | 70 covering tests: pass | new test: J2/Y10 — interrupted run starts a download phase instead of skipping it; `TestJ2SignalClassifiesRemoteStatus` (red ×2, baseline green). |
| `38168288` | `cli.go:318` | remoteRun | statement/return | `- return rc, err`; `- }`; `+ return 0, err`; `+ }` | 10 covering tests: pass | equivalent: return reached only on err or cancellation; runLogged discards rc in both cases, after remote cmd_rc was recorded separately (cli.go:293–300,314–318). |
| `65f3552b` | `cli.go:331` | remoteRun | expression/error-guard | `- if err != nil {`; `- break`; `- }`; `+ if false {`; `+ break`; `+ }` | 8 covering tests: pass | new test: Y10 — later successful get overwrites prior failure and returns zero; `TestY10EarlierGetFailureSurvivesLaterSuccess` (red ×2, baseline green). |
| `046af0a8` | `cli.go:332` | remoteRun | loop/break | `- if err != nil {`; `- break`; `- }`; `- err = mirror.Get(ctx, t, target, g)`; `+ if err != nil {`; `+ continue`; `+ err = mirror.Get(ctx, t, target, g)`; `+ }` | no e2e test executes the line | equivalent: continue skips the remaining --get iterations with unchanged nonnil err; break exits them immediately; neither calls Get again (cli.go:330–335). |
| `ae6aabf6` | `cli.go:332` | remoteRun | branch/if | `- break`; `- }`; `+ `; `+ }` | no e2e test executes the line | new test: Y10 — later successful get overwrites prior failure and returns zero; `TestY10EarlierGetFailureSurvivesLaterSuccess` (red ×2, baseline green). |
| `8f9927ae` | `cli.go:337` | remoteRun | conditional/negated | Changes: `if err != nil {` → `if err == nil {` | 58 covering tests: pass | contract-open: extra successful-download/missing failed-download summary; Y10 still gets tool message + rc 125 from fail(err); no predicate fixes this additional summary. |
| `c75feee2` | `cli.go:337` | remoteRun | expression/error-guard | `- if err != nil {`; `- fmt.Fprintf(stderr, "offload: the command exited %d, but its outputs did not all come back\n", rc)`; `- }`; `+ if false {`; `+ fmt.Fprintf(stderr, "offload: the command exited %d, but its outputs did not all come back\n", rc)`; `+ }` | 58 covering tests: pass | contract-open: failed-download summary omitted; Y10 still gets tool message + rc 125 from fail(err); no predicate requires this additional summary. |
| `03928744` | `cli.go:338` | remoteRun | branch/if | `- fmt.Fprintf(stderr, "offload: the command exited %d, but its outputs did not all come back\n", rc)`; `- }`; `+ _, _, _ = fmt.Fprintf, stderr, rc`; `+ }` | 6 covering tests: pass | contract-open: failed-download summary omitted; Y10 still gets tool message + rc 125 from fail(err); no predicate requires this additional summary. |
| `4a515e30` | `cli.go:338` | remoteRun | statement/remove | `- fmt.Fprintf(stderr, "offload: the command exited %d, but its outputs did not all come back\n", rc)`; `- }`; `+ _, _, _ = fmt.Fprintf, stderr, rc`; `+ }` | 6 covering tests: pass | contract-open: failed-download summary omitted; Y10 still gets tool message + rc 125 from fail(err); no predicate requires this additional summary. |
| `a1adcc80` | `cli.go:348` | upload | statement/defer-remove | Changes: `defer runlog.Time(ledger.Upload)()` → `runlog.Time(ledger.Upload)()` | 67 covering tests: pass | new test: J2 — upload duration excludes transfer time; `TestJ2UploadDurationIncludesTransfer` (red ×2, baseline green). |
| `78fea1e7` | `cli.go:358` | upload | expression/error-guard | `- if err != nil {`; `- return err`; `- }`; `+ if false {`; `+ return err`; `+ }` | 53 covering tests: pass | new test: Y4/Y10 — failed upload listing is ignored and command starts; `TestY4UploadRechecksEntriesAfterWaiting` (red ×2, baseline green). |
| `7a332d20` | `cli.go:359` | upload | statement/return | `- return err`; `- }`; `+ return nil`; `+ }` | 1 covering tests: pass | new test: Y4/Y10 — failed upload listing returns success and command starts; `TestY4UploadRechecksEntriesAfterWaiting` (red ×2, baseline green). |
| `d66b5564` | `cli.go:359` | upload | branch/if | `- return err`; `- }`; `+ _ = err`; `+ }` | 1 covering tests: pass | new test: Y4/Y10 — failed upload listing is ignored and command starts; `TestY4UploadRechecksEntriesAfterWaiting` (red ×2, baseline green). |
| `1ab3dd2a` | `cli.go:361` | upload | expression/error-guard | `- if r.Excludes, err = mirror.GlobalExcludes(ctx, t.Root); err != nil {`; `- return err`; `- }`; `+ if r.Excludes, err = mirror.GlobalExcludes(ctx, t.Root); false {`; `+ return err`; `+ }` | 53 covering tests: pass | new test: Y5/Y10 — unreadable excludes are ignored and command starts; `TestY5UnreadableGlobalExcludesStopsRun` (red ×2, baseline green). |
| `5ae6a6e3` | `cli.go:362` | upload | statement/return | `- return err`; `- }`; `+ return nil`; `+ }` | 1 covering tests: pass | new test: Y5/Y10 — unreadable excludes return success and command starts; `TestY5UnreadableGlobalExcludesStopsRun` (red ×2, baseline green). |
| `7f4a7b42` | `cli.go:362` | upload | branch/if | `- return err`; `- }`; `+ _ = err`; `+ }` | 1 covering tests: pass | new test: Y5/Y10 — unreadable excludes are ignored and command starts; `TestY5UnreadableGlobalExcludesStopsRun` (red ×2, baseline green). |
| `42bc02b4` | `cli.go:379` | lookLocal | branch/case | Changes: `return "", rc, fmt.Sprintf("offload: %s: command not found", name)` → `_, _, _ = rc, fmt.Sprintf, name` | 4 covering tests: pass | contract-open: failed lookup diagnostic loses/reformats command name; C4/L1 fix channel + complaint/pool-only notice, not exact wording; rc remains 127. |
| `ab95db35` | `cli.go:381` | lookLocal | statement/return | `- return "", rc, fmt.Sprintf("offload: %s: %v", path, err)`; `+ _ = err`; `+ return "", rc, ""` | 3 covering tests: pass | new test: L1/C4 — hybrid non-executable lookup loses its complaint and offload prefix; `TestL1NonExecutableLocalCommandExplainsPoolOnly` (red ×2, baseline green). |
| `d0d46da8` | `cli.go:381` | lookLocal | statement/return | Changes: `return "", rc, fmt.Sprintf("offload: %s: %v", path, err)` → `return "", 0, fmt.Sprintf("offload: %s: %v", path, err)` | 3 covering tests: pass | new test: X2 — non-executable local command takes a slot before returning 126; `TestX2NonExecutableLookupBeforeSlot` (red ×2, baseline green). |

## Notes

Baseline = `f0246b8204e893624d73b38dc8562a6a622eda08`; target source unchanged from sweep. Commands below run from `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-10` with `umask 022`. Each red uses the exact `agentic.json` diff, then restores `internal/cli/cli.go`.

### 73ad10cd / e0d93f38

Existing test: `go test -count=2 -run '^TestK7StatusPassesThrough$' ./test/e2e/` → rc 1 each, two failing runs each. `73ad10cd`: `ctr_test.go:434: exit = 127, want 0; stdout="" stderr="offload: __ctr-init: command not found\n"`. `e0d93f38`: same failing line, `stderr="offload: exit: command not found\n"`. Restored baseline command (`-count=1`) → rc 0, `ok github.com/eturkes/offload/test/e2e 35.295s`.

### 1a092486

`go test -count=2 -run '^TestJ2HostAndGPUOptionsRecorded$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:23: start.options = map[string]interface {}{"get":[]interface {}{}, "gpu":false, "host":"alpha", "local":false, "no_sync":true, "pool":true}, want map[string]interface {}{"get":[]interface {}{}, "gpu":true, "host":"alpha", "local":false, "no_sync":true, "pool":true}
reviewer10_mutation_test.go:24: end.options = map[string]interface {}{"get":[]interface {}{}, "gpu":false, "host":"alpha", "local":false, "no_sync":true, "pool":true}, want map[string]interface {}{"get":[]interface {}{}, "gpu":true, "host":"alpha", "local":false, "no_sync":true, "pool":true}
```

### f645377f

`go test -count=2 -run '^TestJ2HostAndGPUOptionsRecorded$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:23: start.options = map[string]interface {}{"get":[]interface {}{}, "gpu":true, "host":"", "local":false, "no_sync":true, "pool":true}, want map[string]interface {}{"get":[]interface {}{}, "gpu":true, "host":"alpha", "local":false, "no_sync":true, "pool":true}
reviewer10_mutation_test.go:24: end.options = map[string]interface {}{"get":[]interface {}{}, "gpu":true, "host":"", "local":false, "no_sync":true, "pool":true}, want map[string]interface {}{"get":[]interface {}{}, "gpu":true, "host":"alpha", "local":false, "no_sync":true, "pool":true}
```

### 47db9207

`go test -count=2 -run '^TestC3RemovedCwdFailsBeforeCommand$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:36: removed cwd: err=<nil> rc=0 output="offload: local slot 1/2\nshell-init: error retrieving current directory: getcwd: cannot access parent directories: No such file or directory\n", want 125
```

### 2f05fb0f

`go test -count=2 -run '^TestC3RemovedCwdFailsBeforeCommand$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:36: removed cwd: err=<nil> rc=0 output="", want 125
```

### e4012e60

`go test -count=2 -run '^TestC3RemovedCwdFailsBeforeCommand$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:36: removed cwd: err=<nil> rc=0 output="offload: local slot 1/2\nshell-init: error retrieving current directory: getcwd: cannot access parent directories: No such file or directory\n", want 125
```

### 7352283d

`go test -count=2 -run '^TestY3LocalOnlyTreeSkipsUploadFileChecks$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:65: exit = 125, want 0; stdout="" stderr="offload: lstat /tmp/TestY3LocalOnlyTreeSkipsUploadFileChecks2996244557/001/project/locked/file: permission denied\n"
reviewer10_mutation_test.go:65: exit = 125, want 0; stdout="" stderr="offload: lstat /tmp/TestY3LocalOnlyTreeSkipsUploadFileChecks4252172435/001/project/locked/file: permission denied\n"
```

### c876ff51

`go test -count=2 -run '^TestY4UnreadableUploadEntryIsFailure$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:55: exit = 0, want 125; stdout="" stderr="offload: local slot 1/2\n"
```

### 747be994

`go test -count=2 -run '^TestY4UnreadableUploadEntryIsFailure$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:55: exit = 0, want 125; stdout="" stderr="offload: local slot 1/2\n"
```

### f70dc17a

`go test -count=2 -run '^TestY4UnreadableUploadEntryIsFailure$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:55: exit = 0, want 125; stdout="" stderr=""
```

### 2ddb8b92

`go test -count=2 -run '^TestS1InvalidHeavySlotsRejectBeforePlacement$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:74: exit = 0, want 2; stdout="" stderr="offload: alpha slot 1/2 (8 CPUs, 25% busy) → ~/offload/project-1d1676ff\n"
reviewer10_mutation_test.go:74: exit = 0, want 2; stdout="" stderr="offload: alpha slot 1/2 (8 CPUs, 25% busy) → ~/offload/project-1f7081ef\n"
```

### 346b2157

`go test -count=2 -run '^TestS1InvalidHeavySlotsRejectBeforePlacement$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:74: exit = 0, want 2; stdout="" stderr="offload: alpha slot 1/2 (8 CPUs, 25% busy) → ~/offload/project-3d1588cc\n"
reviewer10_mutation_test.go:74: exit = 0, want 2; stdout="" stderr="offload: alpha slot 1/2 (8 CPUs, 25% busy) → ~/offload/project-665d62ce\n"
```

### 1356d186

`go test -count=2 -run '^TestS1InvalidHeavySlotsRejectBeforePlacement$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:74: exit = 0, want 2; stdout="" stderr="offload: HEAVY_SLOTS must be a positive shell integer\n"
```

### d04f40ae

`go test -count=2 -run '^TestC3LocalWatcherStartFailureIsNotCommandSuccess$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:84: end.cmd_rc = 0, want <nil>
```

### 97d7d9a4

`go test -count=2 -run '^TestC3LocalWatcherStartFailureIsNotCommandSuccess$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:82: exit = 0, want 125; stdout="" stderr="offload: local slot 1/2\n"
```

### 455035b4

`go test -count=2 -run '^TestC3LocalWatcherStartFailureIsNotCommandSuccess$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:82: exit = 0, want 125; stdout="" stderr="offload: local slot 1/2\n"
```

### 50586361

`go test -count=2 -run '^TestC3LocalWatcherStartFailureIsNotCommandSuccess$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:82: exit = 0, want 125; stdout="" stderr="offload: local slot 1/2\n"
```

### 0c6a2fe0

`go test -count=2 -run '^TestJ2SignalClassifiesRemoteStatus$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:110: end.cmd_rc = 255, want <nil>
```

### 8382e91e

`go test -count=2 -run '^TestJ2RemoteSetupFailureHasNoCommandStatus$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:93: end.cmd_rc = 255, want <nil>
```

### 8ce0aab5

`go test -count=2 -run '^TestJ2SignalClassifiesRemoteStatus$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:110: end.cmd_rc = 125, want <nil>
reviewer10_mutation_test.go:110: end.cmd_rc = 255, want <nil>
```

### 90bd0cd6

`go test -count=2 -run '^TestJ2SignalClassifiesRemoteStatus$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:110: end.cmd_rc = <nil>, want 254
reviewer10_mutation_test.go:110: end.cmd_rc = 255, want <nil>
```

### c8e19e0c

`go test -count=2 -run '^TestJ2SignalClassifiesRemoteStatus$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:110: end.cmd_rc = 255, want <nil>
```

### c9974722

`go test -count=2 -run '^TestJ2RemoteSetupFailureHasNoCommandStatus$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:93: end.cmd_rc = 255, want <nil>
```

### e8764565

`go test -count=2 -run '^TestJ2SignalClassifiesRemoteStatus$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:110: end.cmd_rc = 125, want <nil>
reviewer10_mutation_test.go:110: end.cmd_rc = 255, want <nil>
```

### fcf62b80

`go test -count=2 -run '^TestJ2SignalClassifiesRemoteStatus$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:110: end.cmd_rc = 125, want <nil>
```

### f87d9cad

`go test -count=2 -run '^TestJ2SignalClassifiesRemoteStatus$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:111: end.download = 0, want <nil>
```

### 65f3552b

`go test -count=2 -run '^TestY10EarlierGetFailureSurvivesLaterSuccess$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:119: exit = 0, want 125; stdout="" stderr="offload: alpha slot 1/2 (8 CPUs, 25% busy) → ~/offload/project-33ffc449\nrsync: [sender] link_stat \"/tmp/TestY10EarlierGetFailureSurvivesLaterSuccess3512709816/001/hosts/alpha/offload/project-33ffc449/missing\" failed: No such file or directory (2)\nrsync error: some files/attrs were not transferred (see previous errors) (code 23) at main.c(1983) [Receiver=3.5.1-g04355d27]\nrsync: [Receiver] write error: Broken pipe (32)\n"
reviewer10_mutation_test.go:119: exit = 0, want 125; stdout="" stderr="offload: alpha slot 1/2 (8 CPUs, 25% busy) → ~/offload/project-7bec2a59\nrsync: [sender] link_stat \"/tmp/TestY10EarlierGetFailureSurvivesLaterSuccess1384727812/001/hosts/alpha/offload/project-7bec2a59/missing\" failed: No such file or directory (2)\nrsync error: some files/attrs were not transferred (see previous errors) (code 23) at main.c(1983) [Receiver=3.5.1-g04355d27]\nrsync: [Receiver] write error: Broken pipe (32)\n"
```

### ae6aabf6

`go test -count=2 -run '^TestY10EarlierGetFailureSurvivesLaterSuccess$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:119: exit = 0, want 125; stdout="" stderr="offload: alpha slot 1/2 (8 CPUs, 25% busy) → ~/offload/project-17d3168a\nrsync: [sender] link_stat \"/tmp/TestY10EarlierGetFailureSurvivesLaterSuccess907664406/001/hosts/alpha/offload/project-17d3168a/missing\" failed: No such file or directory (2)\nrsync error: some files/attrs were not transferred (see previous errors) (code 23) at main.c(1983) [Receiver=3.5.1-g04355d27]\nrsync: [Receiver] write error: Broken pipe (32)\n"
reviewer10_mutation_test.go:119: exit = 0, want 125; stdout="" stderr="offload: alpha slot 1/2 (8 CPUs, 25% busy) → ~/offload/project-f37e175a\nrsync: [sender] link_stat \"/tmp/TestY10EarlierGetFailureSurvivesLaterSuccess2469921109/001/hosts/alpha/offload/project-f37e175a/missing\" failed: No such file or directory (2)\nrsync error: some files/attrs were not transferred (see previous errors) (code 23) at main.c(1983) [Receiver=3.5.1-g04355d27]\nrsync: [Receiver] write error: Broken pipe (32)\n"
```

### a1adcc80

`go test -count=2 -run '^TestJ2UploadDurationIncludesTransfer$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:199: upload=0, want at least the transfer's 0.1 seconds
```

### 78fea1e7

`go test -count=2 -run '^TestY4UploadRechecksEntriesAfterWaiting$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:170: exit = 0, want 125; stdout="" stderr="offload: alpha slot 1/2 (8 CPUs, 25% busy) → ~/offload/project-c8f2e6f1\n"
reviewer10_mutation_test.go:170: exit = 0, want 125; stdout="" stderr="offload: alpha slot 1/2 (8 CPUs, 25% busy) → ~/offload/project-b8cad6f1\n"
```

### 7a332d20

`go test -count=2 -run '^TestY4UploadRechecksEntriesAfterWaiting$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:170: exit = 0, want 125; stdout="" stderr="offload: alpha slot 1/2 (8 CPUs, 25% busy) → ~/offload/project-710a9c12\n"
reviewer10_mutation_test.go:170: exit = 0, want 125; stdout="" stderr="offload: alpha slot 1/2 (8 CPUs, 25% busy) → ~/offload/project-e9a41c6c\n"
```

### d66b5564

`go test -count=2 -run '^TestY4UploadRechecksEntriesAfterWaiting$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:170: exit = 0, want 125; stdout="" stderr="offload: alpha slot 1/2 (8 CPUs, 25% busy) → ~/offload/project-b5bad2a8\n"
reviewer10_mutation_test.go:170: exit = 0, want 125; stdout="" stderr="offload: alpha slot 1/2 (8 CPUs, 25% busy) → ~/offload/project-859c4e53\n"
```

### 1ab3dd2a

`go test -count=2 -run '^TestY5UnreadableGlobalExcludesStopsRun$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:183: exit = 0, want 125; stdout="" stderr="offload: alpha slot 1/2 (8 CPUs, 25% busy) → ~/offload/project-a3d00f97\n"
reviewer10_mutation_test.go:183: exit = 0, want 125; stdout="" stderr="offload: alpha slot 1/2 (8 CPUs, 25% busy) → ~/offload/project-80159f93\n"
```

### 5ae6a6e3

`go test -count=2 -run '^TestY5UnreadableGlobalExcludesStopsRun$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:183: exit = 0, want 125; stdout="" stderr="offload: alpha slot 1/2 (8 CPUs, 25% busy) → ~/offload/project-6e699131\n"
reviewer10_mutation_test.go:183: exit = 0, want 125; stdout="" stderr="offload: alpha slot 1/2 (8 CPUs, 25% busy) → ~/offload/project-3a05a079\n"
```

### 7f4a7b42

`go test -count=2 -run '^TestY5UnreadableGlobalExcludesStopsRun$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:183: exit = 0, want 125; stdout="" stderr="offload: alpha slot 1/2 (8 CPUs, 25% busy) → ~/offload/project-3897bef3\n"
reviewer10_mutation_test.go:183: exit = 0, want 125; stdout="" stderr="offload: alpha slot 1/2 (8 CPUs, 25% busy) → ~/offload/project-4cd32008\n"
```

### ab95db35

`go test -count=2 -run '^TestL1NonExecutableLocalCommandExplainsPoolOnly$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:141: "; pool hosts only\noffload: waiting for a free slot (alpha 2/2 held)\n" does not contain "offload: ./noexec: permission denied; pool hosts only"
```

### d0d46da8

`go test -count=2 -run '^TestX2NonExecutableLookupBeforeSlot$' ./test/e2e/` → rc 1, test red twice. Failing output:

```text
reviewer10_mutation_test.go:127: path must be absent: "/tmp/TestX2NonExecutableLookupBeforeSlot2401595428/001/runtime/heavy/slot1" (lstat error=<nil>)
reviewer10_mutation_test.go:127: path must be absent: "/tmp/TestX2NonExecutableLookupBeforeSlot3020406880/001/runtime/heavy/slot1" (lstat error=<nil>)
```

## Verification

- Rulings = 66/66: `new test` 36; `equivalent` 18; `contract-open` 10; `killed` 2; `n/a` 0. No unresolved row.
- Tests = `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-10/test/e2e/reviewer10_mutation_test.go` (14 tests); commit `d7ef33bcf0ef40121dcbf699303a591dea992f83` on `wt/reviewer-10`. Product code unchanged; `cmp` against `09e9969:internal/cli/cli.go` → rc 0.
- Exact mutant replay: 36 new-test rows + 2 existing-test kills, each red twice. Per-row command + failing output above. The existing real-podman test also passed with restored code.
- All 14 new tests, regex = their names joined by `|` inside `^(...)$`: `umask 022; go test -race -shuffle=on -count=2 -run REGEX ./test/e2e/` → rc 0 (24.540 s). Final helper-variable rename followed by `-race -shuffle=on -count=1` over the same regex → rc 0 (16.938 s).
- `golangci-lint fmt --diff test/e2e/reviewer10_mutation_test.go` → rc 0; `golangci-lint run ./test/e2e/` → rc 0, `0 issues.`; `.agent/law` → rc 0. Binary = `/home/eturkes/Projects/offload/.tools/golangci-lint-2.15.0/golangci-lint`.
- Iteration failures resolved: upload-duration fixture initially matched only ssh's final argument, so its intended delay did not run (baseline `upload=0.06`, wanted ≥0.1); matching the complete argv fixed the fixture before mutation credit. Lint initially rejected two variables named `real` (builtin shadow); renamed `wrappedSSH`, lint + tests green.
- Limits: full e2e suite + full gate not run (brief: narrow regexes while controller shared); 18 equivalent + 10 contract-open rulings = source adjudications, not exhaustive dynamic checks. LSP hover on cli.go failed (`No LSP server available for file type: .go`); source + caller/callee reads used. No unconfirmed behavioral claim retained as a new-test ruling.
- Validator: positive control = 66 table-row ids, rc 0; unknown-cell search = no output, rc 1 (0 matches).

reviewer-10-DONE-1
