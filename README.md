# offload

offload runs a heavy command on another machine when this machine is busy.
It uses the local slots of the `heavy` command first.
When every local slot is in use, it runs the command on a free host of an SSH pool on your local network.
When no slot is free anywhere, it waits.

The command sees the same project files on the pool host.
Its output and exit status come back as if it ran here.

## Requirements

On this machine (the controller):

- Linux with `bash`, `ssh`, `rsync`, `git`, util-linux (`setsid`) and procps (`pkill`, `pgrep`).
- Go 1.27.3 or later, to build offload.
- Key-based SSH access to each pool host, without a passphrase prompt.

On each pool host:

- `bash`, `rsync`, `git`, GNU coreutils, util-linux (`setsid`, `flock`) and procps.
- An SSH server. Set `ClientAliveInterval 15` and `ClientAliveCountMax 4`, so that a run stops within about one minute when the controller disappears.

## Install

Build the binary and put it on your `PATH`:

```sh
CGO_ENABLED=0 go build -trimpath -o ~/.local/bin/offload.new ./cmd/offload
mv ~/.local/bin/offload.new ~/.local/bin/offload
```

The `mv` replaces the binary in one step, so runs that are in progress are not affected.

## Configure the pool

Write one `[[host]]` table for each pool host in `~/.config/offload/pool.toml`:

```toml
[[host]]
name = "nixbox"
address = "offload@192.168.40.240"
slots = 2
```

- `name` identifies the host in messages. Use letters, digits, `.`, `_` and `-`.
- `address` is `[user@]host` or a host alias from your `~/.ssh/config`. To use a port or an IPv6 address, define an alias.
- `slots` is the number of runs the host takes at the same time. The default is 2.

To use a different file, set `OFFLOAD_CONFIG` to its path.
`pool.example.toml` contains this example.

## Run a command

Give the command after `run`:

```sh
offload run -- make -j8
```

offload puts the command in the first free place:

1. A free local `heavy` slot. offload shares these slots with `heavy`, so `heavy --status` shows its runs.
2. Otherwise, the pool host with a free slot and the lowest load.
3. Otherwise, it shows one waiting message and tries again every 2 seconds.

A line on stderr names the place, for example `offload: nixbox slot 1/2 (8 CPUs, 3% busy) → ~/offload/myproject-1a2b3c4d`.

Use these options to limit the choice:

| Option | Effect |
|--------|--------|
| `--local` | Use the local slots only. |
| `--pool` | Use the pool hosts only. |
| `--host NAME` | Use this pool host only. |
| `--gpu` | Use pool hosts with an NVIDIA GPU only. For GPU work on this machine, use your local GPU tools instead. |
| `--no-sync` | Do not copy files to or from the host. |
| `--get PATH` | After the run, also copy `PATH` back. Use this option for outputs that git ignores. A path that ends with `/` copies the directory contents. You can give this option more than once. |

`--host`, `--gpu`, `--no-sync` and `--get` also select `--pool`.

### What goes to the pool host

In a git repository, offload copies the repository from its top level, even when you start it in a subdirectory.
The command then runs in the same subdirectory on the host.

- All files that git does not ignore go to the host. Tracked files always go, also when an ignore rule matches them.
- The `.git` directory goes to the host, so the command can read the history.
- Ignored files, for example virtual environments, caches and build output, stay on each side. Each host keeps its own copy and builds it again when necessary.
- Files that you deleted here are also deleted on the host. Ignored files on the host are never deleted.

After the command stops, files that git does not ignore come back.
A file that you changed here during the run is not replaced.
Nothing is deleted here.
Ignored outputs come back only with `--get`.

In a directory that is not a git repository, offload copies the directory with rsync and applies the `.gitignore` files that rsync can read.

These repositories run on the local slots only: git worktrees, repositories with submodules, repositories that contain another repository, and repositories with an external git directory.

Do not change the files while offload copies them. offload does not take a snapshot.

### Exit status

offload exits with the exit status of the command.
It uses these values for its own results:

| Status | Meaning |
|--------|---------|
| 2 | The options or the pool file are not correct. |
| 125 | offload could not do its part, for example a copy to or from the host failed. The message on stderr tells what failed. |
| 129, 130, 143 | offload received HUP, INT or TERM and stopped the command. |

When the command itself exits with 2 or 125, offload also exits with that value.

### Stop a run

Stop offload in any way, also with `kill -9`.
The command and every process that it started in its session stop, on this machine or on the host.
offload sends TERM first and KILL 10 seconds later.
A process that starts its own session is not stopped.

## See the pool

```sh
offload status
```

The output shows the local `heavy` slots and their holders, then one line for each pool host: its slots in use, CPUs, load, free memory and GPUs.

## Give a pool host an idle task

A pool host can run a task of your choice when offload does not use it.
Declare the task in the host's configuration in `machines/<host>/`.
Run it as its own user and give that user's systemd slice idle CPU weight.
offload runs then take the CPU from the task at once.

`machines/nixbox/` is the NixOS configuration of the pool host `nixbox`. See `machines/nixbox/README.md`.

## Develop

Run the full check chain before you commit:

```sh
.agent/gate
```

The chain checks formatting, lint, tests (with the race detector), the build, known vulnerabilities, secrets in the git history, the CI workflows, shell scripts and Nix files.

## License

Apache-2.0 WITH LLVM-exception. See `LICENSE` and `NOTICE`.
