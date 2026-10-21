---
paths:
  - "cmd/**"
  - "internal/**"
  - "test/e2e/**"
---
# offload code

- Law = `.agent/contract.md` (predicate ids C/P/R/S/L/Y/X); `Decisions` outrank it. A behavior change edits the predicate first; tests cite ids in names or comments.
- `internal/run/watch.bash` runs on both sides (local runs + inside `remote.bash` on pool hosts; remote = `<meta>/run-<sha>.bash`, installed by run.go's `boot`, X8) ⇒ bash-only, GNU coreutils/util-linux/procps, no controller paths. Every variable in watch.bash, remote.bash + boot is named `__…`: a bash local shadowing a caller-exported name is exported with the local's value, straight into CMD's environment (X2). `&` jobs in non-interactive bash get stdin `/dev/null` + ignored INT/QUIT ⇒ the explicit `<&0` and `trap - INT QUIT` stay.
- Deletion on a host = `remote.bash` `prune_one` (leaf files/symlinks under real-directory ancestors) + a linked worktree's own `rm -f .git` before its gitfile is written (Y4); rsync `--delete` only ever runs inside `.git` or `<meta>/repo` (a worktree's common dir; plain mode excepted, Y7). A new rsync flag that deletes or replaces outside those = a Y5 change.
- Every ssh/rsync connection takes `remote.SSHOptions`; `test/e2e/testdata/ssh` refuses a call without them (rc 254) and honors `-n`, so the e2e suite checks X5 on every call.
- e2e fake hosts = dirs under `$OFFLOAD_FAKE_HOSTS/<host>` driven by marker files: `.probe` (probe answer), `.transport-fail` (every call 255), `.run-fail` (run-phase call 255; run phase = the boot command, matched on `sha256sum`), `.upload-fail` / `.download-fail` (rsync 23); each call logs `.ssh-call-<pid>`. A host without a dir = down.
- Mutation credit (not a gate step): `TMPDIR=<worktree>/.tmp .agent/tool mutago --workers=4 ./internal/<pkg>/...` in a disposable worktree (it rewrites sources in place; a killed mutant's test leaves its `t.TempDir` tree behind and a broken-root mutant can mirror `/` into it ⇒ TMPDIR on disk, deleted after); one targeted mutant at a time: `.agent/plant <worktree> <name> <file> <old> <new> <test-regex>` (applies, runs, reverts with `git checkout -- <file>` ⇒ commit the change under test inside the disposable worktree first, else the revert restores the old code). A mutant that hangs the suite reads as `rc … without a failing test`: bound the test instead.
