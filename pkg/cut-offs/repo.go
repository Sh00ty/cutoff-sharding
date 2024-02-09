package cutoffs

import (
	"context"
)

type CutOffRepo[K Key, V any] interface {
	Get(ctx context.Context, key K) (*CutOff[K, V], error)
	Create(ctx context.Context, cutoff CutOff[K, V]) error
	GetLast(ctx context.Context) (*CutOff[K, V], error)
	GetPrev(ctx context.Context, cutoffID uint64) (*CutOff[K, V], error)
}
