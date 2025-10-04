package mocks

import (
	"github.com/golang/mock/gomock"
	users "github.com/ooqls/go-auth/records/v1/users"
)

func ReturnUser(ctrl *gomock.Controller, user users.User) *MockReader {
	mock := NewMockReader(ctrl)
	mock.EXPECT().GetUser(gomock.Any(), gomock.Any()).AnyTimes().Return(&user, nil)
	return mock
}
