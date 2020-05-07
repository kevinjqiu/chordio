package chord

import "math"

type ChordID uint64

func (c ChordID) Sub(other ChordID, m Rank) ChordID {
	var newID int = int(c - other)
	if newID >= 0 {
		return ChordID(newID)
	}
	max := pow2(uint32(m))
	inRangeID := uint64(newID) % max
	if inRangeID >= 0 {
		return ChordID(inRangeID)
	}
	return ChordID(inRangeID + max)
}

func (c ChordID) Add(other ChordID, m Rank) ChordID {
	max := pow2(uint32(m))
	return ChordID(uint64(c+other) % max)
}

func (c ChordID) Pow(m int) ChordID {
	max := pow2(uint32(m))
	id := uint64(math.Pow(2, float64(m))) % max
	return ChordID(id)
}

func (c ChordID) In(start, end ChordID, m Rank) bool {
	int := NewInterval(m, start, end)
	return int.Has(c)
}

func (c ChordID) AsU64() uint64 {
	return uint64(c)
}

func (c ChordID) AsInt() int {
	return int(c)
}
