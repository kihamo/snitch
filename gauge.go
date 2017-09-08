package snitch

type Gauge interface {
	Metric
	Collector

	With(...string) Gauge
	Set(float64)
	Add(float64)
	Sub(float64)
	Inc()
	Dec()
	Value() float64
}

type gaugeMetric struct {
	untypedMetric
	selfCollector
}

func NewGauge(name, help string, labels ...string) Gauge {
	g := &gaugeMetric{
		untypedMetric: untypedMetric{
			description: NewDescription(name, help, MetricTypeGauge, labels...),
		},
	}
	g.selfCollector.self = g

	return g
}

func (g *gaugeMetric) With(labels ...string) Gauge {
	return &gaugeMetric{
		untypedMetric: untypedMetric{
			bits:        g.bits,
			description: g.description,
		},
	}
}
