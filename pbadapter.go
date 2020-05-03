package chordio

import "github.com/kevinjqiu/chordio/pb"

func newPBNode(n Node, includeNeighbours bool) *pb.Node {
	pbn := &pb.Node{
		Id:   uint64(n.id),
		Bind: n.bind,
		Pred: nil,
		Succ: nil,
	}
	return pbn
}

func newRemoteNodeFromPB(pbn *pb.Node) (*RemoteNode, error) {
	n := Node{
		id:   ChordID(pbn.Id),
		bind: pbn.Bind,
		pred: ChordID(pbn.Pred.Id),
		succ: ChordID(pbn.Succ.Id),
	}
	return newRemoteNode(n)
}