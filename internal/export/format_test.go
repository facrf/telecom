package export

import "testing"

func TestDocumentValidation(t *testing.T) {
	document := NewDocument()
	if err := Validate(document); err != nil {
		t.Fatal(err)
	}
	document.SchemaVersion = 2
	if err := Validate(document); err == nil {
		t.Fatal("invalid version accepted")
	}
}
