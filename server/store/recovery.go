package store

import (
	"context"

	"github.com/ksankeerth/open-image-registry/types/models"
)

type AccountRecoveryStore interface {
	Create(ctx context.Context, userId, uuid string, reason int) error

	Get(ctx context.Context, uuid string) (*models.AccountRecovery, error)

	GetByUserID(ctx context.Context, userId string) (*models.AccountRecovery, error)

	Delete(ctx context.Context, uuid string) (err error)

	DeleteByUserID(ctx context.Context, userId string) (err error)

	UpdateReason(ctx context.Context, recoveryID string, reason uint) error 
}
