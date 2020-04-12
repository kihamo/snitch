package snitch

import (
	"math"
	"sync/atomic"
)

type Untyped interface {
	Metric
	Collector

	Set(float64)
	Add(float64)
	Sub(float64)
	Inc()
	Dec()
	Value() float64

	With(...string) Untyped
}

type untypedMetric struct {
	valueBits       uint64
	sampleCountBits uint64

	vector

	description *Description
}

func NewUntyped(name, help string, labels ...string) Untyped {
	u := &untypedMetric{
		description: NewDescription(name, help, MetricTypeUntyped, labels...),
	}
	u.init(u, func(l ...string) Metric {
		return NewUntyped(name, help, append(labels, l...)...)
	})

	return u
}

func (u *untypedMetric) Description() *Description {
	return u.description
}

func (u *untypedMetric) Measure() (*MeasureValue, error) {
	return &MeasureValue{
		Value:       Float64(u.Value()),
		SampleCount: Uint64(u.SampleCount()),
	}, nil
}

func (u *untypedMetric) Set(value float64) {
	atomic.StoreUint64(&u.valueBits, math.Float64bits(value))
	atomic.AddUint64(&u.sampleCountBits, 1)
}

func (u *untypedMetric) Add(value float64) {
	for {
		old := atomic.LoadUint64(&u.valueBits)
		current := math.Float64bits(math.Float64frombits(old) + value)

		if atomic.CompareAndSwapUint64(&u.valueBits, old, current) {
			atomic.AddUint64(&u.sampleCountBits, 1)
			return
		}
	}
}

func (u *untypedMetric) Sub(value float64) {
	u.Add(value * -1)
}

func (u *untypedMetric) Inc() {
	u.Add(1)
}

func (u *untypedMetric) Dec() {
	u.Add(-1)
}

func (u *untypedMetric) Value() float64 {
	return math.Float64frombits(atomic.LoadUint64(&u.valueBits))
}

func (u *untypedMetric) SampleCount() uint64 {
	return atomic.LoadUint64(&u.sampleCountBits)
}

func (u *untypedMetric) With(labels ...string) Untyped {
	return u.vector.With(labels...).(Untyped)
}
