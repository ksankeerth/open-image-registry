package namespace

import (
	"context"
	"net/http"

	"github.com/ksankeerth/open-image-registry/constants"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/store"
	"github.com/ksankeerth/open-image-registry/types/api/v1alpha/mgmt"
	"github.com/ksankeerth/open-image-registry/types/models"
)

type namespaceService struct {
	store store.Store
}

type createNsResult struct {
	nsId              string
	conflict          bool
	invalidMaintainer bool
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
		res.conflict = true
		return res, nil
	}

	validAccounts, err := svc.store.Users().AreAccountsActive(ctx, req.Maintainers)
	if err != nil {
		res.invalidMaintainer = false
		log.Logger().Error().Err(err).Msg("Error creating namespace: verification of maintainer accounts returned an error.")
		return nil, err
	}

	if !validAccounts {
		res.invalidMaintainer = true
		return res, nil
	}

	valid, err := svc.store.NamespaceQueries().ValidateMaintainers(ctx, req.Maintainers)
	if err != nil {
		res.invalidMaintainer = false
		log.Logger().Error().Err(err).Msgf("Error creating namespace: verification roles of maintainers returned an error")
		return nil, err
	}

	if !valid {
		res.invalidMaintainer = true
		return res, nil
	}

	res.nsId, err = svc.store.Namespaces().Create(ctx, constants.HostedRegistryID, req.Name, req.Purpose,
		req.Description, req.IsPublic)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error occurred when creating namespace: %s", req.Name)
		return res, err
	}

	for _, userId := range req.Maintainers {
		// TODO: We should provide change "admin" to the actual user later
		_, err = svc.store.Access().GrantAccess(ctx, res.nsId, constants.ResourceTypeNamespace, userId, constants.AccessLevelMaintainer, "admin")
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Error occurred when granting maintainer(%s) access to namespace: %s", userId, req.Name)
			return res, err
		}
	}

	return res, nil
}

func (svc *namespaceService) namespaceExists(reqCtx context.Context, identifier string) (bool, error) {

	exists, err := svc.store.Namespaces().ExistsByIdentifier(reqCtx, constants.HostedRegistryID, identifier)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Error in checking namespace availablity")
		return false, err
	}
	return exists, nil
}

func (svc *namespaceService) deleteNamespace(reqCtx context.Context, identifier string) error {
	err := svc.store.Namespaces().DeleteByIdentifier(reqCtx, constants.HostedRegistryID, identifier)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Error in deleting namespace: %s", identifier)
		return err
	}
	return nil
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
		return nil, nil
	}

	if ns.State == constants.ResourceStateActive && newState == constants.ResourceStateDisabled {
		log.Logger().Warn().Msgf("Not allowed to change namespace(%s) state from 'Active' to 'Disabled'", ns.Id)
		result.httpErrorMsg = "Not allowed to change namespace state from 'Active' to 'Disabled'"
		result.httpStatusCode = http.StatusForbidden
		return result, nil
	}

	if newState != constants.ResourceStateActive {
		log.Logger().Warn().Msgf("Changing state of namespace(%s) to %s. This change will affect associated repositories", ns.Id, newState)
	}

	err = svc.store.Namespaces().SetStateByID(ctx, ns.Id, constants.ResourceStateDeprecated)
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
	conflict            bool // conflict with existing resource access
	userNotFound        bool
	grantedUserNotFound bool
	resourceNotFound    bool
	bodyURLMismatch     bool // true if identifier in url doesn't match with ID provided in request
}

func (svc *namespaceService) grantAccess(reqCtx context.Context, identifier string, req *mgmt.AccessGrantRequest) (result *grantAccessResult, err error) {
	tx, err := svc.store.Begin(reqCtx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Granting namespace access to user failed to due to transaction errors")
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

	result = &grantAccessResult{}

	user, err := svc.store.Users().Get(ctx, req.UserID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Granting namespace access failed due to database errors")
		return nil, err
	}

	if user == nil {
		result.userNotFound = true
		return result, nil
	}

	grantedUser, err := svc.store.Users().GetByUsername(ctx, req.GrantedBy)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Granting namespace access failed due to database errors")
		return nil, err
	}

	if grantedUser == nil {
		result.grantedUserNotFound = true
		return result, nil
	}

	ns, err := svc.store.Namespaces().Get(ctx, req.ResourceID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Granting namespace access failed due to database errors")
		return nil, err
	}
	if ns == nil {
		result.resourceNotFound = true
		return result, nil
	}

	if !(identifier == ns.Id || identifier == ns.Name) {
		result.bodyURLMismatch = true
		return result, nil
	}

	access, err := svc.store.Access().GetUserAccess(ctx, req.ResourceID, req.ResourceType, req.UserID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Granting namespace access failed due to database errors")
		return nil, err
	}

	if access != nil {
		result.conflict = true
		return result, nil
	}

	_, err = svc.store.Access().GrantAccess(ctx, req.ResourceID, req.ResourceType, req.UserID, req.GrantedBy,
		req.AccessLevel)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Granting namespace access failed due to database errors")
		return nil, err
	}

	return result, nil
}

func (svc *namespaceService) revokeAccess(reqCtx context.Context, identifier string, req *mgmt.AccessRevokeRequest) (notFound bool,
	mismatch bool, err error) {
	tx, err := svc.store.Begin(reqCtx)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Revoking namespace access failed to due to transaction errors")
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

	ns, err := svc.store.Namespaces().Get(ctx, req.ResourceID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Revoking namespace access failed due to database errors")
		return false, false, err
	}
	if ns == nil {
		return true, false, nil
	}

	if !(ns.Id == identifier || ns.Name == identifier) {
		return false, true, nil
	}

	err = svc.store.Access().RevokeAccess(ctx, req.ResourceID, req.ResourceType, req.UserID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Revoking namespace access failed due to database errors")
		return false, false, err
	}

	return false, false, nil
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

	accesses, total, err = svc.store.Access().List(ctx, cond)
	if err != nil {
		log.Logger().Error().Err(err).Msgf("Failed to retrieve user accesses of namespace(%s)", identifier)
	}

	return accesses, total, err
}
