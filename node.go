package chordio

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"github.com/sirupsen/logrus"
)

type INode interface {
	GetID() ChordID
	GetBind() string
	GetPred() ChordID
	GetSucc() ChordID
}

type Node struct {
	id   ChordID
	bind string
	pred ChordID
	succ ChordID
}

func (n Node) GetID() ChordID   { return n.id }
func (n Node) GetBind() string  { return n.bind }
func (n Node) GetPred() ChordID { return n.pred }
func (n Node) GetSucc() ChordID { return n.succ }

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

func (n *LocalNode) getPredNode() (INode, error) {
	node, ok := n.neighbourhood.Get(n.pred)
	if !ok {
		return nil, fmt.Errorf("predecessor node %v not found in neighbourhood", n.pred)
	}
	return node, nil
}

func (n *LocalNode) getSuccNode() (INode, error) {
	node, ok := n.neighbourhood.Get(n.succ)
	if !ok {
		return nil, fmt.Errorf("successor node %v not found in neighbourhood", n.pred)
	}
	return node, nil
}

func (n *LocalNode) closestPrecedingFinger(id ChordID) (LocalNode, error) {
	logrus.Info("LocalNode.closestPrecedingFinger: id=", id)
	for i := n.m - 1; i >= 0; i-- {
		if n.ft.entries[i].node.In(n.id, id, n.m) {
			nodeID := n.ft.entries[i].node
			resultNode, ok := n.neighbourhood.Get(nodeID)
			logrus.Info("found result node: ", resultNode)
			if !ok {
				return LocalNode{}, fmt.Errorf("node not found: %d", nodeID)
			}
			return LocalNode{
				Node:          resultNode,
				m:             n.m,
				ft:            n.ft,
				neighbourhood: n.neighbourhood,
			}, nil
		}
	}
	return *n, nil
}

func (n *LocalNode) initFinger(remote *RemoteNode) error {
	logrus.Info("LocalNode.initFinger: using remote node ", remote)
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

func (n *LocalNode) join(introducerNode *RemoteNode) error {
	if err := n.initFinger(introducerNode); err != nil {
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
	ft := newFingerTable(localNode, m)
	localNode.ft = &ft
	neighbourhood := newNeighbourhood(m)
	neighbourhood.Add(localNode.Node)
	localNode.neighbourhood = neighbourhood
	return localNode
}
