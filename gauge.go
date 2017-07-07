package snitch

import (
	"math"
	"sync/atomic"
)

type Gauge interface {
	Metric
	Collector

	With(...string) Gauge
	Set(float64)
	Add(float64)
	Sub(float64)
	Inc()
	Dec()
	Value() float64
}

type GaugeMeasure struct {
	Value float64
}

type gaugeMetric struct {
	selfCollector

	bits        uint64
	description *Description
}

func NewGauge(name, help string, labels ...string) Gauge {
	g := &gaugeMetric{
		description: NewDescription(name, help, MetricTypeGauge, labels...),
	}
	g.selfCollector.self = g

	return g
}

func (g *gaugeMetric) Description() *Description {
	return g.description
}

func (g *gaugeMetric) Write(measure *Measure) error {
	measure.Gauge = &GaugeMeasure{
		Value: g.Value(),
	}

	return nil
}

func (g *gaugeMetric) With(labels ...string) Gauge {
	return &gaugeMetric{
		bits:        g.bits,
		description: g.description,
	}
}

func (g *gaugeMetric) Set(value float64) {
	atomic.StoreUint64(&g.bits, math.Float64bits(value))
}

func (g *gaugeMetric) Add(value float64) {
	for {
		old := atomic.LoadUint64(&g.bits)
		new := math.Float64bits(math.Float64frombits(old) + value)
		if atomic.CompareAndSwapUint64(&g.bits, old, new) {
			return
		}
	}
}

func (g *gaugeMetric) Sub(value float64) {
	g.Add(value * -1)
}

func (g *gaugeMetric) Inc() {
	g.Add(1)
}

func (g *gaugeMetric) Dec() {
	g.Add(-1)
}

func (g *gaugeMetric) Value() float64 {
	return math.Float64frombits(atomic.LoadUint64(&g.bits))
}
