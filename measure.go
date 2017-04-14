package snitch

type Measures []*Measure

type Measure struct {
	Name      string
	Type      MetricType
	Labels    Labels
	Counter   *CounterMeasure
	Gauge     *GaugeMeasure
	Histogram *HistogramMeasure
	Timer     *TimerMeasure
}
