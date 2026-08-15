package devices

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/local/telecom/internal/database"
)

func TestInventoryCRUDAndAddresses(t *testing.T) {
	db, err := database.Open(filepath.Join(t.TempDir(), "inventory.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := database.Migrate(db); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := db.Exec(`INSERT INTO clients(id,name) VALUES('client-1','Cliente')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO projects(id,client_id,name) VALUES('project-1','client-1','Matriz')`); err != nil {
		t.Fatal(err)
	}
	repository := NewRepository(db)
	created, err := repository.Create(ctx, Device{ID: "device-1", ProjectID: "project-1", Name: "Switch principal", CategoryID: "switch", Status: "online"})
	if err != nil {
		t.Fatal(err)
	}
	if created.Name != "Switch principal" {
		t.Fatalf("device name = %q", created.Name)
	}
	address := Address{ID: "address-1", DeviceID: created.ID, Type: "ipv4", Address: "192.168.0.2", Primary: true}
	if err := address.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := repository.AddAddress(ctx, address); err != nil {
		t.Fatal(err)
	}
	addresses, err := repository.Addresses(ctx, created.ID)
	if err != nil || len(addresses) != 1 {
		t.Fatalf("addresses = %d, %v", len(addresses), err)
	}
	invalidAddr := Address{Type: "mac", Address: "invalid"}
	if err := invalidAddr.Validate(); err == nil {
		t.Fatal("invalid MAC was accepted")
	}
	if err := repository.Delete(ctx, created.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.Addresses(ctx, created.ID); err != nil {
		t.Fatal(err)
	}
}
