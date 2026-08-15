package dashboard

import (
	"context"
	"database/sql"
)

type Summary struct {
	Clients  int `json:"clients"`
	Projects int `json:"projects"`
	Devices  int `json:"devices"`
	Online   int `json:"online"`
	Offline  int `json:"offline"`
	Cameras  int `json:"cameras"`
	Routers  int `json:"routers"`
	Switches int `json:"switches"`
	Servers  int `json:"servers"`
}
type Service struct{ db *sql.DB }

func New(db *sql.DB) *Service { return &Service{db} }
func (s *Service) Summary(ctx context.Context, projectID string) (Summary, error) {
	var v Summary
	e := s.db.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM clients),(SELECT count(*) FROM projects),(SELECT count(*) FROM devices WHERE ?='' OR project_id=?),(SELECT count(*) FROM devices WHERE status='online' AND (?='' OR project_id=?)),(SELECT count(*) FROM devices WHERE status='offline' AND (?='' OR project_id=?)),(SELECT count(*) FROM devices WHERE category_id='ip-camera' AND (?='' OR project_id=?)),(SELECT count(*) FROM devices WHERE category_id='router' AND (?='' OR project_id=?)),(SELECT count(*) FROM devices WHERE category_id='switch' AND (?='' OR project_id=?)),(SELECT count(*) FROM devices WHERE category_id='server' AND (?='' OR project_id=?))`, projectID, projectID, projectID, projectID, projectID, projectID, projectID, projectID, projectID, projectID, projectID, projectID, projectID, projectID).Scan(&v.Clients, &v.Projects, &v.Devices, &v.Online, &v.Offline, &v.Cameras, &v.Routers, &v.Switches, &v.Servers)
	return v, e
}
