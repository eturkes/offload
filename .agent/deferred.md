# Deferred

- `machines/nixbox/install` EFI-entry rewrite (stale `Linux Boot Manager` → USB fall-through) landed untested; nested quoting is fragile. Accept: next reinstall boots the disk with no manual `efibootmgr`.
- nixbox BIOS 1.5.4 (old) → firmware updates via `fwupd`. A capsule may change PCR 7 → TPM refuses → recovery key at console. Accept: `fwupdmgr get-updates` empty + box reboots unattended after.
- LUKS header/root-swap attack (attacker-crafted volume boots, asks TPM for the real key under unchanged PCR 7). Accept: TPM policy also binds the volume (systemd `tpm2-measure-pcr=yes` + PCR 15 policy or `systemd-pcrlock`), and a swapped-header boot fails to unseal.
- nixbox on Wi-Fi only → rsync of large inputs is link-bound. Accept: USB-C Ethernet adapter added to `network.nix` + measured throughput recorded.
- Pool hosts have no admission control: concurrent `offload run`s from several sessions all stack on the least-loaded host. Accept: per-host slots (like `heavy`) or load-aware wait, shown by `offload status`.
- Sync honors per-dir `.gitignore` alone (not `.git/info/exclude`, global excludes, or a `.git` file of a git worktree). Accept: upload set = `git ls-files -co --exclude-standard` + `.git` resolved for worktrees.
