---
paths:
  - ".agent/gate"
  - ".agent/tool"
  - ".agent/pins"
  - ".agent/law"
  - ".golangci.yml"
  - "statix.toml"
  - ".github/**"
---
# Gate firing records

Each step fires on a planted bad input (rerun the seed after editing its check):
- fmt: `func  seed( ) int {return 1}` in `internal/cli/zz_seed.go` → rc 1.
- lint: unchecked `os.Remove("x")` → rc 1 (errcheck).
- tidy: `go mod edit -require=golang.org/x/text@v0.3.5` → `go mod tidy -diff` rc 1.
- test: `t.Fatal` in a new test → rc 1.
- build: a `//go:build !cgo` file with a type error → static build rc 1 while vet + the race tests (cgo on) stay green, so only this step catches it.
- vuln: scratch module requiring `golang.org/x/text v0.3.5` + calling `language.Parse` → rc 3 (GO-2021-0113).
- secrets: scratch repo committing a `ghp_` + 36 alnum token → Betterleaks rc 1.
- actions: workflow step with `shell: bsh` → actionlint rc 1; same workflow (unpinned `actions/checkout@v4`, `${{ github.event.issue.title }}` in `run`) → zizmor rc 14.
- shell: `cd $1` in a `#!/bin/sh` file → shellcheck `-S warning` rc 1.
- nix: a tracked `*.nix` file holding `{ a=1; }` on one line → `nixfmt --check` rc 1.
- nixlint: a tracked `*.nix` file `{ ... }:\n{ }` → statix W10 empty pattern, rc 1; one holding `let\n  unused = 1;\nin\n{ }` → deadnix rc 1.
- law: `.claude/rules/zz.md` = `---`/`paths:`/`  - machines/**`/`---` (unquoted glob) → rc 1; `simply` appended to README.md → rc 1; `- seed` appended to `.agent/deferred.md` (no `Accept:`) → rc 1; `internal/cli/zz_seed_test.go` with `func TestSeed(t *testing.T) {}` (untracked counts) → rc 1.
- spec-lint: a history line appended under `## Phase` → rc 1.
- `.agent/tool`: a pin with a wrong sha256 → refuses the asset, rc 2.
- `.agent/pins` (weekly `pins.yml`, not a gate step): a pin behind its latest release → `behind`, rc 1 (a planted `lanzaboote/archive/v1.1.0` in `secureboot.nix` → `lanzaboote 1.1.0 1.2.0 behind`; an untagged archive URL → `no lanzaboote tag`, rc 2); a workflow `go-version` ≠ `go.mod` → rc 1; a planted `runs-on: ubuntu-22.04` → `behind`, rc 1; an unknown source repo → `lookup failed`, rc 2.
