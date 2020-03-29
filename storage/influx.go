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

	var fields map[string]interface{}

	for _, m := range measures {
		if *(m.Value.SampleCount) == 0 {
			continue
		}

		switch m.Description.Type() {
		case snitch.MetricTypeUntyped, snitch.MetricTypeCounter, snitch.MetricTypeGauge:
			fields = map[string]interface{}{
				"value":        *(m.Value.Value),
				"sample_count": float64(*(m.Value.SampleCount)),
			}

		case snitch.MetricTypeHistogram, snitch.MetricTypeTimer:
			fields = map[string]interface{}{
				"sample_count": float64(*(m.Value.SampleCount)),
			}

			if !math.IsNaN(*(m.Value.SampleSum)) {
				fields["sample_sum"] = *(m.Value.SampleSum)
			}

			if !math.IsNaN(*(m.Value.SampleMin)) {
				fields["sample_min"] = *(m.Value.SampleMin)
			}

			if !math.IsNaN(*(m.Value.SampleMax)) {
				fields["sample_max"] = *(m.Value.SampleMax)
			}

			if !math.IsNaN(*(m.Value.SampleVariance)) {
				fields["sample_variance"] = *(m.Value.SampleVariance)
			}

			for q, v := range m.Value.Quantiles {
				if !math.IsNaN(*v) {
					fields[fmt.Sprintf("p%.f", q*100)] = *v
				}
			}

		default:
			continue
		}

		localLabels := globalLabels.WithLabels(m.Description.Labels())

		p, err := influxdb.NewPoint(m.Description.Name(), localLabels.Map(), fields, m.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed create point for %s metric with labels %s because %v",
				m.Description.Name(),
				localLabels,
				err,
			)
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
