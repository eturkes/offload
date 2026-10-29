# IMPLEMENT — whole product (closed 2026-10-21)

Units (sha = landing commit):

- [x] f44db3a nixbox: `eturkes` user + idle slice + trading/Codexify deps → deploy
- [x] cac258b trading → nixbox: copy every file, checksum-verify, rebuild env on nixbox, trading gates there
- [x] 46aa85e trading on nixbox: deployed paused + Luna set up (owner activates via ChatGPT); nixbox state documented + committed in trading (trading 04a1a31)
- [x] be205eb foundation: Go module, LICENSE + NOTICE, entry point, one test, gate (`.agent/gate`), CI + Dependabot, gate rule
- [x] 582209c contract (`.agent/contract.md`, consultant-reviewed) + pool file + probe
- [x] 282d82f slots: heavy interop + pool slots
- [x] 7b96bf0 sync engine: git-exact up/down, remote deletion pass, non-git fallback, repo root, worktree detection
- [x] fc3c575 run: watcher (session kill, remote lock), local + remote execution, placement loop, CLI + `offload status`, `--get`
- [x] cac258b ship: installed binary at `~/.local/bin/offload`, global `CLAUDE.md` rule re-derived (agents repo), README
- [x] (close) controller trading copy deleted (owner confirmed immediately before; nixbox = live copy; controller project snapshots keep older copies)
- [x] (close) review: 5 lens reviewers + security lens, ledger .agent/review.md (37 rows, 0 open); prototype retired (tag prototype-whole-product); Phase → MAINTAIN

- [x] (after close) Codexify on nixbox: owner ran quickstart + Codex login; a key ID pasted as the runtime key gave 401 on every tunnel poll → owner re-ran quickstart with the sk- secret; doctor healthy, polls succeed

Review ledger: .agent/review.md. Owner rulings: commit e45fb7b body + spec Decisions.
