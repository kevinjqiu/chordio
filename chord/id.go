package chord

import "math"

type ID uint64

func (c ID) Sub(other ID, m Rank) ID {
	var newID int = int(c - other)
	if newID >= 0 {
		return ID(newID)
	}
	max := pow2(uint32(m))
	inRangeID := uint64(newID) % max
	if inRangeID >= 0 {
		return ID(inRangeID)
	}
	return ID(inRangeID + max)
}

func (c ID) Add(other ID, m Rank) ID {
	max := pow2(uint32(m))
	return ID(uint64(c+other) % max)
}

func (c ID) Pow(m int) ID {
	id := uint64(math.Pow(2, float64(m)))
	return ID(id)
}

func (c ID) In(start, end ID, m Rank) bool {
	int := NewInterval(m, start, end)
	return int.Has(c)
}

func (c ID) AsU64() uint64 {
	return uint64(c)
}

func (c ID) AsInt() int {
	return int(c)
}
