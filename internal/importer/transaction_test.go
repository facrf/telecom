package importer

import (
	"context"
	"database/sql"
	"errors"
	"github.com/local/telecom/internal/database"
	telecomexport "github.com/local/telecom/internal/export"
	"path/filepath"
	"testing"
)

func TestApplyRollsBack(t *testing.T) {
	db, e := database.Open(filepath.Join(t.TempDir(), "t.sqlite"))
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	if e = database.Migrate(db); e != nil {
		t.Fatal(e)
	}
	doc := telecomexport.NewDocument()
	doc.Client = map[string]any{}
	doc.Project = map[string]any{}
	e = Apply(context.Background(), db, doc, func(ctx context.Context, tx *sql.Tx, d telecomexport.Document) error {
		if _, e := tx.ExecContext(ctx, "INSERT INTO settings(key,value)VALUES('partial','x')"); e != nil {
			return e
		}
		return errors.New("stop")
	})
	if e == nil {
		t.Fatal("failure accepted")
	}
	var count int
	if e = db.QueryRow("SELECT count(*) FROM settings WHERE key='partial'").Scan(&count); e != nil {
		t.Fatal(e)
	}
	if count != 0 {
		t.Fatal("transaction was not rolled back")
	}
}
