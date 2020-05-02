package chordio

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInterval_Has(t *testing.T) {
	t.Run("interval does not cross 0", func(t *testing.T) {
		int := NewInterval(7, 35, 73)
		for i := 35; i < 73; i++ {
			assert.True(t, int.Has(ChordID(i)))
		}

		for i := 73; i < 127; i++ {
			assert.False(t, int.Has(ChordID(i)))
		}
	})

	t.Run("interval does cross 0", func(t *testing.T) {
		int := NewInterval(7, 100, 5)
		for i := 100; i < 127; i++ {
			assert.True(t, int.Has(ChordID(i)))
		}

		for i := 0; i < 5; i++ {
			assert.True(t, int.Has(ChordID(i)))
		}

		for i := 5; i < 100; i++ {
			assert.False(t, int.Has(ChordID(i)))
		}
	})
}