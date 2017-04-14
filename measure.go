package snitch

import (
	"strings"
	"time"
)

type Measures []*Measure

type Measure struct {
	Name      string
	Type      MetricType
	Labels    Labels
	CreatedAt time.Time
	Counter   *CounterMeasure
	Gauge     *GaugeMeasure
	Histogram *HistogramMeasure
	Timer     *TimerMeasure
}

func (m Measures) Len() int {
	return len(m)
}
func (m Measures) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}
func (m Measures) Less(i, j int) bool {
	cmp := strings.Compare(m[i].Name, m[j].Name)
	if cmp == 0 {
		return strings.Compare(m[i].Labels.String(), m[j].Labels.String()) < 0
	}

	return cmp < 0
}
