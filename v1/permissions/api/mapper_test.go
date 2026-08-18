package permissionsapi

import (
	"testing"

	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/permissionsv1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToGenPermission_MapsPermission(t *testing.T) {
	domain := permissionsv1.Permission{
		Metadata:   corev1.PermissionsV1,
		Permission: "resource:read",
	}

	gen := toGenPermission(domain)

	assert.Equal(t, domain.Permission, gen.Permission)
}

func TestToGenPermissionList_MapsEachElement(t *testing.T) {
	perms := []permissionsv1.Permission{
		{Permission: "resource:read"},
		{Permission: "resource:write"},
	}

	gen := toGenPermissionList(perms)

	require.Len(t, gen, 2)
	assert.Equal(t, "resource:read", gen[0].Permission)
	assert.Equal(t, "resource:write", gen[1].Permission)
}

func TestToGenPermissionList_Empty(t *testing.T) {
	assert.Empty(t, toGenPermissionList(nil))
}
