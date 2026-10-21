package probe

import (
	"strings"
	"testing"
)

func TestR1ParseBoundaries(t *testing.T) {
	for _, tc := range []struct {
		out  string
		want Stats
	}{
		{"1 0 0 0", Stats{CPUs: 1}},
		{"1 0.5 4611686018427387904 0", Stats{CPUs: 1, Load: 0.5, MemKB: 1 << 62}},
	} {
		if got, err := Parse(tc.out); err != nil || got != tc.want {
			t.Errorf("R1: Parse(%q) = %+v, %v; want %+v", tc.out, got, err, tc.want)
		}
	}
}

func TestR1MalformedCounters(t *testing.T) {
	for _, out := range []string{
		"18446744073709551616 0 1 0",
		"1 0 not-memory 0",
		"1 0 18446744073709551616 0",
		"1 0 -1 0",
		"1 0 1 not-gpus",
		"1 0 1 18446744073709551616",
	} {
		if got, err := Parse(out); err == nil {
			t.Errorf("R1: malformed counter accepted: Parse(%q) = %+v", out, got)
		}
	}
}

func TestR1MalformedReasonNamesRejectedValue(t *testing.T) {
	for _, tc := range []struct{ out, bad string }{
		{"not-cpus 0.5 123 0", "not-cpus"},
		{"8 not-load 123 0", "not-load"},
		{"8 0.5 not-memory 0", "not-memory"},
		{"8 0.5 123 not-gpus", "not-gpus"},
	} {
		if _, err := Parse(tc.out); err == nil || !strings.Contains(err.Error(), tc.bad) {
			t.Errorf("R1: reason for %q must identify rejected value %q: %v", tc.out, tc.bad, err)
		}
	}
}
