package remote

import (
	"os/exec"
	"testing"
)

func TestJoinRoundTripsThroughBash(t *testing.T) {
	words := []string{"plain", "", "a b", "it's", `"dq"`, "$HOME", "`x`", "a\nb", "-n", "*", "~", "\\", "é", "x;y|z&"}
	out, err := exec.Command("bash", "-c", "printf '%s\\0' "+Join(words)).Output()
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	start := 0
	for i, b := range out {
		if b == 0 {
			got = append(got, string(out[start:i]))
			start = i + 1
		}
	}
	if len(got) != len(words) {
		t.Fatalf("got %d words %q", len(got), got)
	}
	for i := range words {
		if got[i] != words[i] {
			t.Errorf("word %d: got %q want %q", i, got[i], words[i])
		}
	}
}
