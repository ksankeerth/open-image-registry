package access

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/ksankeerth/open-image-registry/constants"
	"github.com/ksankeerth/open-image-registry/log"
	"github.com/ksankeerth/open-image-registry/store"
	"github.com/ksankeerth/open-image-registry/types/models"
)

type Manager struct {
	store store.Store
}

func NewManager(store store.Store) *Manager {
	return &Manager{
		store: store,
	}
}

type AccessOpFailure int

const (
	Success AccessOpFailure = iota // zero = success
	InitiatorNotFound
	InitiatorNotAdmin
	GranteeNotFound
	GranteeIsAdmin

	ResourceNotFound
	ResourceDisabled
	ExceedsRole
	Conflict
	HasSameAccessAlready
	NotAllowedAccessLevel
	RedundantAccess
	AccessNotFound
	CannotRevokeSelf
	UnexpectedError
)

func (g AccessOpFailure) String() string {
	return [...]string{
		"SUCCESS",
		"INITIATOR_NOT_FOUND",
		"INITIATOR_NOT_ADMIN",
		"GRANTEE_NOT_FOUND",
		"GRANTEE_IS_ADMIN",
		"RESOURCE_NOT_FOUND",
		"RESOURCE_DISABLED",
		"EXCEEDS_ROLE",
		"CONFLICT",
		"HAS_SAME_ACCESS_ALREADY",
		"NOT_ALLOWED_ACCESS_LEVEL",
		"REDUNDANT_ACCESS",
		"ACCESS_NOT_FOUND",
		"CANNOT_REVOKE_SELF",
		"UNEXPECTED_ERROR",
	}[g]
}

// GrantAccess will grant resource access to given resource. Before it gives access
// it will assess multiple conditions and return failure reason if any of conditions failed
// If any unexpected error occurred, It will return an error. In successful scenario,
// reason will be zero.
func (m *Manager) GrantAccess(ctx context.Context, resourceID, resourceType, grantorID, granteeID,
	accessLevel string) (reason AccessOpFailure, err error) {

	if tx, _ := store.TxFromContext(ctx); tx == nil {
		tx, err = m.store.Begin(ctx)
		if err != nil {
			log.Logger().Error().Err(err).Msg("Granting resource access failed when creating new database transaction")
			return UnexpectedError, err
		}

		ctx = store.WithTxContext(ctx, tx)
		defer func() {
			if err != nil || reason != Success {
				tx.Rollback()
			} else {
				tx.Commit()
			}
		}()
	}

	// 0. Verify allowed access levels by resource type
	allowedAccessLevels := accessLevelsByType(resourceType)
	if !slices.Contains(allowedAccessLevels, accessLevel) {
		return NotAllowedAccessLevel, nil
	}

	// 1. Verify grantor
	ok, reason, err := m.verifyInitiatorAuthority(ctx, grantorID, accessLevel, resourceType, resourceID)
	if !ok {
		return reason, err
	}

	// 2. Verify resource
	ok, reason, err = m.verifyResource(ctx, resourceID, resourceType)
	if !ok {
		return reason, err
	}

	// 3. Verify grantee
	ok, reason, err = m.verifyGrantee(ctx, granteeID, accessLevel, resourceType)
	if !ok {
		return reason, err
	}

	// 5. Check for conflicts
	hasAccess, sameLevel, err := m.hasAccess(ctx, granteeID, resourceID, resourceType, accessLevel)
	if err != nil {
		return UnexpectedError, err
	}

	if sameLevel {
		return HasSameAccessAlready, nil
	}

	if hasAccess {
		return Conflict, nil
	}

	// 6. Check for redundant access
	if resourceType == constants.ResourceTypeRepository {
		yes, err := m.isRedundantAccess(ctx, resourceID, accessLevel, granteeID)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Granting resource access failed when verifying redundant access")
			return UnexpectedError, err
		}

		if yes {
			return RedundantAccess, nil
		}
	}

	_, err = m.store.Access().GrantAccess(ctx, resourceID, resourceType, granteeID, accessLevel, grantorID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Granting resource access failed")
		return UnexpectedError, err
	}

	return Success, nil
}

// TODO:
// 1. Check allowed access levels by resource type
// 2. Check for redundant access
func (m *Manager) GrantAccessBulk(ctx context.Context, resourceID, resourceType,
	grantorID string, grantees []string, accessLevel string) (reason AccessOpFailure, err error) {

	if tx, _ := store.TxFromContext(ctx); tx == nil {
		tx, err = m.store.Begin(ctx)
		if err != nil {
			log.Logger().Error().Err(err).Msg("Granting resource access failed when creating new database transaction")
			return UnexpectedError, err
		}

		ctx = store.WithTxContext(ctx, tx)
		defer func() {
			if err != nil || reason != Success {
				tx.Rollback()
			} else {
				tx.Commit()
			}
		}()
	}

	// 1. Verify grantor
	ok, reason, err := m.verifyInitiatorAuthority(ctx, grantorID, accessLevel, resourceType, resourceID)
	if !ok {
		return reason, err
	}

	// 2. Verify resource
	ok, reason, err = m.verifyResource(ctx, resourceID, resourceType)
	if !ok {
		return reason, err
	}

	// 3. Verify whether all the grantees exist
	matchingUsers, err := m.store.UserQueries().CountUsersByIDs(ctx, grantees)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Granting access to users in bulk failed when verifying grantees")
		return UnexpectedError, err
	}

	if matchingUsers < len(grantees) {
		log.Logger().Warn().Msg("Granting access to users in bulk failed due to some invalid user accounts")
		return GranteeNotFound, nil
	}

	// 4. Verify if all users are eligible for given access level based on roles.
	// If one user is found not to be eligible, we'll reject the request
	possibleRoles := getRolesSatisfyingAccessLevel(accessLevel)
	matchingUsers, err = m.store.UserQueries().CountUsersByIDsAndRoles(ctx, grantees, possibleRoles)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Granting access to users in bulk failed when verifying grantees' eligibility")
		return UnexpectedError, err
	}

	if matchingUsers < len(grantees) {
		log.Logger().Warn().Msg("Granting access to users in bulk failed due to some grantees' ineligiblity")
		return ExceedsRole, nil
	}

	// 5. Check for conflicts
	conflictingUsers, err := m.store.AccessQueries().CountUsersByResourceAccess(ctx, grantees, resourceType, resourceID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Granting access to users in bulk failed when verifying grantees' current access")
		return UnexpectedError, err
	}
	if conflictingUsers != 0 {
		log.Logger().Warn().Msgf("Granting access to users in bulk failed since %d out of %d users already have conflicting access to resource %s:%s",
			conflictingUsers, len(grantees), resourceType, resourceID)
		return Conflict, nil
	}

	// Since all the checks are completed, grant access
	for _, id := range grantees {
		_, err = m.store.Access().GrantAccess(ctx, resourceID, resourceType, id, accessLevel, grantorID)
		if err != nil {
			log.Logger().Error().Err(err).Msgf("Granting access to users in bulk failed due to error occurred when granting access to user %s", id)
			return UnexpectedError, err
		}
	}

	log.Logger().Info().Msgf("Resource(%s:%s) access is given to users(%d): %s", resourceType, resourceID, len(grantees),
		accessLevel)

	return Success, nil
}

func (m *Manager) isRedundantAccess(ctx context.Context, repositoryID, accessLevel, userID string) (yes bool,
	err error) {
	repo, err := m.store.Repositories().Get(ctx, repositoryID)
	if err != nil {
		return false, err
	}

	if repo == nil {
		return false, fmt.Errorf("repository not found: %s", repositoryID)
	}

	nsAccess, err := m.store.Access().GetUserAccess(ctx, repo.NamespaceID, constants.ResourceTypeNamespace,
		userID)
	if err != nil {
		return false, err
	}

	if nsAccess == nil {
		return false, nil
	}

	if nsAccess.AccessLevel == accessLevel {
		return true, nil
	}

	if nsAccess.AccessLevel == constants.AccessLevelMaintainer {
		return true, nil
	}

	if nsAccess.AccessLevel == constants.AccessLevelDeveloper && accessLevel == constants.AccessLevelGuest {
		return true, nil
	}

	return false, nil
}

func (m *Manager) RevokeAccess(ctx context.Context, resourceID, resourceType, userID, revokerID string) (reason AccessOpFailure, err error) {
	if tx, _ := store.TxFromContext(ctx); tx == nil {
		tx, err = m.store.Begin(ctx)
		if err != nil {
			log.Logger().Error().Err(err).Msg("Revoking resource access failed when creating new database transaction")
			return UnexpectedError, err
		}

		ctx = store.WithTxContext(ctx, tx)
		defer func() {
			if err != nil || reason != Success {
				tx.Rollback()
			} else {
				tx.Commit()
			}
		}()
	}

	// 1. Self revoke check
	if userID == revokerID {
		return CannotRevokeSelf, nil
	}

	// 2.Verify grantor and authority
	currentAccess, err := m.store.Access().GetUserAccess(ctx, resourceID, resourceType, userID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Revoking resource access failed when checking current access")
		return UnexpectedError, err
	}

	if currentAccess == nil {
		return AccessNotFound, nil
	}

	ok, reason, err := m.verifyInitiatorAuthority(ctx, revokerID, currentAccess.AccessLevel, resourceType, resourceID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Revoking resource access failed when verifying initiator authority")
		return UnexpectedError, err
	}

	if !ok {
		return reason, nil
	}

	// now we can revoke
	err = m.store.Access().RevokeAccess(ctx, resourceID, resourceType, userID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Revoking resource access failed")
		return UnexpectedError, err
	}

	return Success, nil
}

// GetUserAccessByLevels find all accesses of resources the user has.
// Currently, this method only supports upto maximum 2 access levels. If nil or empty access levels
// are passed, it will return all the resource access.
func (m *Manager) GetUserAccessByLevels(ctx context.Context, userId string, page, limit uint,
	accessLevels ...string) (accesses []*models.ResourceAccessView, total int, err error) {

	queryConds := store.ListQueryConditions{
		Page:  page,
		Limit: limit,
		Filters: []store.Filter{
			{
				Field:    constants.FilterFieldUserID,
				Operator: store.OpEqual,
				Values:   []any{userId},
			},
		},
	}

	if len(accessLevels) > 0 {
		queryConds.Filters = append(queryConds.Filters, store.Filter{
			Field:    constants.FilterFieldAccessLevel,
			Operator: store.OpIn,
			Values:   []any{accessLevels},
		})
	}

	return m.store.Access().List(ctx, &queryConds)
}

func (m *Manager) getResourceInfo(ctx context.Context, resourceType, id string) (exists bool,
	state, parentState string, err error) {
	switch resourceType {
	case constants.ResourceTypeNamespace:
		ns, err := m.store.Namespaces().Get(ctx, id)
		if err != nil {
			return false, "", "", err
		}
		if ns == nil {
			return false, "", "", nil
		}
		return true, ns.State, "", nil
	case constants.ResourceTypeRepository:
		repo, err := m.store.Repositories().Get(ctx, id)
		if err != nil {
			return false, "", "", err
		}
		if repo == nil {
			return false, "", "", nil
		}

		ns, err := m.store.Namespaces().Get(ctx, repo.NamespaceID)
		if err != nil {
			return false, "", "", err
		}
		return true, repo.State, ns.State, nil
	case constants.ResourceTypeUpstream:
		// TODO: Implement later
		return false, "", "", nil
	}

	return false, "", "", fmt.Errorf("invalid resource type")
}

func (m *Manager) verifyInitiatorAuthority(ctx context.Context, grantorID, accessLevel, resourceType, resourceId string) (ok bool,
	reason AccessOpFailure, err error) {
	grantor, err := m.store.Users().Get(ctx, grantorID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Granting resource access failed when verifing grantor")
		return false, UnexpectedError, err
	}
	if grantor == nil {
		return false, InitiatorNotFound, nil
	}

	role, err := m.store.Users().GetRole(ctx, grantor.Id)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Granting resource access failed when verifing grantor")
		return false, UnexpectedError, err
	}

	if role == constants.RoleMaintainer {

		// allowed access level Developer and Guest
		// allowed resource type namespace and repository
		if !(accessLevel == constants.AccessLevelDeveloper || accessLevel == constants.AccessLevelGuest) {
			return false, InitiatorNotAdmin, nil
		}

		if !(resourceType == constants.ResourceTypeNamespace || resourceType == constants.ResourceTypeRepository) {
			return false, InitiatorNotAdmin, nil
		}

		// the grantor should have Maintainer level access to this resource to able to give access to other users.
		var nsId string
		if resourceType == constants.ResourceTypeRepository {
			repo, err := m.store.Repositories().Get(ctx, resourceId)
			if err != nil {
				return false, 0, err
			}
			if repo == nil {
				return false, 0, fmt.Errorf("repository %s not found", resourceId)
			}
			nsId = repo.NamespaceID
		}
		if resourceType == constants.ResourceTypeNamespace {
			nsId = resourceId
		}

		if nsId != "" {
			currentAccess, err := m.store.Access().GetUserAccess(ctx, resourceId, resourceType, grantorID)
			if err != nil {
				return false, 0, err
			}
			if currentAccess == nil || currentAccess.AccessLevel != constants.AccessLevelMaintainer {
				return false, InitiatorNotAdmin, nil
			}
		}
		return true, 0, nil
	} else if role != constants.RoleAdmin {
		return false, InitiatorNotAdmin, nil
	}

	return true, 0, nil
}

func (m *Manager) verifyResource(ctx context.Context, resourceID, resourceType string) (ok bool,
	reason AccessOpFailure, err error) {

	exists, state, parentState, err := m.getResourceInfo(ctx, resourceType, resourceID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Granting resource access failed when verifying resource")
		return false, UnexpectedError, err
	}
	if !exists {
		return false, ResourceNotFound, nil
	}

	if state == constants.ResourceStateDisabled {
		return false, ResourceDisabled, nil
	}

	if parentState == constants.ResourceStateDisabled {
		return false, ResourceDisabled, nil
	}

	return true, 0, nil
}

func (m *Manager) verifyGrantee(ctx context.Context, granteeID, accessLevel, resourceType string) (ok bool,
	reason AccessOpFailure, err error) {

	grantee, err := m.store.Users().Get(ctx, granteeID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Granting resource access failed when verifying grantee")
		return false, UnexpectedError, err
	}
	if grantee == nil {
		return false, GranteeNotFound, nil
	}

	granteeRole, err := m.store.Users().GetRole(ctx, granteeID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Granting resource access failed when verifying grantee's role")
		return false, UnexpectedError, err
	}

	// if grantee is an admin, it is not allowed to have Guest and Developer access to resources
	// since the grantee already has the access. However, they are allowed to have 'Maintainer'
	// access level on namespaces since it is a must to have at least one maintainer to a namespace.

	allowedAccessLevels := accessLevelsByRole(granteeRole)
	log.Logger().Debug().Msgf("Checking allowed access levels(%s) against the given role(%s)", strings.Join(allowedAccessLevels, ","), granteeRole)
	if !slices.Contains(allowedAccessLevels, accessLevel) {
		return false, NotAllowedAccessLevel, nil
	}

	if granteeRole == constants.RoleAdmin && resourceType != constants.ResourceTypeNamespace {
		return false, NotAllowedAccessLevel, nil
	}

	// // Verify role and access level
	// if m.exceedsRole(accessLevel, granteeRole) {
	// 	return false, ExceedsRole, nil
	// }

	return true, 0, nil
}

func (m *Manager) hasAccess(ctx context.Context, granteeID, resourceID, resourceType, accessLevel string) (yes bool,
	sameLevel bool, err error) {

	currentAccess, err := m.store.Access().GetUserAccess(ctx, resourceID, resourceType,
		granteeID)
	if err != nil {
		log.Logger().Error().Err(err).Msg("Granting resource access failed when verifying grantee's current access")
		return false, false, err
	}
	if currentAccess != nil {
		if currentAccess.AccessLevel == accessLevel {
			// If grantee already have same level of access, then we will return success
			return true, true, nil
		}
		return true, false, nil
	}

	return false, false, nil
}

// we don't need this function until we check redundant access in bulk grant method
// func getInheritedAccessLevels(accessLevel string) []string {
// switch accessLevel {
// case constants.AccessLevelMaintainer:
// return []string{constants.AccessLevelMaintainer, constants.AccessLevelDeveloper, constants.AccessLevelGuest}
// case constants.AccessLevelDeveloper:
// return []string{constants.AccessLevelDeveloper, constants.AccessLevelGuest}
// }
// return []string{}
// }

func getRolesSatisfyingAccessLevel(accessLevel string) []string {
	switch accessLevel {
	case constants.AccessLevelMaintainer:
		return []string{constants.RoleMaintainer, constants.RoleAdmin}
	case constants.AccessLevelDeveloper:
		return []string{constants.RoleDeveloper, constants.RoleMaintainer}
	case constants.AccessLevelGuest:
		return []string{constants.RoleGuest, constants.RoleDeveloper, constants.RoleMaintainer}
	default:
		return []string{}
	}
}

func accessLevelsByRole(role string) []string {
	switch role {
	case constants.RoleAdmin:
		return []string{constants.AccessLevelMaintainer}
	case constants.RoleMaintainer:
		return []string{constants.AccessLevelMaintainer, constants.AccessLevelDeveloper, constants.AccessLevelGuest}
	case constants.RoleDeveloper:
		return []string{constants.AccessLevelDeveloper, constants.AccessLevelGuest}
	case constants.RoleGuest:
		return []string{constants.AccessLevelGuest}
	}

	return []string{}
}

func accessLevelsByType(resourceType string) []string {
	switch resourceType {
	case constants.ResourceTypeNamespace:
		return []string{constants.AccessLevelMaintainer, constants.AccessLevelDeveloper,
			constants.AccessLevelGuest}
	case constants.ResourceTypeRepository:
		return []string{constants.AccessLevelDeveloper,
			constants.AccessLevelGuest}
	case constants.ResourceTypeUpstream:
		//TODO: TBD
		return []string{}
	}
	return []string{}
}