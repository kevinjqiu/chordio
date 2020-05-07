package chordio

import (
	"github.com/kevinjqiu/chordio/chord"
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestNewFingerTable(t *testing.T) {
	m := chord.Rank(5)
	node, _ := newLocalNode(15, "localhost:1234", m)
	assert.Equal(t, chord.ChordID(15), node.id)

	ft := newFingerTable(node, m)
	for i, e := range ft.entries {
		assert.Equal(t, node.id+chord.ChordID(chord.pow2(uint32(i))), e.start)
	}
	ft.Print(nil)
}
