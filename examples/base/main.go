package main

import (
	"expvar"
	"fmt"
	"time"

	"github.com/kihamo/snitch"
	"github.com/kihamo/snitch/collector"
	"github.com/kihamo/snitch/storage"
)

func main() {
	counter := snitch.NewCounter("test-counter", "Counter metric", "label-1", "1")
	gauge := snitch.NewGauge("test-gauge", "Gauge metric", "label-2", "2")
	histogram := snitch.NewHistogram("test-histogram", "Histogram metric", "label-3", "3")
	timer := snitch.NewTimer("test-timer", "Timer metric", "label-4", "4", "label-5", "5")

	register := snitch.NewRegistry(0)

	register.Register(
		collector.NewRuntimeCollector(),
		collector.NewDebugCollector())

	register.Register(counter, gauge, histogram, timer)

	//s, err := storage.NewInflux("http://localhost:8086", "metrics", "metrics", "DE2RLgaPbq", "s")
	// if err != nil {
	//	log.Panic(err.Error())
	//}

	s := storage.NewExpvarWithID("metrics")

	register.AddStorages(s)

	register.SendInterval(time.Second)
	expvarHandler()

	time.Sleep(time.Second * 5)
	register.SendInterval(0)
	expvarHandler()

	time.Sleep(time.Second * 5)
}

func expvarHandler() {
	fmt.Printf("{\n")
	first := true
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Printf(",\n")
		}
		first = false
		fmt.Printf("%q: %s", kv.Key, kv.Value)
	})
	fmt.Printf("\n}\n")
}
