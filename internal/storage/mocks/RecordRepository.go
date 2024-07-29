// Code generated by mockery v2.43.2. DO NOT EDIT.

package mocks

import (
	storage "GYMBRO/internal/storage"

	mock "github.com/stretchr/testify/mock"
)

// RecordRepository is an autogenerated mock type for the RecordRepository type
type RecordRepository struct {
	mock.Mock
}

// DeleteRecord provides a mock function with given fields: id
func (_m *RecordRepository) DeleteRecord(id int) error {
	ret := _m.Called(id)

	if len(ret) == 0 {
		panic("no return value specified for DeleteRecord")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(int) error); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetRecord provides a mock function with given fields: id
func (_m *RecordRepository) GetRecord(id int) (storage.Record, error) {
	ret := _m.Called(id)

	if len(ret) == 0 {
		panic("no return value specified for GetRecord")
	}

	var r0 storage.Record
	var r1 error
	if rf, ok := ret.Get(0).(func(int) (storage.Record, error)); ok {
		return rf(id)
	}
	if rf, ok := ret.Get(0).(func(int) storage.Record); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Get(0).(storage.Record)
	}

	if rf, ok := ret.Get(1).(func(int) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SaveRecord provides a mock function with given fields: ex
func (_m *RecordRepository) SaveRecord(ex storage.Record) (int, error) {
	ret := _m.Called(ex)

	if len(ret) == 0 {
		panic("no return value specified for SaveRecord")
	}

	var r0 int
	var r1 error
	if rf, ok := ret.Get(0).(func(storage.Record) (int, error)); ok {
		return rf(ex)
	}
	if rf, ok := ret.Get(0).(func(storage.Record) int); ok {
		r0 = rf(ex)
	} else {
		r0 = ret.Get(0).(int)
	}

	if rf, ok := ret.Get(1).(func(storage.Record) error); ok {
		r1 = rf(ex)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewRecordRepository creates a new instance of RecordRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewRecordRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *RecordRepository {
	mock := &RecordRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
