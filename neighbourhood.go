package chordio

import (
	"sort"
)

type nodeList []*NodeRef

func (n nodeList) Len() int {
	return len(n)
}

func (n nodeList) Less(i, j int) bool {
	return n[i].id < n[j].id
}

func (n nodeList) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

// A neighbourhood is a group of nodes that a local node knows about
type Neighbourhood struct {
	nodes nodeList
	idMap map[ChordID]interface{}
}

func (n *Neighbourhood) Add(node *NodeRef) error {
	_, ok := n.idMap[node.id]
	if ok {
		return errNodeIDConflict
	}

	n.nodes = append(n.nodes, node)
	n.idMap[node.id] = ""
	sort.Sort(n.nodes)
	return nil
}

func (n *Neighbourhood) Remove(nodeID ChordID) {
	idx := sort.Search(len(n.nodes), func(i int) bool {
		return n.nodes[i].id >= nodeID
	})

	if idx == -1 || idx >= len(n.nodes) || n.nodes[idx].id != nodeID {
		return
	}

	lastIdx := len(n.nodes) - 1
	n.nodes[idx], n.nodes[lastIdx] = n.nodes[lastIdx], n.nodes[idx]
	n.nodes = n.nodes[:lastIdx]
	sort.Sort(n.nodes)
}

// GetEntry the NodeRef for the node given the ID, as well as the ID for the preceding and succeeding nodes
func (n *Neighbourhood) Get(id ChordID) (node NodeRef, predID ChordID, succID ChordID, ok bool) {
	idx := sort.Search(len(n.nodes), func(i int) bool {
		return n.nodes[i].id >= id
	})
	if idx == -1 || idx >= len(n.nodes) || n.nodes[idx].id != id {
		return
	}

	ok = true
	node = NodeRef{
		id: n.nodes[idx].id,
		bind: n.nodes[idx].bind,
	}
	if idx == 0 {
		predID = n.nodes[len(n.nodes)-1].id
	} else {
		predID = n.nodes[idx-1].id
	}
	if idx == len(n.nodes) - 1 {
		succID = n.nodes[0].id
	} else {
		succID = n.nodes[idx+1].id
	}
	return
}

func newNeighbourhood(m Rank) *Neighbourhood {
	return &Neighbourhood{
		nodes: make([]*NodeRef, 0, int(m)),
		idMap: make(map[ChordID]interface{}),
	}
}
