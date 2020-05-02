package chordio

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNeighbourhood_Add(t *testing.T) {
	m := Rank(5)
	n := newNeighbourhood(m)
	n.Add(&Node{id: 5})
	n.Add(&Node{id: 15})
	n.Add(&Node{id: 51})
	n.Add(&Node{id: 43})
	n.Add(&Node{id: 23})
	n.Add(&Node{id: 123})
	n.Add(&Node{id: 101})

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
		n.Add(&Node{id: 5})

		_, ok := n.Get(50)
		assert.False(t, ok)
	})

	t.Run("single node, pred and succ are itself", func(t *testing.T) {
		m := Rank(5)
		n := newNeighbourhood(m)
		n.Add(&Node{id: 5})

		node, ok := n.Get(5)
		assert.True(t, ok)

		assert.Equal(t, ChordID(5), node.id)
		assert.Equal(t, ChordID(5), node.pred)
		assert.Equal(t, ChordID(5), node.succ)
	})

	t.Run("two nodes", func(t *testing.T) {
		m := Rank(5)
		n := newNeighbourhood(m)
		n.Add(&Node{id: 5})
		n.Add(&Node{id: 51})

		node, ok := n.Get(5)
		assert.True(t, ok)

		assert.Equal(t, ChordID(5), node.id)
		assert.Equal(t, ChordID(51), node.pred)
		assert.Equal(t, ChordID(51), node.succ)
	})

	t.Run("three nodes", func(t *testing.T) {
		m := Rank(5)
		n := newNeighbourhood(m)
		n.Add(&Node{id: 5})
		n.Add(&Node{id: 51})
		n.Add(&Node{id: 23})

		node, ok := n.Get(5)
		assert.True(t, ok)

		assert.Equal(t, ChordID(5), node.id)
		assert.Equal(t, ChordID(51), node.pred)
		assert.Equal(t, ChordID(23), node.succ)
	})
}
