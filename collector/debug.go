package collector

import (
	"runtime/debug"
	"sync"
	"time"

	"github.com/kihamo/snitch"
)

const (
	MetricDebugGCLast       = "go_debug_gc_last_duration_seconds"
	MetricDebugGCNum        = "go_debug_gc"
	MetricDebugGCPause      = "go_debug_gc_pauses_duration_seconds"
	MetricDebugGCPauseTotal = "go_debug_gc_pauses_total"
	MetricDebugGCReadStats  = "go_debug_gc_read_stats_duration_seconds"
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
		CGLast:       snitch.NewGauge(MetricDebugGCLast, "Time of last collection"),
		CGNum:        snitch.NewGauge(MetricDebugGCNum, "Number of garbage collections"),
		CGPause:      snitch.NewHistogram(MetricDebugGCPause, ""),
		CGPauseTotal: snitch.NewGauge(MetricDebugGCPauseTotal, "Total pause for all collections"),
		CGReadStats:  snitch.NewTimer(MetricDebugGCReadStats, "Lead time of ReadGCStats"),
	}
}

func (c *debugCollector) Describe(ch chan<- *snitch.Description) {
	c.CGLast.Describe(ch)
	c.CGNum.Describe(ch)
	c.CGPause.Describe(ch)
	c.CGPauseTotal.Describe(ch)
	c.CGReadStats.Describe(ch)
}

func (c *debugCollector) Collect(ch chan<- snitch.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	gcLast := debugGCStats.LastGC

	t := time.Now()
	debug.ReadGCStats(&debugGCStats)
	c.CGReadStats.UpdateSince(t)

	c.CGLast.Set(float64(debugGCStats.LastGC.UTC().UnixNano()))
	c.CGNum.Set(float64(debugGCStats.NumGC))
	c.CGPauseTotal.Set(float64(debugGCStats.PauseTotal))

	if gcLast != debugGCStats.LastGC && len(debugGCStats.Pause) > 0 {
		c.CGPause.Add(float64(debugGCStats.Pause[0]))
	}

	// send metrics
	ch <- c.CGLast
	ch <- c.CGNum
	ch <- c.CGPause
	ch <- c.CGPauseTotal
	ch <- c.CGReadStats
}
