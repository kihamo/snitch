package storage

import (
	"expvar"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/kihamo/snitch"
	"github.com/pborman/uuid"
)

type Var struct {
	expvar.Var

	description *snitch.Description
	value       *snitch.MeasureValue
	labels      func() snitch.Labels
}

func (v *Var) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "{\"help\": %q", v.description.Help())

	switch v.description.Type() {
	case snitch.MetricTypeUntyped, snitch.MetricTypeCounter, snitch.MetricTypeGauge:
		fmt.Fprint(&b, ",\"value\": "+strconv.FormatFloat(*(v.value.Value), 'g', -1, 64))
		fmt.Fprint(&b, ",\"sample_count\": "+strconv.FormatUint(*(v.value.SampleCount), 10))

	case snitch.MetricTypeHistogram, snitch.MetricTypeTimer:
		fmt.Fprint(&b, ",\"sample_count\": "+strconv.FormatUint(*(v.value.SampleCount), 10))

		if !math.IsNaN(*(v.value.SampleSum)) {
			fmt.Fprint(&b, ",\"sample_sum\": "+strconv.FormatFloat(*(v.value.SampleSum), 'g', -1, 64))
		}

		if !math.IsNaN(*(v.value.SampleMin)) {
			fmt.Fprint(&b, ",\"sample_min\": "+strconv.FormatFloat(*(v.value.SampleMin), 'g', -1, 64))
		}

		if !math.IsNaN(*(v.value.SampleMax)) {
			fmt.Fprint(&b, ",\"sample_max\": "+strconv.FormatFloat(*(v.value.SampleMax), 'g', -1, 64))
		}

		if !math.IsNaN(*(v.value.SampleVariance)) {
			fmt.Fprint(&b, ",\"sample_variance\": "+strconv.FormatFloat(*(v.value.SampleVariance), 'g', -1, 64))
		}

		for q, val := range v.value.Quantiles {
			if !math.IsNaN(*val) {
				fmt.Fprint(&b, ",\"p"+strconv.FormatInt(int64(q*100), 10)+"\": "+strconv.FormatFloat(*val, 'g', -1, 64))
			}
		}
	}

	localLabels := v.labels().WithLabels(v.description.Labels())
	if len(localLabels) > 0 {
		fmt.Fprint(&b, ",\"labels\": {")

		for i, label := range localLabels {
			if i != 0 {
				fmt.Fprint(&b, ",")
			}

			fmt.Fprint(&b, "\""+label.Key+"\": \""+label.Value+"\"")
		}

		fmt.Fprint(&b, "}")
	}

	fmt.Fprintf(&b, "}")

	return b.String()
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
	var v *Var

	for _, m := range measures {
		switch m.Description.Type() {
		case snitch.MetricTypeUntyped, snitch.MetricTypeCounter, snitch.MetricTypeGauge, snitch.MetricTypeHistogram, snitch.MetricTypeTimer:
			if exists := s.expvar.Get(m.Description.Name()); exists != nil {
				v = exists.(*Var)
			} else {
				v = &Var{
					description: m.Description,
					labels:      s.getLabels,
				}

				s.expvar.Set(m.Description.Name(), v)
			}

			v.value = m.Value

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
