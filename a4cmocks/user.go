// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/alien4cloud/alien4cloud-go-client/v3/alien4cloud (interfaces: UserService)

// Package a4cmocks is a generated GoMock package.
package a4cmocks

import (
	context "context"
	reflect "reflect"

	alien4cloud "github.com/alien4cloud/alien4cloud-go-client/v3/alien4cloud"
	gomock "github.com/golang/mock/gomock"
)

// MockUserService is a mock of UserService interface.
type MockUserService struct {
	ctrl     *gomock.Controller
	recorder *MockUserServiceMockRecorder
}

// MockUserServiceMockRecorder is the mock recorder for MockUserService.
type MockUserServiceMockRecorder struct {
	mock *MockUserService
}

// NewMockUserService creates a new mock instance.
func NewMockUserService(ctrl *gomock.Controller) *MockUserService {
	mock := &MockUserService{ctrl: ctrl}
	mock.recorder = &MockUserServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserService) EXPECT() *MockUserServiceMockRecorder {
	return m.recorder
}

// AddRole mocks base method.
func (m *MockUserService) AddRole(arg0 context.Context, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddRole", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddRole indicates an expected call of AddRole.
func (mr *MockUserServiceMockRecorder) AddRole(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddRole", reflect.TypeOf((*MockUserService)(nil).AddRole), arg0, arg1, arg2)
}

// CreateGroup mocks base method.
func (m *MockUserService) CreateGroup(arg0 context.Context, arg1 alien4cloud.Group) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateGroup", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateGroup indicates an expected call of CreateGroup.
func (mr *MockUserServiceMockRecorder) CreateGroup(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateGroup", reflect.TypeOf((*MockUserService)(nil).CreateGroup), arg0, arg1)
}

// CreateUser mocks base method.
func (m *MockUserService) CreateUser(arg0 context.Context, arg1 alien4cloud.CreateUpdateUserRequest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUser", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateUser indicates an expected call of CreateUser.
func (mr *MockUserServiceMockRecorder) CreateUser(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUser", reflect.TypeOf((*MockUserService)(nil).CreateUser), arg0, arg1)
}

// DeleteGroup mocks base method.
func (m *MockUserService) DeleteGroup(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteGroup", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteGroup indicates an expected call of DeleteGroup.
func (mr *MockUserServiceMockRecorder) DeleteGroup(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteGroup", reflect.TypeOf((*MockUserService)(nil).DeleteGroup), arg0, arg1)
}

// DeleteUser mocks base method.
func (m *MockUserService) DeleteUser(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteUser", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteUser indicates an expected call of DeleteUser.
func (mr *MockUserServiceMockRecorder) DeleteUser(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteUser", reflect.TypeOf((*MockUserService)(nil).DeleteUser), arg0, arg1)
}

// GetGroup mocks base method.
func (m *MockUserService) GetGroup(arg0 context.Context, arg1 string) (alien4cloud.Group, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGroup", arg0, arg1)
	ret0, _ := ret[0].(alien4cloud.Group)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetGroup indicates an expected call of GetGroup.
func (mr *MockUserServiceMockRecorder) GetGroup(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGroup", reflect.TypeOf((*MockUserService)(nil).GetGroup), arg0, arg1)
}

// GetGroups mocks base method.
func (m *MockUserService) GetGroups(arg0 context.Context, arg1 []string) ([]alien4cloud.Group, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGroups", arg0, arg1)
	ret0, _ := ret[0].([]alien4cloud.Group)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetGroups indicates an expected call of GetGroups.
func (mr *MockUserServiceMockRecorder) GetGroups(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGroups", reflect.TypeOf((*MockUserService)(nil).GetGroups), arg0, arg1)
}

// GetUser mocks base method.
func (m *MockUserService) GetUser(arg0 context.Context, arg1 string) (alien4cloud.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUser", arg0, arg1)
	ret0, _ := ret[0].(alien4cloud.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUser indicates an expected call of GetUser.
func (mr *MockUserServiceMockRecorder) GetUser(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUser", reflect.TypeOf((*MockUserService)(nil).GetUser), arg0, arg1)
}

// GetUsers mocks base method.
func (m *MockUserService) GetUsers(arg0 context.Context, arg1 []string) ([]alien4cloud.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUsers", arg0, arg1)
	ret0, _ := ret[0].([]alien4cloud.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUsers indicates an expected call of GetUsers.
func (mr *MockUserServiceMockRecorder) GetUsers(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUsers", reflect.TypeOf((*MockUserService)(nil).GetUsers), arg0, arg1)
}

// RemoveRole mocks base method.
func (m *MockUserService) RemoveRole(arg0 context.Context, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveRole", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveRole indicates an expected call of RemoveRole.
func (mr *MockUserServiceMockRecorder) RemoveRole(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveRole", reflect.TypeOf((*MockUserService)(nil).RemoveRole), arg0, arg1, arg2)
}

// SearchGroups mocks base method.
func (m *MockUserService) SearchGroups(arg0 context.Context, arg1 alien4cloud.SearchRequest) ([]alien4cloud.Group, int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SearchGroups", arg0, arg1)
	ret0, _ := ret[0].([]alien4cloud.Group)
	ret1, _ := ret[1].(int)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// SearchGroups indicates an expected call of SearchGroups.
func (mr *MockUserServiceMockRecorder) SearchGroups(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SearchGroups", reflect.TypeOf((*MockUserService)(nil).SearchGroups), arg0, arg1)
}

// SearchUsers mocks base method.
func (m *MockUserService) SearchUsers(arg0 context.Context, arg1 alien4cloud.SearchRequest) ([]alien4cloud.User, int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SearchUsers", arg0, arg1)
	ret0, _ := ret[0].([]alien4cloud.User)
	ret1, _ := ret[1].(int)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// SearchUsers indicates an expected call of SearchUsers.
func (mr *MockUserServiceMockRecorder) SearchUsers(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SearchUsers", reflect.TypeOf((*MockUserService)(nil).SearchUsers), arg0, arg1)
}

// UpdateGroup mocks base method.
func (m *MockUserService) UpdateGroup(arg0 context.Context, arg1 string, arg2 alien4cloud.Group) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateGroup", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateGroup indicates an expected call of UpdateGroup.
func (mr *MockUserServiceMockRecorder) UpdateGroup(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateGroup", reflect.TypeOf((*MockUserService)(nil).UpdateGroup), arg0, arg1, arg2)
}

// UpdateUser mocks base method.
func (m *MockUserService) UpdateUser(arg0 context.Context, arg1 string, arg2 alien4cloud.CreateUpdateUserRequest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateUser", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateUser indicates an expected call of UpdateUser.
func (mr *MockUserServiceMockRecorder) UpdateUser(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateUser", reflect.TypeOf((*MockUserService)(nil).UpdateUser), arg0, arg1, arg2)
}
