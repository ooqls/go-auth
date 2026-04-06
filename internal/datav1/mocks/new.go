package mocks

import (
	"github.com/golang/mock/gomock"
	"github.com/ooqls/go-auth/internal/aggsv1"
	"github.com/ooqls/go-auth/internal/datav1"
	"github.com/ooqls/go-auth/internal/permissionbindingsv1"
	"github.com/ooqls/go-auth/internal/permissionsv1"
	"github.com/ooqls/go-auth/internal/resourcesv1"
	"github.com/ooqls/go-auth/internal/rolebindingsv1"
	"github.com/ooqls/go-auth/internal/rolesv1"
	"github.com/ooqls/go-auth/internal/usersv1"
)

type MockParams struct {
	RoleReader              rolesv1.Reader
	RoleWriter              rolesv1.Writer
	RoleBindingsReader      rolebindingsv1.Reader
	RoleBindingsWriter      rolebindingsv1.Writer
	UserReader              usersv1.Reader
	UserWriter              usersv1.Writer
	ResourceReader          resourcesv1.Reader
	ResourceWriter          resourcesv1.Writer
	PermissionReader        permissionsv1.Reader
	PermissionWriter        permissionsv1.Writer
	PermissionBindingReader permissionbindingsv1.Reader
	PermissionBindingWriter permissionbindingsv1.Writer
	AggReader               aggsv1.Reader
}

func NewMockFactoryFromParams(ctrl *gomock.Controller, p MockParams) datav1.Factory {
	f := NewMockFactory(ctrl)
	f.EXPECT().NewRoleReader().Return(p.RoleReader).AnyTimes()
	f.EXPECT().NewRoleWriter().Return(p.RoleWriter).AnyTimes()
	f.EXPECT().NewRoleBindingsReader().Return(p.RoleBindingsReader).AnyTimes()
	f.EXPECT().NewRoleBindingsWriter().Return(p.RoleBindingsWriter).AnyTimes()
	f.EXPECT().NewUserReader().Return(p.UserReader).AnyTimes()
	f.EXPECT().NewUserWriter().Return(p.UserWriter).AnyTimes()
	f.EXPECT().NewResourceReader().Return(p.ResourceReader).AnyTimes()
	f.EXPECT().NewResourceWriter().Return(p.ResourceWriter).AnyTimes()
	f.EXPECT().NewPermissionReader().Return(p.PermissionReader).AnyTimes()
	f.EXPECT().NewPermissionWriter().Return(p.PermissionWriter).AnyTimes()
	f.EXPECT().NewPermissionBindingReader().Return(p.PermissionBindingReader).AnyTimes()
	f.EXPECT().NewPermissionBindingWriter().Return(p.PermissionBindingWriter).AnyTimes()
	f.EXPECT().NewAggReader().Return(p.AggReader).AnyTimes()
	return f
}
