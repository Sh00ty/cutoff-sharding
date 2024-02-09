package inmem

import (
	"context"
	"errors"
	"slices"
	"sync"

	cutoffs "github.com/Sh00ty/cutoff-sharding/pkg/cut-offs"
)

type InMemRepo[K cutoffs.Key, V any] struct {
	mu      *sync.RWMutex
	cutoffs []cutoffs.CutOff[K, V]
}

func NewInMemRepo[K cutoffs.Key, V any]() *InMemRepo[K, V] {
	return &InMemRepo[K, V]{
		mu: &sync.RWMutex{},
	}
}

func (repo InMemRepo[K, V]) Get(ctx context.Context, key K) (*cutoffs.CutOff[K, V], error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	if len(repo.cutoffs) == 0 {
		return nil, errors.New("cutoff doesn't exist: len(cutoffs)=0")
	}

	seqNum := key.GetSeqNum()
	// write optimization
	if repo.cutoffs[len(repo.cutoffs)-1].Bound < seqNum {
		return &repo.cutoffs[len(repo.cutoffs)-1], nil
	}
	l := -1
	r := len(repo.cutoffs)
	for l < r-1 {
		m := (l + r) / 2
		cut := repo.cutoffs[m]
		if cut.Bound <= seqNum && (m == len(repo.cutoffs)-1 || seqNum < repo.cutoffs[m+1].Bound) {
			return &cut, nil
		}
		if seqNum < cut.Bound {
			r = m
			continue
		}
		l = m
	}
	return nil, errors.New("cutoff doesn't exist")
}

func (r *InMemRepo[K, V]) Create(ctx context.Context, cutoff cutoffs.CutOff[K, V]) error {
	r.mu.Lock()
	r.cutoffs = append(r.cutoffs, cutoff)
	r.mu.Unlock()
	return nil
}

func (r *InMemRepo[K, V]) GetLast(ctx context.Context) (*cutoffs.CutOff[K, V], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.cutoffs) == 0 {
		return nil, errors.New("cur cutoff doesn't exist")
	}
	return &r.cutoffs[len(r.cutoffs)-1], nil
}

func idCmp[K cutoffs.Key, V any](co cutoffs.CutOff[K, V], u uint64) int {
	if co.ID == u {
		return 0
	}
	if co.ID < u {
		return -1
	}
	return 1
}

func (r *InMemRepo[K, V]) GetPrev(ctx context.Context, cutoffID uint64) (*cutoffs.CutOff[K, V], error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ind, ok := slices.BinarySearchFunc(r.cutoffs, cutoffID, idCmp)
	if !ok || ind < 1 {
		return nil, errors.New("can't find cutoff by ID")
	}

	return &r.cutoffs[ind-1], nil
}
