---
paths:
  - "machines/**"
---
# Pool machine configs

- `machines/<host>/wifi.nix` = Wi-Fi secret, gitignored → never commit, never print; regenerate from the controller's NetworkManager profile.
- Config change → `machines/<host>/deploy`, then confirm the box answers SSH; Wi-Fi + static IP (`network.nix`) = the only link, a bad network change strands the box (recovery = installer stick + `install`).
- nixbox boot chain (lanzaboote, Secure Boot keys, firmware) changes can move TPM PCR 7 → LUKS stops auto-unlocking → recovery key at the console. Read `machines/nixbox/README.md` before touching `secureboot.nix`, `boot.*`, sbctl, fwupd or the LUKS tokens; re-seal = `machines/nixbox/tpm-bind`.
- Recovery key lives on the controller only (`~/.local/share/offload/nixbox-recovery-key`), never in the repo or logs.
- nixbox Codexify (`~/AGENTS.md`, catalogue entry, worktree mode) → controller agents repo `codex/nixbox/deploy-instructions`, not `machines/nixbox/deploy`; hand edits on the box drift from that source (`--check` reports them).
- nixbox reboot = kills offload runs on it + ends trading's live trials (units stay disabled) ⇒ hold both controller slot locks for the window, then tell the owner which trials need a restart through Codexify. `nixos-rebuild switch` leaves trading's user units running.
