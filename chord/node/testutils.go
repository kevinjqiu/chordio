package node

import (
	"context"
	"github.com/kevinjqiu/chordio/chord"
	"github.com/stretchr/testify/mock"
)

type mockBuilder func(m *mock.Mock)

func newMockLocalNode(id chord.ID, bind string, m chord.Rank, _ ...nodeConstructorOption) (LocalNode, error) {
	return &localNode{
		id:   id,
		bind: bind,
		m:    m,
	}, nil
}

func newMockRemoteNode(ctx context.Context, bind string, _ ...nodeConstructorOption) (RemoteNode, error) {
	return &remoteNode{
		bind: bind,
	}, nil
}

func newMockNode(id chord.ID, bind string, builders ...mockBuilder) Node {
	m := mock.Mock{}
	m.On("GetID").Return(id)
	m.On("GetBind").Return(bind)
	for _, mb := range builders {
		mb(&m)
	}
	return &MockNode{
		Mock: m,
	}
}
