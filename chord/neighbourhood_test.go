package chord

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNeighbourhood_Add(t *testing.T) {
	m := Rank(5)
	n := newNeighbourhood(m)
	n.Add(&NodeRef{ID: 5})
	n.Add(&NodeRef{ID: 15})
	n.Add(&NodeRef{ID: 51})
	n.Add(&NodeRef{ID: 43})
	n.Add(&NodeRef{ID: 23})
	n.Add(&NodeRef{ID: 123})
	n.Add(&NodeRef{ID: 101})

	ids := []ID{}
	for _, nn := range n.nodes {
		ids = append(ids, nn.ID)
	}
	assert.Equal(t, []ID{ID(5), ID(15), ID(23), ID(43), ID(51), ID(101), ID(123)}, ids)
}

func TestNeighbourhood_Get(t *testing.T) {
	t.Run("NodeID not found", func(t *testing.T) {
		m := Rank(5)
		n := newNeighbourhood(m)
		n.Add(&NodeRef{ID: 5})

		_, _, _, ok := n.Get(50)
		assert.False(t, ok)
	})

	t.Run("single NodeID, pred and succ are itself", func(t *testing.T) {
		m := Rank(5)
		n := newNeighbourhood(m)
		n.Add(&NodeRef{ID: 5})

		node, pred, succ, ok := n.Get(5)
		assert.True(t, ok)

		assert.Equal(t, ID(5), node.ID)
		assert.Equal(t, ID(5), pred)
		assert.Equal(t, ID(5), succ)
	})

	t.Run("two nodes", func(t *testing.T) {
		m := Rank(5)
		n := newNeighbourhood(m)
		n.Add(&NodeRef{ID: 5})
		n.Add(&NodeRef{ID: 51})

		node, pred, succ, ok := n.Get(5)
		assert.True(t, ok)

		assert.Equal(t, ID(5), node.ID)
		assert.Equal(t, ID(51), pred)
		assert.Equal(t, ID(51), succ)
	})

	t.Run("three nodes", func(t *testing.T) {
		m := Rank(5)
		n := newNeighbourhood(m)
		n.Add(&NodeRef{ID: 5})
		n.Add(&NodeRef{ID: 51})
		n.Add(&NodeRef{ID: 23})

		node, pred, succ, ok := n.Get(5)
		assert.True(t, ok)

		assert.Equal(t, ID(5), node.ID)
		assert.Equal(t, ID(51), pred)
		assert.Equal(t, ID(23), succ)
	})
}
