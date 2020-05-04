package chordio

import (
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestNewFingerTable(t *testing.T) {
	m := Rank(5)
	node, _ := newLocalNode("localhost:1234", m)
	assert.Equal(t, ChordID(15), node.id)

	ft := newFingerTable(node, m)
	for i, e := range ft.entries {
		assert.Equal(t, node.id + ChordID(pow2(uint32(i))), e.start)
	}
	ft.Print(nil)
}
