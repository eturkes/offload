package e2e_test

import (
	"fmt"
	"testing"
)

func TestReviewY10OutputListFailureOverridesCommand(t *testing.T) {
	t.Parallel()
	for _, code := range []int{0, 42, 255} {
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			s := newSandbox(t, true)
			command := fmt.Sprintf("printf command-ran; printf broken > .git/index; printf output > output; exit %d", code)
			r := s.run("run", "--pool", "--", "bash", "-c", command)
			if r.stdout != "command-ran" {
				t.Fatalf("fixture command did not run: %+v", r)
			}
			requireContains(t, r.stderr, "index file smaller than expected")
			r.requireCode(t, 125)
		})
	}
}
