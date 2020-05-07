package chordio

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

	// findPredecessor for the given id
	findPredecessor(context.Context, chord.ChordID) (Node, error)
	// findSuccessor for the given id
	findSuccessor(context.Context, chord.ChordID) (Node, error)
	// find the closest finger entry that's preceding the id
	closestPrecedingFinger(context.Context, chord.ChordID) (Node, error)
	// update the finger table entry at index i to node s
	updateFingerTable(_ context.Context, s Node, i int) error
}

type NodeRef struct {
	id   chord.ChordID
	bind string
}

func AssignID(key []byte, m chord.Rank) chord.ChordID {
	hasher := sha1.New()
	hasher.Write(key)
	b := hasher.Sum(nil)
	return chord.ChordID(binary.BigEndian.Uint64(b) % chord.pow2(uint32(m)))
}
