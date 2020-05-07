package chord

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestChordID_Sub(t *testing.T) {
	m := Rank(3)
	t.Run("same as normal sub if doesnt cross boundary", func(t *testing.T) {
		a := ChordID(5)
		b := ChordID(2)
		c := a.Sub(b, m)
		assert.Equal(t, ChordID(3), c)
	})

	t.Run("hit boundary", func(t *testing.T) {
		a := ChordID(3)
		b := ChordID(3)
		c := a.Sub(b, m)
		assert.Equal(t, ChordID(0), c)
	})

	t.Run("over boundary", func(t *testing.T) {
		a := ChordID(3)
		b := ChordID(6)
		c := a.Sub(b, m)
		assert.Equal(t, ChordID(5), c)
	})

	t.Run("twice over boundary", func(t *testing.T) {
		a := ChordID(3)
		b := ChordID(14)
		c := a.Sub(b, m)
		assert.Equal(t, ChordID(5), c)
	})
}

func TestChordID_Add(t *testing.T) {
	m := Rank(3)
	t.Run("same as normal add if doesnt cross boundary", func(t *testing.T) {
		a := ChordID(3)
		b := ChordID(2)
		c := a.Add(b, m)
		assert.Equal(t, ChordID(5), c)
	})

	t.Run("hit boundary", func(t *testing.T) {
		a := ChordID(3)
		b := ChordID(5)
		c := a.Add(b, m)
		assert.Equal(t, ChordID(0), c)
	})

	t.Run("over boundary", func(t *testing.T) {
		a := ChordID(3)
		b := ChordID(8)
		c := a.Add(b, m)
		assert.Equal(t, ChordID(3), c)
	})
}
