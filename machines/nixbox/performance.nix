# Performance posture: boot-time parts take effect at the next reboot.
_: {
  # fsync in an on-disk /tmp costs ≈ 11 ms here (dm-crypt ext4) vs ≈ 0.1 ms on tmpfs; the controller's
  # /tmp is tmpfs too. 50 % of RAM: a run writing more gets ENOSPC.
  boot.tmp = {
    useTmpfs = true;
    tmpfsSize = "50%";
  };
}
