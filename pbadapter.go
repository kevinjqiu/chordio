package chordio

import "github.com/kevinjqiu/chordio/pb"

func newPBNodeFromLocalNode(n LocalNode) *pb.Node {
	pbn := &pb.Node{
		Id:   uint64(n.GetID()),
		Bind: n.GetBind(),
		Pred: nil,
		Succ: nil,
	}

	predNode, err := n.getPredNode()
	if err == nil {
		pbn.Pred = &pb.Node{
			Id:   uint64(predNode.GetID()),
			Bind: predNode.GetBind(),
		}
	}

	succNode, err := n.getSuccNode()
	if err == nil {
		pbn.Succ = &pb.Node{
			Id:   uint64(succNode.GetID()),
			Bind: succNode.GetBind(),
		}
	}
	return pbn
}

func newRemoteNodeFromPB(pbn *pb.Node) (*RemoteNode, error) {
	n, err := newINodeFromPB(pbn)
	if err != nil {
		return nil, err
	}
	return newRemoteNode(n.(Node))  // TODO: don't type assert
}

func newINodeFromPB(pbn *pb.Node) (INode, error) {
	n := Node{
		id:   ChordID(pbn.Id),
		bind: pbn.Bind,
		pred: ChordID(pbn.Pred.Id),
		succ: ChordID(pbn.Succ.Id),
	}
	return n, nil
}
