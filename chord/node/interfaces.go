package node

import (
	"context"
	"fmt"
	"github.com/kevinjqiu/chordio/chord"
	"github.com/kevinjqiu/chordio/pb"
)

type (
	NodeRef interface {
		fmt.Stringer
		GetID() chord.ID
		GetBind() string
	}

	Node interface {
		NodeRef

		GetPredNode() NodeRef
		GetSuccNode() NodeRef
		AsProtobufNode() *pb.Node

		setNodeFactory(f factory)

		// FindPredecessor for the given ID
		FindPredecessor(ctx context.Context, id chord.ID) (Node, error)
		// FindSuccessor for the given ID
		FindSuccessor(ctx context.Context, id chord.ID) (Node, error)
		// find the closest finger entry that's preceding the ID
		ClosestPrecedingFinger(ctx context.Context, id chord.ID) (Node, error)
		// update the finger table entry at index i to node s
		UpdateFingerTableEntry(ctx context.Context, s Node, i int) error
	}

	LocalNode interface {
		Node
		GetFingerTable() *FingerTable
		Join(ctx context.Context, introducerNode RemoteNode) error
		GetRank() chord.Rank
		SetPredNode(n NodeRef)
		SetSuccNode(n NodeRef)
	}

	RemoteNode interface {
		Node
	}
)
