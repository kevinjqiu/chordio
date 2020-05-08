package node

import (
	"context"
	"fmt"
	"github.com/kevinjqiu/chordio/chord"
	"github.com/kevinjqiu/chordio/pb"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"
	"sync"
)

type localNode struct {
	trace.Tracer
	mu       *sync.Mutex
	id       chord.ID
	bind     string
	pred     chord.ID
	succ     chord.ID
	predNode *nodeRef
	succNode *nodeRef
	m        chord.Rank
	ft       *FingerTable

	factory  factory
}

func (n *localNode) GetFingerTable() *FingerTable {
	return n.ft
}

func (n *localNode) GetRank() chord.Rank {
	return n.m
}

func (n *localNode) String() string {
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

func (n *localNode) SetPredNode(pn *nodeRef) {
	n.predNode = pn
	n.pred = pn.ID
}

func (n *localNode) SetSuccNode(sn *nodeRef) {
	n.succNode = sn
	n.succ = sn.ID
}

func (n *localNode) GetID() chord.ID {
	return n.id
}

func (n *localNode) GetBind() string {
	return n.bind
}

func (n *localNode) AsProtobufNode() *pb.Node {
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

func (n *localNode) GetPredNode() (*nodeRef, error) {
	if n.predNode != nil && n.predNode.ID != n.pred {
		return n.predNode, nil
	}
	node, ok := n.ft.GetNodeByID(n.pred)
	if !ok {
		return nil, fmt.Errorf("predecessor node %v not found in neighbourhood", n.pred)
	}
	return &nodeRef{Bind: node.GetBind(), ID: node.GetID()}, nil
}

func (n *localNode) GetSuccNode() (*nodeRef, error) {
	if n.succNode != nil && n.succNode.ID != n.succ {
		return n.succNode, nil
	}
	node, ok := n.ft.GetNodeByID(n.succ)
	if !ok {
		return nil, fmt.Errorf("successor node %v not found in neighbourhood", n.pred)
	}
	return &nodeRef{Bind: node.GetBind(), ID: node.GetID()}, nil
}

// TODO: if the node is itself, do not return a remoteNode version of it
func (n *localNode) FindPredecessor(ctx context.Context, id chord.ID) (Node, error) {
	ctx, span := n.Start(ctx, "localNode.FindPredecessor")
	defer span.End()

	logger := logrus.WithField("method", "localNode.FindPredecessor")
	var (
		remoteNode Node
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
	remoteNode, err = n.factory.newRemoteNode(ctx, n_.GetBind())
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

func (n *localNode) FindSuccessor(ctx context.Context, id chord.ID) (Node, error) {
	ctx, span := n.Start(ctx, "localNode.FindSuccessor")
	defer span.End()

	predNode, err := n.FindPredecessor(ctx, id)
	if err != nil {
		return nil, err
	}

	succNode, err := predNode.GetSuccNode()
	if err != nil {
		return nil, err
	}

	return n.factory.newRemoteNode(ctx, succNode.Bind)
}

func (n *localNode) ClosestPrecedingFinger(ctx context.Context, id chord.ID) (Node, error) {
	ctx, span := n.Start(ctx, "localNode.ClosestPrecedingFinger")
	defer span.End()

	logger := logrus.WithField("method", "localNode.ClosestPrecedingFinger")
	logger.Debugf("ID=%d", id)
	// nb: int cast here is IMPORTANT!
	// because n.m is of type uint32
	// i >= 0 is always going to be true
	for i := int(n.m) - 1; i >= 0; i-- {
		fte := n.ft.GetEntry(i)
		interval := chord.NewInterval(n.m, n.id, id, chord.WithLeftOpen, chord.WithRightOpen)
		if interval.Has(fte.Node.GetID()) {
			node, ok := n.ft.GetNodeByID(fte.Node.GetID())
			if !ok {
				return nil, fmt.Errorf("node %s at fte[%d] not found", fte.Node, i)
			}

			if node.GetID() == n.id {
				return n.factory.newLocalNode(node.GetID(), node.GetBind(), n.m)
			} else {
				return n.factory.newRemoteNode(ctx, node.GetBind())
			}
		}
	}
	return n, nil
}

// initialize finger table of the local node
// n' is an arbitrary node already in the network
func (n *localNode) initFinger(ctx context.Context, remote RemoteNode) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	ctx, span := n.Start(ctx, "localNode.initFinger")
	defer span.End()

	logger := logrus.WithField("method", "localNode.initFinger")
	logger.Infof("using remote node %s", remote)
	local := n
	logger.Debugf("Try to find successor for %d on %s", n.ft.GetEntry(0).Start, remote)
	succ, err := remote.FindSuccessor(ctx, n.ft.GetEntry(0).Start)
	if err != nil {
		return err
	}

	logger.Debugf("Successor node for %d is %s", n.ft.GetEntry(0).Start, succ)
	n.ft.SetNodeAtEntry(0, succ)

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
		logger.Debugf("interval=[%d, %d)", local.id, n.ft.GetEntry(i).Node.GetID())

		if n.ft.GetEntry(i+1).Start.In(local.id, n.ft.GetEntry(i).Node.GetID(), n.m) {
			logger.Debugf("interval=[%d, %d)", local.id, n.ft.GetEntry(i).Node.GetID())
			n.ft.ReplaceNodeWithAnotherEntry(i+1, i)
		} else {
			newSucc, err := remote.FindSuccessor(ctx, n.ft.GetEntry(i+1).Start)
			logger.Debugf("new successor for %d is %v", n.ft.GetEntry(i+1).Start, newSucc)
			if err != nil {
				return err
			}
			n.ft.SetNodeAtEntry(i+1, newSucc)
		}
	}
	return nil
}

func (n *localNode) Join(ctx context.Context, introducerNode RemoteNode) error {
	ctx, span := n.Start(ctx, "localNode.Join")
	defer span.End()

	logger := logrus.WithField("method", "localNode.Join")
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

func (n *localNode) updateOthers(ctx context.Context) error {
	newCtx, span := n.Start(ctx, "localNode.updateOthers")
	defer span.End()

	logger := logrus.WithField("method", "localNode.updateOthers")
	for i := 0; i < int(n.m); i++ {
		logger.Debugf("iteration: %d", i)
		newID := n.id.Sub(chord.ID(2).Pow(i), n.m)
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

func (n *localNode) UpdateFingerTableEntry(_ context.Context, s Node, i int) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.ft.SetNodeAtEntry(i, s)
	return nil
}

func (n *localNode) setNodeFactory(f factory) {
	n.factory = f
}

func NewLocal(id chord.ID, bind string, m chord.Rank) (LocalNode, error) {
	localNode := &localNode{
		Tracer: global.Tracer(""),
		mu:     new(sync.Mutex),
		id:     id,
		bind:   bind,
		pred:   id,
		succ:   id,
		ft:     nil,
		m:      m,

		factory: defaultFactory{},
	}
	ft := newFingerTable(localNode, m)
	localNode.ft = &ft

	return localNode, nil
}
