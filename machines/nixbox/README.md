# nixbox

Dell Latitude 7400 (i7-8665U 4C/8T, 62 GiB, 2 TB NVMe, TPM 2.0), Wi-Fi only → `192.168.40.240`. Logins (key-only, controller key): `offload@` = offload runs + deploys (wheel, passwordless sudo); `eturkes@` = trading + Codexify (no sudo). Root SSH off; no account has a password.

## Files

- `host.nix` = shared facts (iface, address, stateVersion, controller key).
- `configuration.nix` = installed system; `network.nix` = NM Wi-Fi profile + static IPv4 (installer + system); `secureboot.nix` = lanzaboote v1.2.0.
- `installer.nix` = headless ISO: Wi-Fi + static IP + sshd with controller key.
- `wifi.nix` = `{ ssid; psk; }`, gitignored, mode 600 on the box. Regenerate from the controller's NM profile `<redacted>`.
- `hardware-configuration.nix` = written by `install`.
- `efi-entry` = run by `install`: one `Linux Boot Manager` entry on the new ESP (stale ones from a reused disk deleted). Dry run on the live box: `ssh offload@<ip> 'sudo DRY_RUN=1 sh -s -- /dev/nvme0n1' < machines/nixbox/efi-entry`.
- `idle.nix` = idle task: user `eturkes` (uid 1001, unprivileged, lingering) owns trading; its `user-1001.slice` runs at `CPUWeight=idle` ⇒ offload runs (user `offload`) preempt it.

## Ops

- Change config → `machines/nixbox/deploy` (rsync → `/etc/nixos`, `nixos-rebuild switch`).
- Installer ISO → `nix-build '<nixpkgs/nixos>' -A config.system.build.isoImage -I nixos-config=machines/nixbox/installer.nix -I nixpkgs=https://channels.nixos.org/nixos-26.05/nixexprs.tar.xz` (controller has no nix → `podman run -v offload-nix:/nix … docker.io/nixos/nix`; `wifi.nix` must sit beside `installer.nix`).
- Reinstall (wipes disk) → stick in, `sudo efibootmgr --bootnext <USB entry>` + reboot on the box, then `machines/nixbox/install /dev/nvme0n1`.
- Recovery key (TPM refuses after BIOS reset/key change) = `~/.local/share/offload/nixbox-recovery-key` on the controller; typed at the console LUKS prompt.

## Disk + boot security

- Root = LUKS2 `cryptroot`; slots = `tpm2` + `recovery`, no passphrase.
- Secure Boot = own keys (sbctl, `/var/lib/sbctl`) + Microsoft (`--microsoft`, keeps option ROMs + fwupd capsules booting).
- TPM slot bound to PCR 7 (Secure Boot policy) ⇒ unattended reboots + upgrades; disabling SB or changing keys ⇒ TPM refuses ⇒ recovery key.
- Re-bind after key changes: `sudo systemd-cryptenroll --wipe-slot=tpm2 --tpm2-device=auto --tpm2-pcrs=7 --unlock-key-file=<(printf %s "$RK") /dev/nvme0n1p2`.
- `/boot/EFI/nixos/kernel-*.efi` shows unsigned in `sbctl verify` by design: the signed lanzaboote stub checks it by hash.

## Idle task

- `user-1001.slice` (`eturkes`) = `CPUWeight=idle` (cgroup `cpu.idle` yields to siblings only ⇒ set on the user slice, sibling of `offload`'s). IO weight unenforced (NVMe scheduler `none`).
- Check: 8 CPU burners as `eturkes` → an 8-way CPU job as `offload` takes ≈ its solo time (1877 ms vs 1532–1789 ms solo); the same burners as `offload` → 3847 ms.
