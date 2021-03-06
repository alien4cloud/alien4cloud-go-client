// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/alien4cloud/alien4cloud-go-client/v3/alien4cloud (interfaces: EventService)

// Package a4cmocks is a generated GoMock package.
package a4cmocks

import (
	context "context"
	reflect "reflect"

	alien4cloud "github.com/alien4cloud/alien4cloud-go-client/v3/alien4cloud"
	gomock "github.com/golang/mock/gomock"
)

// MockEventService is a mock of EventService interface.
type MockEventService struct {
	ctrl     *gomock.Controller
	recorder *MockEventServiceMockRecorder
}

// MockEventServiceMockRecorder is the mock recorder for MockEventService.
type MockEventServiceMockRecorder struct {
	mock *MockEventService
}

// NewMockEventService creates a new mock instance.
func NewMockEventService(ctrl *gomock.Controller) *MockEventService {
	mock := &MockEventService{ctrl: ctrl}
	mock.recorder = &MockEventServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEventService) EXPECT() *MockEventServiceMockRecorder {
	return m.recorder
}

// GetEventsForApplicationEnvironment mocks base method.
func (m *MockEventService) GetEventsForApplicationEnvironment(arg0 context.Context, arg1 string, arg2, arg3 int) ([]alien4cloud.Event, int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEventsForApplicationEnvironment", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].([]alien4cloud.Event)
	ret1, _ := ret[1].(int)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetEventsForApplicationEnvironment indicates an expected call of GetEventsForApplicationEnvironment.
func (mr *MockEventServiceMockRecorder) GetEventsForApplicationEnvironment(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEventsForApplicationEnvironment", reflect.TypeOf((*MockEventService)(nil).GetEventsForApplicationEnvironment), arg0, arg1, arg2, arg3)
}
