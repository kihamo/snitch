package snitch

import (
	"sync"

	"github.com/bsm/histogram"
)

var (
	Quantiles = []float64{0.5, 0.9, 0.99}
)

type Histogram interface {
	Metric
	Collector

	With(...string) Histogram
	Add(float64)
	Quantile(float64) float64
}

type HistogramMeasure struct {
	SampleCount    uint64
	SampleSum      float64
	SampleMin      float64
	SampleMax      float64
	SampleVariance float64
	Quantiles      map[float64]float64
}

type safeHistogram struct {
	sync.RWMutex
	histogram.Histogram
}

type histogramMetric struct {
	selfCollector

	description *Description
	histogram   *safeHistogram
	quantiles   []float64
}

func newSafeHistogram() *safeHistogram {
	return &safeHistogram{
		Histogram: *histogram.New(50),
	}
}

func (h *safeHistogram) Copy() *safeHistogram {
	h.RLock()
	defer h.RUnlock()

	return &safeHistogram{
		Histogram: *h.Histogram.Copy(nil),
	}
}

func (h *safeHistogram) Quantiles(quantiles []float64) map[float64]float64 {
	ret := make(map[float64]float64, len(quantiles))
	for _, q := range quantiles {
		ret[q] = h.Quantile(q)
	}

	return ret
}

func NewHistogram(name string, labels ...string) Histogram {
	return NewHistogramWithQuantiles(name, Quantiles, labels...)
}

func NewHistogramWithQuantiles(name string, quantiles []float64, labels ...string) Histogram {
	if len(quantiles) == 0 {
		quantiles = Quantiles
	}

	h := &histogramMetric{
		description: NewDescription(name, MetricTypeHistogram, labels...),
		histogram:   newSafeHistogram(),
		quantiles:   quantiles,
	}
	h.selfCollector.self = h
	return h
}

func (h *histogramMetric) Description() *Description {
	return h.description
}

func (h *histogramMetric) Write(measure *Measure) error {
	h.histogram.RLock()
	defer h.histogram.RUnlock()

	measure.Histogram = &HistogramMeasure{
		SampleCount:    uint64(h.histogram.Count()),
		SampleSum:      h.histogram.Sum(),
		SampleMin:      h.histogram.Min(),
		SampleMax:      h.histogram.Max(),
		SampleVariance: h.histogram.Variance(),
		Quantiles:      h.histogram.Quantiles(h.quantiles),
	}

	return nil
}

func (h *histogramMetric) With(labels ...string) Histogram {
	return &histogramMetric{
		description: h.description,
		histogram:   h.histogram.Copy(),
		quantiles:   h.quantiles,
	}
}

func (h *histogramMetric) Add(value float64) {
	h.histogram.Lock()
	defer h.histogram.Unlock()

	h.histogram.Add(value)
}

func (h *histogramMetric) Quantile(q float64) float64 {
	h.histogram.RLock()
	defer h.histogram.RUnlock()

	return h.histogram.Quantile(q)
}
