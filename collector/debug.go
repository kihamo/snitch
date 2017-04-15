package collector

import (
	"runtime/debug"
	"sync"
	"time"

	"github.com/kihamo/snitch"
)

const (
	MetricDebugGCLast       = "debug.gc.last"
	MetricDebugGCNum        = "debug.gc.num"
	MetricDebugGCPause      = "debug.gc.pause"
	MetricDebugGCPauseTotal = "debug.gc.pause_total"
	MetricDebugGCReadStats  = "debug.gc.read_stats"
)

var (
	debugGCStats debug.GCStats
)

func init() {
	debugGCStats.Pause = make([]time.Duration, 11)
}

type debugCollector struct {
	mutex sync.Mutex

	CGLast       snitch.Gauge
	CGNum        snitch.Gauge
	CGPause      snitch.Histogram
	CGPauseTotal snitch.Gauge
	CGReadStats  snitch.Timer
}

func NewDebugCollector() snitch.Collector {
	return &debugCollector{
		CGLast:       snitch.NewGauge(MetricDebugGCLast),
		CGNum:        snitch.NewGauge(MetricDebugGCNum),
		CGPause:      snitch.NewHistogram(MetricDebugGCPause),
		CGPauseTotal: snitch.NewGauge(MetricDebugGCPauseTotal),
		CGReadStats:  snitch.NewTimer(MetricDebugGCReadStats),
	}
}

func (c *debugCollector) Describe(ch chan<- *snitch.Description) {
	ch <- c.CGLast.Description()
	ch <- c.CGNum.Description()
	ch <- c.CGPause.Description()
	ch <- c.CGPauseTotal.Description()
	ch <- c.CGReadStats.Description()
}

func (c *debugCollector) Collect(ch chan<- snitch.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	gcLast := debugGCStats.LastGC

	t := time.Now()
	debug.ReadGCStats(&debugGCStats)
	c.CGReadStats.UpdateSince(t)

	c.CGLast.Set(float64(debugGCStats.LastGC.UnixNano()))
	c.CGNum.Set(float64(debugGCStats.NumGC))
	c.CGPauseTotal.Set(float64(debugGCStats.PauseTotal))

	if gcLast != debugGCStats.LastGC && len(debugGCStats.Pause) > 0 {
		c.CGPauseTotal.Add(float64(debugGCStats.Pause[0]))
	}

	ch <- c.CGReadStats
	ch <- c.CGNum
	ch <- c.CGPause
	ch <- c.CGPauseTotal
	ch <- c.CGReadStats
}
