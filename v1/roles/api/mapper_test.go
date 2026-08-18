package rolesapi

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/v1/roles"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleRole() roles.Role {
	return roles.Role{
		Object: corev1.Object{
			Metadata:  corev1.Metadata{Group: "roles", Kind: "Role"},
			Id:        uuid.New(),
			Name:      "admin",
			CreatedAt: time.Now().Add(-time.Hour),
			UpdatedAt: time.Now(),
		},
		Description: "Administrator role",
		Hierarchy:   10,
	}
}

func TestToGenRole_MapsAllFields(t *testing.T) {
	domain := sampleRole()

	gen := toGenRole(domain)

	require.NotNil(t, gen.Id)
	assert.Equal(t, domain.Id, *gen.Id)
	assert.Equal(t, domain.Name, gen.Name)
	assert.Equal(t, domain.Description, gen.Description)
	assert.Equal(t, int(domain.Hierarchy), gen.Hierarchy)
	assert.Equal(t, domain.CreatedAt, gen.CreatedAt)
	assert.Equal(t, domain.UpdatedAt, gen.UpdatedAt)
}

func TestToGenRoleList_MapsEachElement(t *testing.T) {
	a := sampleRole()
	b := sampleRole()

	gen := toGenRoleList([]roles.Role{a, b})

	require.Len(t, gen, 2)
	assert.Equal(t, a.Name, gen[0].Name)
	assert.Equal(t, b.Name, gen[1].Name)
	require.NotNil(t, gen[0].Id)
	require.NotNil(t, gen[1].Id)
	assert.NotSame(t, gen[0].Id, gen[1].Id)
}

func TestToGenRoleList_Empty(t *testing.T) {
	assert.Empty(t, toGenRoleList(nil))
}
