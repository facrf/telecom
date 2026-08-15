package config

import "testing"

func TestHTTPAddress(t *testing.T) {
	cfg := Config{Port: 8080}
	if got := cfg.HTTPAddress(); got != ":8080" {
		t.Fatalf("HTTPAddress() = %q", got)
	}
}
