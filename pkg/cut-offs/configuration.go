package cutoffs

import (
	"context"
)

type ConfType int8

const (
	ConsistentHashingType ConfType = iota + 1
	Vnodes
	Circle
)

type Configuration[K Key, V any] interface {
	Type() ConfType
	Clone() Configuration[K, V]
	GetNode(ctx context.Context, key K) (Node[K, V], error)
}
