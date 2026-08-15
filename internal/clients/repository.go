package clients

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("not found")

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }
func (r *Repository) Database() *sql.DB    { return r.db }

func (r *Repository) ListClients(ctx context.Context, query string) ([]Client, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,name,legal_name,document,phone,email,contact_name,address,city,state,postal_code,description,notes,created_at,updated_at FROM clients WHERE ? = '' OR name LIKE ? OR legal_name LIKE ? OR document LIKE ? OR email LIKE ? ORDER BY name`, query, "%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%")
	if err != nil {
		return nil, fmt.Errorf("list clients: %w", err)
	}
	defer rows.Close()
	var items []Client
	for rows.Next() {
		var item Client
		if err := scanClient(rows, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (r *Repository) GetClient(ctx context.Context, id string) (Client, error) {
	var item Client
	err := scanClient(r.db.QueryRowContext(ctx, `SELECT id,name,legal_name,document,phone,email,contact_name,address,city,state,postal_code,description,notes,created_at,updated_at FROM clients WHERE id=?`, id), &item)
	if errors.Is(err, sql.ErrNoRows) {
		return Client{}, ErrNotFound
	}
	return item, err
}
func (r *Repository) CreateClient(ctx context.Context, item Client) (Client, error) {
	_, err := r.db.ExecContext(ctx, `INSERT INTO clients(id,name,legal_name,document,phone,email,contact_name,address,city,state,postal_code,description,notes) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`, item.ID, item.Name, item.LegalName, item.Document, item.Phone, item.Email, item.ContactName, item.Address, item.City, item.State, item.PostalCode, item.Description, item.Notes)
	if err != nil {
		return Client{}, fmt.Errorf("create client: %w", err)
	}
	return r.GetClient(ctx, item.ID)
}
func (r *Repository) UpdateClient(ctx context.Context, item Client) (Client, error) {
	result, err := r.db.ExecContext(ctx, `UPDATE clients SET name=?,legal_name=?,document=?,phone=?,email=?,contact_name=?,address=?,city=?,state=?,postal_code=?,description=?,notes=?,updated_at=CURRENT_TIMESTAMP WHERE id=?`, item.Name, item.LegalName, item.Document, item.Phone, item.Email, item.ContactName, item.Address, item.City, item.State, item.PostalCode, item.Description, item.Notes, item.ID)
	if err != nil {
		return Client{}, fmt.Errorf("update client: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return Client{}, ErrNotFound
	}
	return r.GetClient(ctx, item.ID)
}
func (r *Repository) DeleteClient(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM clients WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("delete client: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
func (r *Repository) ListProjects(ctx context.Context, clientID, query string) ([]Project, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,client_id,name,description,location,address,local_contact,notes,created_at,updated_at FROM projects WHERE (?='' OR client_id=?) AND (?='' OR name LIKE ? OR location LIKE ?) ORDER BY name`, clientID, clientID, query, "%"+query+"%", "%"+query+"%")
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()
	var items []Project
	for rows.Next() {
		var item Project
		if err := scanProject(rows, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (r *Repository) GetProject(ctx context.Context, id string) (Project, error) {
	var item Project
	err := scanProject(r.db.QueryRowContext(ctx, `SELECT id,client_id,name,description,location,address,local_contact,notes,created_at,updated_at FROM projects WHERE id=?`, id), &item)
	if errors.Is(err, sql.ErrNoRows) {
		return Project{}, ErrNotFound
	}
	return item, err
}
func (r *Repository) CreateProject(ctx context.Context, item Project) (Project, error) {
	_, err := r.db.ExecContext(ctx, `INSERT INTO projects(id,client_id,name,description,location,address,local_contact,notes) VALUES(?,?,?,?,?,?,?,?)`, item.ID, item.ClientID, item.Name, item.Description, item.Location, item.Address, item.LocalContact, item.Notes)
	if err != nil {
		return Project{}, fmt.Errorf("create project: %w", err)
	}
	return r.GetProject(ctx, item.ID)
}
func (r *Repository) UpdateProject(ctx context.Context, item Project) (Project, error) {
	result, err := r.db.ExecContext(ctx, `UPDATE projects SET name=?,description=?,location=?,address=?,local_contact=?,notes=?,updated_at=CURRENT_TIMESTAMP WHERE id=?`, item.Name, item.Description, item.Location, item.Address, item.LocalContact, item.Notes, item.ID)
	if err != nil {
		return Project{}, fmt.Errorf("update project: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return Project{}, ErrNotFound
	}
	return r.GetProject(ctx, item.ID)
}
func (r *Repository) DeleteProject(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM projects WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

type scanner interface{ Scan(...any) error }

func scanClient(row scanner, item *Client) error {
	return row.Scan(&item.ID, &item.Name, &item.LegalName, &item.Document, &item.Phone, &item.Email, &item.ContactName, &item.Address, &item.City, &item.State, &item.PostalCode, &item.Description, &item.Notes, &item.CreatedAt, &item.UpdatedAt)
}
func scanProject(row scanner, item *Project) error {
	return row.Scan(&item.ID, &item.ClientID, &item.Name, &item.Description, &item.Location, &item.Address, &item.LocalContact, &item.Notes, &item.CreatedAt, &item.UpdatedAt)
}
