package mocks

import (
	"github.com/golang/mock/gomock"
	loginchallenges "github.com/ooqls/go-auth/records/v1/login_challenges"
)

func ReturnChallenge(ctrl *gomock.Controller, challenge loginchallenges.Challenge) *MockReader {
	mock := NewMockReader(ctrl)
	mock.EXPECT().GetChallenge(gomock.Any(), gomock.Any()).Return(&challenge, nil)
	return mock
}
