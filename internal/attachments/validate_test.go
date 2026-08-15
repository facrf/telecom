package attachments

import (
	"testing"
)

func TestValidate(t *testing.T) {
	data := []byte("plain text")
	meta, e := Validate("nota.txt", data, 1024)
	if e != nil || meta.StoredFilename == "" {
		t.Fatalf("%#v %v", meta, e)
	}
	if _, e = Validate("../segredo.txt", data, 1024); e == nil {
		t.Fatal("path accepted")
	}
}
