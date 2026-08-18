package resourcesv1

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/stretchr/testify/assert"
)

func TestMarshalResource(t *testing.T) {
	res := Resourcev1{
		Metadata: corev1.Metadata{
			Group: "group",
			Kind:  "kind",
		},
		Id:   uuid.New(),
		Name: "name",
	}

	b, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}

	var retr Resourcev1
	if err := json.Unmarshal(b, &retr); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, retr.Group, res.Group)
	assert.NotEmpty(t, retr.Group)

}
