package store

import (
	"context"
	"time"

	"github.com/ksankeerth/open-image-registry/types/models"
)

type AuthStore interface {
	// PersistScopeRoleBinding(ctx context.Context, roleName string) error

	// GetAllScopeRoleBindings(ctx context.Context) ([]*models.ScopeRoleBinding, error)

	PersistAuthSession(ctx context.Context, session *models.OAuthSession) error

	GetAuthSession(ctx context.Context, scopeHash, userId string) (*models.OAuthSession, error)

	PersistAuthSessionScopeBinding(ctx context.Context, scopes []string, sessionId string) error

	RemoveAuthSession(ctx context.Context, sessionId string) error

	UpdateSessionLastAccess(ctx context.Context, sesssionId string, lastAccessed time.Time) error
}
