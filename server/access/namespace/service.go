package namespace

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

type namespaceService struct {
	store         store.Store
	accessManager *resource.Manager
}

type createNsResult struct {
	nsId       string
	statusCode int
	errMsg     string
}

func (svc *namespaceService) createNamespace(reqCtx context.Context, req *mgmt.CreateNamespaceRequest) (res *createNsResult, err error) {
	tx, err := svc.store.Begin(reqCtx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to create namespace due to transactions errors")
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	ctx := store.WithTxContext(reqCtx, tx)

	res = &createNsResult{}

	nsModel, err := svc.store.Namespaces().GetByName(ctx, constants.HostedRegistryID, req.Name)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to check namespace from database")
		return nil, err
	}
	if nsModel != nil {
		res.statusCode = http.StatusConflict
		res.errMsg = "Another namespace is available with same name"
		return res, nil
	}


	res.nsId, err = svc.store.Namespaces().Create(ctx, constants.HostedRegistryID, req.Name, req.Purpose,
		req.Description, req.IsPublic)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when creating namespace: %s", req.Name)
		return res, err
	}

	// TODO We'll get admin UUID and pass it. Later, We have to get it from Auth context. Exact approch has TBD
	admin, err := svc.store.Users().Get(ctx, "admin")
	if err != nil {
		log.Logger().Error().Err(err).Msg("Error occurred when finding admin's id")
		return res, err
	}
	if admin == nil {
		return res, fmt.Errorf("no admin account found")
	}

	reason, err := svc.accessManager.GrantAccessBulk(ctx, res.nsId, constants.ResourceTypeNamespace, admin.Id, req.Maintainers,
		constants.AccessLevelMaintainer)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Granting access to maintainers during namespace(%s) creation failed", req.Name)
		return res, err
	}

	if reason != resource.Success {
		res.statusCode, res.errMsg = reason.MapToHTTP(true)
		return
	}
	res.statusCode = http.StatusCreated
	res.errMsg = ""

	return res, err
}

func (svc *namespaceService) namespaceExists(reqCtx context.Context, identifier string) (bool, error) {

	exists, err := svc.store.Namespaces().ExistsByIdentifier(reqCtx, constants.HostedRegistryID, identifier)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Error in checking namespace availablity")
		return false, err
	}
	return exists, nil
}

func (svc *namespaceService) deleteNamespace(reqCtx context.Context, identifier string) (notFound bool, err error) {
	tx, err := svc.store.Begin(reqCtx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to delete namespace due to transactions errors")
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

	exists, err := svc.store.Namespaces().ExistsByIdentifier(ctx, constants.HostedRegistryID, identifier)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error in checking namespace: %s", identifier)
		return false, err
	}

	if !exists {
		log.Logger().Warn().Msgf("Attempt to delete non-existing namespace(%s) failed", identifier)
		return true, err
	}

	err = svc.store.Namespaces().DeleteByIdentifier(ctx, constants.HostedRegistryID, identifier)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error in deleting namespace: %s", identifier)
		return false, err
	}
	return false, nil
}

func (svc *namespaceService) getNamespace(reqCtx context.Context, identifier string) (m *models.NamespaceModel, err error) {
	m, err = svc.store.Namespaces().GetByIdentifier(reqCtx, constants.HostedRegistryID, identifier)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error in retriving namespace: %s", identifier)
		return nil, err
	}
	return m, nil
}

func (svc *namespaceService) updateNamespace(reqCtx context.Context, identifier string, req *mgmt.UpdateNamespaceRequest) (notFound bool, err error) {
	tx, err := svc.store.Begin(reqCtx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to update namespace due to transaction errors")
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

	ns, err := svc.store.Namespaces().GetByIdentifier(ctx, constants.HostedRegistryID, identifier)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to update namespace due to database errors")
		return false, err
	}

	if ns == nil {
		log.Logger().Warn().Msgf("Failed to update non existent namespace: %s", req.ID)
		return true, nil
	}

	purpose := req.Purpose

	if req.Purpose == "" {
		purpose = ns.Purpose
	}

	err = svc.store.Namespaces().Update(ctx, ns.Id, req.Description, purpose)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to update namespace due to database errors")
		return false, err
	}

	return false, nil
}

type patchResult struct {
	httpStatusCode int
	httpErrorMsg   string
	success        bool
}

func (svc *namespaceService) changeState(reqCtx context.Context, identifier, newState string) (result *patchResult,
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

	ns, err := svc.store.Namespaces().GetByIdentifier(ctx, constants.HostedRegistryID, identifier)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to change state of namespace due to database errors")
		return nil, err
	}

	if ns == nil {
		log.Logger().Warn().Msgf("Failed to change state of non existent namespace: %s", identifier)
		result.httpErrorMsg = "Namespace " + identifier + " is not found"
		result.httpStatusCode = http.StatusNotFound

		return result, nil
	}

	if ns.State == newState {
		log.Logger().Debug().Msgf("No changes in state. Updating state of namespace(%s) is skipped", identifier)
		result.success = true
		return result, nil
	}

	log.Logger().Debug().Msgf("Namespace state change request received: Current State: %s, New State: %s", ns.State, newState)
	if ns.State == constants.ResourceStateActive && newState == constants.ResourceStateDisabled {
		log.Logger().Warn().Msgf("Not allowed to change namespace(%s) state from 'Active' to 'Disabled'", ns.Id)
		result.httpErrorMsg = "Not allowed to change namespace state from 'Active' to 'Disabled'"
		result.httpStatusCode = http.StatusForbidden
		return result, nil
	}

	if newState != constants.ResourceStateActive {
		log.Logger().Warn().Msgf("Changing state of namespace(%s) to %s. This change will affect associated repositories", ns.Id, newState)
	}

	err = svc.store.Namespaces().SetStateByID(ctx, ns.Id, newState)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Failed to change state of namespace(%s) due to database errors", ns.Id)
		return nil, err
	}
	result.success = true

	return result, nil
}

func (svc *namespaceService) changeVisiblity(reqCtx context.Context, identifier string, isPublic bool) (result *patchResult,
	err error) {
	result = &patchResult{}

	tx, err := svc.store.Begin(reqCtx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to change visibility of namespace due to transaction errors")
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

	ns, err := svc.store.Namespaces().GetByIdentifier(ctx, constants.HostedRegistryID, identifier)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Failed to change visibility of namespace due to database errors")
		return nil, err
	}

	if ns == nil {
		log.Logger().Warn().Msgf("Failed to change visibility of non existent namespace: %s", identifier)
		result.httpErrorMsg = "Namespace " + identifier + " is not found"
		result.httpStatusCode = http.StatusNotFound
		return result, nil
	}

	if ns.IsPublic == isPublic {
		result.success = true
		return result, nil
	}

	if !isPublic {
		log.Logger().Warn().Msgf("Changing visibility of namespace(%s) to private. This change will affect associated repositories", ns.Id)
	}

	if ns.State == constants.ResourceStateDisabled {
		log.Logger().Warn().Msgf("Not allowed to change visibility of namespace(%s) when namespace is in disabled state", identifier)
		result.success = false
		result.httpStatusCode = http.StatusForbidden
		result.httpErrorMsg = "Not allowed to change visibility of a namespace when it is in disabled state"
		return
	}

	err = svc.store.Namespaces().SetVisiblityByID(ctx, ns.Id, isPublic)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Failed to change visibility of namespace(%s) due to errors", ns.Id)
		return nil, err
	}

	result.success = true
	return result, nil
}

type grantAccessResult struct {
	statusCode int
	errMsg     string
}

func (svc *namespaceService) grantAccess(reqCtx context.Context, req *mgmt.AccessGrantRequest) (result *grantAccessResult,
	err error) {
	result = &grantAccessResult{}

	// TODO We'll get admin UUID and pass it. Later, We have to get it from Auth context. Exact approch has TBD
	admin, err := svc.store.Users().Get(reqCtx, "admin")
	if err != nil {
		log.Logger().Error().Err(err).Msg("Error occurred when finding admin's id")
		return result, err
	}
	if admin == nil {
		return result, fmt.Errorf("no admin account found")
	}

	reason, err := svc.accessManager.GrantAccess(reqCtx, req.ResourceID, req.ResourceType, admin.Id, req.UserID,
		req.AccessLevel)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Granting namespace access failed due to errors")
		return nil, err
	}

	result.statusCode, result.errMsg = reason.MapToHTTP(false)
	return result, nil
}

func (svc *namespaceService) revokeAccess(reqCtx context.Context, req *mgmt.AccessRevokeRequest) (status int, msg string,
	err error) {

	//TODO: We pass 'admin' as revoker id. Later, we have to change this to the actual user
	reason, err := svc.accessManager.RevokeAccess(reqCtx, req.ResourceID, req.ResourceType, req.UserID, "admin")
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Unable to revoke resource access(%s:%s) to user(%s)", req.ResourceType,
			req.ResourceID, req.UserID)
	}

	status, msg = reason.MapToHTTP(false)
	return
}

func (svc *namespaceService) listNamespaces(reqCtx context.Context, cond *store.ListQueryConditions) (namespaces []*models.NamespaceView,
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

	namespaces, total, err = svc.store.Namespaces().List(ctx, cond)
	return
}

func (svc *namespaceService) listRepositories(reqCtx context.Context, identifier string, cond *store.ListQueryConditions) (repositories []*models.RepositoryView,
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

	ns, err := svc.store.Namespaces().GetByIdentifier(ctx, constants.HostedRegistryID, identifier)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Listing repositories of namespace(%s) failed", identifier)
		return nil, -1, err
	}

	if ns == nil {
		log.Logger().Warn().Msgf("Unable to retrieve the repositories of non-existent namespace(%s)", identifier)
		return
	}

	if cond == nil {
		cond = &store.ListQueryConditions{
			Filters: []store.Filter{
				{
					Field:    constants.FilterFieldNamespaceID,
					Values:   []any{ns.Id},
					Operator: store.OpEqual,
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
					Field:    constants.FilterFieldNamespaceID,
					Values:   []any{ns.Id},
					Operator: store.OpEqual,
				},
			}
		} else {
			cond.Filters = append(cond.Filters, store.Filter{
				Field:    constants.FilterFieldNamespaceID,
				Values:   []any{ns.Id},
				Operator: store.OpEqual,
			})
		}
	}

	repositories, total, err = svc.store.Repositories().List(ctx, cond)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Failed to retrieve repositories of namespace(%s)", identifier)
	}

	return repositories, total, err
}

func (svc *namespaceService) listUsers(reqCtx context.Context, identifier string, cond *store.ListQueryConditions) (accesses []*models.ResourceAccessView,
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

	ns, err := svc.store.Namespaces().GetByIdentifier(ctx, constants.HostedRegistryID, identifier)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Listing user accesses of namespace(%s) failed", identifier)
		return nil, -1, err
	}

	if ns == nil {
		log.Logger().Warn().Msgf("Unable to retrieve the user accesses of non-existent namespace(%s)", identifier)
		return
	}

	if cond == nil {
		cond = &store.ListQueryConditions{
			Filters: []store.Filter{
				{
					Field:    constants.FilterFieldResourceID,
					Values:   []any{ns.Id},
					Operator: store.OpEqual,
				},
				{
					Field:    constants.FilterFieldResourceType,
					Values:   []any{constants.ResourceTypeNamespace},
					Operator: store.OpEqual,
				},
			},
			Page:       1,
			Limit:      20,
			SortOrder:  store.SortAsc,
			SearchTerm: "",
		}
	} else {
		if cond.Filters == nil {
			cond.Filters = append(cond.Filters,
				store.Filter{
					Field:    constants.FilterFieldResourceID,
					Values:   []any{ns.Id},
					Operator: store.OpEqual,
				},
				store.Filter{
					Field:    constants.FilterFieldResourceType,
					Values:   []any{constants.ResourceTypeNamespace},
					Operator: store.OpEqual,
				})
		} else {
			cond.Filters = append(cond.Filters,
				store.Filter{
					Field:    constants.FilterFieldResourceID,
					Values:   []any{ns.Id},
					Operator: store.OpEqual,
				},
				store.Filter{
					Field:    constants.FilterFieldResourceType,
					Values:   []any{constants.ResourceTypeNamespace},
					Operator: store.OpEqual,
				})
		}
	}

	accesses, total, err = svc.store.Access().List(ctx, cond)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Failed to retrieve user accesses of namespace(%s)", identifier)
	}

	return accesses, total, err
}