# nixbox

Dell Latitude 7400 (i7-8665U 4C/8T, 62 GiB, 2 TB NVMe, TPM 2.0), Wi-Fi only → `192.168.40.240`. Logins (key-only, controller key): `offload@` = offload runs + deploys (wheel, passwordless sudo); `eturkes@` = trading + Codexify (no sudo). Root SSH off; no account has a password.

## Files

- `host.nix` = shared facts (iface, address, stateVersion, controller key).
- `configuration.nix` = installed system; `network.nix` = NM Wi-Fi profile + static IPv4 (installer + system); `secureboot.nix` = lanzaboote v1.2.0 + fwupd's signed-app bind (`Firmware`).
- `installer.nix` = headless ISO: Wi-Fi + static IP + sshd with controller key; sshd's only host key = `installer_host_key`.
- `installer_host_key` (+ `.pub`) = the ISO's ed25519 host key, gitignored, generated on the controller (`ssh-keygen -q -t ed25519 -N '' -C nixbox-installer -f machines/nixbox/installer_host_key`); `install` pins the `.pub` (`StrictHostKeyChecking=yes`) ⇒ a LAN MITM during a reinstall fails host-key verification before `wifi.env` or the recovery key cross. A new key, or a stick built before the pin (ephemeral key), ⇒ rebuild the ISO first: `install` refuses its key.
- `wifi.env` = `WIFI_SSID=…` + `WIFI_PSK=…` (systemd EnvironmentFile syntax), gitignored, `/etc/nixos/wifi.env` mode 600 root on the box; `network.nix` keeps `$WIFI_*` placeholders in the store, envsubst'd into `/run/NetworkManager/system-connections/lan.nmconnection`. Regenerate from the controller's NM profile `JCOM_hSAf`. Check: `sudo bash -c 'grep -rlF -f <(sed -n s/^WIFI_PSK=//p /etc/nixos/wifi.env) $(nix-store -qR /run/current-system)'` prints nothing.
- `hardware-configuration.nix` = written by `install`.
- `tpm-bind` = re-seal the TPM2 LUKS token from the controller (Disk + boot security).
- `check` = findings an agent must act on (rc 1, one `check: …` line each); `deploy` runs it after every switch. Upgrade: the last `nixos-upgrade` start job in the persistent journal must be `done` and the last success ≤ 72 h old (else the box stops getting the channel's fixes unseen). Units: no failed system unit. Config (`config.reviewed`): effective sshd settings, the nftables input policy + allow rules + no temporary ports, non-loopback listeners (sshd, NetworkManager's DHCPv6), no file of the running closure holding the Wi-Fi PSK, `wifi.env` 600 root. Wrappers: each security wrapper's real program, `strings` matching `^/(usr/)?s?bin/` = the set `wrappers.reviewed` rules on (envfs resolves `/bin` + `/usr/bin` from the caller's PATH with no AT_SECURE check). A finding → rule it there (exec after a privilege drop, only behind authentication `eturkes` cannot pass, or a path that cannot exist) or remove the wrapper. `check --dump` prints the facts; `check --facts FILE` evaluates a saved one.
- `efi-entry` = run by `install`: one `Linux Boot Manager` entry on the new ESP (stale ones from a reused disk deleted). Dry run on the live box: `ssh offload@<ip> 'sudo DRY_RUN=1 sh -s -- /dev/nvme0n1' < machines/nixbox/efi-entry`.
- `performance.nix` = performance posture: `/tmp` tmpfs 50 % of RAM (boot-time).
- `idle.nix` = idle task: user `eturkes` (uid 1001, unprivileged, lingering) owns trading; its `user-1001.slice` runs at `CPUWeight=idle` ⇒ offload runs (user `offload`) preempt it.

## Ops

- Change config → `machines/nixbox/deploy` (rsync → `/etc/nixos`, `nixos-rebuild switch`).
- Installer ISO → `nix-build '<nixpkgs/nixos>' -A config.system.build.isoImage -I nixos-config=machines/nixbox/installer.nix -I nixpkgs=https://channels.nixos.org/nixos-26.05/nixexprs.tar.xz` (controller has no nix → `podman run -v offload-nix:/nix … docker.io/nixos/nix`; `wifi.env` + `installer_host_key` must sit beside `installer.nix`; the ISO's store holds both — keep the image + stick private).
- Reinstall (wipes disk) → stick in, `sudo efibootmgr --bootnext <USB entry>` + reboot on the box, then `machines/nixbox/install /dev/nvme0n1`.
- Recovery key (TPM refuses after BIOS reset/key change) = `~/.local/share/offload/nixbox-recovery-key` on the controller; typed at the console LUKS prompt.

## Toolchains

- Offload runs land here first ⇒ every project gate runs here. `configuration.nix` carries: envfs (`/bin`, `/usr/bin` resolve from the caller's PATH, else the system profile; exec, `test -x` + `stat()` see them, coreutils `ls`/`stat` do not), elan, rustup (default stable via `rustup-default.service`), go (`GOTOOLCHAIN=auto`), node 26 for `offload` (system node 24 = trading's), clang, shellcheck, swipl, bwrap, nix-ld libs for headless Chromium, `offload`'s git identity (`~/.config/git/config` → store), Noto fonts as the controller's sans/serif/mono, tz UTC.
- envfs option change → deploy reloads `usr-bin.mount` but the daemon keeps its old options ⇒ `sudo systemctl stop bin.mount usr-bin.mount && sudo systemctl start usr-bin.mount bin.mount`; leave the old `mount.envfs` lingering in the unit's cgroup (killing it unmounts both).
- `offload ctr` (project images): `OFFLOAD_CTR=1` machine-wide + `~/.local/bin` on PATH (`configuration.nix`); the offload binary there = the controller's, installed by `tools/install` beside `heavy` (`tools/heavy`; project scripts wrap commands in it) — rerun after every offload or heavy change; sha256 printed per file per side. An image-adopting project's missing tool goes into its `.devcontainer/Containerfile`, not here.
- A run fails on a missing tool → add it to `environment.systemPackages` (a library a prebuilt binary loads → `programs.nix-ld.libraries`; a version trading must not see → `users.users.offload.packages`), `machines/nixbox/deploy`, rerun with `offload run --pool`. Missing libs: `ldd <binary> | grep 'not found'` (also lists nix-ld's own set, e.g. glib).

## Power

- `cpu-power-limit.service` = platform profile `performance` + MMIO RAPL PL1 22 W; thermald stays off (its `--adaptive` DPTF policy rewrites PL1 to 9–12 W). Check: `sudo cat /sys/class/powercap/intel-rapl-mmio:0/constraint_0_power_limit_uw` → `22000000`. Sustained 8-thread load, lid shut: ≈ 17 W, 2.6 GHz, 98–99 °C (TCC), fan 7000 RPM.

## Disk + boot security

- Root = LUKS2 `cryptroot`; slots = `tpm2` + `recovery`, no passphrase.
- Secure Boot = own keys (sbctl, `/var/lib/sbctl`) + Microsoft (`--microsoft`, keeps option ROMs + fwupd capsules booting).
- TPM slot bound to PCR 7 (Secure Boot policy) + PCR 15 = 0; the initrd measures the volume key into PCR 15 right after unlock (`tpm2-measure-pcr=yes`) ⇒ a swapped LUKS header unlocked first (attacker passphrase) leaves the real token sealed. Unattended reboots + upgrades unaffected; disabling SB or changing keys ⇒ TPM refuses ⇒ recovery key.
- Re-seal (key or firmware changes) = `machines/nixbox/tpm-bind` (unlocks with the controller's recovery key: after boot PCR 15 ≠ 0, so the TPM token itself never authorizes changes).
- Attacker check: attach a scratch LUKS image with `systemd-cryptsetup attach <name> <loop> <keyfile> tpm2-measure-pcr=yes`, then `sudo env LD_LIBRARY_PATH=/run/current-system/sw/lib/cryptsetup cryptsetup open --test-passphrase --token-only --token-id <tpm2 token> <luks dev>` → refused (`TPM policy does not match`). The plugin path is needed: NixOS cryptsetup does not find systemd's token plugin.
- `/boot/EFI/nixos/kernel-*.efi` shows unsigned in `sbctl verify` by design: the signed lanzaboote stub checks it by hash.

## Firmware

- Installed: BIOS 1.43.0 (Dell blocks downgrades), dbx 20260707.
- fwupd capsules boot under our Secure Boot: lanzaboote's `fwupd-efi.service` signs `fwupdx64.efi` with our db key into `/run/fwupd-efi` (`sbverify --cert /var/lib/sbctl/keys/db/db.pem` → OK), `DisableShimForSecureBoot=true`; fwupd copies it to `\EFI\nixos\fwupdx64.efi` + sets BootNext = `Linux Firmware Updater`. Firmware lacks capsule-on-disk (`OsIndicationsSupported` = 0x03) ⇒ this signed app is the only path.
- fwupd ≥ 2.1 reads the app from fwupd-efi's store dir alone (lanzaboote's `FWUPD_EFIAPPDIR` ignored) ⇒ `secureboot.nix` binds `/run/fwupd-efi` over that dir in fwupd's namespace. Check: `sudo nsenter -t $(systemctl show fwupd -p MainPID --value) -m ls <fwupd-efi>/libexec/fwupd/efi` lists `fwupdx64.efi.signed`; absent → staging fails with `fwupdx64.efi.signed cannot be found`.
- PCR 7: the 1.5.4 → 1.43.0 flash left it + the active PK/KEK/db/dbx byte-identical (only the `*Default` stores changed); the dbx update moved it ⇒ the bound token refused, the open token unlocked. Custom keys survived that flash; a later BIOS may reset them ⇒ owner within reach of the console for BIOS updates.
- Update: hold both controller slot locks (`${XDG_RUNTIME_DIR}/offload/hosts/nixbox/slot{1,2}`, `flock`) for the window — each reboot kills runs on the box + ends trading's live trials (units disabled; restart via Codexify) → `machines/nixbox/tpm-bind --open` → on the box `sudo fwupdmgr update -y --no-reboot-check && sudo reboot` → box answers SSH (BIOS flash ≈ 8 min) → `fwupdmgr get-updates` empty → `machines/nixbox/tpm-bind` (re-seal to the new PCR 7, wipes the open token) → reboot once more, unattended.

## Codexify

- `eturkes` runs Codexify ≥ 2.3 (user unit `codexify.service`, outbound OpenAI tunnel): every chat selects a workspace; only that directory's `AGENTS.md` loads.
- Workspace = `~` for every chat: catalogue entry `.` (name `home`, alias `trading`) in `~/.codexify/codexify.config.json`; `~/AGENTS.md` routes work to `~/Projects/trading`. `worktrees.mode` = `never`: `auto` gives each later chat on `~` a git worktree, and `~` holds no repo.
- Source + deploy = controller agents repo: `~/.local/app/agents/codex/nixbox/AGENTS.codexify.md` → `~/.local/app/agents/codex/nixbox/deploy-instructions` (`--check` = drift, rc 1). Check on the box: `~/.codexify/bin/codexify projects list --query trading` → selector `.`.

## Idle task

- `user-1001.slice` (`eturkes`) = `CPUWeight=idle` (cgroup `cpu.idle` yields to siblings only ⇒ set on the user slice, sibling of `offload`'s) + `MemoryHigh=20G` = 2 × the slice's recorded peak (9.9G over a 36 h boot, `codexify.service` incl. page cache): above it the kernel throttles + reclaims into swap, never kills. Check: an `eturkes` allocator of 24 GiB (urandom) holds the slice at 20.0 GiB (`memory.events` high 0 → 856, 4.1 GiB swapped) while a concurrent 4 GiB offload run completes rc 0. IO weight unenforced (NVMe scheduler `none`).
- Check: 8 CPU burners as `eturkes` → an 8-way CPU job as `offload` takes ≈ its solo time (1877 ms vs 1532–1789 ms solo); the same burners as `offload` → 3847 ms.
- Reverse cost (trading's intraday watcher, 90 s 8-way burn via `offload run`, offload at 7.9 of 8 CPUs): the slice keeps its 0.04 CPU, but idle-class wakeups lag (sleep overshoot p99 0.7 → 119 ms) ⇒ frame interval 13.5 → 20.8 s, decision-time quote age 1.3 → 3.9 s median, max 4.6 s against the model's 5 s `max_quote_age_ms` gate. A quote past the gate is dropped for that frame (market coverage `unavailable`, counted as a rejection); seen once live (7.1 s, frame gap 21 s = the burn's signature).
