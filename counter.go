package snitch

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

func NewCounter(name string, labels ...string) Counter {
	c := &counterMetric{
		gaugeMetric: gaugeMetric{
			description: NewDescription(name, MetricTypeCounter, labels...),
		},
	}
	c.selfCollector.self = c

	return c
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
