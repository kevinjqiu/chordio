package chordio

type FingerTableEntry struct {
	start    ChordID
	interval Interval
	node     ChordID
}

type FingerTable struct {
	m       Rank
	entries []FingerTableEntry
}

func newFingerTable(initNode Node, m Rank) FingerTable {
	ft := FingerTable{m: m}
	ft.entries = make([]FingerTableEntry, 0, m)

	maxKey := pow2(uint32(m))

	for k := 0; k < int(m); k++ {
		start := ChordID((uint64(initNode.id) + pow2(uint32(k))) % maxKey)
		end := ChordID((uint64(initNode.id) + pow2(uint32(k+1))) % maxKey)
		ft.entries = append(ft.entries, FingerTableEntry{
			start: start,
			interval: Interval{
				start: start,
				end:   end,
			},
			node: initNode.id,
		})
	}

	return ft
}
