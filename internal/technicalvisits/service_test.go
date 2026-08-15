package technicalvisits_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/local/telecom/internal/database"
	"github.com/local/telecom/internal/technicalvisits"
)

func testService(t *testing.T) (*technicalvisits.Service, func()) {
	t.Helper()
	db, err := database.Open(filepath.Join(t.TempDir(), "telecom.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	if err = database.Migrate(db); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO clients(id,name) VALUES('c1','Empresa ABC'); INSERT INTO projects(id,client_id,name) VALUES('p1','c1','Matriz'),('p2','c1','Filial')`); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO devices(id,project_id,name,category_id) VALUES('d1','p1','Switch Rack','switch'),('d2','p2','Router Filial','router')`); err != nil {
		t.Fatal(err)
	}
	return technicalvisits.NewService(technicalvisits.NewRepository(db)), func() { _ = db.Close() }
}

func TestVisitStructuredDetails(t *testing.T) {
	service, closeDB := testService(t)
	defer closeDB()
	ctx := context.Background()
	visit, err := service.Create(ctx, technicalvisits.Visit{ProjectID: "p1", Title: "VT-2", VisitType: "maintenance", ScheduledAt: "2026-08-15T10:00"})
	if err == nil {
		t.Fatal("expected invalid visit type")
	}
	visit, err = service.Create(ctx, technicalvisits.Visit{ProjectID: "p1", Title: "VT-2", VisitType: "corrective_maintenance", ScheduledAt: "2026-08-15T10:00"})
	if err != nil {
		t.Fatal(err)
	}
	linked, err := service.SaveDevice(ctx, visit.ID, technicalvisits.VisitDevice{DeviceID: "d1", Role: "configured", Notes: "VLAN 30"})
	if err != nil {
		t.Fatal(err)
	}
	if linked.DeviceName != "Switch Rack" {
		t.Fatalf("device name = %q", linked.DeviceName)
	}
	if _, err = service.SaveDevice(ctx, visit.ID, technicalvisits.VisitDevice{DeviceID: "d2", Role: "tested"}); !errors.Is(err, technicalvisits.ErrInvalidDeviceProject) {
		t.Fatalf("expected project mismatch, got %v", err)
	}
	serviceItem, err := service.SaveService(ctx, visit.ID, technicalvisits.VisitService{Description: "Configuração da VLAN", Category: "Configuração", DeviceID: "d1", Technician: "Ana", Order: 1})
	if err != nil {
		t.Fatal(err)
	}
	check, err := service.SaveChecklist(ctx, visit.ID, technicalvisits.ChecklistItem{Text: "Testar conectividade", Status: "completed", Order: 1})
	if err != nil {
		t.Fatal(err)
	}
	material, err := service.SaveMaterial(ctx, visit.ID, technicalvisits.Material{Quantity: 10, Unit: "m", Description: "Cabo CAT6"})
	if err != nil {
		t.Fatal(err)
	}
	pending, err := service.SavePendingItem(ctx, visit.ID, technicalvisits.PendingItem{Description: "Trocar nobreak", Priority: "high", Responsible: "Carlos", Status: "pending"})
	if err != nil {
		t.Fatal(err)
	}
	if values, _ := service.ListServices(ctx, visit.ID); len(values) != 1 || values[0].DeviceName != "Switch Rack" {
		t.Fatalf("services = %+v", values)
	}
	if values, _ := service.ListChecklist(ctx, visit.ID); len(values) != 1 {
		t.Fatalf("checklist = %+v", values)
	}
	if values, _ := service.ListMaterials(ctx, visit.ID); len(values) != 1 {
		t.Fatalf("materials = %+v", values)
	}
	if values, _ := service.ListPendingItems(ctx, visit.ID); len(values) != 1 {
		t.Fatalf("pending = %+v", values)
	}
	for _, query := range []string{"Configuração da VLAN", "Switch Rack"} {
		values, listErr := service.List(ctx, technicalvisits.Filters{Query: query})
		if listErr != nil || len(values) != 1 || values[0].ID != visit.ID {
			t.Fatalf("query %q returned %+v, %v", query, values, listErr)
		}
	}
	check.Status = "pending"
	if _, err = service.SaveChecklist(ctx, visit.ID, check); err != nil {
		t.Fatal(err)
	}
	if err = service.DeleteService(ctx, visit.ID, serviceItem.ID); err != nil {
		t.Fatal(err)
	}
	if err = service.DeleteChecklist(ctx, visit.ID, check.ID); err != nil {
		t.Fatal(err)
	}
	if err = service.DeleteMaterial(ctx, visit.ID, material.ID); err != nil {
		t.Fatal(err)
	}
	if err = service.DeletePendingItem(ctx, visit.ID, pending.ID); err != nil {
		t.Fatal(err)
	}
	if err = service.DeleteDevice(ctx, visit.ID, linked.ID); err != nil {
		t.Fatal(err)
	}
}

func TestCRUDProtocolFiltersAndConcurrency(t *testing.T) {
	service, closeDB := testService(t)
	defer closeDB()
	ctx := context.Background()
	created, err := service.Create(ctx, technicalvisits.Visit{ProjectID: "p1", Title: "Manutenção do link", VisitType: "corrective_maintenance", ScheduledAt: "2026-08-15T13:30", ResponsibleTechnician: "Ana"})
	if err != nil {
		t.Fatal(err)
	}
	if created.Protocol != "VT-2026-000001" || created.ClientID != "c1" || created.ClientName != "Empresa ABC" {
		t.Fatalf("unexpected created visit: %+v", created)
	}
	second, err := service.Create(ctx, technicalvisits.Visit{ProjectID: "p2", Title: "Vistoria", VisitType: "inspection", ScheduledAt: "2026-08-22T09:00"})
	if err != nil {
		t.Fatal(err)
	}
	if second.Protocol != "VT-2026-000002" {
		t.Fatalf("second protocol = %q", second.Protocol)
	}
	listed, err := service.List(ctx, technicalvisits.Filters{ProjectID: "p1", Technician: "An"})
	if err != nil || len(listed) != 1 || listed[0].ID != created.ID {
		t.Fatalf("filtered list = %+v, %v", listed, err)
	}
	stale := created
	created.Title = "Link restaurado"
	created.Status = "completed"
	created.Result = "resolved"
	created.StartedAt = "2026-08-15T13:30"
	created.FinishedAt = "2026-08-15T16:15"
	updated, err := service.Update(ctx, created.ID, created)
	if err != nil {
		t.Fatal(err)
	}
	if updated.DurationMinutes != 165 {
		t.Fatalf("duration = %d", updated.DurationMinutes)
	}
	stale.Title = "Sobrescrita"
	if _, err = service.Update(ctx, stale.ID, stale); !errors.Is(err, technicalvisits.ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
	if err = service.Delete(ctx, created.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = service.Get(ctx, created.ID); !errors.Is(err, technicalvisits.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestVisitValidationAndProjectIntegrity(t *testing.T) {
	service, closeDB := testService(t)
	defer closeDB()
	ctx := context.Background()
	if _, err := service.Create(ctx, technicalvisits.Visit{ProjectID: "missing", Title: "Teste", VisitType: "inspection", ScheduledAt: "2026-08-15T10:00"}); !errors.Is(err, technicalvisits.ErrProjectNotFound) {
		t.Fatalf("expected invalid project, got %v", err)
	}
	if _, err := service.Create(ctx, technicalvisits.Visit{ProjectID: "p1", Title: "Concluída sem resultado", VisitType: "inspection", Status: "completed", ScheduledAt: "2026-08-15T10:00"}); err == nil {
		t.Fatal("expected completed visit validation error")
	}
}
