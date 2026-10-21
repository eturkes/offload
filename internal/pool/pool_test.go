package pool

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func env(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

func TestPathPrecedence(t *testing.T) { // P1
	cases := []struct {
		env  map[string]string
		want string
	}{
		{map[string]string{"OFFLOAD_CONFIG": "/x/p.toml", "XDG_CONFIG_HOME": "/c", "HOME": "/h"}, "/x/p.toml"},
		{map[string]string{"XDG_CONFIG_HOME": "/c", "HOME": "/h"}, "/c/offload/pool.toml"},
		{map[string]string{"HOME": "/h"}, "/h/.config/offload/pool.toml"},
	}
	for _, c := range cases {
		if got := Path(env(c.env)); got != c.want {
			t.Errorf("Path(%v) = %q, want %q", c.env, got, c.want)
		}
	}
}

func write(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "pool.toml")
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLoadDefaultsSlots(t *testing.T) { // P2
	hosts, err := Load(write(t, "[[host]]\nname = \"a\"\naddress = \"u@10.0.0.1\"\n\n[[host]]\nname = \"b\"\naddress = \"b\"\nslots = 5\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(hosts) != 2 || hosts[0].Slots != 2 || hosts[1].Slots != 5 || hosts[0].Address != "u@10.0.0.1" {
		t.Fatalf("hosts = %+v", hosts)
	}
}

func TestLoadMultiplex(t *testing.T) { // P2: multiplex = bool, default false
	hosts, err := Load(write(t, "[[host]]\nname = \"a\"\naddress = \"a\"\n\n[[host]]\nname = \"b\"\naddress = \"b\"\nmultiplex = true\n"))
	if err != nil || len(hosts) != 2 || hosts[0].Multiplex || !hosts[1].Multiplex {
		t.Fatalf("hosts = %+v, err %v", hosts, err)
	}
}

func TestLoadRejects(t *testing.T) { // P2
	cases := map[string]struct{ body, want string }{
		"unknown host key":  {"[[host]]\nname = \"a\"\naddress = \"a\"\nslot = 3\n", "slot"},
		"unknown top key":   {"hosts = 1\n[[host]]\nname = \"a\"\naddress = \"a\"\n", "hosts"},
		"no hosts":          {"", "no [[host]]"},
		"missing name":      {"[[host]]\naddress = \"a\"\n", "name"},
		"name with slash":   {"[[host]]\nname = \"a/b\"\naddress = \"a\"\n", "name"},
		"dot-dot name":      {"[[host]]\nname = \"..\"\naddress = \"a\"\n", "name"},
		"duplicate name":    {"[[host]]\nname = \"a\"\naddress = \"a\"\n[[host]]\nname = \"a\"\naddress = \"b\"\n", "twice"},
		"missing address":   {"[[host]]\nname = \"a\"\n", "address"},
		"option as address": {"[[host]]\nname = \"a\"\naddress = \"-oProxyCommand=x\"\n", "address"},
		"user option":       {"[[host]]\nname = \"a\"\naddress = \"u@-oX\"\n", "address"},
		"uri address":       {"[[host]]\nname = \"a\"\naddress = \"ssh://u@h:22\"\n", "address"},
		"port in address":   {"[[host]]\nname = \"a\"\naddress = \"h:2222\"\n", "address"},
		"space in address":  {"[[host]]\nname = \"a\"\naddress = \"h x\"\n", "address"},
		"duplicate address": {"[[host]]\nname = \"a\"\naddress = \"h\"\n[[host]]\nname = \"b\"\naddress = \"h\"\n", "twice"},
		"negative slots":    {"[[host]]\nname = \"a\"\naddress = \"a\"\nslots = -1\n", "slots"},
		"zero slots":        {"[[host]]\nname = \"a\"\naddress = \"a\"\nslots = 0\n", "slots"},
		"bad toml":          {"[[host]\n", "pool.toml"},
		"string multiplex":  {"[[host]]\nname = \"a\"\naddress = \"a\"\nmultiplex = \"yes\"\n", "pool.toml"},
	}
	for name, c := range cases {
		_, err := Load(write(t, c.body))
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: err = %v, want mention of %q", name, err, c.want)
		}
	}
}

func TestLoadMissingFile(t *testing.T) { // P3
	_, err := Load(filepath.Join(t.TempDir(), "absent.toml"))
	if !errors.Is(err, ErrNotExist) {
		t.Fatalf("err = %v, want ErrNotExist", err)
	}
}
