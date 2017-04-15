package main

import (
	"log"
	"time"

	"github.com/kihamo/snitch"
	_ "github.com/kihamo/snitch/collector"
	"github.com/kihamo/snitch/storage"
)

func main() {
	counter := snitch.NewCounter("test-counter", "label-1", "1")
	gauge := snitch.NewGauge("test-gauge", "label-2", "2")
	histogram := snitch.NewHistogram("test-histogram", "label-3", "3")
	timer := snitch.NewTimer("test-timer", "label-4", "4", "label-5", "5")

	snitch.Register(counter, gauge, histogram, timer)

	s, err := storage.NewInflux("http://localhost:8086", "metrics", "metrics", "DE2RLgaPbq", "s")
	if err != nil {
		log.Panic(err.Error())
	}

	snitch.DefaultRegisterer.AddStorages(s)

	snitch.DefaultRegisterer.SendInterval(time.Second)
	time.Sleep(time.Second * 5)
	snitch.DefaultRegisterer.SendInterval(0)
	time.Sleep(time.Second * 5)
}
