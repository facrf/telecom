package backup

import (
	"context"
	"github.com/local/telecom/internal/database"
	"path/filepath"
	"testing"
)

func TestCreate(t *testing.T) {
	dir := t.TempDir()
	db, e := database.Open(filepath.Join(dir, "telecom.sqlite"))
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	if e = database.Migrate(db); e != nil {
		t.Fatal(e)
	}
	if e = New(db).Create(context.Background(), filepath.Join(dir, "backup.zip"), ""); e != nil {
		t.Fatal(e)
	}
}
