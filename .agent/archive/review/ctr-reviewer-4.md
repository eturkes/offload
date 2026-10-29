# reviewer-4 — lens V: verification integrity (tests of diff 5c11b58..57903b5)

Target fixed = 57903b5; base = 5c11b58. Check set V1–V8 fixed from seed before target read. Contract = disk K1–K11; rules = disk `.claude/rules/{code,gate,gate-checks,machines}.md`. Writes = this validator input + `.scratch/worktrees/reviewer-4`; target + remotes read-only. V1–V8 adjudicated. Target production source restored byte-for-byte after every experiment. Active review tests include intentional reds for VF3 + VF5; the exact-codec proposal remains a nongating testdata fixture pending VF4.

| id | check | verdict (pass / finding) | finding ids | evidence (file:line, command + output) |
|----|-------|---------------------------|-------------|----------------------------------------|
| V1 | each K1–K11 predicate clause has a test that fails when the clause is violated (map clause → test; list uncovered clauses) | finding | VF2, VF3, VF5 | Clause map below; original K4 tests miss other-execute (T012), plus default-operator boundary survivors. Missing K7/K8 Podman-first-death/live-peer test conceals a real indefinite wait (VF3); help after a bad option fails K1 (VF5). |
| V2 | oracles independent of the code under test (no expected value computed by the implementation) | pass | — | Literal argv/status/stdio/env expectations; independent sha1 HOME naming + Podman inspect oracle; K4 metamorphic comparisons. `tagOf` derives only the synchronization resource used in K5, not a tag-correctness expectation. |
| V3 | red-first record in 0d14378 body reproducible (base 5c11b58 + new test files → all K tests fail) | pass | — | Replayed 5c11b58 + exact target e2e files: 29 failing tests + 3 failing subtests; 0 pass/skip; rc 1. `.scratch/worktrees/reviewer-4/.scratch/review/red-base.log`. |
| V4 | skip paths: podman tests cannot silently skip in CI (OFFLOAD_REQUIRE_PODMAN) or locally on a prebuild failure | pass | — | `requirePodman`: `storageErr` + required env → Fatal; `prebuildErr` → unconditional Fatal. 5c11b58 replay's shared-image prebuild failure produces FAIL, not SKIP; CI step sets `OFFLOAD_REQUIRE_PODMAN: "1"`. |
| V5 | timing/flakiness: sleeps, thresholds (9–20 s windows, 6 s awaits), parallel podman contention, pgrep/pkill patterns | pass | — | Exact target K suite: `-race -shuffle=on -count=1`, 29 tests + 3 subtests pass, no skips, rc 0 (14.920 s). Ready markers follow handler installation; shared prebuild stays outside launch clocks; polling/grace windows survive this parallel run. One-run coverage, not a stress/flakiness guarantee. Failed probe status is VF1, not a timing defect. |
| V6 | vacuous passes: checks that cannot fail (e.g. absence checks on markers never present, broad requireContains) | finding | VF1 | `markerGone` equates every `pgrep` error with absence; controlled rc 2 returns while the positive-control helper remains alive. Red probe: `v6-red.log`, rc 1. |
| V7 | mutation: mutago on internal/ctr (+ .agent/plant targeted mutants for code only e2e reaches: ctr.Run, Ensure, Init) → every survivor ruled (specified → red test; open → recorded) | finding | VF2, VF4 | 304 default mutants: 62 killed, 241 escaped, 1 noncompiling. Escapes: 188 killed by added tests, 50 open/equivalent, 3 conditional codec-order probes (VF4). 16 Init/CLI plants all killed. No combined score claimed; scope/limits below. |
| V8 | gate integrity: CI env change (OFFLOAD_REQUIRE_PODMAN) + TestX1NestedRunEndsWithOuterRun change keep every check's strength | pass | — | CI env strengthens absence handling. Nested-run diff retains the original death checks + adds slot-release synchronization; targeted unchanged test passed (0.26 s). No case/threshold removed or demoted. |

## Clause map (target tests, before review additions)

| predicate | observed assertions | clauses without a direct target assertion |
|---|---|---|
| K1 | `TestK1CtrUsage`, `TestK1HelpWinsOverBadSwitch`; `TestK6RunArgs` fixes CMD word boundaries | implicit first-nonflag termination + verbatim remaining CLI words; help behind a conflicting/bad option; PID-1-only hidden entry (K8) |
| K2 | `TestK2SwitchOffExecsInPlace` pins pid/env/status for empty + 0; missing/nonexecutable command; fake Podman tripwire; invalid switch names variable | truly unset env distinct from empty; native stdin/cwd/argv; malformed pool/git tripwires; ENOEXEC; `--layers`/`--check` with invalid switch |
| K3 | missing + symlink definition; damaged repo; colon root; git subdir + symlink cwd via K6; plain runs; common-dir mount | directory definition; additional git failure/ancestor cases via existing mirror tests rather than ctr-specific tests; binary/home/common-dir colon cases |
| K4 | equal trees/location/mtime; content/name/new entry/file-vs-link/link-target/selected modes; FIFO; framed path/content collision pair | owner changes; all 07777 bits (notably other-execute); unreadable file/directory; context-root mode exclusion; explicit global path-byte order |
| K5 | miss/build log/reuse/change; two concurrent starts build once; lock-wait TERM; exists failure; build inherits lock past SIGKILL | build-failure captured stdout+stderr; recheck failure; build argv's `--timestamp 0`; non-TERM setup signals; TERM during active build; lock release/error paths; symlink lock uses shared slots coverage |
| K6 | exact RunArgs; cwd/root/common-dir mounts; invoking uid writes; image ENV + host-env absence; HOME across runs + independent root names; static binary identity helper; byte-exact stdin/stdout/stderr | HOME default cache fallback + 0700; persistence across image changes; image ENTRYPOINT/CMD/USER/WORKDIR poison defaults; live output before CMD ends; running-inode guard wired into Run (helper alone insufficient) |
| K7 | selected statuses incl. 125/126/127/255; TERM status; missing/nonexecutable command; failed/missing Podman; malformed/incomplete record helper | complete record despite independently different Podman rc; successful Podman with no record; ENOEXEC in image; late transport loss |
| K8 | HUP/INT/TERM once; offload + Podman TERM once; SIGKILL death/removal; TERM-ignoring leftover + grace; outer-run death + slot-after-last-write | QUIT/USR1/USR2; signal reaches whole CMD group; before-CMD queued signals; PID-1 refusal; fd3 noninheritance; Podman-first death; reap arbitrary exited children while CMD remains live; failed observation falsely passes (VF1) |
| K9 | real `--layers` equals independent Podman inspect | zero layers; failed inspect with partial stdout; setup/inspect signals |
| K10 | same/different/binary-mismatch; down host; banner with forged layer becomes different; missing installed offload; remote worktree via K6 | ≥2 ordered layers; >1 host order + parallelism; bad pool before local work; local failure before remote work; remote stderr equality; check interruption |
| K11 | TestMain builds both binaries with CGO_ENABLED=0; real rootless scratch-image execution | Linux/rootless/local Podman ≥4.9, uid/gid ranges + architecture are environmental preconditions, not CLI rejection predicates; CI's actual runner version not executed by this review |

Coverage inventory ≠ a claim every absent case is a production defect. Mutation rows decide which gaps demonstrably admit a contract violation. Existing source roots `mirror.Locate/Name` retain Y1/Y2 regression coverage; new `WaitLock` is exercised by K5 lock wait/inheritance. Target `TestSlotFilesNeverFollowSymlinks` calls TryAcquire, not WaitLock: it does not credit the waiter's no-follow clause.

## Mutation survivors (one row each: file:line, operator, ruling, test or reason)

| id | target site | operator | ruling / proof |
|---|---|---|---|
| M001 | `internal/ctr/ctr.go:212` | `arithmetic/base` | test-added → killed: `TestK5ReviewCancelActiveBuild`. `.agent/plant`, rc 0; test FAIL (rc 1); round2. |
| M002 | `internal/ctr/ctr.go:251` | `arithmetic/base` | test-added → killed: `TestK5ReviewSignalAfterEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M003 | `internal/ctr/ctr.go:235` | `arithmetic/bitwise` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M004 | `internal/ctr/ctr.go:142` | `branch/case` | test-added → killed: `TestK5ReviewEnsure`, `TestK7ReviewRunBoundary`, `TestK5ReviewSignalAfterEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M005 | `internal/ctr/ctr.go:57` | `branch/if` | test-added → killed: `TestK3ReviewFindFailures`. `.agent/plant`, rc 0; test FAIL (rc 1); faults replay. |
| M006 | `internal/ctr/ctr.go:144` | `branch/case` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M007 | `internal/ctr/ctr.go:61` | `branch/if` | test-added → killed: `TestK3ReviewFindErrorsBeforePodman`. `.agent/plant`, rc 0; test FAIL (rc 1); round2. |
| M008 | `internal/ctr/ctr.go:64` | `branch/if` | test-added → killed: `TestK3ReviewFindFailures`. `.agent/plant`, rc 0; test FAIL (rc 1); faults replay. |
| M009 | `internal/ctr/ctr.go:67` | `branch/if` | test-added → killed: `TestK6ReviewExternalGitDir`. `.agent/plant`, rc 0; test FAIL (rc 1); round2. |
| M010 | `internal/ctr/ctr.go:69` | `branch/if` | test-added → killed: `TestK3ReviewFindFailures`. `.agent/plant`, rc 0; test FAIL (rc 1); faults replay. |
| M011 | `internal/ctr/ctr.go:72` | `branch/if` | test-added → killed: `TestK6ReviewExternalGitDir`. `.agent/plant`, rc 0; test FAIL (rc 1); round2. |
| M012 | `internal/ctr/ctr.go:77` | `branch/if` | test-added → killed: `TestK3ReviewFind`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M013 | `internal/ctr/ctr.go:81` | `branch/if` | contract-open K9: failure of an otherwise unused executable-path lookup during --layers is unprescribed. Run still rejects an empty/directory/missing binary in running(); --check still fails fileSHA256. K6 replacement refusal retained; diagnostic path differs. |
| M014 | `internal/ctr/ctr.go:90` | `branch/if` | contract-open K9: failure of an otherwise unused executable-path lookup during --layers is unprescribed. Run still rejects an empty/directory/missing binary in running(); --check still fails fileSHA256. K6 replacement refusal retained; diagnostic path differs. |
| M015 | `internal/ctr/ctr.go:106` | `branch/if` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M016 | `internal/ctr/ctr.go:93` | `branch/if` | contract-open K9: failure of an otherwise unused executable-path lookup during --layers is unprescribed. Run still rejects an empty/directory/missing binary in running(); --check still fails fileSHA256. K6 replacement refusal retained; diagnostic path differs. |
| M017 | `internal/ctr/ctr.go:113` | `branch/if` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M018 | `internal/ctr/ctr.go:109` | `branch/if` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1); round2. |
| M019 | `internal/ctr/ctr.go:117` | `branch/if` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M020 | `internal/ctr/ctr.go:125` | `branch/if` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M021 | `internal/ctr/ctr.go:126` | `branch/if` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M022 | `internal/ctr/ctr.go:154` | `branch/if` | test-added → killed: `TestK9ReviewLayers`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M023 | `internal/ctr/ctr.go:163` | `branch/if` | test-added → killed: `TestK6ReviewHomeFallback`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M024 | `internal/ctr/ctr.go:165` | `branch/if` | test-added → killed: `TestK6ReviewHomeFallback`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M025 | `internal/ctr/ctr.go:200` | `branch/if` | test-added → killed: `TestK8ReviewControlSignalWithoutPodman`. `.agent/plant`, rc 0; test FAIL (rc 1); round2. |
| M026 | `internal/ctr/ctr.go:202` | `branch/if` | test-added → killed: `TestK5ReviewSignalAfterEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M027 | `internal/ctr/ctr.go:212` | `branch/if` | test-added → killed: `TestK5ReviewCancelActiveBuild`. `.agent/plant`, rc 0; test FAIL (rc 1); round2. |
| M028 | `internal/ctr/ctr.go:220` | `branch/if` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1); faults replay. |
| M029 | `internal/ctr/ctr.go:224` | `branch/if` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M030 | `internal/ctr/ctr.go:227` | `branch/if` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M031 | `internal/ctr/ctr.go:230` | `branch/if` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M032 | `internal/ctr/ctr.go:233` | `branch/if` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M033 | `internal/ctr/ctr.go:237` | `branch/if` | test-added → killed: `TestK7ReviewSocketpairFailure`. `.agent/plant`, rc 0; test FAIL (rc 1); faults replay. |
| M034 | `internal/ctr/ctr.go:249` | `branch/if` | test-added → killed: `TestK5ReviewSignalAfterEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M035 | `internal/ctr/ctr.go:255` | `branch/if` | test-added → killed: `TestK8ReviewControlSignalWithoutPodman`. `.agent/plant`, rc 0; test FAIL (rc 1); round2. |
| M036 | `internal/ctr/ctr.go:260` | `branch/if` | equivalent K7: failed Start leaves no child + closes theirs; Wait returns ErrNotStarted, readRecord sees EOF, final fail still returns 125 (or recorded setup signal), with an offload diagnostic. |
| M037 | `internal/ctr/ctr.go:268` | `branch/if` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M038 | `internal/ctr/ctr.go:272` | `branch/if` | contract-open K7 diagnostic wording: fallback fail still reports the ExitError + returns 125; no complete record fabricated. |
| M039 | `internal/ctr/ctr.go:292` | `branch/if` | equivalent K6: failed self Stat leaves nil FileInfo; candidate Stat error or SameFile(nil, cur)==false still rejects. Injected self-stat EPERM probe confirms refusal; only diagnostic changes. |
| M040 | `internal/ctr/ctr.go:246` | `composite/field-clear` | contract-open K6/K8: Podman's host-side PGID is not specified; CMD still gets its own group in Init.start. Group-wide proxy duplication is handled by init signal dropping. This is not a claim that every terminal-startup interleaving was reproduced. |
| M041 | `internal/ctr/ctr.go:195` | `concurrency/goroutine-remove` | test-added → killed: `TestK7ReviewRunBoundary`, `TestK5ReviewSignalAfterEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M042 | `internal/ctr/ctr.go:56` | `conditional/negated` | test-added → killed: `TestK3ReviewFind`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M043 | `internal/ctr/ctr.go:60` | `conditional/negated` | test-added → killed: `TestK3ReviewFind`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M044 | `internal/ctr/ctr.go:63` | `conditional/negated` | test-added → killed: `TestK3ReviewFind`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M045 | `internal/ctr/ctr.go:68` | `conditional/negated` | test-added → killed: `TestK3ReviewFind`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M046 | `internal/ctr/ctr.go:76` | `conditional/negated` | test-added → killed: `TestK3ReviewFind`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M047 | `internal/ctr/ctr.go:80` | `conditional/negated` | contract-open K9: failure of an otherwise unused executable-path lookup during --layers is unprescribed. Run still rejects an empty/directory/missing binary in running(); --check still fails fileSHA256. K6 replacement refusal retained; diagnostic path differs. |
| M048 | `internal/ctr/ctr.go:99` | `conditional/negated` | test-added → killed: `TestK6ReviewExternalGitDir`. `.agent/plant`, rc 0; test FAIL (rc 1); round2. |
| M049 | `internal/ctr/ctr.go:105` | `conditional/negated` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M050 | `internal/ctr/ctr.go:108` | `conditional/negated` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M051 | `internal/ctr/ctr.go:112` | `conditional/negated` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M052 | `internal/ctr/ctr.go:116` | `conditional/negated` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M053 | `internal/ctr/ctr.go:124` | `conditional/negated` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M054 | `internal/ctr/ctr.go:125` | `conditional/negated` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M055 | `internal/ctr/ctr.go:141` | `conditional/negated` | test-added → killed: `TestK5ReviewEnsure`, `TestK7ReviewRunBoundary`, `TestK5ReviewSignalAfterEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M056 | `internal/ctr/ctr.go:143` | `conditional/negated` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M057 | `internal/ctr/ctr.go:153` | `conditional/negated` | test-added → killed: `TestK9ReviewLayers`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M058 | `internal/ctr/ctr.go:164` | `conditional/negated` | test-added → killed: `TestK6ReviewHomeFallback`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M059 | `internal/ctr/ctr.go:199` | `conditional/negated` | test-added → killed: `TestK5ReviewSignalAfterEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M060 | `internal/ctr/ctr.go:201` | `conditional/negated` | test-added → killed: `TestK5ReviewSignalAfterEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M061 | `internal/ctr/ctr.go:211` | `conditional/negated` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M062 | `internal/ctr/ctr.go:219` | `conditional/negated` | test-added → killed: `TestK7ReviewRunBoundary`, `TestK5ReviewSignalAfterEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M063 | `internal/ctr/ctr.go:223` | `conditional/negated` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M064 | `internal/ctr/ctr.go:226` | `conditional/negated` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M065 | `internal/ctr/ctr.go:229` | `conditional/negated` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M066 | `internal/ctr/ctr.go:232` | `conditional/negated` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M067 | `internal/ctr/ctr.go:236` | `conditional/negated` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M068 | `internal/ctr/ctr.go:248` | `conditional/negated` | test-added → killed: `TestK7ReviewRunBoundary`, `TestK5ReviewSignalAfterEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M069 | `internal/ctr/ctr.go:254` | `conditional/negated` | test-added → killed: `TestK8ReviewControlSignalWithoutPodman`. `.agent/plant`, rc 0; test FAIL (rc 1); round2. |
| M070 | `internal/ctr/ctr.go:259` | `conditional/negated` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M071 | `internal/ctr/ctr.go:71` | `conditional/not` | test-added → killed: `TestK3ReviewFind`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M072 | `internal/ctr/ctr.go:76` | `conditional/not` | test-added → killed: `TestK3ReviewFind`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M073 | `internal/ctr/ctr.go:307` | `expression/comparison` | equivalent K7 valid producer domain: CMD status is 0..255; decimal record <=4 bytes. Bounds 15/16/17 consume every complete valid record identically; longer malformed inputs are not emitted by init. |
| M074 | `internal/ctr/ctr.go:59` | `expression/context-nil` | test-added → killed: `TestK3ReviewFind`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M075 | `internal/ctr/ctr.go:108` | `expression/context-nil` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M076 | `internal/ctr/ctr.go:111` | `expression/context-nil` | test-added → killed: `TestK5ReviewWaitCancellation`. `.agent/plant`, rc 0; test FAIL (rc 1); round2. |
| M077 | `internal/ctr/ctr.go:116` | `expression/context-nil` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M078 | `internal/ctr/ctr.go:120` | `expression/context-nil` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M079 | `internal/ctr/ctr.go:136` | `expression/context-nil` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M080 | `internal/ctr/ctr.go:151` | `expression/context-nil` | test-added → killed: `TestK9ReviewLayers`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M081 | `internal/ctr/ctr.go:190` | `expression/context-nil` | test-added → killed: `TestK7ReviewCommandStatusWithoutPodman`, `TestK6ReviewStreamsWithoutPodman`, `TestK8ReviewControlSignalWithoutPodman`. `.agent/plant`, rc 0; test FAIL (rc 1); round2. |
| M082 | `internal/ctr/ctr.go:56` | `expression/error-guard` | test-added → killed: `TestK3ReviewFindFailures`. `.agent/plant`, rc 0; test FAIL (rc 1); faults replay. |
| M083 | `internal/ctr/ctr.go:60` | `expression/error-guard` | test-added → killed: `TestK3ReviewFindErrorsBeforePodman`. `.agent/plant`, rc 0; test FAIL (rc 1); round2. |
| M084 | `internal/ctr/ctr.go:63` | `expression/error-guard` | test-added → killed: `TestK3ReviewFindFailures`. `.agent/plant`, rc 0; test FAIL (rc 1); faults replay. |
| M085 | `internal/ctr/ctr.go:68` | `expression/error-guard` | test-added → killed: `TestK3ReviewFindFailures`. `.agent/plant`, rc 0; test FAIL (rc 1); faults replay. |
| M086 | `internal/ctr/ctr.go:80` | `expression/error-guard` | contract-open K9: failure of an otherwise unused executable-path lookup during --layers is unprescribed. Run still rejects an empty/directory/missing binary in running(); --check still fails fileSHA256. K6 replacement refusal retained; diagnostic path differs. |
| M087 | `internal/ctr/ctr.go:89` | `expression/error-guard` | contract-open K9: failure of an otherwise unused executable-path lookup during --layers is unprescribed. Run still rejects an empty/directory/missing binary in running(); --check still fails fileSHA256. K6 replacement refusal retained; diagnostic path differs. |
| M088 | `internal/ctr/ctr.go:92` | `expression/error-guard` | contract-open K9: failure of an otherwise unused executable-path lookup during --layers is unprescribed. Run still rejects an empty/directory/missing binary in running(); --check still fails fileSHA256. K6 replacement refusal retained; diagnostic path differs. |
| M089 | `internal/ctr/ctr.go:105` | `expression/error-guard` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M090 | `internal/ctr/ctr.go:112` | `expression/error-guard` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M091 | `internal/ctr/ctr.go:124` | `expression/error-guard` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M092 | `internal/ctr/ctr.go:125` | `expression/error-guard` | contract-open K5: extra captured build output on interruption; build stops and signal status is unchanged. Contract does not require silence for interrupted build diagnostics. |
| M093 | `internal/ctr/ctr.go:153` | `expression/error-guard` | test-added → killed: `TestK9ReviewLayers`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M094 | `internal/ctr/ctr.go:164` | `expression/error-guard` | test-added → killed: `TestK6ReviewHomeFallback`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M095 | `internal/ctr/ctr.go:219` | `expression/error-guard` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1); faults replay. |
| M096 | `internal/ctr/ctr.go:223` | `expression/error-guard` | equivalent K6 failure boundary: Home error returns empty path; unconditional MkdirAll("") also errors. No Podman launch; same 125, different unprescribed diagnostic. Fault replay preserves refusal. |
| M097 | `internal/ctr/ctr.go:226` | `expression/error-guard` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M098 | `internal/ctr/ctr.go:229` | `expression/error-guard` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M099 | `internal/ctr/ctr.go:232` | `expression/error-guard` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M100 | `internal/ctr/ctr.go:236` | `expression/error-guard` | test-added → killed: `TestK7ReviewSocketpairFailure`. `.agent/plant`, rc 0; test FAIL (rc 1); faults replay. |
| M101 | `internal/ctr/ctr.go:254` | `expression/error-guard` | contract-open K5/K7 boundary: on failed Start only, ctl becomes mine before failure return; no Podman child starts. The winner of overlapping failed launch + asynchronously delivered signal is not linearized by the contract. Non-overlapping result remains 125. Judgment, not an equivalence claim. |
| M102 | `internal/ctr/ctr.go:259` | `expression/error-guard` | equivalent K7: failed Start leaves no child + closes theirs; Wait returns ErrNotStarted, readRecord sees EOF, final fail still returns 125 (or recorded setup signal), with an offload diagnostic. |
| M103 | `internal/ctr/ctr.go:291` | `expression/error-guard` | equivalent K6: failed self Stat leaves nil FileInfo; candidate Stat error or SameFile(nil, cur)==false still rejects. Injected self-stat EPERM probe confirms refusal; only diagnostic changes. |
| M104 | `internal/ctr/ctr.go:93` | `expression/errorf-wrap` | equivalent CLI: %w→%v preserves message; callers inspect only nil/error (no unwrapping of these errors). Status and command admission unchanged. |
| M105 | `internal/ctr/ctr.go:128` | `expression/errorf-wrap` | equivalent CLI: %w→%v preserves message; callers inspect only nil/error (no unwrapping of these errors). Status and command admission unchanged. |
| M106 | `internal/ctr/ctr.go:146` | `expression/errorf-wrap` | equivalent CLI: %w→%v preserves message; callers inspect only nil/error (no unwrapping of these errors). Status and command admission unchanged. |
| M107 | `internal/ctr/ctr.go:154` | `expression/errorf-wrap` | equivalent CLI: %w→%v preserves message; callers inspect only nil/error (no unwrapping of these errors). Status and command admission unchanged. |
| M108 | `internal/ctr/ctr.go:274` | `expression/errorf-wrap` | contract-open K7 diagnostic wording for no record: status still 125 + offload message; %w/%v of nil may render differently, not a specified payload. |
| M109 | `internal/ctr/ctr.go:76` | `expression/logical` | test-added → killed: `TestK3ReviewFind`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M110 | `internal/ctr/ctr.go:99` | `expression/logical` | test-added → killed: `TestK3ReviewFind`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M111 | `internal/ctr/ctr.go:108` | `expression/logical` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1); round2. |
| M112 | `internal/ctr/ctr.go:116` | `expression/logical` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M113 | `internal/ctr/ctr.go:143` | `expression/logical` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M114 | `internal/ctr/ctr.go:99` | `expression/remove` | test-added → killed: `TestK6ReviewExternalGitDir`. `.agent/plant`, rc 0; test FAIL (rc 1); round2. |
| M115 | `internal/ctr/ctr.go:99` | `expression/remove` | test-added → killed: `TestK3ReviewFind`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M116 | `internal/ctr/ctr.go:108` | `expression/remove` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1); round2. |
| M117 | `internal/ctr/ctr.go:116` | `expression/remove` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M118 | `internal/ctr/ctr.go:143` | `expression/remove` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M119 | `internal/ctr/ctr.go:143` | `expression/remove` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M120 | `internal/ctr/ctr.go:196` | `loop/range_break` | test-added → killed: `TestK5ReviewSignalAfterEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M121 | `internal/ctr/ctr.go:37` | `numbers/decrementer` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M122 | `internal/ctr/ctr.go:143` | `numbers/decrementer` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M123 | `internal/ctr/ctr.go:201` | `numbers/decrementer` | test-added → killed: `TestK5ReviewSignalAfterEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M124 | `internal/ctr/ctr.go:211` | `numbers/decrementer` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M125 | `internal/ctr/ctr.go:212` | `numbers/decrementer` | test-added → killed: `TestK5ReviewCancelActiveBuild`. `.agent/plant`, rc 0; test FAIL (rc 1); round2. |
| M126 | `internal/ctr/ctr.go:224` | `numbers/decrementer` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M127 | `internal/ctr/ctr.go:235` | `numbers/decrementer` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M128 | `internal/ctr/ctr.go:239` | `numbers/decrementer` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M129 | `internal/ctr/ctr.go:248` | `numbers/decrementer` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M130 | `internal/ctr/ctr.go:251` | `numbers/decrementer` | test-added → killed: `TestK5ReviewSignalAfterEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M131 | `internal/ctr/ctr.go:307` | `numbers/decrementer` | equivalent K7 valid producer domain: CMD status is 0..255; decimal record <=4 bytes. Bounds 15/16/17 consume every complete valid record identically; longer malformed inputs are not emitted by init. |
| M132 | `internal/ctr/ctr.go:318` | `numbers/decrementer` | equivalent K7: numeric return accompanying ok=false is discarded by the sole caller; only ok gates command status. |
| M133 | `internal/ctr/ctr.go:37` | `numbers/incrementer` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M134 | `internal/ctr/ctr.go:143` | `numbers/incrementer` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M135 | `internal/ctr/ctr.go:201` | `numbers/incrementer` | test-added → killed: `TestK5ReviewSignalAfterEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M136 | `internal/ctr/ctr.go:211` | `numbers/incrementer` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M137 | `internal/ctr/ctr.go:212` | `numbers/incrementer` | test-added → killed: `TestK5ReviewCancelActiveBuild`. `.agent/plant`, rc 0; test FAIL (rc 1); round2. |
| M138 | `internal/ctr/ctr.go:224` | `numbers/incrementer` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M139 | `internal/ctr/ctr.go:235` | `numbers/incrementer` | equivalent K6/K11 Linux AF_UNIX protocol alias: 0 and PF_UNIX(1) create SOCK_STREAM+CLOEXEC sockets. Live positive probe: both SO_TYPE=1, payload b"7\n", inheritable=false. |
| M140 | `internal/ctr/ctr.go:239` | `numbers/incrementer` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M141 | `internal/ctr/ctr.go:248` | `numbers/incrementer` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M142 | `internal/ctr/ctr.go:251` | `numbers/incrementer` | test-added → killed: `TestK5ReviewSignalAfterEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M143 | `internal/ctr/ctr.go:307` | `numbers/incrementer` | equivalent K7 valid producer domain: CMD status is 0..255; decimal record <=4 bytes. Bounds 15/16/17 consume every complete valid record identically; longer malformed inputs are not emitted by init. |
| M144 | `internal/ctr/ctr.go:115` | `statement/defer-remove` | test-added → killed: `TestK5ReviewBuildKeepsLockWithoutPodman`. `.agent/plant`, rc 0; test FAIL (rc 1); round2. |
| M145 | `internal/ctr/ctr.go:318` | `numbers/incrementer` | equivalent K7: numeric return accompanying ok=false is discarded by the sole caller; only ok gates command status. |
| M146 | `internal/ctr/ctr.go:191` | `statement/defer-remove` | test-added → killed: `TestK7ReviewRunBoundary`, `TestK5ReviewSignalAfterEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M147 | `internal/ctr/ctr.go:240` | `statement/defer-remove` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M148 | `internal/ctr/ctr.go:75` | `statement/remove` | test-added → killed: `TestK3ReviewFind`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M149 | `internal/ctr/ctr.go:210` | `statement/defer-remove` | contract-open K5/K7 boundary: releases failure-decision mutex before reading got; only an overlapping failure/signal can distinguish it. Contract does not fix that concurrency linearization. No specific overlapping race-detector replay credited. |
| M150 | `internal/ctr/ctr.go:72` | `statement/remove` | test-added → killed: `TestK6ReviewExternalGitDir`. `.agent/plant`, rc 0; test FAIL (rc 1); round2. |
| M151 | `internal/ctr/ctr.go:79` | `statement/remove` | test-added → killed: `TestK3ReviewFind`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M152 | `internal/ctr/ctr.go:119` | `statement/remove` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M153 | `internal/ctr/ctr.go:122` | `statement/remove` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M154 | `internal/ctr/ctr.go:123` | `statement/remove` | test-added → killed: `TestK5ReviewBuildKeepsLockWithoutPodman`. `.agent/plant`, rc 0; test FAIL (rc 1); round2. |
| M155 | `internal/ctr/ctr.go:137` | `statement/remove` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M156 | `internal/ctr/ctr.go:167` | `statement/remove` | test-added → killed: `TestK6ReviewHomeFallback`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M157 | `internal/ctr/ctr.go:243` | `statement/remove` | test-added → killed: `TestK6ReviewStreamsWithoutPodman`. `.agent/plant`, rc 0; test FAIL (rc 1); round2. |
| M158 | `internal/ctr/ctr.go:244` | `statement/remove` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M159 | `internal/ctr/ctr.go:246` | `statement/remove` | contract-open K6/K8: same host-side Podman PGID question as M040; no change to the specified CMD process group or control-socket protocol. |
| M160 | `internal/ctr/ctr.go:247` | `statement/remove` | test-added → killed: `TestK7ReviewCommandStatusWithoutPodman`, `TestK6ReviewStreamsWithoutPodman`, `TestK8ReviewControlSignalWithoutPodman`. `.agent/plant`, rc 0; test FAIL (rc 1); round2. |
| M161 | `internal/ctr/ctr.go:253` | `statement/remove` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M162 | `internal/ctr/ctr.go:257` | `statement/remove` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M163 | `internal/ctr/ctr.go:263` | `statement/remove` | test-added → killed: `TestK7ReviewCommandStatusWithoutPodman`, `TestK6ReviewStreamsWithoutPodman`, `TestK8ReviewControlSignalWithoutPodman`. `.agent/plant`, rc 0; test FAIL (rc 1); round2. |
| M164 | `internal/ctr/ctr.go:264` | `statement/remove` | contract-open K7/K8 boundary: retains ctl after Podman Wait; a complete record still wins, and only a signal arriving after Podman completion can distinguish the states. Late-signal/record-arrival ordering is unspecified; VF3 separately blocks the unbounded missing-record path. |
| M165 | `internal/ctr/ctr.go:265` | `statement/remove` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M166 | `internal/ctr/ctr.go:198` | `statement/remove` | test-added → killed: `TestK8ReviewControlSignalWithoutPodman`, `TestK5ReviewCancelActiveBuild`. `.agent/plant`, rc 0; test FAIL (rc 1); round2. |
| M167 | `internal/ctr/ctr.go:205` | `statement/remove` | test-added → killed: `TestK5ReviewSignalAfterEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M168 | `internal/ctr/ctr.go:202` | `statement/remove` | test-added → killed: `TestK5ReviewSignalAfterEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M169 | `internal/ctr/ctr.go:203` | `statement/remove` | test-added → killed: `TestK5ReviewCancelActiveBuild`. `.agent/plant`, rc 0; test FAIL (rc 1); round2. |
| M170 | `internal/ctr/ctr.go:209` | `statement/remove` | test-added → killed: `TestK5ReviewCancelActiveBuild`. `.agent/plant`, rc 0; test FAIL (rc 1); round2. |
| M171 | `internal/ctr/ctr.go:224` | `statement/remove` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M172 | `internal/ctr/ctr.go:214` | `statement/remove` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M173 | `internal/ctr/ctr.go:227` | `statement/remove` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M174 | `internal/ctr/ctr.go:230` | `statement/remove` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M175 | `internal/ctr/ctr.go:255` | `statement/remove` | test-added → killed: `TestK8ReviewControlSignalWithoutPodman`. `.agent/plant`, rc 0; test FAIL (rc 1); round2. |
| M176 | `internal/ctr/ctr.go:249` | `statement/remove` | equivalent CLI K5: prelaunch signal returns identical 128+n, closes theirs, starts no command. Retained private mutex affects only a goroutine until this one-command CLI exits; no public reusable-library lifetime contract. |
| M177 | `internal/ctr/ctr.go:324` | `statement/remove` | test-added → killed: `TestK8ReviewNotify`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M178 | `internal/ctr/ctr.go:83` | `statement/return` | test-added → killed: `TestK3ReviewFind`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M179 | `internal/ctr/ctr.go:57` | `statement/return` | equivalent K3: only the unused Project value accompanying nonnil error changes; findProject discards it on every failure. Same error type/message + rc. |
| M180 | `internal/ctr/ctr.go:57` | `statement/return` | test-added → killed: `TestK3ReviewFindFailures`. `.agent/plant`, rc 0; test FAIL (rc 1); faults replay. |
| M181 | `internal/ctr/ctr.go:61` | `statement/return` | equivalent K3: only the unused Project value accompanying nonnil error changes; findProject discards it on every failure. Same error type/message + rc. |
| M182 | `internal/ctr/ctr.go:61` | `statement/return` | test-added → killed: `TestK3ReviewFindErrorsBeforePodman`. `.agent/plant`, rc 0; test FAIL (rc 1); round2. |
| M183 | `internal/ctr/ctr.go:64` | `statement/return` | equivalent K3: only the unused Project value accompanying nonnil error changes; findProject discards it on every failure. Same error type/message + rc. |
| M184 | `internal/ctr/ctr.go:64` | `statement/return` | test-added → killed: `TestK3ReviewFindFailures`. `.agent/plant`, rc 0; test FAIL (rc 1); faults replay. |
| M185 | `internal/ctr/ctr.go:69` | `statement/return` | equivalent K3: only the unused Project value accompanying nonnil error changes; findProject discards it on every failure. Same error type/message + rc. |
| M186 | `internal/ctr/ctr.go:69` | `statement/return` | test-added → killed: `TestK3ReviewFindFailures`. `.agent/plant`, rc 0; test FAIL (rc 1); faults replay. |
| M187 | `internal/ctr/ctr.go:77` | `statement/return` | equivalent K3: only the unused Project value accompanying nonnil error changes; findProject discards it on every failure. Same error type/message + rc. |
| M188 | `internal/ctr/ctr.go:77` | `statement/return` | test-added → killed: `TestK3ReviewFind`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M189 | `internal/ctr/ctr.go:81` | `statement/return` | equivalent K3: only the unused Project value accompanying nonnil error changes; findProject discards it on every failure. Same error type/message + rc. |
| M190 | `internal/ctr/ctr.go:81` | `statement/return` | contract-open K9: failure of an otherwise unused executable-path lookup during --layers is unprescribed. Run still rejects an empty/directory/missing binary in running(); --check still fails fileSHA256. K6 replacement refusal retained; diagnostic path differs. |
| M191 | `internal/ctr/ctr.go:90` | `statement/return` | contract-open K9: failure of an otherwise unused executable-path lookup during --layers is unprescribed. Run still rejects an empty/directory/missing binary in running(); --check still fails fileSHA256. K6 replacement refusal retained; diagnostic path differs. |
| M192 | `internal/ctr/ctr.go:93` | `statement/return` | contract-open K9: failure of an otherwise unused executable-path lookup during --layers is unprescribed. Run still rejects an empty/directory/missing binary in running(); --check still fails fileSHA256. K6 replacement refusal retained; diagnostic path differs. |
| M193 | `internal/ctr/ctr.go:99` | `statement/return` | test-added → killed: `TestK3ReviewFind`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M194 | `internal/ctr/ctr.go:130` | `statement/return` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M195 | `internal/ctr/ctr.go:106` | `statement/return` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M196 | `internal/ctr/ctr.go:109` | `statement/return` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M197 | `internal/ctr/ctr.go:109` | `statement/return` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M198 | `internal/ctr/ctr.go:113` | `statement/return` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M199 | `internal/ctr/ctr.go:117` | `statement/return` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M200 | `internal/ctr/ctr.go:117` | `statement/return` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M201 | `internal/ctr/ctr.go:128` | `statement/return` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M202 | `internal/ctr/ctr.go:146` | `statement/return` | test-added → killed: `TestK5ReviewEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M203 | `internal/ctr/ctr.go:142` | `statement/return` | test-added → killed: `TestK5ReviewEnsure`, `TestK7ReviewRunBoundary`, `TestK5ReviewSignalAfterEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M204 | `internal/ctr/ctr.go:156` | `statement/return` | test-added → killed: `TestK9ReviewLayers`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M205 | `internal/ctr/ctr.go:154` | `statement/return` | test-added → killed: `TestK9ReviewLayers`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M206 | `internal/ctr/ctr.go:165` | `statement/return` | test-added → killed: `TestK6ReviewHomeFallback`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M207 | `internal/ctr/ctr.go:274` | `statement/return` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M208 | `internal/ctr/ctr.go:215` | `statement/return` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M209 | `internal/ctr/ctr.go:212` | `statement/return` | test-added → killed: `TestK5ReviewCancelActiveBuild`. `.agent/plant`, rc 0; test FAIL (rc 1); round2. |
| M210 | `internal/ctr/ctr.go:220` | `statement/return` | test-added → killed: `TestK5ReviewCancelActiveBuild`. `.agent/plant`, rc 0; test FAIL (rc 1); round2. |
| M211 | `internal/ctr/ctr.go:233` | `statement/return` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M212 | `internal/ctr/ctr.go:237` | `statement/return` | test-added → killed: `TestK7ReviewSocketpairFailure`. `.agent/plant`, rc 0; test FAIL (rc 1); faults replay. |
| M213 | `internal/ctr/ctr.go:251` | `statement/return` | test-added → killed: `TestK5ReviewSignalAfterEnsure`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M214 | `internal/ctr/ctr.go:260` | `statement/return` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M215 | `internal/ctr/ctr.go:268` | `statement/return` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M216 | `internal/ctr/ctr.go:272` | `statement/return` | test-added → killed: `TestK7ReviewRunBoundary`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| M217 | `internal/ctr/ctr.go:292` | `statement/return` | test-added → killed: `TestK6ReviewUnavailableSelfIdentity`. `.agent/plant`, rc 0; test FAIL (rc 1); faults replay. |
| M218 | `internal/ctr/ctr.go:325` | `statement/return` | test-added → killed: `TestK8ReviewNotify`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| T001 | `internal/ctr/tag.go:47` | `branch/case` | contract-open K4 codec: an empty directory-kind token remains distinct from f/l under length framing. K4 names abstract kinds, not their byte tokens; equal-tree/sensitivity predicates remain injective. |
| T002 | `internal/ctr/tag.go:30` | `branch/if` | test-added → killed: `TestK4ReviewRootModeAndUnreadable`. `.agent/plant`, rc 0; test FAIL (rc 1); round2. |
| T003 | `internal/ctr/tag.go:27` | `branch/if` | test-added → killed: `TestK4ReviewRootModeAndUnreadable`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| T004 | `internal/ctr/tag.go:38` | `branch/if` | test-added → killed: `TestK4ReviewRootModeAndUnreadable`. `.agent/plant`, rc 0; test FAIL (rc 1); faults replay. |
| T005 | `internal/ctr/tag.go:34` | `branch/if` | equivalent K4/Linux: WalkDir passes root or filepath.Join descendants using the same absolute/relative form; root is skipped. filepath.Rel cannot produce its volume/absolute mismatch error on these paths. |
| T006 | `internal/ctr/tag.go:42` | `branch/if` | equivalent K4/K11: Linux os.FileInfo from WalkDir.Info always carries *syscall.Stat_t. The changed non-Stat_t branch is unreachable on this supported platform. |
| T007 | `internal/ctr/tag.go:65` | `conditional/negated` | specified K4 path-byte order → conditional test supplied: TestK4ReviewProposedCodec passes target + kills this mutant. Stored as testdata fixture, not credited/adopted until MAIN fixes the byte codec (VF4). |
| T008 | `internal/ctr/tag.go:65` | `expression/comparison` | equivalent K4: relative paths are unique; <= and < give the same comparison for distinct entries. The sort never needs a different ordering for an entry against itself. |
| T009 | `internal/ctr/tag.go:26` | `expression/error-guard` | test-added → killed: `TestK4ReviewRootModeAndUnreadable`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| T010 | `internal/ctr/tag.go:37` | `expression/error-guard` | test-added → killed: `TestK4ReviewRootModeAndUnreadable`. `.agent/plant`, rc 0; test FAIL (rc 1); faults replay. |
| T011 | `internal/ctr/tag.go:33` | `expression/error-guard` | equivalent K4/Linux: WalkDir passes root or filepath.Join descendants using the same absolute/relative form; root is skipped. filepath.Rel cannot produce its volume/absolute mismatch error on these paths. |
| T012 | `internal/ctr/tag.go:44` | `numbers/decrementer` | test-added → killed: `TestK4ReviewEveryPermissionBit`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| T013 | `internal/ctr/tag.go:44` | `numbers/decrementer` | contract-open K4 codec: base-7 mode text is still injective over 07777 bits. Contract specifies permission bits, not numeric text radix. |
| T014 | `internal/ctr/tag.go:44` | `numbers/incrementer` | contract-open K4 codec: base-9 mode text is still injective over 07777 bits; same unstated radix boundary as T013. |
| T015 | `internal/ctr/tag.go:65` | `statement/remove` | specified K4 path-byte order → conditional test supplied: TestK4ReviewProposedCodec passes target + kills this mutant. Stored as testdata fixture, not credited/adopted until MAIN fixes the byte codec (VF4). |
| T016 | `internal/ctr/tag.go:49` | `statement/remove` | test-added → killed: TestK4ReviewSymlinkModesAreIgnored. Quiescent owned /proc/<pid>/fd symlink target stays fixed while its real mode changes 0500→0700. Target hash stays equal; mutant changes it. .agent/plant rc 0, test FAIL (rc 1). |
| T017 | `internal/ctr/tag.go:52` | `statement/remove` | contract-open K4 codec: empty file-kind token remains distinct from d/l in the framed tuple; literal kind bytes are unstated. |
| T018 | `internal/ctr/tag.go:27` | `statement/return` | test-added → killed: `TestK4ReviewRootModeAndUnreadable`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| T019 | `internal/ctr/tag.go:60` | `statement/return` | test-added → killed: `TestK4ReviewRootModeAndUnreadable`. `.agent/plant`, rc 0; test FAIL (rc 1). |
| T020 | `internal/ctr/tag.go:38` | `statement/return` | test-added → killed: `TestK4ReviewRootModeAndUnreadable`. `.agent/plant`, rc 0; test FAIL (rc 1); faults replay. |
| T021 | `internal/ctr/tag.go:34` | `statement/return` | equivalent K4/Linux: WalkDir passes root or filepath.Join descendants using the same absolute/relative form; root is skipped. filepath.Rel cannot produce its volume/absolute mismatch error on these paths. |
| T022 | `internal/ctr/tag.go:42` | `statement/return` | equivalent K4/K11: Linux os.FileInfo from WalkDir.Info always carries *syscall.Stat_t. The changed non-Stat_t branch is unreachable on this supported platform. |
| T023 | `internal/ctr/tag.go:65` | `statement/return` | specified K4 path-byte order → conditional test supplied: TestK4ReviewProposedCodec passes target + kills this mutant. Stored as testdata fixture, not credited/adopted until MAIN fixes the byte codec (VF4). |

## Findings (keyed by id)

### VF3 — major — missing-record test closes the peer, concealing an indefinite wait

- Targets: `internal/ctr/ctr.go:304`, `internal/ctr/ctr_test.go:187`; original `sendRecord` always closes the peer before `readRecord`. Thus the negative cases observe EOF, not the promised bounded wait when Podman exits but init/conmon retains the control socket.
- Target bytes: `_ = f.SetReadDeadline(time.Now().Add(time.Second))`; construction = `syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM|syscall.SOCK_CLOEXEC, 0)` + `os.NewFile`. These blocking NewFile descriptors do not support Go deadlines. The ignored error leaves `f.Read` unbounded.
- Red unit command (target production source + review test): `TMPDIR=<worktree>/.tmp go test -count=1 -run '^TestK7ReviewMissingRecordHasBound$' -v ./internal/ctr`; rc 1. Output: `blocking socket's deadline support: file type does not support deadline` / `missing status record blocked past its one-second deadline while the peer stayed open` / `--- FAIL: TestK7ReviewMissingRecordHasBound (5.02s)`.
- Real CLI red: `TestK8ReviewPodmanGoneFirst` starts the TERM-ignoring container helper, observes readiness, enumerates the one direct Podman child, kills that child, then expects rc 125. Target instead times out after 10 s. Test cleanup must kill offload before liveness ends the container. Command = private-temp-namespace runner + `go test -count=1 -race -run 'TestK(7ReviewImageENOEXECAndFD3|8ReviewAllSignals|8ReviewPodmanGoneFirst)$' -v ./test/e2e`; only the Podman-first case fails. Log = `.scratch/worktrees/reviewer-4/.scratch/review/new-real-green.log`.
- Acceptance: K7's no-record path must terminate while the peer remains open; K8 Podman-first death returns 125 and closes liveness so every container process ends. Preserve complete-record precedence. Red inputs = `internal/ctr/reviewer4_missing_record_test.go` + `test/e2e/reviewer4_ctr_test.go` in the review worktree.

### VF2 — minor — contract-relevant mutations survive the target tests

- Target: `internal/ctr/ctr_test.go:73`; exact anchor: `"other x bit":   func(d string) { must(t, os.Chmod(filepath.Join(d, "a/b"), 0o654)) },`. That case changes group execute (0010), not other execute (0001). The suite never varies the latter bit although K4 includes all 07777 bits.
- Demonstration: T012 changes `st.Mode&0o7777` → `st.Mode&0o7776`; original default-operator sweep reports ESCAPED. `TestK4ReviewEveryPermissionBit` kills it via `.agent/plant` (`--- FAIL: TestK4ReviewEveryPermissionBit`), while the unmutated target passes. Full independent survivor inventory below also records missing failure/build/status/signal assertions, not merely the one permission example.
- Acceptance: merge the contract-sensitive added probes after their listed red/green evidence; retain each equivalent/open survivor ruling. The clause map separates uncovered cases from independently demonstrated defects. VF3 is an actual behavior failure; most VF2 survivors are missing regressions around otherwise correct target code.

### VF1 — minor — failed process probes count as successful teardown

- Target: `test/e2e/ctr_test.go:442`; exact anchor: `if exec.Command("pgrep", "-f", "ctrhelper ignore "+marker).Run() != nil {`.
- K8's death checks call `markerGone`; its predicate accepts rc 2, failed exec, and other probe errors as the same outcome as `pgrep` rc 1 (no match). A failed observation therefore credits SIGKILL/leftover/outer-run teardown while the process may remain live.
- Committed red input: `test/e2e/reviewer4_probe_test.go`. Command: `go test -count=1 -race -run '^TestReviewV6ProbeFailureIsNotAbsence$' -v ./test/e2e`; rc 1. Final output: `absence assertion accepted a failed process probe:` / `probe setup complete: expected visible count=1; injected probe rc=2` / `returned-as-gone: helper pid 919691 is live; positive count=1; probe rc=2` / `--- FAIL: TestReviewV6ProbeFailureIsNotAbsence (1.35s)`. Parent requires the child's setup marker, so an unrelated child/build failure cannot count as a successful oracle. Log = `final-helper-usage-red.log`.
- Positive control: real `pgrep` first returned exactly the helper PID (count 1); after the fake probe returned 2, `kill(pid, 0)` still succeeded. Probe cleanup killed/reaped only its own helper.
- Acceptance: absence requires a successfully executed `pgrep` with rc 1; other errors fail the check. Establish probe visibility before signaling; check `podman ps` status before trusting empty output in the adjacent removal assertion too.

### VF4 — minor — exact tag-order credit needs a byte-codec ruling

- Target: `.agent/contract.md:71`; exact bytes: `each length-framed as kind (dir | file | symlink) + permission bits`. K4 specifies semantic fields + path-byte order, but not the kind tokens, integer radix, or length-prefix byte encoding. Those choices change the SHA even when every stated sensitivity/equality property holds.
- T007 reverses ordering; T015 removes sorting; T023 makes the comparator always false. Target tests + the added semantic probes let all three survive. A canonical-stream oracle needs a concrete codec; treating implementation-chosen d/f/l + octal + decimal-colon framing as already owned by the contract would invent part of its expected output.
- Conditional evidence supplied, not credited as an adopted check: `internal/ctr/testdata/tag-codec/candidate_test.go`. Copy into `internal/ctr/reviewer4_codec_candidate_test.go`, then `go test -count=1 -run '^TestK4ReviewProposedCodec$' -v ./internal/ctr` → target rc 0 (0.002 s). The same case kills T007/T015/T023 via `.agent/plant`; it also kills four currently contract-open encoding mutants T001/T013/T014/T017. `codec-proposal.log` records all seven. The temporary source copy was removed; the committed fixture is explicitly nongating.
- Acceptance recommendation: state the existing codec in K4 (d/f/l; octal mode with no prefix; each field = decimal byte length + `:` + bytes; first eight SHA-256 bytes as lower-case hex), activate the supplied fixture, then rerun the order/codec plants. MAIN owns this contract ruling. V7 is not cleared until that decision; no claim that the conditional three already earned mutation credit.

### VF5 — minor — help loses to an earlier malformed flag

- Targets: `internal/cli/ctr.go:39`, `test/e2e/ctr_test.go:552`. Exact production anchor: `if err := fs.Parse(args); err != nil {`. Existing `TestK1HelpWinsOverBadSwitch` tests environment precedence, not parse-error precedence.
- K1 says help wins over any other problem. A prior malformed option stops flag parsing before the help request is seen: `ctr --bogus --help` and `ctr --bogus -h` return 2 on stderr. The help-first permutations return 0, so the positive control proves help support itself is present.
- Red command: `go test -count=1 -race -run '^TestK1ReviewHelpWinsOverBadFlag$' -v ./test/e2e`; rc 1. Output: `args=["ctr" "--bogus" "--help"] rc=2 stderr-first="offload: flag provided but not defined: -bogus"`; same result for `-h`. Input = `test/e2e/reviewer4_usage_test.go`; combined final evidence = `final-helper-usage-red.log`.
- Acceptance: help among wrapper options takes precedence over parse errors; preserve the `--` / first-CMD boundary so CMD's own help words remain verbatim (`TestK1ReviewNativeWordsAndENOEXEC` pins it).

## Targeted e2e-only plants

| id | target | contract assertion | result |
|---|---|---|---|
| P01 | `internal/ctr/init.go` | K8 fd3 noninheritance | `P01: killed by --- FAIL: TestK7ReviewImageENOEXECAndFD3 (1.29s)`; wrapper rc 0 |
| P02 | `internal/ctr/init.go` | K8 PID 1 refusal | `P02: killed by --- FAIL: TestK8ReviewInitRefusesOutsidePID1 (0.00s)`; wrapper rc 0 |
| P03 | `internal/ctr/init.go` | K8 complete signal set | `P03: killed by --- FAIL: TestK8ReviewAllSignals (4.10s)`; wrapper rc 0 |
| P04 | `internal/ctr/init.go` | K8 own CMD group | `P04: killed by --- FAIL: TestK8ReviewWholeCommandGroup (11.02s)`; wrapper rc 0 |
| P05 | `internal/ctr/init.go` | K8 entire CMD group | `P05: killed by --- FAIL: TestK8ReviewWholeCommandGroup (3.09s)`; wrapper rc 0 |
| P06 | `internal/ctr/init.go` | K8 leftover TERM | `P06: killed by --- FAIL: TestK8ReviewLeftoverTERM (3.07s)`; wrapper rc 0 |
| P07 | `internal/ctr/init.go` | K8 liveness EOF kills | `P07: killed by --- FAIL: TestK8ReviewLivenessEndsContainer (16.19s)`; wrapper rc 0 |
| P08 | `internal/ctr/init.go` | K8 grace period | `P08: killed by --- FAIL: TestK8LeftoverProcessesEndAfterCommand (1.07s)`; wrapper rc 0 |
| P09 | `internal/ctr/init.go` | K7 recorded command status | `P09: killed by --- FAIL: TestK7StatusPassesThrough (1.28s)`; wrapper rc 0 |
| P10 | `internal/ctr/init.go` | K7 ENOEXEC status | `P10: killed by --- FAIL: TestK7ReviewImageENOEXECAndFD3 (1.01s)`; wrapper rc 0 |
| P11 | `internal/cli/ctr.go` | K10 ordered digest equality | `P11: killed by --- FAIL: TestK10ReviewOrderedLayersAndHosts (0.77s)`; wrapper rc 0 |
| P12 | `internal/cli/ctr.go` | K10 binary mismatch fails | `P12: killed by --- FAIL: TestK10CheckComparesHosts (1.94s)`; wrapper rc 0 |
| P13 | `internal/cli/ctr.go` | K10 pool order | `P13: killed by --- FAIL: TestK10ReviewOrderedLayersAndHosts (0.40s)`; wrapper rc 0 |
| P14 | `internal/cli/ctr.go` | K1 literal command words | `P14: killed by --- FAIL: TestK1ReviewNativeWordsAndENOEXEC (0.00s)`; wrapper rc 0 |
| P15 | `internal/cli/ctr.go` | K2 native exec identity | `P15: killed by --- FAIL: TestK2SwitchOffExecsInPlace (0.01s)`; wrapper rc 0 |
| P16 | `internal/cli/ctr.go` | K10 parallel host operations | `P16: killed by --- FAIL: TestK10ReviewHostsStartInParallel (5.33s)`; wrapper rc 0 |
