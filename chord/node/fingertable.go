package node

import (
	"bytes"
	"fmt"
	"github.com/kevinjqiu/chordio/chord"
	"github.com/kevinjqiu/chordio/pb"
	"github.com/olekukonko/tablewriter"
	"io"
	"os"
	"strconv"
)

type fingerTableEntry struct {
	Start    chord.ID
	Interval chord.Interval
	Node     chord.NodeRef
}

func (fte *fingerTableEntry) SetStart(start chord.ID) {
	fte.Start = start
}

func (fte *fingerTableEntry) SetInterval(iv chord.Interval) {
	fte.Interval = iv
}

func (fte *fingerTableEntry) SetNode(n chord.NodeRef) {
	fte.Node = n
}

func (fte fingerTableEntry) GetStart() chord.ID {
	return fte.Start
}

func (fte fingerTableEntry) GetInterval() chord.Interval {
	return fte.Interval
}

func (fte fingerTableEntry) GetNode() chord.NodeRef {
	return fte.Node
}

func (fte fingerTableEntry) String() string {
	return fmt.Sprintf("%d\t%s\t%s", fte.Start, fte.Interval.String(), fte.Node)
}

type fingerTable struct {
	ownerID       chord.ID
	m             chord.Rank
	entries       []chord.FingerTableEntry
	neighbourhood map[chord.ID]chord.NodeRef
}

func (ft *fingerTable) String() string {
	var b bytes.Buffer
	for _, fte := range ft.entries {
		b.WriteString(fte.String())
		b.WriteString("|")
	}
	return b.String()
}

func (ft *fingerTable) PrettyPrint(w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	writer := tablewriter.NewWriter(w)
	writer.SetHeader([]string{"Start", "[Start, End)", "Successor Node #"})
	for _, fte := range ft.entries {
		writer.Append([]string{
			strconv.Itoa(int(fte.GetStart())),
			fmt.Sprintf(fte.GetInterval().String()),
			fte.GetNode().String(),
		})
	}
	writer.Render()
}

func (ft *fingerTable) Len() int {
	return len(ft.entries)
}

// SetNodeAtEntry the i'th finger table entry's NodeID to n
func (ft *fingerTable) SetNodeAtEntry(i int, n chord.NodeRef) {
	oldNodeRef := ft.entries[i].GetNode()
	if oldNodeRef.GetID() == n.GetID() {
		return
	}

	newNodeRef, ok := ft.neighbourhood[n.GetID()]
	if !ok {
		newNodeRef = &nodeRef{
			ID:   n.GetID(),
			Bind: n.GetBind(),
		}
		ft.neighbourhood[n.GetID()] = newNodeRef
	}

	ft.entries[i].SetNode(newNodeRef)

	if oldNodeRef.GetID() == ft.ownerID {
		// Do not delete the node from the neighbourhood if it's the owner of the fingertable
		return
	}

	for _, fte := range ft.entries {
		if fte.GetNode().GetID() == oldNodeRef.GetID() {
			// Old node still in the finger table
			// Do not delete it from the neighbourhood
			return
		}
	}
	delete(ft.neighbourhood, oldNodeRef.GetID())
}

func (ft *fingerTable) GetEntry(i int) chord.FingerTableEntry {
	return ft.entries[i]
}

func (ft *fingerTable) GetNodeByID(nodeID chord.ID) (chord.NodeRef, bool) {
	nodeRef, ok := ft.neighbourhood[nodeID]
	return nodeRef, ok
}

func (ft *fingerTable) HasNode(id chord.ID) bool {
	_, ok := ft.neighbourhood[id]
	return ok
}

func (ft *fingerTable) AsProtobufFT() *pb.FingerTable {
	pbft := pb.FingerTable{}
	entries := make([]*pb.FingerTableEntry, 0, len(ft.entries))
	for _, fte := range ft.entries {
		entries = append(entries, &pb.FingerTableEntry{
			Start:  uint64(fte.GetStart()),
			End:    uint64(fte.GetInterval().End),
			NodeID: uint64(fte.GetNode().GetID()),
		})
	}
	pbft.Entries = entries
	return &pbft
}

func newFingerTable(initNode chord.Node, m chord.Rank) chord.FingerTable {
	ft := fingerTable{
		m:             m,
		ownerID:       initNode.GetID(),
		neighbourhood: make(map[chord.ID]chord.NodeRef),
		entries:       make([]chord.FingerTableEntry, 0, m),
	}

	initNodeRef := &nodeRef{
		ID:   initNode.GetID(),
		Bind: initNode.GetBind(),
	}

	for k := 0; k < int(m); k++ {
		start := initNode.GetID().Add(chord.ID(2).Pow(k), m)
		end := initNode.GetID().Add(chord.ID(2).Pow(k+1), m)
		ft.entries = append(ft.entries, &fingerTableEntry{
			Start:    start,
			Interval: chord.NewInterval(m, start, end),
			Node:     initNodeRef,
		})
	}

	ft.neighbourhood[initNode.GetID()] = initNodeRef
	return &ft
}
