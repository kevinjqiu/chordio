package ft

import (
	"fmt"
	"github.com/kevinjqiu/chordio/chord"
	"github.com/kevinjqiu/chordio/chord/node"
	"github.com/kevinjqiu/chordio/pb"
	"github.com/olekukonko/tablewriter"
	"io"
	"os"
	"strconv"
)

type FingerTableEntry struct {
	Start    chord.ID
	Interval chord.Interval
	NodeID   chord.ID
}

type FingerTable struct {
	ownerID       chord.ID
	m             chord.Rank
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
			strconv.Itoa(int(fte.Start)),
			fmt.Sprintf("[%d, %d)", fte.Start, fte.Interval.End),
			strconv.Itoa(int(fte.NodeID)),
		})
	}
	writer.Render()
}

// SetEntry the i'th finger table entry's NodeID to id
// The NodeID represented by the id must already exist
// in the neighbourhood
func (ft FingerTable) SetID(i int, id chord.ID) error {
	oldNodeID := ft.entries[i].NodeID
	if oldNodeID == id {
		return nil
	}

	_, _, _, ok := ft.neighbourhood.Get(id)
	if !ok {
		return fmt.Errorf("cannot set %dth fingertable entry to %d: NodeID %d not found in the neighbourhood", i, id, id)
	}

	ft.entries[i].NodeID = id

	if oldNodeID != ft.ownerID && !ft.HasNode(oldNodeID) {
		ft.neighbourhood.Remove(oldNodeID)
	}
	return nil
}

// SetEntry the i'th finger table entry's NodeID to n
func (ft FingerTable) SetEntry(i int, n node.Node) {
	oldNodeID := ft.entries[i].NodeID
	if oldNodeID == n.GetID() {
		return
	}

	ft.entries[i].NodeID = n.GetID()
	_ = ft.neighbourhood.Add(&node.NodeRef{
		Id:   n.GetID(),
		Bind: n.GetBind(),
	})

	if oldNodeID != ft.ownerID && !ft.HasNode(oldNodeID) {
		ft.neighbourhood.Remove(oldNodeID)
	}
}

func (ft FingerTable) GetEntry(i int) FingerTableEntry {
	return ft.entries[i]
}

// Get the NodeID, pred, succ at fingertable entry index i
func (ft FingerTable) GetNodeByFingerIdx(i int) (node.NodeRef, chord.ID, chord.ID, bool) {
	nodeID := ft.entries[i].NodeID
	return ft.neighbourhood.Get(nodeID)
}

func (ft FingerTable) GetNodeByID(nodeID chord.ID) (node.NodeRef, chord.ID, chord.ID, bool) {
	return ft.neighbourhood.Get(nodeID)
}

func (ft FingerTable) HasNode(id chord.ID) bool {
	for _, fte := range ft.entries {
		if fte.NodeID == id {
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
			Start:  uint64(fte.Start),
			End:    uint64(fte.Interval.End),
			NodeID: uint64(fte.NodeID),
		})
	}
	pbft.Entries = entries
	return &pbft
}

func New(initNode node.Node, m chord.Rank) FingerTable {
	ft := FingerTable{
		m:             m,
		ownerID:       initNode.GetID(),
		neighbourhood: newNeighbourhood(m),
	}

	ft.entries = make([]FingerTableEntry, 0, m)

	for k := 0; k < int(m); k++ {
		//start := initNode.GetID().Add(chord.ID(chord.pow2(uint32(k))), m)
		//end := initNode.GetID().Add(chord.ID(chord.pow2(uint32(k+1))), m)
		start := initNode.GetID().Add(chord.ID(2).Pow(k), m)
		end := initNode.GetID().Add(chord.ID(2).Pow(k+1), m)
		ft.entries = append(ft.entries, FingerTableEntry{
			Start: start,
			Interval: chord.Interval{
				Start: start,
				End:   end,
			},
			NodeID: initNode.GetID(),
		})
	}

	_ = ft.neighbourhood.Add(&node.NodeRef{
		ID:   initNode.GetID(),
		Bind: initNode.GetBind(),
	})

	return ft
}
