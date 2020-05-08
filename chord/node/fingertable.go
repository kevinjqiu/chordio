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
	Node     *NodeRef
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
			fte.Node.String(),
		})
	}
	writer.Render()
}

// SetEntryAt the i'th finger table entry's NodeID to id
// The NodeID represented by the id must already exist
// in the neighbourhood
func (ft FingerTable) SetID(i int, id chord.ID) error {
	oldNodeID := ft.entries[i].Node.ID
	if oldNodeID == id {
		return nil
	}

	newNodeRef, ok := ft.neighbourhood[id]
	if !ok {
		return fmt.Errorf("cannot set %dth fingertable entry to %d: NodeID %d not found in the neighbourhood", i, id, id)
	}

	ft.entries[i].Node = newNodeRef

	if oldNodeID != ft.ownerID && !ft.HasNode(oldNodeID) {
		delete(ft.neighbourhood, oldNodeID)
	}
	return nil
}

// SetEntryAt the i'th finger table entry's NodeID to n
func (ft FingerTable) SetEntryAt(i int, n Node) {
	oldNodeID := ft.entries[i].Node.ID
	if oldNodeID == n.GetID() {
		return
	}

	newNodeRef, ok := ft.neighbourhood[n.GetID()]
	if !ok {
		newNodeRef = &NodeRef{
			ID: n.GetID(),
			Bind: n.GetBind(),
		}
		ft.neighbourhood[n.GetID()] = newNodeRef
	}

	ft.entries[i].Node = newNodeRef

	if oldNodeID != ft.ownerID && !ft.HasNode(oldNodeID) {
		delete(ft.neighbourhood, oldNodeID)
	}
}

func (ft FingerTable) GetEntry(i int) FingerTableEntry {
	return ft.entries[i]
}

func (ft FingerTable) GetNodeByID(nodeID chord.ID) (*NodeRef, bool) {
	nodeRef, ok := ft.neighbourhood[nodeID]
	return nodeRef, ok
}

func (ft FingerTable) HasNode(id chord.ID) bool {
	_, ok := ft.neighbourhood[id]
	return ok
}

func (ft FingerTable) AsProtobufFT() *pb.FingerTable {
	pbft := pb.FingerTable{}
	entries := make([]*pb.FingerTableEntry, 0, len(ft.entries))
	for _, fte := range ft.entries {
		entries = append(entries, &pb.FingerTableEntry{
			Start:  uint64(fte.Start),
			End:    uint64(fte.Interval.End),
			NodeID: uint64(fte.Node.ID),
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

	initNodeRef := &NodeRef{
		ID: initNode.GetID(),
		Bind: initNode.GetBind(),
	}

	for k := 0; k < int(m); k++ {
		start := initNode.GetID().Add(chord.ID(2).Pow(k), m)
		end := initNode.GetID().Add(chord.ID(2).Pow(k+1), m)
		ft.entries = append(ft.entries, FingerTableEntry{
			Start: start,
			Interval: chord.Interval{
				Start: start,
				End:   end,
			},
			Node: initNodeRef,
		})
	}

	ft.neighbourhood[initNode.GetID()] = initNodeRef
	return ft
}
