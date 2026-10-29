# Pool run environment — PROTOTYPE/ITERATE record

Prototype retired at IMPLEMENT close → tag `prototype-pool-run-environment` (`prototype/container/`: cement gate `Containerfile`, `ctr`, `build-host`, `grade`, `d28-hits`, `d28-one`, `results.md` = measurements).

Artifacts entry (retired):
- `prototype/container/` = pool run environment spike (PROTOTYPE law; cement-specific): `Containerfile` (gate image, reproducible: identical layers on every host; build `podman build --timestamp 0 -t localhost/cement-gate:proto prototype/container`), `ctr <cmd…>` (runs it in the image, git root bind-mounted at its own path, venv + uv cache in `~/.cache/ctr/`), `build-host` (ruled distribution: builds the image on the controller + nixbox from the same Containerfile, installs `ctr` at nixbox `~/.local/share/offload-proto/ctr`, compares layer digests), `d28-hits` + `d28-one` (cement D28 key probes), results `results.md`. Run: `prototype/container/grade clone|hits|one|gate <local|nixbox> <native|ctr>|bump` (no args = usage); working state = `.scratch/cement` (clone holding the image verdict) + `.scratch/logs/`.

Decision text superseded by the implementation (`offload ctr`):
- Pool run environment (owner): cross-host parity moves from per-tool nixbox additions to a replicated environment; containerized dev environments = good practice + an orthogonal portability test. PROTOTYPE spike = per-project image alone (Nix two-sided pin dropped): the gate command wraps `podman run` (rootless both sides; offload unchanged). Grading check (owner: a cold fill of all 73 D28 revisions in the image ≈ 3–6 h is not needed to confirm the system ⇒ key equality + one replay suffice): in a local clone of cement at its HEAD (verdict store copied; keys path-independent; the live cement root untouched) — (1) `d28-hits` (cement's own `_d28_content_key` + `_d28_verdict`, no replay) on both hosts; (2) one cold D28 revision replayed in the image on the controller (`d28-one`) → its verdict hits on nixbox (`d28-hits` 0 → 1 there; native today 0: nixbox uv 0.11.21 vs 0.13.1 ⇒ different Python builds); (3) `.agent/gate --omit-d28` in the image on both hosts green + wall-clock vs native per host; (4) upkeep per image bump (steps, seconds, bytes shipped). Distribution measured: `podman save | ssh podman load`, a controller-local registry over an ssh reverse tunnel, a per-host build (the ruled one, above).

IMPLEMENT (closed → MAINTAIN — whole product), Tasks:
- [x] 0d14378 Spikes: container HOME under `--userns=keep-id`; kill guarantee through podman (in-container init on a liveness pipe vs `--cgroups=split`); offload install path on nixbox
- [x] 0d14378 Contract: `offload ctr` predicates (tier kernel) in `.agent/contract.md` + `consultant` review before code
- [x] 0d14378 `offload ctr` run path: switch, definition lookup, content-hash tag, build on miss, mounts, exit status, kill guarantee — red-first tests (fake podman + real podman)
- [x] 0d14378 Layer check across hosts (ruled sync check: equal `RootFS.Layers` digest lists)
- [x] be96752 nixbox: offload installed + switch set + deploy; live runs on both hosts
- [x] b78f436 Agents repo (agents repo commits b78f436 + f40dfde): `CLAUDE.project.md` CI gate line + `CLAUDE.local.md` tier; cachyos `CLAUDE.md` pool line + `offload ctr` usage + switch in settings `env`
- [x] 57903b5 README + prototype retired (`prototype-pool-run-environment` tag) + deferred rows
- [x] 81a86d5 Closing review per lens + security (MAIN) + mutation credit + re-review round → `.agent/review.md` (commits 81a86d5, 6451106, 7a4a2d0, d8503a0, fa6936e, 1d93774)
- Deferred queue → `.agent/deferred.md`


Routed request (Phase lines at close):
IMPLEMENT — pool run environment + CI gate compute
> At this point don't you think it might be a good idea to go containerized with our workloads so we can replicate them across machines?
> Barring that, an advantage of NixOS is the ease at which we can pin multiple versions I thought?
> Sounds good. I think containerized dev environments are good practice anyway. This also gives us an orthogonal way to test portability.
> If we already have CI, then why are we running these long gates on my own machines anyway?
> […] my project gates are very long. It is by far the bottleneck I am facing in development right now.
