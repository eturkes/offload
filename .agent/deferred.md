# Deferred

- `machines/nixbox/install` EFI-entry rewrite (stale `Linux Boot Manager` → USB fall-through) landed untested; nested quoting is fragile. Accept: next reinstall boots the disk with no manual `efibootmgr`.
- nixbox BIOS 1.5.4 (old) → firmware updates via `fwupd`. A capsule may change PCR 7 → TPM refuses → recovery key at console. Accept: `fwupdmgr get-updates` empty + box reboots unattended after.
- LUKS header/root-swap attack (attacker-crafted volume boots, asks TPM for the real key under unchanged PCR 7). Accept: TPM policy also binds the volume (systemd `tpm2-measure-pcr=yes` + PCR 15 policy or `systemd-pcrlock`), and a swapped-header boot fails to unseal.
- `offload run` sync-back pulls every remote file except `.git` (no `.gitignore` filter on the downlink) → remote-built venvs/caches land locally. Accept: sync-back honors `.gitignore`, opt-in flag for ignored outputs.
- `offload run` from one cwd twice in parallel shares one remote workdir → clobbers. Accept: concurrent runs get distinct workdirs or a lock with a clear message.
- nixbox on Wi-Fi only → rsync of large inputs is link-bound. Accept: USB-C Ethernet adapter added to `network.nix` + measured throughput recorded.
