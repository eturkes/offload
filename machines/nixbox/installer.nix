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
  # Host key generated on the controller (gitignored, like wifi.env) ⇒ `install` pins its .pub:
  # a LAN MITM during a reinstall fails host-key verification before wifi.env or the recovery key
  # cross. A mode ⇒ copied into /etc (root, 0600), not a store symlink sshd would refuse.
  services.openssh.hostKeys = [
    {
      type = "ed25519";
      path = "/etc/ssh/ssh_host_ed25519_key";
    }
  ];
  environment.etc."ssh/ssh_host_ed25519_key" = {
    source = ./installer_host_key;
    mode = "0600";
  };

  isoImage.squashfsCompression = "zstd -Xcompression-level 6";
}
