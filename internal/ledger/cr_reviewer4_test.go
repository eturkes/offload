package ledger

import (
	"bytes"
	"os"
	"testing"
)

func TestR4J2EndRecordAndExitShareFinalStatus(t *testing.T) {
	for _, lateSignal := range []bool{false, true} {
		name := "no-signal"
		if lateSignal {
			name = "signal-after-record-built"
		}
		t.Run(name, func(t *testing.T) {
			oldOpen := openLog
			t.Cleanup(func() { openLog = oldOpen })
			r := Begin(env(map[string]string{"XDG_STATE_HOME": t.TempDir()}), "/w", []string{"true"}, Options{}, &bytes.Buffer{})
			rc, sig := 0, ""
			// J2: rc + signal freeze together when the end write holds its lock. openLog
			// follows record construction, so this seam models a signal during the write.
			openLog = func(path string) (*os.File, error) {
				if lateSignal {
					rc, sig = 143, "TERM"
				}
				return oldOpen(path)
			}
			got := r.End(func() (int, string) { return rc, sig })
			records := lines(t, r.path)
			if len(records) != 1 {
				t.Fatalf("records = %d, want one end", len(records))
			}
			rec := records[0]
			if rec["rc"] != float64(got) {
				t.Fatalf("J2 split final status: end record rc=%v signal=%v, End returned %d", rec["rc"], rec["signal"], got)
			}
		})
	}
}
