package technicalvisits

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

type Service struct{ repository *Repository }

func NewService(repository *Repository) *Service { return &Service{repository: repository} }
func (s *Service) List(ctx context.Context, filters Filters) ([]Visit, error) {
	return s.repository.List(ctx, filters)
}
func (s *Service) Get(ctx context.Context, id string) (Visit, error) {
	return s.repository.Get(ctx, id)
}
func (s *Service) Create(ctx context.Context, v Visit) (Visit, error) {
	if v.Status == "" {
		v.Status = StatusDraft
	}
	if err := v.Validate(); err != nil {
		return Visit{}, err
	}
	v.ID = newID()
	return s.repository.Create(ctx, v)
}
func (s *Service) Update(ctx context.Context, id string, v Visit) (Visit, error) {
	v.ID = id
	if v.UpdatedAt == "" {
		return Visit{}, ValidationError{"updatedAt", "Versão do registro é obrigatória"}
	}
	if err := v.Validate(); err != nil {
		return Visit{}, err
	}
	return s.repository.Update(ctx, v)
}
func (s *Service) Delete(ctx context.Context, id string) error { return s.repository.Delete(ctx, id) }

func (s *Service) ListDevices(ctx context.Context, visitID string) ([]VisitDevice, error) {
	return s.repository.ListDevices(ctx, visitID)
}
func (s *Service) SaveDevice(ctx context.Context, visitID string, item VisitDevice) (VisitDevice, error) {
	item.TechnicalVisitID = visitID
	if err := item.Validate(); err != nil {
		return VisitDevice{}, err
	}
	if item.ID == "" {
		item.ID = newID()
		return s.repository.CreateDevice(ctx, item)
	}
	return s.repository.UpdateDevice(ctx, item)
}
func (s *Service) DeleteDevice(ctx context.Context, visitID, id string) error {
	return s.repository.DeleteDevice(ctx, visitID, id)
}

func (s *Service) ListServices(ctx context.Context, visitID string) ([]VisitService, error) {
	return s.repository.ListServices(ctx, visitID)
}
func (s *Service) SaveService(ctx context.Context, visitID string, item VisitService) (VisitService, error) {
	item.TechnicalVisitID = visitID
	if err := item.Validate(); err != nil {
		return VisitService{}, err
	}
	if item.ID == "" {
		item.ID = newID()
		return s.repository.CreateService(ctx, item)
	}
	return s.repository.UpdateService(ctx, item)
}
func (s *Service) DeleteService(ctx context.Context, visitID, id string) error {
	return s.repository.DeleteService(ctx, visitID, id)
}

func (s *Service) ListChecklist(ctx context.Context, visitID string) ([]ChecklistItem, error) {
	return s.repository.ListChecklist(ctx, visitID)
}
func (s *Service) SaveChecklist(ctx context.Context, visitID string, item ChecklistItem) (ChecklistItem, error) {
	item.TechnicalVisitID = visitID
	if err := item.Validate(); err != nil {
		return ChecklistItem{}, err
	}
	if item.ID == "" {
		item.ID = newID()
		return s.repository.CreateChecklist(ctx, item)
	}
	return s.repository.UpdateChecklist(ctx, item)
}
func (s *Service) DeleteChecklist(ctx context.Context, visitID, id string) error {
	return s.repository.DeleteChecklist(ctx, visitID, id)
}

func (s *Service) ListMaterials(ctx context.Context, visitID string) ([]Material, error) {
	return s.repository.ListMaterials(ctx, visitID)
}
func (s *Service) SaveMaterial(ctx context.Context, visitID string, item Material) (Material, error) {
	item.TechnicalVisitID = visitID
	if err := item.Validate(); err != nil {
		return Material{}, err
	}
	if item.ID == "" {
		item.ID = newID()
		return s.repository.CreateMaterial(ctx, item)
	}
	return s.repository.UpdateMaterial(ctx, item)
}
func (s *Service) DeleteMaterial(ctx context.Context, visitID, id string) error {
	return s.repository.DeleteMaterial(ctx, visitID, id)
}

func (s *Service) ListPendingItems(ctx context.Context, visitID string) ([]PendingItem, error) {
	return s.repository.ListPendingItems(ctx, visitID)
}
func (s *Service) SavePendingItem(ctx context.Context, visitID string, item PendingItem) (PendingItem, error) {
	item.TechnicalVisitID = visitID
	if err := item.Validate(); err != nil {
		return PendingItem{}, err
	}
	if item.ID == "" {
		item.ID = newID()
		return s.repository.CreatePendingItem(ctx, item)
	}
	return s.repository.UpdatePendingItem(ctx, item)
}
func (s *Service) DeletePendingItem(ctx context.Context, visitID, id string) error {
	return s.repository.DeletePendingItem(ctx, visitID, id)
}
func newID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		panic(err)
	}
	return hex.EncodeToString(value)
}
