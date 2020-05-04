package chordio

import (
	"context"
	"github.com/kevinjqiu/chordio/pb"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type RemoteNode struct {
	Node
	predNode *pb.Node
	succNode *pb.Node
	client pb.ChordClient
}

func (rn *RemoteNode) getPredNode() (INode, error) {
	return newINodeFromPB(rn.predNode)
}

func (rn *RemoteNode) getSuccNode() (INode, error) {
	return newINodeFromPB(rn.succNode)
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
		KeyID: uint64(id),  // TODO: rename to ID
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

func newRemoteNode(node Node) (*RemoteNode, error) {
	rn := &RemoteNode{
		Node: node,
	}

	conn, err := grpc.Dial(node.bind, grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrapf(err, "unable to initiate grpc client for node: %v", node)
	}

	rn.client = pb.NewChordClient(conn)
	return rn, nil
}
