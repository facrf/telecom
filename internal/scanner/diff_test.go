package scanner

import "testing"

func TestDiff(t *testing.T) {
	old := []SnapshotHost{{IP: "192.168.1.2", MAC: "aa", Hostname: "switch", Ports: map[int]string{80: "HTTP"}}}
	current := []SnapshotHost{{IP: "192.168.1.3", MAC: "aa", Hostname: "switch", Ports: map[int]string{80: "HTTP", 443: "HTTPS"}}, {IP: "192.168.1.4", MAC: "bb"}}
	changes := Diff(old, current)
	if len(changes) != 3 {
		t.Fatalf("changes=%#v", changes)
	}
}
