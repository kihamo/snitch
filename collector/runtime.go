package collector

import (
	"runtime"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/kihamo/snitch"
	r "github.com/kihamo/snitch/collector/runtime"
)

const (
	MetricRuntimeReadMemStats = "go_memstats_read_duration_seconds"

	MetricMemStatsAlloc      = "go_memstats_alloc_bytes"
	MetricMemStatsTotalAlloc = "go_memstats_alloc_bytes_total"
	MetricMemStatsSys        = "go_memstats_sys_bytes"
	MetricMemStatsLookups    = "go_memstats_lookups_total"
	MetricMemStatsMallocs    = "go_memstats_mallocs_total"
	MetricMemStatsFrees      = "go_memstats_frees_total"

	MetricMemStatsHeapAlloc    = "go_memstats_heap_alloc_bytes"
	MetricMemStatsHeapSys      = "go_memstats_heap_sys_bytes"
	MetricMemStatsHeapIdle     = "go_memstats_heap_idle_bytes"
	MetricMemStatsHeapInuse    = "go_memstats_heap_inuse_bytes"
	MetricMemStatsHeapReleased = "go_memstats_heap_released_bytes"
	MetricMemStatsHeapObjects  = "go_memstats_heap_objects"

	MetricMemStatsStackInuse  = "go_memstats_stack_inuse_bytes"
	MetricMemStatsStackSys    = "go_memstats_stack_sys_bytes"
	MetricMemStatsMSpanInuse  = "go_memstats_mspan_inuse_bytes"
	MetricMemStatsMSpanSys    = "go_memstats_mspan_sys_bytes"
	MetricMemStatsMCacheInuse = "go_memstats_mcache_inuse_bytes"
	MetricMemStatsMCacheSys   = "go_memstats_mcache_sys_bytes"
	MetricMemStatsBuckHashSys = "go_memstats_buck_hash_sys_bytes"
	MetricMemStatsGCSys       = "go_memstats_gc_sys_bytes"
	MetricMemStatsOtherSys    = "go_memstats_other_sys_bytes"

	MetricMemStatsNextGC        = "go_memstats_next_gc_bytes"
	MetricMemStatsLastGC        = "go_memstats_last_gc_time_seconds"
	MetricMemStatsPauseTotalNs  = "go_memstats_last_pause_total_nanoseconds"
	MetricMemStatsPauseNs       = "go_memstats_last_pause_nanoseconds"
	MetricMemStatsNumGC         = "go_memstats_gc"
	MetricMemStatsGCCPUFraction = "go_memstats_gc_cpu_fraction"
	MetricMemStatsEnableGC      = "go_memstats_gc_enabled"
	MetricMemStatsDebugGC       = "go_memstats_gc_debug"

	MetricRuntimeNumCgoCall   = "go_cgo_calls"
	MetricRuntimeNumGoroutine = "go_goroutines"
	MetricRuntimeNumThread    = "go_threads"
)

var (
	runtimeNumGC               uint32
	runtimeThreadCreateProfile = pprof.Lookup("threadcreate")
)

type memStatMetrics []struct {
	metric  snitch.Metric
	collect func(snitch.Metric, *runtime.MemStats)
}

type runtimeCollector struct {
	mutex sync.Mutex

	readMemStats snitch.Timer
	pauseNs      snitch.Histogram
	numThread    snitch.Gauge
	numCgoCall   snitch.Gauge
	numGoroutine snitch.Gauge

	memStat memStatMetrics
}

func NewRuntimeCollector() snitch.Collector {
	return &runtimeCollector{
		readMemStats: snitch.NewTimer(MetricRuntimeReadMemStats, "Lead time of ReadMemStats"),
		pauseNs:      snitch.NewHistogram(MetricMemStatsPauseNs, ""),
		numThread:    snitch.NewGauge(MetricRuntimeNumThread, "Number of OS threads created"),
		numCgoCall:   snitch.NewGauge(MetricRuntimeNumCgoCall, "Number of CGO calls"),
		numGoroutine: snitch.NewGauge(MetricRuntimeNumGoroutine, "Number of goroutines that currently exist"),
		memStat: memStatMetrics{
			{
				metric: snitch.NewGauge(MetricMemStatsAlloc, "Number of bytes allocated and still in use"),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.Alloc))
				},
			}, {
				metric: snitch.NewCounter(MetricMemStatsTotalAlloc, "Total number of bytes allocated, even if freed"),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Counter).Add(float64(s.TotalAlloc))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsSys, "Number of bytes obtained from system"),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.Sys))
				},
			}, {
				metric: snitch.NewCounter(MetricMemStatsLookups, "Total number of pointer lookups"),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Counter).Add(float64(s.Lookups))
				},
			}, {
				metric: snitch.NewCounter(MetricMemStatsMallocs, "Total number of mallocs"),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Counter).Add(float64(s.Mallocs))
				},
			}, {
				metric: snitch.NewCounter(MetricMemStatsFrees, "Total number of frees"),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Counter).Add(float64(s.Frees))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsHeapAlloc, "Number of heap bytes allocated and still in use"),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.HeapAlloc))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsHeapSys, "Number of heap bytes obtained from system"),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.HeapSys))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsHeapIdle, "Number of heap bytes waiting to be used"),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.HeapIdle))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsHeapInuse, "Number of heap bytes that are in use"),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.HeapInuse))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsHeapReleased, "Number of heap bytes released to OS"),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.HeapReleased))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsHeapObjects, "Number of allocated objects"),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.HeapObjects))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsStackInuse, "Number of bytes in use by the stack allocator"),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.StackInuse))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsStackSys, "Number of bytes obtained from system for stack allocator"),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.StackSys))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsMSpanInuse, "Number of bytes in use by mspan structures"),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.MSpanInuse))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsMSpanSys, "Number of bytes used for mspan structures obtained from system"),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.MSpanSys))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsMCacheInuse, "Number of bytes in use by mcache structures"),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.MCacheInuse))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsMCacheSys, "Number of bytes used for mcache structures obtained from system"),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.MCacheSys))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsBuckHashSys, "Number of bytes used by the profiling bucket hash table"),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.BuckHashSys))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsGCSys, "Number of bytes used for garbage collection system metadata"),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.GCSys))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsOtherSys, "Number of bytes used for other system allocations"),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.OtherSys))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsNextGC, "Number of heap bytes when next garbage collection will take place"),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.NextGC))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsLastGC, "Number of seconds since 1970 of last garbage collection"),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.LastGC))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsPauseTotalNs, "Cumulative nanoseconds in GC stop-the-world pauses since the program started"),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.PauseTotalNs))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsNumGC, "Number of completed GC cycles"),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.NumGC))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsGCCPUFraction, "The fraction of this program's available CPU time used by the GC since the program started"),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(r.GCCPUFraction(s))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsEnableGC, "Indicates that GC is enabled"),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					if s.EnableGC {
						m.(snitch.Gauge).Set(1)
					} else {
						m.(snitch.Gauge).Set(0)
					}
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsDebugGC, "Debug GC is currently unused"),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					if s.DebugGC {
						m.(snitch.Gauge).Set(1)
					} else {
						m.(snitch.Gauge).Set(0)
					}
				},
			},
		},
	}
}

func (c *runtimeCollector) Describe(ch chan<- *snitch.Description) {
	for _, m := range c.memStat {
		ch <- m.metric.Description()
	}

	ch <- c.readMemStats.Description()
	ch <- c.pauseNs.Description()
	ch <- c.numThread.Description()
	ch <- c.numCgoCall.Description()
	ch <- c.numGoroutine.Description()
}

func (c *runtimeCollector) Collect(ch chan<- snitch.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	t := time.Now()
	ms := &runtime.MemStats{}
	runtime.ReadMemStats(ms)

	c.readMemStats.UpdateSince(t)
	ch <- c.readMemStats

	for _, m := range c.memStat {
		m.collect(m.metric, ms)
		ch <- m.metric
	}

	i := runtimeNumGC % uint32(len(ms.PauseNs))
	ii := ms.NumGC % uint32(len(ms.PauseNs))
	if ms.NumGC-runtimeNumGC >= uint32(len(ms.PauseNs)) {
		for i = 0; i < uint32(len(ms.PauseNs)); i++ {
			c.pauseNs.Add(float64(ms.PauseNs[i]))
		}
	} else {
		if i > ii {
			for ; i < uint32(len(ms.PauseNs)); i++ {
				c.pauseNs.Add(float64(ms.PauseNs[i]))
			}
			i = 0
		}
		for ; i < ii; i++ {
			c.pauseNs.Add(float64(ms.PauseNs[i]))
		}
	}
	ch <- c.pauseNs

	c.numThread.Set(float64(runtimeThreadCreateProfile.Count()))
	ch <- c.numThread

	c.numCgoCall.Set(float64(r.GetNumCgoCall()))
	ch <- c.numCgoCall

	c.numGoroutine.Set(float64(runtime.NumGoroutine()))
	ch <- c.numGoroutine

	runtimeNumGC = ms.NumGC
}
