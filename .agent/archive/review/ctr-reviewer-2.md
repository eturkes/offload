# reviewer-2 — lens B: claim soundness (every claim in the diff ef515f2..2c2744d + agents-repo commits b78f436, f40dfde)

| id | check | verdict (pass / finding) | finding ids | evidence (file:line, command + output) |
|----|-------|---------------------------|-------------|----------------------------------------|
| B1 | README `Run a command in the project image` (all subsections) vs code + tests | finding | B-F4, B-F5, B-F6 | README:161 native-equivalence fails on PATH nonexec (126 vs127); :201 host-env exclusion fails for proxies (MAIN witness); :174 content-change invalidation fails for accepted directory symlink (reviewer-1 red). Other subsections inspected against implementation/tests. |
| B2 | README Requirements + Install (+ Develop line) vs tools/install + code | finding | B-F1, B-F2 | README:39,43; tools/install:17; contract K6. Extracted target parser: valid single-quoted TOML selects 0/1 hosts, rc0; witness rc1. Atomic-install wording contradicts documented pre-launch inode rejection. Requirements/static build/Develop skip wording otherwise match. |
| B3 | contract K1–K11 + C3 scope + Scope line: each sentence true of the code | finding | B-F3, B-F4, B-F6; F-A2 (reviewer-1) | K1 help precedence overstates parsing; K2 PATH nonexec status127≠126 (both local red); K4 context symlink ignores file changes; K7/K8 missing record hangs (reviewer-1 reds inspected). Remaining predicates source/test/evidence checked; limits below. |
| B4 | .agent/spec.md Artifacts, Decisions, Tasks vs repo state | finding | B-F1 | Artifacts:7 pool-wide installer wording inherits B-F1. ctr artifacts/Decisions match code + recorded deployment; tag prototype-pool-run-environment=bbaf2c2 retains retired scripts, absent from 2c2744d tree. README task remains open pending closing bookkeeping; review/mutation correctly open. |
| B5 | code comments (why-claims) in internal/ctr/*, internal/cli/ctr.go, internal/slots WaitLock, internal/mirror Locate/Name, tools/install, configuration.nix | finding | B-F1, B-F6; F-A2 (reviewer-1) | Installer every-address claim fails; Tag comment's every-entry promise fails for root symlink. Reviewer-1 F-A2 also disproves readRecord's bounded-missing-record comment (cross-reference below). Other why-comments inspected with enclosing functions/callees. |
| B6 | commit bodies 2efc60d, bbaf2c2, 2c2744d: each measurement/claim checkable from repo, logs (.scratch/logs/*) or a rerun | finding | B-F1 | bbaf2c2 subject's every-pool-file-host claim inherits B-F1. Recorded gates: 13/13, test85.7/87.9/82.9s; cement1375 tests, unittest255.4/447.8s, wall279/474s; --check same5 layers/8.830s/tag0c9c…; redbase-2=32 FAIL/0 PASS/0 SKIP. Remaining historical checks explicitly unconfirmed below. |
| B7 | .claude/rules/code.md new bullets + machines/nixbox/README.md new bullet | pass | — | code.md:12–13 agree with init.go, e2e requirePodman/removeStorage/TestMain/ctrhelper + gate.yml OFFLOAD_REQUIRE_PODMAN=1. Machine README:27 agrees with configuration.nix + install/deploy evidence (remote state not re-probed). No claim that whole K-domain tests are exhaustive. |
| B8 | agents repo: ~/.local/app/agents/claude/CLAUDE.project.md CI gate + Git + CLAUDE.local.md lines; claude/cachyos/CLAUDE.md pool + Project images lines (git -C ~/.local/app/agents show b78f436 f40dfde) | pass | — | Template:18,20,34 + cachyos:10–11 implement spec's conditional CI adoption, ci/* permission scope, local instruction tier, missing-tool image routing, ctr/check forms. Local source/live switch jq=1/1; deployed instruction cmp rc0. Historical CI timing qualifications = R3. |

## Findings (keyed by id)

### B-F1 — major — valid pool syntax silently omits installation targets

- Anchors: `/home/eturkes/Projects/offload/README.md:39` → `on each host of the pool file`; `/home/eturkes/Projects/offload/tools/install:3` → `default: every \`address\` in`; parser `/home/eturkes/Projects/offload/tools/install:17` → `set -- $(sed`.
- Wrong: P2 accepts TOML `[[host]]` tables, including literal strings. Installer's regex accepts double-quoted values alone. A valid single-quoted address silently becomes zero SSH destinations; installation can report success after updating only the controller, contrary to the documented pool-wide install.
- Red witness on exact target revision; no installer execution, target writes, or SSH:

```sh
python -I - <<'PY'
import subprocess, tomllib
repo='/home/eturkes/Projects/offload'
text=subprocess.check_output(['git','-C',repo,'show','2c2744d:tools/install'],text=True)
line=next(x.strip() for x in text.splitlines() if x.strip().startswith('set -- $('))
for label,value in [('positive-double','"agent@alpha"'),('valid-single',"'agent@alpha'")]:
    config='[[host]]\nname = "alpha"\naddress = '+value+'\n'
    expected=tomllib.loads(config)['host'][0]['address']
    run=subprocess.run(['sh','-c','pool=/dev/stdin; '+line+'; printf "%s\\n" "$#"; printf "%s\\n" "$@"'],input=config,text=True,capture_output=True)
    fields=run.stdout.splitlines()
    print(f'{label}: TOML address={expected!r}; installer parser count={fields[0]}; parser rc={run.returncode}')
    if fields[0]!='1' or fields[1:]!=[expected]:
        print('FAIL: tools/install must enumerate the one valid pool host; got none')
        raise SystemExit(1)
PY
```

Output (rc1):

```text
positive-double: TOML address='agent@alpha'; installer parser count=1; parser rc=0
valid-single: TOML address='agent@alpha'; installer parser count=0; parser rc=0
FAIL: tools/install must enumerate the one valid pool host; got none
```

- Acceptance: enumerate the same host addresses P1/P2 load, or explicitly narrow/validate the installer input before changing anything. Test equivalent double-/single-quoted and quoted-key forms; success must not silently omit a valid configured host.

### B-F2 — minor — atomic-install guarantee excludes the documented replacement rejection

- Exact anchor: `/home/eturkes/Projects/offload/README.md:43` → `runs that are in progress are not affected`.
- Contradiction: `/home/eturkes/Projects/offload/.agent/contract.md:73` K6 promises `replaced → 125`; `/home/eturkes/Projects/offload/internal/ctr/ctr.go:295` says `the offload binary was replaced while it ran; run the command again`. Run resolves the binary before `Ensure` and checks its inode after `Ensure`; installation during an image build therefore deliberately aborts that already-in-progress invocation. Atomic rename protects processes already executing, not every stage of an offload invocation.
- Proof type = exact text/implementation contradiction; no installation or replacement was performed. Acceptance: qualify the README guarantee to already-started workloads and describe retryable rc125 during setup/build.

### B-F3 — minor — K1 help precedence exceeds the parser

- Exact anchor: `/home/eturkes/Projects/offload/.agent/contract.md:68` → `help wins over any other problem`.
- Wrong: `parseCtr` returns the first `flag.FlagSet.Parse` error. An invalid flag before `--help` returns usage error2; help is never reached. The current tests cover bad environment, not earlier bad flags.
- Red witness: `OFFLOAD_CTR=bogus /home/eturkes/.local/bin/offload ctr --help` → rc0, `Usage:` on stdout (positive control). Same prefix + `ctr --bogus --help` → rc2, `offload: flag provided but not defined: -bogus` on stderr. Python assertion `actual.returncode == 0` failed, witness rc1.
- Artifact identity: installed SHA256 `5ab12c60ebc65a5c979276394a0b4e27f9f63c642aca11ff5e28f80eb7027e53` = bbaf2c2's recorded install; `go version -m` = 2efc60d+dirty, CGO_ENABLED=0. Target parser source agrees; bbaf2c2/2c2744d do not change that parser.
- Acceptance: narrow the claim to successfully parsed help beating setup/switch problems, or implement + test the broader stated precedence without consuming CMD's flags.

### B-F4 — major — native PATH lookup reports a non-executable as missing

- Exact anchors: `/home/eturkes/Projects/offload/.agent/contract.md:69` → `not executable or exec refused (ENOEXEC included) → 126`; `/home/eturkes/Projects/offload/README.md:161` → `as if you typed it without \`offload ctr\``.
- Wrong: `execNative` calls `checkLocalCommand`; `exec.LookPath` suppresses a non-executable PATH candidate into `exec.ErrNotFound`, which maps to127. CMD exists, but the diagnostic says it does not. Explicit `./not-executable` is tested; PATH-resolved non-executable is not.
- Red command (installed artifact identity = B-F3; no writes):

```sh
python -I - <<'PY'
import os, subprocess
repo='/home/eturkes/Projects/offload'
env={**os.environ,'PATH':repo,'OFFLOAD_CTR':'0'}
ref=subprocess.run(['/bin/bash','-c','exec README.md'],env=env,text=True,capture_output=True)
actual=subprocess.run(['/home/eturkes/.local/bin/offload','ctr','--','README.md'],env=env,text=True,capture_output=True)
print(f'native reference rc={ref.returncode}; stderr={ref.stderr.strip()!r}')
print(f'offload ctr rc={actual.returncode}; stderr={actual.stderr.strip()!r}')
assert ref.returncode==126
if actual.returncode!=126:
    print('FAIL: K2 not-executable PATH command must return 126')
    raise SystemExit(1)
PY
```

Output (rc1):

```text
native reference rc=126; stderr='/bin/bash: line 1: /home/eturkes/Projects/offload/README.md: Permission denied'
offload ctr rc=127; stderr='offload: README.md: command not found'
FAIL: K2 not-executable PATH command must return 126
```

- Additional reported red witness (reviewer-1 F-A6; not independently rerun here): `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-1/internal/cli/reviewer1_path_test.go` `TestReviewA2NonExecutableOnPath` → `K2: existing nonexecutable PATH entry returned 127, want 126`; `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-1/internal/ctr/reviewer1_path_test.go` `TestReviewA7NonExecutableOnPath` → `pid=0 status=127, want 0/126`. Command: `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-1 && go test -count=1 -run '^TestReviewA[27]NonExecutableOnPath$' ./internal/cli ./internal/ctr)` → rc1. This directly exercises native + init lookup; image process not separately launched for this witness.
- Acceptance: preserve the missing/non-executable distinction over PATH candidates; test both native K2 and image K7.

### B-F5 — minor — README excludes host environment that Podman passes through

- Exact anchor: `/home/eturkes/Projects/offload/README.md:201` → `Variables from the host do not go into the container.`
- Wrong: target RunArgs inherits Podman's default `--http-proxy=true`; ordinary host proxy variables are passed to CMD. A project cannot treat its image ENV as the whole source of those values.
- Independently checked command: `podman run --help` → `--http-proxy Set proxy environment variables in the container based on the host proxy vars (default true)`, rc0; target `internal/ctr/ctr.go:173–184` passes no false override.
- Behavior witness supplied by MAIN (not independently rerun here; brief boundary): installed binary SHA256 `5ab12c60ebc65a5c979276394a0b4e27f9f63c642aca11ff5e28f80eb7027e53`, same2efc60d code carried through target2c2744d. Command: `(cd /home/eturkes/Projects/offload/.scratch/cement && http_proxy=http://proxy.invalid:3128 HTTPS_PROXY=http://p.invalid:1 OFFLOAD_CTR=1 offload ctr -- sh -c 'env | grep -i proxy; echo end')`. stdout: `HTTPS_PROXY=http://p.invalid:1` / `http_proxy=http://proxy.invalid:3128` / `end`; rc0. Both host values survive into CMD, disproving the exclusion claim.
- Acceptance: document Podman's automatic proxy passthrough as an exception, or deliberately disable it and test that behavior. An unchanged blanket exclusion is false.

### B-F6 — major — accepted symlinked context defeats content-change invalidation

- Exact claim anchors: `/home/eturkes/Projects/offload/README.md:174` → `When you change a file in the directory, the next run builds a new image.`; `/home/eturkes/Projects/offload/.agent/contract.md:71` → `a content, name, permission or link change → a new tag`.
- Target code anchors: `/home/eturkes/Projects/offload/internal/ctr/ctr.go:76,79` → `os.Lstat(p.Definition)` / `p.Context = filepath.Dir(p.Definition)`; `/home/eturkes/Projects/offload/internal/ctr/tag.go:25,29` → `filepath.WalkDir(dir` / `if path == dir`.
- Wrong: Find accepts a regular Containerfile reached through a symlinked `.devcontainer` directory. WalkDir does not traverse that root symlink; the root callback returns without hashing it. Distinct accepted context contents all receive empty-tree tag `localhost/offload-ctr:e3b0c44298fc1c14`; changed definitions can reuse stale images.
- Reported red witness from reviewer-1, fixture inspected here: `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-1/internal/ctr/reviewer1_test.go`, `TestReviewA4SymlinkContextChangesTag`. The test creates a git root, a regular `image/Containerfile`, `.devcontainer -> image`, calls Find/Tag, changes Containerfile ENV bytes, and requires a different tag. Worktree HEAD independently checked =2c2744d; ctr.go/tag.go diff to target empty.
- Reviewer-1 command: `(cd /home/eturkes/Projects/offload/.scratch/worktrees/reviewer-1 && TMPDIR="$PWD/.tmp" GOCACHE="$PWD/.cache/go-build" go test -count=1 -run '^TestReviewA(4|7)' ./internal/ctr)` → rc1; failure: `K4: changed Containerfile below accepted symlink context kept tag localhost/offload-ctr:e3b0c44298fc1c14`. Not independently rerun here; duplicate root cause = reviewer-1 F-A1.
- Acceptance: every accepted context hashes its actual entries; ordinary tree and directory-symlink view share a tag, a changed definition changes that tag (or reject the symlinked-context input explicitly).

### F-A2 (reviewer-1) — major — missing-record deadline comment is false (cross-lens duplicate)

- Exact target anchor `/home/eturkes/Projects/offload/internal/ctr/ctr.go:301` → `the deadline only bounds a missing one`; K7/K8 require125 + container teardown after Podman dies.
- Reviewer-1's inspected test `/home/eturkes/Projects/offload/.scratch/worktrees/reviewer-1/internal/ctr/reviewer1_test.go` leaves the socket peer open. Target socket is blocking; ignored `SetReadDeadline` fails; `Read` blocks past1.5s. Reported command = B-F6's combined `^TestReviewA(4|7)` command, rc1; failing output `K7/K8: missing status blocked past its 1 s deadline while peer remained open`.
- Full behavior finding + live Podman witness owned by `/home/eturkes/Projects/offload/.scratch/review/reviewer-1.md` F-A2. No extra root cause or scope expansion here; source assertion belongs to B5, lifecycle contract to B3. Independent rerun not done by this reviewer.

## Scope + method

- Check set frozen before target read: B1–B8 above. Read-only targets; this validator input = sole write.
- Target: offload `ef515f2..2c2744d`; agents commits `b78f436`, `f40dfde`.
- Evidence: exact text anchors + enclosing implementation/tests; measurements checked against recorded logs or permitted cheap local reruns. No remote access, deployment, installer run, full-suite loop, or container/image removal.
- Target drift: MAIN modified ctr/slots/installer during review. Original reads matched target; later changed areas checked through `git show 2c2744d:<path>`. Final cmp: README/contract/spec/code-rule/machine-README + both agents instruction sources equal target; installer differs (concurrent repair, not re-reviewed). Findings concern the fixed target, not unreviewed concurrent repairs.
- Final validator:0 pending-marker matches (grep rc1, expected),8/8 adjudicated rows,6 own finding sections +1 cross-lens duplicate. Positive seed control:8 matches, rc0. B-F5 exact MAIN command/output/rc attached; no pending evidence attachment.
- Validator positive control: seeded table = 8 pending-state lines; completion requires 0 matching lines (grep count exits 1 when count = 0).

## Evidence + material coverage limits

- Target text read: README whole; contract/spec/rules whole; every new ctr/cli function + tests/helper; mirror Locate/Name + slots WaitLock with callers/callees; installer; changed machine/workflow hunks; all three offload commit bodies; both agents commits. Existing unrelated spec decisions were contextual inputs, not newly re-certified claims.
- Historical measurements confirmed from `/home/eturkes/Projects/offload/.scratch/logs/gate-{2,4,5}.log`, `cement-{local,nixbox}.log`, `check-1.log`: counts/timings above; preceding failed gates1/3 correspond to the commit-described test fixes. `redbase-2.log` counts32 FAIL,0 PASS,0 SKIP; older `redbase.log` contains the skip bug, so that initial concern is retired rather than a finding.
- Install/local requirements confirmed: local sha256 = recorded bbaf2c2 value; `go version -m` reports CGO_ENABLED=0, linux/amd64, Go1.27.3; `podman version --format '{{.Client.Version}}'` =6.1.4. Remote checksum/live PATH/OFFLOAD_CTR/trading-unit state = **unconfirmed independently**; looked in deploy-1/check-1 + commit record, no remote calls allowed. deploy-1 establishes activation output, not every post-deploy smoke assertion.
- Historical exact claims **unconfirmed independently**: controller race+shuffle124s; 199/211/239ms slot-barrier samples; 0.09–0.18s spike range; ≤0.5s container-death bound; 20/20 isolated nested-test rerun. Looked in named `.scratch/logs` + archived consultant report + tests; source contains the probes but located logs do not preserve all those outputs. Three timing samples are not a proof of a hard scheduling bound; consultant's accepted XC1 ruling not reopened without a reproduced counterexample.
- Offload-based d28-hits1/73 on both hosts + identical sys.version = **unconfirmed independently as an IMPLEMENT rerun**. hits.log records prototype1/73; close-hits.log records1/74; bbaf2c2 records the later offload run. These establish compatible interpreter/results but not a separately located raw log of that exact later command. No full D28 claim made; deferral explicitly retains it.
- Agents `check-portable rc0` + historical copy cmp = **unconfirmed as historical executions**; target commit records them, no raw log located. Present instruction source/live cmp rc0 + switch1/1 verified. CI34-job79s vs~11min measurement exists in tagged prototype results (R3 qualification); GitHub run not independently fetched.
- Search limits: scratch evidence inventory capped at160 lines; cross-log/archive pattern searches capped100, later focused named-log reads/counts uncapped. Broad hits included other reviewers' worktrees + cement clone; those incidental paths were not used as independent evidence. Initial combined diff tool output was49KB/truncated; target surfaces subsequently read individually + agents diffs separately. LSP call on ctr.go returned `No LSP server available for file type: .go`; source navigation used exact searches/reads.
- Executed here: three red witnesses (installer parser, help precedence, native PATH status); read-only Podman help/version; hashes/buildinfo; local instruction cmp/jq; git/text/log consistency checks. Not run here: full/fast gate, Go suite, container lifecycle tests, installer, deployment, remote probes, CI, mutation suite. Test-source coverage is not reported as a fresh green execution.

## Register

### R3 — CI timing comparison includes different D28 cache states

Agents b78f436 commit body records34-job79s vs~11min one-job. Tagged `/home/eturkes/Projects/offload/prototype/container/results.md` (at prototype-pool-run-environment; now retired) qualifies them: fan-out D28 `74 revisions, 0 replayed`; one-job `2 D28 replays`. Numeric observations agree; do not reuse this ratio as a controlled speedup from parallelism alone. Instruction lines themselves promise no measured ratio, so B8 passes.
