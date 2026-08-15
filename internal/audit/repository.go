package audit

import (
	"context"
	"database/sql"
	"encoding/json"
)

type Entry struct {
	ID         string         `json:"id"`
	Action     string         `json:"action"`
	EntityType string         `json:"entityType"`
	EntityID   string         `json:"entityId"`
	Details    map[string]any `json:"details"`
	CreatedAt  string         `json:"createdAt"`
}
type Repository struct{ db *sql.DB }

func New(db *sql.DB) *Repository { return &Repository{db} }
func (r *Repository) Log(ctx context.Context, e Entry) error {
	data, err := json.Marshal(e.Details)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, "INSERT INTO audit_logs(id,action,entity_type,entity_id,details_json)VALUES(?,?,?,?,?)", e.ID, e.Action, e.EntityType, e.EntityID, string(data))
	return err
}
func (r *Repository) Recent(ctx context.Context, limit int) ([]Entry, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}
	rows, err := r.db.QueryContext(ctx, "SELECT id,action,COALESCE(entity_type,''),COALESCE(entity_id,''),details_json,created_at FROM audit_logs ORDER BY created_at DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []Entry
	for rows.Next() {
		var entry Entry
		var details string
		if err = rows.Scan(&entry.ID, &entry.Action, &entry.EntityType, &entry.EntityID, &details, &entry.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(details), &entry.Details)
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}
