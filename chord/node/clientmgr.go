package node

import (
	"github.com/kevinjqiu/chordio/pb"
	"github.com/kevinjqiu/chordio/telemetry"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/plugin/grpctrace"
	"google.golang.org/grpc"
	"sync"
)

type clientManager struct {
	mu *sync.Mutex
	clients map[string]pb.ChordClient
}

func (cm *clientManager) newCient(bind string) (pb.ChordClient, error) {
	conn, err := grpc.Dial(bind,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(grpctrace.UnaryClientInterceptor(global.Tracer(telemetry.GetServiceName()))),
		grpc.WithStreamInterceptor(grpctrace.StreamClientInterceptor(global.Tracer(telemetry.GetServiceName()))),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to initiate grpc client for node: %v", bind)
	}

	client := pb.NewChordClient(conn)
	return client, nil
}

func (cm *clientManager) getClient(bind string) (pb.ChordClient, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	conn, ok := cm.clients[bind]
	if !ok {
		conn, err := cm.newCient(bind)
		if err != nil {
			return nil, err
		}
		cm.clients[bind] = conn
	}

	return conn, nil
}

func newClientManager() *clientManager {
	return &clientManager{
		mu: new(sync.Mutex),
		clients: make(map[string]pb.ChordClient),
	}
}
