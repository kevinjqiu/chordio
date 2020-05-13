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
	"strconv"
	"strings"
	"testing"
)

var defaultServiceConfig = `{
	"methodConfig": [{
		"waitForReady": true,

		"retryPolicy": {
			"MaxAttempts": 100,
			"InitialBackoff": ".01s",
			"MaxBackoff": ".50s",
			"BackoffMultiplier": 2.0,
			"RetryableStatusCodes": [ "UNAVAILABLE" ]
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
	case "fixFingers":
		node.fixFingers()
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
	c    pb.ChordClient
	addr string
}

func (tn testNode) stop() {
	tn.s.GracefulStop()
}

func (tn testNode) status() *pb.GetNodeInfoResponse {
	resp, err := tn.c.GetNodeInfo(context.Background(), &pb.GetNodeInfoRequest{
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
	resp, err := tn.c.JoinRing(context.Background(), &pb.JoinRingRequest{
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

func (tn testNode) stabilize() {
	_, _ = tn.c.X_Stabilize(context.Background(), &pb.StabilizeRequest{})
}

func (tn testNode) fixFingers() {
	_, _ = tn.c.X_FixFinger(context.Background(), &pb.FixFingerRequest{})
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
	})
	if err != nil {
		panic(err)
	}

	go func() {
		if err := server.Serve(); err != nil {
			fmt.Println(err)
		}
	}()

	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithDefaultServiceConfig(defaultServiceConfig))
	if err != nil {
		panic(err)
	}
	client := pb.NewChordClient(conn)
	return testNode{
		m:    uint32(m),
		id:   uint64(id),
		c:    client,
		s:    server,
		addr: addr,
	}
}
