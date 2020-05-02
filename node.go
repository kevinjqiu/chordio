package chordio

import (
	"context"
	"crypto/sha1"
	"encoding/binary"
	"github.com/kevinjqiu/chordio/pb"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type Node struct {
	id   ChordID
	bind string
	pred ChordID
	succ ChordID
}

func newNode(bind string, m Rank) Node {
	n := Node{
		bind: bind,
	}
	n.id = assignID([]byte(bind), m)
	n.pred = n.id
	n.succ = n.id
	return n
}

func assignID(key []byte, m Rank) ChordID {
	hasher := sha1.New()
	hasher.Write(key)
	b := hasher.Sum(nil)
	return ChordID(binary.BigEndian.Uint64(b) % pow2(uint32(m)))
}

type RemoteNode struct {
	Node
	client pb.ChordClient
}

func (rn *RemoteNode) FindSuccessor(id ChordID) (*RemoteNode, error) {
	req := pb.FindSuccessorRequest{
		KeyID: uint64(id),
	}

	resp, err := rn.client.FindSuccessor(context.Background(), &req)
	if err != nil {
		return nil, err
	}

	n := Node{
		id:   ChordID(resp.Node.Id),
		bind: resp.Node.Bind,
		pred: ChordID(resp.Node.Pred.Id),
		succ: ChordID(resp.Node.Succ.Id),
	}

	return newRemoteNode(n)
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
