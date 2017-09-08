package snitch

import (
	"errors"
)

type Counter interface {
	Metric
	Collector

	With(...string) Counter
	Add(float64)
	Inc()
	Count() float64
}

type counterMetric struct {
	untypedMetric
	selfCollector
}

func NewCounter(name, help string, labels ...string) Counter {
	c := &counterMetric{
		untypedMetric: untypedMetric{
			description: NewDescription(name, help, MetricTypeCounter, labels...),
		},
	}
	c.selfCollector.self = c

	return c
}

func (c *counterMetric) Add(value float64) {
	if value < 0 {
		panic(errors.New("value can't be less than zero"))
	}

	c.untypedMetric.Add(value)
}

func (c *counterMetric) With(labels ...string) Counter {
	return &counterMetric{
		untypedMetric: untypedMetric{
			bits:        c.bits,
			description: c.description,
		},
	}
}

func (c *counterMetric) Count() float64 {
	return c.Value()
}
