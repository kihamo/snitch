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
		switch m.Type {
		case snitch.MetricTypeCounter:
			fields = map[string]interface{}{
				"value": m.Counter.Value,
			}

		case snitch.MetricTypeGauge:
			fields = map[string]interface{}{
				"value": m.Gauge.Value,
			}

		case snitch.MetricTypeHistogram:
			if m.Histogram.SampleCount == 0 {
				continue
			}

			fields = map[string]interface{}{
				"sample_count":    m.Histogram.SampleCount,
				"sample_sum":      m.Histogram.SampleSum,
				"sample_min":      m.Histogram.SampleMin,
				"sample_max":      m.Histogram.SampleMax,
				"sample_variance": m.Histogram.SampleVariance,
			}

			for q, v := range m.Histogram.Quantiles {
				fields[fmt.Sprintf("p%.f", q*100)] = v
			}

		case snitch.MetricTypeTimer:
			if m.Timer.SampleCount == 0 {
				continue
			}

			fields = map[string]interface{}{
				"sample_count":    m.Timer.SampleCount,
				"sample_sum":      m.Timer.SampleSum,
				"sample_min":      m.Timer.SampleMin,
				"sample_max":      m.Timer.SampleMax,
				"sample_variance": m.Timer.SampleVariance,
			}

			for q, v := range m.Timer.Quantiles {
				fields[fmt.Sprintf("p%.f", q*100)] = v
			}

		default:
			continue
		}

		localLabels := globalLabels.WithLabels(m.Labels).Map()

		p, err := influxdb.NewPoint(m.Name, localLabels, fields, m.CreatedAt)
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
