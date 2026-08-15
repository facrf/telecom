package technicalvisits

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrNotFound = errors.New("technical visit not found")
var ErrConflict = errors.New("technical visit changed by another user")
var ErrProjectNotFound = errors.New("project not found")

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

const selectVisit = `SELECT v.id,v.project_id,p.client_id,c.name,p.name,v.protocol,v.title,v.visit_type,v.status,v.result,v.scheduled_at,v.started_at,v.finished_at,v.responsible_technician,v.requester,v.local_contact,v.request_description,v.initial_situation,v.diagnosis,v.work_summary,v.recommendations,v.pending_summary,v.customer_notes,v.internal_notes,v.requires_return,v.return_reason,v.suggested_return_at,v.created_at,v.updated_at FROM technical_visits v JOIN projects p ON p.id=v.project_id JOIN clients c ON c.id=p.client_id`

func (r *Repository) List(ctx context.Context, f Filters) ([]Visit, error) {
	like := "%" + strings.TrimSpace(f.Query) + "%"
	rows, err := r.db.QueryContext(ctx, selectVisit+` WHERE (?='' OR v.protocol LIKE ? OR v.title LIKE ? OR v.responsible_technician LIKE ? OR v.request_description LIKE ? OR v.diagnosis LIKE ? OR c.name LIKE ? OR p.name LIKE ? OR v.work_summary LIKE ? OR EXISTS(SELECT 1 FROM technical_visit_services s WHERE s.technical_visit_id=v.id AND (s.description LIKE ? OR s.technician LIKE ?)) OR EXISTS(SELECT 1 FROM technical_visit_devices vd JOIN devices d ON d.id=vd.device_id WHERE vd.technical_visit_id=v.id AND d.name LIKE ?)) AND (?='' OR p.client_id=?) AND (?='' OR v.project_id=?) AND (?='' OR v.responsible_technician LIKE ?) AND (?='' OR v.status=?) AND (?='' OR v.result=?) AND (?='' OR date(v.scheduled_at)>=date(?)) AND (?='' OR date(v.scheduled_at)<=date(?)) ORDER BY v.scheduled_at DESC,v.created_at DESC`, f.Query, like, like, like, like, like, like, like, like, like, like, like, f.ClientID, f.ClientID, f.ProjectID, f.ProjectID, f.Technician, "%"+f.Technician+"%", f.Status, f.Status, f.Result, f.Result, f.DateFrom, f.DateFrom, f.DateTo, f.DateTo)
	if err != nil {
		return nil, fmt.Errorf("list technical visits: %w", err)
	}
	defer rows.Close()
	result := []Visit{}
	for rows.Next() {
		var v Visit
		if err = scanVisit(rows, &v); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}

func (r *Repository) Get(ctx context.Context, id string) (Visit, error) {
	var v Visit
	err := scanVisit(r.db.QueryRowContext(ctx, selectVisit+` WHERE v.id=?`, id), &v)
	if errors.Is(err, sql.ErrNoRows) {
		return Visit{}, ErrNotFound
	}
	return v, err
}

func (r *Repository) Create(ctx context.Context, v Visit) (Visit, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Visit{}, err
	}
	defer tx.Rollback()
	var exists bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM projects WHERE id=?)`, v.ProjectID).Scan(&exists); err != nil {
		return Visit{}, err
	}
	if !exists {
		return Visit{}, ErrProjectNotFound
	}
	year := protocolYear(v.ScheduledAt)
	var sequence int
	if err = tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(CAST(substr(protocol,9) AS INTEGER)),0)+1 FROM technical_visits WHERE protocol LIKE ?`, `VT-`+year+`-%`).Scan(&sequence); err != nil {
		return Visit{}, err
	}
	v.Protocol = fmt.Sprintf("VT-%s-%06d", year, sequence)
	_, err = tx.ExecContext(ctx, `INSERT INTO technical_visits(id,project_id,protocol,title,visit_type,status,result,scheduled_at,started_at,finished_at,responsible_technician,requester,local_contact,request_description,initial_situation,diagnosis,work_summary,recommendations,pending_summary,customer_notes,internal_notes,requires_return,return_reason,suggested_return_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, visitArgs(v)...)
	if err != nil {
		return Visit{}, fmt.Errorf("create technical visit: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return Visit{}, err
	}
	return r.Get(ctx, v.ID)
}

func (r *Repository) Update(ctx context.Context, v Visit) (Visit, error) {
	result, err := r.db.ExecContext(ctx, `UPDATE technical_visits SET project_id=?,title=?,visit_type=?,status=?,result=?,scheduled_at=?,started_at=?,finished_at=?,responsible_technician=?,requester=?,local_contact=?,request_description=?,initial_situation=?,diagnosis=?,work_summary=?,recommendations=?,pending_summary=?,customer_notes=?,internal_notes=?,requires_return=?,return_reason=?,suggested_return_at=?,updated_at=CASE WHEN strftime('%Y-%m-%dT%H:%M:%fZ','now')<=updated_at THEN strftime('%Y-%m-%dT%H:%M:%fZ',updated_at,'+0.001 seconds') ELSE strftime('%Y-%m-%dT%H:%M:%fZ','now') END WHERE id=? AND updated_at=?`, v.ProjectID, v.Title, v.VisitType, v.Status, v.Result, v.ScheduledAt, v.StartedAt, v.FinishedAt, v.ResponsibleTechnician, v.Requester, v.LocalContact, v.RequestDescription, v.InitialSituation, v.Diagnosis, v.WorkSummary, v.Recommendations, v.PendingSummary, v.CustomerNotes, v.InternalNotes, v.RequiresReturn, v.ReturnReason, v.SuggestedReturnAt, v.ID, v.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "FOREIGN KEY") {
			return Visit{}, ErrProjectNotFound
		}
		return Visit{}, fmt.Errorf("update technical visit: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		var exists bool
		if e := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM technical_visits WHERE id=?)`, v.ID).Scan(&exists); e != nil {
			return Visit{}, e
		}
		if exists {
			return Visit{}, ErrConflict
		}
		return Visit{}, ErrNotFound
	}
	return r.Get(ctx, v.ID)
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM technical_visits WHERE id=?`, id)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func visitArgs(v Visit) []any {
	return []any{v.ID, v.ProjectID, v.Protocol, v.Title, v.VisitType, v.Status, v.Result, v.ScheduledAt, v.StartedAt, v.FinishedAt, v.ResponsibleTechnician, v.Requester, v.LocalContact, v.RequestDescription, v.InitialSituation, v.Diagnosis, v.WorkSummary, v.Recommendations, v.PendingSummary, v.CustomerNotes, v.InternalNotes, v.RequiresReturn, v.ReturnReason, v.SuggestedReturnAt}
}

type scanner interface{ Scan(...any) error }

func scanVisit(row scanner, v *Visit) error {
	err := row.Scan(&v.ID, &v.ProjectID, &v.ClientID, &v.ClientName, &v.ProjectName, &v.Protocol, &v.Title, &v.VisitType, &v.Status, &v.Result, &v.ScheduledAt, &v.StartedAt, &v.FinishedAt, &v.ResponsibleTechnician, &v.Requester, &v.LocalContact, &v.RequestDescription, &v.InitialSituation, &v.Diagnosis, &v.WorkSummary, &v.Recommendations, &v.PendingSummary, &v.CustomerNotes, &v.InternalNotes, &v.RequiresReturn, &v.ReturnReason, &v.SuggestedReturnAt, &v.CreatedAt, &v.UpdatedAt)
	if err == nil {
		v.CalculateDuration()
	}
	return err
}
func protocolYear(value string) string {
	if parsed, err := parseDate(value); err == nil {
		return parsed.Format("2006")
	}
	if len(value) >= 4 {
		return value[:4]
	}
	return time.Now().Format("2006")
}
