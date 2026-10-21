## Intent

A simple CLI for personal use to offload heavy compute work, at a moment's notice, to any available machine within a designated pool of machines accessible within the local network via SSH.

## Artifacts

- `prototype/offload/offload.py` = PROTOTYPE CLI (Python, stdlib, uv script). Run: `prototype/offload/offload.py status` | `prototype/offload/offload.py run [--host H] [--gpu] [--no-sync] -- <cmd…>`. Pool = `~/.config/offload/pool.toml` (`OFFLOAD_CONFIG` overrides); example `prototype/offload/pool.example.toml`.
- `machines/nixbox/` = NixOS config of pool host `nixbox` (`offload@192.168.40.240`); runbook `machines/nixbox/README.md`. Deploy: `machines/nixbox/deploy`; reinstall: `machines/nixbox/install /dev/nvme0n1` (box on installer ISO from `installer.nix`).

## Decisions

- Prototype stack = Python single file under `uv run --script`, stdlib only; shells out to `ssh` + `rsync`.
- `run` model = pick least-loaded reachable host (score = 1-min loadavg / nproc; `--gpu` filters to NVIDIA hosts) → rsync cwd up (`.git` excluded, `.gitignore` honored) → `bash -lc` the command over ssh, live output, rc passed through → rsync back (`--update`, no delete).
- Remote workdir = `~/offload/<dirname>-<sha1(abs cwd)[:8]>`, persistent across runs ⇒ incremental syncs.
- Pool = TOML `[[host]]` tables (`name`, `address` = any ssh destination). SSH = BatchMode + accept-new host keys.
- Pool hosts = dedicated machines owned by the agent; login = user `offload`, key-only SSH, LAN-only (nftables: port 22 from 192.168.40.0/24 alone), no passwords on any account.
- Pool host OS = NixOS, config declarative in `machines/<host>/`; Wi-Fi secret stays on controller + box (gitignored `wifi.nix`).
- nixbox disk = LUKS2 root, TPM2 auto-unlock bound to PCR 7 + Secure Boot with own keys (lanzaboote) ⇒ unattended reboots; recovery key on the controller only.
- nixbox compute posture = performance governor, thermald, earlyoom, zram, lid/suspend ignored, battery capped 80 %; nix-ld + `UV_PYTHON_PREFERENCE=only-managed` so prebuilt wheels run.

## Tasks

- [x] b989704 Prototype CLI `status` + `run`
- [x] 0738ff6 nixbox: find on LAN, key SSH, install NixOS, static IP, `offload` user, LUKS + TPM + Secure Boot
- [x] b989704 Drive CLI against nixbox
- Deferred queue → `.agent/deferred.md`

## Phase

ITERATE — whole product
