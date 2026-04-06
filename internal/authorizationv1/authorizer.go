package authorizationv1

//go:generate go run github.com/golang/mock/mockgen -source=authorizer.go -destination=mocks/mock_authorizer.go -package=mocks

import (
	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/aggsv1"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/datav1"
	"github.com/ooqls/go-auth/internal/permissionbindingsv1"
	"github.com/ooqls/go-auth/internal/permissionsv1"
	"github.com/ooqls/go-auth/internal/rolebindingsv1"
	"github.com/ooqls/go-auth/internal/rolesv1"
	"go.uber.org/zap"
)

type Authorizer interface {
	IsAuthorizedToAssignPermission(ctx *Context, perm permissionsv1.Permission, role rolesv1.Role) error
	IsAuthorizedToUnassignPermission(ctx *Context, target permissionsv1.Permission, role rolesv1.Role) error
	IsAuthorizedToAssignRole(ctx *Context, user uuid.UUID, role aggsv1.RoleAgg) error
	IsAuthorizedToReadUserRoles(ctx *Context, user uuid.UUID) error
	IsAuthorizedToReadRolePermissions(ctx *Context, roleID uuid.UUID) error
	IsAuthorizedToPerformAction(ctx *Context, action Action, target corev1.Object) error
}

type AuthorizerImpl struct {
	pReader  permissionsv1.Reader
	rbReader rolebindingsv1.Reader
	aReader  aggsv1.Reader
}

func NewAuthorizerImpl(fact datav1.Factory) Authorizer {
	return &AuthorizerImpl{
		pReader:  fact.NewPermissionReader(),
		rbReader: fact.NewRoleBindingsReader(),
		aReader:  fact.NewAggReader(),
	}
}

func (a *AuthorizerImpl) IsAuthorizedToAssignPermission(ctx *Context, perm permissionsv1.Permission, role rolesv1.Role) error {
	// todo:
	// use the rolebindingsv1.Reader to determine
	userHierarchy, err := a.aReader.GetUserHierarchy(ctx, ctx.GetAuthedUser().Id)
	if err != nil {
		return err
	}

	if role.Hierarchy > userHierarchy {
		return ErrPermissionDenied
	}

	hasPerm, err := a.pReader.HasPermission(ctx, ctx.GetAuthedUser().Id, perm.Name, perm.Group, perm.Kind, perm.Actions)
	if err != nil {
		return err
	}

	if !hasPerm {
		return ErrPermissionDenied
	}

	return nil

}

func (a *AuthorizerImpl) IsAuthorizedToUnassignPermission(ctx *Context, target permissionsv1.Permission, role rolesv1.Role) error {
	authed, err := a.pReader.HasPermission(ctx, ctx.GetAuthedUser().Id, target.Group, target.Kind, target.Name, target.Actions)
	if err != nil {
		return err
	}

	if !authed {
		return ErrPermissionDenied
	}

	return nil
}

func (a *AuthorizerImpl) IsAuthorizedToAssignRole(ctx *Context, user uuid.UUID, role aggsv1.RoleAgg) error {
	rbMeta := rolebindingsv1.Metadata
	authed, err := a.pReader.HasPermission(ctx, ctx.GetAuthedUser().Id, "*", rbMeta.Group, rbMeta.Kind, string(AssignAction))
	if err != nil {
		return err
	}

	if !authed {
		ctx.L().Debug("role assignment failed because user does not have the permission to assign roles")
		return ErrPermissionDenied
	}

	assignerHierarchy, err := a.aReader.GetUserHierarchy(ctx, ctx.GetAuthedUser().Id)
	if err != nil {
		return err
	}

	assigneeHierarchy, err := a.aReader.GetUserHierarchy(ctx, user)
	if err != nil {
		return err
	}

	if assignerHierarchy < assigneeHierarchy {
		ctx.L().Debug("Role assignment failed because assignee has a higher user hierarchy than assigner")
		return ErrPermissionDenied
	}

	return nil
}

func (a *AuthorizerImpl) IsAuthorizedToPerformAction(ctx *Context, action Action, target corev1.Object) error {
	authed, err := a.pReader.HasPermission(ctx, ctx.GetAuthedUser().Id, target.Name, target.Group, target.Kind, action)
	if err != nil {
		return err
	}

	if !authed {
		ctx.L().Debug("failed to perform action because requester does not have permission",
			zap.String("name", target.Name),
			zap.String("group", target.Group),
			zap.String("kind", target.Kind),
		)
		return ErrPermissionDenied
	}

	return nil
}

func (a *AuthorizerImpl) IsAuthorizedToReadUserRoles(ctx *Context, userID uuid.UUID) error {
	if ctx.IsInternalOperation() {
		return nil
	}

	requesterHierarchy, err := a.aReader.GetUserHierarchy(ctx, ctx.GetAuthedUser().Id)
	if err != nil {
		return err
	}

	targetHierarchy, err := a.aReader.GetUserHierarchy(ctx, userID)
	if err != nil {
		return err
	}

	if requesterHierarchy < targetHierarchy {
		ctx.L().Debug("read user roles denied: requester hierarchy is lower than target user hierarchy")
		return ErrPermissionDenied
	}

	rbMeta := rolebindingsv1.Metadata
	hasRolebindingsPerm, err := a.pReader.HasPermission(ctx, ctx.GetAuthedUser().Id, "*", rbMeta.Group, rbMeta.Kind, string(ReadAction))
	if err != nil {
		return err
	}

	if !hasRolebindingsPerm {
		ctx.L().Debug("read user roles denied: requester lacks rolebindings read permission")
		return ErrPermissionDenied
	}

	hasRolesPerm, err := a.pReader.HasPermission(ctx, ctx.GetAuthedUser().Id, "*", "roles", rolesv1.Kind, string(ReadAction))
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
	roleAgg, err := a.aReader.GetRoleAgg(ctx, roleID)
	if err != nil {
		return err
	}

	requesterHierarchy, err := a.aReader.GetUserHierarchy(ctx, ctx.GetAuthedUser().Id)
	if err != nil {
		return err
	}

	if roleAgg != nil && requesterHierarchy <= roleAgg.Hierarchy {
		ctx.L().Debug("read role permissions denied: requester hierarchy does not exceed role hierarchy")
		return ErrPermissionDenied
	}

	pbMeta := permissionbindingsv1.Metadata
	hasBindingsPerm, err := a.pReader.HasPermission(ctx, ctx.GetAuthedUser().Id, "*", pbMeta.Group, pbMeta.Kind, string(ReadAction))
	if err != nil {
		return err
	}

	if !hasBindingsPerm {
		ctx.L().Debug("read role permissions denied: requester lacks permission bindings read permission")
		return ErrPermissionDenied
	}

	pMeta := permissionsv1.Metadata
	hasPermsPerm, err := a.pReader.HasPermission(ctx, ctx.GetAuthedUser().Id, "*", pMeta.Group, pMeta.Kind, string(ReadAction))
	if err != nil {
		return err
	}

	if !hasPermsPerm {
		ctx.L().Debug("read role permissions denied: requester lacks permissions read permission")
		return ErrPermissionDenied
	}

	return nil
}
