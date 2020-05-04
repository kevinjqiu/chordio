package chordio

import "github.com/kevinjqiu/chordio/pb"

//func newPBNodeFromRemoteNode(n RemoteNode) *pb.Node {
//	// can't simply have ID
//	pbn:= &pb.Node{
//		Id:   uint64(n.id),
//		Bind: n.bind,
//		Pred: newPBNodeFromLocalNode(pred),
//		Succ: newPBNodeFromLocalNode(succ),
//	}
//}

func newPBNodeFromLocalNode(n LocalNode) *pb.Node {
	pbn := &pb.Node{
		Id:   uint64(n.GetID()),
		Bind: n.GetBind(),
		Pred: nil,
		Succ: nil,
	}

	predNode, err := n.GetPredNode()
	if err == nil {
		pbn.Pred = &pb.Node{
			Id:   uint64(predNode.id),
			Bind: predNode.bind,
		}
	}

	succNode, err := n.GetSuccNode()
	if err == nil {
		pbn.Succ = &pb.Node{
			Id:   uint64(succNode.id),
			Bind: succNode.bind,
		}
	}
	return pbn
}

func newRemoteNodeFromPB(pbn *pb.Node) (*RemoteNode, error) {
	return newRemoteNode(pbn.Bind)
}
