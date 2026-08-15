package technicalvisits

import (
	"strings"
	"time"
)

type VisitDevice struct {
	ID               string `json:"id"`
	TechnicalVisitID string `json:"technicalVisitId"`
	DeviceID         string `json:"deviceId"`
	DeviceName       string `json:"deviceName"`
	DeviceCategory   string `json:"deviceCategory"`
	Role             string `json:"role"`
	Notes            string `json:"notes"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
}

type VisitService struct {
	ID               string `json:"id"`
	TechnicalVisitID string `json:"technicalVisitId"`
	Description      string `json:"description"`
	Category         string `json:"category"`
	DeviceID         string `json:"deviceId"`
	DeviceName       string `json:"deviceName"`
	PerformedAt      string `json:"performedAt"`
	Technician       string `json:"technician"`
	Notes            string `json:"notes"`
	Order            int    `json:"order"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
}

type ChecklistItem struct {
	ID               string `json:"id"`
	TechnicalVisitID string `json:"technicalVisitId"`
	Text             string `json:"text"`
	Status           string `json:"status"`
	Notes            string `json:"notes"`
	Order            int    `json:"order"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
}

type Material struct {
	ID               string  `json:"id"`
	TechnicalVisitID string  `json:"technicalVisitId"`
	Quantity         float64 `json:"quantity"`
	Unit             string  `json:"unit"`
	Description      string  `json:"description"`
	Brand            string  `json:"brand"`
	Model            string  `json:"model"`
	Notes            string  `json:"notes"`
	CreatedAt        string  `json:"createdAt"`
	UpdatedAt        string  `json:"updatedAt"`
}

type PendingItem struct {
	ID               string `json:"id"`
	TechnicalVisitID string `json:"technicalVisitId"`
	Description      string `json:"description"`
	Priority         string `json:"priority"`
	Responsible      string `json:"responsible"`
	DueAt            string `json:"dueAt"`
	Status           string `json:"status"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
}

var deviceRoles = setOf("inspected", "configured", "installed", "removed", "replaced", "defective", "tested", "updated")
var checklistStatuses = setOf("pending", "completed", "not_applicable")
var pendingPriorities = setOf("low", "normal", "high", "critical")
var pendingStatuses = setOf("pending", "in_progress", "resolved", "cancelled")

func (v *VisitDevice) Validate() error {
	v.DeviceID = strings.TrimSpace(v.DeviceID)
	v.Role = strings.TrimSpace(v.Role)
	if v.DeviceID == "" {
		return ValidationError{"deviceId", "Equipamento é obrigatório"}
	}
	if !deviceRoles[v.Role] {
		return ValidationError{"role", "Papel do equipamento inválido"}
	}
	return validateDetailText(map[string]string{"notes": v.Notes})
}

func (v *VisitService) Validate() error {
	v.Description = strings.TrimSpace(v.Description)
	v.DeviceID = strings.TrimSpace(v.DeviceID)
	if v.Description == "" {
		return ValidationError{"description", "Descrição do serviço é obrigatória"}
	}
	if v.Order < 0 {
		return ValidationError{"order", "Ordem não pode ser negativa"}
	}
	if v.PerformedAt != "" {
		if _, err := parseDate(v.PerformedAt); err != nil {
			return ValidationError{"performedAt", "Horário do serviço inválido"}
		}
	}
	return validateDetailText(map[string]string{"description": v.Description, "category": v.Category, "technician": v.Technician, "notes": v.Notes})
}

func (v *ChecklistItem) Validate() error {
	v.Text = strings.TrimSpace(v.Text)
	if v.Text == "" {
		return ValidationError{"text", "Texto do checklist é obrigatório"}
	}
	if v.Status == "" {
		v.Status = "pending"
	}
	if !checklistStatuses[v.Status] {
		return ValidationError{"status", "Status do checklist inválido"}
	}
	if v.Order < 0 {
		return ValidationError{"order", "Ordem não pode ser negativa"}
	}
	return validateDetailText(map[string]string{"text": v.Text, "notes": v.Notes})
}

func (v *Material) Validate() error {
	v.Unit = strings.TrimSpace(v.Unit)
	v.Description = strings.TrimSpace(v.Description)
	if v.Quantity <= 0 {
		return ValidationError{"quantity", "Quantidade deve ser maior que zero"}
	}
	if v.Unit == "" {
		return ValidationError{"unit", "Unidade é obrigatória"}
	}
	if v.Description == "" {
		return ValidationError{"description", "Descrição do material é obrigatória"}
	}
	return validateDetailText(map[string]string{"unit": v.Unit, "description": v.Description, "brand": v.Brand, "model": v.Model, "notes": v.Notes})
}

func (v *PendingItem) Validate() error {
	v.Description = strings.TrimSpace(v.Description)
	if v.Description == "" {
		return ValidationError{"description", "Descrição da pendência é obrigatória"}
	}
	if v.Priority == "" {
		v.Priority = "normal"
	}
	if v.Status == "" {
		v.Status = "pending"
	}
	if !pendingPriorities[v.Priority] {
		return ValidationError{"priority", "Prioridade inválida"}
	}
	if !pendingStatuses[v.Status] {
		return ValidationError{"status", "Status da pendência inválido"}
	}
	if v.DueAt != "" {
		if _, err := time.Parse("2006-01-02", v.DueAt); err != nil {
			return ValidationError{"dueAt", "Prazo inválido"}
		}
	}
	return validateDetailText(map[string]string{"description": v.Description, "responsible": v.Responsible})
}

func validateDetailText(values map[string]string) error {
	for field, value := range values {
		if len(value) > 4000 {
			return ValidationError{field, "Campo deve ter no máximo 4000 caracteres"}
		}
	}
	return nil
}
