package namespace

import (
	"github.com/ksankeerth/open-image-registry/types/api/v1alpha/mgmt"
	"github.com/ksankeerth/open-image-registry/types/models"
)

func makeGetNamespaceResponse(m *models.NamespaceModel) *mgmt.NamespaceResponse {
	if m == nil {
		return nil
	}

	return &mgmt.NamespaceResponse{
		ID:          m.Id,
		RegistryID:  m.RegistryId,
		Name:        m.Name,
		Description: m.Description,
		Purpose:     m.Purpose,
		IsPublic:    m.IsPublic,
		State:       m.State,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func toNamespaceViewDTO(view *models.NamespaceView) *mgmt.NamespaceViewDTO {
	if view == nil {
		return nil
	}

	return &mgmt.NamespaceViewDTO{
		RegistryID:  view.RegistryID,
		ID:          view.ID,
		Name:        view.Name,
		Description: view.Description,
		State:       view.State,
		IsPublic:    view.IsPublic,
		Purpose:     view.Purpose,
		CreatedAt:   view.CreatedAt,
		UpdatedAt:   view.UpdatedAt,
		Developers:  view.Developers,
		Maintainers: view.Maintainers,
	}
}