package chordio

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

func withCluster(m int, nodeIDs []int, f func(nodes map[int]testNode)) {
	testNodes := map[int]testNode{}

	for _, nodeID := range nodeIDs {
		testNodes[nodeID] = newNode(nodeID, m)
	}

	wg := sync.WaitGroup{}
	wg.Add(len(testNodes))
	for _, tn := range testNodes {
		go func() {
			tn.status()
			wg.Done()
		}()
	}
	wg.Wait()

	f(testNodes)

	for _, n := range testNodes {
		n.stop()
	}
}

func TestServer(t *testing.T) {
	t.Run("initially the finger tables contain their owner nodes", func(t *testing.T) {
		withCluster(3, []int{0, 1}, func(nodes map[int]testNode) {
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
			{"0.stabilize", "1.stabilize"},
			{"1.stabilize", "0.stabilize", "1.stabilize"},
			{"0.stabilize", "0.stabilize", "1.stabilize"},
			{"1.stabilize", "1.stabilize", "0.stabilize", "1.stabilize"},
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("%v", tc), func(t *testing.T) {
				withCluster(3, []int{0, 1}, func(nodes map[int]testNode) {
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

	t.Run("after n3 join n1", func(t *testing.T) {
		withCluster(3, []int{0, 1, 3}, func(nodes map[int]testNode) {
			nodes[0].join(nodes[1])
			nodes[3].join(nodes[0])

			//ctx, _ := context.WithTimeout(context.Background(), 1*time.Minute)
			waitForStabilization(context.Background(), nodes)

			fmt.Println(nodes[0].status().Node.GetPred())
			fmt.Println(nodes[0].status().Node.GetSucc())

			fmt.Println(nodes[1].status().Node.GetPred())
			fmt.Println(nodes[1].status().Node.GetSucc())

			fmt.Println(nodes[3].status().Node.GetPred())
			fmt.Println(nodes[3].status().Node.GetSucc())

			fmt.Println(ftCSV(nodes[0].status().GetFt()))
			fmt.Println(ftCSV(nodes[1].status().GetFt()))
			fmt.Println(ftCSV(nodes[3].status().GetFt()))
		})
	})
}
