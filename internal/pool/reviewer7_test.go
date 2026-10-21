package pool

import "testing"

func TestP2OneSlotAndDotName(t *testing.T) {
	t.Run("one-slot", func(t *testing.T) {
		hosts, err := Load(write(t, "[[host]]\nname = \"a\"\naddress = \"a\"\nslots = 1\n"))
		if err != nil || len(hosts) != 1 || hosts[0].Slots != 1 {
			t.Fatalf("P2: slots=1 → %+v, %v", hosts, err)
		}
	})
	t.Run("dot-name", func(t *testing.T) {
		if hosts, err := Load(write(t, "[[host]]\nname = \".\"\naddress = \"a\"\n")); err == nil {
			t.Fatalf("P2: dot name accepted: %+v", hosts)
		}
	})
}
