# Wi-Fi + static LAN address through NetworkManager (the 26.05 installer ISO already runs it).
{ lib, ... }:
let
  host = import ./host.nix;
  wifi = if builtins.pathExists ./wifi.nix then import ./wifi.nix else null;
in
{
  networking.networkmanager.enable = true;
  networking.networkmanager.ensureProfiles.profiles = lib.optionalAttrs (wifi != null) {
    lan = {
      connection = {
        id = "lan";
        type = "wifi";
        interface-name = host.wlan;
        autoconnect = true;
      };
      wifi = {
        inherit (wifi) ssid;
        mode = "infrastructure";
      };
      wifi-security = {
        key-mgmt = "wpa-psk";
        inherit (wifi) psk;
      };
      ipv4 = {
        method = "manual";
        address1 = "${host.address}/24,192.168.40.1";
        dns = "192.168.40.1;";
      };
      ipv6.method = "auto";
    };
  };
}
