package importer

import "testing"

func TestValidateRejectsIncompleteDocument(t *testing.T) {
	if _, err := Validate([]byte(`{"format":"telecom","schema_version":1}`)); err == nil {
		t.Fatal("incomplete document accepted")
	}
}
