package chordio

import (
	"errors"
	"sort"
)

var (
	errNodeIDConflict = errors.New("conflict node id")
)

type nodeList []*Node

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
	idMap map[ChordID]*Node
}

func (n *Neighbourhood) Add(node *Node) error {
	_, ok := n.idMap[node.id]
	if ok {
		return errNodeIDConflict
	}

	n.nodes = append(n.nodes, node)
	n.idMap[node.id] = node
	sort.Sort(n.nodes)
	return nil
}

func (n *Neighbourhood) Get(id ChordID) (Node, bool) {
	idx := sort.Search(len(n.nodes), func(i int) bool {
		return n.nodes[i].id >= id
	})
	if idx == -1 {
		return Node{}, false
	}
	if idx >= len(n.nodes) || n.nodes[idx].id != id {
		return Node{}, false
	}
	node := Node{
		id: n.nodes[idx].id,
		bind: n.nodes[idx].bind,
	}
	if idx == 0 {
		node.pred = n.nodes[len(n.nodes)-1].id
	} else {
		node.pred = n.nodes[idx-1].id
	}
	if idx == len(n.nodes) - 1 {
		node.succ = n.nodes[0].id
	} else {
		node.succ = n.nodes[idx+1].id
	}
	return node, true
}

func newNeighbourhood(m Rank) *Neighbourhood {
	return &Neighbourhood{
		nodes: make([]*Node, 0, int(m)),
		idMap: make(map[ChordID]*Node),
	}
}
