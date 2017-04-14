package main

import (
	"fmt"

	"time"

	"github.com/kihamo/snitch"
	_ "github.com/kihamo/snitch/collector"
)

func main() {
	counter := snitch.NewCounter("test-counter", "label-1", "1")
	gauge := snitch.NewGauge("test-gauge", "label-2", "2")
	histogram := snitch.NewHistogram("test-histogram", "label-3", "3")
	timer := snitch.NewTimer("test-timer", "label-4", "4", "label-5", "5")

	time.Sleep(time.Second)

	snitch.Register(counter, gauge, histogram, timer)

	snitch.DefaultRegisterer.Walk(func(d *snitch.Description) {
		fmt.Println(d.Name(), d.Type())
	})

	fmt.Println("--------")

	collect, _ := snitch.DefaultRegisterer.Gather()
	for _, measure := range collect {
		fmt.Println(measure.Name, measure.Type, measure.Counter, measure.Gauge, measure.Histogram, measure.Timer, measure.Labels)
	}

	fmt.Println("========")

	counter.Add(1234)
	gauge.Sub(39)
	histogram.Add(10)
	histogram.Add(9)
	histogram.Add(10)
	histogram.Add(9)
	timer.Time()
	timer.Time()
	timer.Time()

	time.Sleep(time.Second * 5)

	collect, _ = snitch.DefaultRegisterer.Gather()
	for _, measure := range collect {
		fmt.Println(measure.Name, measure.Type, measure.Counter, measure.Gauge, measure.Histogram, measure.Timer, measure.Labels)
	}
}
