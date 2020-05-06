package chordio

import (
	"fmt"
	"github.com/kevinjqiu/chordio/pb"
	"github.com/olekukonko/tablewriter"
	"io"
	"os"
	"strconv"
)

type FingerTableEntry struct {
	start    ChordID
	interval Interval
	node     ChordID
}

type FingerTable struct {
	m       Rank
	entries []FingerTableEntry
}

func (ft FingerTable) Print(w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	writer := tablewriter.NewWriter(w)
	writer.SetHeader([]string{"Start", "[Start, End)", "Successor Node #"})
	for _, fte := range ft.entries {
		writer.Append([]string{
			strconv.Itoa(int(fte.start)),
			fmt.Sprintf("[%d, %d)", fte.start, fte.interval.end),
			strconv.Itoa(int(fte.node)),
		})
	}
	writer.Render()
}

func (ft FingerTable) HasNode(id ChordID) bool {
	for _, fte := range ft.entries {
		if fte.node == id {
			return true
		}
	}
	return false
}

func (ft FingerTable) AsProtobufFT() *pb.FingerTable {
	pbft := pb.FingerTable{}
	entries := make([]*pb.FingerTableEntry, 0, len(ft.entries))
	for _, fte := range ft.entries {
		entries = append(entries, &pb.FingerTableEntry{
			Start:  uint64(fte.start),
			End:    uint64(fte.interval.end),
			NodeID: uint64(fte.node),
		})
	}
	pbft.Entries = entries
	return &pbft
}

func newFingerTable(initNode Node, m Rank) FingerTable {
	ft := FingerTable{m: m}
	ft.entries = make([]FingerTableEntry, 0, m)

	maxKey := pow2(uint32(m))

	for k := 0; k < int(m); k++ {
		start := ChordID((uint64(initNode.GetID()) + pow2(uint32(k))) % maxKey)
		end := ChordID((uint64(initNode.GetID()) + pow2(uint32(k+1))) % maxKey)
		ft.entries = append(ft.entries, FingerTableEntry{
			start: start,
			interval: Interval{
				start: start,
				end:   end,
			},
			node: initNode.GetID(),
		})
	}

	return ft
}
