package backup

import (
	"context"
	"github.com/local/telecom/internal/database"
	"path/filepath"
	"testing"
)

func TestInspect(t *testing.T) {
	dir := t.TempDir()
	db, e := database.Open(filepath.Join(dir, "telecom.sqlite"))
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	if e = database.Migrate(db); e != nil {
		t.Fatal(e)
	}
	path := filepath.Join(dir, "backup.zip")
	if e = New(db).Create(context.Background(), path, ""); e != nil {
		t.Fatal(e)
	}
	inspection, e := Inspect(path)
	if e != nil || !inspection.HasDatabase {
		t.Fatalf("%#v %v", inspection, e)
	}
}
