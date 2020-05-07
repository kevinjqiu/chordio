package chord

import (
	"context"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"github.com/kevinjqiu/chordio/pb"
)

type Node interface {
	fmt.Stringer

	GetID() ID
	GetBind() string
	GetPredNode() (*NodeRef, error)
	GetSuccNode() (*NodeRef, error)
	AsProtobufNode() *pb.Node

	// FindPredecessor for the given ID
	FindPredecessor(context.Context, ID) (Node, error)
	// FindSuccessor for the given ID
	FindSuccessor(context.Context, ID) (Node, error)
	// find the closest finger entry that's preceding the ID
	ClosestPrecedingFinger(context.Context, ID) (Node, error)
	// update the finger table entry at index i to node s
	UpdateFingerTableEntry(_ context.Context, s Node, i int) error
}

// A node ref only contains ID and Bind info
// It's used to reference a node with minimal information
type NodeRef struct {
	ID   ID
	Bind string
}

func AssignID(key []byte, m Rank) ID {
	hasher := sha1.New()
	hasher.Write(key)
	b := hasher.Sum(nil)
	return ID(binary.BigEndian.Uint64(b) % (ID(2).Pow(m.AsInt())).AsU64())
}
