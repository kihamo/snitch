package snitch

type Gauge interface {
	Metric
	Collector

	Set(float64)
	Add(float64)
	Sub(float64)
	Inc()
	Dec()
	Value() float64

	With(...string) Gauge
}

type gaugeMetric struct {
	untypedMetric
}

func NewGauge(name, help string, labels ...string) Gauge {
	g := &gaugeMetric{
		untypedMetric: untypedMetric{
			description: NewDescription(name, help, MetricTypeGauge, labels...),
		},
	}
	g.init(g, func(l ...string) Metric {
		return NewGauge(name, help, append(labels, l...)...)
	})

	return g
}

func (g *gaugeMetric) With(labels ...string) Gauge {
	return g.vector.With(labels...).(Gauge)
}
