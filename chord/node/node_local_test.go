package node

import (
	"context"
	"fmt"
	"github.com/kevinjqiu/chordio/chord"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func setupNetwork(t *testing.T) chord.LocalNode {
	n0, n0mock := newMockNode(0, "n0")
	n1, n1mock := newMockNode(1, "n1")
	n3, n3mock := newMockNode(3, "n3")

	n, err := NewLocal(0, "n0", chord.Rank(3))
	factory, m := newMockFactory()
	n.setNodeFactory(factory)
	n.SetPredNode(context.Background(), n3)
	n.SetSuccNode(context.Background(), n1)

	m.On("newRemoteNode", mock.Anything, "n0").Return(n0, nil)
	m.On("newRemoteNode", mock.Anything, "n1").Return(n1, nil)
	m.On("newRemoteNode", mock.Anything, "n3").Return(n3, nil)

	n0mock.On("GetPredNode").Return(n3)
	n0mock.On("GetSuccNode").Return(n1)

	n1mock.On("GetPredNode").Return(n0)
	n1mock.On("GetSuccNode").Return(n3)

	n3mock.On("GetPredNode").Return(n1)
	n3mock.On("GetSuccNode").Return(n0)

	n.GetFingerTable().SetNodeAtEntry(0, n1)
	n.GetFingerTable().SetNodeAtEntry(1, n3)
	n.GetFingerTable().SetNodeAtEntry(2, n0)

	assert.Nil(t, err)

	return n
}

func setupRank3Network(t *testing.T) (n chord.LocalNode, n2, n5 chord.Node, n2mock, n5mock *mock.Mock){
	n2, n2mock = newMockNode(2, "n2")
	n5, n5mock = newMockNode(5, "n5")

	n, err := NewLocal(1, "n1", chord.Rank(3))

	assert.Nil(t, err)

	factory, m := newMockFactory()
	n.setNodeFactory(factory)
	m.On("newRemoteNode", mock.Anything, "n2").Return(n2, nil)
	m.On("newRemoteNode", mock.Anything, "n5").Return(n5, nil)

	n.GetFingerTable().SetNodeAtEntry(0, n2)
	n.GetFingerTable().SetNodeAtEntry(1, n2)
	n.GetFingerTable().SetNodeAtEntry(2, n5)
	//n.GetFingerTable().PrettyPrint(nil)
	return
}

func TestLocalNode_ClosestPrecedingFinger(t *testing.T) {
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
			n, _, _, _, _ := setupRank3Network(t)
			cpf, err := n.ClosestPrecedingFinger(context.Background(), tc.id)
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedNode, cpf.GetID())
		})
	}
}

func TestLocalNode_FindPredecessor(t *testing.T) {
	n := setupNetwork(t)
	var results = make([]chord.ID, 0, 8)
	for i := 0; i < 8; i++ {
		pred, err := n.FindPredecessor(context.Background(), chord.ID(i))
		assert.Nil(t, err)
		results = append(results, pred.GetID())
	}
	expected := []chord.ID{3, 0, 1, 1, 3, 3, 3, 3}
	assert.Equal(t, expected, results)
}

func TestLocalNode_FindSuccessor(t *testing.T) {
	n := setupNetwork(t)
	var results = make([]chord.ID, 0, 8)
	for i := 0; i < 8; i++ {
		succ, err := n.FindSuccessor(context.Background(), chord.ID(i))
		assert.Nil(t, err)
		results = append(results, succ.GetID())
	}
	fmt.Println(results)
	expected := []chord.ID{0, 1, 3, 3, 0, 0, 0, 0}
	assert.Equal(t, expected, results)
}
