package node

import (
	"github.com/kevinjqiu/chordio/chord"
	"sort"
)

type nodeList []*NodeRef

func (n nodeList) Len() int {
	return len(n)
}

func (n nodeList) Less(i, j int) bool {
	return n[i].ID < n[j].ID
}

func (n nodeList) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

// A neighbourhood is a group of nodes that a local NodeID knows about
type Neighbourhood struct {
	nodes nodeList
	idMap map[chord.ID]interface{}
}

func (neigh *Neighbourhood) Add(node *NodeRef) error {
	_, ok := neigh.idMap[node.ID]
	if ok {
		return chord.ErrNodeIDConflict
	}

	neigh.nodes = append(neigh.nodes, node)
	neigh.idMap[node.ID] = ""
	sort.Sort(neigh.nodes)
	return nil
}

func (neigh *Neighbourhood) Remove(nodeID chord.ID) {
	idx := sort.Search(len(neigh.nodes), func(i int) bool {
		return neigh.nodes[i].ID >= nodeID
	})

	if idx == -1 || idx >= len(neigh.nodes) || neigh.nodes[idx].ID != nodeID {
		return
	}

	lastIdx := len(neigh.nodes) - 1
	neigh.nodes[idx], neigh.nodes[lastIdx] = neigh.nodes[lastIdx], neigh.nodes[idx]
	neigh.nodes = neigh.nodes[:lastIdx]
	sort.Sort(neigh.nodes)
}

// GetEntry the NodeRef for the NodeID given the ID, as well as the ID for the preceding and succeeding nodes
func (neigh *Neighbourhood) Get(id chord.ID) (n NodeRef, ok bool) {
	idx := sort.Search(len(neigh.nodes), func(i int) bool {
		return neigh.nodes[i].ID >= id
	})
	if idx == -1 || idx >= len(neigh.nodes) || neigh.nodes[idx].ID != id {
		return
	}

	ok = true
	n = NodeRef{
		ID:   neigh.nodes[idx].ID,
		Bind: neigh.nodes[idx].Bind,
	}
	return
}

func newNeighbourhood(m chord.Rank) *Neighbourhood {
	return &Neighbourhood{
		nodes: make([]*NodeRef, 0, int(m)),
		idMap: make(map[chord.ID]interface{}),
	}
}
