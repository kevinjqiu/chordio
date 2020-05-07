package chord

import (
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestNewFingerTable(t *testing.T) {
	m := Rank(5)
	node, _ := NewLocal(15, "localhost:1234", m)
	assert.Equal(t, ID(15), node.GetID())

	ft := newFingerTable(node, m)
	for i, e := range ft.entries {
		assert.Equal(t, node.GetID().Add(ID(2).Pow(i), m), e.Start)
	}
	ft.Print(nil)
}
