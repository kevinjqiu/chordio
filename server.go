package chordio

import (
	context "context"
	"github.com/kevinjqiu/chordio/pb"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
)

type Server struct {
	m    Rank
	bind string

	local  Node
	finger FingerTable
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
	panic("implement me")
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
	n.initFinger(rn)

	return nil
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
	s := Server{
		m:    config.M,
		bind: config.Bind,

		local: newNode(config.Bind, config.M),
	}

	s.finger = newFingerTable(s.local, s.m)
	return &s, nil
}
