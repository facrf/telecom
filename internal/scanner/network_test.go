package scanner

import "testing"

func TestParseTarget(t *testing.T) {
	tests := []struct {
		input string
		count uint64
		valid bool
	}{{"192.168.1.0/24", 256, true}, {"10.0.0.2-10.0.0.5", 4, true}, {"8.8.8.0/24", 0, false}, {"192.168.1.2-192.168.1.1", 0, false}}
	for _, test := range tests {
		target, err := ParseTarget(test.input, 512, false)
		if test.valid && err != nil {
			t.Fatalf("%s: %v", test.input, err)
		}
		if !test.valid && err == nil {
			t.Fatalf("%s should fail", test.input)
		}
		if test.valid && target.Count != test.count {
			t.Fatalf("%s count=%d", test.input, target.Count)
		}
	}
}
