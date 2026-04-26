package corev1

import (
	"fmt"
)

func ToPermissionString(meta Metadata, action string) string {
	return fmt.Sprintf("%s:%s", ToTargetString(meta), action)
}

func ToTargetString(meta Metadata) string {
	return fmt.Sprintf("%s:%s", meta.Group, meta.Kind)
}
