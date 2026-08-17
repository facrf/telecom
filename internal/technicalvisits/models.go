package technicalvisits

import (
	"fmt"
	"strings"
	"time"
)

const (
	StatusDraft      = "draft"
	StatusScheduled  = "scheduled"
	StatusInProgress = "in_progress"
	StatusCompleted  = "completed"
	StatusCancelled  = "cancelled"
)

var VisitTypes = []string{
	"installation", "preventive_maintenance", "corrective_maintenance", "inspection", "diagnosis",
	"network_expansion", "cctv", "cabling", "fiber_optic", "wifi", "telephony", "migration",
	"device_replacement", "configuration", "documentation", "other",
}

var statuses = setOf(StatusDraft, StatusScheduled, StatusInProgress, StatusCompleted, StatusCancelled)
var results = setOf("", "resolved", "partially_resolved", "not_resolved", "waiting_material", "waiting_customer", "requires_return", "no_fault_found")
var visitTypes = setOf(VisitTypes...)

type Visit struct {
	ID                    string `json:"id"`
	ProjectID             string `json:"projectId"`
	ClientID              string `json:"clientId"`
	ClientName            string `json:"clientName"`
	ProjectName           string `json:"projectName"`
	Protocol              string `json:"protocol"`
	Title                 string `json:"title"`
	VisitType             string `json:"visitType"`
	Status                string `json:"status"`
	Result                string `json:"result"`
	ScheduledAt           string `json:"scheduledAt"`
	StartedAt             string `json:"startedAt"`
	FinishedAt            string `json:"finishedAt"`
	DurationMinutes       int    `json:"durationMinutes"`
	ResponsibleTechnician string `json:"responsibleTechnician"`
	Requester             string `json:"requester"`
	LocalContact          string `json:"localContact"`
	RequestDescription    string `json:"requestDescription"`
	InitialSituation      string `json:"initialSituation"`
	Diagnosis             string `json:"diagnosis"`
	WorkSummary           string `json:"workSummary"`
	Recommendations       string `json:"recommendations"`
	PendingSummary        string `json:"pendingSummary"`
	CustomerNotes         string `json:"customerNotes"`
	InternalNotes         string `json:"internalNotes"`
	RequiresReturn        bool   `json:"requiresReturn"`
	ReturnReason          string `json:"returnReason"`
	SuggestedReturnAt     string `json:"suggestedReturnAt"`
	CreatedAt             string `json:"createdAt"`
	UpdatedAt             string `json:"updatedAt"`
}

type Filters struct {
	Query, ClientID, ProjectID, Technician, Status, Result, DateFrom, DateTo string
}

type ValidationError struct{ Field, Message string }

func (e ValidationError) Error() string { return e.Message }

func (v *Visit) Validate() error {
	v.ProjectID = strings.TrimSpace(v.ProjectID)
	v.Title = strings.TrimSpace(v.Title)
	v.VisitType = strings.TrimSpace(v.VisitType)
	v.Status = strings.TrimSpace(v.Status)
	v.Result = strings.TrimSpace(v.Result)
	if v.ProjectID == "" {
		return ValidationError{"projectId", "Projeto é obrigatório"}
	}
	if v.Title == "" {
		return ValidationError{"title", "Título é obrigatório"}
	}
	if len(v.Title) > 160 {
		return ValidationError{"title", "Título deve ter no máximo 160 caracteres"}
	}
	if !visitTypes[v.VisitType] {
		return ValidationError{"visitType", "Tipo de atendimento inválido"}
	}
	if !statuses[v.Status] {
		return ValidationError{"status", "Status inválido"}
	}
	if !results[v.Result] {
		return ValidationError{"result", "Resultado inválido"}
	}
	if v.Result == "requires_return" {
		v.RequiresReturn = true
	}
	if v.ScheduledAt == "" {
		return ValidationError{"scheduledAt", "Data da visita é obrigatória"}
	}
	if _, err := parseDate(v.ScheduledAt); err != nil {
		return ValidationError{"scheduledAt", "Data da visita inválida"}
	}
	for field, value := range map[string]string{
		"responsibleTechnician": v.ResponsibleTechnician, "requester": v.Requester, "localContact": v.LocalContact,
		"requestDescription": v.RequestDescription, "initialSituation": v.InitialSituation, "diagnosis": v.Diagnosis,
		"workSummary": v.WorkSummary, "recommendations": v.Recommendations, "pendingSummary": v.PendingSummary,
		"customerNotes": v.CustomerNotes, "internalNotes": v.InternalNotes, "returnReason": v.ReturnReason,
	} {
		if len(value) > 12000 {
			return ValidationError{field, "Campo deve ter no máximo 12000 caracteres"}
		}
	}
	if v.Status == StatusCompleted && v.Result == "" {
		return ValidationError{"result", "Resultado é obrigatório para visita concluída"}
	}
	if v.SuggestedReturnAt != "" {
		if _, err := parseDate(v.SuggestedReturnAt); err != nil {
			return ValidationError{"suggestedReturnAt", "Data sugerida para retorno é inválida"}
		}
	}
	if v.FinishedAt != "" && v.StartedAt == "" {
		return ValidationError{"startedAt", "Hora inicial é obrigatória quando houver hora final"}
	}
	if v.StartedAt != "" && v.FinishedAt != "" {
		start, errStart := parseDate(v.StartedAt)
		end, errEnd := parseDate(v.FinishedAt)
		if errStart != nil {
			return ValidationError{"startedAt", "Hora inicial inválida"}
		}
		if errEnd != nil {
			return ValidationError{"finishedAt", "Hora final inválida"}
		}
		if end.Before(start) {
			return ValidationError{"finishedAt", "Hora final não pode ser anterior à inicial"}
		}
	}
	return nil
}

func (v *Visit) CalculateDuration() {
	v.DurationMinutes = 0
	start, e1 := parseDate(v.StartedAt)
	end, e2 := parseDate(v.FinishedAt)
	if e1 == nil && e2 == nil && !end.Before(start) {
		v.DurationMinutes = int(end.Sub(start).Minutes())
	}
}

func parseDate(value string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05", "2006-01-02T15:04", "2006-01-02 15:04:05", "2006-01-02 15:04", "2006-01-02"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date")
}

func setOf(values ...string) map[string]bool {
	result := make(map[string]bool, len(values))
	for _, value := range values {
		result[value] = true
	}
	return result
}
