package pkg

import "math"

// An interval between [Start, End)
type Interval struct {
	m          Rank
	start, end ChordID
}

func (i Interval) Has(id ChordID) bool {
	if i.start < i.end {
		return i.start <= id && id < i.end
	}
	max := ChordID(pow2(uint32(i.m)))
	return i.start <= id && id < max || 0 <= id && id < i.end
}

func NewInterval(m Rank, start, end ChordID) Interval {
	return Interval{
		m:     m,
		start: start,
		end:   end,
	}
}

func pow2(exp uint32) uint64 {
	return uint64(math.Pow(2, float64(exp)))
}
