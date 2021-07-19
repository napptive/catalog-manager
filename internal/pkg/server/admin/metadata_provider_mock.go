// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/napptive/catalog-manager/internal/pkg/provider/metadata (interfaces: MetadataProvider)

// Package admin is a generated GoMock package.
package admin

import (
	gomock "github.com/golang/mock/gomock"
	entities "github.com/napptive/catalog-manager/internal/pkg/entities"
	reflect "reflect"
)

// MockMetadataProvider is a mock of MetadataProvider interface
type MockMetadataProvider struct {
	ctrl     *gomock.Controller
	recorder *MockMetadataProviderMockRecorder
}

// MockMetadataProviderMockRecorder is the mock recorder for MockMetadataProvider
type MockMetadataProviderMockRecorder struct {
	mock *MockMetadataProvider
}

// NewMockMetadataProvider creates a new mock instance
func NewMockMetadataProvider(ctrl *gomock.Controller) *MockMetadataProvider {
	mock := &MockMetadataProvider{ctrl: ctrl}
	mock.recorder = &MockMetadataProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockMetadataProvider) EXPECT() *MockMetadataProviderMockRecorder {
	return m.recorder
}

// Add mocks base method
func (m *MockMetadataProvider) Add(arg0 *entities.ApplicationInfo) (*entities.ApplicationInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add", arg0)
	ret0, _ := ret[0].(*entities.ApplicationInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Add indicates an expected call of Add
func (mr *MockMetadataProviderMockRecorder) Add(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockMetadataProvider)(nil).Add), arg0)
}

// Exists mocks base method
func (m *MockMetadataProvider) Exists(arg0 *entities.ApplicationID) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Exists", arg0)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Exists indicates an expected call of Exists
func (mr *MockMetadataProviderMockRecorder) Exists(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exists", reflect.TypeOf((*MockMetadataProvider)(nil).Exists), arg0)
}

// Get mocks base method
func (m *MockMetadataProvider) Get(arg0 entities.ApplicationID) (*entities.ApplicationInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0)
	ret0, _ := ret[0].(*entities.ApplicationInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get
func (mr *MockMetadataProviderMockRecorder) Get(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockMetadataProvider)(nil).Get), arg0)
}

// GetSummary mocks base method
func (m *MockMetadataProvider) GetSummary() (*entities.Summary, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSummary")
	ret0, _ := ret[0].(*entities.Summary)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSummary indicates an expected call of GetSummary
func (mr *MockMetadataProviderMockRecorder) GetSummary() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSummary", reflect.TypeOf((*MockMetadataProvider)(nil).GetSummary))
}

// List mocks base method
func (m *MockMetadataProvider) List(arg0 string) ([]*entities.ApplicationInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", arg0)
	ret0, _ := ret[0].([]*entities.ApplicationInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List
func (mr *MockMetadataProviderMockRecorder) List(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockMetadataProvider)(nil).List), arg0)
}

// ListSummary mocks base method
func (m *MockMetadataProvider) ListSummary(arg0 string) ([]*entities.AppSummary, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListSummary", arg0)
	ret0, _ := ret[0].([]*entities.AppSummary)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListSummary indicates an expected call of ListSummary
func (mr *MockMetadataProviderMockRecorder) ListSummary(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListSummary", reflect.TypeOf((*MockMetadataProvider)(nil).ListSummary), arg0)
}

// Remove mocks base method
func (m *MockMetadataProvider) Remove(arg0 *entities.ApplicationID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Remove", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Remove indicates an expected call of Remove
func (mr *MockMetadataProviderMockRecorder) Remove(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Remove", reflect.TypeOf((*MockMetadataProvider)(nil).Remove), arg0)
}