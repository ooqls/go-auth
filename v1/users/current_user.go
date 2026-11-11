package users

import (
	"errors"

	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/corev1"
	v1 "github.com/ooqls/go-auth/v1"
)

// CurrentUserCapabilities are fresh UI hints, not authority for later requests.
// ResourcesRead applies only to the unfiltered resource list (group=*; kind=*).
type CurrentUserCapabilities struct {
	UsersRead         bool
	RolesRead         bool
	PermissionsRead   bool
	PermissionsCreate bool
	ResourcesRead     bool
}

type CurrentUser struct {
	User         User
	Capabilities CurrentUserCapabilities
}

// GetCurrentUser permits authenticated self-read without granting access to other users.
func (u *UserServiceImpl) GetCurrentUser(ctx *authorizationv1.Context) (*CurrentUser, error) {
	if ctx == nil || !ctx.IsAuthenticated() || ctx.GetAuthedUser().Id == uuid.Nil {
		return nil, v1.ErrPermissionDenied(errors.New("authenticated user required"), nil)
	}

	user, err := u.userR.GetUser(ctx, ctx.GetAuthedUser().Id)
	if err != nil {
		return nil, v1.ErrInternal(err, nil)
	}
	if user == nil {
		return nil, v1.ErrNotFound(errors.New("current user not found"), nil)
	}

	capabilities := CurrentUserCapabilities{}
	for _, check := range []struct {
		metadata corev1.Metadata
		action   authorizationv1.Action
		allowed  *bool
	}{
		{corev1.UsersV1, authorizationv1.ReadAction, &capabilities.UsersRead},
		{corev1.RolesV1, authorizationv1.ReadAction, &capabilities.RolesRead},
		{corev1.PermissionsV1, authorizationv1.ReadAction, &capabilities.PermissionsRead},
		{corev1.PermissionsV1, authorizationv1.CreateAction, &capabilities.PermissionsCreate},
	} {
		err := u.ua.IsAuthorizedToPerformCoreAction(ctx, check.action, check.metadata.Group, check.metadata.Kind)
		if errors.Is(err, authorizationv1.ErrPermissionDenied) {
			continue
		}
		if err != nil {
			return nil, v1.ErrInternal(err, nil)
		}
		*check.allowed = true
	}

	err = u.ua.IsAuthorizedToPerformResourceAction(ctx, authorizationv1.ReadAction, "*", "*")
	if err != nil && !errors.Is(err, authorizationv1.ErrPermissionDenied) {
		return nil, v1.ErrInternal(err, nil)
	}
	capabilities.ResourcesRead = err == nil

	return &CurrentUser{User: *user, Capabilities: capabilities}, nil
}
