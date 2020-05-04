package chordio

import (
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
