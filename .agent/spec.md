## Intent

A simple CLI for personal use to offload heavy compute work, at a moment's notice, to any available machine within a designated pool of machines accessible within the local network via SSH. Whenever a pool machine is idle, it works on an indefinite task that I designate for it.

## Artifacts

- `cmd/offload` = the CLI (Go; `internal/{cli,mirror,run,slots,pool,probe,remote}`); contract `.agent/contract.md`. Install: `CGO_ENABLED=0 go build -trimpath -o ~/.local/bin/offload.new ./cmd/offload && mv ~/.local/bin/offload.new ~/.local/bin/offload` (installed). Run: `offload status` | `offload run [--local|--pool] [--host H] [--gpu] [--no-sync] [--get PATH]… -- <cmd…>`. Pool = `~/.config/offload/pool.toml` (`OFFLOAD_CONFIG` overrides); example `pool.example.toml`. User docs `README.md`.
- `machines/nixbox/` = NixOS config of pool host `nixbox` (`offload@192.168.40.240`); runbook `machines/nixbox/README.md`. Deploy: `machines/nixbox/deploy`; reinstall: `machines/nixbox/install /dev/nvme0n1` (box on installer ISO from `installer.nix`).

## Decisions

- Stack = Go, one static binary; shells out to `ssh`, `rsync`, `git`. License = Apache-2.0 WITH LLVM-exception (`LICENSE` + `NOTICE`, as cnl-ckc).
- User = agents alone (owner never calls it). Trigger = hybrid `offload run [--local|--pool] -- <cmd>`: the agent opts a command in (= offloadable: project-scoped, no GPU/desktop, timing not recorded); the tool runs it in a free local `heavy` slot, else a free pool slot, else waits on both (one waiting message); stderr names where it ran. `--host`, `--gpu`, `--no-sync` imply `--pool`; `--gpu` = pool hosts with an NVIDIA GPU (local GPU work → `gpu-run`). Timing-recorded runs stay on plain `heavy`.
- `heavy` interop: offload takes heavy's slots itself — `${XDG_RUNTIME_DIR:-/tmp}/heavy/slot<i>`, i = 1…`HEAVY_SLOTS` (default 2), flock(2) exclusive non-blocking, owner line `<pid> since HH:MM:SS: <cmd %q>` in `slot<i>.owner`, emptied at exit ⇒ `heavy --status` lists offload's local runs. `heavy` stays a separate host script.
- Admission = per-host `slots` in the pool file (default 2); slot locks on the controller (`${XDG_RUNTIME_DIR}/offload/hosts/<name>/slot<i>`). Loadavg ranks hosts with a free slot (score = 1-min load / nproc), never admits: an idle task's threads inflate it.
- Remote run = pick host → sync up → `bash -lc` the command in the remote workdir over ssh, live output, rc passed through → sync down; `--get PATH` also fetches an ignored path; `--no-sync` skips both syncs.
- Sync root = git toplevel (a run from a subdir syncs the repo, runs in that subdir remotely); non-git dir = cwd. Remote workdir = `~/offload/<root name>-<sha1(abs root)[:8]>`, persistent ⇒ incremental syncs.
- Git-exact sync: upload set = local `git ls-files -co --exclude-standard` (existing files) + `.git` mirrored (deletions included); after upload the remote deletes its non-ignored files absent from that set (remote git, same ignore sources; local global excludes file shipped along); download set = remote `git ls-files -co --exclude-standard` after the command, `--update`, no deletes, `.git` never returns. Pool hosts need git + GNU coreutils. Non-git dir = rsync mirror with the `.gitignore` dir-merge filter. Tracked files always cross; ignored paths (venvs, toolchains, caches) stay on each side.
- A git-worktree checkout runs in local slots only (`--pool` → error).
- Remote runs from one sync root serialize per (host, workdir) on a controller lock; local runs never lock.
- Killing offload (any signal, incl. SIGKILL) kills the command's session, local or remote: the command runs under a watcher that kills its process group at stdin EOF.
- Pool = TOML `[[host]]` tables (`name`, `address` = `[user@]host` or an ssh_config alias — ports + IPv6 via an alias, `slots`). SSH = BatchMode + accept-new host keys.
- Idle task: each pool host may run one designated indefinite task, declared in `machines/<host>/`, in a systemd slice at idle CPU weight ⇒ offload work preempts it at once; offload code unaffected. nixbox's idle task = trading, user `eturkes`: deployed paused (units + guards installed, nothing enabled, hold kept), Luna set up (pinned Codex); the owner activates + assigns workloads through ChatGPT (Codexify). The intraday trial expired during the migration pause; Luna v1 is source-stopped (a reviewed v2 successor precedes any start).
- Pool hosts = dedicated machines; LAN-only (nftables: port 22 from 192.168.40.0/24 alone), key-only SSH, no passwords on any account. nixbox accounts: `offload` (offload runs + deploys; wheel, passwordless sudo) + `eturkes` (unprivileged; trading, its idle task, Codexify; home `/home/eturkes` ⇒ trading's tracked absolute paths hold).
- Trading lives on nixbox at `/home/eturkes/Projects/trading`, every file incl. key files (owner ruling over trading's never-copy rule); ChatGPT iterates on it through Codexify on nixbox (single project). Controller copy deleted after a checksum-verified copy (confirm immediately before).
- Pool host OS = NixOS, config declarative in `machines/<host>/`; Wi-Fi secret stays on controller + box (gitignored `wifi.nix`).
- nixbox disk = LUKS2 root, TPM2 auto-unlock bound to PCR 7 + PCR 15 = 0 (volume key measured into PCR 15 at unlock ⇒ a swapped header cannot unseal it) + Secure Boot with own keys (lanzaboote) ⇒ unattended reboots; recovery key on the controller only.
- nixbox compute posture = performance governor, thermald, earlyoom, zram, lid/suspend ignored, battery capped 80 %; nix-ld + `UV_PYTHON_PREFERENCE=only-managed` so prebuilt wheels run.

## Tasks

- [ ] MAINTAIN request: fund every `.agent/deferred.md` row in rank order; owner parts → owed on the row; one owner question once agent-side work ends, then fund its answers. Resume = first row without an `Owed` note or closing commit
- Deferred queue → `.agent/deferred.md`

## Phase

MAINTAIN — whole product
