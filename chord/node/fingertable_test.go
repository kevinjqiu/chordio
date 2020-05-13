package node

import (
	"github.com/kevinjqiu/chordio/chord"
	"github.com/stretchr/testify/assert"
	"testing"
)

type extraSetupStep func(t *testing.T, node chord.LocalNode, ft *fingerTable)

func setup(t *testing.T, extraSteps ...extraSetupStep) (chord.LocalNode, *fingerTable) {
	m := chord.Rank(5)
	node, _ := NewLocal(15, "localhost:1234", m)
	assert.Equal(t, chord.ID(15), node.GetID())

	ft := newFingerTable(node, m)

	for _, step := range extraSteps {
		step(t, node, ft)
	}
	return node, ft
}

func TestNewFingerTable(t *testing.T) {
	_, ft := setup(t)
	assert.Equal(t, 5, len(ft.entries))
	assert.Equal(t, chord.ID(16), ft.entries[0].Start)
	assert.Equal(t, chord.ID(17), ft.entries[1].Start)
	assert.Equal(t, chord.ID(19), ft.entries[2].Start)
	assert.Equal(t, chord.ID(23), ft.entries[3].Start)
	assert.Equal(t, chord.ID(31), ft.entries[4].Start)
	ft.PrettyPrint(nil)
}

func TestFingerTable_ReplaceNodeAt(t *testing.T) {
	t.Run("the replaced node is the owner node", func(t *testing.T) {
		_, ft := setup(t, func(t *testing.T, node chord.LocalNode, ft *fingerTable) {
			replacingNode := &nodeRef{
				ID:   35,
				Bind: "n35",
			}

			for i := 1; i < len(ft.entries); i++ {
				ft.entries[i].Node = replacingNode
			}
			ft.neighbourhood[chord.ID(35)] = replacingNode
		})

		ft.ReplaceNodeWithAnotherEntry(0, 1)
		assert.True(t, ft.entries[0].Node.GetID() == 35)
		assert.True(t, ft.HasNode(chord.ID(15)))
	})

	t.Run("the replaced node is no longer in the finger table", func(t *testing.T) {
		_, ft := setup(t, func(t *testing.T, node chord.LocalNode, ft *fingerTable) {
			nodeToBeReplaced := &nodeRef{
				ID:   35,
				Bind: "n35",
			}
			ft.entries[1].Node = nodeToBeReplaced
			ft.neighbourhood[chord.ID(35)] = nodeToBeReplaced
		})

		assert.True(t, ft.entries[1].Node.GetID() == 35)
		ft.ReplaceNodeWithAnotherEntry(1, 2)
		assert.True(t, ft.entries[1].Node.GetID() == 15)
		assert.False(t, ft.HasNode(chord.ID(35)))
	})

	t.Run("the replacing node is the same as the replaced node", func(t *testing.T) {
		_, ft := setup(t)
		ft.ReplaceNodeWithAnotherEntry(0, 1)
		for _, fte := range ft.entries {
			assert.Equal(t, chord.ID(15), fte.Node.GetID())
		}
	})
}

func TestFingerTable_SetEntryAt(t *testing.T) {
	t.Run("the replaced node is the owner node", func(t *testing.T) {
		_, ft := setup(t, func(t *testing.T, node chord.LocalNode, ft *fingerTable) {
			replacingNode := &nodeRef{
				ID:   35,
				Bind: "n35",
			}

			for i := 1; i < len(ft.entries); i++ {
				ft.entries[i].Node = replacingNode
			}
			ft.neighbourhood[chord.ID(35)] = replacingNode
		})

		newNode, _ := newMockNode(25, "N25")
		ft.SetNodeAtEntry(0, newNode)
		assert.True(t, ft.entries[0].Node.GetID() == newNode.GetID())
		assert.True(t, ft.HasNode(chord.ID(15)))
	})

	t.Run("the replaced node is no longer in the finger table", func(t *testing.T) {
		_, ft := setup(t, func(t *testing.T, node chord.LocalNode, ft *fingerTable) {
			nodeToBeReplaced := &nodeRef{
				ID:   35,
				Bind: "n35",
			}
			ft.entries[1].Node = nodeToBeReplaced
			ft.neighbourhood[chord.ID(35)] = nodeToBeReplaced
		})

		assert.True(t, ft.entries[1].Node.GetID() == 35)
		newNode, _ := newMockNode(25, "N25")
		ft.SetNodeAtEntry(1, newNode)
		assert.True(t, ft.entries[1].Node.GetID() == newNode.GetID())
		assert.False(t, ft.HasNode(chord.ID(35)))
	})
}

func TestFingerTable_AsProtobufFT(t *testing.T) {
	_, ft := setup(t)
	pbft := ft.AsProtobufFT()
	assert.Equal(t, len(pbft.Entries), len(ft.entries))

	for i, fte := range ft.entries {
		pbfte := pbft.Entries[i]
		assert.Equal(t, pbfte.Start, fte.Start.AsU64())
		assert.Equal(t, pbfte.End, fte.Interval.End.AsU64())
		assert.Equal(t, pbfte.NodeID, fte.Node.GetID().AsU64())
	}
}
