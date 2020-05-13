package chordio

import (
	"context"
	"fmt"
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

type PBNodeRef pb.Node

func (p *PBNodeRef) GetID() chord.ID {
	return chord.ID(p.Id)
}

func (p *PBNodeRef) GetBind() string {
	return p.Bind
}

func (p *PBNodeRef) String() string {
	return fmt.Sprintf("<PB @%d %s>", p.GetID(), p.GetBind())
}

type Server struct {
	localNode     chord.LocalNode
	grpcServer    *grpc.Server
}

func (s *Server) X_Stabilize(ctx context.Context, _ *pb.StabilizeRequest) (*pb.StabilizeResponse, error) {
	err := s.localNode.Stabilize(ctx)
	return &pb.StabilizeResponse{}, err
}

func (s *Server) X_FixFingers(ctx context.Context, _ *pb.FixFingersRequest) (*pb.FixFingersResponse, error) {
	err := s.localNode.FixFingers(ctx)
	return &pb.FixFingersResponse{}, err
}

func (s *Server) SetPredecessorNode(ctx context.Context, req *pb.SetPredecessorNodeRequest) (*pb.SetPredecessorNodeResponse, error) {
	var nodeRef = PBNodeRef(*req.Node)
	s.localNode.SetPredNode(ctx, &nodeRef)
	return &pb.SetPredecessorNodeResponse{}, nil
}

func (s *Server) SetSuccessorNode(ctx context.Context, req *pb.SetSuccessorNodeRequest) (*pb.SetSuccessorNodeResponse, error) {
	var nodeRef = PBNodeRef(*req.Node)
	s.localNode.SetSuccNode(ctx, &nodeRef)
	return &pb.SetSuccessorNodeResponse{}, nil
}

func (s *Server) GetNodeInfo(_ context.Context, req *pb.GetNodeInfoRequest) (*pb.GetNodeInfoResponse, error) {
	logger := logrus.WithField("method", "Server.GetNodeInfo")
	logger.Debug("[Server] GetNodeInfo")

	var ft *pb.FingerTable
	if req.IncludeFingerTable {
		ft = s.localNode.GetFingerTable().AsProtobufFT()
	}
	return &pb.GetNodeInfoResponse{
		Node: s.localNode.AsProtobufNode(),
		Ft:   ft,
	}, nil
}

func (s *Server) FindPredecessor(ctx context.Context, request *pb.FindPredecessorRequest) (*pb.FindPredecessorResponse, error) {
	logger := logrus.WithField("method", "server.findPredecessor")
	logger.Debug("id=", request.Id)

	node, err := s.localNode.FindPredecessor(ctx, chord.ID(request.Id))
	if err != nil {
		return nil, err
	}
	return &pb.FindPredecessorResponse{
		Node: node.AsProtobufNode(),
	}, nil
}

func (s *Server) FindSuccessor(ctx context.Context, request *pb.FindSuccessorRequest) (*pb.FindSuccessorResponse, error) {
	logger := logrus.WithField("method", "Server.findSuccessor")
	logger.Debugf("id=%d", request.Id)
	node, err := s.localNode.FindSuccessor(ctx, chord.ID(request.Id))
	if err != nil {
		return nil, err
	}
	return &pb.FindSuccessorResponse{
		Node: node.AsProtobufNode(),
	}, nil
}

func (s *Server) ClosestPrecedingFinger(ctx context.Context, request *pb.ClosestPrecedingFingerRequest) (*pb.ClosestPrecedingFingerResponse, error) {
	logger := logrus.WithField("method", "Server.closestPrecedingFinger")
	logger.Debugf("id=%d", request.Id)
	node, err := s.localNode.ClosestPrecedingFinger(ctx, chord.ID(request.Id))
	if err != nil {
		return nil, err
	}
	return &pb.ClosestPrecedingFingerResponse{
		Node: node.AsProtobufNode(),
	}, nil
}

func (s *Server) JoinRing(ctx context.Context, request *pb.JoinRingRequest) (*pb.JoinRingResponse, error) {
	logger := logrus.WithField("method", "Server.JoinRing")
	logger.Debugf("introducer=%v", request.Introducer)
	introNode, err := node.NewRemote(ctx, request.Introducer.Bind)
	if err != nil {
		return nil, err
	}
	if err := s.localNode.Join(ctx, introNode); err != nil {
		return nil, err
	}
	return &pb.JoinRingResponse{}, nil
}

func (s *Server) Notify(ctx context.Context, request *pb.NotifyRequest) (*pb.NotifyResponse, error) {
	logger := logrus.WithField("method", "Server.Notify")
	logger.Debugf("node=%v", request.Node)
	//node, err := node.NewLocal(chord.ID(request.Node.Id), request.Node.Bind, s.localNode.GetRank())
	n, err := node.NewRemote(ctx, request.Node.Bind)
	if err != nil {
		return nil, err
	}
	if err := s.localNode.Notify(ctx, n); err != nil {
		return nil, err
	}
	return &pb.NotifyResponse{}, nil
}

func (s *Server) Serve() error {
	lis, err := net.Listen("tcp", s.localNode.GetBind())
	if err != nil {
		return err
	}

	pb.RegisterChordServer(s.grpcServer, s)
	logrus.Info("serving chord grpc server at: ", s.localNode.GetBind())
	logrus.Infof("nodeID: %d", s.localNode.GetID())

	//jitter := rand.Int() % 3000
	//tickerStabilize := time.Tick(3 * time.Second + (time.Duration(jitter) * time.Millisecond))
	//tickerFixFingers := time.Tick(5 * time.Second + (time.Duration(jitter) * time.Millisecond))
	//
	//go func() {
	//	for {
	//		select {
	//		case <-tickerStabilize:
	//			logrus.Info("Run Stabilize()")
	//			if err := s.localNode.Stabilize(context.Background()); err != nil {
	//				logrus.Error("Stabilize failed", err)
	//			}
	//		case <-tickerFixFingers:
	//			logrus.Info("Run FixFingers()")
	//			if err := s.localNode.FixFingers(context.Background()); err != nil {
	//				logrus.Error("FixFinger failed", err)
	//			}
	//		}
	//	}
	//}()
	return s.grpcServer.Serve(lis)
}

func (s *Server) GracefulStop() {
	logrus.Infof("Stopping server: %s", s.localNode.String())
	s.grpcServer.GracefulStop()
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
		localNode:     localNode,
		grpcServer:    grpcServer,
	}
	return &s, nil
}
