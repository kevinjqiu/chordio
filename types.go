package chordio

type (
	ChordID uint64
	Rank    uint32 // otherwise known as the m value
)

func (c ChordID) In(start, end ChordID, m Rank) bool {
	int := NewInterval(m, start, end)
	return int.Has(c)
}