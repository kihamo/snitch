package snitch

import (
	"github.com/kihamo/snitch/internal"
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

type histogramMetric struct {
	selfCollector

	description *Description
	histogram   *internal.SafeHistogram
	quantiles   []float64
}

func NewHistogram(name, help string, labels ...string) Histogram {
	return NewHistogramWithQuantiles(name, help, Quantiles, labels...)
}

func NewHistogramWithQuantiles(name, help string, quantiles []float64, labels ...string) Histogram {
	if len(quantiles) == 0 {
		quantiles = Quantiles
	}

	h := &histogramMetric{
		description: NewDescription(name, help, MetricTypeHistogram, labels...),
		histogram:   internal.NewSafeHistogram(),
		quantiles:   quantiles,
	}
	h.selfCollector.self = h
	return h
}

func (h *histogramMetric) Description() *Description {
	return h.description
}

func (h *histogramMetric) Measure() (*MeasureValue, error) {
	h.histogram.RLock()
	defer h.histogram.RUnlock()

	return &MeasureValue{
		SampleCount:    &(&struct{ v uint64 }{uint64(h.histogram.Count())}).v,
		SampleSum:      &(&struct{ v float64 }{h.histogram.Sum()}).v,
		SampleMin:      &(&struct{ v float64 }{h.histogram.Min()}).v,
		SampleMax:      &(&struct{ v float64 }{h.histogram.Max()}).v,
		SampleVariance: &(&struct{ v float64 }{h.histogram.Variance()}).v,
		Quantiles:      &(&struct{ v map[float64]float64 }{h.histogram.Quantiles(h.quantiles)}).v,
	}, nil
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
