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

	ctx, span := n.Start(ctx, "localNode.initFinger",
		trace.WithAttributes(core.Key("remoteNode").String(remote.String())))
	defer span.End()

	span.AddEvent(ctx, "find successor", core.Key("fte").Int(0), core.Key("id").Int(n.ft.GetEntry(0).Start.AsInt()))
	succ, err := remote.FindSuccessor(ctx, n.ft.GetEntry(0).Start)
	if err != nil {
		span.RecordError(ctx, err)
		return errors.Wrap(err, "error finding successor")
	}
	span.AddEvent(ctx, "found successor node", core.Key("node").String(succ.String()))
	n.ft.SetNodeAtEntry(0, succ)
	span.AddEvent(ctx, "set predecessor", core.Key("pred").String(succ.GetPredNode().String()))
	n.SetPredNode(ctx, succ.GetPredNode())
	span.AddEvent(ctx, "set predecessor for pred's successor",
		core.Key("succ").String(succ.String()),
		core.Key("pred").String(n.String()),
	)
	succ.SetPredNode(ctx, n)

	span.AddEvent(ctx, "FT before update", core.Key("fte").String(n.ft.String()))
	for i := 0; i < n.m.AsInt()-1; i++ {
		interval := chord.NewInterval(n.m, n.id, n.ft.GetEntry(i).Node.GetID())
		attrs := []core.KeyValue{
			core.Key("i").Int(i),
			core.Key("interval").String(interval.String()),
		}
		_ = n.WithSpan(ctx, "localNode.initFinger#updateFingerTable", func(ctx context.Context) error {
			if interval.Has(n.ft.GetEntry(i + 1).Start) {
				n.ft.ReplaceNodeWithAnotherEntry(i+1, i)
			} else {
				newSucc, err := remote.FindSuccessor(ctx, n.ft.GetEntry(i+1).Start)
				if err != nil {
					return err
				}
				n.ft.SetNodeAtEntry(i+1, newSucc)
			}
			return nil
		}, trace.WithAttributes(attrs...))
	}
	span.AddEvent(ctx, "FT after update", core.Key("fte").String(n.ft.String()))

	return nil
}

func (n *localNode) Join(ctx context.Context, introducerNode RemoteNode) error {
	ctx, span := n.Start(ctx, "localNode.Join",
		trace.WithAttributes(core.Key("introducerNode").String(introducerNode.String())),
	)
	defer span.End()

	span.AddEvent(ctx, "before updating FT", core.Key("ft").String(n.ft.String()))
	if err := n.initFinger(ctx, introducerNode); err != nil {
		span.RecordError(ctx, err)
		return errors.Wrap(err, "error while init'ing fingertable")
	}
	span.AddEvent(ctx, "local node FT updated", core.Key("ft").String(n.ft.String()))
	n.ft.PrettyPrint(nil)

	span.AddEvent(ctx, "before updateOthers")
	if err := n.updateOthers(ctx); err != nil {
		span.RecordError(ctx, err)
		return errors.Wrap(err, "error while updating other node's fingertables")
	}
	span.AddEvent(ctx, "after updateOthers")
	n.ft.PrettyPrint(nil)
	return nil
}

func (n *localNode) updateOthers(ctx context.Context) error {
	ctx, span := n.Start(ctx, "localNode.updateOthers")
	defer span.End()

	logger := logrus.WithField("method", "localNode.updateOthers")
	for i := 0; i < int(n.m); i++ {
		logger.Debugf("iteration: %d", i)
		newID := n.id.Sub(chord.ID(2).Pow(i), n.m)
		logger.Debugf(fmt.Sprintf("FindPredecessor for ft[%d]=%d", i, newID))
		span.AddEvent(ctx, fmt.Sprintf("FindPredecessor for ft[%d]=%d", i, newID))
		p, err := n.FindPredecessor(ctx, newID)
		if err != nil {
			span.RecordError(ctx, err)
			return err
		}
		logger.Debugf(fmt.Sprintf("found predecessor node: %v", p.GetID()))
		span.AddEvent(ctx, fmt.Sprintf("found predecessor node: %v", p.GetID()))

		if err := p.UpdateFingerTableEntry(ctx, n, i); err != nil {
			span.RecordError(ctx, err)
			return err
		}
	}
	return nil
}

func (n *localNode) UpdateFingerTableEntry(ctx context.Context, s Node, i int) error {
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
			span.AddEvent(ctx, "s is not the local node, update FTE", core.Key("i").Int(i))
			span.AddEvent(ctx, "FTE before update", core.Key("entry").String(n.ft.GetEntry(i).String()))
			n.mu.Lock()
			n.ft.SetNodeAtEntry(i, s)
			n.mu.Unlock()
			span.AddEvent(ctx, "FTE after update", core.Key("entry").String(n.ft.GetEntry(i).String()))
		}

		span.AddEvent(ctx, "update the predecessor node's FTE at i", core.Key("i").Int(i))
		var (
			predNode Node
			err      error
		)
		predNode, err = n.factory.newRemoteNode(ctx, n.GetPredNode().GetBind())
		span.AddEvent(ctx, "predecessor node", core.Key("pred").String(predNode.String()))
		if err != nil {
			span.RecordError(ctx, err)
			return err
		}
		if err := predNode.UpdateFingerTableEntry(ctx, s, i); err != nil {
			span.RecordError(ctx, err)
			return err
		}
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
