package api

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

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
