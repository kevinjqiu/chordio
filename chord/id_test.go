package chord

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestChordID_Sub(t *testing.T) {
	m := Rank(3)
	t.Run("same as normal sub if doesnt cross boundary", func(t *testing.T) {
		a := ID(5)
		b := ID(2)
		c := a.Sub(b, m)
		assert.Equal(t, ID(3), c)
	})

	t.Run("hit boundary", func(t *testing.T) {
		a := ID(3)
		b := ID(3)
		c := a.Sub(b, m)
		assert.Equal(t, ID(0), c)
	})

	t.Run("over boundary", func(t *testing.T) {
		a := ID(3)
		b := ID(6)
		c := a.Sub(b, m)
		assert.Equal(t, ID(5), c)
	})

	t.Run("twice over boundary", func(t *testing.T) {
		a := ID(3)
		b := ID(14)
		c := a.Sub(b, m)
		assert.Equal(t, ID(5), c)
	})
}

func TestChordID_Add(t *testing.T) {
	m := Rank(3)
	t.Run("same as normal add if doesnt cross boundary", func(t *testing.T) {
		a := ID(3)
		b := ID(2)
		c := a.Add(b, m)
		assert.Equal(t, ID(5), c)
	})

	t.Run("hit boundary", func(t *testing.T) {
		a := ID(3)
		b := ID(5)
		c := a.Add(b, m)
		assert.Equal(t, ID(0), c)
	})

	t.Run("over boundary", func(t *testing.T) {
		a := ID(3)
		b := ID(8)
		c := a.Add(b, m)
		assert.Equal(t, ID(3), c)
	})
}

func TestChordID_Pow(t *testing.T) {
	assert.Equal(t, ID(8), ID(2).Pow(3))
}
