// +build go1.12

package collector

import (
	"runtime/debug"

	"github.com/kihamo/snitch"
)

const (
	MetricModInfo = "go_mod_info"
)

type modCollector struct {
	info snitch.Gauge
}

func NewModCollector() snitch.Collector {
	return &modCollector{
		info: snitch.NewGauge(MetricModInfo, "Package info"),
	}
}

func (c *modCollector) Describe(ch chan<- *snitch.Description) {
	c.info.Describe(ch)
}

func (c *modCollector) Collect(ch chan<- snitch.Metric) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	c.collectDep(info.Main)

	for _, dep := range info.Deps {
		c.collectDep(*dep)
	}

	c.info.Collect(ch)
}

func (c *modCollector) collectDep(m debug.Module) {
	if m.Replace != nil {
		c.collectDep(*m.Replace)
	} else {
		c.info.With("path", m.Path, "version", m.Version).Set(1)
	}
}
