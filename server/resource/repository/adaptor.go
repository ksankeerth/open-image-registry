package repository

import (
	"github.com/ksankeerth/open-image-registry/types/api/v1alpha/mgmt"
	"github.com/ksankeerth/open-image-registry/types/models"
)

func ToRepositoryViewDTO(view *models.RepositoryView) *mgmt.RepositoryViewDTO {
	if view == nil {
		return nil
	}

	return &mgmt.RepositoryViewDTO{
		RegistryID:  view.RegistryID,
		NamespaceID: view.NamespaceID,
		ID:          view.ID,
		Namespace:   view.Namespace,
		Name:        view.Name,
		Description: view.Description,
		State:       view.State,
		IsPublic:    view.IsPublic,
		CreatedBy:   view.CreatedBy,
		TagsCount:   view.TagsCount,
		CreatedAt:   view.CreatedAt,
		UpdatedAt:   view.UpdatedAt,
	}
}

func makeGetRepositoryResponse(m *models.RepositoryModel) *mgmt.RepositoryResponse {
	if m == nil {
		return &mgmt.RepositoryResponse{}
	}

	return &mgmt.RepositoryResponse{
		ID:          m.ID,
		RegistryID:  m.RegistryID,
		NamespaceID: m.NamespaceID,
		Name:        m.Name,
		Description: m.Description,
		IsPublic:    m.IsPublic,
		State:       m.State,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}