package tags

import (
	"context"
	"database/sql"
	"fmt"
)

type Tag struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}
type Repository struct{ db *sql.DB }

func New(db *sql.DB) *Repository { return &Repository{db} }
func (r *Repository) List(ctx context.Context) ([]Tag, error) {
	rows, e := r.db.QueryContext(ctx, "SELECT id,name,color FROM tags ORDER BY name")
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var tags []Tag
	for rows.Next() {
		var tag Tag
		if e = rows.Scan(&tag.ID, &tag.Name, &tag.Color); e != nil {
			return nil, e
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}
func (r *Repository) Create(ctx context.Context, tag Tag) (Tag, error) {
	if tag.Name == "" {
		return tag, fmt.Errorf("nome é obrigatório")
	}
	_, e := r.db.ExecContext(ctx, "INSERT INTO tags(id,name,color)VALUES(?,?,?)", tag.ID, tag.Name, tag.Color)
	return tag, e
}
func (r *Repository) Assign(ctx context.Context, entityType, entityID, tagID string) error {
	if entityType != "client" && entityType != "project" && entityType != "device" {
		return fmt.Errorf("entidade inválida")
	}
	_, e := r.db.ExecContext(ctx, "INSERT OR IGNORE INTO entity_tags(entity_type,entity_id,tag_id)VALUES(?,?,?)", entityType, entityID, tagID)
	return e
}
func (r *Repository) Assigned(ctx context.Context, entityType, entityID string) ([]Tag, error) {
	rows, e := r.db.QueryContext(ctx, "SELECT t.id,t.name,t.color FROM tags t JOIN entity_tags et ON et.tag_id=t.id WHERE et.entity_type=? AND et.entity_id=? ORDER BY t.name", entityType, entityID)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var values []Tag
	for rows.Next() {
		var value Tag
		if e = rows.Scan(&value.ID, &value.Name, &value.Color); e != nil {
			return nil, e
		}
		values = append(values, value)
	}
	return values, rows.Err()
}
func (r *Repository) Unassign(ctx context.Context, entityType, entityID, tagID string) error {
	_, e := r.db.ExecContext(ctx, "DELETE FROM entity_tags WHERE entity_type=? AND entity_id=? AND tag_id=?", entityType, entityID, tagID)
	return e
}
func (r *Repository) Delete(ctx context.Context, id string) error {
	_, e := r.db.ExecContext(ctx, "DELETE FROM tags WHERE id=?", id)
	return e
}
