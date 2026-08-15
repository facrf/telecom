package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/local/telecom/internal/technicalvisits"
)

type technicalVisitHandlers struct {
	service *technicalvisits.Service
}

func (h technicalVisitHandlers) list(w http.ResponseWriter, r *http.Request) {
	filters := technicalvisits.Filters{
		Query:      strings.TrimSpace(r.URL.Query().Get("q")),
		ClientID:   firstNonEmpty(chi.URLParam(r, "clientID"), r.URL.Query().Get("client_id")),
		ProjectID:  firstNonEmpty(chi.URLParam(r, "projectID"), r.URL.Query().Get("project_id")),
		Technician: strings.TrimSpace(r.URL.Query().Get("technician")),
		Status:     r.URL.Query().Get("status"),
		Result:     r.URL.Query().Get("result"),
		DateFrom:   r.URL.Query().Get("date_from"),
		DateTo:     r.URL.Query().Get("date_to"),
	}
	visits, err := h.service.List(r.Context(), filters)
	if err != nil {
		serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": visits})
}

func (h technicalVisitHandlers) get(w http.ResponseWriter, r *http.Request) {
	visit, err := h.service.Get(r.Context(), chi.URLParam(r, "visitID"))
	if !respondTechnicalVisitError(w, err) {
		writeJSON(w, http.StatusOK, visit)
	}
}

func (h technicalVisitHandlers) create(w http.ResponseWriter, r *http.Request) {
	var visit technicalvisits.Visit
	if !decodeJSON(w, r, &visit) {
		return
	}
	created, err := h.service.Create(r.Context(), visit)
	if !respondTechnicalVisitError(w, err) {
		writeJSON(w, http.StatusCreated, created)
	}
}

func (h technicalVisitHandlers) update(w http.ResponseWriter, r *http.Request) {
	var visit technicalvisits.Visit
	if !decodeJSON(w, r, &visit) {
		return
	}
	updated, err := h.service.Update(r.Context(), chi.URLParam(r, "visitID"), visit)
	if !respondTechnicalVisitError(w, err) {
		writeJSON(w, http.StatusOK, updated)
	}
}

func (h technicalVisitHandlers) delete(w http.ResponseWriter, r *http.Request) {
	err := h.service.Delete(r.Context(), chi.URLParam(r, "visitID"))
	if !respondTechnicalVisitError(w, err) {
		w.WriteHeader(http.StatusNoContent)
	}
}

func respondTechnicalVisitError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	var validation technicalvisits.ValidationError
	switch {
	case errors.As(err, &validation):
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": map[string]string{"code": "VALIDATION_ERROR", "message": validation.Message, "field": validation.Field}})
	case errors.Is(err, technicalvisits.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Visita técnica não encontrada")
	case errors.Is(err, technicalvisits.ErrItemNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Item da visita técnica não encontrado")
	case errors.Is(err, technicalvisits.ErrProjectNotFound):
		writeError(w, http.StatusUnprocessableEntity, "INVALID_PROJECT", "Projeto não encontrado")
	case errors.Is(err, technicalvisits.ErrInvalidDeviceProject):
		writeError(w, http.StatusUnprocessableEntity, "INVALID_DEVICE", "O equipamento não pertence ao projeto da visita")
	case errors.Is(err, technicalvisits.ErrDuplicateDeviceRole):
		writeError(w, http.StatusConflict, "DUPLICATE_DEVICE_ROLE", "O equipamento já possui esse papel na visita")
	case errors.Is(err, technicalvisits.ErrConflict):
		writeError(w, http.StatusConflict, "CONCURRENT_UPDATE", "A visita foi alterada em outra sessão. Recarregue antes de salvar.")
	default:
		serverError(w, err)
	}
	return true
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
