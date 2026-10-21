# Idle task = trading, owned by `eturkes`: that user's whole slice runs at idle CPU weight ⇒ offload
# runs (user `offload`) preempt it at once. cgroup `cpu.idle` yields only to siblings ⇒ the weight sits
# on user-1001.slice, sibling of offload's user-1000.slice; a slice nested under user-1001 would still
# split CPU 50/50 with offload. IO weight is unenforced here (NVMe scheduler `none`, no io.cost).
{ ... }:
let
  host = import ./host.nix;
in
{
  users.users.eturkes = {
    isNormalUser = true;
    uid = 1001;
    openssh.authorizedKeys.keys = host.authorizedKeys;
    linger = true; # trading's user services run without a login
  };

  systemd.slices."user-1001" = {
    overrideStrategy = "asDropin"; # keep the user-.slice template's settings
    sliceConfig.CPUWeight = "idle";
  };
}
