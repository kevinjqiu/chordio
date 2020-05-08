package node

import (
	"github.com/kevinjqiu/chordio/chord"
	"github.com/stretchr/testify/mock"
)

type mockBuilder func(m *mock.Mock)

func newMockNode(id chord.ID, bind string, builders ...mockBuilder) Node {
	m := mock.Mock{}
	m.On("GetID").Return(id)
	m.On("GetBind").Return(bind)
	for _, mb := range builders {
		mb(&m)
	}
	return nil
	//return &MockNode{
	//	Mock: m,
	//}
}
