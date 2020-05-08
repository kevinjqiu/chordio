package node

import (
	"github.com/kevinjqiu/chordio/chord"
	"github.com/stretchr/testify/mock"
)

type mockBuilder func(m *mock.Mock)

func newMockFactory() (*mockFactory, *mock.Mock) {
	mf := mockFactory{}
	return &mf, &mf.Mock
}

func newMockNode(id chord.ID, bind string, builders ...mockBuilder) (Node, *mock.Mock) {
	mn := MockNode{}
	mn.Mock.On("GetID").Return(id)
	mn.Mock.On("GetBind").Return(bind)
	for _, mb := range builders {
		mb(&mn.Mock)
	}
	return &mn, &mn.Mock
}
