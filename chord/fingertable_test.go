package chord

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewFingerTable(t *testing.T) {
	m := Rank(5)
	node, _ := NewLocal(15, "localhost:1234", m)
	assert.Equal(t, ID(15), node.GetID())

	ft := newFingerTable(node, m)
	assert.Equal(t, 5, len(ft.entries))
	assert.Equal(t, ID(16), ft.entries[0].Start)
	assert.Equal(t, ID(17), ft.entries[1].Start)
	assert.Equal(t, ID(19), ft.entries[2].Start)
	assert.Equal(t, ID(23), ft.entries[3].Start)
	assert.Equal(t, ID(31), ft.entries[4].Start)
	ft.Print(nil)
}
