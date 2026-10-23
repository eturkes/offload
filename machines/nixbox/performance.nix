# Performance posture = the controller's (CachyOS: kernel command line + cachyos-settings), owner
# ruling: performance first where it conflicts with hardening. Boot-time parts (command line, /tmp)
# take effect at the next reboot.
_: {
  boot.kernelParams = [
    "mitigations=off"
    "nowatchdog"
    "transparent_hugepage=always"
    "zswap.enabled=0" # zram is the swap (cachyos-settings 30-zram.rules)
  ];
  boot.kernel.sysctl = {
    "vm.swappiness" = 150; # cachyos-settings' value once zram is up
    "vm.vfs_cache_pressure" = 50;
    "vm.dirty_bytes" = 268435456;
    "vm.dirty_background_bytes" = 67108864;
    "vm.dirty_writeback_centisecs" = 1500;
    "vm.page-cluster" = 0;
    "kernel.nmi_watchdog" = 0;
    "kernel.split_lock_mitigate" = 0;
    "kernel.kptr_restrict" = 2;
    "net.core.netdev_max_backlog" = 4096;
  };
  # zstd = the controller's algorithm (≈ 4:1 on its live pages); 2 × RAM (owner).
  zramSwap = {
    enable = true;
    algorithm = "zstd";
    memoryPercent = 200;
    priority = 100;
  };
  # fsync in an on-disk /tmp costs ≈ 11 ms here (dm-crypt ext4) vs ≈ 0.1 ms on tmpfs; the controller's
  # /tmp is tmpfs too. 50 % of RAM: a run writing more gets ENOSPC.
  boot.tmp = {
    useTmpfs = true;
    tmpfsSize = "50%";
  };
  systemd.tmpfiles.rules = [
    "w! /sys/kernel/mm/transparent_hugepage/defrag - - - - defer+madvise"
    "w! /sys/kernel/mm/transparent_hugepage/khugepaged/max_ptes_none - - - - 409"
  ];
  services.udev.extraRules = ''
    ACTION=="add|change", KERNEL=="nvme[0-9]*", ATTR{queue/rotational}=="0", ATTR{queue/scheduler}="kyber"
  '';
  systemd.settings.Manager.DefaultLimitNOFILE = "2048:2097152";
}
