// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/alien4cloud/alien4cloud-go-client/v3/alien4cloud (interfaces: OrchestratorService)

// Package a4cmocks is a generated GoMock package.
package a4cmocks

import (
	context "context"
	reflect "reflect"

	alien4cloud "github.com/alien4cloud/alien4cloud-go-client/v3/alien4cloud"
	gomock "github.com/golang/mock/gomock"
)

// MockOrchestratorService is a mock of OrchestratorService interface.
type MockOrchestratorService struct {
	ctrl     *gomock.Controller
	recorder *MockOrchestratorServiceMockRecorder
}

// MockOrchestratorServiceMockRecorder is the mock recorder for MockOrchestratorService.
type MockOrchestratorServiceMockRecorder struct {
	mock *MockOrchestratorService
}

// NewMockOrchestratorService creates a new mock instance.
func NewMockOrchestratorService(ctrl *gomock.Controller) *MockOrchestratorService {
	mock := &MockOrchestratorService{ctrl: ctrl}
	mock.recorder = &MockOrchestratorServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockOrchestratorService) EXPECT() *MockOrchestratorServiceMockRecorder {
	return m.recorder
}

// GetOrchestratorIDbyName mocks base method.
func (m *MockOrchestratorService) GetOrchestratorIDbyName(arg0 context.Context, arg1 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrchestratorIDbyName", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrchestratorIDbyName indicates an expected call of GetOrchestratorIDbyName.
func (mr *MockOrchestratorServiceMockRecorder) GetOrchestratorIDbyName(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrchestratorIDbyName", reflect.TypeOf((*MockOrchestratorService)(nil).GetOrchestratorIDbyName), arg0, arg1)
}

// GetOrchestratorLocations mocks base method.
func (m *MockOrchestratorService) GetOrchestratorLocations(arg0 context.Context, arg1 string) ([]alien4cloud.Location, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrchestratorLocations", arg0, arg1)
	ret0, _ := ret[0].([]alien4cloud.Location)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrchestratorLocations indicates an expected call of GetOrchestratorLocations.
func (mr *MockOrchestratorServiceMockRecorder) GetOrchestratorLocations(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrchestratorLocations", reflect.TypeOf((*MockOrchestratorService)(nil).GetOrchestratorLocations), arg0, arg1)
}