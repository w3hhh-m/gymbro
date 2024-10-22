// Code generated by mockery v2.43.2. DO NOT EDIT.

package mocks

import (
	storage "GYMBRO/internal/storage"

	mock "github.com/stretchr/testify/mock"
)

// WorkoutRepository is an autogenerated mock type for the WorkoutRepository type
type WorkoutRepository struct {
	mock.Mock
}

// GetWorkout provides a mock function with given fields: _a0
func (_m *WorkoutRepository) GetWorkout(_a0 *string) (*storage.WorkoutWithRecords, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for GetWorkout")
	}

	var r0 *storage.WorkoutWithRecords
	var r1 error
	if rf, ok := ret.Get(0).(func(*string) (*storage.WorkoutWithRecords, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(*string) *storage.WorkoutWithRecords); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*storage.WorkoutWithRecords)
		}
	}

	if rf, ok := ret.Get(1).(func(*string) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SaveWorkout provides a mock function with given fields: _a0
func (_m *WorkoutRepository) SaveWorkout(_a0 *storage.WorkoutSession) error {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for SaveWorkout")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*storage.WorkoutSession) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewWorkoutRepository creates a new instance of WorkoutRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewWorkoutRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *WorkoutRepository {
	mock := &WorkoutRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
