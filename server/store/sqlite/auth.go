package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ksankeerth/open-image-registry/errors/dberrors"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/store"
	"github.com/ksankeerth/open-image-registry/types/models"
	"github.com/ksankeerth/open-image-registry/utils"
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

func (a *authStore) PersistAuthSession(ctx context.Context, session *models.OAuthSession) error {
	q := a.getQuerier(ctx)

	_, err := q.ExecContext(ctx, AuthPersistSessionQuery, session.ID, session.UserID, session.ScopeHash, session.IssuedAt,
		session.ExpiresAt, session.UserAgent, session.ClientIP, session.GrantType)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to persist auth session")
		return dberrors.ClassifyError(err, AuthPersistSessionQuery)
	}

	return nil
}

func (a *authStore) GetAuthSession(ctx context.Context, scopeHash, userID string) (*models.OAuthSession, error) {
	var m models.OAuthSession
	var issuedAt, expiresAt, lastAccessesAt string

	q := a.getQuerier(ctx)

	err := q.QueryRowContext(ctx, AuthSessionGetQuery, scopeHash, userID).Scan(&m.ID, &m.UserAgent, &m.ClientIP,
		&m.GrantType, &issuedAt, &expiresAt, &lastAccessesAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		log.Logger().Error().Err(err).Msg("failed to retrieve auth session")
		return nil, dberrors.ClassifyError(err, AuthSessionGetQuery)
	}

	if issuedAt != "" {
		issuedTime, err := utils.ParseSqliteTimestamp(issuedAt)
		if err != nil {
			log.Logger().Error().Err(err).Msg("failed to parse sqlite timestamp")
			return nil, dberrors.ClassifyError(err, AuthSessionGetQuery)
		}
		m.IssuedAt = *issuedTime
	}

	if expiresAt != "" {
		expiresTime, err := utils.ParseSqliteTimestamp(expiresAt)
		if err != nil {
			log.Logger().Error().Err(err).Msg("failed to parse sqlite timestamp")
			return nil, dberrors.ClassifyError(err, AuthSessionGetQuery)
		}
		m.ExpiresAt = expiresTime
	}

	if lastAccessesAt != "" {
		lastAccessedTime, err := utils.ParseSqliteTimestamp(lastAccessesAt)
		if err != nil {
			log.Logger().Error().Err(err).Msg("failed to parse sqlite timestamp")
			return nil, dberrors.ClassifyError(err, AuthSessionGetQuery)
		}
		m.LastAccessedAt = lastAccessedTime
	}

	return &m, nil
}

func (a *authStore) PersistAuthSessionScopeBinding(ctx context.Context, scopes []string, sessionID string) error {
	q := a.getQuerier(ctx)

	stmt, err := q.PrepareContext(ctx, AuthPersistSessionScopeBindingsQuery)
	if err != nil {
		log.Logger().Error().Err(err).Msg("failed to prepare scope binding statement")
		return dberrors.ClassifyError(err, AuthPersistSessionScopeBindingsQuery)
	}
	defer stmt.Close()

	// Insert each scope binding
	for _, scope := range scopes {
		_, err := stmt.ExecContext(ctx, sessionID, scope)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("failed to persist scope binding for session: %s, scope: %s", sessionID, scope)
			return dberrors.ClassifyError(err, AuthPersistSessionScopeBindingsQuery)
		}
	}

	return nil
}

func (a *authStore) RemoveAuthSession(ctx context.Context, sessionID string) error {
	q := a.getQuerier(ctx)

	_, err := q.ExecContext(ctx, AuthRemoveSessionQuery, sessionID)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("failed to remove auth session: %s", sessionID)
		return dberrors.ClassifyError(err, AuthRemoveSessionQuery)
	}

	return nil
}

func (a *authStore) UpdateSessionLastAccess(ctx context.Context, sessionID string, lastAccessed time.Time) error {
	q := a.getQuerier(ctx)

	_, err := q.ExecContext(ctx, AuthUpdateSessionQuery, lastAccessed, sessionID)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("failed to update session last accessed time: %s", sessionID)
		return dberrors.ClassifyError(err, AuthUpdateSessionQuery)
	}

	return nil
}

// func (a *authStore) GetAllScopeRoleBindings(ctx context.Context) ([]*models.ScopeRoleBinding, error) {
// 	q := a.getQuerier(ctx)

// 	rows, err := q.QueryContext(ctx, AuthGetAllScopeRoleBindingsQuery)
// 	if err != nil {
// 		log.Logger().Error().Err(err).Msg("failed to retrive scope-role bindings")
// 		return nil, dberrors.ClassifyError(err, AuthGetAllScopeRoleBindingsQuery)
// 	}

// 	bindings := make([]*models.ScopeRoleBinding, 0)

// 	for rows.Next() {
// 		var binding models.ScopeRoleBinding

// 		err = rows.Scan(&binding.ScopeName, &binding.RoleName)
// 		if err != nil {
// 			log.Logger().Error().Err(err).Msg("failed to read scope binding results")
// 			return nil, dberrors.ClassifyError(err, AuthGetAllScopeRoleBindingsQuery)
// 		}
// 		bindings = append(bindings, &binding)
// 	}

// 	return bindings, nil
// }
