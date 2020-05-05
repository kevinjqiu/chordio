package chordio

import (
	"context"
	"fmt"
	"github.com/kevinjqiu/chordio/pb"
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

func (n *LocalNode) String() string {
	var pred, succ string

	if n.predNode == nil {
		pred = "<nil>"
	} else {
		pred = fmt.Sprintf("%d@%s", n.predNode.id, n.predNode.bind)
	}

	if n.succNode == nil {
		succ = "<nil>"
	} else {
		succ = fmt.Sprintf("%d@%s", n.succNode.id, n.succNode.bind)
	}
	return fmt.Sprintf("<L: %d@%s, p=%s, s=%s>", n.id, n.bind, pred, succ)
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

func (n *LocalNode) AsProtobufNode() *pb.Node {
	pbn := &pb.Node{
		Id:   uint64(n.GetID()),
		Bind: n.GetBind(),
		Pred: nil,
		Succ: nil,
	}

	predNode, err := n.GetPredNode()
	if err == nil {
		pbn.Pred = &pb.Node{
			Id:   uint64(predNode.id),
			Bind: predNode.bind,
		}
	}

	succNode, err := n.GetSuccNode()
	if err == nil {
		pbn.Succ = &pb.Node{
			Id:   uint64(succNode.id),
			Bind: succNode.bind,
		}
	}
	return pbn
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
	logger := logrus.WithField("method", "LocalNode.closestPrecedingFinger")
	logger.Debugf("id=%d", id)
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

func (n *LocalNode) initFinger(ctx context.Context, remote *RemoteNode) error {
	logger := logrus.WithField("method", "LocalNode.initFinger")
	logger.Infof("using remote node %s", remote)
	local := n
	logger.Debugf("Try to find success for %d on %s", n.ft.entries[0].start, remote)
	succ, err := remote.FindSuccessor(ctx, n.ft.entries[0].start)
	if err != nil {
		return err
	}

	logger.Debugf("Successor node for %d is %s", n.ft.entries[0].start, succ)
	n.ft.entries[0].node = succ.id
	predNode, err := succ.GetPredNode()
	if err != nil {
		return err
	}
	local.SetPredNode(predNode)
	logger.Debugf("Local node's predecessor set to %v", predNode)

	logger.Debug("recalc finger table")
	for i := 0; i < int(n.m)-1; i++ {
		logger.Debugf("i=%d", i)
		logger.Debugf("finger[i+1].start=%d", n.ft.entries[i+1].start)
		logger.Debugf("interval=[%d, %d)", local.id, n.ft.entries[i].node)

		if n.ft.entries[i+1].start.In(local.id, n.ft.entries[i].node, n.m) {
			logger.Debugf("interval=[%d, %d)", local.id, n.ft.entries[i].node)
			n.ft.entries[i+1].node = n.ft.entries[i].node
		} else {
			newSucc, err := remote.FindSuccessor(ctx, n.ft.entries[i+1].start)
			logger.Debugf("new successor for %d is %v", n.ft.entries[i+1].start, newSucc)
			if err != nil {
				return err
			}
			n.ft.entries[i+1].node = newSucc.id
		}
	}
	return nil
}

func (n *LocalNode) join(ctx context.Context, introducerNode *RemoteNode) error {
	logger := logrus.WithField("method", "LocalNode.join")
	logger.Debugf("introducerNode: %s", introducerNode)

	if err := n.initFinger(ctx, introducerNode); err != nil {
		return err
	}
	n.ft.Print(nil)

	// updateOthers()
	return nil
}

func newLocalNode(id ChordID, bind string, m Rank) (*LocalNode, error) {
	//id := AssignID([]byte(bind), m)
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
	if err := neighbourhood.Add(&NodeRef{id: id, bind: bind}); err != nil {
		return nil, err
	}
	localNode.neighbourhood = neighbourhood
	return localNode, nil
}
