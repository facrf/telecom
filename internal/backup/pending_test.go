package backup

import (
	"bytes"
	"context"
	"github.com/local/telecom/internal/database"
	"os"
	"path/filepath"
	"testing"
)

func TestQueueAndApplyPending(t *testing.T) {
	sourceDir := t.TempDir()
	source, err := database.Open(filepath.Join(sourceDir, "telecom.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	if err = database.Migrate(source); err != nil {
		t.Fatal(err)
	}
	if _, err = source.Exec("INSERT INTO settings(key,value)VALUES('restored','yes')"); err != nil {
		t.Fatal(err)
	}
	backupPath := filepath.Join(sourceDir, "backup.zip")
	if err = New(source).Create(context.Background(), backupPath, ""); err != nil {
		t.Fatal(err)
	}
	source.Close()
	data, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatal(err)
	}
	targetDir := t.TempDir()
	if _, err = QueueRestore(targetDir, bytes.NewReader(data)); err != nil {
		t.Fatal(err)
	}
	applied, err := ApplyPending(targetDir)
	if err != nil || !applied {
		t.Fatalf("applied=%v error=%v", applied, err)
	}
	target, err := database.Open(filepath.Join(targetDir, "telecom.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()
	var value string
	if err = target.QueryRow("SELECT value FROM settings WHERE key='restored'").Scan(&value); err != nil || value != "yes" {
		t.Fatalf("value=%q error=%v", value, err)
	}
}
