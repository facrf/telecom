package technicalvisits

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var ErrItemNotFound = errors.New("technical visit item not found")
var ErrInvalidDeviceProject = errors.New("device does not belong to technical visit project")
var ErrDuplicateDeviceRole = errors.New("device role already linked")

func (r *Repository) ListDevices(ctx context.Context, visitID string) ([]VisitDevice, error) {
	if err := r.ensureVisit(ctx, visitID); err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT x.id,x.technical_visit_id,x.device_id,d.name,c.name,x.role,x.notes,x.created_at,x.updated_at FROM technical_visit_devices x JOIN devices d ON d.id=x.device_id JOIN device_categories c ON c.id=d.category_id WHERE x.technical_visit_id=? ORDER BY d.name,x.role`, visitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []VisitDevice{}
	for rows.Next() {
		var item VisitDevice
		if err = rows.Scan(&item.ID, &item.TechnicalVisitID, &item.DeviceID, &item.DeviceName, &item.DeviceCategory, &item.Role, &item.Notes, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) CreateDevice(ctx context.Context, item VisitDevice) (VisitDevice, error) {
	if err := r.ensureDeviceForVisit(ctx, item.TechnicalVisitID, item.DeviceID); err != nil {
		return VisitDevice{}, err
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO technical_visit_devices(id,technical_visit_id,device_id,role,notes) VALUES(?,?,?,?,?)`, item.ID, item.TechnicalVisitID, item.DeviceID, item.Role, item.Notes)
	if err != nil {
		return VisitDevice{}, detailDBError(err)
	}
	return r.getDevice(ctx, item.TechnicalVisitID, item.ID)
}

func (r *Repository) UpdateDevice(ctx context.Context, item VisitDevice) (VisitDevice, error) {
	if err := r.ensureDeviceForVisit(ctx, item.TechnicalVisitID, item.DeviceID); err != nil {
		return VisitDevice{}, err
	}
	result, err := r.db.ExecContext(ctx, `UPDATE technical_visit_devices SET device_id=?,role=?,notes=?,updated_at=strftime('%Y-%m-%dT%H:%M:%fZ','now') WHERE id=? AND technical_visit_id=?`, item.DeviceID, item.Role, item.Notes, item.ID, item.TechnicalVisitID)
	if err != nil {
		return VisitDevice{}, detailDBError(err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return VisitDevice{}, ErrItemNotFound
	}
	return r.getDevice(ctx, item.TechnicalVisitID, item.ID)
}

func (r *Repository) getDevice(ctx context.Context, visitID, id string) (VisitDevice, error) {
	var item VisitDevice
	err := r.db.QueryRowContext(ctx, `SELECT x.id,x.technical_visit_id,x.device_id,d.name,c.name,x.role,x.notes,x.created_at,x.updated_at FROM technical_visit_devices x JOIN devices d ON d.id=x.device_id JOIN device_categories c ON c.id=d.category_id WHERE x.technical_visit_id=? AND x.id=?`, visitID, id).Scan(&item.ID, &item.TechnicalVisitID, &item.DeviceID, &item.DeviceName, &item.DeviceCategory, &item.Role, &item.Notes, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return VisitDevice{}, ErrItemNotFound
	}
	return item, err
}
func (r *Repository) DeleteDevice(ctx context.Context, visitID, id string) error {
	return r.deleteDetail(ctx, `DELETE FROM technical_visit_devices WHERE id=? AND technical_visit_id=?`, id, visitID)
}

func (r *Repository) ListServices(ctx context.Context, visitID string) ([]VisitService, error) {
	if err := r.ensureVisit(ctx, visitID); err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT s.id,s.technical_visit_id,s.description,s.category,COALESCE(s.device_id,''),COALESCE(d.name,''),s.performed_at,s.technician,s.notes,s.sort_order,s.created_at,s.updated_at FROM technical_visit_services s LEFT JOIN devices d ON d.id=s.device_id WHERE s.technical_visit_id=? ORDER BY s.sort_order,s.created_at`, visitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []VisitService{}
	for rows.Next() {
		var item VisitService
		if err = rows.Scan(&item.ID, &item.TechnicalVisitID, &item.Description, &item.Category, &item.DeviceID, &item.DeviceName, &item.PerformedAt, &item.Technician, &item.Notes, &item.Order, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
func (r *Repository) CreateService(ctx context.Context, item VisitService) (VisitService, error) {
	if err := r.ensureOptionalDeviceForVisit(ctx, item.TechnicalVisitID, item.DeviceID); err != nil {
		return VisitService{}, err
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO technical_visit_services(id,technical_visit_id,description,category,device_id,performed_at,technician,notes,sort_order) VALUES(?,?,?,?,?,?,?,?,?)`, item.ID, item.TechnicalVisitID, item.Description, item.Category, nullIfEmpty(item.DeviceID), item.PerformedAt, item.Technician, item.Notes, item.Order)
	if err != nil {
		return VisitService{}, detailDBError(err)
	}
	return r.getService(ctx, item.TechnicalVisitID, item.ID)
}
func (r *Repository) UpdateService(ctx context.Context, item VisitService) (VisitService, error) {
	if err := r.ensureOptionalDeviceForVisit(ctx, item.TechnicalVisitID, item.DeviceID); err != nil {
		return VisitService{}, err
	}
	result, err := r.db.ExecContext(ctx, `UPDATE technical_visit_services SET description=?,category=?,device_id=?,performed_at=?,technician=?,notes=?,sort_order=?,updated_at=strftime('%Y-%m-%dT%H:%M:%fZ','now') WHERE id=? AND technical_visit_id=?`, item.Description, item.Category, nullIfEmpty(item.DeviceID), item.PerformedAt, item.Technician, item.Notes, item.Order, item.ID, item.TechnicalVisitID)
	if err != nil {
		return VisitService{}, detailDBError(err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return VisitService{}, ErrItemNotFound
	}
	return r.getService(ctx, item.TechnicalVisitID, item.ID)
}
func (r *Repository) getService(ctx context.Context, visitID, id string) (VisitService, error) {
	var item VisitService
	err := r.db.QueryRowContext(ctx, `SELECT s.id,s.technical_visit_id,s.description,s.category,COALESCE(s.device_id,''),COALESCE(d.name,''),s.performed_at,s.technician,s.notes,s.sort_order,s.created_at,s.updated_at FROM technical_visit_services s LEFT JOIN devices d ON d.id=s.device_id WHERE s.technical_visit_id=? AND s.id=?`, visitID, id).Scan(&item.ID, &item.TechnicalVisitID, &item.Description, &item.Category, &item.DeviceID, &item.DeviceName, &item.PerformedAt, &item.Technician, &item.Notes, &item.Order, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return VisitService{}, ErrItemNotFound
	}
	return item, err
}
func (r *Repository) DeleteService(ctx context.Context, visitID, id string) error {
	return r.deleteDetail(ctx, `DELETE FROM technical_visit_services WHERE id=? AND technical_visit_id=?`, id, visitID)
}

func (r *Repository) ListChecklist(ctx context.Context, visitID string) ([]ChecklistItem, error) {
	if err := r.ensureVisit(ctx, visitID); err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id,technical_visit_id,text,status,notes,sort_order,created_at,updated_at FROM technical_visit_checklist WHERE technical_visit_id=? ORDER BY sort_order,created_at`, visitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []ChecklistItem{}
	for rows.Next() {
		var item ChecklistItem
		if err = rows.Scan(&item.ID, &item.TechnicalVisitID, &item.Text, &item.Status, &item.Notes, &item.Order, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
func (r *Repository) CreateChecklist(ctx context.Context, item ChecklistItem) (ChecklistItem, error) {
	if err := r.ensureVisit(ctx, item.TechnicalVisitID); err != nil {
		return ChecklistItem{}, err
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO technical_visit_checklist(id,technical_visit_id,text,status,notes,sort_order)VALUES(?,?,?,?,?,?)`, item.ID, item.TechnicalVisitID, item.Text, item.Status, item.Notes, item.Order)
	if err != nil {
		return ChecklistItem{}, err
	}
	return r.getChecklist(ctx, item.TechnicalVisitID, item.ID)
}
func (r *Repository) UpdateChecklist(ctx context.Context, item ChecklistItem) (ChecklistItem, error) {
	result, err := r.db.ExecContext(ctx, `UPDATE technical_visit_checklist SET text=?,status=?,notes=?,sort_order=?,updated_at=strftime('%Y-%m-%dT%H:%M:%fZ','now') WHERE id=? AND technical_visit_id=?`, item.Text, item.Status, item.Notes, item.Order, item.ID, item.TechnicalVisitID)
	if err != nil {
		return ChecklistItem{}, err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ChecklistItem{}, ErrItemNotFound
	}
	return r.getChecklist(ctx, item.TechnicalVisitID, item.ID)
}
func (r *Repository) getChecklist(ctx context.Context, visitID, id string) (ChecklistItem, error) {
	var item ChecklistItem
	err := r.db.QueryRowContext(ctx, `SELECT id,technical_visit_id,text,status,notes,sort_order,created_at,updated_at FROM technical_visit_checklist WHERE technical_visit_id=? AND id=?`, visitID, id).Scan(&item.ID, &item.TechnicalVisitID, &item.Text, &item.Status, &item.Notes, &item.Order, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ChecklistItem{}, ErrItemNotFound
	}
	return item, err
}
func (r *Repository) DeleteChecklist(ctx context.Context, visitID, id string) error {
	return r.deleteDetail(ctx, `DELETE FROM technical_visit_checklist WHERE id=? AND technical_visit_id=?`, id, visitID)
}

func (r *Repository) ListMaterials(ctx context.Context, visitID string) ([]Material, error) {
	if err := r.ensureVisit(ctx, visitID); err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id,technical_visit_id,quantity,unit,description,brand,model,notes,created_at,updated_at FROM technical_visit_materials WHERE technical_visit_id=? ORDER BY created_at`, visitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Material{}
	for rows.Next() {
		var item Material
		if err = rows.Scan(&item.ID, &item.TechnicalVisitID, &item.Quantity, &item.Unit, &item.Description, &item.Brand, &item.Model, &item.Notes, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
func (r *Repository) CreateMaterial(ctx context.Context, item Material) (Material, error) {
	if err := r.ensureVisit(ctx, item.TechnicalVisitID); err != nil {
		return Material{}, err
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO technical_visit_materials(id,technical_visit_id,quantity,unit,description,brand,model,notes)VALUES(?,?,?,?,?,?,?,?)`, item.ID, item.TechnicalVisitID, item.Quantity, item.Unit, item.Description, item.Brand, item.Model, item.Notes)
	if err != nil {
		return Material{}, err
	}
	return r.getMaterial(ctx, item.TechnicalVisitID, item.ID)
}
func (r *Repository) UpdateMaterial(ctx context.Context, item Material) (Material, error) {
	result, err := r.db.ExecContext(ctx, `UPDATE technical_visit_materials SET quantity=?,unit=?,description=?,brand=?,model=?,notes=?,updated_at=strftime('%Y-%m-%dT%H:%M:%fZ','now') WHERE id=? AND technical_visit_id=?`, item.Quantity, item.Unit, item.Description, item.Brand, item.Model, item.Notes, item.ID, item.TechnicalVisitID)
	if err != nil {
		return Material{}, err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return Material{}, ErrItemNotFound
	}
	return r.getMaterial(ctx, item.TechnicalVisitID, item.ID)
}
func (r *Repository) getMaterial(ctx context.Context, visitID, id string) (Material, error) {
	var item Material
	err := r.db.QueryRowContext(ctx, `SELECT id,technical_visit_id,quantity,unit,description,brand,model,notes,created_at,updated_at FROM technical_visit_materials WHERE technical_visit_id=? AND id=?`, visitID, id).Scan(&item.ID, &item.TechnicalVisitID, &item.Quantity, &item.Unit, &item.Description, &item.Brand, &item.Model, &item.Notes, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Material{}, ErrItemNotFound
	}
	return item, err
}
func (r *Repository) DeleteMaterial(ctx context.Context, visitID, id string) error {
	return r.deleteDetail(ctx, `DELETE FROM technical_visit_materials WHERE id=? AND technical_visit_id=?`, id, visitID)
}

func (r *Repository) ListPendingItems(ctx context.Context, visitID string) ([]PendingItem, error) {
	if err := r.ensureVisit(ctx, visitID); err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id,technical_visit_id,description,priority,responsible,due_at,status,created_at,updated_at FROM technical_visit_pending_items WHERE technical_visit_id=? ORDER BY CASE priority WHEN 'critical' THEN 0 WHEN 'high' THEN 1 WHEN 'normal' THEN 2 ELSE 3 END,due_at,created_at`, visitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []PendingItem{}
	for rows.Next() {
		var item PendingItem
		if err = rows.Scan(&item.ID, &item.TechnicalVisitID, &item.Description, &item.Priority, &item.Responsible, &item.DueAt, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
func (r *Repository) CreatePendingItem(ctx context.Context, item PendingItem) (PendingItem, error) {
	if err := r.ensureVisit(ctx, item.TechnicalVisitID); err != nil {
		return PendingItem{}, err
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO technical_visit_pending_items(id,technical_visit_id,description,priority,responsible,due_at,status)VALUES(?,?,?,?,?,?,?)`, item.ID, item.TechnicalVisitID, item.Description, item.Priority, item.Responsible, item.DueAt, item.Status)
	if err != nil {
		return PendingItem{}, err
	}
	return r.getPendingItem(ctx, item.TechnicalVisitID, item.ID)
}
func (r *Repository) UpdatePendingItem(ctx context.Context, item PendingItem) (PendingItem, error) {
	result, err := r.db.ExecContext(ctx, `UPDATE technical_visit_pending_items SET description=?,priority=?,responsible=?,due_at=?,status=?,updated_at=strftime('%Y-%m-%dT%H:%M:%fZ','now') WHERE id=? AND technical_visit_id=?`, item.Description, item.Priority, item.Responsible, item.DueAt, item.Status, item.ID, item.TechnicalVisitID)
	if err != nil {
		return PendingItem{}, err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return PendingItem{}, ErrItemNotFound
	}
	return r.getPendingItem(ctx, item.TechnicalVisitID, item.ID)
}
func (r *Repository) getPendingItem(ctx context.Context, visitID, id string) (PendingItem, error) {
	var item PendingItem
	err := r.db.QueryRowContext(ctx, `SELECT id,technical_visit_id,description,priority,responsible,due_at,status,created_at,updated_at FROM technical_visit_pending_items WHERE technical_visit_id=? AND id=?`, visitID, id).Scan(&item.ID, &item.TechnicalVisitID, &item.Description, &item.Priority, &item.Responsible, &item.DueAt, &item.Status, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return PendingItem{}, ErrItemNotFound
	}
	return item, err
}
func (r *Repository) DeletePendingItem(ctx context.Context, visitID, id string) error {
	return r.deleteDetail(ctx, `DELETE FROM technical_visit_pending_items WHERE id=? AND technical_visit_id=?`, id, visitID)
}

func (r *Repository) ensureVisit(ctx context.Context, visitID string) error {
	var exists bool
	if err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM technical_visits WHERE id=?)`, visitID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}
	return nil
}
func (r *Repository) ensureDeviceForVisit(ctx context.Context, visitID, deviceID string) error {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM technical_visits v JOIN devices d ON d.project_id=v.project_id WHERE v.id=? AND d.id=?)`, visitID, deviceID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		if err = r.ensureVisit(ctx, visitID); err != nil {
			return err
		}
		return ErrInvalidDeviceProject
	}
	return nil
}
func (r *Repository) ensureOptionalDeviceForVisit(ctx context.Context, visitID, deviceID string) error {
	if deviceID == "" {
		return r.ensureVisit(ctx, visitID)
	}
	return r.ensureDeviceForVisit(ctx, visitID, deviceID)
}
func (r *Repository) deleteDetail(ctx context.Context, query, id, visitID string) error {
	result, err := r.db.ExecContext(ctx, query, id, visitID)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrItemNotFound
	}
	return nil
}
func detailDBError(err error) error {
	message := err.Error()
	if strings.Contains(message, "same project") {
		return ErrInvalidDeviceProject
	}
	if strings.Contains(message, "UNIQUE constraint failed") {
		return ErrDuplicateDeviceRole
	}
	return fmt.Errorf("save technical visit detail: %w", err)
}
func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}
