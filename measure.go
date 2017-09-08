package snitch

import (
	"strings"
	"time"
)

type Measures []*Measure

type Measure struct {
	Description *Description
	CreatedAt   time.Time
	Value       *MeasureValue
}

type MeasureValue struct {
	Value          *float64
	SampleCount    *uint64
	SampleSum      *float64
	SampleMin      *float64
	SampleMax      *float64
	SampleVariance *float64
	Quantiles      *map[float64]float64
}

func (m Measures) Len() int {
	return len(m)
}
func (m Measures) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}
func (m Measures) Less(i, j int) bool {
	cmp := strings.Compare(m[i].Description.Name(), m[j].Description.Name())
	if cmp == 0 {
		return strings.Compare(m[i].Description.Labels().String(), m[j].Description.Labels().String()) < 0
	}

	return cmp < 0
}
