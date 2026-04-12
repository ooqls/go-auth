package roles

import (
	"slices"

	"github.com/ooqls/go-auth/internal/aggsv1"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/datav1"
	"github.com/ooqls/go-auth/internal/rolesv1"
	v1 "github.com/ooqls/go-auth/v1"
	"github.com/pkg/errors"
)

var _ Service = &ServiceImpl{}

type Service interface {
	CreateRole(ctx *authorizationv1.Context, params CreateRoleParams) (*Id, error)
	DeleteRole(ctx *authorizationv1.Context, id Id) error
	GetRole(ctx *authorizationv1.Context, id Id) (*Role, error)
	GetRoleAgg(ctx *authorizationv1.Context, id Id) (*RoleAgg, error)
	GetRoleByName(ctx *authorizationv1.Context, name string) (*Role, error)
	ListRoles(ctx *authorizationv1.Context, page, pageSize int32) ([]Role, error)
	UpdateRole(ctx *authorizationv1.Context, params UpdateRoleParams) error
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
	params CreateRoleParams) (*Id, error) {

	role := params.ToRole()

	if err := r.ra.IsAuthorizedToPerformAction(ctx, authorizationv1.CreateAction, role.Object); err != nil {
		return nil, v1.ErrPermissionDenied(err, v1.M{"params": params})
	}

	createdRole, err := r.rw.CreateRole(ctx, params.Name, params.Description, params.Hierarchy)
	if err != nil {
		return nil, v1.ErrInternal(err,
			v1.M{
				"params": params,
			})
	}

	return &createdRole.Id, nil
}

func (r *ServiceImpl) DeleteRole(ctx *authorizationv1.Context, id Id) error {
	role, err := r.rr.GetRole(ctx, id)
	if err != nil {
		return errors.Wrap(err, "failed to get role")
	}

	if err := r.ra.IsAuthorizedToPerformAction(ctx,
		authorizationv1.DeleteAction,
		role.Object); err != nil {
		return errors.Wrap(err, "failed to authorize role action")
	}

	return r.rw.DeleteRole(ctx, id)
}

func (r *ServiceImpl) GetRole(ctx *authorizationv1.Context, id Id) (*rolesv1.Role, error) {
	role, err := r.rr.GetRole(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get role")
	}

	if err := r.ra.IsAuthorizedToPerformAction(ctx, authorizationv1.ReadAction, role.Object); err != nil {
		return nil, errors.Wrap(err, "failed to authorize role action")
	}

	return role, nil
}

func (r *ServiceImpl) GetRoleAgg(ctx *authorizationv1.Context, id Id) (*RoleAgg, error) {
	err := r.ra.IsAuthorizedToReadUserRoles(ctx, id)
	if err != nil {
		return nil, err
	}

	roleAgg, err := r.GetRoleAgg(ctx, id)
	if err != nil {
		return nil, err
	}

	return roleAgg, nil
}

// func (r *ServiceImpl) GetRoleAgg(ctx authorization.Context, id records.RoleId) (*records.RoleAgg, error) {
// 	if err := r.ra.IsAuthorizedToPerformRoleAction(ctx,
// 		authorization.ReadAction,
// 		records.Role{ID: id}); err != nil {
// 		return nil, errors.Wrap(err, "failed to authorize role action")
// 	}

// 	roleAgg, err := r.rar.GetRoleAgg(&ctx, id)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "failed to get role agg")
// 	}

// 	return roleAgg, nil
// }

func (r *ServiceImpl) GetRoleByName(ctx *authorizationv1.Context, name string) (*rolesv1.Role, error) {
	datagenRole, err := r.rr.GetRoleByName(ctx, name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get role by name")
	}
	return datagenRole, nil
}

func (r *ServiceImpl) ListRoles(ctx *authorizationv1.Context, page, pageSize int32) ([]rolesv1.Role, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 100
	}
	roles, err := r.rr.GetRoles(ctx, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get roles")
	}

	roles = slices.DeleteFunc(roles, func(role rolesv1.Role) bool {
		if err := r.ra.IsAuthorizedToPerformAction(ctx, authorizationv1.ReadAction, role.Object); err != nil {
			return true
		}
		return false
	})

	return roles, nil
}

func (r *ServiceImpl) UpdateRole(ctx *authorizationv1.Context, params UpdateRoleParams) error {
	curRole, err := r.rr.GetRole(ctx, params.ID)
	if err != nil {
		return errors.Wrap(err, "failed to get role")
	}

	if curRole == nil {
		return v1.ErrNotFound(err, v1.M{
			"role_id": params.ID,
		})
	}

	if err := r.ra.IsAuthorizedToPerformAction(ctx, authorizationv1.UpdateAction, curRole.Object); err != nil {
		return errors.Wrap(err, "failed to authorize role action")
	}

	_, err = r.rw.UpdateRole(ctx, params.ID, params.Name, params.Description, params.Hierarchy)
	if err != nil {
		return errors.Wrap(err, "failed to update role")
	}

	return nil
}

// func (r *ServiceImpl) AssignRole(ctx authorization.Context, id records.RoleId, userID records.UserId) error {
// 	role, err := r.rr.GetRole(&ctx, id)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to get role")
// 	}

// 	if err := r.ra.IsAuthorizedToPerformRoleAction(ctx, authorization.AssignAction, *role); err != nil {
// 		return errors.Wrap(err, "failed to authorize role action")
// 	}

// 	return r.rw.AddRoleToUser(ctx, userID, id)
// }

// func (r *ServiceImpl) UnassignRole(ctx authorization.Context, id records.RoleId, userID records.UserId) error {
// 	role, err := r.rr.GetRole(&ctx, id)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to get role")
// 	}

// 	if err := r.ra.IsAuthorizedToPerformRoleAction(ctx, authorization.UnassignAction, *role); err != nil {
// 		return errors.Wrap(err, "failed to authorize role action")
// 	}

// 	return r.rw.RemoveRoleFromUser(ctx, userID, id)
// }

// func (r *ServiceImpl) GetAssignedRoles(ctx authorization.Context, id records.UserId) error {
// 	roles, err := r.rr.GetRolesForUser(&ctx, id)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to get roles for user")
// 	}

// 	if err := r.ra.IsAuthorizedToPerformRoleAction(ctx, v1.ReadAction, roles...); err != nil {
// 		return errors.Wrap(err, "failed to authorize role action")
// 	}

// 	return nil
// }
