package records

import (
	"github.com/google/uuid"
	"github.com/ooqls/go-auth/records/gen"
	"github.com/ooqls/go-eventsource/eventsourcingv1"
)

func ParseUserId(id string) (UserId, error) {
	return uuid.Parse(id)
}

type RoleAgg struct {
	RoleId        RoleId
	RoleHierarchy int32
	Permissions   []Permission
}

type UserAgg struct {
	UserId UserId
	Roles  []RoleAgg
}

//go:generate sqlc generate --file sqlc/sqlc.yaml
type User = gen.Authv1User
type Role = gen.Authv1Role

type UserId = uuid.UUID

func NewUserID() UserId {
	return uuid.New()
}

func NewRoleID() RoleId {
	return uuid.New()
}

type RoleId = uuid.UUID
type PermissionId = uuid.UUID

type Permission = gen.Authv1Permission
type Resource = gen.Authv1Resource

const (
	UserEventSource = eventsourcingv1.EventSource("user-events")
	RoleEventSource = eventsourcingv1.EventSource("role-events")
)
