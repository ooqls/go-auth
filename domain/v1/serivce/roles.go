package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/ooqls/go-auth/domain/v1/authorization"
	"github.com/ooqls/go-auth/records"
	"github.com/ooqls/go-auth/records/v1/roles"
)

type RolesService interface {
	CreateRole(ctx authorization.Context, roleName, description string) error
	DeleteRole(ctx authorization.Context, id records.RoleId) error
	GetRole(ctx authorization.Context, id records.RoleId) (*records.Role, error)
	ListRoles(ctx authorization.Context) ([]records.Role, error)
	UpdateRole(ctx authorization.Context, id records.RoleId, roleName, description string) error
}

type RolesServiceImpl struct {
	rr roles.Reader
	rw roles.Writer
	ra authorization.RoleAuthorizer
}

func NewRolesService(rr roles.Reader, rw roles.Writer, ra authorization.RoleAuthorizer) *RolesServiceImpl {
	return &RolesServiceImpl{rr: rr, rw: rw, ra: ra}
}

func (r *RolesServiceImpl) CreateRole(ctx authorization.Context, roleName, description string, hierarchy int32) error {
	role := records.Role{
		ID:            uuid.New(),
		Domain:        ctx.Domain,
		RoleName:      roleName,
		Description:   description,
		RoleHierarchy: hierarchy,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := r.ra.IsAuthorizedToPerformRoleAction(&ctx, authorization.CreateAction, role); err != nil {
		return err
	}

	_, err := r.rw.CreateRole(ctx, role)
	if err != nil {
		return err
	}

	return nil
}

func (r *RolesServiceImpl) DeleteRole(ctx authorization.Context, id records.RoleId) error {
	role, err := r.rr.GetRole(ctx, id)
	if err != nil {
		return err
	}

	if err := r.ra.IsAuthorizedToPerformRoleAction(&ctx, authorization.DeleteAction, *role); err != nil {
		return err
	}

	return r.rw.DeleteRole(ctx, id)
}

func (r *RolesServiceImpl) GetRole(ctx authorization.Context, id records.RoleId) (*records.Role, error) {
	role, err := r.rr.GetRole(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := r.ra.IsAuthorizedToPerformRoleAction(&ctx, authorization.ReadAction, *role); err != nil {
		return nil, err
	}

	return role, nil
}

func (r *RolesServiceImpl) ListRoles(ctx authorization.Context, limit, offset int32) ([]records.Role, error) {
	if err := r.ra.IsAuthorizedToPerformRoleAction(&ctx, authorization.ReadAction, records.Role{}); err != nil {
		return nil, err
	}

	roles, err := r.rr.GetRoles(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *RolesServiceImpl) UpdateRole(ctx authorization.Context, id records.RoleId, roleName, description string, hierarchy int32) error {
	role, err := r.rr.GetRole(ctx, id)
	if err != nil {
		return err
	}

	role.RoleName = roleName
	role.Description = description
	role.RoleHierarchy = hierarchy
	role.UpdatedAt = time.Now()

	if err := r.ra.IsAuthorizedToPerformRoleAction(&ctx, authorization.UpdateAction, *role); err != nil {
		return err
	}

	_, err = r.rw.UpdateRole(ctx, id, *role)
	if err != nil {
		return err
	}

	return nil
}

func (r *RolesServiceImpl) AssignRole(ctx authorization.Context, id records.RoleId, userID records.UserId) error {
	role, err := r.rr.GetRole(ctx, id)
	if err != nil {
		return err
	}

	if err := r.ra.IsAuthorizedToPerformRoleAction(&ctx, authorization.AssignAction, *role); err != nil {
		return err
	}

	return r.rw.AddRoleToUser(ctx, userID, id)
}

func (r *RolesServiceImpl) UnassignRole(ctx authorization.Context, id records.RoleId, userID records.UserId) error {
	role, err := r.rr.GetRole(ctx, id)
	if err != nil {
		return err
	}

	if err := r.ra.IsAuthorizedToPerformRoleAction(&ctx, authorization.UnassignAction, *role); err != nil {
		return err
	}

	return r.rw.RemoveRoleFromUser(ctx, userID, id)
}
