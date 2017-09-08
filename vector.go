package snitch

import (
	"io"
	"sort"
	"strings"
	"sync"

	"github.com/OneOfOne/xxhash"
)

type vector struct {
	mutex sync.RWMutex

	metric   Metric
	children map[uint64]Metric
	creator  func(...string) Metric
}

func (v *vector) init(metric Metric, creator func(...string) Metric) {
	v.metric = metric
	v.children = map[uint64]Metric{}
	v.creator = creator
}

func (v *vector) Describe(ch chan<- *Description) {
	ch <- v.metric.Description()
}

func (v *vector) Collect(ch chan<- Metric) {
	v.mutex.RLock()
	defer v.mutex.RUnlock()

	if len(v.children) == 0 {
		ch <- v.metric
	} else {
		for _, m := range v.children {
			ch <- m
		}
	}
}

func (v *vector) With(labels ...string) Metric {
	return v.loadOrStore(labels...)
}

func (v *vector) hash(labels ...string) uint64 {
	l := Labels{}.With(labels...)
	sort.Sort(l)

	h := xxhash.New64()
	r := strings.NewReader(l.String())

	io.Copy(h, r)

	return h.Sum64()
}

func (v *vector) loadOrStore(labels ...string) Metric {
	hash := v.hash(labels...)

	v.mutex.RLock()
	metric, ok := v.children[hash]
	v.mutex.RUnlock()

	if !ok {
		metric = v.creator(labels...)

		v.mutex.Lock()
		v.children[hash] = metric
		v.mutex.Unlock()
	}

	return metric
}
