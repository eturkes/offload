---
paths:
  - "machines/**"
---
# Pool machine configs

- `machines/<host>/wifi.nix` = Wi-Fi secret, gitignored → never commit, never print; regenerate from the controller's NetworkManager profile.
- Config change → `machines/<host>/deploy`, then confirm the box answers SSH; Wi-Fi + static IP (`network.nix`) = the only link, a bad network change strands the box (recovery = installer stick + `install`).
- nixbox boot chain (lanzaboote, Secure Boot keys, firmware) changes can move TPM PCR 7 → LUKS stops auto-unlocking → recovery key at the console. Read `machines/nixbox/README.md` before touching `secureboot.nix`, `boot.*`, sbctl or fwupd.
- Recovery key lives on the controller only (`~/.local/share/offload/nixbox-recovery-key`), never in the repo or logs.
