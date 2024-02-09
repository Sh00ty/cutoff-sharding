package snowflake

import (
	"encoding/binary"
	"sync/atomic"

	cutoffs "github.com/Sh00ty/cutoff-sharding/pkg/cut-offs"
	"github.com/cespare/xxhash/v2"
)

type ID uint64

func (id ID) GetSeqNum() uint64 {
	return id.GetTime()
}

func (id ID) Hash() uint64 {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(id))
	hash := uint64(xxhash.Sum64(buf))
	return hash
}

const (
	countLen ID = 8
	podLen   ID = 12
	timeLen  ID = 44

	countMask ID = 1<<countLen - 1
	podMask   ID = 1<<podLen - 1
	timeMask  ID = 1<<timeLen - 1

	podOffset  ID = countLen
	timeOffset ID = podOffset + podLen
)

func newID() ID {
	return ID(0)
}
func (id ID) setCount(count int32) ID {
	return id | ID(count)&countMask
}

func (id ID) GetCount() uint8 {
	return uint8(id & countMask)
}

func (id ID) setPod(pid uint16) ID {
	return id | (ID(pid)&podMask)<<podOffset
}

func (id ID) GetPod() uint16 {
	return uint16(id & (podMask << podOffset) >> podOffset)
}

func (id ID) setTime(time uint64) ID {
	return id | (ID(time)&timeMask)<<timeOffset
}

func (id ID) GetTime() uint64 {
	return uint64((id & (timeMask << timeOffset)) >> timeOffset)
}

func NewGenerator(podID uint16, timer Timer) *Generator {
	return &Generator{
		podID:   podID,
		timer:   timer,
		counter: &atomic.Int32{},
	}
}

type Generator struct {
	counter *atomic.Int32
	podID   uint16
	timer   Timer
}

type Timer interface {
	GetTime() uint64
}

func (g *Generator) GenerateID() cutoffs.Key {
	time := g.timer.GetTime()
	return newID().
		setCount(g.counter.Add(1)).
		setPod(g.podID).
		setTime(time)
}
