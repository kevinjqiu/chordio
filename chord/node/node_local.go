package node

import (
	"context"
	"fmt"
	"github.com/kevinjqiu/chordio/chord"
	"github.com/kevinjqiu/chordio/pb"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/api/core"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"
	"math/rand"
	"sync"
)

type localNode struct {
	trace.Tracer
	mu       *sync.Mutex
	id       chord.ID
	bind     string
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
	n.mu.Lock()
	defer n.mu.Unlock()
	n.predNode = pn
}

func (n *localNode) SetSuccNode(_ context.Context, sn NodeRef) {
	n.mu.Lock()
	defer n.mu.Unlock()
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
		n_  Node = n
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
	ctx, span := n.Start(ctx, "localNode.ClosestPrecedingFinger", trace.WithAttributes(core.Key("id").Int(id.AsInt())))
	defer span.End()
	// nb: int cast here is IMPORTANT!
	// because n.m is of type uint32
	// i >= 0 is always going to be true
	for i := int(n.m) - 1; i >= 0; i-- {
		fte := n.ft.GetEntry(i)
		interval := chord.NewInterval(n.m, n.id, id, chord.WithLeftOpen, chord.WithRightOpen)
		span.AddEvent(ctx, fmt.Sprintf("i=%d,interval=%s, fte=%s", i, interval, fte))
		if interval.Has(fte.Node.GetID()) {
			span.AddEvent(ctx, fmt.Sprintf("node %s is in the interval %s", fte.Node, interval))
			node, ok := n.ft.GetNodeByID(fte.Node.GetID())
			if !ok {
				err := errNodeNotFound(fte.Node.GetID())
				span.RecordError(ctx, err)
				return nil, errNodeNotFound(fte.Node.GetID())
			}

			span.AddEvent(ctx, fmt.Sprintf("ClosestPrecedingFinger is : %s", node.String()))
			if node.GetID() == n.id {
				return n.factory.newLocalNode(node.GetID(), node.GetBind(), n.m)
			} else {
				return n.factory.newRemoteNode(ctx, node.GetBind())
			}
		}
	}
	span.AddEvent(ctx, fmt.Sprintf("ClosestPrecedingFinger is the local node: %s", n.String()))
	return n, nil
}

func (n *localNode) Join(ctx context.Context, introducerNode RemoteNode) error {
	ctx, span := n.Start(ctx, "localNode.Join",
		trace.WithAttributes(core.Key("introducerNode").String(introducerNode.String())),
	)
	defer span.End()

	span.AddEvent(ctx, fmt.Sprintf("before updating FT: %s", n.ft.String()))

	n.SetPredNode(ctx, nil)
	succNode, err := introducerNode.FindSuccessor(ctx, n.GetID())
	if err != nil {
		span.RecordError(ctx, err)
		return errors.Wrap(err, "unable to join")
	}
	n.SetSuccNode(ctx, succNode)
	return nil
}

func (n *localNode) UpdateFingerTableEntry(ctx context.Context, s Node, i int) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	ctx, span := n.Start(ctx, "LocalNode.updateFingerTableEntry",
		trace.WithAttributes(core.Key("node").String(s.String()), core.Key("i").Int(i)),
	)
	defer span.End()

	// if s is the ith finger of n, then update n's finger table with s
	interval := chord.NewInterval(n.m, n.id, n.ft.GetEntry(i).Node.GetID())
	span.AddEvent(ctx, "interval.Has(s.GetID())",
		core.Key("interval").String(interval.String()),
		core.Key("id").Int(s.GetID().AsInt()),
	)
	if interval.Has(s.GetID()) {
		span.AddEvent(ctx, "s.GetID() in interval")
		if s.GetID() != n.id {
			span.AddEvent(ctx, fmt.Sprintf("s is not the local node, update FT[%d]", i))
			span.AddEvent(ctx, fmt.Sprintf("FTE before update: %s", n.ft.GetEntry(i).String()))
			n.mu.Lock()
			n.ft.SetNodeAtEntry(i, s)
			n.mu.Unlock()
			span.AddEvent(ctx, fmt.Sprintf("FTE after update: %s", n.ft.GetEntry(i).String()))
		}

		span.AddEvent(ctx, fmt.Sprintf("update the predecessor node's FTE[%d]", i))
		var (
			predNode Node
			err      error
		)
		if predNode, err = n.factory.newRemoteNode(ctx, n.GetPredNode().GetBind()); err != nil {
			span.RecordError(ctx, err)
			return err
		}

		span.AddEvent(ctx, fmt.Sprintf("predecessor node: %s", predNode.String()))
		if err := predNode.UpdateFingerTableEntry(ctx, s, i); err != nil {
			span.RecordError(ctx, err)
			return err
		}
	}
	return nil
}

func (n *localNode) Notify(ctx context.Context, n_ Node) error {
	int := chord.NewInterval(n.m, n.GetPredNode().GetID(), n.GetID(), chord.WithLeftClosed, chord.WithRightOpen)
	if n.GetPredNode() == nil || int.Has(n_.GetID()) {
		n.SetPredNode(ctx, n_)
	}
	logrus.Info("After notify(): %s", n.String())
	return nil
}

func (n *localNode) setNodeFactory(f factory) {
	n.factory = f
}

func (n *localNode) Stabilize(ctx context.Context) error {
	// TODO: do not use remote node if the node is local
	x, err := n.factory.newRemoteNode(ctx, n.GetSuccNode().GetBind())
	if err != nil {
		return err
	}

	int := chord.NewInterval(n.m, n.GetID(), n.GetSuccNode().GetID(), chord.WithLeftOpen, chord.WithRightOpen)
	if int.Has(x.GetID()) {
		n.SetSuccNode(ctx, x)
	}

	succNode, err := n.factory.newRemoteNode(ctx, n.GetSuccNode().GetBind())
	if err != nil {
		return err
	}

	if err := succNode.Notify(ctx, succNode); err != nil {
		return err
	}
	return nil
}

func (n *localNode) FixFingers(ctx context.Context) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	// i is the entry to fix
	i := rand.Int() % n.m.AsInt()

	succNode, err := n.FindSuccessor(ctx, n.GetFingerTable().GetEntry(i).Start)
	if err != nil {
		return err
	}

	n.GetFingerTable().SetNodeAtEntry(i, succNode)
	n.GetFingerTable().PrettyPrint(nil)
	return nil
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
