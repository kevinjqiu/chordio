// Code generated by mockery v1.0.0. DO NOT EDIT.

package node

import (
	context "context"

	chord "github.com/kevinjqiu/chordio/chord"

	mock "github.com/stretchr/testify/mock"

	pb "github.com/kevinjqiu/chordio/pb"
)

// MockLocalNode is an autogenerated mock type for the LocalNode type
type MockLocalNode struct {
	mock.Mock
}

// AsProtobufNode provides a mock function with given fields:
func (_m *MockLocalNode) AsProtobufNode() *pb.Node {
	ret := _m.Called()

	var r0 *pb.Node
	if rf, ok := ret.Get(0).(func() *pb.Node); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pb.Node)
		}
	}

	return r0
}

// ClosestPrecedingFinger provides a mock function with given fields: ctx, id
func (_m *MockLocalNode) ClosestPrecedingFinger(ctx context.Context, id chord.ID) (Node, error) {
	ret := _m.Called(ctx, id)

	var r0 Node
	if rf, ok := ret.Get(0).(func(context.Context, chord.ID) Node); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(Node)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, chord.ID) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindPredecessor provides a mock function with given fields: ctx, id
func (_m *MockLocalNode) FindPredecessor(ctx context.Context, id chord.ID) (Node, error) {
	ret := _m.Called(ctx, id)

	var r0 Node
	if rf, ok := ret.Get(0).(func(context.Context, chord.ID) Node); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(Node)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, chord.ID) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindSuccessor provides a mock function with given fields: ctx, id
func (_m *MockLocalNode) FindSuccessor(ctx context.Context, id chord.ID) (Node, error) {
	ret := _m.Called(ctx, id)

	var r0 Node
	if rf, ok := ret.Get(0).(func(context.Context, chord.ID) Node); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(Node)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, chord.ID) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetBind provides a mock function with given fields:
func (_m *MockLocalNode) GetBind() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetFingerTable provides a mock function with given fields:
func (_m *MockLocalNode) GetFingerTable() *FingerTable {
	ret := _m.Called()

	var r0 *FingerTable
	if rf, ok := ret.Get(0).(func() *FingerTable); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*FingerTable)
		}
	}

	return r0
}

// GetID provides a mock function with given fields:
func (_m *MockLocalNode) GetID() chord.ID {
	ret := _m.Called()

	var r0 chord.ID
	if rf, ok := ret.Get(0).(func() chord.ID); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(chord.ID)
	}

	return r0
}

// GetPredNode provides a mock function with given fields:
func (_m *MockLocalNode) GetPredNode() NodeRef {
	ret := _m.Called()

	var r0 NodeRef
	if rf, ok := ret.Get(0).(func() NodeRef); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(NodeRef)
		}
	}

	return r0
}

// GetRank provides a mock function with given fields:
func (_m *MockLocalNode) GetRank() chord.Rank {
	ret := _m.Called()

	var r0 chord.Rank
	if rf, ok := ret.Get(0).(func() chord.Rank); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(chord.Rank)
	}

	return r0
}

// GetSuccNode provides a mock function with given fields:
func (_m *MockLocalNode) GetSuccNode() NodeRef {
	ret := _m.Called()

	var r0 NodeRef
	if rf, ok := ret.Get(0).(func() NodeRef); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(NodeRef)
		}
	}

	return r0
}

// Join provides a mock function with given fields: ctx, introducerNode
func (_m *MockLocalNode) Join(ctx context.Context, introducerNode RemoteNode) error {
	ret := _m.Called(ctx, introducerNode)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, RemoteNode) error); ok {
		r0 = rf(ctx, introducerNode)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetPredNode provides a mock function with given fields: ctx, n
func (_m *MockLocalNode) SetPredNode(ctx context.Context, n NodeRef) {
	_m.Called(ctx, n)
}

// SetSuccNode provides a mock function with given fields: ctx, n
func (_m *MockLocalNode) SetSuccNode(ctx context.Context, n NodeRef) {
	_m.Called(ctx, n)
}

// String provides a mock function with given fields:
func (_m *MockLocalNode) String() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// UpdateFingerTableEntry provides a mock function with given fields: ctx, s, i
func (_m *MockLocalNode) UpdateFingerTableEntry(ctx context.Context, s Node, i int) error {
	ret := _m.Called(ctx, s, i)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, Node, int) error); ok {
		r0 = rf(ctx, s, i)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// setNodeFactory provides a mock function with given fields: f
func (_m *MockLocalNode) setNodeFactory(f factory) {
	_m.Called(f)
}
