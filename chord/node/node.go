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

	GetID() chord.ChordID
	GetBind() string
	GetPredNode() (*NodeRef, error)
	GetSuccNode() (*NodeRef, error)
	AsProtobufNode() *pb.Node

	// FindPredecessor for the given ID
	FindPredecessor(context.Context, chord.ChordID) (Node, error)
	// FindSuccessor for the given ID
	FindSuccessor(context.Context, chord.ChordID) (Node, error)
	// find the closest finger entry that's preceding the ID
	ClosestPrecedingFinger(context.Context, chord.ChordID) (Node, error)
	// update the finger table entry at index i to node s
	UpdateFingerTableEntry(_ context.Context, s Node, i int) error
}

type NodeRef struct {
	ID   chord.ChordID
	Bind string
}

func AssignID(key []byte, m chord.Rank) chord.ChordID {
	hasher := sha1.New()
	hasher.Write(key)
	b := hasher.Sum(nil)
	return chord.ChordID(binary.BigEndian.Uint64(b) % (chord.ChordID(2).Pow(m.AsInt())).AsU64())
}
