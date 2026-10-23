# Wi-Fi + static LAN address through NetworkManager (the 26.05 installer ISO already runs it).
# SSID + PSK come from /etc/nixos/wifi.env (root, 600; `deploy` + `install` copy it): the store is
# world-readable ⇒ the profile there holds `$WIFI_*` placeholders, envsubst'd at activation.
_:
let
  host = import ./host.nix;
in
{
  networking.networkmanager.enable = true;
  networking.networkmanager.ensureProfiles.environmentFiles = [ "/etc/nixos/wifi.env" ];
  networking.networkmanager.ensureProfiles.profiles.lan = {
    connection = {
      id = "lan";
      type = "wifi";
      interface-name = host.wlan;
      autoconnect = true;
    };
    wifi = {
      ssid = "$WIFI_SSID";
      mode = "infrastructure";
    };
    wifi-security = {
      key-mgmt = "wpa-psk";
      psk = "$WIFI_PSK";
    };
    ipv4 = {
      method = "manual";
      address1 = "${host.address}/24,192.168.40.1";
      dns = "192.168.40.1;";
    };
    ipv6.method = "auto";
  };
}
