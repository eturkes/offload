# Secure Boot with own keys (lanzaboote signs the UKIs; keys in /var/lib/sbctl on the box).
# Firmware enforcement + TPM PCR 7 binding = runbook in README.md.
{ pkgs, lib, ... }:
let
  lanzaboote = import (builtins.fetchTarball {
    url = "https://github.com/nix-community/lanzaboote/archive/v1.2.0.tar.gz";
    sha256 = "0syvqpi77ia1vd189658g53mqjgw7i39zym2r0g7asdjkrz0mgmj";
  }) { inherit pkgs; };
in
{
  imports = [ lanzaboote.nixosModules.lanzaboote ];
  boot.loader.systemd-boot.enable = lib.mkForce false;
  boot.lanzaboote = { enable = true; pkiBundle = "/var/lib/sbctl"; configurationLimit = 10; };
  environment.systemPackages = [ pkgs.sbctl ];
}
