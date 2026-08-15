package scanner

import "testing"

func TestParsePorts(t *testing.T) {
	ports, e := ParsePorts("80, 443,8000-8002", 10)
	if e != nil {
		t.Fatal(e)
	}
	if len(ports) != 5 || ports[0] != 80 || ports[4] != 8002 {
		t.Fatalf("ports=%v", ports)
	}
	if _, e := ParsePorts("0", 10); e == nil {
		t.Fatal("zero accepted")
	}
	if _, e := ParsePorts("1-65535", 10); e == nil {
		t.Fatal("limit ignored")
	}
}
