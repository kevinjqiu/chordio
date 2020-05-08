package node

import (
	"context"
	"fmt"
	"github.com/kevinjqiu/chordio/chord"
	"github.com/kevinjqiu/chordio/pb"
	"github.com/kevinjqiu/chordio/telemetry"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/api/core"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/plugin/grpctrace"
	"google.golang.org/grpc"
)

type remoteNode struct {
	trace.Tracer
	id       chord.ID
	bind     string
	predNode *pb.Node
	succNode *pb.Node
	client   pb.ChordClient

	newLocalNodeFn  localNodeConstructor
	newRemoteNodeFn remoteNodeConstructor
}

func (rn *remoteNode) setLocalNodeConstructor(fn localNodeConstructor) {
	rn.newLocalNodeFn = fn
}

func (rn *remoteNode) setRemoteNodeConstructor(fn remoteNodeConstructor) {
	rn.newRemoteNodeFn = fn
}

func (rn *remoteNode) String() string {
	var pred, succ string

	if rn.predNode == nil {
		pred = "<nil>"
	} else {
		pred = fmt.Sprintf("%d@%s", rn.predNode.GetId(), rn.predNode.GetBind())
	}

	if rn.succNode == nil {
		succ = "<nil>"
	} else {
		succ = fmt.Sprintf("%d@%s", rn.succNode.GetId(), rn.succNode.GetBind())
	}

	return fmt.Sprintf("<R: %d@%s, p=%s, s=%s>", rn.id, rn.bind, pred, succ)
}

func (rn *remoteNode) GetID() chord.ID {
	return rn.id
}

func (rn *remoteNode) GetBind() string {
	return rn.bind
}

func (rn *remoteNode) GetPredNode() (*NodeRef, error) {
	return &NodeRef{chord.ID(rn.predNode.Id), rn.predNode.Bind}, nil
}

func (rn *remoteNode) GetSuccNode() (*NodeRef, error) {
	return &NodeRef{chord.ID(rn.succNode.Id), rn.succNode.Bind}, nil
}

func (rn *remoteNode) FindPredecessor(ctx context.Context, id chord.ID) (Node, error) {
	var n RemoteNode
	err := rn.WithSpan(ctx, "remoteNode.FindPredecessor", func(ctx context.Context) error {
		logrus.Debug("[remoteNode] FindPredecessor: ", id)
		req := pb.FindPredecessorRequest{
			Id: uint64(id),
		}

		resp, err := rn.client.FindPredecessor(ctx, &req)
		if err != nil {
			return err
		}

		n, err = rn.newRemoteNodeFn(ctx, resp.Node.Bind)
		return err
	})
	return n, err
}

func (rn *remoteNode) FindSuccessor(ctx context.Context, id chord.ID) (Node, error) {
	var n RemoteNode
	err := rn.WithSpan(ctx, "remoteNode.FindSuccessor", func(ctx context.Context) error {
		logrus.Debug("[remoteNode] FindSuccessor: ", id)
		req := pb.FindSuccessorRequest{
			Id: uint64(id),
		}

		resp, err := rn.client.FindSuccessor(ctx, &req)
		if err != nil {
			return err
		}

		n, err = rn.newRemoteNodeFn(ctx, resp.Node.Bind)
		return err
	})

	return n, err
}

func (rn *remoteNode) ClosestPrecedingFinger(ctx context.Context, id chord.ID) (Node, error) {
	var n RemoteNode

	err := rn.WithSpan(ctx, "remoteNode.ClosestPrecedingFinger", func(ctx context.Context) error {
		logrus.Debug("[remoteNode] ClosestPrecedingFinger: ", id)
		req := pb.ClosestPrecedingFingerRequest{
			Id: uint64(id),
		}

		resp, err := rn.client.ClosestPrecedingFinger(ctx, &req)
		if err != nil {
			return err
		}

		n, err = rn.newRemoteNodeFn(ctx, resp.Node.Bind)
		return err
	})

	return n, err
}

func (rn *remoteNode) AsProtobufNode() *pb.Node {
	pbn := &pb.Node{
		Id:   uint64(rn.GetID()),
		Bind: rn.GetBind(),
		Pred: nil,
		Succ: nil,
	}

	pred, err := rn.GetPredNode()
	if err != nil {
		pbn.Pred = &pb.Node{
			Id:   uint64(pred.ID),
			Bind: pred.Bind,
		}
	}

	succ, err := rn.GetSuccNode()
	if err != nil {
		pbn.Succ = &pb.Node{
			Id:   uint64(succ.ID),
			Bind: succ.Bind,
		}
	}

	return pbn
}

func (rn *remoteNode) UpdateFingerTableEntry(ctx context.Context, s Node, i int) error {
	ctx, span := rn.Start(ctx, "remoteNode.UpdateFingerTableEntry",
		trace.WithAttributes(core.Key("s").String(s.String())))
	defer span.End()

	logrus.Debugf("[remoteNode] UpdateFingerTableEntry: ID=%d, i=%d", s.GetID(), i)
	req := pb.UpdateFingerTableRequest{
		Node: s.AsProtobufNode(),
		I:    int64(i),
	}

	_, err := rn.client.UpdateFingerTable(ctx, &req)
	return err
}

func NewRemote(ctx context.Context, bind string, opts ...nodeConstructorOption) (RemoteNode, error) {
	conn, err := grpc.Dial(bind,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(grpctrace.UnaryClientInterceptor(global.Tracer(telemetry.GetServiceName()))),
		grpc.WithStreamInterceptor(grpctrace.StreamClientInterceptor(global.Tracer(telemetry.GetServiceName()))),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to initiate grpc client for node: %v", bind)
	}

	client := pb.NewChordClient(conn)
	resp, err := client.GetNodeInfo(ctx, &pb.GetNodeInfoRequest{})
	if err != nil {
		return nil, err
	}

	rn := &remoteNode{
		Tracer:   global.Tracer(""),
		id:       chord.ID(resp.Node.GetId()),
		bind:     bind,
		predNode: resp.Node.GetPred(),
		succNode: resp.Node.GetSucc(),
		client:   client,

		newLocalNodeFn:  NewLocal,
		newRemoteNodeFn: NewRemote,
	}

	for _, opt := range opts {
		opt.apply(rn)
	}
	return rn, nil
}
