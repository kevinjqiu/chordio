package attrs

import (
	"github.com/kevinjqiu/chordio/chord"
	"github.com/kevinjqiu/chordio/chord/node"
	"go.opentelemetry.io/otel/api/core"
)

func Node(key string, node node.NodeRef) core.KeyValue {
	nodeStr := "<nil>"
	if node != nil {
		nodeStr = node.String()
	}

	return core.Key(key).String(nodeStr)
}

func ID(key string, id chord.ID) core.KeyValue {
	return core.Key(key).Int(id.AsInt())
}
