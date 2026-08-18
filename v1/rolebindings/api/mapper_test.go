package rolebindingsapi

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/rolebindingsv1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleBinding() rolebindingsv1.Rolebinding {
	return rolebindingsv1.Rolebinding{
		Metadata: corev1.RoleBindingsV1,
		RoleID:   uuid.New(),
		UserID:   uuid.New(),
	}
}

func TestToGenRoleBinding_MapsIDs(t *testing.T) {
	domain := sampleBinding()

	gen := toGenRoleBinding(domain)

	require.NotNil(t, gen.RoleID)
	require.NotNil(t, gen.UserID)
	assert.Equal(t, domain.RoleID, uuid.UUID(*gen.RoleID))
	assert.Equal(t, domain.UserID, uuid.UUID(*gen.UserID))
}

func TestToGenRoleBinding_UsesContractFieldNames(t *testing.T) {
	domain := sampleBinding()

	b, err := json.Marshal(toGenRoleBinding(domain))
	require.NoError(t, err)

	body := string(b)
	// API contract uses roleID/userID, not the domain's role_id/user_id.
	assert.Contains(t, body, "roleID")
	assert.Contains(t, body, "userID")
	assert.NotContains(t, body, "role_id")
	assert.NotContains(t, body, "user_id")
}

func TestToGenRoleBindingList_MapsEachElement(t *testing.T) {
	a := sampleBinding()
	b := sampleBinding()

	gen := toGenRoleBindingList([]rolebindingsv1.Rolebinding{a, b})

	require.Len(t, gen, 2)
	require.NotNil(t, gen[0].RoleID)
	require.NotNil(t, gen[1].RoleID)
	assert.Equal(t, a.RoleID, uuid.UUID(*gen[0].RoleID))
	assert.Equal(t, b.RoleID, uuid.UUID(*gen[1].RoleID))
	assert.NotSame(t, gen[0].RoleID, gen[1].RoleID)
}

func TestToGenRoleBindingList_Empty(t *testing.T) {
	assert.Empty(t, toGenRoleBindingList(nil))
}
