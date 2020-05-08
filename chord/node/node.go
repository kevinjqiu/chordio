package node

import (
	"context"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"github.com/kevinjqiu/chordio/chord"
	"github.com/kevinjqiu/chordio/pb"
)

type (
	canSetNodeConstructors interface {
		setLocalNodeConstructor(localNodeConstructor)
		setRemoteNodeConstructor(remoteNodeConstructor)
	}
)

type Node interface {
	fmt.Stringer
	canSetNodeConstructors

	GetID() chord.ID
	GetBind() string
	GetPredNode() (*NodeRef, error)
	GetSuccNode() (*NodeRef, error)
	AsProtobufNode() *pb.Node

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

// TODO: all these to support testing... an awful lot of structs and interfaces
// any way to simplify??
type (
	localNodeConstructor  func(id chord.ID, bind string, m chord.Rank, opts ...nodeConstructorOption) (LocalNode, error)
	remoteNodeConstructor func(ctx context.Context, bind string, opts ...nodeConstructorOption) (RemoteNode, error)
	nodeConstructorOption interface {
		apply(n canSetNodeConstructors)
	}

)

type _nodeConstructorOption struct {fn func(n canSetNodeConstructors)}

func (n _nodeConstructorOption) apply(o canSetNodeConstructors) {
	n.fn(o)
}

func nodeConstructorOptionFunc(fn func(n canSetNodeConstructors)) nodeConstructorOption {
	return _nodeConstructorOption{fn}
}

func withLocalNodeConstructor(fn localNodeConstructor) nodeConstructorOption {
	return nodeConstructorOptionFunc(func(n canSetNodeConstructors) {
		n.setLocalNodeConstructor(fn)
	})
}

func withRemoteNodeConstructor(fn remoteNodeConstructor) nodeConstructorOption {
	return nodeConstructorOptionFunc(func(n canSetNodeConstructors) {
		n.setRemoteNodeConstructor(fn)
	})
}

func AssignID(key []byte, m chord.Rank) chord.ID {
	hasher := sha1.New()
	hasher.Write(key)
	b := hasher.Sum(nil)
	return chord.ID(binary.BigEndian.Uint64(b) % (chord.ID(2).Pow(m.AsInt())).AsU64())
}
