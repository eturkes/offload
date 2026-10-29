# reviewer-5 verdicts — closing diff fed6498..cac258b

Check set fixed before target reads = M1–M7. Target = `cac258b87885d725fbbf45d78afa10e1f9a1cd60`; root `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-5`. Primary spec checked separately through `46aa85e`; later implementation changes excluded. Verdict = `pass` | `finding + ids`. Write allowlist = this validator input; target + git read-only. Human-facing scope = README + CLI usage/help/errors; M2 grades README, CLI observations outside another row → register.

| id | check | verdict | evidence |
|----|-------|---------|----------|
| M1 | text artifacts agent-optimized (comments, rules, contract, commit subjects/bodies, scratch excluded) per CLAUDE.md Authoring | finding F1 | Minor: live contract retains amendment provenance + ignored scratch pointer; remaining new internal text dense, boundary/constraint-oriented. Rule provenance → F2. |
| M2 | README.md in human-facing ASD-STE100 register per CLAUDE.md Authoring Human-facing | pass | Full README: direct instructions, condition-first alternatives, stable terminology; no register defect worth blocking. Guarantee accuracy outside M2 → R1. |
| M3 | comments = why not what; dead code/unused helpers absent | pass | Production Go + embedded shell + unit tests + e2e harness read; comments explain watcher/lock/git/cgroup constraints or document boundary formats. All Go helper names have non-comment references; no dead helper found. Type-aware verification unavailable. |
| M4 | commit subjects '<scope>: <cause> → <fix>' + bodies (measurements, SHAs, Teammates line) | finding F3 | Minor: 9/9 subjects + teammate lines conform; 8/9 bodies lack commit revision refs. Two red claims identify neither runnable red revision nor complete command. |
| M5 | .agent/spec.md shape + currency (Tasks ticked, Artifacts current, Decisions binding) | finding F4 | Minor: Decision `address = any ssh destination` contradicts P2's narrower address domain. Both spec shape checks rc 0; primary ship/copy ticks current; phase-close/prototype retirement correctly remain open. |
| M6 | .claude/rules/ files: live law + why, paths: frontmatter where area-bound, no origin stories | finding F2 | Minor: gate-checks retains nixfmt adoption history + undefined pre-nixfmt seed. Gate law bare; machine/check law scoped; quoted paths structurally parse. |
| M7 | .gitignore current; temp artifacts/dangling files absent from the tree | pass | 65 tracked target entries present; 0 tracked symlinks; 5/5 cache/secret/scratch ignore probes match. Primary nonignored status clean at 46aa85e; active ignored caches/worktrees excluded from close-cleanup judgment. |

## F1 — minor — amendment history in the live contract

- Anchor: `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-5/.agent/contract.md:3` → `Amended after consultant-1 (` followed by `.scratch/consult-1/verdicts.md`.
- Clause: `CLAUDE.md:29` = `omit provenance — dates, verification/discovery events, origin stories`; `CLAUDE.md:28` = compress + prune obsolete/redundant internal text.
- Wrong: amendment event changes no predicate; its scratch pointer is not committed contract material. The commit body already owns the rulings. Keeping it in the live introduction adds history, not the contract a fresh agent needs.
- Proof: `rg -Fn 'Amended after consultant-1' /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-5/.agent/contract.md` → line 3, rc 0; same bytes rematched from `git show cac258b:.agent/contract.md`.
- Rewrite: retain the first three sentences of line 3; delete `Amended after consultant-1 (` through the line's final `).`. Predicate ids, tiers + precedence unchanged.
- Acceptance: live introduction states ids/tiers/precedence alone; amendment provenance stays in git.

## F2 — minor — historical rule text instead of a runnable nixfmt seed

- Anchors: `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-5/.claude/rules/gate-checks.md:19` → `the six pre-nixfmt`; line 22 → `- nixfmt adoption:`.
- Clause: `CLAUDE.md:34` = `Each rule = live law + the why` and `its origin story ... → the commit body that lands it`; `CLAUDE.md:80` = rules upkeep removes origin stories.
- Wrong: adoption/AST-equivalence row reports a completed reformat, not a future obligation; `the six pre-nixfmt` names neither a current fixture nor its revision. A fresh agent cannot rerun that seed from this rule alone. This does not object to the other live firing recipes or their expected rc values.
- Proof: `rg -Fn -e 'the six pre-nixfmt' -e 'nixfmt adoption:' /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-5/.claude/rules/gate-checks.md` → lines 19,22, rc 0; both rematched at cac258b. Commit be205eb already records adoption/AST evidence.
- Rewrite line 19: `- nix: \`printf '{a=1;}\\n' | .agent/tool nixfmt --check -\` → rc 1; formatted stdin → rc 0.` Delete line 22; keep historical evidence in be205eb.
- Replacement seed verified read-only with the existing pinned binary: `printf '{a=1;}\n' | /home/eturkes/Projects/offload/.tools/nixfmt-1.5.0/nixfmt --check -` → `<stdin>: not formatted`, rc 1. Positive control = same stdin piped through `nixfmt -` then `nixfmt --check -` → rc 0. Initial bare-stdin invocation emitted a deprecation warning; final command explicitly uses `-`.
- Acceptance: rule carries an immediately regenerable bad input, not the adoption event.

## F3 — minor — red-test commit records omit reproducible revisions

- Clause: `CLAUDE.md:18` = commit body `measurements + SHAs + each teammate ...`; `CLAUDE.md:48` = red test counts once `that revision + command recorded in the unit's commit body`.
- Primary anchors:
  - 582209c, subject = ``offload: no reviewed contract, pool loader or host probe for the Go build → `.agent/contract.md` (C/P/R/S/L/Y/X predicates, consultant-reviewed) + `internal/pool` (strict TOML) + `internal/probe` + shared ssh options``; body exact bytes: `Red on pre-fix code (same tests, fixes reverted in place): TestParse`.
  - 7b96bf0, subject = ``mirror: no git-exact sync (prototype's rsync `:- .gitignore` filter dropped tracked files matching ignore patterns) → `internal/mirror`: git-listed upload set U, `.git` mirror, remote output list download, plain-mode fallback, --get scoping, remote-eligibility checks``; body exact bytes: `Red-first: TestInspectRefusesBrokenRepo red before the damaged-repo rule (plain mode chosen), green`.
- Wrong: these bodies assert red→green evidence but identify neither a replayable red revision/patch nor the full test command. `fixes reverted in place` cannot identify the historical target. This is an evidence-record gap, not a claim that the tests never failed.
- Proof: `git -C /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-5 show -s --format='%h%n%s%n%b' 582209c 7b96bf0` contains both anchors. Full 9-commit log inspected. Body commit-ref scan (7–40 hex tokens resolved with `git rev-parse --verify <token>^{commit}`): fc3c575 has 282d82f + 0b8eb3b; e45fb7b, f44db3a, be205eb, 582209c, 282d82f, 7b96bf0, fcf6e75, cac258b have none. Positive control = fc3c575's two resolving refs. cac258b's cross-repo e7d3007 is in its subject, not body. Missing optional comparison SHAs in non-red units are not independently graded as eight defects.
- Rewrite template for each red claim: `Red: <actual red revision or preserved patch ref>; <full command>; rc <n>; <distinctive failure>. Green: <tested implementation revision>; <same command>; rc 0. Teammates: <name, role, verdict>.` Recover actual evidence; do not invent an old SHA or infer which in-place edits ran. If unrecoverable, replace the claim with `Historical red revision/command unconfirmed; <new reproducible evidence ref>` and record MAIN's ruling in the close evidence/commit. Historical commit repair is MAIN's decision, not authorization to rewrite history.
- Unconfirmed: original in-place red states; looked in the nine commit bodies + committed tests. Scratch transcripts not searched (brief excludes scratch text); no red tests rerun under this read-only brief.
- Acceptance: each counted red claim has a durable runnable target + exact command, or an explicit evidence-limit ruling instead of credited historical red coverage.

## F4 — minor — stale authoritative pool-address Decision

- Anchor: `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-5/.agent/spec.md:23` and `/home/eturkes/Projects/offload/.agent/spec.md:23` → `` `address` = any ssh destination ``.
- Clause: `CLAUDE.md:74` = spec lines bind current/future work; superseded text dies with its replacement; `CLAUDE.md:81` = spec current at every checkpoint. Contract introduction makes Decisions authoritative.
- Wrong: `.agent/contract.md:18` limits addresses to `[user@]host`/alias and explicitly excludes `:`, `/`, whitespace + leading `-`. README requires an alias for ports/IPv6. `any ssh destination` also admits valid SSH URI/IPv6 forms that this contract intentionally rejects. The authoritative state therefore overstates its accepted domain.
- Proof: `rg -Fn 'any ssh destination' /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-5/.agent/spec.md /home/eturkes/Projects/offload/.agent/spec.md` → line 23 in each; P2's full address predicate at `.agent/contract.md:18` read + matched. Finding anchors rematched from cac258b.
- Rewrite: `- Pool = TOML \`[[host]]\` tables (\`name\`, \`address\` = \`[user@]host\` or an ssh_config alias, \`slots\`); ports + IPv6 → alias. SSH = BatchMode + accept-new host keys.`
- Acceptance: spec + P2 + README name the same domain; this proposes documenting the existing restriction, not silently expanding implementation behavior.

## Register — outside fixed rows

- R1 — README guarantee overreach, not M2 register/style: `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-5/README.md:93` exact anchor `A file that you changed here during the run is not replaced.` Contract Y6 (`.agent/contract.md:48`) protects only newer local regular files; symlinks + type changes take the remote version. Proposed rewrite: `A regular file with a newer local modification time is not replaced. Symlinks and file-type changes take the host's version.` Text conflict confirmed; behavior not rerun. MAIN notified.

## Checks + limits

- Read: all changed non-generated internal docs/rules/config; README; production Go + embedded shell; all unit tests; e2e harness + fixture; complete machine diff; 9 commit subjects/bodies. Other e2e case bodies checked by comment/declaration/reference queries, not a correctness pass. Legal boilerplate + generated dependency sums not prose-rewritten; original unchanged machine/agent instructions read only where governing this lens.
- Read-only checks: spec-lint on target + primary → rc 0; stdin spec with one history line → rc 1 + expected Phase finding. Quoted-rule paths structurally parsed as the exact `paths:` + JSON-string-list subset (2 scoped files, 1 bare). Generic YAML parser unavailable: Python `import yaml` → ModuleNotFoundError; no install attempted under write allowlist.
- Go LSP documentSymbol on internal/cli/cli.go → `No LSP server available for file type: .go`. Dead-helper assessment = source/reference inspection, not a type-aware unused check. Full lint/test/gate/mutation suite not run; target never modified.
- Hygiene bound: tracked target + primary nonignored status only; active ignored `.scratch`/`.tools` excluded. No claim about unrelated host files or future phase-close cleanup. No remote calls.
- Seed validator: expected 7 unresolved table rows → count 7, rc 0. Final unresolved-row validator → 0, rc 1 (expected no-match); `grep -c '^| M[1-7] |'` → 7, rc 0. All 7 rows adjudicated; target status empty, rc 0.
