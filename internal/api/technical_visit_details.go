package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/local/telecom/internal/technicalvisits"
)

func (h technicalVisitHandlers) detailRoutes(r chi.Router) {
	r.Route("/devices", func(r chi.Router) {
		r.Get("/", h.listVisitDevices)
		r.Post("/", h.saveVisitDevice)
		r.Put("/{itemID}", h.saveVisitDevice)
		r.Delete("/{itemID}", h.deleteVisitDevice)
	})
	r.Route("/services", func(r chi.Router) {
		r.Get("/", h.listVisitServices)
		r.Post("/", h.saveVisitService)
		r.Put("/{itemID}", h.saveVisitService)
		r.Delete("/{itemID}", h.deleteVisitService)
	})
	r.Route("/checklist", func(r chi.Router) {
		r.Get("/", h.listVisitChecklist)
		r.Post("/", h.saveVisitChecklist)
		r.Put("/{itemID}", h.saveVisitChecklist)
		r.Delete("/{itemID}", h.deleteVisitChecklist)
	})
	r.Route("/materials", func(r chi.Router) {
		r.Get("/", h.listVisitMaterials)
		r.Post("/", h.saveVisitMaterial)
		r.Put("/{itemID}", h.saveVisitMaterial)
		r.Delete("/{itemID}", h.deleteVisitMaterial)
	})
	r.Route("/pending-items", func(r chi.Router) {
		r.Get("/", h.listVisitPendingItems)
		r.Post("/", h.saveVisitPendingItem)
		r.Put("/{itemID}", h.saveVisitPendingItem)
		r.Delete("/{itemID}", h.deleteVisitPendingItem)
	})
}

func (h technicalVisitHandlers) listVisitDevices(w http.ResponseWriter, r *http.Request) {
	values, err := h.service.ListDevices(r.Context(), chi.URLParam(r, "visitID"))
	writeVisitItems(w, values, err)
}
func (h technicalVisitHandlers) saveVisitDevice(w http.ResponseWriter, r *http.Request) {
	var item technicalvisits.VisitDevice
	if !decodeJSON(w, r, &item) {
		return
	}
	item.ID = chi.URLParam(r, "itemID")
	saved, err := h.service.SaveDevice(r.Context(), chi.URLParam(r, "visitID"), item)
	writeVisitItem(w, r, saved, err)
}
func (h technicalVisitHandlers) deleteVisitDevice(w http.ResponseWriter, r *http.Request) {
	writeVisitDelete(w, h.service.DeleteDevice(r.Context(), chi.URLParam(r, "visitID"), chi.URLParam(r, "itemID")))
}

func (h technicalVisitHandlers) listVisitServices(w http.ResponseWriter, r *http.Request) {
	values, err := h.service.ListServices(r.Context(), chi.URLParam(r, "visitID"))
	writeVisitItems(w, values, err)
}
func (h technicalVisitHandlers) saveVisitService(w http.ResponseWriter, r *http.Request) {
	var item technicalvisits.VisitService
	if !decodeJSON(w, r, &item) {
		return
	}
	item.ID = chi.URLParam(r, "itemID")
	saved, err := h.service.SaveService(r.Context(), chi.URLParam(r, "visitID"), item)
	writeVisitItem(w, r, saved, err)
}
func (h technicalVisitHandlers) deleteVisitService(w http.ResponseWriter, r *http.Request) {
	writeVisitDelete(w, h.service.DeleteService(r.Context(), chi.URLParam(r, "visitID"), chi.URLParam(r, "itemID")))
}

func (h technicalVisitHandlers) listVisitChecklist(w http.ResponseWriter, r *http.Request) {
	values, err := h.service.ListChecklist(r.Context(), chi.URLParam(r, "visitID"))
	writeVisitItems(w, values, err)
}
func (h technicalVisitHandlers) saveVisitChecklist(w http.ResponseWriter, r *http.Request) {
	var item technicalvisits.ChecklistItem
	if !decodeJSON(w, r, &item) {
		return
	}
	item.ID = chi.URLParam(r, "itemID")
	saved, err := h.service.SaveChecklist(r.Context(), chi.URLParam(r, "visitID"), item)
	writeVisitItem(w, r, saved, err)
}
func (h technicalVisitHandlers) deleteVisitChecklist(w http.ResponseWriter, r *http.Request) {
	writeVisitDelete(w, h.service.DeleteChecklist(r.Context(), chi.URLParam(r, "visitID"), chi.URLParam(r, "itemID")))
}

func (h technicalVisitHandlers) listVisitMaterials(w http.ResponseWriter, r *http.Request) {
	values, err := h.service.ListMaterials(r.Context(), chi.URLParam(r, "visitID"))
	writeVisitItems(w, values, err)
}
func (h technicalVisitHandlers) saveVisitMaterial(w http.ResponseWriter, r *http.Request) {
	var item technicalvisits.Material
	if !decodeJSON(w, r, &item) {
		return
	}
	item.ID = chi.URLParam(r, "itemID")
	saved, err := h.service.SaveMaterial(r.Context(), chi.URLParam(r, "visitID"), item)
	writeVisitItem(w, r, saved, err)
}
func (h technicalVisitHandlers) deleteVisitMaterial(w http.ResponseWriter, r *http.Request) {
	writeVisitDelete(w, h.service.DeleteMaterial(r.Context(), chi.URLParam(r, "visitID"), chi.URLParam(r, "itemID")))
}

func (h technicalVisitHandlers) listVisitPendingItems(w http.ResponseWriter, r *http.Request) {
	values, err := h.service.ListPendingItems(r.Context(), chi.URLParam(r, "visitID"))
	writeVisitItems(w, values, err)
}
func (h technicalVisitHandlers) saveVisitPendingItem(w http.ResponseWriter, r *http.Request) {
	var item technicalvisits.PendingItem
	if !decodeJSON(w, r, &item) {
		return
	}
	item.ID = chi.URLParam(r, "itemID")
	saved, err := h.service.SavePendingItem(r.Context(), chi.URLParam(r, "visitID"), item)
	writeVisitItem(w, r, saved, err)
}
func (h technicalVisitHandlers) deleteVisitPendingItem(w http.ResponseWriter, r *http.Request) {
	writeVisitDelete(w, h.service.DeletePendingItem(r.Context(), chi.URLParam(r, "visitID"), chi.URLParam(r, "itemID")))
}

func writeVisitItems(w http.ResponseWriter, values any, err error) {
	if respondTechnicalVisitError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": values})
}
func writeVisitItem(w http.ResponseWriter, r *http.Request, value any, err error) {
	if respondTechnicalVisitError(w, err) {
		return
	}
	status := http.StatusOK
	if r.Method == http.MethodPost {
		status = http.StatusCreated
	}
	writeJSON(w, status, value)
}
func writeVisitDelete(w http.ResponseWriter, err error) {
	if respondTechnicalVisitError(w, err) {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
