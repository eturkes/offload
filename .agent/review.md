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
