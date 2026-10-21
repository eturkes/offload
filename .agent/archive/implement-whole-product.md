# IMPLEMENT — whole product (closed 2026-10-21)

Units (sha = landing commit):

- [x] 0886be2 nixbox: `eturkes` user + idle slice + trading/Codexify deps → deploy
- [x] 47a8417 trading → nixbox: copy every file, checksum-verify, rebuild env on nixbox, trading gates there
- [x] 97f8ffe trading on nixbox: deployed paused + Luna set up (owner activates via ChatGPT); nixbox state documented + committed in trading (trading 04a1a31)
- [x] c9539b5 foundation: Go module, LICENSE + NOTICE, entry point, one test, gate (`.agent/gate`), CI + Dependabot, gate rule
- [x] e7cb43a contract (`.agent/contract.md`, consultant-reviewed) + pool file + probe
- [x] d4d4828 slots: heavy interop + pool slots
- [x] 5bde281 sync engine: git-exact up/down, remote deletion pass, non-git fallback, repo root, worktree detection
- [x] 4ff8ad4 run: watcher (session kill, remote lock), local + remote execution, placement loop, CLI + `offload status`, `--get`
- [x] 47a8417 ship: installed binary at `~/.local/bin/offload`, global `CLAUDE.md` rule re-derived (agents repo), README
- [x] (close) controller trading copy deleted (owner confirmed immediately before; nixbox = live copy; controller project snapshots keep older copies)
- [x] (close) review: 5 lens reviewers + security lens, ledger .agent/review.md (37 rows, 0 open); prototype retired (tag prototype-whole-product); Phase → MAINTAIN

- [x] (after close) Codexify on nixbox: owner ran quickstart + Codex login; a key ID pasted as the runtime key gave 401 on every tunnel poll → owner re-ran quickstart with the sk- secret; doctor healthy, polls succeed

Review ledger: .agent/review.md. Owner rulings: commit 11d6dd7 body + spec Decisions.
