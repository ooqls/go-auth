package aggsv1

import (
	"github.com/ooqls/go-auth/internal/permissionsv1"
	"github.com/ooqls/go-auth/internal/rolesv1"
	"github.com/ooqls/go-auth/internal/usersv1"
)

type RoleAgg struct {
	rolesv1.Role
	Permissions []permissionsv1.Permission `json:"permissions"`
}

type UserAgg struct {
	usersv1.User
	Hash  string    `json:"hash"`
	Roles []RoleAgg `json:"roles"`
}
