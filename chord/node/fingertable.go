package node

import (
	"fmt"
	"github.com/kevinjqiu/chordio/chord"
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
	neighbourhood map[chord.ID]*NodeRef
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

	_, ok := ft.neighbourhood[id]
	if !ok {
		return fmt.Errorf("cannot set %dth fingertable entry to %d: NodeID %d not found in the neighbourhood", i, id, id)
	}

	ft.entries[i].NodeID = id

	if oldNodeID != ft.ownerID && !ft.HasNode(oldNodeID) {
		delete(ft.neighbourhood, oldNodeID)
	}
	return nil
}

// SetEntry the i'th finger table entry's NodeID to n
func (ft FingerTable) SetEntry(i int, n Node) {
	oldNodeID := ft.entries[i].NodeID
	if oldNodeID == n.GetID() {
		return
	}

	ft.entries[i].NodeID = n.GetID()
	ft.neighbourhood[n.GetID()] = &NodeRef{
		ID:   n.GetID(),
		Bind: n.GetBind(),
	}

	if oldNodeID != ft.ownerID && !ft.HasNode(oldNodeID) {
		delete(ft.neighbourhood, oldNodeID)
	}
}

func (ft FingerTable) GetEntry(i int) FingerTableEntry {
	return ft.entries[i]
}

// Get the NodeID, pred, succ at fingertable entry index i
func (ft FingerTable) GetNodeByFingerIdx(i int) (NodeRef, bool) {
	nodeID := ft.entries[i].NodeID
	nodeRef, ok := ft.neighbourhood[nodeID]
	return *nodeRef, ok
}

func (ft FingerTable) GetNodeByID(nodeID chord.ID) (NodeRef, bool) {
	nodeRef, ok := ft.neighbourhood[nodeID]
	return *nodeRef, ok
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

func newFingerTable(initNode Node, m chord.Rank) FingerTable {
	ft := FingerTable{
		m:             m,
		ownerID:       initNode.GetID(),
		neighbourhood: make(map[chord.ID]*NodeRef),
	}

	ft.entries = make([]FingerTableEntry, 0, m)

	for k := 0; k < int(m); k++ {
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

	ft.neighbourhood[initNode.GetID()] = &NodeRef{
		ID: initNode.GetID(),
		Bind: initNode.GetBind(),
	}
	return ft
}
