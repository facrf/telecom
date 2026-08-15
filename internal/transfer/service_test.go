package transfer

import (
	"context"
	"github.com/local/telecom/internal/database"
	"path/filepath"
	"testing"
)

func TestExportImportProject(t *testing.T) {
	ctx := context.Background()
	source, err := database.Open(filepath.Join(t.TempDir(), "source.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	if err = database.Migrate(source); err != nil {
		t.Fatal(err)
	}
	if _, err = source.Exec("INSERT INTO clients(id,name)VALUES('c','Cliente');INSERT INTO projects(id,client_id,name)VALUES('p','c','Matriz');INSERT INTO devices(id,project_id,name,category_id)VALUES('d','p','Switch','switch')"); err != nil {
		t.Fatal(err)
	}
	document, err := ExportProject(ctx, source, "p")
	if err != nil {
		t.Fatal(err)
	}
	target, err := database.Open(filepath.Join(t.TempDir(), "target.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()
	if err = database.Migrate(target); err != nil {
		t.Fatal(err)
	}
	if err = ImportProject(ctx, target, document); err != nil {
		t.Fatal(err)
	}
	var count int
	if err = target.QueryRow("SELECT count(*) FROM devices").Scan(&count); err != nil || count != 1 {
		t.Fatalf("count=%d error=%v", count, err)
	}
}
