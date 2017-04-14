package snitch

import (
	"sync"

	"github.com/bsm/histogram"
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
}

type safeHistogram struct {
	sync.RWMutex
	histogram.Histogram
}

type histogramMetric struct {
	selfCollector

	description *Description
	histogram   *safeHistogram
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

func NewHistogram(name string, labels ...string) Histogram {
	h := &histogramMetric{
		description: NewDescription(name, MetricTypeHistogram, labels...),
		histogram:   newSafeHistogram(),
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
	}

	return nil
}

func (h *histogramMetric) With(labels ...string) Histogram {
	return &histogramMetric{
		description: h.description,
		histogram:   h.histogram.Copy(),
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
