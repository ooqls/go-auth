package roles

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/aggsv1"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/datav1"
	"github.com/ooqls/go-auth/internal/rolesv1"
	v1 "github.com/ooqls/go-auth/v1"
)

var _ Service = &ServiceImpl{}

type Service interface {
	CreateRole(ctx *authorizationv1.Context, name, description string, hierarchy int32) (*Id, error)
	DeleteRole(ctx *authorizationv1.Context, id Id) error
	GetRole(ctx *authorizationv1.Context, id Id) (*Role, error)
	GetRoleAgg(ctx *authorizationv1.Context, id Id) (*RoleAgg, error)
	GetRoleByName(ctx *authorizationv1.Context, name string) (*Role, error)
	ListRoles(ctx *authorizationv1.Context, page, pageSize int32) (*corev1.Result[[]rolesv1.Role], error)
	UpdateRole(ctx *authorizationv1.Context, roleId uuid.UUID, name, description *string, hierarchy *int32) error
}

type ServiceImpl struct {
	rr rolesv1.Reader
	rw rolesv1.Writer
	ar aggsv1.Reader
	ra authorizationv1.Authorizer
}

func NewServiceImpl(
	factory datav1.Factory,
) *ServiceImpl {
	return &ServiceImpl{
		rr: factory.NewRoleReader(),
		rw: factory.NewRoleWriter(),
		ar: factory.NewAggReader(),
		ra: authorizationv1.NewAuthorizerImpl(factory),
	}
}

func (r *ServiceImpl) CreateRole(
	ctx *authorizationv1.Context,
	name, description string, hierarchy int32) (*Id, error) {

	if err := r.ra.IsAuthorizedToPerformCoreAction(ctx, authorizationv1.CreateAction, corev1.RolesV1.Group, corev1.RolesV1.Kind); err != nil {
		return nil, v1.ErrPermissionDenied(err, v1.M{"name": name, "description": description, "hierarchy": hierarchy})
	}

	if err := r.ra.HasHigherHierarchy(ctx, hierarchy); err != nil {
		if err == authorizationv1.ErrPermissionDenied {
			return nil, v1.ErrPermissionDenied(err, v1.M{"name": name, "description": description, "hierarchy": hierarchy})
		}
		return nil, v1.ErrInternal(err,
			v1.M{
				"name": name,
			})
	}

	createdRole, err := r.rw.CreateRole(ctx, name, description, hierarchy)
	if err != nil {
		return nil, v1.ErrInternal(err,
			v1.M{
				"name": name,
			})
	}

	return &createdRole.Id, nil
}

func (r *ServiceImpl) DeleteRole(ctx *authorizationv1.Context, id Id) error {
	if err := r.ra.IsAuthorizedToPerformCoreAction(ctx,
		authorizationv1.DeleteAction,
		corev1.RolesV1.Group, corev1.RolesV1.Kind); err != nil {
		return v1.ErrPermissionDenied(err, v1.M{"id": id})
	}

	return r.rw.DeleteRole(ctx, id)
}

func (r *ServiceImpl) GetRole(ctx *authorizationv1.Context, id Id) (*rolesv1.Role, error) {
	if err := r.ra.IsAuthorizedToPerformCoreAction(ctx,
		authorizationv1.ReadAction,
		corev1.RolesV1.Group, corev1.RolesV1.Kind); err != nil {
		return nil, v1.ErrPermissionDenied(err, v1.M{"id": id})
	}

	role, err := r.rr.GetRole(ctx, id)
	if err != nil {
		return nil, v1.ErrInternal(err, v1.M{"id": id})
	}

	return role, nil
}

func (r *ServiceImpl) GetRoleAgg(ctx *authorizationv1.Context, id Id) (*RoleAgg, error) {
	if err := r.ra.IsAuthorizedToPerformCoreAction(ctx,
		authorizationv1.ReadAction,
		corev1.RolesV1.Group, corev1.RolesV1.Kind); err != nil {
		return nil, v1.ErrPermissionDenied(err, v1.M{"id": id})
	}

	err := r.ra.IsAuthorizedToReadRolePermissions(ctx, id)
	if err != nil {
		return nil, v1.ErrPermissionDenied(err, v1.M{"id": id})
	}

	roleAgg, err := r.GetRoleAgg(ctx, id)
	if err != nil {
		return nil, v1.ErrInternal(err, v1.M{"id": id})
	}

	return roleAgg, nil
}

// func (r *ServiceImpl) GetRoleAgg(ctx authorization.Context, id records.RoleId) (*records.RoleAgg, error) {
// 	if err := r.ra.IsAuthorizedToPerformRoleAction(ctx,
// 		authorization.ReadAction,
// 		records.Role{ID: id}); err != nil {
// 		return nil, v1.ErrPermissionDenied(err, v1.M{"id": id})
// 	}

// 	roleAgg, err := r.rar.GetRoleAgg(&ctx, id)
// 	if err != nil {
// 		return nil, v1.ErrInternal(err, v1.M{"id": id})
// 	}

// 	return roleAgg, nil
// }

func (r *ServiceImpl) GetRoleByName(ctx *authorizationv1.Context, name string) (*rolesv1.Role, error) {
	if err := r.ra.IsAuthorizedToPerformCoreAction(ctx, authorizationv1.ReadAction, corev1.RolesV1.Group, corev1.RolesV1.Kind); err != nil {
		return nil, v1.ErrPermissionDenied(err, v1.M{"name": name})
	}

	datagenRole, err := r.rr.GetRoleByName(ctx, name)
	if err != nil {
		return nil, v1.ErrInternal(err, v1.M{"name": name})
	}

	if datagenRole == nil {
		return nil, v1.ErrPermissionDenied(fmt.Errorf("role %s not found", name), v1.M{"name": name})
	}

	return datagenRole, nil
}

func (r *ServiceImpl) ListRoles(ctx *authorizationv1.Context, page, pageSize int32) (*corev1.Result[[]rolesv1.Role], error) {
	if err := r.ra.IsAuthorizedToPerformCoreAction(ctx, authorizationv1.ReadAction, corev1.RolesV1.Group, corev1.RolesV1.Kind); err != nil {
		return nil, v1.ErrPermissionDenied(err, v1.M{"page": page, "pageSize": pageSize})
	}

	roles, err := r.rr.GetRoles(ctx, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, v1.ErrInternal(err, v1.M{"page": page, "pageSize": pageSize})
	}

	return roles, nil
}

func (r *ServiceImpl) UpdateRole(ctx *authorizationv1.Context, roleId uuid.UUID, name, description *string, hierarchy *int32) error {
	if err := r.ra.IsAuthorizedToPerformCoreAction(ctx, authorizationv1.UpdateAction, corev1.RolesV1.Group, corev1.RolesV1.Kind); err != nil {
		return v1.ErrPermissionDenied(err, v1.M{"name": name, "description": description, "hierarchy": hierarchy})
	}

	if hierarchy != nil {
		requesterHierarchy, err := r.ar.GetUserHierarchy(ctx, ctx.GetAuthedUser().Id)
		if err != nil {
			return v1.ErrInternal(err, v1.M{"name": name, "description": description, "hierarchy": hierarchy})
		}
		if requesterHierarchy <= *hierarchy {
			return v1.ErrPermissionDenied(fmt.Errorf("cannot set hierarchy higher than your own"), v1.M{"name": name, "description": description, "hierarchy": hierarchy})
		}
	}

	_, err := r.rw.UpdateRole(ctx, roleId, name, description, hierarchy)
	if err != nil {
		return v1.ErrInternal(err, v1.M{"name": name, "description": description})
	}

	return nil
}
