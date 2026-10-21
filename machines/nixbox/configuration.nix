# nixbox = dedicated offload pool host: Dell Latitude 7400 (i7-8665U 4C/8T, 62 GiB, 2 TB NVMe),
# lid shut on AC, Wi-Fi only (no Ethernet port).
{ pkgs, lib, ... }:
let
  host = import ./host.nix;
in
{
  imports = [ ./hardware-configuration.nix ./network.nix ./secureboot.nix ];

  boot.loader.efi.canTouchEfiVariables = true;
  boot.tmp.cleanOnBoot = true;
  # Root = LUKS2 unlocked by the TPM → unattended reboots; no passphrase slot (recovery key alone).
  boot.initrd.systemd.enable = true;
  boot.initrd.luks.devices.cryptroot.crypttabExtraOpts = [ "tpm2-device=auto" ];
  # consoleblank: panel off after 60 s → less heat under a shut lid.
  # intel_iommu: DMA isolation against Thunderbolt-port attacks on the unlocked RAM.
  boot.kernelParams = [ "consoleblank=60" "intel_iommu=on" ];

  # iwlwifi firmware = the box's only network link; microcode rides the same switch.
  hardware.enableRedistributableFirmware = true;
  hardware.cpu.intel.updateMicrocode = true;
  services.fwupd.enable = true;

  networking.hostName = "nixbox";
  networking.networkmanager.wifi.powersave = false;

  # SSH reachable from the home LAN alone, IPv6 global addresses included.
  networking.nftables.enable = true;
  networking.firewall.extraInputRules = "ip saddr 192.168.40.0/24 tcp dport 22 accept";
  services.openssh = {
    enable = true;
    openFirewall = false;
    settings = { PasswordAuthentication = false; KbdInteractiveAuthentication = false; PermitRootLogin = "no"; };
  };

  users.users.offload = {
    isNormalUser = true;
    extraGroups = [ "wheel" ];
    openssh.authorizedKeys.keys = host.authorizedKeys;
    linger = true; # user services + tmux sessions outlive the ssh connection
  };
  security.sudo.wheelNeedsPassword = false;

  # Compute posture: full clocks, thermald to hold them without hitting the trip point.
  powerManagement.cpuFreqGovernor = "performance";
  services.thermald.enable = true;
  services.earlyoom = { enable = true; freeMemThreshold = 3; };
  zramSwap.enable = true;
  services.fstrim.enable = true;

  # Always on AC → cap charge at 80 % to spare the battery.
  systemd.services.battery-charge-limit = {
    wantedBy = [ "multi-user.target" ];
    serviceConfig = { Type = "oneshot"; RemainAfterExit = true; };
    script = ''
      b=/sys/class/power_supply/BAT0
      echo 75 > $b/charge_control_start_threshold
      echo 80 > $b/charge_control_end_threshold
    '';
  };

  # Headless compute box: never suspend, lid shut is fine.
  services.logind.settings.Login = { HandleLidSwitch = "ignore"; HandleLidSwitchExternalPower = "ignore"; HandleLidSwitchDocked = "ignore"; };
  systemd.targets = lib.genAttrs [ "sleep" "suspend" "hibernate" "hybrid-sleep" ] (_: { enable = false; });

  nix.settings = {
    experimental-features = [ "nix-command" "flakes" ];
    trusted-users = [ "offload" ];
    auto-optimise-store = true;
  };
  nix.gc = { automatic = true; dates = "weekly"; options = "--delete-older-than 14d"; };
  system.autoUpgrade.enable = true;

  # Prebuilt toolchains (uv pythons, pip wheels, rustup) expect an FHS loader.
  programs.nix-ld = {
    enable = true;
    libraries = with pkgs; [ stdenv.cc.cc.lib zlib zstd bzip2 xz openssl libffi curl glib libxml2 ];
  };
  # Nix-built python bypasses nix-ld → wheels miss libstdc++; uv's standalone builds go through it.
  environment.variables.UV_PYTHON_PREFERENCE = "only-managed";
  virtualisation.podman.enable = true;
  environment.systemPackages = with pkgs; [
    rsync git gcc gnumake cmake pkg-config python3 uv tmux htop btop nvme-cli lm_sensors
    # Project bootstraps fetch pinned tools themselves (gitleaks, rustup, Verus, `just tools`).
    curl jq unzip xz just file
  ];

  system.stateVersion = host.stateVersion;
}
