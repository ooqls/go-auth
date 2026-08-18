package usersapi

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/v1/users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleUser() users.User {
	return users.User{
		Object: corev1.Object{
			Metadata:  corev1.Metadata{Group: corev1.Group, Kind: "User"},
			Id:        uuid.New(),
			Name:      "alice",
			CreatedAt: time.Now().Add(-time.Hour),
			UpdatedAt: time.Now(),
		},
		Username: "alice",
		Email:    "alice@example.com",
		Key:      []byte("super-secret-key"),
		Salt:     []byte("super-secret-salt"),
	}
}

func TestToGenUser_MapsAllFields(t *testing.T) {
	domain := sampleUser()

	gen := toGenUser(domain)

	require.NotNil(t, gen.Id)
	assert.Equal(t, domain.Id, *gen.Id)
	assert.Equal(t, domain.Username, gen.Username)
	assert.Equal(t, domain.Email, gen.Email)
	assert.Equal(t, domain.CreatedAt, gen.CreatedAt)
	assert.Equal(t, domain.UpdatedAt, gen.UpdatedAt)
}

func TestToGenUser_DoesNotLeakKeyOrSalt(t *testing.T) {
	domain := sampleUser()

	b, err := json.Marshal(toGenUser(domain))
	require.NoError(t, err)

	body := string(b)
	assert.NotContains(t, body, "super-secret-key")
	assert.NotContains(t, body, "super-secret-salt")
	assert.NotContains(t, body, "key")
	assert.NotContains(t, body, "salt")
}

func TestToGenUserList_MapsEachElement(t *testing.T) {
	a := sampleUser()
	b := sampleUser()

	gen := toGenUserList([]users.User{a, b})

	require.Len(t, gen, 2)
	assert.Equal(t, a.Username, gen[0].Username)
	assert.Equal(t, b.Username, gen[1].Username)
	require.NotNil(t, gen[0].Id)
	require.NotNil(t, gen[1].Id)
	assert.NotSame(t, gen[0].Id, gen[1].Id)
}

func TestToGenUserList_Empty(t *testing.T) {
	assert.Empty(t, toGenUserList(nil))
}
