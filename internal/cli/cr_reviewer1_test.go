package cli

import (
	"slices"
	"strings"
	"testing"
)

// K10: unframed output must not hide subsequent framed layer digests.
func TestReviewR1K10LongNoiseCannotTruncateLayers(t *testing.T) {
	const nonce = "review-nonce"
	out := "offload-check " + nonce + " same-binary\n" + strings.Repeat("x", 65536) + "\nlayer " + nonce + " sha256:different\n"
	sum, layers, ok := parseCheck(out, nonce)
	if !ok || sum != "same-binary" || !slices.Equal(layers, []string{"sha256:different"}) {
		t.Fatalf("parseCheck accepted=%v sum=%q layers=%q; long unframed noise hid the differing layer (an empty local image now compares same)", ok, sum, layers)
	}
}
