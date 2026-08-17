package api

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/local/telecom/internal/database"
)

func TestHealth(t *testing.T) {
	db, err := database.Open(filepath.Join(t.TempDir(), "telecom.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := database.Migrate(db); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	res := httptest.NewRecorder()
	NewRouter(db, slog.Default(), 2, t.TempDir(), func() {}).ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("health returned %d", res.Code)
	}
	if got := res.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Fatalf("X-Frame-Options = %q", got)
	}
}

func TestClientCRUD(t *testing.T) {
	db, err := database.Open(filepath.Join(t.TempDir(), "telecom.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := database.Migrate(db); err != nil {
		t.Fatal(err)
	}
	router := NewRouter(db, slog.Default(), 2, t.TempDir(), func() {})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/clients", bytes.NewBufferString(`{"name":"Empresa ABC"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("create returned %d: %s", response.Code, response.Body.String())
	}
	var client map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &client); err != nil {
		t.Fatal(err)
	}
	get := httptest.NewRequest(http.MethodGet, "/api/v1/clients?q=Empresa", nil)
	list := httptest.NewRecorder()
	router.ServeHTTP(list, get)
	if list.Code != http.StatusOK || !bytes.Contains(list.Body.Bytes(), []byte("Empresa ABC")) {
		t.Fatalf("list failed: %d %s", list.Code, list.Body.String())
	}
}

func TestTechnicalVisitCRUDAPI(t *testing.T) {
	db, err := database.Open(filepath.Join(t.TempDir(), "telecom.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = database.Migrate(db); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO clients(id,name) VALUES('client-api','Empresa API'); INSERT INTO projects(id,client_id,name) VALUES('project-api','client-api','Matriz'); INSERT INTO devices(id,project_id,name,category_id) VALUES('device-api','project-api','Switch API','switch')`); err != nil {
		t.Fatal(err)
	}
	router := NewRouter(db, slog.Default(), 2, t.TempDir(), func() {})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/technical-visits", bytes.NewBufferString(`{"projectId":"project-api","title":"Diagnóstico do link","visitType":"diagnosis","scheduledAt":"2026-08-15T10:00"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("create returned %d: %s", response.Code, response.Body.String())
	}
	var created struct{ ID, Protocol, UpdatedAt string }
	if err = json.Unmarshal(response.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Protocol != "VT-2026-000001" {
		t.Fatalf("protocol = %q", created.Protocol)
	}
	listRequest := httptest.NewRequest(http.MethodGet, "/api/v1/projects/project-api/technical-visits?status=draft", nil)
	listResponse := httptest.NewRecorder()
	router.ServeHTTP(listResponse, listRequest)
	if listResponse.Code != http.StatusOK || !bytes.Contains(listResponse.Body.Bytes(), []byte(created.ID)) {
		t.Fatalf("list returned %d: %s", listResponse.Code, listResponse.Body.String())
	}
	detailRequest := httptest.NewRequest(http.MethodPost, "/api/v1/technical-visits/"+created.ID+"/materials", bytes.NewBufferString(`{"quantity":10,"unit":"m","description":"Cabo CAT6"}`))
	detailRequest.Header.Set("Content-Type", "application/json")
	detailResponse := httptest.NewRecorder()
	router.ServeHTTP(detailResponse, detailRequest)
	if detailResponse.Code != http.StatusCreated {
		t.Fatalf("create material returned %d: %s", detailResponse.Code, detailResponse.Body.String())
	}
	deviceRequest := httptest.NewRequest(http.MethodPost, "/api/v1/technical-visits/"+created.ID+"/devices", bytes.NewBufferString(`{"deviceId":"device-api","role":"tested"}`))
	deviceRequest.Header.Set("Content-Type", "application/json")
	deviceResponse := httptest.NewRecorder()
	router.ServeHTTP(deviceResponse, deviceRequest)
	if deviceResponse.Code != http.StatusCreated || !bytes.Contains(deviceResponse.Body.Bytes(), []byte("Switch API")) {
		t.Fatalf("create device relation returned %d: %s", deviceResponse.Code, deviceResponse.Body.String())
	}
	deleteRequest := httptest.NewRequest(http.MethodDelete, "/api/v1/technical-visits/"+created.ID, nil)
	deleteResponse := httptest.NewRecorder()
	router.ServeHTTP(deleteResponse, deleteRequest)
	if deleteResponse.Code != http.StatusNoContent {
		t.Fatalf("delete returned %d: %s", deleteResponse.Code, deleteResponse.Body.String())
	}
	var details int
	if err = db.QueryRow(`SELECT COUNT(*) FROM technical_visit_materials WHERE technical_visit_id=?`, created.ID).Scan(&details); err != nil || details != 0 {
		t.Fatalf("visit details were not cascaded: count=%d err=%v", details, err)
	}
}

func TestScanWithoutProjectAndAssign(t *testing.T) {
	db, err := database.Open(filepath.Join(t.TempDir(), "telecom.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = database.Migrate(db); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO clients(id,name) VALUES('c1','Cliente Teste'); INSERT INTO projects(id,client_id,name) VALUES('p1','c1','Projeto Teste')`); err != nil {
		t.Fatal(err)
	}
	router := NewRouter(db, slog.Default(), 2, t.TempDir(), func() {})

	// Inicia scan sem projeto
	req := httptest.NewRequest(http.MethodPost, "/api/v1/scans", bytes.NewBufferString(`{"network":"192.168.1.0/28"}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", res.Code, res.Body.String())
	}
	var created struct{ ID, ProjectID, Status string }
	if err = json.Unmarshal(res.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.ProjectID != "" {
		t.Fatalf("expected empty ProjectID, got %q", created.ProjectID)
	}

	// Lista scans gerais
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/scans", nil)
	listRes := httptest.NewRecorder()
	router.ServeHTTP(listRes, listReq)
	if listRes.Code != http.StatusOK || !bytes.Contains(listRes.Body.Bytes(), []byte(created.ID)) {
		t.Fatalf("list without project filter failed: %d %s", listRes.Code, listRes.Body.String())
	}

	// Atualiza e vincula ao projeto 'p1'
	patchReq := httptest.NewRequest(http.MethodPatch, "/api/v1/scans/"+created.ID, bytes.NewBufferString(`{"projectId":"p1"}`))
	patchReq.Header.Set("Content-Type", "application/json")
	patchRes := httptest.NewRecorder()
	router.ServeHTTP(patchRes, patchReq)
	if patchRes.Code != http.StatusOK {
		t.Fatalf("patch returned %d: %s", patchRes.Code, patchRes.Body.String())
	}
	var updated struct{ ID, ProjectID string }
	if err = json.Unmarshal(patchRes.Body.Bytes(), &updated); err != nil {
		t.Fatal(err)
	}
	if updated.ProjectID != "p1" {
		t.Fatalf("expected ProjectID 'p1', got %q", updated.ProjectID)
	}
}

func TestScanInventoryBulkIsAtomicAndSkipsDuplicates(t *testing.T) {
	db, err := database.Open(filepath.Join(t.TempDir(), "telecom.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = database.Migrate(db); err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`
		INSERT INTO clients(id,name) VALUES('bulk-client','Cliente Bulk');
		INSERT INTO projects(id,client_id,name) VALUES('bulk-project','bulk-client','Projeto Bulk');
		INSERT INTO network_scans(id,network,status) VALUES('bulk-scan','192.168.20.0/24','completed');
		INSERT INTO scan_hosts(id,scan_id,ip,mac,hostname,status,manufacturer,device_type) VALUES
		('bulk-host-1','bulk-scan','192.168.20.10','00:11:22:33:44:55','camera-entrada','online','Intelbras','Câmera IP'),
		('bulk-host-2','bulk-scan','192.168.20.11','','switch-core','online','TP-Link','Equipamento de rede');`)
	if err != nil {
		t.Fatal(err)
	}
	router := NewRouter(db, slog.Default(), 2, t.TempDir(), func() {})
	payload := `{"projectId":"bulk-project","items":[{"ip":"192.168.20.10","name":"Câmera Entrada","categoryId":"ip-camera"},{"ip":"192.168.20.11","name":"Switch Core","categoryId":"switch"}]}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/scans/bulk-scan/inventory", bytes.NewBufferString(payload))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusCreated || !bytes.Contains(response.Body.Bytes(), []byte(`"created":2`)) {
		t.Fatalf("bulk inventory returned %d: %s", response.Code, response.Body.String())
	}
	request = httptest.NewRequest(http.MethodPost, "/api/v1/scans/bulk-scan/inventory", bytes.NewBufferString(payload))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusCreated || !bytes.Contains(response.Body.Bytes(), []byte(`"skipped":2`)) {
		t.Fatalf("duplicate bulk inventory returned %d: %s", response.Code, response.Body.String())
	}
	var devices, addresses int
	if err = db.QueryRow(`SELECT count(*) FROM devices WHERE project_id='bulk-project'`).Scan(&devices); err != nil {
		t.Fatal(err)
	}
	if err = db.QueryRow(`SELECT count(*) FROM device_addresses`).Scan(&addresses); err != nil {
		t.Fatal(err)
	}
	if devices != 2 || addresses != 3 {
		t.Fatalf("devices=%d addresses=%d", devices, addresses)
	}
}

func TestOperationalDashboard(t *testing.T) {
	db, err := database.Open(filepath.Join(t.TempDir(), "telecom.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = database.Migrate(db); err != nil {
		t.Fatal(err)
	}
	today := time.Now().Format("2006-01-02")
	_, err = db.Exec(`INSERT INTO clients(id,name) VALUES('ops-client','Cliente Operacional'); INSERT INTO projects(id,client_id,name) VALUES('ops-project','ops-client','Matriz'); INSERT INTO technical_visits(id,project_id,protocol,title,visit_type,status,scheduled_at,requires_return) VALUES('ops-visit','ops-project','VT-2026-999999','Atendimento de hoje','diagnosis','draft',?,1); INSERT INTO technical_visit_pending_items(id,technical_visit_id,description,priority,due_at) VALUES('ops-pending','ops-visit','Trocar conector','critical','2020-01-01')`, today+"T10:00")
	if err != nil {
		t.Fatal(err)
	}
	router := NewRouter(db, slog.Default(), 2, t.TempDir(), func() {})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/operations?project_id=ops-project", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !bytes.Contains(response.Body.Bytes(), []byte("Atendimento de hoje")) || !bytes.Contains(response.Body.Bytes(), []byte(`"overdue":1`)) {
		t.Fatalf("operational dashboard returned %d: %s", response.Code, response.Body.String())
	}
}
