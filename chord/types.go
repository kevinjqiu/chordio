package chord

import (
	"context"
	"fmt"
	"github.com/kevinjqiu/chordio/pb"
	"io"
)

type (
	NodeRef interface {
		fmt.Stringer
		GetID() ID
		GetBind() string
	}

	Node interface {
		NodeRef

		GetPredNode() NodeRef
		GetSuccNode() NodeRef
		AsProtobufNode() *pb.Node

		// FindPredecessor for the given ID
		FindPredecessor(ctx context.Context, id ID) (Node, error)
		// FindSuccessor for the given ID
		FindSuccessor(ctx context.Context, id ID) (Node, error)
		// find the closest finger entry that's preceding the ID
		ClosestPrecedingFinger(ctx context.Context, id ID) (Node, error)

		SetPredNode(ctx context.Context, n NodeRef) error
		SetSuccNode(ctx context.Context, n NodeRef) error

		// For stabilization
		Notify(ctx context.Context, n_ RemoteNode) error
	}

	LocalNode interface {
		Node
		GetFingerTable() FingerTable
		Join(ctx context.Context, introducerNode RemoteNode) error
		GetRank() Rank
		// For stabilization
		Stabilize(ctx context.Context) error
		FixFingers(ctx context.Context) error
	}

	RemoteNode interface {
		Node
	}

	FingerTable interface {
		fmt.Stringer
		PrettyPrint(writer io.Writer)
		ReplaceNodeWithAnotherEntry(i, j int)
		SetNodeAtEntry(i int, n NodeRef)
		GetEntry(i int) FingerTableEntry
		GetNodeByID(nodeID ID) (NodeRef, bool)
		HasNode(id ID) bool
		AsProtobufFT() *pb.FingerTable
	}

	FingerTableEntry interface {
		fmt.Stringer
		GetStart() ID
		SetStart(start ID)
		GetInterval() Interval
		SetInterval(iv Interval)
		GetNode() NodeRef
		SetNode(n NodeRef)
	}
)
