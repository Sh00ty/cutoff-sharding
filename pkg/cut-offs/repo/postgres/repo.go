package postgres

import cutoffs "github.com/Sh00ty/cutoff-sharding/pkg/cut-offs"

type PgRepo[K cutoffs.Key, V any] struct {
	Current cutoffs.CutOff[K, V]
}
