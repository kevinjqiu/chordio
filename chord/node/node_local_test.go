package node

import (
	"context"
	"fmt"
	"github.com/kevinjqiu/chordio/chord"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestLocalNode_ClosestPrecedingFinger(t *testing.T) {
	n2, _ := newMockNode(2, "n2")
	n5, _ := newMockNode(5, "n5")

	n, err := NewLocal(1, "n1", chord.Rank(3))

	assert.Nil(t, err)

	factory, m := newMockFactory()
	n.setNodeFactory(factory)
	m.On("newRemoteNode", mock.Anything, "n2").Return(n2, nil)
	m.On("newRemoteNode", mock.Anything, "n5").Return(n5, nil)

	n.GetFingerTable().SetEntry(0, n2)
	n.GetFingerTable().SetEntry(1, n2)
	n.GetFingerTable().SetEntry(2, n5)
	n.GetFingerTable().Print(nil)

	tcs := []struct {
		id           chord.ID
		expectedNode chord.ID
	}{
		{
			id:           0,
			expectedNode: 5,
		},
		{
			id:           1,
			expectedNode: 5,
		},
		{
			id:           2,
			expectedNode: 1,
		},
		{
			id:           3,
			expectedNode: 2,
		},
		{
			id:           4,
			expectedNode: 2,
		},
		{
			id:           5,
			expectedNode: 2,
		},
		{
			id:           6,
			expectedNode: 5,
		},
		{
			id:           7,
			expectedNode: 5,
		},
	}

	for _, tc := range tcs {
		t.Run(fmt.Sprintf("id=%d, node=%d", tc.id, tc.expectedNode), func(t *testing.T) {
			cpf, err := n.ClosestPrecedingFinger(context.Background(), tc.id)
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedNode, cpf.GetID())
		})
	}
}
