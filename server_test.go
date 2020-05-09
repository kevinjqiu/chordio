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
	"strings"
	"testing"
)

var defaultServiceConfig = `{
	"methodConfig": [{
		"waitForReady": true,

		"retryPolicy": {
			"MaxAttempts": 100,
			"InitialBackoff": ".01s",
			"MaxBackoff": ".01s",
			"BackoffMultiplier": 1.0,
			"RetryableStatusCodes": [ "UNAVAILABLE" ]
		}
	}]
}`

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

func (tn testNode) status() {
	resp, err := tn.c.GetNodeInfo(context.Background(), &pb.GetNodeInfoRequest{
		IncludeFingerTable: true,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(ftCSV(resp.Ft))
}

func (tn testNode) assertFingerTable(t *testing.T, expectedFTEs []string) {
	resp, err := tn.c.GetNodeInfo(context.Background(), &pb.GetNodeInfoRequest{
		IncludeFingerTable: true,
	})
	if err != nil {
		panic(err)
	}
	actualFTEs := strings.Split(strings.TrimSpace(ftCSV(resp.Ft)), "\n")
	assert.Equal(t, expectedFTEs, actualFTEs)
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

func TestServer(t *testing.T) {
	n0 := newNode(0, 3)
	n1 := newNode(1, 3)
	n3 := newNode(3, 3)

	defer n0.stop()
	defer n1.stop()
	defer n3.stop()

	t.Run("initially the finger tables contain their owner nodes", func(t *testing.T) {
		n0.assertFingerTable(t, []string{
			"1,2,0",
			"2,4,0",
			"4,0,0",
		})

		n1.assertFingerTable(t, []string{
			"2,3,1",
			"3,5,1",
			"5,1,1",
		})

		n3.assertFingerTable(t, []string{
			"4,5,3",
			"5,7,3",
			"7,3,3",
		})
	})

	t.Run("after n0 and n1 join to each other, they have each other in their finger tables", func(t *testing.T) {
		n0.join(n1)
		n0.assertFingerTable(t, []string{
			"1,2,1",
			"2,4,1",
			"4,0,1",
		})
		n1.assertFingerTable(t, []string{
			"2,3,0",
			"3,5,0",
			"5,1,0",
		})
	})

	t.Run("after n3 join n1", func(t *testing.T) {
		n3.join(n0)
		n0.assertFingerTable(t, []string{
			"1,2,1",
			"2,4,1",
			"4,0,1",
		})
		n1.assertFingerTable(t, []string{
			"2,3,3",
			"3,5,0",  // FIXME: verify this
			"5,1,3",
		})
		n3.assertFingerTable(t, []string{
			"4,5,0",
			"5,7,0",
			"7,3,0",
		})
	})
}
