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
	"math/rand"
	"net"
	"time"
)

type PBNodeRef pb.Node

func (p *PBNodeRef) GetID() chord.ID {
	return chord.ID(p.Id)
}

func (p *PBNodeRef) GetBind() string {
	return p.Bind
}

func (p *PBNodeRef) String() string {
	return fmt.Sprintf("<P %d@%s>", p.GetID(), p.GetBind())
}

type Server struct {
	localNode           chord.LocalNode
	grpcServer          *grpc.Server
	stabilizationConfig StabilizationConfig
}

func (s *Server) X_Stabilize(ctx context.Context, _ *pb.StabilizeRequest) (*pb.StabilizeResponse, error) {
	numChanges, err := s.localNode.Stabilize(ctx)
	return &pb.StabilizeResponse{
		NumFingerTableEntryChanges: int32(numChanges),
	}, err
}

func (s *Server) SetPredecessorNode(ctx context.Context, req *pb.SetPredecessorNodeRequest) (*pb.SetPredecessorNodeResponse, error) {
	var nodeRef = PBNodeRef(*req.Node)
	err := s.localNode.SetPredNode(ctx, &nodeRef)
	return &pb.SetPredecessorNodeResponse{}, err
}

func (s *Server) SetSuccessorNode(ctx context.Context, req *pb.SetSuccessorNodeRequest) (*pb.SetSuccessorNodeResponse, error) {
	var nodeRef = PBNodeRef(*req.Node)
	err := s.localNode.SetSuccNode(ctx, &nodeRef)
	return &pb.SetSuccessorNodeResponse{}, err
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

	hops := request.Hops
	hops = append(hops, &pb.Hop{
		Id:   s.localNode.GetID().AsU64(),
		Bind: s.localNode.GetBind(),
	})
	n, err := s.localNode.FindPredecessor(ctx, chord.ID(request.Id))
	if err != nil {
		return nil, err
	}
	return &pb.FindPredecessorResponse{
		Node: n.AsProtobufNode(),
		Hops: hops,
	}, nil
}

func (s *Server) FindSuccessor(ctx context.Context, request *pb.FindSuccessorRequest) (*pb.FindSuccessorResponse, error) {
	logger := logrus.WithField("method", "Server.findSuccessor")
	logger.Debugf("id=%d", request.Id)

	hops := request.Hops
	hops = append(hops, &pb.Hop{
		Id: s.localNode.GetID().AsU64(),
		Bind: s.localNode.GetBind(),
	})

	n, err := s.localNode.FindSuccessor(ctx, chord.ID(request.Id))
	if err != nil {
		return nil, err
	}
	return &pb.FindSuccessorResponse{
		Node: n.AsProtobufNode(),
		Hops: hops,
	}, nil
}

func (s *Server) ClosestPrecedingFinger(ctx context.Context, request *pb.ClosestPrecedingFingerRequest) (*pb.ClosestPrecedingFingerResponse, error) {
	logger := logrus.WithField("method", "Server.closestPrecedingFinger")
	logger.Debugf("id=%d", request.Id)
	n, err := s.localNode.ClosestPrecedingFinger(ctx, chord.ID(request.Id))
	if err != nil {
		return nil, err
	}
	return &pb.ClosestPrecedingFingerResponse{
		Node: n.AsProtobufNode(),
	}, nil
}

func (s *Server) JoinRing(ctx context.Context, request *pb.JoinRingRequest) (*pb.JoinRingResponse, error) {
	logger := logrus.WithField("method", "Server.JoinRing")
	logger.WithField("introducer", request.Introducer.String()).Info("join request")
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
	logger.Infof("node=%v", request.Node)
	n, err := node.NewRemote(ctx, request.Node.Bind)
	if err != nil {
		return nil, err
	}
	if err := s.localNode.Notify(ctx, n); err != nil {
		return nil, err
	}
	return &pb.NotifyResponse{}, nil
}

func (s *Server) runStabilizer(ticker *time.Ticker) {
	for {
		select {
		case <-ticker.C:
			logrus.Info("Run Stabilize()")
			numChanges, err := s.localNode.Stabilize(context.Background())
			if err != nil {
				logrus.Error("Stabilize failed", err)
				continue
			}
			logrus.Infof("Number of finger table entries changed by stabilization: %d", numChanges)
		}
	}
}

func (s *Server) Serve() error {
	lis, err := net.Listen("tcp", s.localNode.GetBind())
	if err != nil {
		return err
	}

	pb.RegisterChordServer(s.grpcServer, s)
	logrus.Info("serving chord grpc server at: ", s.localNode.GetBind())
	logrus.Infof("nodeID: %d", s.localNode.GetID())

	if !s.stabilizationConfig.Disabled {
		rand.Seed(time.Now().UnixNano())
		jitter := time.Duration(rand.Int63() % s.stabilizationConfig.Jitter.Nanoseconds())
		runInterval := jitter + s.stabilizationConfig.Period
		logrus.Infof("jitter: %s, interval: %s", jitter, runInterval)
		tickerStabilize := time.NewTicker(runInterval)
		go s.runStabilizer(tickerStabilize)
	}

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
		localNode:           localNode,
		grpcServer:          grpcServer,
		stabilizationConfig: config.Stabilization,
	}
	return &s, nil
}
