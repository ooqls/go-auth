package resourcesapi

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/resourcesv1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleResource() resourcesv1.Resourcev1 {
	return resourcesv1.Resourcev1{
		Metadata: corev1.Metadata{
			Group: "test-group",
			Kind:  "test-kind",
		},
		Id:        uuid.New(),
		Name:      "test-resource",
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now(),
	}
}

func TestToGenResource_MapsAllFields(t *testing.T) {
	domain := sampleResource()

	gen := toGenResource(domain)

	require.NotNil(t, gen.Id)
	assert.Equal(t, domain.Id, *gen.Id)
	assert.Equal(t, domain.Name, gen.Name)
	assert.Equal(t, domain.Group, gen.Group)
	assert.Equal(t, domain.Kind, gen.Kind)
	assert.Equal(t, domain.CreatedAt, gen.CreatedAt)
	require.NotNil(t, gen.UpdatedAt)
	assert.Equal(t, domain.UpdatedAt, *gen.UpdatedAt)
}

func TestToGenResourceList_MapsEachElement(t *testing.T) {
	a := sampleResource()
	b := sampleResource()

	gen := toGenResourceList([]resourcesv1.Resourcev1{a, b})

	require.Len(t, gen, 2)
	assert.Equal(t, a.Group, gen[0].Group)
	assert.Equal(t, a.Kind, gen[0].Kind)
	assert.Equal(t, b.Group, gen[1].Group)
	assert.Equal(t, b.Kind, gen[1].Kind)
	// Pointers must be distinct per element, not aliased to a shared loop var.
	require.NotNil(t, gen[0].Id)
	require.NotNil(t, gen[1].Id)
	assert.NotSame(t, gen[0].Id, gen[1].Id)
}

func TestToGenResourceList_Empty(t *testing.T) {
	gen := toGenResourceList(nil)
	assert.Empty(t, gen)
}
