package snitch

import (
	"errors"
)

type Counter interface {
	Metric
	Collector

	Add(float64)
	Inc()
	Count() float64

	With(...string) Counter
}

type counterMetric struct {
	untypedMetric
}

func NewCounter(name, help string, labels ...string) Counter {
	c := &counterMetric{
		untypedMetric: untypedMetric{
			description: NewDescription(name, help, MetricTypeCounter, labels...),
		},
	}
	c.init(c, func(l ...string) Metric {
		return NewCounter(name, help, append(labels, l...)...)
	})

	return c
}

func (c *counterMetric) Add(value float64) {
	if value < 0 {
		panic(errors.New("value can't be less than zero"))
	}

	c.untypedMetric.Add(value)
}

func (c *counterMetric) Count() float64 {
	return c.Value()
}

func (c *counterMetric) With(labels ...string) Counter {
	return c.vector.With(labels...).(Counter)
}
