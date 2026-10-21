// Package pool loads the pool file: the SSH hosts offload may run commands on.
package pool

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// DefaultSlots = concurrent offload runs per host when the pool file names none (heavy's default).
const DefaultSlots = 2

// Host is one [[host]] table.
type Host struct {
	Name      string `toml:"name"`
	Address   string `toml:"address"`
	Slots     int    `toml:"slots"`
	Multiplex bool   `toml:"multiplex"`
}

// rawHost tells an absent `slots` (default) from an explicit 0 (refused).
type rawHost struct {
	Name      string `toml:"name"`
	Address   string `toml:"address"`
	Slots     *int   `toml:"slots"`
	Multiplex bool   `toml:"multiplex"`
}

type file struct {
	Host []rawHost `toml:"host"`
}

// ErrNotExist marks a missing pool file (run → local slots only; status/--pool → error).
var ErrNotExist = errors.New("no pool file")

// Name doubles as a lock-directory component ⇒ no separators, no dot-only names.
var validName = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// Address feeds rsync's `host:path` syntax ⇒ no `:` or `/` (ports + IPv6 → an ssh_config alias), and
// ssh's argv ⇒ no leading `-`.
var validAddress = regexp.MustCompile(`^[^\s:/@-][^\s:/@]*(@[^\s:/@-][^\s:/@]*)?$`)

// Path resolves the pool file: $OFFLOAD_CONFIG, else $XDG_CONFIG_HOME/offload/pool.toml, else
// ~/.config/offload/pool.toml.
func Path(getenv func(string) string) string {
	if p := getenv("OFFLOAD_CONFIG"); p != "" {
		return p
	}
	if d := getenv("XDG_CONFIG_HOME"); d != "" {
		return filepath.Join(d, "offload", "pool.toml")
	}
	return filepath.Join(getenv("HOME"), ".config", "offload", "pool.toml")
}

// Load reads and validates the pool file at path.
func Load(path string) ([]Host, error) {
	data, err := os.ReadFile(path) //nolint:gosec // G304: path = the operator's own pool file (env/XDG), trusted
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("%w at %s", ErrNotExist, path)
	}
	if err != nil {
		return nil, err
	}
	var f file
	dec := toml.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&f); err != nil {
		var strict *toml.StrictMissingError
		if errors.As(err, &strict) {
			keys := make([]string, len(strict.Errors))
			for i, e := range strict.Errors {
				keys[i] = strings.Join(e.Key(), ".")
			}
			return nil, fmt.Errorf("%s: unknown key %s", path, strings.Join(keys, ", "))
		}
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if len(f.Host) == 0 {
		return nil, fmt.Errorf("%s lists no [[host]] tables", path)
	}
	seen, seenAddr := map[string]bool{}, map[string]bool{}
	hosts := make([]Host, len(f.Host))
	for i, r := range f.Host {
		h := &hosts[i]
		h.Name, h.Address, h.Slots, h.Multiplex = r.Name, r.Address, DefaultSlots, r.Multiplex
		if r.Slots != nil {
			h.Slots = *r.Slots
		}
		switch {
		case !validName.MatchString(h.Name) || h.Name == "." || h.Name == "..":
			return nil, fmt.Errorf("%s: host %d: name %q must match [A-Za-z0-9._-]+", path, i+1, h.Name)
		case seen[h.Name]:
			return nil, fmt.Errorf("%s: host name %q appears twice", path, h.Name)
		case !validAddress.MatchString(h.Address):
			return nil, fmt.Errorf("%s: host %s: address %q must be [user@]host or an ssh_config alias", path, h.Name, h.Address)
		case seenAddr[h.Address]:
			return nil, fmt.Errorf("%s: address %q appears twice", path, h.Address)
		case h.Slots < 1:
			return nil, fmt.Errorf("%s: host %s: slots must be at least 1", path, h.Name)
		}
		seen[h.Name], seenAddr[h.Address] = true, true
	}
	return hosts, nil
}
