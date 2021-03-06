// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/alien4cloud/alien4cloud-go-client/v3/alien4cloud (interfaces: CatalogService)

// Package a4cmocks is a generated GoMock package.
package a4cmocks

import (
	context "context"
	io "io"
	reflect "reflect"

	alien4cloud "github.com/alien4cloud/alien4cloud-go-client/v3/alien4cloud"
	gomock "github.com/golang/mock/gomock"
)

// MockCatalogService is a mock of CatalogService interface.
type MockCatalogService struct {
	ctrl     *gomock.Controller
	recorder *MockCatalogServiceMockRecorder
}

// MockCatalogServiceMockRecorder is the mock recorder for MockCatalogService.
type MockCatalogServiceMockRecorder struct {
	mock *MockCatalogService
}

// NewMockCatalogService creates a new mock instance.
func NewMockCatalogService(ctrl *gomock.Controller) *MockCatalogService {
	mock := &MockCatalogService{ctrl: ctrl}
	mock.recorder = &MockCatalogServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCatalogService) EXPECT() *MockCatalogServiceMockRecorder {
	return m.recorder
}

// UploadCSAR mocks base method.
func (m *MockCatalogService) UploadCSAR(arg0 context.Context, arg1 io.Reader, arg2 string) (alien4cloud.CSAR, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UploadCSAR", arg0, arg1, arg2)
	ret0, _ := ret[0].(alien4cloud.CSAR)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UploadCSAR indicates an expected call of UploadCSAR.
func (mr *MockCatalogServiceMockRecorder) UploadCSAR(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UploadCSAR", reflect.TypeOf((*MockCatalogService)(nil).UploadCSAR), arg0, arg1, arg2)
}
