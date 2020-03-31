package storage

import (
	"expvar"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/kihamo/snitch"
	"github.com/pborman/uuid"
)

type Var struct {
	expvar.Var

	lock        sync.RWMutex
	description *snitch.Description
	value       *snitch.MeasureValue
	labels      func() snitch.Labels
}

func (v *Var) String() string {
	var b strings.Builder

	v.lock.RLock()
	val := v.value
	v.lock.RUnlock()

	b.WriteString("{\"help\": \"" + v.description.Help() + "\"")

	switch v.description.Type() {
	case snitch.MetricTypeUntyped, snitch.MetricTypeCounter, snitch.MetricTypeGauge:
		b.WriteString(",\"value\": " + strconv.FormatFloat(*(val.Value), 'g', -1, 64))
		b.WriteString(",\"sample_count\": " + strconv.FormatUint(*(val.SampleCount), 10))

	case snitch.MetricTypeHistogram, snitch.MetricTypeTimer:
		b.WriteString(",\"sample_count\": " + strconv.FormatUint(*(val.SampleCount), 10))

		if !math.IsNaN(*(val.SampleSum)) {
			b.WriteString(",\"sample_sum\": " + strconv.FormatFloat(*(val.SampleSum), 'g', -1, 64))
		}

		if !math.IsNaN(*(val.SampleMin)) {
			b.WriteString(",\"sample_min\": " + strconv.FormatFloat(*(val.SampleMin), 'g', -1, 64))
		}

		if !math.IsNaN(*(val.SampleMax)) {
			b.WriteString(",\"sample_max\": " + strconv.FormatFloat(*(val.SampleMax), 'g', -1, 64))
		}

		if !math.IsNaN(*(val.SampleVariance)) {
			b.WriteString(",\"sample_variance\": " + strconv.FormatFloat(*(val.SampleVariance), 'g', -1, 64))
		}

		for q, val := range val.Quantiles {
			if !math.IsNaN(*val) {
				b.WriteString(",\"p" + strconv.FormatInt(int64(q*100), 10) + "\": " + strconv.FormatFloat(*val, 'g', -1, 64))
			}
		}
	}

	labels := v.labels().WithLabels(v.description.Labels())
	if len(labels) > 0 {
		b.WriteString(",\"labels\": {")

		for i, label := range labels {
			if i != 0 {
				b.WriteString(",")
			}

			b.WriteString("\"" + label.Key + "\": \"" + label.Value + "\"")
		}

		b.WriteString("}")
	}

	b.WriteString("}")

	return b.String()
}

func (v *Var) update(value *snitch.MeasureValue) {
	v.lock.Lock()
	v.value = value
	v.lock.Unlock()
}

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
	for _, m := range measures {
		switch m.Description.Type() {
		case snitch.MetricTypeUntyped, snitch.MetricTypeCounter, snitch.MetricTypeGauge, snitch.MetricTypeHistogram, snitch.MetricTypeTimer:
			if exists := s.expvar.Get(m.Description.Name()); exists != nil {
				exists.(*Var).update(m.Value)
			} else {
				s.expvar.Set(m.Description.Name(), &Var{
					description: m.Description,
					labels:      s.getLabels,
					value:       m.Value,
				})
			}

		default:
			continue
		}
	}

	return nil
}

func (s *Expvar) SetLabels(l snitch.Labels) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.labels = l
}

func (s *Expvar) getLabels() snitch.Labels {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.labels
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
