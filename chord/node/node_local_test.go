package node

import (
	"context"
	"fmt"
	"github.com/kevinjqiu/chordio/chord"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLocalNode_ClosestPrecedingFinger(t *testing.T) {
	n, err := NewLocal(1, "n1", chord.Rank(3),
		withLocalNodeConstructor(newMockLocalNode),
		withRemoteNodeConstructor(func(ctx context.Context, bind string, opts ...nodeConstructorOption) (node RemoteNode, err error) {
			var id chord.ID
			switch bind {
			case "n2":
				id = 2
			case "n5":
				id = 5
			}
			return &remoteNode{bind: bind, id: id}, nil
		}),
	)
	assert.Nil(t, err)

	n2 := newMockNode(2, "n2")
	n5 := newMockNode(5, "n5")

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
			expectedNode: 3,
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
			expectedNode: 5,
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
