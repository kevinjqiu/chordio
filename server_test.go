package chordio

import (
	"fmt"
	"testing"
)

func withTwoNodeCluster(f func(nodes map[int]testNode)) {
	n0 := newNode(0, 3)
	n1 := newNode(1, 3)

	defer n0.stop()
	defer n1.stop()

	f(map[int]testNode{
		0: n0,
		1: n1,
	})
}

func TestServer(t *testing.T) {
	t.Run("initially the finger tables contain their owner nodes", func(t *testing.T) {
		withTwoNodeCluster(func(nodes map[int]testNode) {
			nodes[0].assertNeighbours(t, 0, 0)
			nodes[0].assertFingerTable(t, []string{
				"1,2,0",
				"2,4,0",
				"4,0,0",
			})

			nodes[1].assertNeighbours(t, 1, 1)
			nodes[1].assertFingerTable(t, []string{
				"2,3,1",
				"3,5,1",
				"5,1,1",
			})
		})
	})

	t.Run("after n0 and n1 join to each other, they have each other in their finger tables", func(t *testing.T) {
		testCases := [][]string{
			//{"0.stabilize", "0.fixFingers", "1.stabilize", "1.fixFingers"},
			{"0.stabilize", "1.stabilize", "0.fixFingers", "1.fixFingers"},
			//{"1.stabilize", "1.fixFingers", "0.stabilize", "0.fixFingers"},
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("%v", tc), func(t *testing.T) {
				withTwoNodeCluster(func(nodes map[int]testNode) {
					nodes[0].join(nodes[1])

					for _, op := range tc {
						runOperation(op, nodes)
					}

					nodes[0].assertNeighbours(t, 1, 1)
					nodes[0].assertFingerTable(t, []string{
						"1,2,1",
						"2,4,1",
						"4,0,1",
					})

					nodes[1].assertNeighbours(t, 0, 0)
					nodes[1].assertFingerTable(t, []string{
						"2,3,0",
						"3,5,0",
						"5,1,0",
					})
				})
			})
		}
	})
	//
	//t.Run("after n3 join n1", func(t *testing.T) {
	//	n3.join(n0)
	//	n0.assertFingerTable(t, []string{
	//		"1,2,1",
	//		"2,4,1",
	//		"4,0,1",
	//	})
	//	n1.assertFingerTable(t, []string{
	//		"2,3,3",
	//		"3,5,0",  // FIXME: verify this
	//		"5,1,3",
	//	})
	//	n3.assertFingerTable(t, []string{
	//		"4,5,0",
	//		"5,7,0",
	//		"7,3,0",
	//	})
	//})
}
