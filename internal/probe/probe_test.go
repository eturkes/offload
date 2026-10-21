package probe

import (
	"os/exec"
	"testing"
)

func TestParse(t *testing.T) { // R1
	s, err := Parse("8\n0.22\n64216840\n0\n")
	if err != nil || s != (Stats{CPUs: 8, Load: 0.22, MemKB: 64216840}) {
		t.Fatalf("Parse = %+v, %v", s, err)
	}
	if b := s.Busy(); b < 0.0274 || b > 0.0276 {
		t.Fatalf("Busy = %v", b)
	}
	for _, bad := range []string{"", "8\n0.2\n1\n", "8\n0.2\n1\n0\nextra\n", "x\n0.2\n1\n0\n", "0\n0.2\n1\n0\n", "8\nnan?\n1\n0\n", "8\nNaN\n1\n0\n", "8\n+Inf\n1\n0\n", "8\n1e999\n1\n0\n", "8\n-1\n1\n0\n", "8\n0.2\n-5\n0\n", "8\n0.2\n1\n-1\n", "Welcome!\n8\n0.2\n1\n0\n"} {
		if _, err := Parse(bad); err == nil {
			t.Errorf("Parse(%q) accepted", bad)
		}
	}
}

func TestScriptOnThisMachine(t *testing.T) { // R1: the script's own output parses here
	out, err := exec.Command("bash", "-c", Script).Output()
	if err != nil {
		t.Fatal(err)
	}
	s, err := Parse(string(out))
	if err != nil || s.CPUs < 1 {
		t.Fatalf("Parse(%q) = %+v, %v", out, s, err)
	}
}
