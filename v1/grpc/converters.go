package grpc

import (
	"strings"
	"time"

	"github.com/google/uuid"
	commonpb "github.com/ooqls/go-auth/gen/grpc/common/v1"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/permissionbindingsv1"
	"github.com/ooqls/go-auth/internal/permissionsv1"
	"github.com/ooqls/go-auth/internal/resourcesv1"
	"github.com/ooqls/go-auth/internal/rolebindingsv1"
	"github.com/ooqls/go-auth/internal/rolesv1"
	"github.com/ooqls/go-auth/internal/usersv1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// --- Timestamp helpers ---

func TimestampToProto(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

func TimestampFromProto(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime()
}

// --- User converters ---

func UserToProto(u *usersv1.User) *commonpb.User {
	if u == nil {
		return nil
	}
	return &commonpb.User{
		Id:        u.Id.String(),
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: TimestampToProto(u.CreatedAt),
		UpdatedAt: TimestampToProto(u.UpdatedAt),
	}
}

// --- Role converters ---

func RoleToProto(r *rolesv1.Role) *commonpb.Role {
	if r == nil {
		return nil
	}
	return &commonpb.Role{
		Id:          r.Id.String(),
		Name:        r.Name,
		Description: r.Description,
		Hierarchy:   r.Hierarchy,
		CreatedAt:   TimestampToProto(r.CreatedAt),
		UpdatedAt:   TimestampToProto(r.UpdatedAt),
	}
}

func RolesToProto(roles []rolesv1.Role) []*commonpb.Role {
	result := make([]*commonpb.Role, 0, len(roles))
	for i := range roles {
		result = append(result, RoleToProto(&roles[i]))
	}
	return result
}

// --- Permission converters ---

func PermissionToProto(p *permissionsv1.Permission) *commonpb.Permission {
	if p == nil {
		return nil
	}
	actions := strings.Split(p.Actions, ",")
	return &commonpb.Permission{
		Id:            p.Id.String(),
		ResourceKind:  p.Resource.Metadata.Kind,
		ResourceGroup: p.Resource.Metadata.Group,
		ResourceName:  p.Resource.Name,
		Actions:       actions,
		CreatedAt:     TimestampToProto(p.CreatedAt),
		UpdatedAt:     TimestampToProto(p.UpdatedAt),
	}
}

func PermissionsToProto(perms []permissionsv1.Permission) []*commonpb.Permission {
	result := make([]*commonpb.Permission, 0, len(perms))
	for i := range perms {
		result = append(result, PermissionToProto(&perms[i]))
	}
	return result
}

// --- Resource converters ---

func ResourceToProto(r *resourcesv1.Resourcev1) *commonpb.Resource {
	if r == nil {
		return nil
	}
	return &commonpb.Resource{
		Id:          r.Id.String(),
		Name:        r.Name,
		Group:       r.Metadata.Group,
		Kind:        r.Metadata.Kind,
		Description: r.Description,
		CreatedAt:   TimestampToProto(r.CreatedAt),
		UpdatedAt:   TimestampToProto(r.UpdatedAt),
	}
}

func ResourcesToProto(resources []resourcesv1.Resourcev1) []*commonpb.Resource {
	result := make([]*commonpb.Resource, 0, len(resources))
	for i := range resources {
		result = append(result, ResourceToProto(&resources[i]))
	}
	return result
}

// --- RoleBinding converters ---

func RoleBindingToProto(rb *rolebindingsv1.Rolebinding) *commonpb.RoleBinding {
	if rb == nil {
		return nil
	}
	return &commonpb.RoleBinding{
		RoleId: rb.RoleID.String(),
		UserId: rb.UserID.String(),
	}
}

func RoleBindingsToProto(rbs []rolebindingsv1.Rolebinding) []*commonpb.RoleBinding {
	result := make([]*commonpb.RoleBinding, 0, len(rbs))
	for i := range rbs {
		result = append(result, RoleBindingToProto(&rbs[i]))
	}
	return result
}

// --- PermissionBinding converters ---

func PermissionBindingToProto(pb *permissionbindingsv1.Permissionbindingv1) *commonpb.PermissionBinding {
	if pb == nil {
		return nil
	}
	return &commonpb.PermissionBinding{
		RoleId:       pb.RoleID.String(),
		PermissionId: pb.PermissionID.String(),
	}
}

func PermissionBindingsToProto(pbs []permissionbindingsv1.Permissionbindingv1) []*commonpb.PermissionBinding {
	result := make([]*commonpb.PermissionBinding, 0, len(pbs))
	for i := range pbs {
		result = append(result, PermissionBindingToProto(&pbs[i]))
	}
	return result
}

// --- UUID helpers ---

func ParseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

func ParseUUIDs(ss []string) ([]uuid.UUID, error) {
	uuids := make([]uuid.UUID, 0, len(ss))
	for _, s := range ss {
		id, err := uuid.Parse(s)
		if err != nil {
			return nil, err
		}
		uuids = append(uuids, id)
	}
	return uuids, nil
}

// --- Object converters ---

func ObjectToProto(o corev1.Object) *commonpb.Resource {
	return &commonpb.Resource{
		Id:        o.Id.String(),
		Name:      o.Name,
		Group:     o.Metadata.Group,
		Kind:      o.Metadata.Kind,
		CreatedAt: TimestampToProto(o.CreatedAt),
		UpdatedAt: TimestampToProto(o.UpdatedAt),
	}
}
