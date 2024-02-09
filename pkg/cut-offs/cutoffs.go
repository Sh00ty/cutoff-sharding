package cutoffs

import (
	"context"
)

type Key interface {
	GetSeqNum() uint64
}

type CutOff[K Key, V any] struct {
	ID    uint64
	Conf  Configuration[K, V]
	Bound uint64
}

type CutOffManager[K Key, V any] struct {
	repo              CutOffRepo[K, V]
	seqReactionOffset uint64
	generator         IDGenerator
}

func NewCutOffManager[K Key, V any](repo CutOffRepo[K, V], seqReactionOffset uint64, generator IDGenerator) *CutOffManager[K, V] {
	return &CutOffManager[K, V]{
		repo:              repo,
		seqReactionOffset: seqReactionOffset,
		generator:         generator,
	}
}

type CurSeq interface {
	Get(ctx context.Context) uint64
}

type IDGenerator interface {
	GenerateID() Key
}

func (m *CutOffManager[K, V]) Get(ctx context.Context, key K) (*V, error) {
	cutOff, err := m.repo.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	node, err := cutOff.Conf.GetNode(ctx, key)
	if err != nil {
		return nil, err
	}
	return node.Repo.Get(ctx, key)
}

func (m *CutOffManager[K, V]) Set(ctx context.Context, key K, val V) error {
	// can't take last one, due to reaction offset of new configurations
	cutOff, err := m.repo.Get(ctx, key)
	if err != nil {
		return err
	}
	node, err := cutOff.Conf.GetNode(ctx, key)
	if err != nil {
		return err
	}
	return node.Repo.Set(ctx, key, val)
}

func (m *CutOffManager[K, V]) CreateCutOff(ctx context.Context, conf Configuration[K, V]) error {
	cutOffNum := m.generator.GenerateID().GetSeqNum()

	return m.repo.Create(ctx, CutOff[K, V]{
		ID:    cutOffNum,
		Bound: cutOffNum + m.seqReactionOffset,
		Conf:  conf,
	})
}

func (m *CutOffManager[K, V]) GetLast(ctx context.Context) (*CutOff[K, V], error) {
	return m.repo.GetLast(ctx)
}

// это на всех отсечках
// TODO: предлагаю пока что удаление ноды реализовать в виде чистой замены одной
// физ ноды на другую, таким образом используем replace в конфигурации
// Cutoffs наверное не должна сама мигрировать тк от без есть свои супер плюшки
// Вохможно сделать option: WithExceptValues([]{CutOffID, configuration})
func (m *CutOffManager[K, V]) ReplaceNode(
	ctx context.Context,
	old NodeID,
	new NodeID,
	repo NodeRepo[K, V],
	exceptions ...Exception,
) error {
	return nil
}

type Exception struct {
}

type Iterator interface {
	Key
	Next() Iterator
}

func Merge[K Iterator, V any](
	ctx context.Context,
	from, to *CutOff[K, V],
	start K,
	limit uint64,
) error {
	// TODO: нужен какой-то мьютекс на это

	return nil
}
