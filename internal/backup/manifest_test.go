package backup

import (
	"encoding/json"
	"testing"
)

func TestManifest(t *testing.T) {
	data, e := json.Marshal(NewManifest())
	if e != nil {
		t.Fatal(e)
	}
	if _, e := ValidateManifest(data); e != nil {
		t.Fatal(e)
	}
}
