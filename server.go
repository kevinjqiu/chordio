package chordio

import (
	"context"
	"fmt"
	"github.com/kevinjqiu/chordio/pb"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/plugin/grpctrace"
	"google.golang.org/grpc"
	"net"
)

var (
	telemetryServiceName = ""
)

type Server struct {
	*LocalNode
	grpcServer *grpc.Server
}

func (n *Server) GetNodeInfo(_ context.Context, _ *pb.GetNodeInfoRequest) (*pb.GetNodeInfoResponse, error) {
	logger := logrus.WithField("method", "Server.GetNodeInfo")
	logger.Debug("[Server] GetNodeInfo")
	return &pb.GetNodeInfoResponse{
		Node: n.LocalNode.AsProtobufNode(),
	}, nil
}

func (n *Server) FindPredecessor(ctx context.Context, request *pb.FindPredecessorRequest) (*pb.FindPredecessorResponse, error) {
	logger := logrus.WithField("method", "server.FindPredecessor")
	logger.Debug("id=", request.Id)
	var err error
	id := ChordID(request.Id)

	if !id.In(n.id, n.succ, n.m) {
		logger.Debugf("id is within %v, the predecessor is the local node", n.LocalNode)
		return &pb.FindPredecessorResponse{
			Node: n.LocalNode.AsProtobufNode(),
		}, nil
	}

	n_, err := n.closestPrecedingFinger(ctx, id)
	if err != nil {
		return nil, err
	}

	logger.Debugf("the closest preceding node is %v", n_)
	remoteNode, err := newRemoteNode(ctx, n_.GetBind())
	if err != nil {
		return nil, err
	}

	for {
		if !id.In(n_.GetID(), n_.succ, n.m) { // FIXME: not in (a, b]
			logger.Debugf("id is not in %v's range", n_)
			remoteNode, err = remoteNode.ClosestPrecedingFinger(ctx, id)
			if err != nil {
				return nil, err
			}
			logger.Debugf("the closest preceding node in %s's finger table is: ", remoteNode)
		} else {
			logger.Debugf("id is in %v's range", n_)
			break
		}
	}

	return &pb.FindPredecessorResponse{
		Node: remoteNode.AsProtobufNode(),
	}, nil
}

func (n *Server) FindSuccessor(ctx context.Context, request *pb.FindSuccessorRequest) (*pb.FindSuccessorResponse, error) {
	logger := logrus.WithField("method", "Server.FindSuccessor")
	logger.Debugf("id=%d", request.Id)
	resp, err := n.FindPredecessor(ctx, &pb.FindPredecessorRequest{
		Id: request.Id,
	})
	if err != nil {
		return nil, err
	}
	return &pb.FindSuccessorResponse{
		Node: resp.Node,
	}, nil
}

func (n *Server) ClosestPrecedingFinger(ctx context.Context, request *pb.ClosestPrecedingFingerRequest) (*pb.ClosestPrecedingFingerResponse, error) {
	logger := logrus.WithField("method", "Server.ClosestPrecedingFinger")
	logger.Debugf("id=%d", request.Id)
	node, err := n.closestPrecedingFinger(ctx, ChordID(request.Id))
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
	introNode, err := newRemoteNode(ctx, request.Introducer.Bind)
	if err != nil {
		return nil, err
	}
	if err := n.join(ctx, introNode); err != nil {
		return nil, err
	}
	return &pb.JoinRingResponse{}, nil
}

func (n *Server) Serve() error {
	lis, err := net.Listen("tcp", n.bind)
	if err != nil {
		return err
	}

	pb.RegisterChordServer(n.grpcServer, n)
	logrus.Info("serving chord grpc server at: ", n.bind)
	logrus.Infof("m: %d, nodeID: %d", n.m, n.id)
	return n.grpcServer.Serve(lis)
}

func NewServer(config Config) (*Server, error) {
	var err error

	localNode, err := newLocalNode(config.ID, config.Bind, config.M)
	if err != nil {
		return nil, errors.Wrap(err, "unable to initiate local node")
	}

	telemetryServiceName = fmt.Sprintf("chordio/id=%d", localNode.id)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpctrace.UnaryServerInterceptor(global.Tracer(telemetryServiceName))),
		grpc.StreamInterceptor(grpctrace.StreamServerInterceptor(global.Tracer(telemetryServiceName))),
	)

	s := Server{
		LocalNode:  localNode,
		grpcServer: grpcServer,
	}
	return &s, nil
}
