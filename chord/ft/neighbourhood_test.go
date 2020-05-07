package ft

import (
	"github.com/kevinjqiu/chordio/chord"
	"github.com/kevinjqiu/chordio/chord/node"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNeighbourhood_Add(t *testing.T) {
	m := chord.Rank(5)
	n := newNeighbourhood(m)
	n.Add(&node.NodeRef{ID: 5})
	n.Add(&node.NodeRef{ID: 15})
	n.Add(&node.NodeRef{ID: 51})
	n.Add(&node.NodeRef{ID: 43})
	n.Add(&node.NodeRef{ID: 23})
	n.Add(&node.NodeRef{ID: 123})
	n.Add(&node.NodeRef{ID: 101})

	ids := []chord.ChordID{}
	for _, nn := range n.nodes {
		ids = append(ids, nn.ID)
	}
	assert.Equal(t, []chord.ChordID{chord.ChordID(5), chord.ChordID(15), chord.ChordID(23), chord.ChordID(43), chord.ChordID(51), chord.ChordID(101), chord.ChordID(123)}, ids)
}

func TestNeighbourhood_Get(t *testing.T) {
	t.Run("NodeID not found", func(t *testing.T) {
		m := chord.Rank(5)
		n := newNeighbourhood(m)
		n.Add(&node.NodeRef{ID: 5})

		_, _, _, ok := n.Get(50)
		assert.False(t, ok)
	})

	t.Run("single NodeID, pred and succ are itself", func(t *testing.T) {
		m := chord.Rank(5)
		n := newNeighbourhood(m)
		n.Add(&node.NodeRef{ID: 5})

		node, pred, succ, ok := n.Get(5)
		assert.True(t, ok)

		assert.Equal(t, chord.ChordID(5), node.ID)
		assert.Equal(t, chord.ChordID(5), pred)
		assert.Equal(t, chord.ChordID(5), succ)
	})

	t.Run("two nodes", func(t *testing.T) {
		m := chord.Rank(5)
		n := newNeighbourhood(m)
		n.Add(&node.NodeRef{ID: 5})
		n.Add(&node.NodeRef{ID: 51})

		node, pred, succ, ok := n.Get(5)
		assert.True(t, ok)

		assert.Equal(t, chord.ChordID(5), node.ID)
		assert.Equal(t, chord.ChordID(51), pred)
		assert.Equal(t, chord.ChordID(51), succ)
	})

	t.Run("three nodes", func(t *testing.T) {
		m := chord.Rank(5)
		n := newNeighbourhood(m)
		n.Add(&node.NodeRef{ID: 5})
		n.Add(&node.NodeRef{ID: 51})
		n.Add(&node.NodeRef{ID: 23})

		node, pred, succ, ok := n.Get(5)
		assert.True(t, ok)

		assert.Equal(t, chord.ChordID(5), node.ID)
		assert.Equal(t, chord.ChordID(51), pred)
		assert.Equal(t, chord.ChordID(23), succ)
	})
}
