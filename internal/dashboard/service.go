package dashboard

import (
	"context"
	"database/sql"
	"time"
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

type ActionItem struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Subtitle  string `json:"subtitle"`
	Status    string `json:"status"`
	DueAt     string `json:"dueAt"`
	ProjectID string `json:"projectId"`
}

type Operations struct {
	Visits   []ActionItem `json:"visits"`
	Pending  []ActionItem `json:"pending"`
	Drafts   int          `json:"drafts"`
	Returns  int          `json:"returns"`
	Overdue  int          `json:"overdue"`
}

func New(db *sql.DB) *Service { return &Service{db} }
func (s *Service) Summary(ctx context.Context, projectID string) (Summary, error) {
	var v Summary
	e := s.db.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM clients),(SELECT count(*) FROM projects),(SELECT count(*) FROM devices WHERE ?='' OR project_id=?),(SELECT count(*) FROM devices WHERE status='online' AND (?='' OR project_id=?)),(SELECT count(*) FROM devices WHERE status='offline' AND (?='' OR project_id=?)),(SELECT count(*) FROM devices WHERE category_id='ip-camera' AND (?='' OR project_id=?)),(SELECT count(*) FROM devices WHERE category_id='router' AND (?='' OR project_id=?)),(SELECT count(*) FROM devices WHERE category_id='switch' AND (?='' OR project_id=?)),(SELECT count(*) FROM devices WHERE category_id='server' AND (?='' OR project_id=?))`, projectID, projectID, projectID, projectID, projectID, projectID, projectID, projectID, projectID, projectID, projectID, projectID, projectID, projectID).Scan(&v.Clients, &v.Projects, &v.Devices, &v.Online, &v.Offline, &v.Cameras, &v.Routers, &v.Switches, &v.Servers)
	return v, e
}

func (s *Service) Operations(ctx context.Context, projectID string) (Operations, error) {
	today := time.Now().Format("2006-01-02")
	result := Operations{Visits: []ActionItem{}, Pending: []ActionItem{}}
	rows, err := s.db.QueryContext(ctx, `SELECT v.id,'technical_visit',v.protocol || ' · ' || v.title,c.name || ' · ' || p.name,v.status,v.scheduled_at,v.project_id FROM technical_visits v JOIN projects p ON p.id=v.project_id JOIN clients c ON c.id=p.client_id WHERE (?='' OR v.project_id=?) AND (date(v.scheduled_at)=date(?) OR v.status='in_progress') AND v.status<>'cancelled' ORDER BY CASE WHEN v.status='in_progress' THEN 0 ELSE 1 END,v.scheduled_at LIMIT 20`, projectID, projectID, today)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var item ActionItem
		if err = rows.Scan(&item.ID, &item.Type, &item.Title, &item.Subtitle, &item.Status, &item.DueAt, &item.ProjectID); err != nil {
			rows.Close()
			return result, err
		}
		result.Visits = append(result.Visits, item)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return result, err
	}
	rows.Close()

	rows, err = s.db.QueryContext(ctx, `SELECT i.id,'pending_item',i.description,c.name || ' · ' || p.name,i.status,i.due_at,v.project_id FROM technical_visit_pending_items i JOIN technical_visits v ON v.id=i.technical_visit_id JOIN projects p ON p.id=v.project_id JOIN clients c ON c.id=p.client_id WHERE (?='' OR v.project_id=?) AND i.status IN ('pending','in_progress') ORDER BY CASE i.priority WHEN 'critical' THEN 0 WHEN 'high' THEN 1 ELSE 2 END,CASE WHEN i.due_at='' THEN 1 ELSE 0 END,i.due_at LIMIT 20`, projectID, projectID)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var item ActionItem
		if err = rows.Scan(&item.ID, &item.Type, &item.Title, &item.Subtitle, &item.Status, &item.DueAt, &item.ProjectID); err != nil {
			return result, err
		}
		result.Pending = append(result.Pending, item)
		if item.DueAt != "" && item.DueAt < today {
			result.Overdue++
		}
	}
	if err = rows.Err(); err != nil {
		return result, err
	}
	if err = s.db.QueryRowContext(ctx, `SELECT count(*) FROM technical_visits WHERE status='draft' AND (?='' OR project_id=?)`, projectID, projectID).Scan(&result.Drafts); err != nil {
		return result, err
	}
	err = s.db.QueryRowContext(ctx, `SELECT count(*) FROM technical_visits WHERE requires_return=1 AND status<>'cancelled' AND (?='' OR project_id=?)`, projectID, projectID).Scan(&result.Returns)
	return result, err
}
