package chordio

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

type LocalNode struct {
	id            ChordID
	bind          string
	pred          ChordID
	succ          ChordID
	m             Rank
	ft            *FingerTable
	neighbourhood *Neighbourhood
}

func (n *LocalNode) GetID() ChordID {
	return n.id
}

func (n *LocalNode) GetBind() string {
	return n.bind
}

func (n *LocalNode) GetPredNode() (*NodeRef, error) {
	node, _, _, ok := n.neighbourhood.Get(n.pred)
	if !ok {
		return nil, fmt.Errorf("predecessor node %v not found in neighbourhood", n.pred)
	}
	return &NodeRef{bind: node.bind, id: node.id}, nil
}

func (n *LocalNode) GetSuccNode() (*NodeRef, error) {
	node, _, _, ok := n.neighbourhood.Get(n.succ)
	if !ok {
		return nil, fmt.Errorf("successor node %v not found in neighbourhood", n.pred)
	}
	return &NodeRef{bind: node.bind, id: node.id}, nil
}

func (n *LocalNode) closestPrecedingFinger(id ChordID) (LocalNode, error) {
	logrus.Info("LocalNode.closestPrecedingFinger: id=", id)
	for i := n.m - 1; i >= 0; i-- {
		if n.ft.entries[i].node.In(n.id, id, n.m) {
			nodeID := n.ft.entries[i].node
			resultNode, predID, succID, ok := n.neighbourhood.Get(nodeID)
			logrus.Info("found result node: ", resultNode)
			if !ok {
				return LocalNode{}, fmt.Errorf("node not found: %d", nodeID)
			}
			return LocalNode{
				id:            resultNode.id,
				pred:          predID,
				succ:          succID,
				bind:          resultNode.bind,
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
	id := assignID([]byte(bind), m)
	localNode := &LocalNode{
		id:   id,
		pred: id,
		succ: id,
		ft:   nil,
		m:    m,
	}
	ft := newFingerTable(localNode, m)
	localNode.ft = &ft
	neighbourhood := newNeighbourhood(m)
	neighbourhood.Add(&NodeRef{id: id, bind: bind}) // TODO: handle conflict
	localNode.neighbourhood = neighbourhood
	return localNode
}
