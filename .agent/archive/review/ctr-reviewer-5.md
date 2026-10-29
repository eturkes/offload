# reviewer-5 — lens M: CLAUDE.md conformance

Target = `5c11b58..57903b5` (0d14378, be96752, 57903b5) + staged `.agent/review.md` → `.agent/archive/review/ledger-implement-whole-product.md`. Check set = seeded M1–M7, fixed before target reads. **7/7 adjudicated; 6 pass, 1 finding (minor).** Target read-only; this validator input alone written.

| id | check | verdict (pass / finding) | finding ids | evidence (file:line, command + output) |
|----|-------|---------------------------|-------------|----------------------------------------|
| M1 | Authoring: agent-optimized internal text (comments, rules, spec, commit bodies, archive); human-facing README + CLI help in the ASD-STE100 register | pass | — | `README.md:31,151–242` + `internal/cli/cli.go:30–62`: direct instructions, stable terminology, explicit outcomes. Internal K predicates/rules/comments use compact identifiers + notation. Retired history is in `.agent/archive/`, as Session flow prescribes. No substantive register defect. |
| M2 | Engineering: comments = why only; dedupe (helpers reused); no dead code; idiomatic + concise | pass | — | `internal/mirror/mirror.go:43,74,266`: root discovery + Y2 naming shared with ctr; `internal/cli/ctr.go:86–89`: native lookup reuses `checkLocalCommand`; `internal/ctr/ctr.go:111,121`: slot-lock facility + remote command cancellation reused. PID1/signal/FD/mount comments explain constraints. New implementation + test helper callers inspected; no merge-blocking cleanup defect. |
| M3 | .agent/spec.md shape + content (Tasks ticked with SHAs, Decisions current + no superseded text, Artifacts = entry points + run commands, Phase) | finding | M-F1 | `.agent/spec.md:44` still marks the completed 57903b5 unit open. Other shape/content checks pass: 5 sections, runnable artifact entries, current ctr decisions, prototype-only decision archived, routed Phase retained; `sh .agent/spec-lint .agent/spec.md` rc0. |
| M4 | .claude/rules/*: live law + needed why, no origin stories, paths: frontmatter scope right | pass | — | All four rule files read. `code.md:1–6`, `machines.md:1–4`, `gate-checks.md:1–8` area-scoped; `gate.md` repo-wide. Added `code.md:12–13` = PID1/signal/static-build + isolated-Podman-test constraints, not incident history. Frontmatter uses quoted repo-relative globs. |
| M5 | commits: subject <scope>: <cause> → <fix>; body measurements + SHAs + red revision + teammates | pass | — | `git log --reverse --format=fuller 5c11b58..57903b5`: all 3 subjects use cause → fix; bodies record gate outcomes/timings + teammate verdicts (`consultant-1`, then `none`). 0d14378 names red base 5c11b58, copied test paths, exact `go test` command + 32 FAIL; be96752 names cement/agents revisions + live checks. No claim that pending unit-test mutation credit is complete; see register R-M1. |
| M6 | Session flow: prototype retired (tag + git rm + archive), deferred rows each carry an acceptance check, structures per rules | pass | — | `show-ref --tags prototype-pool-run-environment` → be96752; `ls-tree` retains all 7 retired prototype paths; target deletes those 7. `.agent/archive/pool-run-environment.md:3–9` preserves retired entry/decision. `.agent/deferred.md`: 11 entries, 11 `Accept` clauses. Staged old ledger archive = R100; every old row closed. New closing ledger is in-flight, not claimed complete. |
| M7 | Execution: .gitignore current, no stray artifacts tracked, gate command + fast gate recorded | pass | — | `.gitignore:1,3,9`: scratch/local policy/tool outputs ignored; `git check-ignore` returns all 3 sampled paths. Full 30-path target inventory has no stray generated artifact. `.claude/rules/gate.md:3–4` records full + fast gates; `.agent/gate:25–63` implements both + timings/slowest tests. Dominant test-step optimization deferred with equal-strength acceptance (`.agent/deferred.md:13`). |

## Findings

### M-F1 — minor — committed README/prototype unit remains open in the live checkpoint

- **Location:** `/home/eturkes/Projects/offload/.agent/spec.md:44` at 57903b5; disk anchor matched too.
- **Exact bytes:** `- [ ] README + prototype retired`
- **Clause:** `/home/eturkes/Projects/offload/CLAUDE.md:81`, Session flow / Checkpointing: `` `spec.md` current (`Tasks` ticked) at every commit ``. Also `CLAUDE.md:74`: Tasks use `- [x] <sha>` once committed. Both clause searches matched, rc0.
- **Why wrong:** 57903b5 completes this exact unit (README rewrite, seven prototype deletions, archive, deferred rows), yet the sole attached resume state still presents it as unfinished. A resumed session cannot distinguish finished work from the remaining closing review without re-deriving the commit. This is checkpoint correctness, not wording preference.
- **Proof commands + output:**

  ```text
  rg -Fn -- '- [ ] README + prototype retired' /home/eturkes/Projects/offload/.agent/spec.md
  44:- [ ] README + prototype retired (`prototype-pool-run-environment` tag) + deferred rows

  git -C /home/eturkes/Projects/offload show 57903b5:.agent/spec.md | rg -nF -- '- [ ] README + prototype retired'
  44:- [ ] README + prototype retired (`prototype-pool-run-environment` tag) + deferred rows

  git -C /home/eturkes/Projects/offload show --format='%h %s' --stat 57903b5
  57903b5 docs: README still installed offload by hand on the controller alone ... → README: tools/install ...; prototype/container retired ...
   .agent/archive/pool-run-environment.md |   9 +++
   .agent/deferred.md                     |   3 +-
   .agent/spec.md                         |   7 +-
   README.md                             | 105 ++++++++++++++++++++++++++++--
   prototype/container/Containerfile     |  36 -----------
   prototype/container/build-host        |  34 ----------
   prototype/container/ctr               |  20 ------
   prototype/container/d28-hits          |  21 -------
   prototype/container/d28-one           |  35 ----------
   prototype/container/grade             |  62 ------------------
   prototype/container/results.md        | 115 ---------------------------------
  ```

  Both anchor commands → rc0 (pipeline with pipefail). Subject above abbreviated for display; exact file bytes are the two matching anchors. No behavioral red required: text/state defect.
- **Acceptance:** while IMPLEMENT remains open, tick this row with `57903b5`; if the same close finishes the phase, clear completed rows under Session flow and retain only actual open work. Recheck spec shape + matching state; no history rewrite needed.

## Register

### R-M1 — in-flight verification/ledger close; not a second finding

0d14378's body says `internal/ctr unit tests: package absent at 5c11b58.` That sentence does not supply mutation credit for tests never observed red. `.agent/spec.md:45` explicitly leaves `Closing review per lens + security (MAIN) + mutation credit` open. The old ledger's R100 archive move preserves the completed prior phase; MAIN still owns the new `.agent/review.md` + survivor rulings. Treat these as phase-close prerequisites, not completed checks. Verification lens reviewer-4 owns detailed adjudication; this review does not award missing credit or infer a false completion claim.

## Checks + limits

- Read all target-added source/test files, changed source hunks with enclosing functions, current spec/contract/deferred, all 4 rule files, README/help, installer, retired-state archives + all 3 commit bodies. Reviewed prototype retirement by complete deleted-path list + tag tree (all 7); did not re-audit retired prototype behavior.
- Spec shape check: `sh /home/eturkes/Projects/offload/.agent/spec-lint /home/eturkes/Projects/offload/.agent/spec.md` → no diagnostics, **rc0**. This check does not detect M-F1's semantic checkpoint error.
- Artifact negative query: `git ls-files '.scratch/**' '.tools/**' '*.log' '*.tmp' '*.out' '*.png' '*.new' '*.test'` → 0 paths, **rc0**; positive control `git ls-files README.md` → expected 1 path, **rc0**. Full target path inventory also inspected. No claim about unrelated ignored scratch contents.
- Ignore probes: `.scratch/review/reviewer-5.md`, `.tools/mutago`, `CLAUDE.local.md` → all 3 ignored, **rc0**.
- LSP real-file call (`documentSymbol`, `internal/ctr/ctr.go`) → **No LSP server available for file type: .go**. Used source/caller searches; no language-server all-clear claimed.
- MAIN concurrently edited `internal/ctr/{ctr.go,ctr_test.go}`, `internal/slots/{slots.go,slots_test.go}`, `test/e2e/ctr_test.go`. Those post-target changes are excluded; original committed hunks/initial reads establish this review's boundary. Spec finding rematched against 57903b5. No target edits by this reviewer.
- **Not run:** suites, full/fast gate, mutation, builds, real Podman, remote access, deployments, GitHub CI. Commit-measured runtime/CI/deploy outcomes remain **unconfirmed independently here**; inspected the three commit bodies + archived consultant report only. Frontmatter inspected, not parsed by a YAML tool.
- Validator positive control: interim report → expected 5 pending-marker matches, output5, **rc0**. Final validator → expected0, output0, **rc1** (grep no-match status); adjudicated-row query → expected7, output7, **rc0**.
