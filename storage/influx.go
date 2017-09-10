package storage

import (
	"fmt"
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
	return NewInfluxWithId("", url, database, username, password, precision)
}

func NewInfluxWithId(id, url, database, username, password, precision string) (*Influx, error) {
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

func (s *Influx) Id() string {
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

	var fields map[string]interface{}

	for _, m := range measures {
		switch m.Description.Type() {
		case snitch.MetricTypeUntyped, snitch.MetricTypeCounter, snitch.MetricTypeGauge:
			fields = map[string]interface{}{
				"value": *(m.Value.Value),
			}

		case snitch.MetricTypeHistogram, snitch.MetricTypeTimer:
			if *(m.Value.SampleCount) == 0 {
				continue
			}

			fields = map[string]interface{}{
				"sample_count":    float64(*(m.Value.SampleCount)),
				"sample_sum":      *(m.Value.SampleSum),
				"sample_min":      *(m.Value.SampleMin),
				"sample_max":      *(m.Value.SampleMax),
				"sample_variance": *(m.Value.SampleVariance),
			}

			for q, v := range m.Value.Quantiles {
				fields[fmt.Sprintf("p%.f", q*100)] = *v
			}

		default:
			continue
		}

		localLabels := globalLabels.WithLabels(m.Description.Labels()).Map()

		p, err := influxdb.NewPoint(m.Description.Name(), localLabels, fields, m.CreatedAt)
		if err != nil {
			return err
		}

		bp.AddPoint(p)
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
