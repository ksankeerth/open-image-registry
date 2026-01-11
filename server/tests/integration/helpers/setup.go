package helpers

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/ksankeerth/open-image-registry/config"
	"github.com/ksankeerth/open-image-registry/constants"
	"github.com/ksankeerth/open-image-registry/security"
	"github.com/ksankeerth/open-image-registry/store"
	"github.com/ksankeerth/open-image-registry/tests/testdata"
)

func WaitForServer(baseURL string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// TODO: We don't have healthcheck endpoint yet
	healthCheckEndpoint := baseURL + testdata.EndpointHealthCheck

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for server")
		case <-ticker.C:
			resp, err := http.Get(healthCheckEndpoint)
			if err == nil && resp.StatusCode == http.StatusOK {
				resp.Body.Close()
				return nil
			}
			if resp != nil {
				resp.Body.Close()
			}
		}
	}
}

func FindFreePort() uint {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Fatalf("Failed to find free port: %v", err)
	}
	defer listener.Close()

	return uint(listener.Addr().(*net.TCPAddr).Port)
}

func InitializeAdminUserAccount(s store.Store, adminConfig *config.AdminUserAccountConfig) error {
	ctx := context.Background()
	tx, err := s.Begin(ctx)
	if err != nil {
		return err
	}
	ctx = store.WithTxContext(ctx, tx)

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	userAcc, err := s.Users().GetByUsername(ctx, adminConfig.Username)
	if err != nil {
		return err
	}

	if userAcc != nil && userAcc.Locked {
		err = s.Users().UnlockAccount(ctx, adminConfig.Username)
		if err != nil {
			return err
		}
	}

	if userAcc == nil && adminConfig.CreateAccount {
		salt, err := security.GenerateSalt(16)
		if err != nil {
			return err
		}
		passwordHash := security.GeneratePasswordHash(adminConfig.Password, salt)
		userId, err := s.Users().Create(ctx, adminConfig.Username, adminConfig.Email, adminConfig.Username, passwordHash,
			salt)
		if err != nil {
			return err
		}
		err = s.Users().UnlockAccount(ctx, adminConfig.Username)
		if err != nil {
			return err
		}

		err = s.Users().AssignRole(ctx, userId, constants.RoleAdmin)
		if err != nil {
			return err
		}
	}

	return nil
}

func SetAuthCookie(r *http.Request, token string) *http.Request {
	r.AddCookie(&http.Cookie{
		Name:     constants.AuthTokenCookie,
		Value:    token,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
		MaxAge:   900,
	})
	return r
}