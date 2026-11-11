package permissionbindings

import (
	"testing"

	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/stretchr/testify/assert"
)

func TestHasCoreGrant(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		grants []string
		want   bool
	}{
		{name: "no grants"},
		{name: "exact match", grants: []string{"core:auth:permission:assign"}, want: true},
		{name: "all core actions", grants: []string{"core:*"}, want: true},
		{name: "prefix wildcard", grants: []string{"*:auth:permission:assign"}, want: true},
		{name: "group wildcard", grants: []string{"core:*:permission:assign"}, want: true},
		{name: "kind wildcard", grants: []string{"core:auth:*:assign"}, want: true},
		{name: "action wildcard", grants: []string{"core:auth:permission:*"}, want: true},
		{name: "all fields wildcard", grants: []string{"*:*:*:*"}, want: true},
		{name: "find matching grant", grants: []string{"core:auth:role:update", "core:auth:permission:assign"}, want: true},
		{name: "wrong prefix", grants: []string{"resource:auth:permission:assign"}},
		{name: "wrong group", grants: []string{"core:other:permission:assign"}},
		{name: "wrong kind", grants: []string{"core:auth:role:assign"}},
		{name: "wrong action", grants: []string{"core:auth:permission:unassign"}},
		{name: "incomplete grant", grants: []string{"core:auth:permission"}},
		{name: "empty grant", grants: []string{""}},
		// Match the existing split_part-based semantics, including extra fields.
		{name: "trailing fields", grants: []string{"core:auth:permission:assign:extra"}, want: true},
		{name: "bare wildcard is not a core grant", grants: []string{"*"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, hasCoreGrant(tc.grants, corev1.PermissionsV1, "assign"))
		})
	}
}
