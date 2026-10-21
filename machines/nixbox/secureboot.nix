# Secure Boot with own keys (lanzaboote signs the UKIs; keys in /var/lib/sbctl on the box).
# Firmware enforcement + TPM PCR 7 binding = runbook in README.md.
{
  config,
  pkgs,
  lib,
  ...
}:
let
  lanzaboote = import (builtins.fetchTarball {
    url = "https://github.com/nix-community/lanzaboote/archive/v1.2.0.tar.gz";
    sha256 = "0syvqpi77ia1vd189658g53mqjgw7i39zym2r0g7asdjkrz0mgmj";
  }) { inherit pkgs; };
in
{
  imports = [ lanzaboote.nixosModules.lanzaboote ];
  boot.loader.systemd-boot.enable = lib.mkForce false;
  boot.lanzaboote = {
    enable = true;
    pkiBundle = "/var/lib/sbctl";
    configurationLimit = 10;
  };
  environment.systemPackages = [ pkgs.sbctl ];
  # fwupd ≥ 2.1 ignores lanzaboote's FWUPD_EFIAPPDIR (fwupd#10202) → it looks for the signed app in
  # fwupd-efi's store dir ⇒ capsule staging fails under Secure Boot (`fwupdx64.efi.signed cannot be
  # found`). Bind lanzaboote's /run/fwupd-efi over that dir, in fwupd's namespace alone. Drop once
  # nixpkgs builds fwupd with efi_app_location = /run/fwupd-efi (nixpkgs#553136: master, not 26.05).
  systemd.services.fwupd = lib.mkIf config.services.fwupd.enable {
    serviceConfig.BindReadOnlyPaths = [
      "-/run/fwupd-efi:${config.services.fwupd.package.fwupd-efi}/libexec/fwupd/efi"
    ];
  };
}
