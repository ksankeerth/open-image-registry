package store

import (
	"context"

	"github.com/ksankeerth/open-image-registry/types/models"
)

type AuthStore interface {
	RecordTokenRevocation(ctx context.Context, m *models.RevokedToken) error

	GetRevokedToken(ctx context.Context, signatureHash string) (m *models.RevokedToken, err error)
}