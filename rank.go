package chordio

type Rank uint32 // otherwise known as the m value

func (r Rank) AsInt() int {
	return int(r)
}

func (r Rank) AsU32() uint32 {
	return uint32(r)
}
