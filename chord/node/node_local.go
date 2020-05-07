package node

import (
	"context"
	"fmt"
	"github.com/kevinjqiu/chordio/chord"
	ft2 "github.com/kevinjqiu/chordio/chord/ft"
	"github.com/kevinjqiu/chordio/pb"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"
)

type LocalNode struct {
	trace.Tracer
	id       chord.ID
	bind     string
	pred     chord.ID
	succ     chord.ID
	predNode *chord.NodeRef
	succNode *chord.NodeRef
	m        chord.Rank
	ft       *ft2.FingerTable
}

func (n *LocalNode) GetFingerTable() *ft2.FingerTable {
	return n.ft
}

func (n *LocalNode) GetM() chord.Rank {
	return n.m
}

func (n *LocalNode) String() string {
	var pred, succ string

	if n.predNode == nil {
		pred = "<nil>"
	} else {
		pred = fmt.Sprintf("%d@%s", n.predNode.ID, n.predNode.Bind)
	}

	if n.succNode == nil {
		succ = "<nil>"
	} else {
		succ = fmt.Sprintf("%d@%s", n.succNode.ID, n.succNode.Bind)
	}
	return fmt.Sprintf("<L: %d@%s, p=%s, s=%s>", n.id, n.bind, pred, succ)
}

func (n *LocalNode) SetPredNode(pn *chord.NodeRef) {
	n.predNode = pn
	n.pred = pn.ID
}

func (n *LocalNode) SetSuccNode(sn *chord.NodeRef) {
	n.succNode = sn
	n.succ = sn.ID
}

func (n *LocalNode) GetID() chord.ID {
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
			Id:   uint64(predNode.ID),
			Bind: predNode.Bind,
		}
	}

	succNode, err := n.GetSuccNode()
	if err == nil {
		pbn.Succ = &pb.Node{
			Id:   uint64(succNode.ID),
			Bind: succNode.Bind,
		}
	}
	return pbn
}

func (n *LocalNode) GetPredNode() (*chord.NodeRef, error) {
	if n.predNode != nil && n.predNode.ID != n.pred {
		return n.predNode, nil
	}
	node, _, _, ok := n.ft.GetNodeByID(n.pred)
	if !ok {
		return nil, fmt.Errorf("predecessor node %v not found in neighbourhood", n.pred)
	}
	return &chord.NodeRef{Bind: node.Bind, ID: node.ID}, nil
}

func (n *LocalNode) GetSuccNode() (*chord.NodeRef, error) {
	if n.succNode != nil && n.succNode.ID != n.succ {
		return n.succNode, nil
	}
	node, _, _, ok := n.ft.GetNodeByID(n.succ)
	if !ok {
		return nil, fmt.Errorf("successor node %v not found in neighbourhood", n.pred)
	}
	return &chord.NodeRef{Bind: node.Bind, ID: node.ID}, nil
}

// TODO: if the node is itself, do not return a RemoteNode version of it
func (n *LocalNode) FindPredecessor(ctx context.Context, id chord.ID) (chord.Node, error) {
	ctx, span := n.Start(ctx, "LocalNode.FindPredecessor")
	defer span.End()

	logger := logrus.WithField("method", "LocalNode.FindPredecessor")
	var (
		remoteNode chord.Node
		err        error
	)

	if !id.In(n.id, n.succ, n.m) {
		logger.Debugf("ID is within %v, the predecessor is the local node", n)
		return n, nil
	}

	n_, err := n.ClosestPrecedingFinger(ctx, id)
	if err != nil {
		return nil, err
	}

	logger.Debugf("the closest preceding node is %v", n_)
	remoteNode, err = NewRemote(ctx, n_.GetBind())
	if err != nil {
		return nil, err
	}

	for {
		succNode, err := n_.GetSuccNode()
		if err != nil {
			return nil, err
		}
		interval := chord.NewInterval(n.m, n_.GetID(), succNode.ID, chord.WithLeftOpen, chord.WithRightClosed)
		if !interval.Has(id) {
			logger.Debugf("ID is not in %v's range", n_)
			remoteNode, err = remoteNode.ClosestPrecedingFinger(ctx, id)
			if err != nil {
				return nil, err
			}
			logger.Debugf("the closest preceding node in %s's finger table is: ", remoteNode)
		} else {
			logger.Debugf("ID is in %v's range", n_)
			break
		}
	}

	return remoteNode, nil

}

func (n *LocalNode) FindSuccessor(ctx context.Context, id chord.ID) (chord.Node, error) {
	ctx, span := n.Start(ctx, "LocalNode.FindSuccessor")
	defer span.End()

	predNode, err := n.FindPredecessor(ctx, id)
	if err != nil {
		return nil, err
	}

	succNode, err := predNode.GetSuccNode()
	if err != nil {
		return nil, err
	}

	return NewRemote(ctx, succNode.Bind)
}

func (n *LocalNode) ClosestPrecedingFinger(ctx context.Context, id chord.ID) (chord.Node, error) {
	ctx, span := n.Start(ctx, "LocalNode.ClosestPrecedingFinger")
	defer span.End()

	logger := logrus.WithField("method", "LocalNode.ClosestPrecedingFinger")
	logger.Debugf("ID=%d", id)
	// nb: int cast here is IMPORTANT!
	// because n.m is of type uint32
	// i >= 0 is always going to be true
	for i := int(n.m) - 1; i >= 0; i-- {
		fte := n.ft.GetEntry(i)
		interval := chord.NewInterval(n.m, n.id, id, chord.WithLeftOpen, chord.WithRightOpen)
		if interval.Has(fte.NodeID) {
			node, _, _, ok := n.ft.GetNodeByID(fte.NodeID)
			if !ok {
				return nil, fmt.Errorf("node %d at fte[%d] not found", fte.NodeID, i)
			}

			if node.ID == n.id {
				return NewLocal(node.ID, node.Bind, n.m)
			} else {
				return NewRemote(ctx, node.Bind)
			}
		}
	}
	return n, nil
}

// initialize finger table of the local node
// n' is an arbitrary node already in the network
func (n *LocalNode) initFinger(ctx context.Context, remote *RemoteNode) error {
	ctx, span := n.Start(ctx, "LocalNode.initFinger")
	defer span.End()

	logger := logrus.WithField("method", "LocalNode.initFinger")
	logger.Infof("using remote node %s", remote)
	local := n
	logger.Debugf("Try to find successor for %d on %s", n.ft.GetEntry(0).Start, remote)
	succ, err := remote.FindSuccessor(ctx, n.ft.GetEntry(0).Start)
	if err != nil {
		return err
	}

	logger.Debugf("Successor node for %d is %s", n.ft.GetEntry(0).Start, succ)
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
		logger.Debugf("finger[i+1].start=%d", n.ft.GetEntry(i+1).Start)
		logger.Debugf("interval=[%d, %d)", local.id, n.ft.GetEntry(i).NodeID)

		if n.ft.GetEntry(i+1).Start.In(local.id, n.ft.GetEntry(i).NodeID, n.m) {
			logger.Debugf("interval=[%d, %d)", local.id, n.ft.GetEntry(i).NodeID)
			_ = n.ft.SetID(i+1, n.ft.GetEntry(i).NodeID)  // TODO: handle error
		} else {
			newSucc, err := remote.FindSuccessor(ctx, n.ft.GetEntry(i+1).Start)
			logger.Debugf("new successor for %d is %v", n.ft.GetEntry(i+1).Start, newSucc)
			if err != nil {
				return err
			}
			n.ft.SetEntry(i+1, newSucc)
		}
	}
	return nil
}

func (n *LocalNode) Join(ctx context.Context, introducerNode *RemoteNode) error {
	ctx, span := n.Start(ctx, "LocalNode.Join")
	defer span.End()

	logger := logrus.WithField("method", "LocalNode.Join")
	logger.Debugf("introducerNode: %s", introducerNode)

	if err := n.initFinger(ctx, introducerNode); err != nil {
		return errors.Wrap(err, "error while init'ing fingertable")
	}
	n.ft.Print(nil)
	if err := n.updateOthers(ctx); err != nil {
		return errors.Wrap(err, "error while updating other node's fingertables")
	}
	return nil
}

func (n *LocalNode) updateOthers(ctx context.Context) error {
	newCtx, span := n.Start(ctx, "LocalNode.updateOthers")
	defer span.End()

	logger := logrus.WithField("method", "LocalNode.updateOthers")
	for i := 0; i < int(n.m); i++ {
		logger.Debugf("iteration: %d", i)
		//newID := n.ID.Sub(chord.ID(chord.pow2(uint32(i))), n.m)
		newID := n.id.Sub(chord.ID(chord.ID(2).Pow(i)), n.m)
		logger.Debugf(fmt.Sprintf("FindPredecessor for ft[%d]=%d", i, newID))
		span.AddEvent(newCtx, fmt.Sprintf("FindPredecessor for ft[%d]=%d", i, newID))
		p, err := n.FindPredecessor(newCtx, newID)
		if err != nil {
			span.RecordError(newCtx, err)
			return err
		}
		logger.Debugf(fmt.Sprintf("found predecessor node: %v", p.GetID()))
		span.AddEvent(newCtx, fmt.Sprintf("found predecessor node: %v", p.GetID()))

		if err := p.UpdateFingerTableEntry(ctx, n, i); err != nil {
			span.RecordError(newCtx, err)
			return err
		}
	}
	return nil
}

func (n *LocalNode) UpdateFingerTableEntry(_ context.Context, s chord.Node, i int) error {
	n.ft.SetEntry(i, s)
	return nil
}

func NewLocal(id chord.ID, bind string, m chord.Rank) (*LocalNode, error) {
	localNode := &LocalNode{
		Tracer: global.Tracer(""),
		id:     id,
		bind:   bind,
		pred:   id,
		succ:   id,
		ft:     nil,
		m:      m,
	}
	ft := ft2.New(localNode, m)
	localNode.ft = &ft
	return localNode, nil
}
