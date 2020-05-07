package chord

import (
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestRank(t *testing.T) {
	t.Run("As int", func(t *testing.T) {
		m := Rank(10)
		assert.Equal(t, 10, m.AsInt())
	})

	t.Run("As uint32", func(t *testing.T) {
		m := Rank(10)
		assert.Equal(t, uint32(10), m.AsU32())
	})
}
