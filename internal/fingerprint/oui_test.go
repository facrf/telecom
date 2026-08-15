package fingerprint

import (
	"context"
	"github.com/local/telecom/internal/database"
	"path/filepath"
	"strings"
	"testing"
)

func TestImportOUI(t *testing.T) {
	db, err := database.Open(filepath.Join(t.TempDir(), "oui.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = database.Migrate(db); err != nil {
		t.Fatal(err)
	}
	count, err := ImportOUI(context.Background(), db, strings.NewReader("prefix,vendor\nAA:BB:CC,Vendor Test\n"))
	if err != nil || count != 1 {
		t.Fatalf("count=%d error=%v", count, err)
	}
	var vendor string
	if err = db.QueryRow("SELECT vendor FROM oui_vendors WHERE prefix='AABBCC'").Scan(&vendor); err != nil || vendor != "Vendor Test" {
		t.Fatalf("vendor=%q error=%v", vendor, err)
	}
}
