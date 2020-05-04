package chordio

import (
	"context"
	"fmt"
	"github.com/kevinjqiu/chordio/pb"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"strings"
)

type Server struct {
	*LocalNode
	grpcServer *grpc.Server
}

func (n *Server) GetNodeInfo(_ context.Context, _ *pb.GetNodeInfoRequest) (*pb.GetNodeInfoResponse, error) {
	return &pb.GetNodeInfoResponse{
		Node: newPBNodeFromLocalNode(*n.LocalNode),
	}, nil
}

func (n *Server) FindPredecessor(_ context.Context, request *pb.FindPredecessorRequest) (*pb.FindPredecessorResponse, error) {
	var err error
	id := ChordID(request.Id)

	if !id.In(n.id, n.succ, n.m) {
		return &pb.FindPredecessorResponse{
			Node: newPBNodeFromLocalNode(*n.LocalNode),
		}, nil
	}

	n_, err := n.closestPrecedingFinger(id)
	if err != nil {
		return nil, err
	}

	remoteNode, err := newRemoteNode(n_.GetBind())
	if err != nil {
		return nil, err
	}

	for {
		if !id.In(n_.GetID(), n_.succ, n.m) { // FIXME: not in (a, b]
			remoteNode, err = remoteNode.ClosestPrecedingFinger(id)
			if err != nil {
				return nil, err
			}
		} else {
			break
		}
	}

	// TODO: newPBNodeFromRemoteNode?
	// how does the remote node know its pred/succ?
	// answer: through GetNodeInfo
	//pred, err := remoteNode.getPredNode()
	//if err != nil {
	//	return nil, err
	//}
	//succ, err := remoteNode.getSuccNode()
	//if err != nil {
	//	return nil, err
	//}
	return &pb.FindPredecessorResponse{
		Node: &pb.Node{
			Id:   uint64(remoteNode.id),
			Bind: remoteNode.bind,
			//Pred: newPBNodeFromLocalNode(pred),
			//Succ: newPBNodeFromLocalNode(succ),
		},
	}, nil
}

func (n *Server) FindSuccessor(ctx context.Context, request *pb.FindSuccessorRequest) (*pb.FindSuccessorResponse, error) {
	resp, err := n.FindPredecessor(ctx, &pb.FindPredecessorRequest{
		Id: request.KeyID,
	})
	if err != nil {
		return nil, err
	}
	return &pb.FindSuccessorResponse{
		Node: resp.Node,
	}, nil
}

func (n *Server) ClosestPrecedingFinger(_ context.Context, request *pb.ClosestPrecedingFingerRequest) (*pb.ClosestPrecedingFingerResponse, error) {
	node, err := n.closestPrecedingFinger(ChordID(request.Id))
	if err != nil {
		return nil, err
	}
	return &pb.ClosestPrecedingFingerResponse{
		Node: newPBNodeFromLocalNode(node),
	}, nil
}

func (n *Server) JoinRing(_ context.Context, request *pb.JoinRingRequest) (*pb.JoinRingResponse, error) {
	logrus.Info("JoinRing: introducer=", request.Introducer)
	introNode, err := newRemoteNode(request.Introducer.Bind)
	if err != nil {
		return nil, err
	}
	if err := n.join(introNode); err != nil {
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

	parts := strings.Split(config.Bind, ":")
	if len(parts) != 2 {
		return nil, errInvalidBindFormat
	}

	ip := parts[0]
	if ip == "" {
		ip, err = getFirstAvailableBindIP()
		if err != nil  {
			return nil, errors.Wrap(errUnableToGetBindIP, err.Error())
		}
	} else {
		if !canBindIP(ip) {
			return nil, errors.Wrap(errInvalidBindIP, ip)
		}
	}
	bind := fmt.Sprintf("%s:%s", ip, parts[1])

	localNode := newLocalNode(bind, config.M)

	grpcServer := grpc.NewServer()

	s := Server{
		LocalNode:  localNode,
		grpcServer: grpcServer,
	}
	return &s, nil
}
