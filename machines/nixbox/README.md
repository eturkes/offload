# nixbox

Dell Latitude 7400 (i7-8665U 4C/8T, 62 GiB, 2 TB NVMe, TPM 2.0), Wi-Fi only → `192.168.40.240`. Logins (key-only, controller key): `offload@` = offload runs + deploys (wheel, passwordless sudo); `eturkes@` = trading + Codexify (no sudo). Root SSH off; no account has a password.

## Files

- `host.nix` = shared facts (iface, address, stateVersion, controller key).
- `configuration.nix` = installed system; `network.nix` = NM Wi-Fi profile + static IPv4 (installer + system); `secureboot.nix` = lanzaboote v1.2.0.
- `installer.nix` = headless ISO: Wi-Fi + static IP + sshd with controller key.
- `wifi.nix` = `{ ssid; psk; }`, gitignored, mode 600 on the box. Regenerate from the controller's NM profile `<redacted>`.
- `hardware-configuration.nix` = written by `install`.
- `tpm-bind` = re-seal the TPM2 LUKS token from the controller (Disk + boot security).
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
- TPM slot bound to PCR 7 (Secure Boot policy) + PCR 15 = 0; the initrd measures the volume key into PCR 15 right after unlock (`tpm2-measure-pcr=yes`) ⇒ a swapped LUKS header unlocked first (attacker passphrase) leaves the real token sealed. Unattended reboots + upgrades unaffected; disabling SB or changing keys ⇒ TPM refuses ⇒ recovery key.
- Re-seal (key or firmware changes) = `machines/nixbox/tpm-bind` (unlocks with the controller's recovery key: after boot PCR 15 ≠ 0, so the TPM token itself never authorizes changes).
- Attacker check: attach a scratch LUKS image with `systemd-cryptsetup attach <name> <loop> <keyfile> tpm2-measure-pcr=yes`, then `sudo env LD_LIBRARY_PATH=/run/current-system/sw/lib/cryptsetup cryptsetup open --test-passphrase --token-only --token-id <tpm2 token> <luks dev>` → refused (`TPM policy does not match`). The plugin path is needed: NixOS cryptsetup does not find systemd's token plugin.
- `/boot/EFI/nixos/kernel-*.efi` shows unsigned in `sbctl verify` by design: the signed lanzaboote stub checks it by hash.

## Firmware

- fwupd capsules boot under our Secure Boot: lanzaboote's `fwupd-efi.service` signs `fwupdx64.efi` with our db key into `/run/fwupd-efi` (`sbverify --cert /var/lib/sbctl/keys/db/db.pem` → OK), `DisableShimForSecureBoot=true`. Firmware lacks capsule-on-disk (`OsIndicationsSupported` = 0x03) ⇒ this signed app is the only path.
- BIOS + dbx updates move PCR 7 ⇒ the bound token refuses at the next boot. A BIOS update may also reset custom Secure Boot keys (Dell keeps active + default stores apart; no source confirms custom keys survive) ⇒ owner within reach of the console.
- Update: `machines/nixbox/tpm-bind --open` → on the box `sudo fwupdmgr update -y --no-reboot-check && sudo reboot` → box answers SSH → `fwupdmgr get-updates` empty → `machines/nixbox/tpm-bind` (re-seal to the new PCR 7, wipes the open token) → reboot once more, unattended.

## Idle task

- `user-1001.slice` (`eturkes`) = `CPUWeight=idle` (cgroup `cpu.idle` yields to siblings only ⇒ set on the user slice, sibling of `offload`'s). IO weight unenforced (NVMe scheduler `none`).
- Check: 8 CPU burners as `eturkes` → an 8-way CPU job as `offload` takes ≈ its solo time (1877 ms vs 1532–1789 ms solo); the same burners as `offload` → 3847 ms.
