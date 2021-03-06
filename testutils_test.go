package chordio

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kevinjqiu/chordio/chord"
	"github.com/kevinjqiu/chordio/pb"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"
)

var defaultServiceConfig = `{
	"methodConfig": [{
		"name": [{
			"service": "Chord"
		}],
		"waitForReady": true,

		"retryPolicy": {
			"MaxAttempts": 100,
			"InitialBackoff": "1s",
			"MaxBackoff": "5s",
			"BackoffMultiplier": 2.0
		}
	}]
}`

func runOperation(op string, nodes map[int]testNode) {
	parts := strings.Split(op, ".")
	id, err := strconv.Atoi(parts[0])
	if err != nil {
		panic(err)
	}
	node := nodes[id]
	switch parts[1] {
	case "stabilize":
		node.stabilize()
	default:
		panic(fmt.Sprintf("unrecognized command: %s", parts[1]))
	}
}

func ftCSV(table *pb.FingerTable) string {
	var b bytes.Buffer
	for _, e := range table.Entries {
		b.WriteString(fmt.Sprintf("%d,%d,%d", e.Start, e.End, e.NodeID))
		b.WriteString("\n")
	}
	return b.String()
}

type testNode struct {
	m    uint32
	id   uint64
	s    *Server
	addr string
}

func (tn testNode) stop() {
	tn.s.GracefulStop()
}

func (tn testNode) status() *pb.GetNodeInfoResponse {
	c, close := tn.getClient()
	defer close()
	resp, err := c.GetNodeInfo(context.Background(), &pb.GetNodeInfoRequest{
		IncludeFingerTable: true,
	})
	if err != nil {
		panic(err)
	}
	return resp
}

func (tn testNode) assertFingerTable(t *testing.T, expectedFTEs []string) {
	resp := tn.status()
	actualFTEs := strings.Split(strings.TrimSpace(ftCSV(resp.Ft)), "\n")
	assert.Equal(t, expectedFTEs, actualFTEs)
}

func (tn testNode) assertNeighbours(t *testing.T, predID, succID uint64) {
	resp := tn.status()
	assert.Equal(t, predID, resp.Node.GetPred().GetId())
	assert.Equal(t, succID, resp.Node.GetSucc().GetId())
}

func (tn testNode) join(other testNode) {
	c, close := tn.getClient()
	defer close()

	resp, err := c.JoinRing(context.Background(), &pb.JoinRingRequest{
		Introducer: &pb.Node{
			Id:   other.id,
			Bind: other.addr,
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(resp)
}

func (tn testNode) stabilize() (int, error) {
	c, close := tn.getClient()
	defer close()

	resp, err := c.X_Stabilize(context.Background(), &pb.StabilizeRequest{})
	return int(resp.NumFingerTableEntryChanges), err
}

func (tn testNode) getClient() (pb.ChordClient, func() error) {
	conn, err := grpc.Dial(tn.addr, grpc.WithInsecure(), grpc.WithDefaultServiceConfig(defaultServiceConfig))

	if err != nil {
		panic(err)
	}
	client := pb.NewChordClient(conn)

	return client, func() error { return conn.Close() }
}

func newNode(id int, m int) testNode {
	port, err := freeport.GetFreePort()
	if err != nil {
		panic(err)
	}

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	server, err := NewServer(Config{
		ID:   chord.ID(id),
		M:    chord.Rank(m),
		Bind: addr,
		Stabilization: StabilizationConfig{
			Disabled: true,
		},
	})
	if err != nil {
		panic(err)
	}

	go func() {
		if err := server.Serve(); err != nil {
			fmt.Println(err)
		}
	}()

	return testNode{
		m:    uint32(m),
		id:   uint64(id),
		s:    server,
		addr: addr,
	}
}

func isStable(threshold int, numChangesHistory []int) bool {
	if len(numChangesHistory) < threshold {
		return false
	}

	thresholdIdx := len(numChangesHistory) - threshold
	for i := thresholdIdx; i < len(numChangesHistory); i++ {
		if numChangesHistory[i] != 0 {
			return false
		}
	}
	return true
}

// X consecutive stabilization calls with 0 numChanges
// and then it's deemed "stable"
func waitForStabilization(threshold int, nodes map[int]testNode) {
	numChangesChan := make(chan [2]int)
	stopChans := make(map[int]chan bool)

	for id, n := range nodes {
		shouldStop := make(chan bool)
		stopChans[id] = shouldStop

		go func(id int, shouldStop <- chan bool) {
			rand.Seed(time.Now().UnixNano())
			for {
				duration := time.Duration(rand.Int63() % 5e9) + time.Second
				fmt.Println("sleep duration: ", duration)
				timer := time.NewTimer(duration)

				select {
				case <- timer.C:
					n, err := n.stabilize()
					if err == nil {
						numChangesChan <- [2]int{id, n}
					}
				case <- shouldStop:
					return
				}
			}
		}(id, shouldStop)
	}

	nodeChangeHistory := map[int][]int{}

	for {
		select {
		case res := <- numChangesChan:
			id := res[0]
			n := res[1]
			hist, ok := nodeChangeHistory[id]
			if !ok {
				hist = make([]int, 0, 0)
			}
			hist = append(hist, n)
			nodeChangeHistory[id] = hist

			for id, hist := range nodeChangeHistory {
				fmt.Printf("%d: %v\n", id, hist)
			}

			var notStable bool
			for _, hist := range nodeChangeHistory {
				if !isStable(threshold, hist) {
					notStable = true
					break
				}
			}
			if !notStable {
				// if everything is stable, send signal to all proc to stop stabilizing
				for _, stopChan := range stopChans {
					stopChan <- true
				}
				return
			}
		}
	}
}