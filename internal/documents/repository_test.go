package documents

import (
	"context"
	"github.com/local/telecom/internal/database"
	"path/filepath"
	"testing"
)

func TestSave(t *testing.T) {
	db, e := database.Open(filepath.Join(t.TempDir(), "t.sqlite"))
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	if e = database.Migrate(db); e != nil {
		t.Fatal(e)
	}
	if _, e = db.Exec("INSERT INTO clients(id,name)VALUES('c','C')"); e != nil {
		t.Fatal(e)
	}
	if _, e = db.Exec("INSERT INTO projects(id,client_id,name)VALUES('p','c','P')"); e != nil {
		t.Fatal(e)
	}
	r := New(db)
	d, e := r.Save(context.Background(), Document{ID: "d", ProjectID: "p", Title: "Rede"})
	if e != nil || d.Title != "Rede" {
		t.Fatalf("%v %#v", e, d)
	}
}
