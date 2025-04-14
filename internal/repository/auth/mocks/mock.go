package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	pgx "github.com/jackc/pgx/v5"
)

// Mockpool is a mock of pool interface.
type Mockpool struct {
	ctrl     *gomock.Controller
	recorder *MockpoolMockRecorder
}

// MockpoolMockRecorder is the mock recorder for Mockpool.
type MockpoolMockRecorder struct {
	mock *Mockpool
}

// NewMockpool creates a new mock instance.
func NewMockpool(ctrl *gomock.Controller) *Mockpool {
	mock := &Mockpool{ctrl: ctrl}
	mock.recorder = &MockpoolMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Mockpool) EXPECT() *MockpoolMockRecorder {
	return m.recorder
}

// Begin mocks base method.
func (m *Mockpool) Begin(ctx context.Context) (pgx.Tx, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Begin", ctx)
	ret0, _ := ret[0].(pgx.Tx)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Begin indicates an expected call of Begin.
func (mr *MockpoolMockRecorder) Begin(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Begin", reflect.TypeOf((*Mockpool)(nil).Begin), ctx)
}

// QueryRow mocks base method.
func (m *Mockpool) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, sql}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "QueryRow", varargs...)
	ret0, _ := ret[0].(pgx.Row)
	return ret0
}

// QueryRow indicates an expected call of QueryRow.
func (mr *MockpoolMockRecorder) QueryRow(ctx, sql interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, sql}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryRow", reflect.TypeOf((*Mockpool)(nil).QueryRow), varargs...)
}
