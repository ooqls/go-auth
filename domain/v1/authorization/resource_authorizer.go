package authorization

import (
	"go.uber.org/zap"
)

type ResourceAuthorizer struct {
	l *zap.Logger
}

func NewResourceAuthorizer() *ResourceAuthorizer {
	return &ResourceAuthorizer{}
}

func (ra *ResourceAuthorizer) IsAuthorizedToPerformAction(ctx *Context, action Action, resource Resource) error {
	resourceCheckpoints := []*ResourceCheckpoint{
		UserHasResourcePermission(),
	}

	for _, checkpoint := range resourceCheckpoints {
		if !checkpoint.IsAuthorized(ctx, action, resource) {
			ra.l.Error("user is not authorized to perform action", zap.String("action", string(action)), zap.String("resource", resource.ResourceName), zap.String("checkpoint", checkpoint.GetName()))
			return ErrPermissionDenied
		}
	}

	return nil
}
