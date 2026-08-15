package search

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/local/telecom/internal/database"
)

func TestFindTechnicalVisitByStructuredDetails(t *testing.T) {
	db, err := database.Open(filepath.Join(t.TempDir(), "telecom.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = database.Migrate(db); err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO clients(id,name)VALUES('c','Cliente');INSERT INTO projects(id,client_id,name)VALUES('p','c','Projeto');INSERT INTO devices(id,project_id,name,category_id)VALUES('d','p','Switch Rack 01','switch');INSERT INTO technical_visits(id,project_id,protocol,title,visit_type,status,result,scheduled_at)VALUES('v','p','VT-2026-000001','Manutenção','diagnosis','draft','','2026-08-15T10:00');INSERT INTO technical_visit_devices(id,technical_visit_id,device_id,role)VALUES('vd','v','d','tested');INSERT INTO technical_visit_services(id,technical_visit_id,description,technician)VALUES('s','v','Configuração VLAN 30','Ana')`)
	if err != nil {
		t.Fatal(err)
	}
	service := New(db)
	for _, query := range []string{"VLAN 30", "Switch Rack 01"} {
		results, findErr := service.Find(context.Background(), query)
		if findErr != nil {
			t.Fatal(findErr)
		}
		foundVisit := false
		for _, result := range results {
			if result.ID == "v" && result.Type == "technical_visit" {
				foundVisit = true
			}
		}
		if !foundVisit {
			t.Fatalf("query %q returned %+v", query, results)
		}
	}
}
