package diagrams

import (
	"context"
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("not found")

type Diagram struct {
	ID          string `json:"id"`
	ProjectID   string `json:"projectId"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
type Node struct {
	ID        string  `json:"id"`
	DiagramID string  `json:"diagramId"`
	DeviceID  string  `json:"deviceId"`
	Label     string  `json:"label"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Width     float64 `json:"width"`
	Height    float64 `json:"height"`
	StyleJSON string  `json:"styleJson"`
}
type Edge struct {
	ID              string `json:"id"`
	DiagramID       string `json:"diagramId"`
	SourceNodeID    string `json:"sourceNodeId"`
	TargetNodeID    string `json:"targetNodeId"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Type            string `json:"type"`
	SourceInterface string `json:"sourceInterface"`
	TargetInterface string `json:"targetInterface"`
	Speed           string `json:"speed"`
	VLAN            string `json:"vlan"`
	Technology      string `json:"technology"`
	Color           string `json:"color"`
	LineStyle       string `json:"lineStyle"`
	Notes           string `json:"notes"`
}
type Repository struct{ db *sql.DB }

func New(db *sql.DB) *Repository { return &Repository{db} }
func (r *Repository) List(ctx context.Context, projectID string) ([]Diagram, error) {
	rows, e := r.db.QueryContext(ctx, "SELECT id,project_id,name,description FROM diagrams WHERE project_id=? ORDER BY name", projectID)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var values []Diagram
	for rows.Next() {
		var v Diagram
		if e = rows.Scan(&v.ID, &v.ProjectID, &v.Name, &v.Description); e != nil {
			return nil, e
		}
		values = append(values, v)
	}
	return values, rows.Err()
}
func (r *Repository) Save(ctx context.Context, v Diagram) (Diagram, error) {
	_, e := r.db.ExecContext(ctx, "INSERT INTO diagrams(id,project_id,name,description)VALUES(?,?,?,?)", v.ID, v.ProjectID, v.Name, v.Description)
	if e != nil {
		return v, e
	}
	return v, nil
}
func (r *Repository) UpdateDiagram(ctx context.Context, v Diagram) (Diagram, error) {
	result, e := r.db.ExecContext(ctx, "UPDATE diagrams SET name=?,description=?,updated_at=CURRENT_TIMESTAMP WHERE id=?", v.Name, v.Description, v.ID)
	if e != nil {
		return v, e
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return v, ErrNotFound
	}
	return v, nil
}
func (r *Repository) DeleteDiagram(ctx context.Context, id string) error {
	result, e := r.db.ExecContext(ctx, "DELETE FROM diagrams WHERE id=?", id)
	if e != nil {
		return e
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
func (r *Repository) Graph(ctx context.Context, id string) (Diagram, []Node, []Edge, error) {
	var d Diagram
	e := r.db.QueryRowContext(ctx, "SELECT id,project_id,name,description FROM diagrams WHERE id=?", id).Scan(&d.ID, &d.ProjectID, &d.Name, &d.Description)
	if errors.Is(e, sql.ErrNoRows) {
		return d, nil, nil, ErrNotFound
	}
	if e != nil {
		return d, nil, nil, e
	}
	nodes, e := r.nodes(ctx, id)
	if e != nil {
		return d, nil, nil, e
	}
	edges, e := r.edges(ctx, id)
	return d, nodes, edges, e
}
func (r *Repository) SaveNode(ctx context.Context, v Node) (Node, error) {
	_, e := r.db.ExecContext(ctx, "INSERT INTO diagram_nodes(id,diagram_id,device_id,label,x,y,width,height,style_json)VALUES(?,?,?,?,?,?,?,?,?)", v.ID, v.DiagramID, nullIfEmpty(v.DeviceID), v.Label, v.X, v.Y, v.Width, v.Height, v.StyleJSON)
	return v, e
}
func (r *Repository) UpdateNode(ctx context.Context, v Node) (Node, error) {
	result, e := r.db.ExecContext(ctx, "UPDATE diagram_nodes SET device_id=?,label=?,x=?,y=?,width=?,height=?,style_json=? WHERE id=? AND diagram_id=?", nullIfEmpty(v.DeviceID), v.Label, v.X, v.Y, v.Width, v.Height, v.StyleJSON, v.ID, v.DiagramID)
	if e != nil {
		return v, e
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return v, ErrNotFound
	}
	return v, nil
}
func (r *Repository) DeleteNode(ctx context.Context, id string) error {
	result, e := r.db.ExecContext(ctx, "DELETE FROM diagram_nodes WHERE id=?", id)
	if e != nil {
		return e
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
func (r *Repository) SaveEdge(ctx context.Context, v Edge) (Edge, error) {
	_, e := r.db.ExecContext(ctx, "INSERT INTO diagram_edges(id,diagram_id,source_node_id,target_node_id,name,description,type,source_interface,target_interface,speed,vlan,technology,color,line_style,notes)VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", v.ID, v.DiagramID, v.SourceNodeID, v.TargetNodeID, v.Name, v.Description, v.Type, v.SourceInterface, v.TargetInterface, v.Speed, v.VLAN, v.Technology, v.Color, v.LineStyle, v.Notes)
	return v, e
}
func (r *Repository) UpdateEdge(ctx context.Context, v Edge) (Edge, error) {
	result, e := r.db.ExecContext(ctx, `UPDATE diagram_edges SET source_node_id=?,target_node_id=?,name=?,description=?,type=?,source_interface=?,target_interface=?,speed=?,vlan=?,technology=?,color=?,line_style=?,notes=? WHERE id=? AND diagram_id=?`, v.SourceNodeID, v.TargetNodeID, v.Name, v.Description, v.Type, v.SourceInterface, v.TargetInterface, v.Speed, v.VLAN, v.Technology, v.Color, v.LineStyle, v.Notes, v.ID, v.DiagramID)
	if e != nil {
		return v, e
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return v, ErrNotFound
	}
	return v, nil
}
func (r *Repository) DeleteEdge(ctx context.Context, id string) error {
	result, e := r.db.ExecContext(ctx, "DELETE FROM diagram_edges WHERE id=?", id)
	if e != nil {
		return e
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
func (r *Repository) nodes(ctx context.Context, id string) ([]Node, error) {
	rows, e := r.db.QueryContext(ctx, "SELECT id,diagram_id,COALESCE(device_id,''),label,x,y,width,height,style_json FROM diagram_nodes WHERE diagram_id=?", id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var values []Node
	for rows.Next() {
		var v Node
		if e = rows.Scan(&v.ID, &v.DiagramID, &v.DeviceID, &v.Label, &v.X, &v.Y, &v.Width, &v.Height, &v.StyleJSON); e != nil {
			return nil, e
		}
		values = append(values, v)
	}
	return values, rows.Err()
}
func (r *Repository) edges(ctx context.Context, id string) ([]Edge, error) {
	rows, e := r.db.QueryContext(ctx, "SELECT id,diagram_id,source_node_id,target_node_id,name,description,type,source_interface,target_interface,speed,vlan,technology,color,line_style,notes FROM diagram_edges WHERE diagram_id=?", id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var values []Edge
	for rows.Next() {
		var v Edge
		if e = rows.Scan(&v.ID, &v.DiagramID, &v.SourceNodeID, &v.TargetNodeID, &v.Name, &v.Description, &v.Type, &v.SourceInterface, &v.TargetInterface, &v.Speed, &v.VLAN, &v.Technology, &v.Color, &v.LineStyle, &v.Notes); e != nil {
			return nil, e
		}
		values = append(values, v)
	}
	return values, rows.Err()
}
func nullIfEmpty(v string) any {
	if v == "" {
		return nil
	}
	return v
}
