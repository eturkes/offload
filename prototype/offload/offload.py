#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.11"
# ///
"""offload — run heavy commands on the least-loaded machine in an SSH pool."""

import argparse
import concurrent.futures as cf
import fcntl
import hashlib
import os
import shlex
import subprocess
import sys
import tomllib
from pathlib import Path

CONFIG = Path(os.environ.get("OFFLOAD_CONFIG", "~/.config/offload/pool.toml")).expanduser()
SSH_OPTS = ["-o", "BatchMode=yes", "-o", "ConnectTimeout=3", "-o", "StrictHostKeyChecking=accept-new"]
REMOTE_ROOT = "offload"  # relative to the remote $HOME
LOCKS = Path(os.environ.get("XDG_RUNTIME_DIR", "/tmp")) / "offload"
# Command runs in its own session with stdin closed; the watcher holds ssh's stdin and kills that
# session at EOF, which arrives when the local offload/ssh dies (timeout, kill -9) → no remote orphans.
# fd 3 = ssh's stdin: bash points background jobs' stdin at /dev/null.
WRAP = ('cd "$1" && exec 3<&0 && { setsid bash -lc "$2" </dev/null 3<&- & p=$!; '
        '{ cat <&3; kill -TERM -- -$p; } >/dev/null 2>&1 & wait $p; }')
# One round trip → nproc, loadavg, MemAvailable kB, GPU count.
PROBE = (
    "nproc; cut -d' ' -f1 /proc/loadavg; "
    "awk '/MemAvailable/{print $2}' /proc/meminfo; "
    "(nvidia-smi -L 2>/dev/null || true) | grep -c GPU || true"
)


def load_pool():
    if not CONFIG.exists():
        sys.exit(f"offload: no pool file at {CONFIG}. Create it with one [[host]] table per machine.")
    hosts = tomllib.loads(CONFIG.read_text()).get("host", [])
    if not hosts:
        sys.exit(f"offload: {CONFIG} lists no [[host]] entries.")
    return hosts


def probe(host):
    try:
        r = subprocess.run(["ssh", *SSH_OPTS, host["address"], PROBE],
                           capture_output=True, text=True, timeout=8)
    except subprocess.TimeoutExpired:
        return {**host, "up": False, "error": "timeout"}
    if r.returncode:
        return {**host, "up": False, "error": (r.stderr.strip().splitlines() or ["ssh failed"])[-1]}
    cpus, load, mem_kb, gpus = r.stdout.split()
    cpus, load = int(cpus), float(load)
    return {**host, "up": True, "cpus": cpus, "load": load, "mem_gb": int(mem_kb) / 2**20,
            "gpus": int(gpus), "score": load / cpus}


def probe_all(hosts):
    with cf.ThreadPoolExecutor(len(hosts)) as ex:
        return list(ex.map(probe, hosts))


def cmd_status(_args):
    rows = probe_all(load_pool())
    print(f"{'HOST':<14}{'STATE':<7}{'CPUS':>5}{'LOAD':>7}{'BUSY':>6}{'FREE MEM':>10}{'GPUS':>6}")
    for h in rows:
        if h["up"]:
            print(f"{h['name']:<14}{'up':<7}{h['cpus']:>5}{h['load']:>7.2f}{h['score']:>6.0%}"
                  f"{h['mem_gb']:>8.1f}GB{h['gpus']:>6}")
        else:
            print(f"{h['name']:<14}{'down':<7}  {h['error']}")


def pick(hosts, name=None, gpu=False):
    if name:
        hosts = [h for h in hosts if h["name"] == name]
        if not hosts:
            sys.exit(f"offload: no host named {name} in {CONFIG}.")
    up = [h for h in probe_all(hosts) if h["up"] and (h["gpus"] or not gpu)]
    if not up:
        sys.exit("offload: no reachable host matches. Run `offload status` to see the pool.")
    return min(up, key=lambda h: h["score"])


def rsync(src, dst, *opts):
    subprocess.run(["rsync", "-az", "--mkpath", *opts, "-e", "ssh " + " ".join(SSH_OPTS), src, dst], check=True)


# .gitignore'd paths never cross either way, so each side keeps its own venvs, toolchains + build
# dirs (remote ones are linked against the remote, local ones survive). .git goes up (gates read
# history: gitleaks, replays) and never comes back.
FILTER = ["--filter=:- .gitignore"]


def cmd_run(args):
    if not args.command:
        sys.exit("offload: give a command after `run`, e.g. `offload run -- make -j`.")
    cwd = Path.cwd()
    workdir = f"{REMOTE_ROOT}/{cwd.name}-{hashlib.sha1(str(cwd).encode()).hexdigest()[:8]}"
    # One controller derives every workdir ⇒ a local lock serialises runs that share one.
    LOCKS.mkdir(parents=True, exist_ok=True)
    lock = open(LOCKS / f"{Path(workdir).name}.lock", "w")
    try:
        fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
    except BlockingIOError:
        print(f"offload: another run from {cwd} holds ~/{workdir}; waiting.", file=sys.stderr)
        fcntl.flock(lock, fcntl.LOCK_EX)
    host = pick(load_pool(), args.host, args.gpu)
    addr = host["address"]
    print(f"offload: {host['name']} ({host['cpus']} CPUs, {host['score']:.0%} busy) → ~/{workdir}",
          file=sys.stderr)
    if args.sync:
        # Mirror: files deleted locally go remotely too; ignored remote files are kept (filter protects them).
        rsync(f"{cwd}/", f"{addr}:{workdir}/", "--delete", *FILTER)
    remote = shlex.join(["bash", "-c", WRAP, "offload", workdir, shlex.join(args.command)])
    ssh = subprocess.Popen(["ssh", *SSH_OPTS, addr, f"mkdir -p {shlex.quote(workdir)} && {remote}"],
                           stdin=subprocess.PIPE)
    try:
        rc = ssh.wait()
    except KeyboardInterrupt:
        ssh.stdin.close()
        rc = ssh.wait()
    if args.sync:
        # --update: a file edited locally during the run stays.
        rsync(f"{addr}:{workdir}/", f"{cwd}/", "--update", "--exclude=.git/", *FILTER)
    for path in args.get:
        rsync(f"{addr}:{workdir}/{path}", f"{cwd}/{path}", "--update")
    return rc


def main():
    p = argparse.ArgumentParser(prog="offload", description=__doc__)
    sub = p.add_subparsers(dest="cmd", required=True)
    sub.add_parser("status", help="probe every pool host for load, memory and GPUs")
    r = sub.add_parser("run", help="sync the current dir to the best host, run COMMAND, sync back")
    r.add_argument("--host", help="use this pool host instead of the least-loaded one")
    r.add_argument("--gpu", action="store_true", help="only consider hosts with an NVIDIA GPU")
    r.add_argument("--no-sync", dest="sync", action="store_false", help="run without file sync")
    r.add_argument("--get", action="append", default=[], metavar="PATH",
                   help="also copy back this ignored path (file, or dir with trailing /); repeatable")
    r.add_argument("command", nargs=argparse.REMAINDER)
    args = p.parse_args()
    if args.cmd == "run" and args.command[:1] == ["--"]:
        args.command = args.command[1:]
    sys.exit({"status": cmd_status, "run": cmd_run}[args.cmd](args))


if __name__ == "__main__":
    main()
