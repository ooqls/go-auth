package authorizationv1

//go:generate go run go.uber.org/mock/mockgen -source=authorizer.go -destination=mocks/mock_authorizer.go -package=mocks

import (
	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/aggsv1"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/datav1"
	"github.com/ooqls/go-auth/internal/permissionsv1"
	"github.com/ooqls/go-auth/internal/rolebindingsv1"
	"github.com/ooqls/go-auth/internal/rolesv1"
	"go.uber.org/zap"
)

var _ Authorizer = &AuthorizerImpl{}

type Authorizer interface {
	IsAuthorizedToAssignPermission(ctx *Context, roleID uuid.UUID, permission string) error
	IsAuthorizedToUnassignPermission(ctx *Context, roleID uuid.UUID) error
	IsAuthorizedToModifyUser(ctx *Context, targetUserID uuid.UUID) error
	IsAuthorizedToReadUserRoles(ctx *Context, targetUserID uuid.UUID) error
	HasHigherHierarchy(ctx *Context, hierarchy int32) error
	IsAuthorizedToReadRolePermissions(ctx *Context, roleID uuid.UUID) error
	IsAuthorizedToPerformResourceAction(ctx *Context, action Action, group, kind string) error
	IsAuthorizedToPerformCoreAction(ctx *Context, action Action, group, kind string) error
}

type AuthorizerImpl struct {
	pReader  permissionsv1.Reader
	rbReader rolebindingsv1.Reader
	rReader  rolesv1.Reader
	aReader  aggsv1.Reader
}

func NewAuthorizerImpl(fact datav1.Factory) Authorizer {
	return &AuthorizerImpl{
		pReader:  fact.NewPermissionReader(),
		rbReader: fact.NewRoleBindingsReader(),
		rReader:  fact.NewRoleReader(),
		aReader:  fact.NewAggReader(),
	}
}

func (a *AuthorizerImpl) IsAuthorizedToAssignPermission(ctx *Context, roleID uuid.UUID, permission string) error {
	if ctx.IsInternalOperation() {
		return nil
	}

	requester := ctx.GetAuthedUser()
	requesterHierarchy, err := a.aReader.GetUserHierarchy(ctx, requester.Id)
	if err != nil {
		return err
	}

	role, err := a.rReader.GetRole(ctx, roleID)
	if err != nil {
		return err
	}

	if role != nil && role.Hierarchy > requesterHierarchy {
		ctx.L().Debug("assign permission denied: role hierarchy exceeds requester hierarchy")
		return ErrPermissionDenied
	}

	pMeta := corev1.PermissionsV1
	authed, err := a.pReader.HasCorePermission(ctx, requester.Id, pMeta.Group, pMeta.Kind, string(AssignAction))
	if !authed {
		ctx.L().Debug("assign permission denied: requester lacks permission binding assign permission")
		return ErrPermissionDenied
	}

	return nil
}

func (a *AuthorizerImpl) HasHigherHierarchy(ctx *Context, hierarchy int32) error {
	if ctx.IsInternalOperation() {
		return nil
	}

	requesterHierarchy, err := a.aReader.GetUserHierarchy(ctx, ctx.GetAuthedUser().Id)
	if err != nil {
		return err
	}

	if requesterHierarchy < hierarchy {
		return ErrPermissionDenied
	}
	return nil
}

func (a *AuthorizerImpl) IsAuthorizedToUnassignPermission(ctx *Context, roleID uuid.UUID) error {
	if ctx.IsInternalOperation() {
		return nil
	}

	requester := ctx.GetAuthedUser()

	requesterHierarchy, err := a.aReader.GetUserHierarchy(ctx, requester.Id)
	if err != nil {
		return err
	}

	role, err := a.rReader.GetRole(ctx, roleID)
	if err != nil {
		return err
	}

	if role != nil && role.Hierarchy > requesterHierarchy {
		ctx.L().Debug("unassign permission denied: role hierarchy exceeds requester hierarchy")
		return ErrPermissionDenied
	}

	authed, err := a.pReader.HasCorePermission(ctx, requester.Id, corev1.PermissionsV1.Group, corev1.PermissionsV1.Kind, string(UnassignAction))
	if err != nil {
		return err
	}

	if !authed {
		ctx.L().Debug("unassign permission denied: requester lacks permission binding delete permission")
		return ErrPermissionDenied
	}

	return nil
}

func (a *AuthorizerImpl) IsAuthorizedToModifyUser(ctx *Context, targetUserID uuid.UUID) error {
	if ctx.IsInternalOperation() {
		return nil
	}

	requester := ctx.GetAuthedUser()
	if ctx.IsInternalOperation() {
		return nil
	}

	requesterHierarchy, err := a.aReader.GetUserHierarchy(ctx, requester.Id)
	if err != nil {
		return err
	}

	assigneeHierarchy, err := a.aReader.GetUserHierarchy(ctx, targetUserID)
	if err != nil {
		return err
	}

	if requesterHierarchy <= assigneeHierarchy {
		ctx.L().Debug("role assignment failed because assignee has a higher user hierarchy than assigner")
		return ErrPermissionDenied
	}

	return nil
}

func (a *AuthorizerImpl) IsAuthorizedToReadUserRoles(ctx *Context, targetUserID uuid.UUID) error {
	if ctx.IsInternalOperation() {
		return nil
	}

	requester := ctx.GetAuthedUser()
	requesterHierarchy, err := a.aReader.GetUserHierarchy(ctx, requester.Id)
	if err != nil {
		return err
	}

	targetHierarchy, err := a.aReader.GetUserHierarchy(ctx, targetUserID)
	if err != nil {
		return err
	}

	if requesterHierarchy < targetHierarchy {
		ctx.L().Debug("read user roles denied: requester hierarchy is lower than target user hierarchy")
		return ErrPermissionDenied
	}

	hasRolesPerm, err := a.pReader.HasCorePermission(ctx, requester.Id, corev1.RolesV1.Group, corev1.RolesV1.Kind, string(ReadAction))
	if err != nil {
		return err
	}

	if !hasRolesPerm {
		ctx.L().Debug("read user roles denied: requester lacks roles read permission")
		return ErrPermissionDenied
	}

	return nil
}

func (a *AuthorizerImpl) IsAuthorizedToReadRolePermissions(ctx *Context, roleID uuid.UUID) error {
	if ctx.IsInternalOperation() {
		return nil
	}

	role, err := a.rReader.GetRole(ctx, roleID)
	if err != nil {
		return err
	}
	requester := ctx.GetAuthedUser()
	requesterHierarchy, err := a.aReader.GetUserHierarchy(ctx, requester.Id)
	if err != nil {
		return err
	}

	if role != nil && requesterHierarchy <= role.Hierarchy {
		ctx.L().Debug("read role permissions denied: requester hierarchy does not exceed role hierarchy")
		return ErrPermissionDenied
	}

	permissionMeta := corev1.PermissionsV1
	hasPermissionsPerm, err := a.pReader.HasCorePermission(ctx, requester.Id, permissionMeta.Group, permissionMeta.Kind, string(ReadAction))
	if err != nil {
		ctx.L().Error("failed to check if user has permission", zap.Error(err))
		return err
	}

	if !hasPermissionsPerm {
		ctx.L().Debug("read role permissions denied: requester lacks permissions read permission")
		return ErrPermissionDenied
	}

	return nil
}

func (a *AuthorizerImpl) IsAuthorizedToPerformResourceAction(ctx *Context, action Action, group, kind string) error {
	if ctx.IsInternalOperation() {
		return nil
	}

	requester := ctx.GetAuthedUser()
	authed, err := a.pReader.HasResourcePermission(ctx, requester.Id, group, kind, action)
	if err != nil {
		return err
	}

	if !authed {
		ctx.L().Debug("failed to perform action because requester does not have permission",
			zap.String("group", group),
			zap.String("kind", kind),
			zap.String("action", string(action)),
		)
		return ErrPermissionDenied
	}

	return nil
}

func (a *AuthorizerImpl) IsAuthorizedToPerformCoreAction(ctx *Context, action Action, group, kind string) error {
	if ctx.IsInternalOperation() {
		return nil
	}

	requester := ctx.GetAuthedUser()
	authed, err := a.pReader.HasCorePermission(ctx, requester.Id, group, kind, action)
	if err != nil {
		return err
	}

	if !authed {
		ctx.L().Debug("failed to perform core action because requester does not have permission",
			zap.String("group", group),
			zap.String("kind", kind),
			zap.String("action", string(action)),
		)
		return ErrPermissionDenied
	}

	return nil
}

// func (a *AuthorizerImpl) IsAuthorizedToPerformGlobalAction(ctx *Context, action Action, target corev1.Metadata) error {
// 	if ctx.IsInternalOperation() {
// 		return nil
// 	}

// 	requester := ctx.GetAuthedUser()

// 	authed, err := a.pReader.HasPermission(ctx, requester.Id, action)
// 	if err != nil {
// 		return err
// 	}

// 	if !authed {
// 		ctx.L().Debug("failed to perform global action because requester does not have permission",
// 			zap.String("group", target.Group),
// 			zap.String("kind", target.Kind),
// 			zap.String("action", string(action)),
// 		)
// 		return ErrPermissionDenied
// 	}

// 	return nil
// }
