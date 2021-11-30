package snitch

import (
	"sort"
	"sync"

	"github.com/OneOfOne/xxhash"
)

type Vector struct {
	metric   Metric
	children sync.Map
	creator  func(...string) Metric
}

func (v *Vector) SetMetric(metric Metric) *Vector {
	v.metric = metric
	return v
}

func (v *Vector) SetCreator(creator func(...string) Metric) *Vector {
	v.creator = creator
	return v
}

func (v *Vector) Describe(ch chan<- *Description) {
	ch <- v.metric.Description()
}

func (v *Vector) Collect(ch chan<- Metric) {
	var found bool

	v.children.Range(func(_, value interface{}) bool {
		found = true
		ch <- value.(Metric)

		return true
	})

	if !found {
		ch <- v.metric
	}
}

func (v *Vector) With(labels ...string) (metric Metric) {
	l := Labels{}.With(labels...)
	sort.Sort(l)

	h := xxhash.New64()
	h.WriteString(l.String())
	hash := h.Sum64()

	if val, ok := v.children.Load(hash); ok {
		return val.(Metric)
	}

	metric = v.creator(labels...)
	v.children.Store(hash, metric)

	return metric
}
