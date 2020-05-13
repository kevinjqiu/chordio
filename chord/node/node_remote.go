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

type closeFunc func() error

// remoteNode is a proxy that implements the Node interface but
// delegate the calls to the remote node via grpc
type remoteNode struct {
	trace.Tracer
	id       chord.ID
	bind     string
	predNode *pb.Node
	succNode *pb.Node

	factory factory
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

func (rn *remoteNode) SetPredNode(ctx context.Context, n NodeRef) error {
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

func (rn *remoteNode) SetSuccNode(ctx context.Context, n NodeRef) error {
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

func (rn *remoteNode) setNodeFactory(f factory) {
	rn.factory = f
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

func (rn *remoteNode) GetPredNode() NodeRef {
	return &nodeRef{chord.ID(rn.predNode.Id), rn.predNode.Bind}
}

func (rn *remoteNode) GetSuccNode() NodeRef {
	return &nodeRef{chord.ID(rn.succNode.Id), rn.succNode.Bind}
}

func (rn *remoteNode) FindPredecessor(ctx context.Context, id chord.ID) (Node, error) {
	ctx, span := rn.Start(ctx, "remoteNode.FindPredecessor", trace.WithAttributes(core.Key("id").Int(id.AsInt())))
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

	return rn.factory.newRemoteNode(ctx, resp.Node.Bind)
}

func (rn *remoteNode) FindSuccessor(ctx context.Context, id chord.ID) (Node, error) {
	ctx, span := rn.Start(ctx, "remoteNode.FindSuccessor", trace.WithAttributes(core.Key("id").Int(id.AsInt())))
	defer span.End()
	logrus.Debug("[remoteNode] FindSuccessor: ", id)
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

	return rn.factory.newRemoteNode(ctx, resp.Node.Bind)
}

func (rn *remoteNode) ClosestPrecedingFinger(ctx context.Context, id chord.ID) (Node, error) {
	ctx, span := rn.Start(ctx, "remoteNode.ClosestPrecedingFinger", trace.WithAttributes(core.Key("id").Int(id.AsInt())))
	defer span.End()

	logrus.Debug("[remoteNode] ClosestPrecedingFinger: ", id)
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

	return rn.factory.newRemoteNode(ctx, resp.Node.Bind)
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

func (rn *remoteNode) UpdateFingerTableEntry(ctx context.Context, s Node, i int) error {
	ctx, span := rn.Start(ctx, "remoteNode.UpdateFingerTableEntry",
		trace.WithAttributes(core.Key("s").String(s.String())))
	defer span.End()

	req := pb.UpdateFingerTableRequest{
		Node: s.AsProtobufNode(),
		I:    int64(i),
	}

	client, close, err := rn.getClient()
	if err != nil {
		return err
	}
	defer close()

	_, err = client.UpdateFingerTable(ctx, &req)
	return err
}

func (rn *remoteNode) Notify(ctx context.Context, node Node) error {
	ctx, span := rn.Start(ctx, "remoteNode.Notify", trace.WithAttributes(core.Key("node").String(node.String())))
	defer span.End()

	req := pb.NotifyRequest{
		Node: node.AsProtobufNode(),
	}

	client, close, err := rn.getClient()
	if err != nil {
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

func NewRemote(ctx context.Context, bind string) (RemoteNode, error) {
	rn := &remoteNode{
		Tracer:  global.Tracer(""),
		bind:    bind,
		factory: defaultFactory{},
	}
	if err := rn.init(ctx); err != nil {
		return nil, err
	}
	return rn, nil
}
