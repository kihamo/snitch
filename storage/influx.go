package storage

import (
	"fmt"
	"math"
	"sync"

	influxdb "github.com/influxdata/influxdb/client/v2"
	"github.com/kihamo/snitch"
	"github.com/pborman/uuid"
)

type Influx struct {
	mutex sync.RWMutex

	id     string
	client influxdb.Client
	config influxdb.BatchPointsConfig
	labels snitch.Labels
}

func NewInflux(url, database, username, password, precision string) (*Influx, error) {
	return NewInfluxWithID("", url, database, username, password, precision)
}

func NewInfluxWithID(id, url, database, username, password, precision string) (*Influx, error) {
	if id == "" {
		id = uuid.New()
	}

	storage := &Influx{
		id: id,
	}

	err := storage.Reinitialization(url, database, username, password, precision)
	if err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *Influx) ID() string {
	return s.id
}

func (s *Influx) Write(measures snitch.Measures) error {
	s.mutex.RLock()
	bp, err := influxdb.NewBatchPoints(s.config)
	globalLabels := s.labels
	s.mutex.RUnlock()

	if err != nil {
		return err
	}

	var (
		fieldsOne, fieldsTwo map[string]interface{}
		point                *influxdb.Point
	)

	for _, m := range measures {
		if *(m.Value.SampleCount) == 0 {
			continue
		}

		switch m.Description.Type() {
		case snitch.MetricTypeUntyped, snitch.MetricTypeCounter, snitch.MetricTypeGauge:
			if fieldsOne == nil {
				fieldsOne = make(map[string]interface{}, 2)
			}

			fieldsOne["value"] = *(m.Value.Value)
			fieldsOne["sample_count"] = float64(*(m.Value.SampleCount))

			point, err = influxdb.NewPoint(
				m.Description.Name(),
				globalLabels.WithLabels(m.Description.Labels()).Map(),
				fieldsOne,
				m.CreatedAt)

		case snitch.MetricTypeHistogram, snitch.MetricTypeTimer:
			if fieldsTwo == nil {
				fieldsTwo = make(map[string]interface{}, 5+100)
			} else {
				for i := range fieldsTwo {
					delete(fieldsTwo, i)
				}
			}

			fieldsTwo["sample_count"] = float64(*(m.Value.SampleCount))

			if !math.IsNaN(*(m.Value.SampleSum)) {
				fieldsTwo["sample_sum"] = *(m.Value.SampleSum)
			}

			if !math.IsNaN(*(m.Value.SampleMin)) {
				fieldsTwo["sample_min"] = *(m.Value.SampleMin)
			}

			if !math.IsNaN(*(m.Value.SampleMax)) {
				fieldsTwo["sample_max"] = *(m.Value.SampleMax)
			}

			if !math.IsNaN(*(m.Value.SampleVariance)) {
				fieldsTwo["sample_variance"] = *(m.Value.SampleVariance)
			}

			for q, v := range m.Value.Quantiles {
				if !math.IsNaN(*v) {
					fieldsTwo[fmt.Sprintf("p%.f", q*100)] = *v
				}
			}

			point, err = influxdb.NewPoint(
				m.Description.Name(),
				globalLabels.WithLabels(m.Description.Labels()).Map(),
				fieldsTwo,
				m.CreatedAt)

		default:
			continue
		}

		if err != nil {
			return fmt.Errorf("failed create point for %s metric with labels %s because %v",
				m.Description.Name(),
				globalLabels.WithLabels(m.Description.Labels()).Map(),
				err,
			)
		}

		bp.AddPoint(point)
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.client.Write(bp)
}

func (s *Influx) SetLabels(l snitch.Labels) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.labels = l
}

func (s *Influx) Reinitialization(url, database, username, password, precision string) error {
	client, err := influxdb.NewHTTPClient(influxdb.HTTPConfig{
		Addr:     url,
		Username: username,
		Password: password,
	})

	if err != nil {
		return err
	}

	config := influxdb.BatchPointsConfig{
		Database:  database,
		Precision: precision,
	}

	s.mutex.Lock()
	s.client = client
	s.config = config
	s.mutex.Unlock()

	return nil
}
