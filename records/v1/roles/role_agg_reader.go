package roles

import (
	"context"

	authv1 "github.com/ooqls/go-auth/records"
	"github.com/ooqls/go-auth/records/v1/gen"
	"go.uber.org/zap"
)

type RoleAgg = authv1.RoleAgg
type RoleId = authv1.RoleId
type Role = authv1.Role
type Permission = authv1.Permission
type UserId = authv1.UserId

type AggRoleReader interface {
	GetRoleAggForUser(ctx context.Context, id UserId) (*RoleAgg, error)
}

type AggRoleReaderImpl struct {
	l *zap.Logger
	q *gen.Queries
}

func NewAggRoleReaderImpl(l *zap.Logger, q *gen.Queries) *AggRoleReaderImpl {
	return &AggRoleReaderImpl{
		l: l,
		q: q,
	}
}

func (r *AggRoleReaderImpl) GetRoleAggForUser(ctx context.Context, id UserId) ([]RoleAgg, error) {
	roleAggs, err := r.q.GetRoleAggregate(ctx, id)
	if err != nil {
		return nil, err
	}

	roleMap := map[string]*RoleAgg{}
	roles := make([]RoleAgg, 0)

	for _, r := range roleAggs {
		aggR, ok := roleMap[r.RoleID.UUID.String()]
		if !ok {
			newAggRole := &RoleAgg{
				RoleId:        RoleId(r.RoleID.UUID),
				RoleHierarchy: int32(r.RoleHierarchy),
				Permissions: []Permission{
					{
						ID:            r.ID.UUID,
						ResourceKind:  r.ResourceKind.String,
						ResourceGroup: r.ResourceGroup.String,
						ResourceName:  r.ResourceName.String,
						Actions:       r.Actions.String,
						CreatedAt:     r.CreatedAt.Time,
						UpdatedAt:     r.UpdatedAt.Time,
					},
				},
			}
			roleMap[r.RoleID.UUID.String()] = newAggRole
			roles = append(roles, *newAggRole)
		} else {
			aggR.Permissions = append(aggR.Permissions, Permission{
				ID:            r.ID.UUID,
				ResourceKind:  r.ResourceKind.String,
				ResourceGroup: r.ResourceGroup.String,
				ResourceName:  r.ResourceName.String,
				Actions:       r.Actions.String,
				CreatedAt:     r.CreatedAt.Time,
				UpdatedAt:     r.UpdatedAt.Time,
			})
		}
	}
	return roles, nil
}
