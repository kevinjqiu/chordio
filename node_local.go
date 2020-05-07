package chordio

import (
	"context"
	"fmt"
	"github.com/kevinjqiu/chordio/pb"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"
)

type LocalNode struct {
	trace.Tracer
	id            ChordID
	bind          string
	pred          ChordID
	succ          ChordID
	predNode      *NodeRef
	succNode      *NodeRef
	m             Rank
	ft            *FingerTable
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
	node, _, _, ok := n.ft.neighbourhood.Get(n.pred)
	if !ok {
		return nil, fmt.Errorf("predecessor node %v not found in neighbourhood", n.pred)
	}
	return &NodeRef{bind: node.bind, id: node.id}, nil
}

func (n *LocalNode) GetSuccNode() (*NodeRef, error) {
	if n.succNode != nil && n.succNode.id != n.succ {
		return n.succNode, nil
	}
	node, _, _, ok := n.ft.neighbourhood.Get(n.succ)
	if !ok {
		return nil, fmt.Errorf("successor node %v not found in neighbourhood", n.pred)
	}
	return &NodeRef{bind: node.bind, id: node.id}, nil
}

// TODO: if the node is itself, do not return a RemoteNode version of it
func (n *LocalNode) findPredecessor(ctx context.Context, id ChordID) (Node, error) {
	logger := logrus.WithField("method", "LocalNode.findPredecessor")
	var (
		node       Node
		remoteNode Node
	)

	err := n.WithSpan(ctx, "LocalNode.findPredecessor", func(ctx context.Context) error {
		var err error

		if !id.In(n.id, n.succ, n.m) {
			logger.Debugf("id is within %v, the predecessor is the local node", n)
			node = n
			return nil
		}

		n_, err := n.closestPrecedingFinger(ctx, id)
		if err != nil {
			return err
		}

		logger.Debugf("the closest preceding node is %v", n_)
		remoteNode, err = newRemoteNode(ctx, n_.GetBind())
		if err != nil {
			return err
		}

		for {
			succNode, err := n_.GetSuccNode()
			if err != nil {
				return err
			}
			if !id.In(n_.GetID(), succNode.id, n.m) { // FIXME: not in (a, b]
				logger.Debugf("id is not in %v's range", n_)
				remoteNode, err = remoteNode.closestPrecedingFinger(ctx, id)
				if err != nil {
					return err
				}
				logger.Debugf("the closest preceding node in %s's finger table is: ", remoteNode)
			} else {
				logger.Debugf("id is in %v's range", n_)
				break
			}
		}

		node = remoteNode
		return nil
	})

	return node, err
}

func (n *LocalNode) findSuccessor(ctx context.Context, id ChordID) (Node, error) {
	predNode, err := n.findPredecessor(ctx, id)
	if err != nil {
		return nil, err
	}

	succNode, err := predNode.GetSuccNode()
	if err != nil {
		return nil, err
	}

	return newRemoteNode(ctx, succNode.bind)
}

func (n *LocalNode) closestPrecedingFinger(ctx context.Context, id ChordID) (Node, error) {
	var ln *LocalNode
	err := n.WithSpan(ctx, "LocalNode.closestPrecedingFinger", func(ctx context.Context) error {
		logger := logrus.WithField("method", "LocalNode.closestPrecedingFinger")
		logger.Debugf("id=%d", id)
		// nb: int cast here is IMPORTANT!
		// because n.m is of type uint32
		// i >= 0 is always going to be true
		for i := int(n.m) - 1; i >= 0; i-- {
			fte := n.ft.GetEntry(i)
			if fte.node.In(n.id, id, n.m) {
				nodeID := fte.node
				resultNode, predID, succID, ok := n.ft.neighbourhood.Get(nodeID)
				logger.Info("found result node: ", resultNode)
				if !ok {
					return fmt.Errorf("node not found for id: %d", nodeID)
				}
				ln = &LocalNode{
					id:            resultNode.id,
					pred:          predID,
					succ:          succID,
					bind:          resultNode.bind,
					m:             n.m,
					ft:            n.ft,
				}
				return nil
			}
		}
		ln = n
		return nil
	})
	return ln, err
}

// initialize finger table of the local node
// n' is an arbitrary node already in the network
func (n *LocalNode) initFinger(ctx context.Context, remote *RemoteNode) error {
	ctx, span := n.Start(ctx, "LocalNode.initFinger")
	defer span.End()

	logger := logrus.WithField("method", "LocalNode.initFinger")
	logger.Infof("using remote node %s", remote)
	local := n
	logger.Debugf("Try to find successor for %d on %s", n.ft.GetEntry(0).start, remote)
	succ, err := remote.findSuccessor(ctx, n.ft.GetEntry(0).start)
	if err != nil {
		return err
	}

	logger.Debugf("Successor node for %d is %s", n.ft.GetEntry(0).start, succ)
	n.ft.SetEntry(0, succ)

	predNode, err := succ.GetPredNode()
	if err != nil {
		return err
	}
	local.SetPredNode(predNode)
	logger.Debugf("Local node's predecessor set to %v", predNode)

	logger.Debug("recalc finger table")
	for i := 0; i < int(n.m)-1; i++ {
		logger.Debugf("i=%d", i)
		logger.Debugf("finger[i+1].start=%d", n.ft.GetEntry(i+1).start)
		logger.Debugf("interval=[%d, %d)", local.id, n.ft.GetEntry(i).node)

		if n.ft.GetEntry(i+1).start.In(local.id, n.ft.GetEntry(i).node, n.m) {
			logger.Debugf("interval=[%d, %d)", local.id, n.ft.GetEntry(i).node)
			n.ft.SetID(i+1, n.ft.GetEntry(i).node)
		} else {
			newSucc, err := remote.findSuccessor(ctx, n.ft.GetEntry(i+1).start)
			logger.Debugf("new successor for %d is %v", n.ft.GetEntry(i+1).start, newSucc)
			if err != nil {
				return err
			}
			n.ft.SetEntry(i+1, newSucc)
		}
	}
	return nil
}

func (n *LocalNode) join(ctx context.Context, introducerNode *RemoteNode) error {
	return n.WithSpan(ctx, "LocalNode.join", func(ctx context.Context) error {
		logger := logrus.WithField("method", "LocalNode.join")
		logger.Debugf("introducerNode: %s", introducerNode)

		if err := n.initFinger(ctx, introducerNode); err != nil {
			return errors.Wrap(err, "error while init'ing fingertable")
		}
		n.ft.Print(nil)
		if err := n.updateOthers(ctx); err != nil {
			return errors.Wrap(err, "error while updating other node's fingertables")
		}
		return nil
	})
}

func (n *LocalNode) updateOthers(ctx context.Context) error {
	logger := logrus.WithField("method", "LocalNode.updateOthers")
	newCtx, span := n.Start(ctx, "LocalNode.updateOthers")
	defer span.End()

	for i := 0; i < int(n.m); i++ {
		logger.Debugf("iteration: %d", i)
		newID := n.id.Sub(ChordID(pow2(uint32(i))), n.m)
		logger.Debugf(fmt.Sprintf("findPredecessor for ft[%d]=%d", i, newID))
		span.AddEvent(newCtx, fmt.Sprintf("findPredecessor for ft[%d]=%d", i, newID))
		p, err := n.findPredecessor(newCtx, newID)
		if err != nil {
			span.RecordError(newCtx, err)
			return err
		}
		logger.Debugf(fmt.Sprintf("found predecessor node: %v", p.GetID()))
		span.AddEvent(newCtx, fmt.Sprintf("found predecessor node: %v", p.GetID()))

		if err := p.updateFingerTable(ctx, n, i); err != nil {
			span.RecordError(newCtx, err)
			return err
		}
	}
	return nil
}

func (n *LocalNode) updateFingerTable(_ context.Context, s Node, i int) error {
	n.ft.SetEntry(i, s)
	return nil
}

func newLocalNode(id ChordID, bind string, m Rank) (*LocalNode, error) {
	localNode := &LocalNode{
		Tracer: global.Tracer(""),
		id:     id,
		bind:   bind,
		pred:   id,
		succ:   id,
		ft:     nil,
		m:      m,
	}
	ft := newFingerTable(localNode, m)
	localNode.ft = &ft
	return localNode, nil
}
