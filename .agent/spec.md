## Intent

A simple CLI for personal use to offload heavy compute work, at a moment's notice, to any available machine within a designated pool of machines accessible within the local network via SSH.

## Artifacts

- `prototype/offload/offload.py` = PROTOTYPE CLI (Python, stdlib, uv script). Run: `prototype/offload/offload.py status` | `prototype/offload/offload.py run [--host H] [--gpu] [--no-sync] [--get PATH]… -- <cmd…>`; on PATH as `offload` (`~/.local/bin/offload` symlink). Pool = `~/.config/offload/pool.toml` (`OFFLOAD_CONFIG` overrides); example `prototype/offload/pool.example.toml`.
- `machines/nixbox/` = NixOS config of pool host `nixbox` (`offload@192.168.40.240`); runbook `machines/nixbox/README.md`. Deploy: `machines/nixbox/deploy`; reinstall: `machines/nixbox/install /dev/nvme0n1` (box on installer ISO from `installer.nix`).

## Decisions

- Prototype stack = Python single file under `uv run --script`, stdlib only; shells out to `ssh` + `rsync`.
- `run` model = pick least-loaded reachable host (score = 1-min loadavg / nproc; `--gpu` filters to NVIDIA hosts) → rsync cwd up as a mirror (`.git` included, deletions propagate, `.gitignore`d paths never cross) → `bash -lc` the command over ssh, live output, rc passed through → rsync back (`--update`, no delete, `.git` + ignored paths excluded; `--get PATH` fetches an ignored output).
- User = agents alone (owner never calls it). Trigger = `heavy` contention (every local slot held) → work moves to the pool. Prototype form = agent's own call (`heavy --status` → `offload run`), rule in the global agent `CLAUDE.md` beside `heavy`. Final form (agent rule vs tool-decided) = open, decided in IMPLEMENT (`Tasks`).
- Runs from one cwd serialize on a local lock (waiting message); killing `offload` (any signal, incl. SIGKILL) kills the remote command's session.
- Remote workdir = `~/offload/<dirname>-<sha1(abs cwd)[:8]>`, persistent across runs ⇒ incremental syncs.
- Pool = TOML `[[host]]` tables (`name`, `address` = any ssh destination). SSH = BatchMode + accept-new host keys.
- Pool hosts = dedicated machines owned by the agent; login = user `offload`, key-only SSH, LAN-only (nftables: port 22 from 192.168.40.0/24 alone), no passwords on any account.
- Pool host OS = NixOS, config declarative in `machines/<host>/`; Wi-Fi secret stays on controller + box (gitignored `wifi.nix`).
- nixbox disk = LUKS2 root, TPM2 auto-unlock bound to PCR 7 + Secure Boot with own keys (lanzaboote) ⇒ unattended reboots; recovery key on the controller only.
- nixbox compute posture = performance governor, thermald, earlyoom, zram, lid/suspend ignored, battery capped 80 %; nix-ld + `UV_PYTHON_PREFERENCE=only-managed` so prebuilt wheels run.

## Tasks

- [ ] Trigger form, decided first: agent rule (prototype) vs tool-decided vs hybrid `offload heavy -- <cmd>` (agent opts a command in = offloadable; tool runs it in a free local `heavy` slot, else on the pool, else keeps waiting on both; stderr names where it ran; timing-recorded runs stay on plain `heavy`). MAIN's ITERATE recommendation = hybrid: offloadability needs the agent's judgment, contention timing needs the tool (an agent blocked inside `heavy` never sees contention arise).
- [ ] Ship as a real installable command: `~/.local/bin/offload` currently symlinks into `prototype/` → repoint at the IMPLEMENT entry point before the prototype retires; global `CLAUDE.md` rule re-derived from what ships (its nixbox specs + "both slots" = prototype-time facts).
- Deferred queue → `.agent/deferred.md`

## Phase

IMPLEMENT — whole product
