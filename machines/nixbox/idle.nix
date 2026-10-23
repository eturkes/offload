# Idle task = trading, owned by `eturkes`: that user's whole slice runs at idle CPU weight ⇒ offload
# runs (user `offload`) preempt it at once. cgroup `cpu.idle` yields only to siblings ⇒ the weight sits
# on user-1001.slice, sibling of offload's user-1000.slice; a slice nested under user-1001 would still
# split CPU 50/50 with offload. IO weight is unenforced here (NVMe scheduler `none`, no io.cost).
_:
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

  # MemoryHigh = 2 × the slice's recorded peak (9.9G over a 36 h boot, codexify.service incl. page
  # cache; trading's trial units ≤ 384M): above it the kernel throttles + reclaims, never kills ⇒ a
  # runaway chat job slows down instead of driving the box into earlyoom, whose kill can hit an
  # offload run.
  systemd.slices."user-1001" = {
    overrideStrategy = "asDropin"; # keep the user-.slice template's settings
    sliceConfig = {
      CPUWeight = "idle";
      MemoryHigh = "20G";
    };
  };
}
