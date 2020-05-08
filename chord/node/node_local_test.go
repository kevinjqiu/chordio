package node

import (
	"context"
	"fmt"
	"github.com/kevinjqiu/chordio/chord"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLocalNode_ClosestPrecedingFinger(t *testing.T) {
	n, err := NewLocal(1, "n1", chord.Rank(3))
	assert.Nil(t, err)

	n2 := newMockNode(2, "n2")
	n5 := newMockNode(5, "n5")

	n.GetFingerTable().SetEntry(0, n2)
	n.GetFingerTable().SetEntry(1, n2)
	n.GetFingerTable().SetEntry(2, n5)
	n.GetFingerTable().Print(nil)

	n_, err := n.ClosestPrecedingFinger(context.Background(), 1)
	fmt.Println(err)
	fmt.Println(n_)
}
