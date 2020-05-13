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

		SetPredNode(ctx context.Context, n NodeRef) error
		SetSuccNode(ctx context.Context, n NodeRef) error

		// For stabilization
		Notify(ctx context.Context, n_ Node) error
	}

	LocalNode interface {
		Node
		GetFingerTable() *FingerTable
		Join(ctx context.Context, introducerNode RemoteNode) error
		GetRank() chord.Rank
		// For stabilization
		Stabilize(ctx context.Context) error
		FixFingers(ctx context.Context) error
	}

	RemoteNode interface {
		Node
	}
)
