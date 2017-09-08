package snitch

import (
	"math"
	"sync/atomic"
)

type Untyped interface {
	Metric
	Collector

	With(...string) Untyped
	Set(float64)
	Add(float64)
	Sub(float64)
	Inc()
	Dec()
	Value() float64
}

type untypedMetric struct {
	selfCollector

	bits        uint64
	description *Description
}

func NewUntyped(name, help string, labels ...string) Untyped {
	g := &untypedMetric{
		description: NewDescription(name, help, MetricTypeUntyped, labels...),
	}
	g.selfCollector.self = g

	return g
}

func (u *untypedMetric) Description() *Description {
	return u.description
}

func (u *untypedMetric) Measure() (*MeasureValue, error) {
	return &MeasureValue{
		Value: &(&struct{ v float64 }{u.Value()}).v,
	}, nil
}

func (u *untypedMetric) With(labels ...string) Untyped {
	return &untypedMetric{
		bits:        u.bits,
		description: u.description,
	}
}

func (u *untypedMetric) Set(value float64) {
	atomic.StoreUint64(&u.bits, math.Float64bits(value))
}

func (u *untypedMetric) Add(value float64) {
	for {
		old := atomic.LoadUint64(&u.bits)
		new := math.Float64bits(math.Float64frombits(old) + value)
		if atomic.CompareAndSwapUint64(&u.bits, old, new) {
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
	return math.Float64frombits(atomic.LoadUint64(&u.bits))
}
