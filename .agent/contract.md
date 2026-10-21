# offload contract — IMPLEMENT units

Predicates = testable; ids (`C3`, `S7`…) are cited by tests + review rows. Tier per unit. `Decisions` (spec) outrank this file; a conflict = a defect here.

Scope: one controller (this machine) drives the pool; pool hosts are distinct machines; the source tree stays quiescent during an upload (offload takes no snapshot).

## CLI — tier kernel

- C1 `offload run [--local|--pool] [--host NAME] [--gpu] [--no-sync] [--get PATH]… [--] CMD [ARG…]`; `offload status`; `-h`/`--help` on both. Flags parse up to `--` or the first non-flag word; everything after belongs to CMD.
- C2 `--local` with any of `--pool`, `--host`, `--gpu`, `--no-sync`, `--get` → usage error. `--host`, `--gpu`, `--no-sync`, `--get` imply `--pool`.
- C3 Exit status precedence: offload got HUP/INT/TERM → 129/130/143; else an offload failure → 125 (sync up/down failed, remote setup failed, worktree or unsupported repo layout under `--pool`, every host up and none with a GPU under `--gpu`); else the command's status unchanged (any value, 2/125/255 included; 128+n when signal n killed it; 127 = not found, 126 = not executable). Usage errors, an invalid pool file, an unknown `--host` → 2 before any work.
- C4 stderr carries every offload message, prefixed `offload: `; stdout carries the command's output alone (`status` prints to stdout).
- C5 Placement line on stderr before the command starts: `offload: local slot <i>/<n>` | `offload: <host> slot <i>/<n> (<cpus> CPUs, <busy>% busy) → ~/<workdir>` (busy = 100 × load/nproc, rounded).

## Pool file — tier data

- P1 Path = `$OFFLOAD_CONFIG`, else `$XDG_CONFIG_HOME/offload/pool.toml`, else `~/.config/offload/pool.toml` (empty env values count as unset).
- P2 `[[host]]` tables: `name` (required, `[A-Za-z0-9._-]+`, not `.`/`..`, unique), `address` (required, `[user@]host` or an ssh_config alias: no whitespace, `/`, `:`, leading `-`; unique), `slots` (int ≥ 1, default 2). Unknown keys → error naming them. A present file is validated on every `run` (bad file → rc 2) except under `--local`, which never reads it.
- P3 Missing file: `run` → local slots only + one stderr note; `run --pool` + `status` → rc 2 naming the path. A file without hosts = an invalid file.

## Probe + status — tier data

- R1 One ssh round trip per host (shared SSH options, X5), all hosts in parallel, ≤ 8 s end to end, killed + reaped at the deadline: nproc (int ≥ 1), 1-min loadavg (finite ≥ 0), MemAvailable kB (int ≥ 0), NVIDIA GPU count (int ≥ 0). Nonzero ssh exit, timeout, or output ≠ exactly those 4 fields → host `down` with the reason; never a crash.
- R2 `status`: local heavy slots (held/total + `heavy --status`-style holder lines) then one row per host in pool-file order: state, slots held/total, CPUs, load, busy %, free memory, GPUs; a down host shows its reason. A slot counts as held by its lock alone (stale owner text ignored). rc 0 with hosts down.

## Slots — tier kernel

- S1 Local slots = heavy's: dir `${XDG_RUNTIME_DIR:-/tmp}/heavy`, files `slot1…slotN`, N = `HEAVY_SLOTS` (unset/empty = 2) accepted + rejected exactly as heavy does (differential test vs a copy of heavy).
- S2 Acquire = open `slot<i>` (create, never unlink) + flock(2) `LOCK_EX|LOCK_NB`, i ascending; first success wins. heavy (and `flock -n`) and offload exclude each other on the same file. Contention ≠ I/O error (the latter → 125).
- S3 While held, `slot<i>.owner` = `<pid> since <HH:MM:SS>:` + ` <word>` per CMD word, shell-quoted; emptied before unlock. The command never inherits a lock fd.
- S4 Pool slots = same mechanism under `${XDG_RUNTIME_DIR:-/tmp}/offload/hosts/<name>/slot<i>`, i = 1…`slots`; owner text = ` <host>: <root> <CMD…>`. Controller workdir lock `…/hosts/<name>/wd-<workdir name>.lock` taken non-blocking; busy → host unavailable this round (its slot released, never held while waiting).
- S5 A local run's heavy slot stays held until the watcher (X1) finishes tearing down the command, even if offload dies first. A remote run's controller locks may free early on offload death; workdir safety then rests on the remote lock (Y11).

## Placement loop — tier kernel

- L1 Eligibility: local unless pool-only (`--pool` or implied); pool unless `--local`, missing pool file, or remote-ineligible tree (Y3).
- L2 Round: try a local slot; else among eligible hosts (`--host` = that one) with a free slot + free workdir lock, probe (≤ 8 s), keep up hosts (with ≥ 1 GPU under `--gpu`), pick the lowest load/nproc (ties → pool-file order), acquire its slot non-blocking (lost race → next round).
- L3 Nothing acquired → one `offload: waiting …` line naming held/total per place; next round 2 s after the previous ends; a down host is re-probed at most every 30 s. INT/TERM/HUP while waiting → 130/143/129, nothing held.
- L4 `--gpu` with every eligible host up and none showing a GPU → 125. Hosts down or busy → keep waiting.

## Sync — tier kernel

- Y1 Sync root = `git rev-parse --show-toplevel` when cwd is inside a git work tree, else cwd ("plain"). A git error other than "not a git repository" → 125, never plain mode; "not a git repository" with a `.git` entry in cwd or any ancestor (a damaged repo) → 125 too. CMD runs in `<workdir>/<cwd relative to root>`. A root that is `/` or the home directory is local-only (`--pool` → 125); every transfer refuses a root that is not absolute or is `/`.
- Y2 Workdir = `offload/<basename(root), chars outside [A-Za-z0-9._-] → _>-<first 8 hex of sha1(abs root)>` under the remote `$HOME`; meta dir = `offload/.meta/<workdir name>/`.
- Y3 Remote-eligible git layouts: `.git` is a directory at the root, git dir = common dir, no `objects/info/alternates`, no gitlink in the index, no untracked nested repository. Anything else (linked worktree, external git dir, submodules, nested repo) → local slots only; under `--pool` → 125 naming the reason. Checked before any remote work.
- Y4 Git upload: U = `git ls-files -z -co --exclude-standard` at root, entries that exist (lstat; a non-ENOENT lstat error → 125) below real-directory ancestors (a path reached through a symlink is not in the tree), + every self-ignored `.gitignore` in a non-ignored directory. rsync sends U (non-recursive: files + symlinks) and mirrors `.git` (recursive, `--delete` inside `.git`, `*.lock` never sent). Every tracked file crosses whatever the ignore rules say; every ignored untracked file stays local.
- Y5 Git remote prune (before CMD, under the remote lock): R = remote `git ls-files -z -co --exclude-standard` with the local global excludes file supplied; delete each path in R − U whose every ancestor below the workdir is a real directory (no symlink) and which is itself a file or symlink (never a directory; never recursive). U reaches the remote framed with its record count + a run nonce; a count/nonce mismatch, a git failure or a missing `.git` → 125 before any prune deletion (the `.git` mirror may already have changed).
- Y6 Git download (after CMD, any status): D = remote `git ls-files -z -co --exclude-standard`, filtered to existing paths below real-directory ancestors; rsync D → local with `--update` (a regular file newer locally stays; symlinks + type changes take the remote version), no deletes. `.git` and ignored untracked files never return.
- Y7 Plain mode: upload = mirror with `--delete` + rsync `--filter=':- .gitignore'` (rsync's dir-merge semantics, not git's); download = `--update` + same filter.
- Y8 `--get PATH` (repeatable; relative to the invocation cwd; must stay inside the sync root after normalization; `.git` + absolute paths refused → rc 2): after the automatic download, PATH copied remote → local with `--update` whatever its ignore status; trailing `/` = directory contents. Missing remote PATH → 125.
- Y9 `--no-sync`: no upload, prune or automatic download; workdir created if absent; `--get` still runs.
- Y10 Upload/prune/download/get failure → 125 with the tool's message; CMD never starts after a failed upload or prune. Interrupted runs (C3 signals) skip every download.
- Y11 Remote lock: every remote phase of a run (upload rsync via `--rsync-path`, the run script, download rsyncs) holds flock on `<meta dir>/lock`, waiting ≤ 300 s (then 125); the run script keeps it until X1's teardown ends ⇒ a successor never syncs into a workdir with a live predecessor.

## Execution — tier kernel

- X1 CMD runs under a watcher in a new session (setsid), stdin `/dev/null`, default INT/QUIT dispositions. The watcher holds offload's pipe; at EOF, or when CMD's leader exits, it sends TERM to every process in that session (`pkill -s`), KILL after 10 s, and exits once none remain. A process that starts its own session escapes (documented limit).
- X2 Local: argv executed as given (no shell), cwd unchanged, env inherited; a missing command → 127 before any slot is taken. Remote: CMD words shell-quoted into `bash -lc` (login env ⇒ host PATH).
- X3 Offload on HUP/INT/TERM: close the watcher pipe, wait for the watcher to finish, skip downloads, release locks, exit 129/130/143.
- X4 Output streams live, byte-exact (no buffering or rewriting by offload).
- X5 Every ssh + rsync connection: BatchMode=yes, StrictHostKeyChecking=accept-new, ConnectTimeout=5, ServerAliveInterval=15, ServerAliveCountMax=4. Pool hosts set sshd ClientAliveInterval 15 × 4 ⇒ a vanished controller's session ends within ~1 min (X1 teardown follows).
- X6 Remote status vs transport: the run script records `<nonce> <status>` in the meta dir; ssh exit 255 → offload reads that record (one extra ssh): matching nonce → the command's 255; else transport failure → 125.
- X7 Requirements: controller = ssh, rsync, git, bash, util-linux (`setsid`), procps (`pkill`/`pgrep`); pool host = bash, rsync, git, GNU coreutils (`sort -z`, `comm -z`, `head -z`), util-linux (`setsid`, `flock`), procps.
