package attachments

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStore(t *testing.T) {
	dir := t.TempDir()
	metadata, hash, e := NewStore(dir, 1024).Save("device", "nota.txt", []byte("test"))
	if e != nil || hash == "" {
		t.Fatalf("%#v %s %v", metadata, hash, e)
	}
	if _, e = os.Stat(filepath.Join(dir, "attachments", "device", metadata.StoredFilename)); e != nil {
		t.Fatal(e)
	}
}

func TestStoreAcceptsTechnicalVisitEvidence(t *testing.T) {
	dir := t.TempDir()
	metadata, _, err := NewStore(dir, 1024).Save("technical_visit", "antes.txt", []byte("evidencia"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(filepath.Join(dir, "attachments", "technical_visit", metadata.StoredFilename)); err != nil {
		t.Fatal(err)
	}
}
