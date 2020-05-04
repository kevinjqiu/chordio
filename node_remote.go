package chordio

import (
	"context"
	"fmt"
	"github.com/kevinjqiu/chordio/pb"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/plugin/grpctrace"
	"google.golang.org/grpc"
)

type RemoteNode struct {
	id       ChordID
	bind     string
	predNode *pb.Node
	succNode *pb.Node
	client   pb.ChordClient
}

func (rn *RemoteNode) String() string {
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

func (rn *RemoteNode) GetID() ChordID {
	return rn.id
}

func (rn *RemoteNode) GetBind() string {
	return rn.bind
}

func (rn *RemoteNode) GetPredNode() (*NodeRef, error) {
	return &NodeRef{ChordID(rn.predNode.Id), rn.predNode.Bind}, nil
}

func (rn *RemoteNode) GetSuccNode() (*NodeRef, error) {
	return &NodeRef{ChordID(rn.succNode.Id), rn.succNode.Bind}, nil
}

func (rn *RemoteNode) FindPredecessor(ctx context.Context, id ChordID) (*RemoteNode, error) {
	logrus.Debug("[RemoteNode] FindPredecessor: ", id)
	req := pb.FindPredecessorRequest{
		Id: uint64(id),
	}

	resp, err := rn.client.FindPredecessor(ctx, &req)
	if err != nil {
		return nil, err
	}

	return newRemoteNode(ctx, resp.Node.Bind)
}

func (rn *RemoteNode) FindSuccessor(ctx context.Context, id ChordID) (*RemoteNode, error) {
	logrus.Debug("[RemoteNode] FindSuccessor: ", id)
	req := pb.FindSuccessorRequest{
		Id: uint64(id),
	}

	resp, err := rn.client.FindSuccessor(ctx, &req)
	if err != nil {
		return nil, err
	}

	return newRemoteNode(ctx, resp.Node.Bind)
}

func (rn *RemoteNode) ClosestPrecedingFinger(ctx context.Context, id ChordID) (*RemoteNode, error) {
	logrus.Debug("[RemoteNode] ClosestPrecedingFinger: ", id)
	req := pb.ClosestPrecedingFingerRequest{
		Id: uint64(id),
	}

	resp, err := rn.client.ClosestPrecedingFinger(ctx, &req)
	if err != nil {
		return nil, err
	}

	return newRemoteNode(ctx, resp.Node.Bind)
}

func (rn *RemoteNode) AsProtobufNode() *pb.Node {
	pbn := &pb.Node{
		Id:   uint64(rn.GetID()),
		Bind: rn.GetBind(),
		Pred: nil,
		Succ: nil,
	}

	pred, err := rn.GetPredNode()
	if err != nil {
		pbn.Pred = &pb.Node{
			Id:   uint64(pred.id),
			Bind: pred.bind,
		}
	}

	succ, err := rn.GetSuccNode()
	if err != nil {
		pbn.Succ = &pb.Node{
			Id:   uint64(succ.id),
			Bind: succ.bind,
		}
	}

	return pbn
}

func newRemoteNode(ctx context.Context, bind string) (*RemoteNode, error) {
	conn, err := grpc.Dial(bind,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(grpctrace.UnaryClientInterceptor(global.Tracer(telemetryServiceName))),
		grpc.WithStreamInterceptor(grpctrace.StreamClientInterceptor(global.Tracer(telemetryServiceName))),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to initiate grpc client for node: %v", bind)
	}

	client := pb.NewChordClient(conn)
	resp, err := client.GetNodeInfo(ctx, &pb.GetNodeInfoRequest{})
	if err != nil {
		return nil, err
	}

	rn := &RemoteNode{
		id:       ChordID(resp.Node.GetId()),
		bind:     bind,
		predNode: resp.Node.GetPred(),
		succNode: resp.Node.GetSucc(),
		client:   client,
	}
	return rn, nil
}
