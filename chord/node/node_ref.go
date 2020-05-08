package node

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"github.com/kevinjqiu/chordio/chord"
)

// A node ref only contains ID and Bind info
// It's used to reference a node with minimal information
type nodeRef struct {
	ID   chord.ID
	Bind string
}

func (nr *nodeRef) GetID() chord.ID {
	return nr.ID
}

func (nr *nodeRef) GetBind() string {
	return nr.Bind
}

func (nr nodeRef) String() string {
	return fmt.Sprintf("<@%d %s>", nr.ID, nr.Bind)
}

func AssignID(key []byte, m chord.Rank) chord.ID {
	hasher := sha1.New()
	hasher.Write(key)
	b := hasher.Sum(nil)
	return chord.ID(binary.BigEndian.Uint64(b) % (chord.ID(2).Pow(m.AsInt())).AsU64())
}
