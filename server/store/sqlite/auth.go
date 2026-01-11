package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ksankeerth/open-image-registry/errors/dberrors"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/store"
	"github.com/ksankeerth/open-image-registry/types/models"
)

type authStore struct {
	db *sql.DB
}

func newAuthStore(db *sql.DB) *authStore {
	return &authStore{db: db}
}

func (a *authStore) getQuerier(ctx context.Context) store.Querier {
	if tx, ok := store.TxFromContext(ctx); ok {
		return tx
	}
	return a.db
}

func (a *authStore) RecordTokenRevocation(ctx context.Context, m *models.RevokedToken) error {
	q := a.getQuerier(ctx)

	_, err := q.ExecContext(ctx, RecordTokenRevocationQuery, m.SignatureHash, m.ExpiresAt, m.IssuedAt,
		m.UserID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to record revoked token")
		return dberrors.ClassifyError(err, RecordTokenRevocationQuery)
	}
	return nil
}

func (a *authStore) GetRevokedToken(ctx context.Context, signatureHash string) (m *models.RevokedToken,
	err error) {
	q := a.getQuerier(ctx)

	m = &models.RevokedToken{}

	err = q.QueryRowContext(ctx, GetRevokedTokenQuery, signatureHash).Scan(&m.SignatureHash, &m.ExpiresAt, &m.IssuedAt,
		&m.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to retrive rovked token")
		return nil, dberrors.ClassifyError(err, GetRevokedTokenQuery)
	}

	return m, nil
}
