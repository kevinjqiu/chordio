package chordio

import (
	"context"
	"fmt"
	"github.com/kevinjqiu/chordio/chord"
	"github.com/kevinjqiu/chordio/pb"
	"github.com/phayes/freeport"
	"google.golang.org/grpc"
	"testing"
	"time"
)

func newNode(id int, m int) (*Server, pb.ChordClient) {
	port, err := freeport.GetFreePort()
	if err != nil {
		panic(err)
	}

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	n, err := NewServer(Config{
		ID:   chord.ID(id),
		M:    chord.Rank(m),
		Bind: addr,
	})
	if err != nil {
		panic(err)
	}

	go func() {
		if err := n.Serve(); err != nil {
			fmt.Println(err)
		}
	}()

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	client := pb.NewChordClient(conn)
	return n, client
}

func TestServer(t *testing.T) {
	n0s, n0c := newNode(0, 3)
	n1s, n1c := newNode(1, 3)
	n3s, n3c := newNode(3, 3)

	defer n0s.GracefulStop()
	defer n1s.GracefulStop()
	defer n3s.GracefulStop()

	time.Sleep(500 * time.Millisecond)

	resp, err := n0c.GetNodeInfo(context.Background(), &pb.GetNodeInfoRequest{
		IncludeFingerTable:   true,
	})
	if err != nil { panic(err) }
	fmt.Println(resp)

	resp, err = n1c.GetNodeInfo(context.Background(), &pb.GetNodeInfoRequest{})
	if err != nil { panic(err) }
	fmt.Println(resp)

	resp, err = n3c.GetNodeInfo(context.Background(), &pb.GetNodeInfoRequest{})
	if err != nil { panic(err) }
	fmt.Println(resp)
}
