package node

import (
	"context"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"github.com/kevinjqiu/chordio/chord"
	"github.com/kevinjqiu/chordio/pb"
)

type Node interface {
	fmt.Stringer

	GetID() chord.ID
	GetBind() string
	GetPredNode() (*NodeRef, error)
	GetSuccNode() (*NodeRef, error)
	AsProtobufNode() *pb.Node

	setNodeFactory(f factory)

	// FindPredecessor for the given ID
	FindPredecessor(context.Context, chord.ID) (Node, error)
	// FindSuccessor for the given ID
	FindSuccessor(context.Context, chord.ID) (Node, error)
	// find the closest finger entry that's preceding the ID
	ClosestPrecedingFinger(context.Context, chord.ID) (Node, error)
	// update the finger table entry at index i to node s
	UpdateFingerTableEntry(ctx context.Context, s Node, i int) error
}

type LocalNode interface {
	Node
	GetFingerTable() *FingerTable
	Join(ctx context.Context, introducerNode RemoteNode) error
	GetRank() chord.Rank
}

type RemoteNode interface {
	Node
}

// A node ref only contains ID and Bind info
// It's used to reference a node with minimal information
type NodeRef struct {
	ID   chord.ID
	Bind string
}

func AssignID(key []byte, m chord.Rank) chord.ID {
	hasher := sha1.New()
	hasher.Write(key)
	b := hasher.Sum(nil)
	return chord.ID(binary.BigEndian.Uint64(b) % (chord.ID(2).Pow(m.AsInt())).AsU64())
}
