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
	ownerID       ChordID
	m             Rank
	entries       []FingerTableEntry
	neighbourhood *Neighbourhood
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

// Set the i'th finger table entry's node to id
// The node represented by the id must already exist
// in the neighbourhood
func (ft FingerTable) SetID(i int, id ChordID) error {
	oldNodeID := ft.entries[i].node
	if oldNodeID == id {
		return nil
	}

	_, _, _, ok := ft.neighbourhood.Get(id)
	if !ok {
		return fmt.Errorf("cannot set %dth fingertable entry to %d: node %d not found in the neighbourhood", i, id, id)
	}

	ft.entries[i].node = id

	if oldNodeID != ft.ownerID && !ft.HasNode(oldNodeID) {
		ft.neighbourhood.Remove(oldNodeID)
	}
	return nil
}

// Set the i'th finger table entry's node to n
func (ft FingerTable) Set(i int, n Node) {
	oldNodeID := ft.entries[i].node
	if oldNodeID == n.GetID() {
		return
	}

	ft.entries[i].node = n.GetID()
	_ = ft.neighbourhood.Add(&NodeRef{
		id: n.GetID(),
		bind: n.GetBind(),
	})

	if oldNodeID != ft.ownerID && !ft.HasNode(oldNodeID) {
		ft.neighbourhood.Remove(oldNodeID)
	}
}

func (ft FingerTable) Get(i int) FingerTableEntry {
	return ft.entries[i]
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
	ft := FingerTable{
		m: m,
		ownerID: initNode.GetID(),
		neighbourhood: newNeighbourhood(m),
	}

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

	_ = ft.neighbourhood.Add(&NodeRef{
		id:   initNode.GetID(),
		bind: initNode.GetBind(),
	})

	return ft
}
