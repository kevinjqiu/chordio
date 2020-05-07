package chordio

import (
	"context"
	"github.com/kevinjqiu/chordio/chord"
	"github.com/kevinjqiu/chordio/chord/node"
	"github.com/kevinjqiu/chordio/pb"
	"github.com/kevinjqiu/chordio/telemetry"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/plugin/grpctrace"
	"google.golang.org/grpc"
	"net"
)

type Server struct {
	localNode  *node.LocalNode
	grpcServer *grpc.Server
}

func (n *Server) GetNodeInfo(_ context.Context, req *pb.GetNodeInfoRequest) (*pb.GetNodeInfoResponse, error) {
	logger := logrus.WithField("method", "Server.GetNodeInfo")
	logger.Debug("[Server] GetNodeInfo")

	var ft *pb.FingerTable
	if req.IncludeFingerTable {
		ft = n.localNode.GetFingerTable().AsProtobufFT()
	}
	return &pb.GetNodeInfoResponse{
		Node: n.localNode.AsProtobufNode(),
		Ft:   ft,
	}, nil
}

func (n *Server) FindPredecessor(ctx context.Context, request *pb.FindPredecessorRequest) (*pb.FindPredecessorResponse, error) {
	logger := logrus.WithField("method", "server.findPredecessor")
	logger.Debug("id=", request.Id)

	node, err := n.localNode.FindPredecessor(ctx, chord.ID(request.Id))
	if err != nil {
		return nil, err
	}
	return &pb.FindPredecessorResponse{
		Node: node.AsProtobufNode(),
	}, nil
}

func (n *Server) FindSuccessor(ctx context.Context, request *pb.FindSuccessorRequest) (*pb.FindSuccessorResponse, error) {
	logger := logrus.WithField("method", "Server.findSuccessor")
	logger.Debugf("id=%d", request.Id)
	node, err := n.localNode.FindSuccessor(ctx, chord.ID(request.Id))
	if err != nil {
		return nil, err
	}
	return &pb.FindSuccessorResponse{
		Node: node.AsProtobufNode(),
	}, nil
}

func (n *Server) ClosestPrecedingFinger(ctx context.Context, request *pb.ClosestPrecedingFingerRequest) (*pb.ClosestPrecedingFingerResponse, error) {
	logger := logrus.WithField("method", "Server.closestPrecedingFinger")
	logger.Debugf("id=%d", request.Id)
	node, err := n.localNode.ClosestPrecedingFinger(ctx, chord.ID(request.Id))
	if err != nil {
		return nil, err
	}
	return &pb.ClosestPrecedingFingerResponse{
		Node: node.AsProtobufNode(),
	}, nil
}

func (n *Server) JoinRing(ctx context.Context, request *pb.JoinRingRequest) (*pb.JoinRingResponse, error) {
	logger := logrus.WithField("method", "Server.JoinRing")
	logger.Debugf("introducer=%v", request.Introducer)
	introNode, err := node.NewRemote(ctx, request.Introducer.Bind)
	if err != nil {
		return nil, err
	}
	if err := n.localNode.Join(ctx, introNode); err != nil {
		return nil, err
	}
	return &pb.JoinRingResponse{}, nil
}

func (n *Server) UpdateFingerTable(ctx context.Context, request *pb.UpdateFingerTableRequest) (*pb.UpdateFingerTableResponse, error) {
	logger := logrus.WithField("method", "Server.UpdateFingerTable")
	logger.Debugf("node=%v, i=%d", request.Node, request.I)
	node, err := node.NewLocal(chord.ID(request.Node.Id), request.Node.Bind, n.localNode.GetRank())
	if err != nil {
		return nil, err
	}
	if err := n.localNode.UpdateFingerTableEntry(ctx, node, int(request.I)); err != nil {
		return nil, err
	}

	return &pb.UpdateFingerTableResponse{}, nil
}

func (n *Server) Serve() error {
	lis, err := net.Listen("tcp", n.localNode.GetBind())
	if err != nil {
		return err
	}

	pb.RegisterChordServer(n.grpcServer, n)
	logrus.Info("serving chord grpc server at: ", n.localNode.GetBind())
	logrus.Infof("nodeID: %d", n.localNode.GetID())
	return n.grpcServer.Serve(lis)
}

func NewServer(config Config) (*Server, error) {
	var err error

	localNode, err := node.NewLocal(config.ID, config.Bind, config.M)
	if err != nil {
		return nil, errors.Wrap(err, "unable to initiate local node")
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpctrace.UnaryServerInterceptor(global.Tracer(telemetry.GetServiceName()))),
		grpc.StreamInterceptor(grpctrace.StreamServerInterceptor(global.Tracer(telemetry.GetServiceName()))),
	)

	s := Server{
		localNode:  localNode,
		grpcServer: grpcServer,
	}
	return &s, nil
}
