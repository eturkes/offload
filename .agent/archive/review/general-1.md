# general-1 — trading deployment on nixbox

Scope = `eturkes@192.168.40.240:/home/eturkes/Projects/trading`; controller trading copy read-only. Latest disposition = Round 2 below: setup only, globally paused; no activation. Round 1 records the earlier blocked activation attempt and is retained as history.

## Round 1 — activation blocked

Software preparation only at that checkpoint: Codex 0.158.0 + Ruff 0.17.1 under remote `/home/eturkes/.local/`; no units or repository edits yet.

| id | step | status | evidence |
|----|------|--------|----------|
| D1 | owner approval recorded in trading repo + committed | blocked (activation decision unresolved) | Delegated approval quotes received; current immutable intraday deadline expired + Luna source verification fails. No approval receipt fabricated, no repo changes. MAIN notified for exact-scope ruling/confirmation. |
| D2 | global hold removed + committed | blocked (reactivation prerequisites fail) | `.agents/trading-paused` retained. `/home/eturkes/Projects/trading/AGENTS.md:12` requires explicit approval + retained-state/source checks. Intraday bindings pass; Luna bindings fail (D8). |
| D3 | Luna retired → active; reviewed permit + finite stop rule | refused (current policy rejects Luna; frozen sources changed) | `/home/eturkes/Projects/trading/scripts/trading_operations.py:63–64` accepts SCALE/INTRADAY alone. `/home/eturkes/Projects/trading/docs/operations.md:104` requires reviewed permit + source/state verification. No allowlist bypass or policy rewrite. Proposed campaign limits already frozen: 48 trials, 3/day, 5000 requests/snapshots. |
| D4 | nixbox deployment docs + commit | blocked (deployment not established) | Existing CachyOS absence section remains truthful for controller; nixbox activation cannot be documented as complete. Verified remote manager PATH contains `/run/current-system/sw/bin`, `systemctl show user-1001.slice -p CPUWeight --value` → `idle`, `loginctl show-user eturkes -p Linger --value` → `yes`. |
| D5 | pinned Codex installed; archive SHA256 equality | done | nixbox `/home/eturkes/.local/lib/trading-codex/0.158.0/codex --version` → `codex-cli 0.158.0`; archive/install SHA256 = `167c0148a849d2444f1b5a7fb5f8bb2de1de5ae13a2a504b833fc765980f5cd9`; `cmp` binary + release.json → rc0 each. |
| D6 | units + guards/resource drop-ins installed | blocked (reviewed activation prerequisites) | No units/drop-ins installed. `systemctl --user list-unit-files 'bsc-*' 'paper-*'` → 0 files, rc1 (expected empty census); no guarded/disabled deployment claimed. CPU gate requires 11 dormant definitions + original resource overrides; retired `bsc-luna` explicitly says do not install. |
| D7 | intraday running + freshness checked | blocked (original trial expired; indefinite run forbidden) | nixbox `timeout 180 ./bsc-intraday status` → rc0, `binding_status=current`, `runtime_blocked=false`, `paused=true`, `worker_status=stopped`, seq549, fresh=false. All 9 books flat: positions=0, pending=0, unresolved=false. Original final deadline = 2026-10-18T07:31:34.437Z; host clock 2026-10-21. `/home/eturkes/Projects/trading/docs/intraday-paper.md:89,162` forbids indefinite restart/reset. Normal approved `watch` would append an interruption, mark terminal, and exit without market collection (runtime.py:634–650; model.py:1045). Not run. |
| D8 | Luna healthy OR installed-disabled + owner login blocker | blocked (login + frozen source binding) | Pinned Codex installed, service absent/off. Installed `codex login status` → rc1 `Not logged in`. nixbox `./bsc-autonomy verify` → rc2 `Gateway repair source binding changed`; only historical_lab/model.py differs in 18-file binding; Python unchanged 3.13.14. No inference/rebind/repair. See D8 detail. |
| D9 | operations check | blocked (no approved installed deployment) | nixbox `uv run --locked python -m scripts.trading_operations check` → rc2, `Unable to census user units: `; list-unit-files empty/rc1. Expected predeployment failure, not green. |
| D10 | CPU budget check | blocked (units absent) | nixbox `uv run --locked python -m scripts.check_trading_cpu_budget` → rc1, `ValueError: bsc-orderbook.service: unit not loaded`. No quota changed or checker weakened. |
| D11 | docs/operations.md Verification commands | done (executed; installed checks fail as D9–D10) | Exact six-module resource runner → 68 tests/5.353s, OK, resource_warnings=0, rc0. Ruff check + format initially rc2 missing executable; installed user-local Ruff 0.17.1, reran exact commands with `$HOME/.local/bin` on PATH → rc0 each (`All checks passed!`, `2 files already formatted`). Installed checks retained failures above. Full repository suite not rerun (MAIN already performed census); no test skipped/xfail/threshold change. |
| D12 | clean trading repo + scoped commit census | done (unchanged repository; no new commits) | nixbox `git status --porcelain=v1` → 0 bytes, rc0; positive control `git ls-files -- AGENTS.md` → expected 1 line. HEAD remains `f45c9d3df6f01e3f123a80caccb289d2fbbc4120`. No repo edit → no commit due. |

## D8

`./bsc-autonomy verify` output: `{"execution_ready": false, "error": "ValueError", "message": "Gateway repair source binding changed"}` (rc2).

Read-only `autonomous_lab.runtime.binding()` versus `.bsc-autonomy/v1/gateway-repair.json.binding`:
- Python before/current = 3.13.14; file census before/current = 18.
- Sole mismatch = `historical_lab/model.py`:
  - frozen `49753228da4205a2a864d40afa070b9cea69993a0bfa345dd517df35224fd588`
  - current `348e3a1468f320c7f09d4a866f3cdcfd6fcc3a2646d2c827b3901f1868d921a3`
- `git diff --no-index` archive/current → rc1; 67 insertions, 4 deletions. Material matched-horizon controls/benchmark qualification extension, not host-path or Python drift.
- `/home/eturkes/Projects/trading/autonomous_lab/gateway_repair.py:46–57` permits only its five named autonomous_lab source changes; historical_lab/model.py is outside them. Existing repair is not a legitimate bypass.
- `/home/eturkes/Projects/trading/AGENTS.md:59` + `docs/autonomy.md:223–226`: changed sources stop campaign; preserve state/archives; no automatic successor/reset. Future remedy requires separately scoped ruling/review; unconfirmed whether archived-runtime deployment is an approved mechanism (not found in inspected docs).
- Login needed independently: `! ssh -t eturkes@192.168.40.240 /home/eturkes/.local/lib/trading-codex/0.158.0/codex login --device-auth`. Owner performs login. Login alone does not clear source blocker.

## Evidence / coverage

- Read AGENTS.md, operations/resource/intraday/autonomy guides, fixed operations review, complete active/retired policy + systemd template census, controller, CPU checker, archived CachyOS operations receipts. Large runtime JSON inspected through selected fields; source-drift comparison covers all 18 binding entries. No full financial replay claimed.
- Gates log = `/tmp/claude-1000/-home-eturkes-Projects-offload/8af52555-1798-4470-93ee-a9a6dc5d8187/tasks/b0ajl10sz.output` on controller. Shell overall rc0 is last printf only; per-command rc captured above.
- No secret file read/copied/printed. Codex owns login-status lookup. No market/provider calls, account resets, synthetic exits, restart, sudo, offload-user access, NixOS changes, or Git push.
- Notify-set edits = none. Remote repository edits = none. Controller edit = this validator input only.

## Round 2 — owner ruling: Luna set up, nothing activated

| id | step | status | evidence |
|----|------|--------|----------|
| D13 | deployment installed paused on nixbox in the pre-removal CachyOS shape (all 18 templates linked/defined, pause + operational-scope guards, resource overrides; nixbox PATH drop-in for bsc-autonomy), every unit disabled + inactive; file list | done (17 definitions; optional eighteenth refused per law) | Archive = 17 loaded/1 absent; `.agents/operations-retired.json:4` forbids installing `bsc-luna.service`, preserved absent. 18 pause + 18 operational guards; 15 disabled + 2 static definitions, every policy row inactive/PID0, no boot links. All 85 installed files + hashes enumerated in nixbox `/home/eturkes/Projects/trading/docs/operations-setup-nixbox-20261021.json`. Ten selected policy/plan/publication/receipt hashes unchanged; source-bound producers untouched. |
| D14 | `uv run --locked python -m scripts.trading_operations check` on nixbox → result | done | rc0, `operational_scope_ok=true`, issues=[], 18 units quiescent. New retirement receipt `/home/eturkes/Projects/trading/docs/operations-retirement-nixbox-20261021.jsonl`; no enable/start commands. Paused installation health ≠ campaign readiness. |
| D15 | `uv run --locked python -m scripts.check_trading_cpu_budget` on nixbox → result | done | rc0; original 11 units, aggregate300%, period50ms, learner2GiB; independent intraday15%/384MiB/16tasks intact. Receipt `/home/eturkes/Projects/trading/docs/trading-cpu-nixbox-20261021.json`. |
| D16 | docs: nixbox host deployment section (AGENTS.md + docs/operations.md): deployed-paused state, intraday expired during the migration pause, Luna v1 source-stopped (historical_lab/model.py 49753228… → 348e3a14…, commit 5337f74), Luna activation prerequisites (pinned Codex path, owner's `codex login`, a reviewed v2 successor + controller support + permit before any start) — committed | done | Commit `04a1a312bbabfb389954b22e96fdd1b418885f83`: AGENTS.md host rule; docs/operations.md nixbox setup/register/recipe + retained blockers + CachyOS absence; docs/trading-resources.md removes stale ongoing-intraday observation instruction. No activation approval/permit/v2 created. |
| D17 | operations docs/tests still green after the doc change (docs/operations.md Verification set) | done | Committed-state rerun: six-module resource runner 68 tests/5.328s, OK/resource_warnings=0; Ruff check + format, operations check, CPU check, git diff --check all rc0. Precommit unit syntax (autonomy/intraday/learner) rc0; actual hold condition rc1 with positive exists control rc0; service PATH resolves uv + codex0.158.0. No worker/provider live run (explicitly excluded). |
| D18 | trading `git status` clean on nixbox; commits (sha + subject) | done | `04a1a312bbabfb389954b22e96fdd1b418885f83` — `operations: nixbox had no guarded deployment → install paused Luna setup`; 7 scoped files. After committed-state checks, status_bytes=0. No push. |

## Round 2 — delivery detail

Nixbox paths (controller copies remain unchanged):
- `/home/eturkes/Projects/trading/AGENTS.md`: host policy now nixbox/paused; absent optional bsc-luna, fixed trial/source stops, owner-controlled activation prerequisites. **Instruction-file edit: MAIN must name this in final delivery.**
- `/home/eturkes/Projects/trading/docs/operations.md`: paused-host runbook + activation blockers/register; CachyOS absence remains explicit.
- `/home/eturkes/Projects/trading/docs/trading-resources.md`: paused/expired current workload scope; original resource recipe/budgets unchanged.
- `/home/eturkes/Projects/trading/.agents/systemd/bsc-autonomy.service.d/45-nixbox-runtime.conf`: private pinned Codex PATH + `/run/current-system/sw/bin`; managed Python.
- `/home/eturkes/Projects/trading/docs/operations-retirement-nixbox-20261021.jsonl`: original controller output, guards/actions/full before-after snapshots; rc0.
- `/home/eturkes/Projects/trading/docs/trading-cpu-nixbox-20261021.json`: unchanged 11-unit installed CPU check; rc0.
- `/home/eturkes/Projects/trading/docs/operations-setup-nixbox-20261021.json`: complete 85-file inventory/hashes; all18 policy states; ten selected preservation hashes; 13 setup checks with expected failures; fixed original ten-row operations review adjudicated.
- `/home/eturkes/Projects/trading/.bsc-host-archive/nixbox/state/setup-20261021/`: local-only stdout/stderr/rc logs, including committed-state checks.

Installed file census = 17 definitions + 18 `95-trading-pause.conf` + 18 `96-operational-scope.conf` + 7 priority drop-ins + learner override + pinned-Codex override + nixbox override + 22 persistent CPU property files = 85. Configuration roots = `/home/eturkes/.config/systemd/user/` and `/home/eturkes/.config/systemd/user.control/`. All definition bytes equal repository templates. Every trading boot link absent. Two timer-triggered service definitions are static, not enabled.

Retained blockers are not setup failures: global hold intentionally present; Codex not logged in (rc1); intraday original deadline expired; Luna v1 source-stop unchanged; future v2/controller/permit requires owner review. No activation, account/stop/deadline edit, credential access, external provider/market call, or full financial replay. Preservation evidence samples ten selected control/plan/publication/receipt files; it does not claim whole-journal checksum coverage. Retired legacy NixOS executable paths were not run or certified.

Round 2 controller log = `/tmp/claude-1000/-home-eturkes-Projects-offload/8af52555-1798-4470-93ee-a9a6dc5d8187/tasks/ba22kj8r2.output`; committed-state log = `/tmp/claude-1000/-home-eturkes-Projects-offload/8af52555-1798-4470-93ee-a9a6dc5d8187/tasks/bq2ekatm8.output`.

