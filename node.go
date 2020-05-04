package chordio

import (
	"crypto/sha1"
	"encoding/binary"
)

type INode interface {
	GetID() ChordID
	GetBind() string
	//GetPred() ChordID
	//GetSucc() ChordID
	GetPredNode() (*NodeRef, error)
	GetSuccNode() (*NodeRef, error)
}

type NodeRef struct {
	id   ChordID
	bind string
}

type Node struct {
	id   ChordID
	bind string
	pred ChordID
	succ ChordID
}

func (n Node) GetID() ChordID   { return n.id }
func (n Node) GetBind() string  { return n.bind }
func (n Node) GetPred() ChordID { return n.pred }
func (n Node) GetSucc() ChordID { return n.succ }

func newNode(bind string, m Rank) Node {
	n := Node{
		bind: bind,
	}
	n.id = assignID([]byte(bind), m)
	n.pred = n.id
	n.succ = n.id
	return n
}

func assignID(key []byte, m Rank) ChordID {
	hasher := sha1.New()
	hasher.Write(key)
	b := hasher.Sum(nil)
	return ChordID(binary.BigEndian.Uint64(b) % pow2(uint32(m)))
}
