package snitch

import (
	"sync"

	"github.com/golang/sync/syncmap"
	"github.com/pborman/uuid"
)

const (
	sizeOfDescribeChannel = 10000
	sizeOfCollectChannel  = 10000
)

var (
	DefaultRegisterer Registerer = NewRegistry()
)

type Registerer interface {
	Register(...Collector)
	Gather() (Measures, error)
	Walk(func(*Description))
}

type Registry struct {
	collectors   *syncmap.Map
	descriptions *syncmap.Map
}

func NewRegistry() Registerer {
	return &Registry{
		collectors:   &syncmap.Map{},
		descriptions: &syncmap.Map{},
	}
}

func (r *Registry) Register(cs ...Collector) {
	for _, c := range cs {
		descriptionsChan := make(chan *Description, sizeOfDescribeChannel)

		go func() {
			c.Describe(descriptionsChan)
			close(descriptionsChan)
		}()

		for d := range descriptionsChan {
			r.descriptions.Store(d.Id(), d)
		}

		r.collectors.Store(uuid.New(), c)
	}
}

func (r *Registry) Gather() (Measures, error) {
	var wg sync.WaitGroup
	metricsChan := make(chan Metric, sizeOfCollectChannel)
	measures := Measures{}

	r.collectors.Range(func(_, _ interface{}) bool {
		wg.Add(1)
		return true
	})

	go func() {
		wg.Wait()
		close(metricsChan)
	}()

	r.collectors.Range(func(key, value interface{}) bool {
		go func(c Collector) {
			defer wg.Done()
			c.Collect(metricsChan)
		}(value.(Collector))

		return true
	})

	for metric := range metricsChan {
		d := metric.Description()
		m := &Measure{
			Name:   d.Name(),
			Type:   d.Type(),
			Labels: d.Labels(),
		}

		if err := metric.Write(m); err != nil {
			return nil, err
		}

		measures = append(measures, m)
	}

	return measures, nil
}

func (r *Registry) Walk(f func(*Description)) {
	r.descriptions.Range(func(_, value interface{}) bool {
		f(value.(*Description))
		return true
	})
}

func Register(c ...Collector) {
	DefaultRegisterer.Register(c...)
}
