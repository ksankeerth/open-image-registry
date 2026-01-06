package sqlite

import (
	"context"
	"database/sql"

	"github.com/ksankeerth/open-image-registry/errors/dberrors"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/store"
	"github.com/ksankeerth/open-image-registry/types/models"
	"github.com/ksankeerth/open-image-registry/utils"
)

type accountRecoveryStore struct {
	db *sql.DB
}

func (r *accountRecoveryStore) getQuerier(ctx context.Context) store.Querier {
	if tx, ok := store.TxFromContext(ctx); ok {
		return tx
	}
	return r.db
}

func newAccountRecoveryStore(db *sql.DB) *accountRecoveryStore {
	return &accountRecoveryStore{db: db}
}

func (r *accountRecoveryStore) Create(ctx context.Context, userId, uuid string, reason int) error {
	q := r.getQuerier(ctx)

	_, err := q.ExecContext(ctx, AccountRecoveryCreateQuery, uuid, userId, reason)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to create account recovery record")
		return dberrors.ClassifyError(err, AccountRecoveryCreateQuery)
	}

	return nil
}

func (r *accountRecoveryStore) Get(ctx context.Context, uuid string) (*models.AccountRecovery, error) {
	q := r.getQuerier(ctx)

	row := q.QueryRowContext(ctx, AccountRecoveryGetQuery, uuid)

	var createdAt string
	var m models.AccountRecovery

	err := row.Scan(&m.RecoveryID, &m.UserID, &m.ReasonType, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // not found
		}
		log.Logger().Error().Err(err).Msg("failed to get account recovery record")
		return nil, dberrors.ClassifyError(err, AccountRecoveryGetQuery)
	}

	t, err := utils.ParseSqliteTimestamp(createdAt)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to parse created_at timestamp")
		return nil, dberrors.ClassifyError(err, AccountRecoveryGetQuery)
	}
	m.CreatedAt = *t

	return &m, nil
}

func (r *accountRecoveryStore) GetByUserID(ctx context.Context, userId string) (*models.AccountRecovery, error) {
	q := r.getQuerier(ctx)

	row := q.QueryRowContext(ctx, AccountRecoveryGetByUserIDQuery, userId)

	var createdAt string
	var m models.AccountRecovery

	err := row.Scan(&m.RecoveryID, &m.UserID, &m.ReasonType, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // not found
		}
		log.Logger().Error().Err(err).Msg("failed to get account recovery by user id")
		return nil, dberrors.ClassifyError(err, AccountRecoveryGetByUserIDQuery)
	}

	t, err := utils.ParseSqliteTimestamp(createdAt)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to parse created_at timestamp")
		return nil, dberrors.ClassifyError(err, AccountRecoveryGetByUserIDQuery)
	}
	m.CreatedAt = *t

	return &m, nil
}

func (r *accountRecoveryStore) Delete(ctx context.Context, uuid string) error {
	q := r.getQuerier(ctx)

	_, err := q.ExecContext(ctx, AccountRecoveeryDeleteQuery, uuid)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to delete account recovery record")
		return dberrors.ClassifyError(err, AccountRecoveeryDeleteQuery)
	}

	return nil
}

func (r *accountRecoveryStore) DeleteByUserID(ctx context.Context, userId string) error {
	q := r.getQuerier(ctx)

	_, err := q.ExecContext(ctx, AccountRecoveryDeleteByUserIDQuery, userId)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to delete account recovery by user id")
		return dberrors.ClassifyError(err, AccountRecoveryDeleteByUserIDQuery)
	}

	return nil
}

func (r *accountRecoveryStore) UpdateReason(ctx context.Context, recoveryID string, reason uint) error {
	q := r.getQuerier(ctx)

	_, err := q.ExecContext(ctx, AccountRecoveryUpdateReasonQuery, reason, recoveryID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("faile dto update account recovery reason")
		return dberrors.ClassifyError(err, AccountRecoveryUpdateReasonQuery)
	}

	return nil
}
