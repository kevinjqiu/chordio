package chordio

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNeighbourhood_Add(t *testing.T) {
	m := Rank(5)
	n := newNeighbourhood(m)
	n.Add(&NodeRef{id: 5})
	n.Add(&NodeRef{id: 15})
	n.Add(&NodeRef{id: 51})
	n.Add(&NodeRef{id: 43})
	n.Add(&NodeRef{id: 23})
	n.Add(&NodeRef{id: 123})
	n.Add(&NodeRef{id: 101})

	ids := []ChordID{}
	for _, nn := range n.nodes {
		ids = append(ids, nn.id)
	}
	assert.Equal(t, []ChordID{ChordID(5), ChordID(15), ChordID(23), ChordID(43), ChordID(51), ChordID(101), ChordID(123)}, ids)
}

func TestNeighbourhood_Get(t *testing.T) {
	t.Run("node not found", func(t *testing.T) {
		m := Rank(5)
		n := newNeighbourhood(m)
		n.Add(&NodeRef{id: 5})

		_, _, _, ok := n.Get(50)
		assert.False(t, ok)
	})

	t.Run("single node, pred and succ are itself", func(t *testing.T) {
		m := Rank(5)
		n := newNeighbourhood(m)
		n.Add(&NodeRef{id: 5})

		node, pred, succ, ok := n.Get(5)
		assert.True(t, ok)

		assert.Equal(t, ChordID(5), node.id)
		assert.Equal(t, ChordID(5), pred)
		assert.Equal(t, ChordID(5), succ)
	})

	t.Run("two nodes", func(t *testing.T) {
		m := Rank(5)
		n := newNeighbourhood(m)
		n.Add(&NodeRef{id: 5})
		n.Add(&NodeRef{id: 51})

		node, pred, succ, ok := n.Get(5)
		assert.True(t, ok)

		assert.Equal(t, ChordID(5), node.id)
		assert.Equal(t, ChordID(51), pred)
		assert.Equal(t, ChordID(51), succ)
	})

	t.Run("three nodes", func(t *testing.T) {
		m := Rank(5)
		n := newNeighbourhood(m)
		n.Add(&NodeRef{id: 5})
		n.Add(&NodeRef{id: 51})
		n.Add(&NodeRef{id: 23})

		node, pred, succ, ok := n.Get(5)
		assert.True(t, ok)

		assert.Equal(t, ChordID(5), node.id)
		assert.Equal(t, ChordID(51), pred)
		assert.Equal(t, ChordID(23), succ)
	})
}
