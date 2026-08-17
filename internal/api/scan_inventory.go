package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

type inventoryHostRequest struct {
	IP         string `json:"ip"`
	Name       string `json:"name"`
	CategoryID string `json:"categoryId"`
}

func (h scanHandlers) inventory(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		ProjectID string                 `json:"projectId"`
		Items     []inventoryHostRequest `json:"items"`
	}
	if !decodeJSON(w, r, &payload) {
		return
	}
	payload.ProjectID = strings.TrimSpace(payload.ProjectID)
	if payload.ProjectID == "" || len(payload.Items) == 0 {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Projeto e ao menos um host são obrigatórios")
		return
	}
	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		serverError(w, err)
		return
	}
	defer tx.Rollback()
	var projectExists bool
	if err = tx.QueryRowContext(r.Context(), `SELECT EXISTS(SELECT 1 FROM projects WHERE id=?)`, payload.ProjectID).Scan(&projectExists); err != nil {
		serverError(w, err)
		return
	}
	if !projectExists {
		writeError(w, http.StatusUnprocessableEntity, "INVALID_PROJECT", "Projeto não encontrado")
		return
	}
	created, skipped := 0, 0
	createdIDs := []string{}
	for _, requested := range payload.Items {
		var ip, mac, hostname, manufacturer, deviceType string
		err = tx.QueryRowContext(r.Context(), `SELECT ip,mac,hostname,manufacturer,device_type FROM scan_hosts WHERE scan_id=? AND ip=?`, chi.URLParam(r, "scanID"), strings.TrimSpace(requested.IP)).Scan(&ip, &mac, &hostname, &manufacturer, &deviceType)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusUnprocessableEntity, "INVALID_HOST", "Um host selecionado não pertence ao scan")
			return
		}
		if err != nil {
			serverError(w, err)
			return
		}
		var existing string
		err = tx.QueryRowContext(r.Context(), `SELECT d.id FROM devices d JOIN device_addresses a ON a.device_id=d.id WHERE d.project_id=? AND (a.address=? OR (?<>'' AND a.address=?)) LIMIT 1`, payload.ProjectID, ip, mac, mac).Scan(&existing)
		if err == nil {
			skipped++
			continue
		}
		if !errors.Is(err, sql.ErrNoRows) {
			serverError(w, err)
			return
		}
		category := strings.TrimSpace(requested.CategoryID)
		if category == "" {
			category = map[string]string{"Câmera IP": "ip-camera", "Servidor ou equipamento de rede": "server"}[deviceType]
		}
		if category == "" {
			category = "other"
		}
		var categoryExists bool
		if err = tx.QueryRowContext(r.Context(), `SELECT EXISTS(SELECT 1 FROM device_categories WHERE id=?)`, category).Scan(&categoryExists); err != nil {
			serverError(w, err)
			return
		}
		if !categoryExists {
			category = "other"
		}
		name := strings.TrimSpace(requested.Name)
		if name == "" {
			name = hostname
		}
		if name == "" {
			name = strings.TrimSpace(deviceType + " " + ip)
		}
		if name == "" {
			name = "Dispositivo " + ip
		}
		deviceID := newID()
		if _, err = tx.ExecContext(r.Context(), `INSERT INTO devices(id,project_id,name,category_id,manufacturer,hostname,status)VALUES(?,?,?,?,?,?,'online')`, deviceID, payload.ProjectID, name, category, manufacturer, hostname); err != nil {
			serverError(w, err)
			return
		}
		if _, err = tx.ExecContext(r.Context(), `INSERT INTO device_addresses(id,device_id,type,address,is_primary)VALUES(?,?,'ipv4',?,1)`, newID(), deviceID, ip); err != nil {
			serverError(w, err)
			return
		}
		if mac != "" {
			if _, err = tx.ExecContext(r.Context(), `INSERT INTO device_addresses(id,device_id,type,address,is_primary)VALUES(?,?,'mac',?,1)`, newID(), deviceID, mac); err != nil {
				serverError(w, err)
				return
			}
		}
		created++
		createdIDs = append(createdIDs, deviceID)
	}
	if _, err = tx.ExecContext(r.Context(), `UPDATE network_scans SET project_id=? WHERE id=?`, payload.ProjectID, chi.URLParam(r, "scanID")); err != nil {
		serverError(w, err)
		return
	}
	if err = tx.Commit(); err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"created": created, "skipped": skipped, "deviceIds": createdIDs})
}
