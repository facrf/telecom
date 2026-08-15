package database

import (
	"path/filepath"
	"testing"
)

func TestMigrate(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "telecom.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 15 {
		t.Fatalf("expected fifteen migrations, got %d", count)
	}
}
