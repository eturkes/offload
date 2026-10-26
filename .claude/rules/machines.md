---
paths:
  - "machines/**"
---
# Pool machine configs

- `machines/<host>/wifi.env` = Wi-Fi secret (`WIFI_SSID=`, `WIFI_PSK=`), gitignored → never commit, never print; regenerate from the controller's NetworkManager profile. It reaches the box as root-only `/etc/nixos/wifi.env` (`deploy`, `install`); config references it by string path alone: a Nix path value or an inline secret lands in the world-readable `/nix/store`.
- `machines/nixbox/*_test.go` = Go tests of the host scripts (package `nixbox_test`: stub tools first on PATH, no host contact) ⇒ `go test ./...` + the gate's test step enter `machines/nixbox`.
- Config change → `machines/<host>/deploy`, then confirm the box answers SSH; Wi-Fi + static IP (`network.nix`) = the only link, a bad network change strands the box (recovery = installer stick + `install`).
- nixbox boot chain (lanzaboote, Secure Boot keys, firmware) changes can move TPM PCR 7 → LUKS stops auto-unlocking → recovery key at the console. Read `machines/nixbox/README.md` before touching `secureboot.nix`, `boot.*`, sbctl, fwupd or the LUKS tokens; re-seal = `machines/nixbox/tpm-bind`.
- `machines/<host>/installer_host_key` = the installer ISO's sshd host key (private; `.pub` pinned by `install`), gitignored → never commit, never copy to a pool host (its store is world-readable); regenerate with `ssh-keygen` (runbook `Files`), then rebuild the ISO.
- `machines/nixbox/check` fires (rc 1) on planted facts: `check --dump` output + `wrapper pkexec /bin/sh` (new string), + `wrapper ping -` (new wrapper), + `wrapper doas ?` (real program not found), minus a reviewed line (gone wrapper); upgrade lines removed + `upgrade done <now - 10 d>` (stale), + `upgrade failed <now - 1 h>` (failed last run), no upgrade line, + `failed <unit>`; `sshd passwordauthentication yes`, + `listen tcp 0.0.0.0:8080 python3`, `fw temp-ports 1`, `fw input-policy accept`, `secret psk-files 1` (the PSK scan's positive control: the pre-fix `-lan` store path → 1 file); the saved live facts and a last success 71 h old → rc 0.
- Recovery key lives on the controller only (`~/.local/share/offload/nixbox-recovery-key`), never in the repo or logs.
- nixbox Codexify (`~/AGENTS.md`, catalogue entry, worktree mode) → controller agents repo `codex/nixbox/deploy-instructions`, not `machines/nixbox/deploy`; hand edits on the box drift from that source (`--check` reports them).
- nixbox reboot = kills offload runs on it + ends trading's live trials (units stay disabled) ⇒ hold both controller slot locks for the window, then tell the owner which trials need a restart through Codexify. `nixos-rebuild switch` leaves trading's user units running.
