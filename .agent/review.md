# Review ledger — IMPLEMENT close (diff c32846f..HEAD)

Check sets fixed before reading: lenses A (correctness/spec, reviewer-1), B (claim soundness, reviewer-2), G (guarantee-vs-claim, reviewer-3), V (verification integrity, reviewer-4), M (CLAUDE.md conformance, reviewer-5) — row texts in `.scratch/reviewer-N/verdicts.md`; S (security, MAIN alone). Row = verdict + MAIN's ruling; `open` rows block the phase close.

| id | check | verdict | ruling / evidence | state |
|----|-------|---------|-------------------|-------|
| S1 | pool-host responses → local tree (download list + contents, remote-controlled) | accepted | Pool hosts = dedicated, agent-owned (Decisions) ⇒ trusted; a compromised host already holds the whole repo + .git. rsync 3.5.1 receiver (post-CVE-2024-12086/7/8 path fixes), no `-K`, symlinks arrive as links | closed |
| S2 | probe output → parser | pass | exactly 4 fields; CPUs ≥ 1, load finite ≥ 0 (NaN/Inf rejected, red-first), mem/GPUs ≥ 0; bad output = host down | closed |
| S3 | remote status record → exit status | pass | exact string match against a crypto/rand nonce (`rand.Text`) | closed |
| S4 | project file names → lists, rsync, remote prune | pass | NUL-framed, never shell-parsed; remote: quoted, `rm -f --`, `in_tree` refuses `..`/absolute/symlinked ancestors (tests: TestRemotePrune, TestUploadList, real-host victim check) | closed |
| S5 | command words → remote `bash -lc` | pass | `remote.Join` single-quote quoting; 14 hostile words round-trip through bash (TestJoinRoundTripsThroughBash); e2e TestX2 injection words stay literal | closed |
| S6 | runtime lock/owner files under the `/tmp` fallback (no XDG_RUNTIME_DIR) | fixed | planted symlink at `slotN`/`slotN.owner` made offload acquire through it + truncate the target → `O_NOFOLLOW` on both opens; TestSlotFilesNeverFollowSymlinks red before, green after | closed |
| S7 | SSH host keys | accepted | BatchMode + accept-new (trust on first use) = owner Decision; LAN-only hosts | closed |
| S8 | offloaded code on nixbox runs as `offload` (passwordless sudo); trading's wallet key lives there under `eturkes` | accepted | owner ruling "Add eturkes, keep sudo" after MAIN stated offloaded code could read the key | closed |
| S9 | Codexify on nixbox: ChatGPT account → shell as `eturkes` (wallet key readable) | accepted | owner request ("set up Codexify on it, so I can iterate on that repo through ChatGPT"); outbound OpenAI tunnel, no inbound port | closed |
| S10 | secrets in logs/owner lines | accepted | slot owner lines carry command words (heavy parity); offload logs nothing else; never pass secrets as arguments | closed |
| S11 | CI supply chain | pass | actions pinned by commit SHA, `contents: read`, `persist-credentials: false`, CodeQL `security-events: write` scoped to its job; zizmor clean; Dependabot gomod + actions weekly | closed |
| S12 | dependencies | pass | one module (pelletier/go-toml/v2 v2.4.3, 0 transitive); govulncheck clean in the gate | closed |
| R1 | P2 explicit `slots = 0` accepted as 2 (A-F1, B-F2, V-F05) | fixed | `*int` decode: absent → 2, explicit < 1 refused; TestReviewP2ExplicitZeroSlots + pool `zero slots` case | closed |
| R2 | R2 fresh `status` reports slots held before their dir exists (A-F2, B-F3) | fixed | `Holders` skips never-created slot files; TestReviewR2FreshStatusHasNoHolders | closed |
| R3 | C3 signal during inspection/upload prep → 125 (A-F3, G-F7) | fixed | every runtime failure yields to a pending signal; TestReviewC3SignalDuringInspection, TestG7SignalDuringInspectionWins | closed |
| R4 | Y6/Y10 failed remote output listing → rc 0, nothing downloaded (A-F4, G-F3, V-F06) | fixed | list written only after `git ls-files` succeeds, else download fails → 125; TestReviewY6DownloadListFailure, TestG3DownloadListFailureIsNotSuccess, TestReviewY10OutputListFailureOverridesCommand | closed |
| R5 | Y8 `--get dir` nests `dir/dir` on a repeat (A-F5) | fixed | destination = parent dir; TestReviewY8GetExistingDirectory | closed |
| R6 | X6 status lookup ignores signals (A-F6, G-F7b) | fixed | bound to the run's context; TestReviewX3SignalDuringStatusRecovery, TestG7SignalDuringStatusLookupCancels | closed |
| R7 | L4 stale GPU inventory exits 125 while a host is down (A-F7) | fixed | failed probe forgets the host's inventory; "no GPU" needs every host up now; TestReviewL4PreviouslyUpHostGoesDown | closed |
| R8 | S4 pool owner text omits the host (A-F8) | fixed | ` <host>: <root> <CMD…>`; TestReviewS4PoolOwnerNamesHost | closed |
| R9 | Y10 failed prune still runs the command (A-F9) | fixed | `rm` failure → 125 before CMD; TestReviewY10FailedPrunePreventsCommand | closed |
| R10 | X1 EOF before the leader's setsid() → command survives (A-F10, G-F1) | fixed | EOF path also signals the leader pid while it is still the watcher's child (no reused-pid kill); TestReviewX1DeathBeforeSessionCreation, TestG1CancellationBeforeSetsidKillsSession | closed |
| R11 | R1 probe outlives 8 s when ssh descendants hold its pipes (A-F11) | fixed | deadline 7.5 s + 0.5 s WaitDelay; TestReviewR1ProbeDeadlineWithInheritedOutput | closed |
| R12 | Y5 upload replaced a remote ignored dir with a local file, deleting its contents (G-F5) | fixed | U uploads without `--delete` (type conflict fails closed → 125); `.git` mirrored by its own rsync; TestG5UploadTypeChangeKeepsIgnoredRemoteFile | closed |
| R13 | README promised every local edit survives the download (G-F4, B-F1) | text | README states the `--update` guarantee (newer regular files stay; symlinks + type changes take the host's); the two counterexample tests retire with the claim | closed |
| R14 | README slot count read as a hard bound across controller death (G-F6) | text | README: a killed offload frees its host slot at once, its command needs ≤ 10 s more (S5); test retired with the claim | closed |
| R15 | spec Decision `address = any ssh destination` vs P2 (B-F4, M-F4) | text | Decision narrowed to `[user@]host` or an alias; IPv6 counterexample test retired with the claim | closed |
| R16 | gate firing records lacked a `build` seed (B-F5) | fixed | `//go:build !cgo` type error → static build rc 1 while vet + race tests stay green (gate-checks.md) | closed |
| R17 | fake ssh ignored `-n` + every `-o` option (V-F01, V-F04 X5) | fixed | harness honors `-n`, refuses a call missing any X5 option (rc 254) ⇒ every e2e test checks X5; harness self-tests `missing-x5-option-refused`, `null-stdin` | closed |
| R18 | X6 transport-failure fallback never exercised (V-F02) | fixed | `.run-fail` marker fails the run-phase call alone; TestX6TransportFailureDuringRunIs125 | closed |
| R19 | R2 status numbers never asserted (V-F03) | fixed | TestR2StatusShowsProbeNumbers (row = up 0/3 12 3.00 25% 1.0GB 2) | closed |
| R20 | L3 30 s down-host throttle + S4 controller workdir lock untested (V-F04) | fixed | TestL3DownHostProbedAtMostEvery30s (1 contact in ~6 s), TestS4SameTreeRunsSerializeAtController | closed |
| R21 | TestRemoteLockSerializes timed its own fixture (V-F07) | fixed | elapsed measured before joining the holder | closed |
| R22 | contract carried amendment provenance (M-F1) | fixed | line removed; history lives in commit bodies | closed |
| R23 | gate-checks.md carried nixfmt adoption history + an undefined seed (M-F2) | fixed | history line removed (commit c9539b5 keeps it); nix seed = `{ a=1; }` → rc 1 | closed |
| R24 | 8/9 commit bodies name no red revision SHA (M-F3) | accepted | history stays unrewritten; this close commit records revision + command for every review red | closed |
| R25 | pass rows: A-none; B6; G2; V3–V6; M2, M3, M7 | pass | see `.scratch/reviewer-N/verdicts.md` (tags archive/reviewer-1…5) | closed |
