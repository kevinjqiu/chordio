package chordio

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLocalNode_findPredecessor(t *testing.T) {
	m := Rank(3)

	n0, err := newLocalNode(0, "n0", m)
	n1 := NodeRef{id: 1, bind: "n1"}
	n3 := NodeRef{id: 3, bind: "n3"}
	assert.Nil(t, err)

	n0.neighbourhood.Add(&n1)
	n0.neighbourhood.Add(&n3)

	// Setup the finger table
	n0.ft.entries[0].node = n1.id
	n0.ft.entries[1].node = n3.id
	n0.ft.entries[2].node = n0.id

	n, err := n0.findPredecessor(context.Background(), 6)
	assert.Nil(t, err)
	assert.Equal(t, 0, n.GetID())
}
