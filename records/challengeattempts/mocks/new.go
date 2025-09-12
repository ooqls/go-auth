package mocks

import "github.com/golang/mock/gomock"

func NoopWriter(ctrl *gomock.Controller) *MockWriter {
	mock := NewMockWriter(ctrl)
	mock.EXPECT().CreateChallengeAttempt(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
	return mock
}