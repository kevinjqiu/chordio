package node

import (
	"context"
	"fmt"
	"github.com/kevinjqiu/chordio/attrs"
	"github.com/kevinjqiu/chordio/chord"
	"github.com/kevinjqiu/chordio/pb"
	"github.com/kevinjqiu/chordio/telemetry"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/plugin/grpctrace"
	"google.golang.org/grpc"
)

type closeFunc func() error

// remoteNode is a proxy that implements the Node interface but
// delegate the calls to the remote node via grpc
type remoteNode struct {
	trace.Tracer
	id       chord.ID
	bind     string
	predNode *pb.Node
	succNode *pb.Node
}

func (rn *remoteNode) getClient() (pb.ChordClient, closeFunc, error) {
	conn, err := grpc.Dial(rn.bind,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(grpctrace.UnaryClientInterceptor(global.Tracer(telemetry.GetServiceName()))),
		grpc.WithStreamInterceptor(grpctrace.StreamClientInterceptor(global.Tracer(telemetry.GetServiceName()))),
	)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "unable to initiate grpc client for node: %v", rn.bind)
	}

	client := pb.NewChordClient(conn)
	return client, func() error { return conn.Close() }, nil
}

func (rn *remoteNode) SetPredNode(ctx context.Context, n chord.NodeRef) error {
	client, close, err := rn.getClient()
	if err != nil {
		return err
	}
	defer close()

	_, err = client.SetPredecessorNode(ctx, &pb.SetPredecessorNodeRequest{
		Node: &pb.Node{
			Id:   n.GetID().AsU64(),
			Bind: n.GetBind(),
		},
	})
	return err
}

func (rn *remoteNode) SetSuccNode(ctx context.Context, n chord.NodeRef) error {
	client, close, err := rn.getClient()
	if err != nil {
		return err
	}
	defer close()

	_, err = client.SetSuccessorNode(ctx, &pb.SetSuccessorNodeRequest{
		Node: &pb.Node{
			Id:   n.GetID().AsU64(),
			Bind: n.GetBind(),
		},
	})
	return err
}

func (rn *remoteNode) String() string {
	return fmt.Sprintf("<R %d@%s>", rn.id, rn.bind)
}

func (rn *remoteNode) GetID() chord.ID {
	return rn.id
}

func (rn *remoteNode) GetBind() string {
	return rn.bind
}

func (rn *remoteNode) GetPredNode() chord.NodeRef {
	return &nodeRef{chord.ID(rn.predNode.Id), rn.predNode.Bind}
}

func (rn *remoteNode) GetSuccNode() chord.NodeRef {
	return &nodeRef{chord.ID(rn.succNode.Id), rn.succNode.Bind}
}

func (rn *remoteNode) FindPredecessor(ctx context.Context, id chord.ID) (chord.Node, error) {
	ctx, span := rn.Start(ctx, "remoteNode.FindPredecessor", trace.WithAttributes(attrs.ID("id", id)))
	defer span.End()
	req := pb.FindPredecessorRequest{
		Id: uint64(id),
	}

	client, close, err := rn.getClient()
	if err != nil {
		return nil, err
	}
	defer close()

	resp, err := client.FindPredecessor(ctx, &req)
	if err != nil {
		return nil, err
	}

	return NewRemote(ctx, resp.Node.Bind)
}

func (rn *remoteNode) FindSuccessor(ctx context.Context, id chord.ID) (chord.Node, error) {
	ctx, span := rn.Start(ctx, "remoteNode.FindSuccessor", trace.WithAttributes(attrs.ID("id", id)))
	defer span.End()

	req := pb.FindSuccessorRequest{
		Id: uint64(id),
	}

	client, close, err := rn.getClient()
	if err != nil {
		return nil, err
	}
	defer close()

	resp, err := client.FindSuccessor(ctx, &req)
	if err != nil {
		return nil, err
	}

	return NewRemote(ctx, resp.Node.Bind)
}

func (rn *remoteNode) ClosestPrecedingFinger(ctx context.Context, id chord.ID) (chord.Node, error) {
	ctx, span := rn.Start(ctx, "remoteNode.ClosestPrecedingFinger", trace.WithAttributes(attrs.ID("id", id)))
	defer span.End()

	client, close, err := rn.getClient()
	if err != nil {
		return nil, err
	}
	defer close()

	req := pb.ClosestPrecedingFingerRequest{
		Id: uint64(id),
	}

	resp, err := client.ClosestPrecedingFinger(ctx, &req)
	if err != nil {
		return nil, err
	}

	return NewRemote(ctx, resp.Node.Bind)
}

func (rn *remoteNode) AsProtobufNode() *pb.Node {
	pbn := &pb.Node{
		Id:   uint64(rn.GetID()),
		Bind: rn.GetBind(),
		Pred: nil,
		Succ: nil,
	}

	pred := rn.GetPredNode()
	if pred != nil {
		pbn.Pred = &pb.Node{
			Id:   uint64(pred.GetID()),
			Bind: pred.GetBind(),
		}
	}

	succ := rn.GetSuccNode()
	if succ != nil {
		pbn.Succ = &pb.Node{
			Id:   uint64(succ.GetID()),
			Bind: succ.GetBind(),
		}
	}

	return pbn
}

func (rn *remoteNode) Notify(ctx context.Context, n_ chord.RemoteNode) error {
	ctx, span := rn.Start(ctx, "remoteNode.Notify", trace.WithAttributes(attrs.Node("n_", n_)))
	defer span.End()

	req := pb.NotifyRequest{
		Node: n_.AsProtobufNode(),
	}

	client, close, err := rn.getClient()
	if err != nil {
		span.RecordError(ctx, err)
		return err
	}
	defer close()

	_, err = client.Notify(ctx, &req)
	return err
}

func (rn *remoteNode) init(ctx context.Context) error {
	client, close, err := rn.getClient()
	if err != nil {
		return err
	}
	defer close()

	resp, err := client.GetNodeInfo(ctx, &pb.GetNodeInfoRequest{})
	if err != nil {
		return err
	}

	rn.id = chord.ID(resp.Node.GetId())
	rn.predNode = resp.Node.GetPred()
	rn.succNode = resp.Node.GetSucc()
	return nil
}

func NewRemote(ctx context.Context, bind string) (chord.RemoteNode, error) {
	rn := &remoteNode{
		Tracer:  global.Tracer(""),
		bind:    bind,
	}
	if err := rn.init(ctx); err != nil {
		return nil, err
	}
	return rn, nil
}
