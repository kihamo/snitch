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
	metric := &gaugeMetric{
		untypedMetric: untypedMetric{
			description: NewDescription(name, help, MetricTypeGauge, labels...),
		},
	}
	metric.SetMetric(metric).SetCreator(func(l ...string) Metric {
		return NewGauge(name, help, append(labels, l...)...)
	})

	return metric
}

func (g *gaugeMetric) With(labels ...string) Gauge {
	return g.Vector.With(labels...).(Gauge)
}
