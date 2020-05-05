package chordio

import (
	"context"
	"crypto/sha1"
	"encoding/binary"
	"github.com/kevinjqiu/chordio/pb"
)

type Node interface {
	GetID() ChordID
	GetBind() string
	GetPredNode() (*NodeRef, error)
	GetSuccNode() (*NodeRef, error)
	AsProtobufNode() *pb.Node

	// findPredecessor for the given id
	findPredecessor(context.Context, ChordID) (Node, error)
	// findSuccessor for the given id
	findSuccessor(context.Context, ChordID) (Node, error)
	// find the closest finger entry that's preceding the id
	closestPrecedingFinger(context.Context, ChordID) (Node, error)
	// update the finger table entry at index i to node s
	updateFingerTable(_ context.Context, s Node, i int) error
}

type NodeRef struct {
	id   ChordID
	bind string
}

func AssignID(key []byte, m Rank) ChordID {
	hasher := sha1.New()
	hasher.Write(key)
	b := hasher.Sum(nil)
	return ChordID(binary.BigEndian.Uint64(b) % pow2(uint32(m)))
}
