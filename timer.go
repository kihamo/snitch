package snitch

import (
	"time"

	"github.com/kihamo/snitch/internal"
)

type Timer interface {
	Metric
	Collector

	With(...string) Timer
	Update(time.Duration)
	UpdateSince(time.Time)
	Time()
	Quantile(float64) float64
}

type timerMetric struct {
	selfCollector
	histogramMetric

	begin time.Time
}

func NewTimer(name, help string, labels ...string) Timer {
	return NewTimerWithQuantiles(name, help, Quantiles, labels...)
}

func NewTimerWithQuantiles(name, help string, quantiles []float64, labels ...string) Timer {
	if len(quantiles) == 0 {
		quantiles = Quantiles
	}

	t := &timerMetric{
		histogramMetric: histogramMetric{
			description: NewDescription(name, help, MetricTypeTimer, labels...),
			histogram:   internal.NewSafeHistogram(),
			quantiles:   quantiles,
		},
		begin: time.Now(),
	}
	t.selfCollector.self = t
	return t
}

func (t *timerMetric) With(labels ...string) Timer {
	return &timerMetric{
		histogramMetric: histogramMetric{
			description: t.description,
			histogram:   t.histogram.Copy(),
		},
		begin: t.begin,
	}
}

func (t *timerMetric) Update(d time.Duration) {
	t.Add(d.Seconds())
}

func (t *timerMetric) UpdateSince(ts time.Time) {
	d := time.Since(ts).Seconds()
	if d < 0 {
		d = 0
	}

	t.Add(d)
}

func (t *timerMetric) Time() {
	t.UpdateSince(t.begin)
}
