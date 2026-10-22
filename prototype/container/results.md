# Pool run environment spike: results

This page records what the container spike measured, so that you can decide the direction in ITERATE.
The spike is a prototype. Its scripts are shortcuts, and the image is specific to cement.

## What was run

- Project: a local clone of cement at `3e16e4e`, with a copy of its D28 verdict store (100 verdicts).
- Image: `localhost/cement-gate:proto`, built from `Containerfile`. It contains Debian trixie-slim, uv 0.13.1, Python 3.13.16 (`3.13.16 (main, Oct 14 2026, 01:04:01) [Clang 23.1.3]`) and git. The base image and uv are pinned by digest.
- Hosts: the controller (t25-480) and nixbox.
- Commands: `prototype/container/grade <step>`. The logs of each run are in `.scratch/logs/` (not committed).

## Result 1: D28 verdicts travel between hosts in the image

`grade hits` counts the D28 range revisions that have a green verdict for the current interpreter. It uses cement's own key and lookup functions and does not replay.
The range has 73 revisions.

| Host | Environment | Python build | Hits |
|------|-------------|--------------|------|
| controller | native | 3.13.14, Jul 23 2026 | 73/73 |
| nixbox | native | 3.13.14, Jun 11 2026 | 0/73 |
| controller | image | 3.13.16, Oct 14 2026 | 1/73 |
| nixbox | image | 3.13.16, Oct 14 2026 | 1/73 |

- The single image verdict came from `grade one`: one cold D28 replay of `e325bdc` in the image on the controller (1337 tests, 170 s; a second run took 130 s).
- After `offload` copied the store to nixbox, the same verdict hit there. Natively, nixbox hits none of the 73 controller verdicts.
- The verdict also hit in a rebuilt image (an extra apt package, and later precompiled bytecode). A rebuild keeps the verdicts while the Python bytes stay the same.
- Not run (your ruling): a cold fill of all 73 revisions in the image (about 3 to 6 hours), and then a full gate on nixbox that prints `0 replayed`. Result 1 shows that the keys are equal. It does not show a full D28 run in the image. See `.agent/deferred.md`.

## Result 2: the gate passes in the image on both hosts

`grade gate <host> <env>` runs `.agent/gate --omit-d28` (every step except D28). All 11 runs exited with 0.
The table shows the `unittest` step in seconds, because all the other steps together take 16 seconds or less.
"Busy" is the host's load at the start of the run.

| Run | Host | Image | Native |
|-----|------|-------|--------|
| before the bytecode fix | nixbox | 466 (96 % busy) | 632 (37 % busy) |
| before the bytecode fix | nixbox | 291 (15 % busy) | 446 (14 % busy) |
| before the bytecode fix | controller | 455 | 654 (other gates running) |
| after the bytecode fix | controller | 338 | 330 |
| after the bytecode fix | nixbox | 282 (85 % busy) | 455 (15 % busy) |
| after the bytecode fix | nixbox | 275 (23 % busy) | 467 (13 % busy) |
| ITERATE rerun, both hosts at once | nixbox | 260 (0 % busy) | not run |
| ITERATE rerun, both hosts at once | controller | 322 | not run |

- Controller: the image adds about 2 % (one pair).
- nixbox: the image is about 40 % faster than native in all four pairs. The cause is not known. Process start, `/usr/bin/env` through envfs, git commits, CPU-bound Python and SQLite all measure about the same in both environments. See `.agent/deferred.md`.
- Bytecode fix: the first image had no precompiled stdlib bytecode, and the container user cannot write it. Each Python process then started 4 times slower than native. The image now precompiles at build time (60 starts: 1.96 s in the image, 1.85 s native).

## Result 3: upkeep per image bump

| Change | Build | `ship save` | `ship registry` |
|--------|-------|-------------|-----------------|
| first ship | 75 s | 112.6 MB, 12.6 s | 127.6 MB, 9.8 s |
| no change | none | 112.6 MB (the full image every time) | 2.2 kB, 2.9 s |
| one apt package added (cold) | 75 s | 113.6 MB, 9.7 s | 96.8 MB, 13.1 s |
| bytecode added to the Python layer | 14 s | not measured | 43.5 MB, 6.3 s |

- `ship save` sends the whole image each time (`podman save | zstd | ssh podman load`). It needs nothing on the controller except podman.
- `ship registry` sends only the layers that the host does not have. It needs a registry container on the controller (`ctr-registry`, port 5005 on localhost). nixbox reaches it through an SSH reverse tunnel, so no firewall changes are necessary.
- Both methods give the same image id on both hosts.
- Before Result 4, a rebuild sent the Python layer (about 43 MB) again even when Python did not change, because the layer was not byte-reproducible. The reproducible build fixes this.
- A change of the Python version in the image makes all D28 verdicts miss. Then one host must replay all 73 revisions once (about 2 to 5 minutes each).

## Result 4: each host can build the image itself, with identical layers

nixbox built the image from the same `Containerfile`, and nothing was shipped. The first build took 25 s.

- Without changes to the `Containerfile`, the Python binary was byte-identical on both hosts and `grade hits` gave 1/73. But only the base layer was identical. The apt layer differed because `apt-get update` takes the current Debian archive. The Python layer differed because each `.pyc` file records the time when uv unpacked the stdlib.
- The `Containerfile` now makes the build reproducible:
  - The Debian archive is pinned to a `snapshot.debian.org` date.
  - The apt logs and caches are removed.
  - The bytecode is compiled with `--invalidation-mode unchecked-hash`, `-f` and `PYTHONHASHSEED=0`.
  - The image is built with `podman build --timestamp 0`.
- Result: all 5 layers have the same digest on both hosts, although the hosts use different podman versions (6.1.4 and 5.8.8). The image id is still different, because the history records the local tag of the base image. To compare two hosts, use the layer digests (`podman image inspect --format '{{.RootFS.Layers}}'`).
- The reproducible image still gives 1/73 hits, and 60 Python starts take 0.74 s.
- Each fix was necessary. These are the numbers of different files in the Python layer: 1045 with only the snapshot pin and `--timestamp 0`, 59 with `unchecked-hash` added, 22 with `PYTHONHASHSEED=0` added, and 0 with `-f` added.

## Kill behaviour

- When `offload` is killed with SIGKILL during a run on nixbox, the container stops within 5 seconds. The remote watcher sends TERM, and `podman` forwards it to the container.
- When `ctr` itself gets TERM, the container stops.
- When `ctr` itself gets SIGKILL without a TERM first, the container keeps running, because podman's conmon process is outside the killed session. See `.agent/deferred.md`.

## Decisions for ITERATE

- Where does each project keep its image definition? This spike keeps cement's `Containerfile` in the offload repository. A project repository is the natural place.
- How does a project select the image? The options are a wrapper in the gate command (this spike) or an `offload run --image` option.
- How does the image get to each host? `registry` sends less after the first ship. `save` has fewer moving parts. A build on each host (Result 4) sends nothing, and every host gets identical layers.
- Is a full cold D28 fill in the image necessary before you decide?
