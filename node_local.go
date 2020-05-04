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
	predNode      *NodeRef
	succNode      *NodeRef
	m             Rank
	ft            *FingerTable
	neighbourhood *Neighbourhood
}

func (n *LocalNode) SetPredNode(pn *NodeRef) {
	n.predNode = pn
	n.pred = pn.id
}

func (n *LocalNode) SetSuccNode(sn *NodeRef) {
	n.succNode = sn
	n.succ = sn.id
}

func (n *LocalNode) GetID() ChordID {
	return n.id
}

func (n *LocalNode) GetBind() string {
	return n.bind
}

func (n *LocalNode) GetPredNode() (*NodeRef, error) {
	if n.predNode != nil && n.predNode.id != n.pred {
		return n.predNode, nil
	}
	node, _, _, ok := n.neighbourhood.Get(n.pred)
	if !ok {
		return nil, fmt.Errorf("predecessor node %v not found in neighbourhood", n.pred)
	}
	return &NodeRef{bind: node.bind, id: node.id}, nil
}

func (n *LocalNode) GetSuccNode() (*NodeRef, error) {
	if n.succNode != nil && n.succNode.id != n.succ {
		return n.succNode, nil
	}
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

	succNode, err := succ.GetSuccNode()
	if err != nil {
		return err
	}
	local.SetPredNode(succNode)

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
		bind: bind,
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
