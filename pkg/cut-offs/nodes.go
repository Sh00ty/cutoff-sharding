package cutoffs

import (
	"context"
)

type NodeID string

type Node[K Key, V any] struct {
	ID   NodeID
	Repo NodeRepo[K, V]
}

type NodeRepo[K Key, V any] interface {
	Get(ctx context.Context, key K) (*V, error)
	Set(ctx context.Context, key K, val V) error
	Remove(ctx context.Context, key K) error
}
