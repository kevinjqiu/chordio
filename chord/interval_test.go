package chord

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInterval_Has(t *testing.T) {
	t.Run("[start, end) - interval does not cross 0", func(t *testing.T) {
		int := NewInterval(7, 35, 73)
		for i := 35; i < 73; i++ {
			assert.True(t, int.Has(ChordID(i)), i)
		}

		for i := 73; i < 127; i++ {
			assert.False(t, int.Has(ChordID(i)), i)
		}
	})

	t.Run("[start, end) - interval does cross 0", func(t *testing.T) {
		int := NewInterval(7, 100, 5)
		for i := 100; i < 127; i++ {
			assert.True(t, int.Has(ChordID(i)), i)
		}

		for i := 0; i < 5; i++ {
			assert.True(t, int.Has(ChordID(i)), i)
		}

		for i := 5; i < 100; i++ {
			assert.False(t, int.Has(ChordID(i)), i)
		}
	})

	t.Run("(start, end) - does not cross 0", func(t *testing.T) {
		int := NewInterval(7, 35, 73, WithLeftOpen, WithRightOpen)

		assert.False(t, int.Has(ChordID(35)))
		for i := 36; i < 73; i++ {
			assert.True(t, int.Has(ChordID(i)), i)
		}

		for i := 73; i < 127; i++ {
			assert.False(t, int.Has(ChordID(i)), i)
		}
	})

	t.Run("(start, end) - interval does cross 0", func(t *testing.T) {
		int := NewInterval(7, 100, 5, WithLeftOpen, WithRightOpen)
		assert.False(t, int.Has(ChordID(100)))
		for i := 101; i < 127; i++ {
			assert.True(t, int.Has(ChordID(i)), i)
		}

		for i := 0; i < 5; i++ {
			assert.True(t, int.Has(ChordID(i)), i)
		}

		for i := 5; i < 100; i++ {
			assert.False(t, int.Has(ChordID(i)), i)
		}
	})

	t.Run("[start, end] - does not cross 0", func(t *testing.T) {
		int := NewInterval(7, 35, 73, WithLeftClosed, WithRightClosed)

		for i := 35; i < 74; i++ {
			assert.True(t, int.Has(ChordID(i)), i)
		}

		for i := 74; i < 127; i++ {
			assert.False(t, int.Has(ChordID(i)), i)
		}
	})

	t.Run("[start, end] - interval does cross 0", func(t *testing.T) {
		int := NewInterval(7, 100, 5, WithLeftClosed, WithRightClosed)
		for i := 100; i < 127; i++ {
			assert.True(t, int.Has(ChordID(i)), i)
		}

		for i := 0; i < 6; i++ {
			assert.True(t, int.Has(ChordID(i)), i)
		}

		for i := 6; i < 100; i++ {
			assert.False(t, int.Has(ChordID(i)), i)
		}
	})

	t.Run("(start, end] - does not cross 0", func(t *testing.T) {
		int := NewInterval(7, 35, 73, WithLeftOpen, WithRightClosed)
		assert.False(t, int.Has(ChordID(35)))
		for i := 36; i < 74; i++ {
			assert.True(t, int.Has(ChordID(i)), i)
		}

		for i := 74; i < 127; i++ {
			assert.False(t, int.Has(ChordID(i)), i)
		}
	})

	t.Run("(start, end] - interval does cross 0", func(t *testing.T) {
		int := NewInterval(7, 100, 5, WithLeftOpen, WithRightClosed)
		assert.False(t, int.Has(ChordID(100)))
		for i := 101; i < 127; i++ {
			assert.True(t, int.Has(ChordID(i)), i)
		}

		for i := 0; i < 6; i++ {
			assert.True(t, int.Has(ChordID(i)), i)
		}

		for i := 6; i < 100; i++ {
			assert.False(t, int.Has(ChordID(i)), i)
		}
	})
}

