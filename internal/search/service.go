package search

import (
	"context"
	"database/sql"
	"strings"
)

type Result struct {
	Type      string `json:"type"`
	ID        string `json:"id"`
	Title     string `json:"title"`
	Subtitle  string `json:"subtitle"`
	ProjectID string `json:"projectId"`
}
type Service struct{ db *sql.DB }

func New(db *sql.DB) *Service { return &Service{db} }
func (s *Service) Find(ctx context.Context, query string) ([]Result, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []Result{}, nil
	}
	p := "%" + query + "%"
	rows, e := s.db.QueryContext(ctx, `SELECT type,id,title,subtitle,project_id FROM (SELECT 'client' type,id,name title,COALESCE(city,'') subtitle,'' project_id FROM clients WHERE name LIKE ? OR legal_name LIKE ? OR document LIKE ? UNION ALL SELECT 'project',id,name,location,id FROM projects WHERE name LIKE ? OR location LIKE ? UNION ALL SELECT 'device',d.id,d.name,COALESCE(a.address,d.hostname),d.project_id FROM devices d LEFT JOIN device_addresses a ON a.device_id=d.id AND a.is_primary=1 WHERE d.name LIKE ? OR d.hostname LIKE ? OR d.manufacturer LIKE ? OR d.model LIKE ? OR d.location LIKE ? OR a.address LIKE ? UNION ALL SELECT 'technical_visit',v.id,v.protocol || ' · ' || v.title,c.name || ' · ' || p.name,v.project_id FROM technical_visits v JOIN projects p ON p.id=v.project_id JOIN clients c ON c.id=p.client_id WHERE v.protocol LIKE ? OR v.title LIKE ? OR v.responsible_technician LIKE ? OR v.request_description LIKE ? OR v.diagnosis LIKE ? OR v.work_summary LIKE ? OR EXISTS(SELECT 1 FROM technical_visit_services s WHERE s.technical_visit_id=v.id AND (s.description LIKE ? OR s.technician LIKE ?)) OR EXISTS(SELECT 1 FROM technical_visit_devices vd JOIN devices d ON d.id=vd.device_id WHERE vd.technical_visit_id=v.id AND d.name LIKE ?)) ORDER BY title LIMIT 100`, p, p, p, p, p, p, p, p, p, p, p, p, p, p, p, p, p, p, p, p)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var results []Result
	for rows.Next() {
		var r Result
		if e = rows.Scan(&r.Type, &r.ID, &r.Title, &r.Subtitle, &r.ProjectID); e != nil {
			return nil, e
		}
		results = append(results, r)
	}
	return results, rows.Err()
}
