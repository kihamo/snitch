package storage

import (
	"expvar"
	"fmt"
	"math"
	"sync"

	"github.com/kihamo/snitch"
	"github.com/pborman/uuid"
)

type Expvar struct {
	mutex sync.RWMutex

	id       string
	callback func() (snitch.Measures, error)
	labels   snitch.Labels
	expvar   *expvar.Map
}

func NewExpvar() *Expvar {
	return NewExpvarWithID("")
}

func NewExpvarWithID(id string) *Expvar {
	if id == "" {
		id = uuid.New()
	}

	storage := &Expvar{
		id:     id,
		expvar: new(expvar.Map).Init(),
	}

	if r := expvar.Get(storage.id); r == nil {
		expvar.Publish(storage.id, storage)
	}

	return storage
}

func (s *Expvar) ID() string {
	return s.id
}

func (s *Expvar) Write(measures snitch.Measures) error {
	s.mutex.RLock()
	globalLabels := s.labels
	s.mutex.RUnlock()

	for _, m := range measures {
		exp := new(expvar.Map).Init()

		switch m.Description.Type() {
		case snitch.MetricTypeUntyped, snitch.MetricTypeCounter, snitch.MetricTypeGauge:
			value := new(expvar.Float)
			value.Set(*(m.Value.Value))
			exp.Set("value", value)

			count := new(expvar.Int)
			count.Set(int64(*(m.Value.SampleCount)))
			exp.Set("sample_count", count)

		case snitch.MetricTypeHistogram, snitch.MetricTypeTimer:
			count := new(expvar.Int)
			count.Set(int64(*(m.Value.SampleCount)))
			exp.Set("sample_count", count)

			if !math.IsNaN(*(m.Value.SampleSum)) {
				sum := new(expvar.Float)
				sum.Set(*(m.Value.SampleSum))
				exp.Set("sample_sum", sum)
			}

			if !math.IsNaN(*(m.Value.SampleMin)) {
				min := new(expvar.Float)
				min.Set(*(m.Value.SampleMin))
				exp.Set("sample_min", min)
			}

			if !math.IsNaN(*(m.Value.SampleMax)) {
				max := new(expvar.Float)
				max.Set(*(m.Value.SampleMax))
				exp.Set("sample_max", max)
			}

			if !math.IsNaN(*(m.Value.SampleVariance)) {
				variance := new(expvar.Float)
				variance.Set(*(m.Value.SampleVariance))
				exp.Set("sample_variance", variance)
			}

			for q, v := range m.Value.Quantiles {
				if !math.IsNaN(*v) {
					quantile := new(expvar.Float)
					quantile.Set(*v)
					exp.Set(fmt.Sprintf("p%.f", q*100), quantile)
				}
			}

		default:
			continue
		}

		localLabels := globalLabels.WithLabels(m.Description.Labels())
		if len(localLabels) > 0 {
			expLabels := new(expvar.Map).Init()

			for _, label := range localLabels {
				labelExp := new(expvar.String)
				labelExp.Set(label.Value)

				expLabels.Set(label.Key, labelExp)
			}

			exp.Set("labels", expLabels)
		}

		help := new(expvar.String)
		help.Set(m.Description.Help())

		exp.Set("help", help)

		s.expvar.Set(m.Description.Name(), exp)
	}

	return nil
}

func (s *Expvar) SetLabels(l snitch.Labels) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.labels = l
}

func (s *Expvar) SetCallback(f func() (snitch.Measures, error)) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.callback = f
}

func (s *Expvar) String() string {
	s.mutex.RLock()
	if s.callback != nil {
		if measures, err := s.callback(); err == nil {
			s.Write(measures)
		}
	}
	s.mutex.RUnlock()

	return s.expvar.String()
}
