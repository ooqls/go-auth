package permissionbindings

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/corev1"
	bindingsdata "github.com/ooqls/go-auth/internal/permissionbindingsv1/datagen"
	v1 "github.com/ooqls/go-auth/v1"
)

type bindingMutator interface {
	// Mutations return nil or a typed error from the root v1 package.
	AssignPermissions(ctx *authorizationv1.Context, roleIDs []uuid.UUID, permissions []string) error
	UnassignPermissions(ctx *authorizationv1.Context, roleIDs []uuid.UUID, permissions []string) error
}

type mutator struct {
	transactionRunner transaction
}

func (m *mutator) AssignPermissions(ctx *authorizationv1.Context, roleIDs []uuid.UUID, permissions []string) error {
	return m.mutate(ctx, roleIDs, permissions, "assign")
}

func (m *mutator) UnassignPermissions(ctx *authorizationv1.Context, roleIDs []uuid.UUID, permissions []string) error {
	return m.mutate(ctx, roleIDs, permissions, "unassign")
}

// mutate owns permission-binding policy. Authorization and writes use the same
// fresh snapshot, while internal packages only supply transactional data access.
func (m *mutator) mutate(ctx *authorizationv1.Context, roleIDs []uuid.UUID, permissions []string, action string) error {
	if len(roleIDs) == 0 {
		return nil
	}
	return m.transactionRunner.Run(ctx, func(data transactionData) error {
		var hierarchy int32
		if !ctx.IsInternalOperation() {
			requester := ctx.GetAuthedUser().Id
			grants, err := data.Permissions.GetPermissionsForUser(ctx, requester)
			if err != nil {
				return v1.ErrInternal(fmt.Errorf("read requester permissions: %w", err), nil)
			}
			if !hasCoreGrant(grants, corev1.RolesV1, "update") || !hasCoreGrant(grants, corev1.PermissionsV1, action) {
				return v1.ErrPermissionDenied(errors.New("permission-binding mutation denied"), nil)
			}
			hierarchy, err = data.Aggregates.GetUserHierarchy(ctx, requester)
			if errors.Is(err, pgx.ErrNoRows) {
				return v1.ErrPermissionDenied(errors.New("requester has no role hierarchy"), nil)
			}
			if err != nil {
				return v1.ErrInternal(fmt.Errorf("read requester hierarchy: %w", err), nil)
			}
		}

		// Validate every target before writing. Internal operations bypass
		// authorization, but missing roles remain invalid for every caller.
		for _, roleID := range roleIDs {
			role, err := data.Roles.GetRole(ctx, roleID)
			if errors.Is(err, pgx.ErrNoRows) {
				return v1.ErrNotFound(fmt.Errorf("role %s not found", roleID), v1.M{"role_id": roleID})
			}
			if err != nil {
				return v1.ErrInternal(fmt.Errorf("read target role: %w", err), v1.M{"role_id": roleID})
			}
			if !ctx.IsInternalOperation() && role.Hierarchy > hierarchy {
				return v1.ErrPermissionDenied(errors.New("target role exceeds requester hierarchy"), v1.M{"role_id": roleID})
			}
		}

		for _, roleID := range roleIDs {
			for _, permission := range permissions {
				var err error
				if action == "assign" {
					err = data.Bindings.AssignPermission(ctx, bindingsdata.AssignPermissionParams{RoleID: roleID, Permission: permission})
				} else {
					err = data.Bindings.UnassignPermission(ctx, bindingsdata.UnassignPermissionParams{RoleID: roleID, Permission: permission})
				}
				if err != nil {
					return v1.ErrInternal(fmt.Errorf("%s permission binding: %w", action, err), v1.M{"role_id": roleID, "permission": permission})
				}
			}
		}
		return nil
	})
}

// hasCoreGrant evaluates stored grants here, rather than asking internal data
// readers to make authorization decisions. Preserve the existing core wildcard
// semantics: core:* grants all core actions, and '*' matches each of four fields.
func hasCoreGrant(grants []string, metadata corev1.Metadata, action string) bool {
	want := [4]string{"core", metadata.Group, metadata.Kind, action}
	for _, grant := range grants {
		if grant == "core:*" {
			return true
		}
		parts := strings.SplitN(grant, ":", 5)
		if len(parts) < len(want) {
			continue
		}
		matches := true
		for i, value := range want {
			if parts[i] != "*" && parts[i] != value {
				matches = false
				break
			}
		}
		if matches {
			return true
		}
	}
	return false
}
