package chordio

import (
	"context"
	"github.com/kevinjqiu/chordio/pb"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type RemoteNode struct {
	id       ChordID
	bind     string
	predNode *pb.Node
	succNode *pb.Node
	client   pb.ChordClient
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

func (rn *RemoteNode) FindPredecessor(id ChordID) (*RemoteNode, error) {
	req := pb.FindPredecessorRequest{
		Id: uint64(id),
	}

	resp, err := rn.client.FindPredecessor(context.Background(), &req)
	if err != nil {
		return nil, err
	}

	return newRemoteNodeFromPB(resp.Node)
}

func (rn *RemoteNode) FindSuccessor(id ChordID) (*RemoteNode, error) {
	logrus.Info("RemoteNode.FindSuccessor: ", id)
	req := pb.FindSuccessorRequest{
		KeyID: uint64(id), // TODO: rename to ID
	}

	resp, err := rn.client.FindSuccessor(context.Background(), &req)
	if err != nil {
		return nil, err
	}

	return newRemoteNodeFromPB(resp.Node)
}

func (rn *RemoteNode) ClosestPrecedingFinger(id ChordID) (*RemoteNode, error) {
	req := pb.ClosestPrecedingFingerRequest{
		Id: uint64(id),
	}

	resp, err := rn.client.ClosestPrecedingFinger(context.Background(), &req)
	if err != nil {
		return nil, err
	}

	return newRemoteNodeFromPB(resp.Node)
}

func newRemoteNode(bind string) (*RemoteNode, error) {
	conn, err := grpc.Dial(bind, grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrapf(err, "unable to initiate grpc client for node: %v", bind)
	}

	client := pb.NewChordClient(conn)
	resp, err := client.GetNodeInfo(context.Background(), &pb.GetNodeInfoRequest{})
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