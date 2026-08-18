package resources

import (
	"fmt"

	"github.com/ooqls/go-auth/internal/resourcesv1"
)

func ToTargetString(group, kind string) string {
	return fmt.Sprintf("resource:%s:%s", group, kind)
}

type Resourcev1 = resourcesv1.Resourcev1
