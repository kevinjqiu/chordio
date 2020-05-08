package node

import (
	"context"
	"github.com/kevinjqiu/chordio/chord"
)

type factory interface {
	newLocalNode(id chord.ID, bind string, m chord.Rank) (LocalNode, error)
	newRemoteNode(ctx context.Context, bind string) (RemoteNode, error)
}

type defaultFactory struct{}

func (d defaultFactory) newLocalNode(id chord.ID, bind string, m chord.Rank) (LocalNode, error) {
	return NewLocal(id, bind, m)
}
func (d defaultFactory) newRemoteNode(ctx context.Context, bind string) (RemoteNode, error) {
	return NewRemote(ctx, bind)
}
