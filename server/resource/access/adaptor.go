package access

import (
	"github.com/ksankeerth/open-image-registry/types/api/v1alpha/mgmt"
	"github.com/ksankeerth/open-image-registry/types/models"
)

func ToResourceAccessViewDTO(view *models.ResourceAccessView) *mgmt.ResourceAccessViewDTO {
	if view == nil {
		return nil
	}

	return &mgmt.ResourceAccessViewDTO{
		ID:           view.ID,
		ResourceType: view.ResourceType,
		ResourceName: view.ResourceName,
		ResourceID:   view.ResourceID,
		AccessLevel:  view.AccessLevel,
		UserId:       view.UserId,
		Username:     view.Username,
		GrantedBy:    view.GrantedBy,
		GrantedUser:  view.GrantedUser,
		GrantedAt:    view.GrantedAt,
	}
}