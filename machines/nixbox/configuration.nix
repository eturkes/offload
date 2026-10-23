# nixbox = dedicated offload pool host: Dell Latitude 7400 (i7-8665U 4C/8T, 62 GiB, 2 TB NVMe),
# lid shut on AC, Wi-Fi only (no Ethernet port).
{ pkgs, lib, ... }:
let
  host = import ./host.nix;
  # nixos-unstable channel rev (Hydra-built ⇒ binary cache) for packages 26.05 ships too old.
  unstable = import (builtins.fetchTarball {
    url = "https://github.com/NixOS/nixpkgs/archive/02c087f3fe673c9e317f206a627cb6b910336dbe.tar.gz";
    sha256 = "1l60ar0j7j7jrrg4r7j4kscrih0qn2n0lslffzflhv86lyq650r2";
  }) { inherit (pkgs.stdenv.hostPlatform) system; };
in
{
  imports = [
    ./hardware-configuration.nix
    ./network.nix
    ./secureboot.nix
    ./idle.nix
    ./performance.nix
  ];

  boot.loader.efi.canTouchEfiVariables = true;
  # Root = LUKS2 unlocked by the TPM → unattended reboots; no passphrase slot (recovery key alone).
  # The volume key goes into PCR 15 right after unlock; the token demands PCR 15 = 0 (`tpm-bind`) ⇒
  # a swapped-header volume unlocked first leaves the real token sealed.
  boot.initrd.systemd.enable = true;
  boot.initrd.luks.devices.cryptroot.crypttabExtraOpts = [
    "tpm2-device=auto"
    "tpm2-measure-pcr=yes"
  ];
  # consoleblank: panel off after 60 s → less heat under a shut lid.
  # intel_iommu: DMA isolation against Thunderbolt-port attacks on the unlocked RAM.
  boot.kernelParams = [
    "consoleblank=60"
    "intel_iommu=on"
  ];

  # iwlwifi firmware = the box's only network link; microcode rides the same switch.
  hardware.enableRedistributableFirmware = true;
  hardware.cpu.intel.updateMicrocode = true;
  services.fwupd.enable = true;

  networking.hostName = "nixbox";
  # Writes /etc/localtime (absent without it: Lean's zone lookup fails); UTC = glibc's no-file
  # behavior ⇒ trading's clocks unchanged.
  time.timeZone = "UTC";
  networking.networkmanager.wifi.powersave = false;

  # SSH reachable from the home LAN alone, IPv6 global addresses included.
  networking.nftables.enable = true;
  networking.firewall.extraInputRules = "ip saddr 192.168.40.0/24 tcp dport 22 accept";
  services.openssh = {
    enable = true;
    openFirewall = false;
    settings = {
      PasswordAuthentication = false;
      KbdInteractiveAuthentication = false;
      PermitRootLogin = "no";
      # A vanished offload controller → its session closes within ~1 min → the run's watcher sees
      # stdin EOF and tears the command down (offload contract X5).
      ClientAliveInterval = 15;
      ClientAliveCountMax = 4;
    };
  };

  users.users.offload = {
    isNormalUser = true;
    extraGroups = [ "wheel" ];
    openssh.authorizedKeys.keys = host.authorizedKeys;
    linger = true; # user services + tmux sessions outlive the ssh connection
    # Projects pin the controller's node 26 (`node --test <dir>` needs it); trading stays on the
    # system node.
    packages = [ pkgs.nodejs_26 ];
  };
  security.sudo.wheelNeedsPassword = false;

  # Compute posture: full clocks. thermald --adaptive follows Dell's DPTF tables, which held PL1 at
  # 9–12 W ⇒ off; PL1 = 22 W (the firmware's own PL1 max), the CPU's TCC (98 °C) paces the rest.
  # Lid shut, fan at max throughout, 8-thread stress-ng: 12 W → 2.3 GHz 84 °C; 15 W → 2.5 GHz 93 °C;
  # 22 W → 2.6 GHz 99 °C (cooling-bound at ~17 W; 30 W adds nothing).
  powerManagement.cpuFreqGovernor = "performance";
  systemd.services.cpu-power-limit = {
    wantedBy = [ "multi-user.target" ];
    serviceConfig = {
      Type = "oneshot";
      RemainAfterExit = true;
    };
    script = ''
      echo performance > /sys/firmware/acpi/platform_profile
      echo 22000000 > /sys/class/powercap/intel-rapl-mmio:0/constraint_0_power_limit_uw
    '';
  };
  services.earlyoom = {
    enable = true;
    freeMemThreshold = 3;
  };
  zramSwap.enable = true;
  services.fstrim.enable = true;

  # Always on AC → cap charge at 80 % to spare the battery. dell-laptop threshold writes only store
  # the custom values; the firmware keeps its current mode (Adaptive) until charge_types = Custom.
  systemd.services.battery-charge-limit = {
    wantedBy = [ "multi-user.target" ];
    serviceConfig = {
      Type = "oneshot";
      RemainAfterExit = true;
    };
    script = ''
      b=/sys/class/power_supply/BAT0
      echo 75 > $b/charge_control_start_threshold
      echo 80 > $b/charge_control_end_threshold
      echo Custom > $b/charge_types
    '';
  };

  # Headless compute box: never suspend, lid shut is fine.
  services.logind.settings.Login = {
    HandleLidSwitch = "ignore";
    HandleLidSwitchExternalPower = "ignore";
    HandleLidSwitchDocked = "ignore";
  };
  systemd.targets = lib.genAttrs [ "sleep" "suspend" "hibernate" "hybrid-sleep" ] (_: {
    enable = false;
  });

  nix.settings = {
    experimental-features = [
      "nix-command"
      "flakes"
    ];
    trusted-users = [ "offload" ];
    auto-optimise-store = true;
  };
  nix.gc = {
    automatic = true;
    dates = "weekly";
    options = "--delete-older-than 14d";
  };
  system.autoUpgrade.enable = true;

  # Prebuilt toolchains (uv pythons, pip wheels, rustup) expect an FHS loader.
  programs.nix-ld = {
    enable = true;
    libraries = with pkgs; [
      stdenv.cc.cc.lib
      zlib
      zstd
      bzip2
      xz
      openssl
      libffi
      curl
      glib
      libxml2
      # Headless Chromium (Playwright's chrome-headless-shell, as projects' e2e suites fetch it).
      alsa-lib
      at-spi2-core
      dbus
      expat
      libgbm
      nspr
      nss
      systemdLibs
      libx11
      libxcb
      libxcomposite
      libxdamage
      libxext
      libxfixes
      libxrandr
      libxkbcommon
    ];
  };
  # Nix-built python bypasses nix-ld → wheels miss libstdc++; uv's standalone builds go through it.
  environment.variables.UV_PYTHON_PREFERENCE = "only-managed";
  # uv's standalone pythons (OpenSSL dir /etc/ssl) read /etc/ssl/cert.pem, which NixOS lacks ⇒ 0
  # trusted CAs, every stdlib TLS call fails verification — in services as in offload's `bash -lc`.
  environment.etc."ssl/cert.pem".source = "/etc/ssl/certs/ca-certificates.crt";
  # Offload runs land here first ⇒ FHS paths in project code (`#!/bin/bash`, `/usr/bin/git`) must
  # resolve: envfs serves /bin + /usr/bin from the calling process's PATH — a setuid caller's too (no
  # AT_SECURE check) ⇒ `check` rules every wrapper's /bin-style exec (wrappers.reviewed).
  services.envfs.enable = true;
  # A caller whose PATH holds no Nix dir (suites pinning PATH=<mock>:/usr/bin:/bin) falls back to the
  # system profile, not just envfs's own sh + env.
  fileSystems."/usr/bin".options = [ "fallback-path=/run/current-system/sw/bin" ];
  # Headless browsers lay out with these ⇒ the controller's resolved families (Noto CJK JP), so UI
  # suites tap the same targets there as here.
  fonts.packages = with pkgs; [
    noto-fonts
    noto-fonts-cjk-sans
    noto-fonts-cjk-serif
    noto-fonts-color-emoji
  ];
  fonts.fontconfig.defaultFonts = {
    sansSerif = [ "Noto Sans CJK JP" ];
    serif = [ "Noto Serif CJK JP" ];
    monospace = [ "Noto Sans Mono CJK JP" ];
    emoji = [ "Noto Color Emoji" ];
  };
  # rustup ships no default toolchain ⇒ a bare `cargo` (outside any rust-toolchain.toml) fails.
  systemd.services.rustup-default = {
    wantedBy = [ "multi-user.target" ];
    wants = [ "network-online.target" ];
    after = [ "network-online.target" ];
    path = [ pkgs.rustup ];
    serviceConfig = {
      Type = "oneshot";
      RemainAfterExit = true;
      User = "offload";
    };
    script = "rustup default >/dev/null 2>&1 || rustup default stable";
  };
  # Suites commit in scratch repos ⇒ offload runs need a git identity; offload's alone (a system
  # one would sign trading's commits as eturkes).
  # tmpfiles would create missing parents root-owned ⇒ both dirs declared.
  systemd.tmpfiles.rules =
    let
      gitconfig = pkgs.writeText "offload-gitconfig" ''
        [user]
          name = offload
          email = offload@nixbox.invalid
      '';
    in
    [
      "d /home/offload/.config 0755 offload users -"
      "d /home/offload/.config/git 0755 offload users -"
      "L+ /home/offload/.config/git/config - - - - ${gitconfig}"
    ];
  virtualisation.podman.enable = true;
  # `offload ctr` (pool run environment): tools/install puts the controller's offload binary in
  # ~/.local/bin; OFFLOAD_CTR=1 = the per-machine switch, so ctr commands run in the project's image.
  environment.localBinInPath = true;
  environment.variables.OFFLOAD_CTR = "1";
  environment.systemPackages = with pkgs; [
    rsync
    git
    gcc
    gnumake
    cmake
    pkg-config
    python3
    uv
    tmux
    htop
    btop
    nvme-cli
    lm_sensors
    # Project bootstraps fetch pinned tools themselves (gitleaks, rustup, Verus, `just tools`).
    curl
    jq
    unzip
    xz
    just
    file
    # Toolchain managers: each fetches the repo's own pin (lean-toolchain, rust-toolchain.toml,
    # go.mod via GOTOOLCHAIN=auto).
    elan
    rustup
    go
    # Gate tools projects take from the host: shellcheck; swipl 10.0.2 = cnl-ckc's pin (26.05: 9.2.9,
    # refused by its identity gate) → unstable, frozen at its rev; bwrap =
    # `lake check`'s sandbox; clang's analyzer (lowPrio: gcc keeps `cc`).
    shellcheck
    unstable.swi-prolog
    bubblewrap
    (lib.lowPrio clang)
    # Trading (contracts build + tests) + Codexify's tool checks.
    nodejs
    pnpm
    ripgrep
    gh
  ];

  system.stateVersion = host.stateVersion;
}
