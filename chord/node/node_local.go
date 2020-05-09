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
	mu   *sync.Mutex
	id   chord.ID
	bind string
	predNode NodeRef
	m        chord.Rank
	ft       *FingerTable

	factory factory
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
		pred = n.predNode.String()
	}

	if n.GetSuccNode() == nil {
		succ = "<nil>"
	} else {
		succ = n.GetSuccNode().String()
	}
	return fmt.Sprintf("<L: %d@%s, p=%s, s=%s>", n.id, n.bind, pred, succ)
}

func (n *localNode) SetPredNode(_ context.Context, pn NodeRef) {
	n.predNode = pn
}

func (n *localNode) SetSuccNode(_ context.Context, sn NodeRef) {
	n.ft.SetNodeAtEntry(0, sn)
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

	predNode := n.GetPredNode()
	if predNode != nil {
		pbn.Pred = &pb.Node{
			Id:   uint64(predNode.GetID()),
			Bind: predNode.GetBind(),
		}
	}

	succNode := n.GetSuccNode()
	if succNode != nil {
		pbn.Succ = &pb.Node{
			Id:   uint64(succNode.GetID()),
			Bind: succNode.GetBind(),
		}
	}
	return pbn
}

func (n *localNode) GetPredNode() NodeRef {
	return n.predNode
}

func (n *localNode) GetSuccNode() NodeRef {
	return n.ft.GetEntry(0).Node
}

func (n *localNode) FindPredecessor(ctx context.Context, id chord.ID) (Node, error) {
	ctx, span := n.Start(ctx, "localNode.FindPredecessor")
	defer span.End()

	var (
		n_ Node = n
		err error
	)
	for {
		interval := chord.NewInterval(n.m, n_.GetID(), n_.GetSuccNode().GetID(), chord.WithLeftOpen, chord.WithRightClosed)
		if !interval.Has(id) {
			if n_, err = n_.ClosestPrecedingFinger(ctx, id); err != nil {
				return nil, err
			}
		} else {
			break
		}
	}

	return n_, nil
}

func (n *localNode) FindSuccessor(ctx context.Context, id chord.ID) (Node, error) {
	ctx, span := n.Start(ctx, "localNode.FindSuccessor")
	defer span.End()

	predNode, err := n.FindPredecessor(ctx, id)
	if err != nil {
		return nil, err
	}

	succNode := predNode.GetSuccNode()
	if succNode == nil {
		return nil, errNoSuccessorNode
	}

	if succNode.GetID() == n.id {
		return n, nil
	}

	return n.factory.newRemoteNode(ctx, succNode.GetBind())
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
				return nil, errNodeNotFound(fte.Node.GetID())
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

	succ, err := remote.FindSuccessor(ctx, n.ft.GetEntry(0).Start)
	if err != nil {
		return err
	}
	n.ft.SetNodeAtEntry(0, succ)
	n.SetPredNode(ctx, succ.GetPredNode())
	succ.SetPredNode(ctx, n)

	for i := 0; i < n.m.AsInt() - 1; i++ {
		interval := chord.NewInterval(n.m, n.id, n.ft.GetEntry(i).Node.GetID())
		if interval.Has(n.ft.GetEntry(i+1).Start) {
			n.ft.ReplaceNodeWithAnotherEntry(i+1, i)
		} else {
			newSucc, err := remote.FindSuccessor(ctx, n.ft.GetEntry(i+1).Start)
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
	fmt.Println("After initFinger(n)")
	n.ft.Print(nil)

	if err := n.updateOthers(ctx); err != nil {
		return errors.Wrap(err, "error while updating other node's fingertables")
	}
	fmt.Println("After updateOthers()")
	n.ft.Print(nil)
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

func (n *localNode) UpdateFingerTableEntry(ctx context.Context, s Node, i int) error {
	// if s is the ith finger of n, then update n's finger table with s
	interval := chord.NewInterval(n.m, n.id, n.ft.GetEntry(i).Node.GetID())
	if interval.Has(s.GetID()) {
		if s.GetID() != n.id {
			n.mu.Lock()
			n.ft.SetNodeAtEntry(i, s)
			n.mu.Unlock()
		}

		var (
			predNode Node
			err error
		)
		//if n.GetPredNode().GetID() == n.id {
		//	predNode, err = n.factory.newLocalNode(n.id, n.bind, n.m)
		//} else {
		//	predNode, err = n.factory.newRemoteNode(ctx, n.GetPredNode().GetBind())
		//}
		//if err != nil {
		//	return err
		//}
		predNode, err = n.factory.newRemoteNode(ctx, n.GetPredNode().GetBind())
		if err != nil {
			return err
		}
		return predNode.UpdateFingerTableEntry(ctx, s, i)
	}
	return nil
}

func (n *localNode) setNodeFactory(f factory) {
	n.factory = f
}

func NewLocal(id chord.ID, bind string, m chord.Rank) (LocalNode, error) {
	localNodeRef := &nodeRef{
		ID: id, Bind: bind,
	}
	localNode := &localNode{
		Tracer:   global.Tracer(""),
		mu:       new(sync.Mutex),
		id:       id,
		bind:     bind,
		predNode: localNodeRef,
		ft:       nil,
		m:        m,

		factory: defaultFactory{},
	}
	localNode.ft = newFingerTable(localNode, m)
	return localNode, nil
}
