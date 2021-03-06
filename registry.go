package snitch

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/pborman/uuid"
	"go.uber.org/multierr"
)

const (
	sizeOfDescribeChannel = 10000
	sizeOfCollectChannel  = 10000

	notRunningSendInterval = time.Hour
)

type Registerer interface {
	Register(...Collector)
	Walk(func(*Description))
	Gather() (Measures, error)
	GatherAndSend() error
	AddStorages(...Storage)
	GetStorage(string) (Storage, error)
	SetLabels(Labels)
	SendInterval(time.Duration)
}

type Registry struct {
	mutex        sync.RWMutex
	collectors   *sync.Map
	descriptions *sync.Map
	storages     *sync.Map
	labels       Labels

	sendTicker chan time.Duration
}

func NewRegistry(d time.Duration) Registerer {
	r := &Registry{
		collectors:   &sync.Map{},
		descriptions: &sync.Map{},
		storages:     &sync.Map{},
		sendTicker:   make(chan time.Duration),
	}

	go func() {
		r.send(d)
	}()

	return r
}

func (r *Registry) Register(cs ...Collector) {
	for _, c := range cs {
		descriptionsChan := make(chan *Description, sizeOfDescribeChannel)

		go func() {
			c.Describe(descriptionsChan)
			close(descriptionsChan)
		}()

		for d := range descriptionsChan {
			r.descriptions.Store(d.ID(), d)
		}

		r.collectors.Store(uuid.New(), c)
	}
}

func (r *Registry) Walk(f func(*Description)) {
	r.descriptions.Range(func(_, value interface{}) bool {
		f(value.(*Description))
		return true
	})
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

	r.collectors.Range(func(_, value interface{}) bool {
		go func(c Collector) {
			defer wg.Done()
			c.Collect(metricsChan)
		}(value.(Collector))

		return true
	})

	for metric := range metricsChan {
		value, err := metric.Measure()
		if err != nil {
			return nil, err
		}

		measures = append(measures, &Measure{
			Description: metric.Description(),
			CreatedAt:   time.Now(),
			Value:       value,
		})
	}

	return measures, nil
}

func (r *Registry) GatherAndSend() error {
	measures, err := r.Gather()
	if err != nil {
		return err
	}

	var sizeOfStorages int

	r.storages.Range(func(_, _ interface{}) bool {
		sizeOfStorages++
		return true
	})

	if sizeOfStorages == 0 {
		return nil
	}

	var wg sync.WaitGroup

	wg.Add(sizeOfStorages)
	errorChan := make(chan error, sizeOfStorages)

	go func() {
		wg.Wait()
		close(errorChan)
	}()

	r.mutex.RLock()
	l := r.labels
	r.mutex.RUnlock()

	r.storages.Range(func(_, value interface{}) bool {
		go func(s Storage, m Measures, l Labels) {
			defer wg.Done()

			s.SetLabels(l)
			if e := s.Write(m); e != nil {
				errorChan <- e
			}
		}(value.(Storage), measures, l)

		return true
	})

	for e := range errorChan {
		err = multierr.Append(err, e)
	}

	return err
}

func (r *Registry) AddStorages(ss ...Storage) {
	for _, s := range ss {
		if rs, ok := s.(StorageRealtime); ok {
			rs.SetCallback(r.Gather)
		}

		r.storages.Store(s.ID(), s)
	}
}

func (r *Registry) GetStorage(id string) (Storage, error) {
	if s, ok := r.storages.Load(id); ok {
		return s.(Storage), nil
	}

	return nil, fmt.Errorf("storage %s not exists", id)
}

func (r *Registry) SetLabels(l Labels) {
	r.mutex.Lock()
	r.labels = l
	r.mutex.Unlock()

	r.storages.Range(func(_, value interface{}) bool {
		value.(Storage).SetLabels(l)
		return true
	})
}

func (r *Registry) SendInterval(d time.Duration) {
	r.sendTicker <- d
}

func (r *Registry) send(d time.Duration) {
	running := true
	if d <= 0 {
		running = false
		d = notRunningSendInterval
	}

	ticker := time.NewTicker(d)

	for {
		select {
		case <-ticker.C:
			if running {
				err := r.GatherAndSend()

				if err != nil {
					log.Print(err.Error())
				}
			}
		case d := <-r.sendTicker:
			if d <= 0 {
				running = false
				d = notRunningSendInterval
			} else {
				running = true
			}

			ticker = time.NewTicker(d)
		}
	}
}
