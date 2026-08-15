package scanner

import (
	"net/netip"
	"testing"
)

func TestParseTarget(t *testing.T) {
	tests := []struct {
		input string
		start string
		end   string
		count uint64
		valid bool
	}{
		{"192.168.1.0/24", "192.168.1.0", "192.168.1.255", 256, true},
		{"10.0.0.2-10.0.0.5", "10.0.0.2", "10.0.0.5", 4, true},
		{"192.168.1.50", "192.168.1.50", "192.168.1.50", 1, true},
		{"8.8.8.0/24", "", "", 0, false},
		{"192.168.1.2-192.168.1.1", "", "", 0, false},
	}
	for _, test := range tests {
		target, err := ParseTarget(test.input, 512, false)
		if test.valid && err != nil {
			t.Fatalf("%s: %v", test.input, err)
		}
		if !test.valid && err == nil {
			t.Fatalf("%s should fail", test.input)
		}
		if test.valid {
			if target.Count != test.count {
				t.Fatalf("%s count=%d want %d", test.input, target.Count, test.count)
			}
			if target.Start != netip.MustParseAddr(test.start) {
				t.Fatalf("%s start=%s want %s", test.input, target.Start, test.start)
			}
			if target.End != netip.MustParseAddr(test.end) {
				t.Fatalf("%s end=%s want %s", test.input, target.End, test.end)
			}
		}
	}
}
