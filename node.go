package chordio

import (
	"crypto/sha1"
	"encoding/binary"
)

type Node interface {
	GetID() ChordID
	GetBind() string
	GetPredNode() (*NodeRef, error)
	GetSuccNode() (*NodeRef, error)
}

type NodeRef struct {
	id   ChordID
	bind string
}

func assignID(key []byte, m Rank) ChordID {
	hasher := sha1.New()
	hasher.Write(key)
	b := hasher.Sum(nil)
	return ChordID(binary.BigEndian.Uint64(b) % pow2(uint32(m)))
}
