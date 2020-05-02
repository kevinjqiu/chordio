package chordio

import (
	context "context"
	"fmt"
	"github.com/kevinjqiu/chordio/pb"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
)

type Server struct {
	m    Rank
	bind string

	neighbourhood *Neighbourhood
	local         Node
	finger        FingerTable
}

func (n *Server) FindSuccessor(_ context.Context, request *pb.FindSuccessorRequest) (*pb.FindSuccessorResponse, error) {
	panic("implement me")
}

func (n *Server) GetID(_ context.Context, _ *pb.GetIDRequest) (*pb.GetIDResponse, error) {
	return &pb.GetIDResponse{
		Id: uint64(n.local.id),
	}, nil
}

func (n *Server) JoinRing(_ context.Context, request *pb.JoinRingRequest) (*pb.JoinRingResponse, error) {
	introNode := newNode(request.Introducer.Bind, n.m)
	if err := n.join(introNode); err != nil {
		return nil, err
	}
	return &pb.JoinRingResponse{}, nil
}

func (n *Server) initFinger(remote *RemoteNode) error {
	local := &n.local
	succ, err := remote.FindSuccessor(n.finger.entries[0].start)
	if err != nil {
		return err
	}
	local.pred = succ.pred

	for i := 0; i < int(n.m)-1; i++ {
		if n.finger.entries[i+1].start.In(local.id, n.finger.entries[i].node, n.m) {
			n.finger.entries[i+1].node = n.finger.entries[i].node
		} else {
			newSucc, err := remote.FindSuccessor(n.finger.entries[i+1].start)
			if err != nil {
				return err
			}
			n.finger.entries[i+1].node = newSucc.id
		}
	}
	return nil
}

func (n *Server) join(introducerNode Node) error {
	rn, err := newRemoteNode(introducerNode)
	if err != nil {
		return err
	}

	if err := n.initFinger(rn); err != nil {
		return err
	}

	// updateOthers()
	return nil
}

func (n *Server) closestPrecedingFinger(id ChordID) (Node, error) {
	for i := n.m - 1; i >= 0; i-- {
		if n.finger.entries[i].node.In(n.local.id, id, n.m) {
			nodeID := n.finger.entries[i].node
			n, ok := n.neighbourhood.Get(nodeID)
			if !ok {
				return Node{}, fmt.Errorf("node not found: %d", nodeID)
			}
			return n, nil
		}
	}
	return n.local, nil
}

func (n *Server) Serve() error {
	lis, err := net.Listen("tcp", n.local.bind)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	pb.RegisterChordServer(grpcServer, n)
	logrus.Info("serving chord grpc server at: ", n.local.bind)
	logrus.Infof("m: %d, nodeID: %d", n.m, n.local.id)
	return grpcServer.Serve(lis)
}

func NewServer(config Config) (*Server, error) {
	localNode := newNode(config.Bind, config.M)
	neighbourhood := newNeighbourhood(config.M)
	neighbourhood.Add(localNode)

	s := Server{
		m:    config.M,
		bind: config.Bind,

		neighbourhood: neighbourhood,
		local:         localNode,
	}

	s.finger = newFingerTable(s.local, s.m)
	return &s, nil
}
