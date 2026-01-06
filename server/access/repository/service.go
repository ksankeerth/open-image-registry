package repository

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ksankeerth/open-image-registry/access/resource"
	"github.com/ksankeerth/open-image-registry/constants"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/store"
	"github.com/ksankeerth/open-image-registry/types/api/v1alpha/mgmt"
	"github.com/ksankeerth/open-image-registry/types/models"
)

type repositoryService struct {
	store         store.Store
	accessManager *resource.Manager
}

type createRepoResult struct {
	conflict  bool
	invalidNs bool
	id        string
}

func (svc *repositoryService) createRepository(reqCtx context.Context, req *mgmt.CreateRepositoryRequest) (*createRepoResult,
	error) {

	tx, err := svc.store.Begin(reqCtx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to create repository due to transaction errors")
		return nil, err
	}

	ctx := store.WithTxContext(reqCtx, tx)
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	res := &createRepoResult{}

	nsExists, err := svc.store.Namespaces().ExistsByIdentifier(ctx, constants.HostedRegistryID, req.NamespaceId)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Creating repository failed due to database errrors: %s:%s", req.NamespaceId, req.Name)
		return nil, err
	}

	if !nsExists {
		res.invalidNs = true
		log.Logger().Warn().Msgf("Not allowed to create repository with invalid namespace: %s", req.NamespaceId)
		return res, nil
	}

	repoExists, err := svc.store.Repositories().ExistsByIdentifier(ctx, req.NamespaceId, req.Name)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Creating repository failed due to database errrors: %s:%s", req.NamespaceId, req.Name)
		return nil, err
	}

	if repoExists {
		res.conflict = true
		log.Logger().Warn().Msgf("Not allowed to create another repository with same identifier: %s:%s", req.NamespaceId, req.Name)
		return res, nil
	}

	id, err := svc.store.Repositories().Create(ctx, constants.HostedRegistryID, req.NamespaceId, req.Name, req.Description, req.IsPublic, req.CreatedBy)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Creating repository failed due to database errrors: %s:%s", req.NamespaceId, req.Name)
		return nil, err
	}

	res.id = id

	return res, nil
}

func (svc *repositoryService) listRepositories(reqCtx context.Context, cond *store.ListQueryConditions) (repositories []*models.RepositoryView,
	total int, err error) {
	tx, err := svc.store.Begin(reqCtx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("error occurred when starting transaction")
		return nil, -1, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	ctx := store.WithTxContext(reqCtx, tx)

	repositories, total, err = svc.store.Repositories().List(ctx, cond)
	return
}

func (svc *repositoryService) repositoryExists(reqCtx context.Context, identifier string, namespaceId string) (exits bool, err error) {

	if namespaceId == "" {
		exits, err = svc.store.Repositories().Exists(reqCtx, identifier)
	} else {
		exits, err = svc.store.Repositories().ExistsByIdentifier(reqCtx, identifier, namespaceId)
	}

	return
}

func (svc *repositoryService) getRepository(reqCtx context.Context, identifier, namespaceId string) (*models.RepositoryModel, error) {
	if namespaceId == "" {
		return svc.store.Repositories().Get(reqCtx, identifier)
	} else {
		return svc.store.Repositories().GetByIdentifier(reqCtx, identifier, namespaceId)
	}
}

func (svc *repositoryService) deleteRepository(reqCtx context.Context, identifier, namespaceId string) error {
	if namespaceId == "" {
		return svc.store.Repositories().Delete(reqCtx, identifier)
	} else {
		return svc.store.Repositories().DeleteByIdentifier(reqCtx, identifier, namespaceId)
	}
}

func (svc *repositoryService) updateRepsitory(reqCtx context.Context, id string, req *mgmt.UpdateRepositoryRequest) (notFound bool, err error) {
	tx, err := svc.store.Begin(reqCtx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Updating repository failed due to dabase errors")
		return false, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	ctx := store.WithTxContext(reqCtx, tx)

	exists, err := svc.store.Repositories().Exists(ctx, id)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Updating repository failed due to database errors: %s", id)
		return false, err
	}

	if !exists {
		return true, nil
	}

	err = svc.store.Repositories().Update(ctx, id, req.Description)
	return
}

type patchResult struct {
	httpStatusCode int
	httpErrorMsg   string
	success        bool
}

func (svc *repositoryService) changeState(reqCtx context.Context, id string, newState string) (result *patchResult,
	err error) {
	result = &patchResult{}
	tx, err := svc.store.Begin(reqCtx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to change state of namespace due to transaction errors")
		return nil, err
	}

	ctx := store.WithTxContext(reqCtx, tx)

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	repo, err := svc.store.Repositories().Get(ctx, id)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to change state of repository due to database errors")
		return nil, err
	}

	if repo == nil {
		log.Logger().Warn().Msgf("Failed to change state of non existent repository: %s", id)
		result.httpErrorMsg = "Repository " + id + " is not found"
		result.httpStatusCode = http.StatusNotFound

		return result, nil
	}

	if repo.State == newState {
		log.Logger().Debug().Msgf("No changes in state. Updating state of repository(%s) is skipped", id)
		result.success = true
		return nil, nil
	}

	if repo.State == constants.ResourceStateActive && newState == constants.ResourceStateDisabled {
		log.Logger().Warn().Msgf("Not allowed to change repository(%s) state from 'Active' to 'Disabled'", id)
		result.httpErrorMsg = "Not allowed to change repository state from 'Active' to 'Disabled'"
		result.httpStatusCode = http.StatusForbidden
		return result, nil
	}

	ns, err := svc.store.Namespaces().Get(ctx, repo.NamespaceID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to change state of namespace due to database errors")
		return nil, err
	}

	if ns == nil {
		log.Logger().Error().Msgf("Failed to change state of repository due to non-existent namespace")
		// this must be an internal error. because due to the FK constraints it should not be allowed to have a repository
		// without a namespace
		result.httpStatusCode = http.StatusInternalServerError
		return result, fmt.Errorf("invalid repository without namespace")
	}

	if ns.State == constants.ResourceStateDisabled {
		result.success = false
		result.httpErrorMsg = "Not allowed to change state of a repository when namespace is in disabled state"
		result.httpStatusCode = http.StatusForbidden
		log.Logger().Warn().Msgf("Not allowed to change state of a repository when namespace is in disabled state: %s", id)
		return
	}

	if ns.State == constants.ResourceStateDeprecated && newState == constants.ResourceStateActive {
		result.success = false
		result.httpErrorMsg = "Not allowed to change state of a repository to 'active' when namespace is in deprecated state"
		result.httpStatusCode = http.StatusForbidden
		log.Logger().Warn().Msgf("Not allowed to change state of a repository to 'active' when namespace is in deprecated state: %s", id)
		return
	}

	err = svc.store.Repositories().SetState(ctx, id, newState)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Failed to change state of repository(%s) to %s", id, newState)
		return nil, err
	}

	result.success = true
	return result, nil
}

func (svc *repositoryService) changeVisiblity(reqCtx context.Context, id string, public bool) (result *patchResult,
	err error) {
	tx, err := svc.store.Begin(reqCtx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to change visibility of repository due to transaction errors")
		return nil, err
	}

	ctx := store.WithTxContext(reqCtx, tx)

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	repo, err := svc.store.Repositories().Get(ctx, id)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to change state of repository due to database errors")
		return nil, err
	}

	if repo == nil {
		log.Logger().Warn().Msgf("Failed to change visibility of non existent repository: %s", id)
		result.httpErrorMsg = "Repository " + id + " is not found"
		result.httpStatusCode = http.StatusNotFound

		return result, nil
	}

	if repo.IsPublic == public {
		log.Logger().Debug().Msgf("No changes in visibility. Updating visibility of repository(%s) is skipped", id)
		result.success = true
		return result, nil
	}

	ns, err := svc.store.Namespaces().Get(ctx, repo.NamespaceID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to change visibility of repository due to database errors")
		return nil, err
	}

	if ns == nil {
		log.Logger().Error().Msgf("Failed to change visibility of repository due to non-existent namespace")
		// this must be an internal error. because due to the FK constraints it should not be allowed to have a repository
		// without a namespace
		result.httpStatusCode = http.StatusInternalServerError
		return result, fmt.Errorf("invalid repository without namespace")
	}

	if ns.State == constants.ResourceStateDisabled || repo.State == constants.ResourceStateDisabled {
		result.success = false
		result.httpErrorMsg = "Not allowed to change visibility of a repository when namespace or repository is in disabled state"
		result.httpStatusCode = http.StatusForbidden
		log.Logger().Warn().Msgf("Not allowed to change state of a repository when namespace or visibility is in disabled state: %s", id)
		return
	}

	err = svc.store.Repositories().SetVisibility(ctx, id, public)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Failed to change visibility of repository(%s) to public=%t", id, public)
		return nil, err
	}

	result.success = true
	return result, nil
}

type grantAccessResult struct {
	conflict            bool // conflict with existing resource access
	userNotFound        bool
	grantedUserNotFound bool
	resourceNotFound    bool
	bodyURLMismatch     bool // true if id in url doesn't match with ID provided in request
}

func (svc *repositoryService) grantAccess(reqCtx context.Context, id string, req *mgmt.AccessGrantRequest) (result *grantAccessResult, err error) {
	result = &grantAccessResult{}
	if id != req.ResourceID {
		result.bodyURLMismatch = true
		return result, nil
	}

	tx, err := svc.store.Begin(reqCtx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Granting repository access to user failed to due to transaction errors")
		return nil, err
	}

	ctx := store.WithTxContext(reqCtx, tx)

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	user, err := svc.store.Users().Get(ctx, req.UserID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Granting repository access failed due to database errors")
		return nil, err
	}

	if user == nil {
		result.userNotFound = true
		return result, nil
	}

	grantedUser, err := svc.store.Users().GetByUsername(ctx, req.GrantedBy)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Granting repository access failed due to database errors")
		return nil, err
	}

	if grantedUser == nil {
		result.grantedUserNotFound = true
		return result, nil
	}

	ns, err := svc.store.Repositories().Get(ctx, req.ResourceID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Granting repository access failed due to database errors")
		return nil, err
	}
	if ns == nil {
		result.resourceNotFound = true
		return result, nil
	}

	access, err := svc.store.Access().GetUserAccess(ctx, req.ResourceID, req.ResourceType, req.UserID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Granting repository access failed due to database errors")
		return nil, err
	}

	if access != nil {
		result.conflict = true
		return result, nil
	}

	_, err = svc.store.Access().GrantAccess(ctx, req.ResourceID, req.ResourceType, req.UserID, req.GrantedBy,
		req.AccessLevel)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Granting repository access failed due to database errors")
		return nil, err
	}

	return result, nil
}

func (svc *repositoryService) revokeAccess(reqCtx context.Context, id string, req *mgmt.AccessRevokeRequest) (notFound bool,
	mismatch bool, err error) {

	if id != req.ResourceID {
		return false, true, nil
	}

	tx, err := svc.store.Begin(reqCtx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Revoking repository access failed to due to transaction errors")
		return false, false, err
	}

	ctx := store.WithTxContext(reqCtx, tx)

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	ns, err := svc.store.Repositories().Get(ctx, req.ResourceID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Revoking repository access failed due to database errors")
		return false, false, err
	}
	if ns == nil {
		return true, false, nil
	}

	err = svc.store.Access().RevokeAccess(ctx, req.ResourceID, req.ResourceType, req.UserID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Revoking repository access failed due to database errors")
		return false, false, err
	}

	return false, false, nil
}

func (svc *repositoryService) listUsers(reqCtx context.Context, id string, cond *store.ListQueryConditions) (accesses []*models.ResourceAccessView,
	total int, err error) {

	tx, err := svc.store.Begin(reqCtx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to fetch user access due to database errors")
		return nil, -1, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	ctx := store.WithTxContext(reqCtx, tx)

	repo, err := svc.store.Repositories().Get(ctx, id)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Listing user accesses of repository(%s) failed", id)
		return nil, -1, err
	}
	if repo == nil {
		log.Logger().Warn().Msgf("Unable to retrieve the user accesses of non-existent repository(%s)", id)
		return
	}

	ns, err := svc.store.Namespaces().Get(ctx, repo.NamespaceID)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Listing user accesses of repository(%s) failed due to namespace validation errors", id)
		return nil, -1, err
	}

	if ns == nil {
		log.Logger().Warn().Msgf("Unable to retrieve the user accesses of repository(%s) due to namespace validation errors", id)
		return
	}

	if cond == nil {
		cond = &store.ListQueryConditions{
			Filters: []store.Filter{
				{
					Field: "",
					Values: []any{
						store.Filter{
							Field:    constants.FilterFieldNamespaceID,
							Values:   []any{ns.Id},
							Operator: store.OpEqual,
						},
						store.Filter{
							Field:    constants.FilterFieldRepositoryID,
							Values:   []any{id},
							Operator: store.OpEqual,
						},
					},
					Operator: store.OpOR,
				},
			},
			Page:       1,
			Limit:      20,
			SortOrder:  store.SortAsc,
			SearchTerm: "",
		}
	} else {
		if cond.Filters == nil {
			cond.Filters = []store.Filter{
				{
					Field: "",
					Values: []any{
						store.Filter{
							Field:    constants.FilterFieldNamespaceID,
							Values:   []any{ns.Id},
							Operator: store.OpEqual,
						},
						store.Filter{
							Field:    constants.FilterFieldRepositoryID,
							Values:   []any{id},
							Operator: store.OpEqual,
						},
					},
					Operator: store.OpOR,
				},
			}
		} else {
			cond.Filters = append(cond.Filters, store.Filter{
				Field: "",
				Values: []any{
					store.Filter{
						Field:    constants.FilterFieldNamespaceID,
						Values:   []any{ns.Id},
						Operator: store.OpEqual,
					},
					store.Filter{
						Field:    constants.FilterFieldRepositoryID,
						Values:   []any{id},
						Operator: store.OpEqual,
					},
				},
				Operator: store.OpOR,
			})
		}
	}

	accesses, total, err = svc.store.Access().List(ctx, cond)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Failed to retrieve user accesses of repository(%s)", id)
	}

	return accesses, total, err
}