package node

import (
	"fmt"
	"github.com/kevinjqiu/chordio/chord"
	"github.com/pkg/errors"
)

var (
	errNoSuccessorNode   = errors.New("no successor node found")
	errNoPredecessorNode = errors.New("no predecessor node found")
)

func errNodeNotFound(id chord.ID) error {
	return errors.New(fmt.Sprintf("node %d s not known to the current local node", id))
}
