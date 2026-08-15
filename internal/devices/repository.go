package devices

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strings"
)

var ErrNotFound = errors.New("not found")

type Device struct {
	ID              string `json:"id"`
	ProjectID       string `json:"projectId"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	CategoryID      string `json:"categoryId"`
	Manufacturer    string `json:"manufacturer"`
	Model           string `json:"model"`
	SerialNumber    string `json:"serialNumber"`
	Hostname        string `json:"hostname"`
	VLAN            string `json:"vlan"`
	Location        string `json:"location"`
	Room            string `json:"room"`
	Rack            string `json:"rack"`
	RackPosition    string `json:"rackPosition"`
	OperatingSystem string `json:"operatingSystem"`
	Firmware        string `json:"firmware"`
	AdminURL        string `json:"adminUrl"`
	Status          string `json:"status"`
	Notes           string `json:"notes"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}
type Address struct {
	ID        string `json:"id"`
	DeviceID  string `json:"deviceId"`
	Type      string `json:"type"`
	Address   string `json:"address"`
	Interface string `json:"interface"`
	VLAN      string `json:"vlan"`
	Primary   bool   `json:"primary"`
}
type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db} }
func (r *Repository) Categories(ctx context.Context) ([]Category, error) {
	rows, e := r.db.QueryContext(ctx, "SELECT id,name FROM device_categories ORDER BY name")
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var a []Category
	for rows.Next() {
		var v Category
		if e = rows.Scan(&v.ID, &v.Name); e != nil {
			return nil, e
		}
		a = append(a, v)
	}
	return a, rows.Err()
}
func (r *Repository) List(ctx context.Context, project, q string) ([]Device, error) {
	rows, e := r.db.QueryContext(ctx, `SELECT id,project_id,name,description,category_id,manufacturer,model,serial_number,hostname,vlan,location,room,rack,rack_position,operating_system,firmware,admin_url,status,notes,created_at,updated_at FROM devices WHERE (?='' OR project_id=?) AND (?='' OR name LIKE ? OR hostname LIKE ? OR manufacturer LIKE ? OR model LIKE ?) ORDER BY name`, project, project, q, "%"+q+"%", "%"+q+"%", "%"+q+"%", "%"+q+"%")
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var a []Device
	for rows.Next() {
		var v Device
		if e = scan(rows, &v); e != nil {
			return nil, e
		}
		a = append(a, v)
	}
	return a, rows.Err()
}
func (r *Repository) Get(ctx context.Context, id string) (Device, error) {
	var v Device
	e := scan(r.db.QueryRowContext(ctx, `SELECT id,project_id,name,description,category_id,manufacturer,model,serial_number,hostname,vlan,location,room,rack,rack_position,operating_system,firmware,admin_url,status,notes,created_at,updated_at FROM devices WHERE id=?`, id), &v)
	if errors.Is(e, sql.ErrNoRows) {
		return v, ErrNotFound
	}
	return v, e
}
func (r *Repository) Create(ctx context.Context, v Device) (Device, error) {
	_, e := r.db.ExecContext(ctx, `INSERT INTO devices(id,project_id,name,description,category_id,manufacturer,model,serial_number,hostname,vlan,location,room,rack,rack_position,operating_system,firmware,admin_url,status,notes)VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, v.ID, v.ProjectID, v.Name, v.Description, v.CategoryID, v.Manufacturer, v.Model, v.SerialNumber, v.Hostname, v.VLAN, v.Location, v.Room, v.Rack, v.RackPosition, v.OperatingSystem, v.Firmware, v.AdminURL, v.Status, v.Notes)
	if e != nil {
		return v, e
	}
	return r.Get(ctx, v.ID)
}
func (r *Repository) Update(ctx context.Context, v Device) (Device, error) {
	result, e := r.db.ExecContext(ctx, `UPDATE devices SET name=?,description=?,category_id=?,manufacturer=?,model=?,serial_number=?,hostname=?,vlan=?,location=?,room=?,rack=?,rack_position=?,operating_system=?,firmware=?,admin_url=?,status=?,notes=?,updated_at=CURRENT_TIMESTAMP WHERE id=?`, v.Name, v.Description, v.CategoryID, v.Manufacturer, v.Model, v.SerialNumber, v.Hostname, v.VLAN, v.Location, v.Room, v.Rack, v.RackPosition, v.OperatingSystem, v.Firmware, v.AdminURL, v.Status, v.Notes, v.ID)
	if e != nil {
		return v, e
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return v, ErrNotFound
	}
	return r.Get(ctx, v.ID)
}
func (r *Repository) Addresses(ctx context.Context, id string) ([]Address, error) {
	rows, e := r.db.QueryContext(ctx, "SELECT id,device_id,type,address,interface,vlan,is_primary FROM device_addresses WHERE device_id=? ORDER BY is_primary DESC,address", id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var a []Address
	for rows.Next() {
		var v Address
		if e = rows.Scan(&v.ID, &v.DeviceID, &v.Type, &v.Address, &v.Interface, &v.VLAN, &v.Primary); e != nil {
			return nil, e
		}
		a = append(a, v)
	}
	return a, rows.Err()
}
func (r *Repository) AddAddress(ctx context.Context, v Address) error {
	_, e := r.db.ExecContext(ctx, "INSERT INTO device_addresses(id,device_id,type,address,interface,vlan,is_primary)VALUES(?,?,?,?,?,?,?)", v.ID, v.DeviceID, v.Type, v.Address, v.Interface, v.VLAN, v.Primary)
	return e
}
func (r *Repository) DeleteAddress(ctx context.Context, id string) error {
	result, e := r.db.ExecContext(ctx, "DELETE FROM device_addresses WHERE id=?", id)
	if e != nil {
		return e
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}
func (r *Repository) Delete(ctx context.Context, id string) error {
	x, e := r.db.ExecContext(ctx, "DELETE FROM devices WHERE id=?", id)
	if e != nil {
		return e
	}
	n, _ := x.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
func (v *Device) Validate() error {
	v.Name = strings.TrimSpace(v.Name)
	if v.Name == "" {
		return fmt.Errorf("nome é obrigatório")
	}
	if v.ProjectID == "" || v.CategoryID == "" {
		return fmt.Errorf("projeto e categoria são obrigatórios")
	}
	if v.Status == "" {
		v.Status = "unknown"
	}
	return nil
}
func (v Address) Validate() error {
	if v.Type != "ipv4" && v.Type != "ipv6" && v.Type != "mac" {
		return fmt.Errorf("tipo de endereço inválido")
	}
	v.Address = strings.TrimSpace(v.Address)
	if v.Type != "mac" {
		if _, e := netip.ParseAddr(v.Address); e != nil {
			return fmt.Errorf("endereço IP inválido")
		}
	} else if _, e := net.ParseMAC(v.Address); e != nil {
		return fmt.Errorf("endereço MAC inválido")
	}
	return nil
}

type scanner interface{ Scan(...any) error }

func scan(s scanner, v *Device) error {
	return s.Scan(&v.ID, &v.ProjectID, &v.Name, &v.Description, &v.CategoryID, &v.Manufacturer, &v.Model, &v.SerialNumber, &v.Hostname, &v.VLAN, &v.Location, &v.Room, &v.Rack, &v.RackPosition, &v.OperatingSystem, &v.Firmware, &v.AdminURL, &v.Status, &v.Notes, &v.CreatedAt, &v.UpdatedAt)
}
