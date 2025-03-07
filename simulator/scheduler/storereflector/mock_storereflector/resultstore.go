// Code generated by MockGen. DO NOT EDIT.
// Source: sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/storereflector (interfaces: ResultStore)

// Package mock_storereflector is a generated GoMock package.
package mock_storereflector

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1 "k8s.io/api/core/v1"
)

// MockResultStore is a mock of ResultStore interface.
type MockResultStore struct {
	ctrl     *gomock.Controller
	recorder *MockResultStoreMockRecorder
}

// MockResultStoreMockRecorder is the mock recorder for MockResultStore.
type MockResultStoreMockRecorder struct {
	mock *MockResultStore
}

// NewMockResultStore creates a new mock instance.
func NewMockResultStore(ctrl *gomock.Controller) *MockResultStore {
	mock := &MockResultStore{ctrl: ctrl}
	mock.recorder = &MockResultStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockResultStore) EXPECT() *MockResultStoreMockRecorder {
	return m.recorder
}

// AddStoredResultToPod mocks base method.
func (m *MockResultStore) AddStoredResultToPod(arg0 *v1.Pod) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddStoredResultToPod", arg0)
}

// AddStoredResultToPod indicates an expected call of AddStoredResultToPod.
func (mr *MockResultStoreMockRecorder) AddStoredResultToPod(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddStoredResultToPod", reflect.TypeOf((*MockResultStore)(nil).AddStoredResultToPod), arg0)
}

// DeleteData mocks base method.
func (m *MockResultStore) DeleteData(arg0 v1.Pod) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "DeleteData", arg0)
}

// DeleteData indicates an expected call of DeleteData.
func (mr *MockResultStoreMockRecorder) DeleteData(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteData", reflect.TypeOf((*MockResultStore)(nil).DeleteData), arg0)
}
