package chord

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInterval_Has(t *testing.T) {
	t.Run("[Start, End) - interval does not cross 0", func(t *testing.T) {
		int := NewInterval(7, 35, 73)
		assert.Equal(t, "[35, 73)", int.String())
		for i := 35; i < 73; i++ {
			assert.True(t, int.Has(ID(i)), i)
		}

		for i := 73; i < 127; i++ {
			assert.False(t, int.Has(ID(i)), i)
		}
	})

	t.Run("[Start, End) - interval does cross 0", func(t *testing.T) {
		int := NewInterval(7, 100, 5)
		assert.Equal(t, "[100, 5)", int.String())
		for i := 100; i < 127; i++ {
			assert.True(t, int.Has(ID(i)), i)
		}

		for i := 0; i < 5; i++ {
			assert.True(t, int.Has(ID(i)), i)
		}

		for i := 5; i < 100; i++ {
			assert.False(t, int.Has(ID(i)), i)
		}
	})

	t.Run("(Start, End) - does not cross 0", func(t *testing.T) {
		int := NewInterval(7, 35, 73, WithLeftOpen, WithRightOpen)
		assert.Equal(t, "(35, 73)", int.String())

		assert.False(t, int.Has(ID(35)))
		for i := 36; i < 73; i++ {
			assert.True(t, int.Has(ID(i)), i)
		}

		for i := 73; i < 127; i++ {
			assert.False(t, int.Has(ID(i)), i)
		}
	})

	t.Run("(Start, End) - interval does cross 0", func(t *testing.T) {
		int := NewInterval(7, 100, 5, WithLeftOpen, WithRightOpen)
		assert.Equal(t, "(100, 5)", int.String())

		assert.False(t, int.Has(ID(100)))
		for i := 101; i < 127; i++ {
			assert.True(t, int.Has(ID(i)), i)
		}

		for i := 0; i < 5; i++ {
			assert.True(t, int.Has(ID(i)), i)
		}

		for i := 5; i < 100; i++ {
			assert.False(t, int.Has(ID(i)), i)
		}
	})

	t.Run("[Start, End] - does not cross 0", func(t *testing.T) {
		int := NewInterval(7, 35, 73, WithLeftClosed, WithRightClosed)
		assert.Equal(t, "[35, 73]", int.String())

		for i := 35; i < 74; i++ {
			assert.True(t, int.Has(ID(i)), i)
		}

		for i := 74; i < 127; i++ {
			assert.False(t, int.Has(ID(i)), i)
		}
	})

	t.Run("[Start, End] - interval does cross 0", func(t *testing.T) {
		int := NewInterval(7, 100, 5, WithLeftClosed, WithRightClosed)
		assert.Equal(t, "[100, 5]", int.String())

		for i := 100; i < 127; i++ {
			assert.True(t, int.Has(ID(i)), i)
		}

		for i := 0; i < 6; i++ {
			assert.True(t, int.Has(ID(i)), i)
		}

		for i := 6; i < 100; i++ {
			assert.False(t, int.Has(ID(i)), i)
		}
	})

	t.Run("(Start, End] - does not cross 0", func(t *testing.T) {
		int := NewInterval(7, 35, 73, WithLeftOpen, WithRightClosed)
		assert.Equal(t, "(35, 73]", int.String())

		assert.False(t, int.Has(ID(35)))
		for i := 36; i < 74; i++ {
			assert.True(t, int.Has(ID(i)), i)
		}

		for i := 74; i < 127; i++ {
			assert.False(t, int.Has(ID(i)), i)
		}
	})

	t.Run("(Start, End] - interval does cross 0", func(t *testing.T) {
		int := NewInterval(7, 100, 5, WithLeftOpen, WithRightClosed)
		assert.Equal(t, "(100, 5]", int.String())

		assert.False(t, int.Has(ID(100)))
		for i := 101; i < 127; i++ {
			assert.True(t, int.Has(ID(i)), i)
		}

		for i := 0; i < 6; i++ {
			assert.True(t, int.Has(ID(i)), i)
		}

		for i := 6; i < 100; i++ {
			assert.False(t, int.Has(ID(i)), i)
		}
	})

	t.Run("(start, start) - should contain all", func(t *testing.T) {
		iv := NewInterval(7, 0, 0, WithLeftOpen, WithRightOpen)
		assert.Equal(t, "(0, 0)", iv.String())
		assert.False(t, iv.Has(0), 0)
		for i := 1; i < 128; i++ {
			assert.True(t, iv.Has(ID(i)), i)
		}

		iv = NewInterval(7, 1, 1, WithLeftOpen, WithRightOpen)
		assert.Equal(t, "(1, 1)", iv.String())
		assert.True(t, iv.Has(0), 0)
		assert.False(t, iv.Has(1), 1)
		for i := 2; i < 128; i++ {
			assert.True(t, iv.Has(ID(i)), i)
		}
	})
}
