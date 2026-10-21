---
paths:
  - ".agent/gate"
  - ".agent/tool"
  - ".golangci.yml"
  - ".github/**"
---
# Gate firing records

Each step fires on a planted bad input (rerun the seed after editing its check):
- fmt: `func  seed( ) int {return 1}` in `internal/cli/zz_seed.go` → rc 1.
- lint: unchecked `os.Remove("x")` → rc 1 (errcheck).
- tidy: `go mod edit -require=golang.org/x/text@v0.3.5` → `go mod tidy -diff` rc 1.
- test: `t.Fatal` in a new test → rc 1.
- vuln: scratch module requiring `golang.org/x/text v0.3.5` + calling `language.Parse` → rc 3 (GO-2021-0113).
- secrets: scratch repo committing a `ghp_` + 36 alnum token → Betterleaks rc 1.
- actions: workflow step with `shell: bsh` → actionlint rc 1; same workflow (unpinned `actions/checkout@v4`, `${{ github.event.issue.title }}` in `run`) → zizmor rc 14.
- shell: `cd $1` in a `#!/bin/sh` file → shellcheck `-S warning` rc 1.
- nix: the six pre-nixfmt `machines/nixbox/*.nix` files → rc 1.
- spec-lint: a history line appended under `## Phase` → rc 1.
- `.agent/tool`: a pin with a wrong sha256 → refuses the asset, rc 2.
- nixfmt adoption: every tracked `*.nix` parses (`nix-instantiate --parse`, both versions from one directory, on nixbox) to the same AST before + after formatting; control = a one-byte value change → different AST.
