package chordio

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
)

type Node struct {
	id   ChordID
	bind string
	pred ChordID
	succ ChordID
}

func newNode(bind string, m Rank) Node {
	n := Node{
		bind: bind,
	}
	n.id = assignID([]byte(bind), m)
	n.pred = n.id
	n.succ = n.id
	return n
}

func assignID(key []byte, m Rank) ChordID {
	hasher := sha1.New()
	hasher.Write(key)
	b := hasher.Sum(nil)
	return ChordID(binary.BigEndian.Uint64(b) % pow2(uint32(m)))
}

type LocalNode struct {
	Node
	m             Rank
	ft            *FingerTable
	neighbourhood *Neighbourhood
}

func (n *LocalNode) closestPrecedingFinger(id ChordID) (Node, error) {
	for i := n.m - 1; i >= 0; i-- {
		if n.ft.entries[i].node.In(n.id, id, n.m) {
			nodeID := n.ft.entries[i].node
			n, ok := n.neighbourhood.Get(nodeID)
			if !ok {
				return Node{}, fmt.Errorf("node not found: %d", nodeID)
			}
			return n, nil
		}
	}
	return n.Node, nil
}

func (n *LocalNode) initFinger(remote *RemoteNode) error {
	local := n
	succ, err := remote.FindSuccessor(n.ft.entries[0].start)
	if err != nil {
		return err
	}
	local.pred = succ.pred

	for i := 0; i < int(n.m)-1; i++ {
		if n.ft.entries[i+1].start.In(local.id, n.ft.entries[i].node, n.m) {
			n.ft.entries[i+1].node = n.ft.entries[i].node
		} else {
			newSucc, err := remote.FindSuccessor(n.ft.entries[i+1].start)
			if err != nil {
				return err
			}
			n.ft.entries[i+1].node = newSucc.id
		}
	}
	return nil
}

func (n *LocalNode) join(introducerNode Node) error {
	rn, err := newRemoteNode(introducerNode)
	if err != nil {
		return err
	}

	if err := n.initFinger(rn); err != nil {
		return err
	}

	// updateOthers()
	return nil
}

func newLocalNode(bind string, m Rank) *LocalNode {
	n := Node{
		bind: bind,
	}
	n.id = assignID([]byte(bind), m)
	n.pred = n.id
	n.succ = n.id

	localNode := &LocalNode{
		Node: n,
		ft:   nil,
		m:    m,
	}
	ft := newFingerTable(localNode.Node, m)  // TODO: make a Node interface
	localNode.ft = &ft
	neighbourhood := newNeighbourhood(m)
	neighbourhood.Add(localNode.Node)
	localNode.neighbourhood = neighbourhood
	return localNode
}
