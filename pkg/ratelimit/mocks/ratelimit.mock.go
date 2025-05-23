// Code generated by MockGen. DO NOT EDIT.
// Source: ./pkg/ratelimit/types.go
//
// Generated by this command:
//
//	mockgen -source=./pkg/ratelimit/types.go -package=limitmocks -destination=./pkg/ratelimit/mocks/ratelimit.mock.go
//

// Package limitmocks is a generated GoMock package.
package limitmocks

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockLimiter is a mock of Limiter interface.
type MockLimiter struct {
	ctrl     *gomock.Controller
	recorder *MockLimiterMockRecorder
	isgomock struct{}
}

// MockLimiterMockRecorder is the mock recorder for MockLimiter.
type MockLimiterMockRecorder struct {
	mock *MockLimiter
}

// NewMockLimiter creates a new mock instance.
func NewMockLimiter(ctrl *gomock.Controller) *MockLimiter {
	mock := &MockLimiter{ctrl: ctrl}
	mock.recorder = &MockLimiterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLimiter) EXPECT() *MockLimiterMockRecorder {
	return m.recorder
}

// Limited mocks base method.
func (m *MockLimiter) Limited(ctx context.Context, key string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Limited", ctx, key)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Limited indicates an expected call of Limited.
func (mr *MockLimiterMockRecorder) Limited(ctx, key any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Limited", reflect.TypeOf((*MockLimiter)(nil).Limited), ctx, key)
}
