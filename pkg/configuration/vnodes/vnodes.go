package vnodes

import (
	"context"
	"slices"

	cutoffs "github.com/Sh00ty/cutoff-sharding/pkg/cut-offs"
	"github.com/pkg/errors"
)

type Configuration[K cutoffs.Key, V any] struct {
	Nodes []cutoffs.Node[K, V]
}

func (Configuration[K, V]) Type() cutoffs.ConfType {
	return cutoffs.Vnodes
}

func (c *Configuration[K, V]) Clone() cutoffs.Configuration[K, V] {
	return &Configuration[K, V]{Nodes: slices.Clone(c.Nodes)}
}

func (c *Configuration[K, V]) GetNode(ctx context.Context, key K) (res cutoffs.Node[K, V], err error) {
	if len(c.Nodes) == 0 {
		return res, errors.New("not count is 0")
	}
	return c.Nodes[key.GetSeqNum()%uint64(len(c.Nodes))], nil
}

func (c *Configuration[K, V]) MergeTo(ctx context.Context, target cutoffs.Configuration[K, V]) error {
	return nil
}

func (c *Configuration[K, V]) RemoveNode(
	ctx context.Context,
	nodeID cutoffs.NodeID,
	helpValue interface{},
) (res cutoffs.Node[K, V], err error) {
	return
}

type Positions []uint

func (c *Configuration[K, V]) AddNode(ctx context.Context, node cutoffs.Node[K, V], helpValue interface{}) error {
	positions, ok := helpValue.(Positions)
	if !ok {
		return errors.New("Invalid help value type: expected positions")
	}
	for _, pos := range positions {
		c.Nodes[pos] = node
	}
	return errors.New("node doesn't exists")
}
