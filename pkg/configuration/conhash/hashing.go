package conhash

import cutoffs "github.com/Sh00ty/cutoff-sharding/pkg/cut-offs"

type Point[K cutoffs.Key, V any] struct {
	Node cutoffs.Node[K, V]
}

type ConsistentHashing[K cutoffs.Key, V any] struct {
	circle []Point[K, V]
}
