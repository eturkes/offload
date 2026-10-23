# Headless installer ISO: joins Wi-Fi, takes the static address, sshd up with the controller key.
{ modulesPath, lib, ... }:
let
  host = import ./host.nix;
in
{
  imports = [
    "${modulesPath}/installer/cd-dvd/installation-cd-minimal.nix"
    ./network.nix
  ];

  systemd.services.sshd.wantedBy = lib.mkForce [ "multi-user.target" ];
  users.users.nixos.openssh.authorizedKeys.keys = host.authorizedKeys;
  users.users.root.openssh.authorizedKeys.keys = host.authorizedKeys;
  # network.nix's env file; the ISO's store is the live system's alone.
  environment.etc."nixos/wifi.env".source = ./wifi.env;

  isoImage.squashfsCompression = "zstd -Xcompression-level 6";
}
