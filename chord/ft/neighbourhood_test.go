package ft

import (
	"github.com/kevinjqiu/chordio/chord"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNeighbourhood_Add(t *testing.T) {
	m := chord.Rank(5)
	n := newNeighbourhood(m)
	n.Add(&chord.NodeRef{ID: 5})
	n.Add(&chord.NodeRef{ID: 15})
	n.Add(&chord.NodeRef{ID: 51})
	n.Add(&chord.NodeRef{ID: 43})
	n.Add(&chord.NodeRef{ID: 23})
	n.Add(&chord.NodeRef{ID: 123})
	n.Add(&chord.NodeRef{ID: 101})

	ids := []chord.ID{}
	for _, nn := range n.nodes {
		ids = append(ids, nn.ID)
	}
	assert.Equal(t, []chord.ID{chord.ID(5), chord.ID(15), chord.ID(23), chord.ID(43), chord.ID(51), chord.ID(101), chord.ID(123)}, ids)
}

func TestNeighbourhood_Get(t *testing.T) {
	t.Run("NodeID not found", func(t *testing.T) {
		m := chord.Rank(5)
		n := newNeighbourhood(m)
		n.Add(&chord.NodeRef{ID: 5})

		_, _, _, ok := n.Get(50)
		assert.False(t, ok)
	})

	t.Run("single NodeID, pred and succ are itself", func(t *testing.T) {
		m := chord.Rank(5)
		n := newNeighbourhood(m)
		n.Add(&chord.NodeRef{ID: 5})

		node, pred, succ, ok := n.Get(5)
		assert.True(t, ok)

		assert.Equal(t, chord.ID(5), node.ID)
		assert.Equal(t, chord.ID(5), pred)
		assert.Equal(t, chord.ID(5), succ)
	})

	t.Run("two nodes", func(t *testing.T) {
		m := chord.Rank(5)
		n := newNeighbourhood(m)
		n.Add(&chord.NodeRef{ID: 5})
		n.Add(&chord.NodeRef{ID: 51})

		node, pred, succ, ok := n.Get(5)
		assert.True(t, ok)

		assert.Equal(t, chord.ID(5), node.ID)
		assert.Equal(t, chord.ID(51), pred)
		assert.Equal(t, chord.ID(51), succ)
	})

	t.Run("three nodes", func(t *testing.T) {
		m := chord.Rank(5)
		n := newNeighbourhood(m)
		n.Add(&chord.NodeRef{ID: 5})
		n.Add(&chord.NodeRef{ID: 51})
		n.Add(&chord.NodeRef{ID: 23})

		node, pred, succ, ok := n.Get(5)
		assert.True(t, ok)

		assert.Equal(t, chord.ID(5), node.ID)
		assert.Equal(t, chord.ID(51), pred)
		assert.Equal(t, chord.ID(23), succ)
	})
}
