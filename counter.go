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

type CounterMeasure struct {
	Value float64
}

type counterMetric struct {
	gaugeMetric
	selfCollector
}

func NewCounter(name, help string, labels ...string) Counter {
	c := &counterMetric{
		gaugeMetric: gaugeMetric{
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

	c.gaugeMetric.Add(value)
}

func (c *counterMetric) Write(measure *Measure) error {
	measure.Counter = &CounterMeasure{
		Value: c.Value(),
	}

	return nil
}

func (c *counterMetric) With(labels ...string) Counter {
	return &counterMetric{
		gaugeMetric: gaugeMetric{
			bits:        c.bits,
			description: c.description,
		},
	}
}

func (c *counterMetric) Count() float64 {
	return c.Value()
}
