# nixbox

Dell Latitude 7400 (i7-8665U 4C/8T, 62 GiB, 2 TB NVMe, TPM 2.0), Wi-Fi only → `192.168.40.240`. Login `offload@192.168.40.240`, key-only; root SSH off; no account has a password.

## Files

- `host.nix` = shared facts (iface, address, stateVersion, controller key).
- `configuration.nix` = installed system; `network.nix` = NM Wi-Fi profile + static IPv4 (installer + system); `secureboot.nix` = lanzaboote v1.2.0.
- `installer.nix` = headless ISO: Wi-Fi + static IP + sshd with controller key.
- `wifi.nix` = `{ ssid; psk; }`, gitignored, mode 600 on the box. Regenerate from the controller's NM profile `<redacted>`.
- `hardware-configuration.nix` = written by `install`.

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
