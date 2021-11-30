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

	Add(float64)
	Quantile(float64) float64

	With(...string) Histogram
}

type histogramMetric struct {
	Vector

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

	metric := &histogramMetric{
		description: NewDescription(name, help, MetricTypeHistogram, labels...),
		histogram:   internal.NewSafeHistogram(),
		quantiles:   quantiles,
	}
	metric.SetMetric(metric).SetCreator(func(l ...string) Metric {
		return NewHistogramWithQuantiles(name, help, quantiles, append(labels, l...)...)
	})

	return metric
}

func (h *histogramMetric) Description() *Description {
	return h.description
}

func (h *histogramMetric) Measure() (*MeasureValue, error) {
	h.histogram.RLock()
	defer h.histogram.RUnlock()

	return &MeasureValue{
		SampleCount:    Uint64(uint64(h.histogram.Count())),
		SampleSum:      Float64(h.histogram.Sum()),
		SampleMin:      Float64(h.histogram.Min()),
		SampleMax:      Float64(h.histogram.Max()),
		SampleVariance: Float64(h.histogram.Variance()),
		Quantiles:      Float64Map(h.histogram.Quantiles(h.quantiles)),
	}, nil
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

func (h *histogramMetric) With(labels ...string) Histogram {
	return h.Vector.With(labels...).(Histogram)
}
