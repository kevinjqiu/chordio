package ft

import (
	"github.com/kevinjqiu/chordio/chord"
	node2 "github.com/kevinjqiu/chordio/chord/node"
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestNewFingerTable(t *testing.T) {
	m := chord.Rank(5)
	node, _ := node2.NewLocal(15, "localhost:1234", m)
	assert.Equal(t, chord.ChordID(15), node.GetID())

	ft := New(node, m)
	for i, e := range ft.entries {
		assert.Equal(t, node.GetID().Add(chord.ChordID(2).Pow(i), m), e.Start)
	}
	ft.Print(nil)
}
