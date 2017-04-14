package collector

import (
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/kihamo/snitch"
	r "github.com/kihamo/snitch/collector/runtime"
)

const (
	MetricRuntimeReadMemStats = "runtime.read_mem_stats"

	MetricMemStatsAlloc      = "runtime.mem_stats.alloc"
	MetricMemStatsTotalAlloc = "runtime.mem_stats.total_alloc"
	MetricMemStatsSys        = "runtime.mem_stats.sys"
	MetricMemStatsLookups    = "runtime.mem_stats.lookups"
	MetricMemStatsMallocs    = "runtime.mem_stats.mallocs"
	MetricMemStatsFrees      = "runtime.mem_stats.frees"

	MetricMemStatsHeapAlloc    = "runtime.mem_stats.heap_alloc"
	MetricMemStatsHeapSys      = "runtime.mem_stats.heap_sys"
	MetricMemStatsHeapIdle     = "runtime.mem_stats.heap_idle"
	MetricMemStatsHeapInuse    = "runtime.mem_stats.heap_inuse"
	MetricMemStatsHeapReleased = "runtime.mem_stats.heap_released"
	MetricMemStatsHeapObjects  = "runtime.mem_stats.heap_objects"

	MetricMemStatsStackInuse  = "runtime.mem_stats.stack_inuse"
	MetricMemStatsStackSys    = "runtime.mem_stats.stack_sys"
	MetricMemStatsMSpanInuse  = "runtime.mem_stats.m_span_inuse"
	MetricMemStatsMSpanSys    = "runtime.mem_stats.m_span_sys"
	MetricMemStatsMCacheInuse = "runtime.mem_stats.m_cache_inuse"
	MetricMemStatsMCacheSys   = "runtime.mem_stats.m_cache_sys"
	MetricMemStatsBuckHashSys = "runtime.mem_stats.buck_hash_sys"
	MetricMemStatsGCSys       = "runtime.mem_stats.gc_sys"
	MetricMemStatsOtherSys    = "runtime.mem_stats.other_sys"

	MetricMemStatsNextGC        = "runtime.mem_stats.next_gc"
	MetricMemStatsLastGC        = "runtime.mem_stats.last_gc"
	MetricMemStatsPauseTotalNs  = "runtime.mem_stats.pause_total_ns"
	MetricMemStatsPauseNs       = "runtime.mem_stats.pause_ns"
	MetricMemStatsNumGC         = "runtime.mem_stats.num_gc"
	MetricMemStatsGCCPUFraction = "runtime.mem_stats.gc_cpu_fraction"
	MetricMemStatsEnableGC      = "runtime.mem_stats.enabled_gc"
	MetricMemStatsDebugGC       = "runtime.mem_stats.debug_gc"

	MetricRuntimeNumCgoCall   = "runtime.num_cgo_call"
	MetricRuntimeNumGoroutine = "runtime.num_goroutine"
	MetricRuntimeNumThread    = "runtime.num_thread"
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
	readMemStats snitch.Timer
	pauseNs      snitch.Histogram
	numThread    snitch.Gauge
	numCgoCall   snitch.Gauge
	numGoroutine snitch.Gauge

	memStat memStatMetrics
}

func NewRuntimeCollector() snitch.Collector {
	return &runtimeCollector{
		readMemStats: snitch.NewTimer(MetricRuntimeReadMemStats),
		pauseNs:      snitch.NewHistogram(MetricMemStatsPauseNs),
		numThread:    snitch.NewGauge(MetricRuntimeNumThread),
		numCgoCall:   snitch.NewGauge(MetricRuntimeNumCgoCall),
		numGoroutine: snitch.NewGauge(MetricRuntimeNumGoroutine),
		memStat: memStatMetrics{
			{
				metric: snitch.NewGauge(MetricMemStatsAlloc),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.Alloc))
				},
			}, {
				metric: snitch.NewCounter(MetricMemStatsTotalAlloc),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Counter).Add(float64(s.TotalAlloc))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsSys),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.Sys))
				},
			}, {
				metric: snitch.NewCounter(MetricMemStatsLookups),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Counter).Add(float64(s.Lookups))
				},
			}, {
				metric: snitch.NewCounter(MetricMemStatsMallocs),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Counter).Add(float64(s.Mallocs))
				},
			}, {
				metric: snitch.NewCounter(MetricMemStatsFrees),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Counter).Add(float64(s.Frees))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsHeapAlloc),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.HeapAlloc))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsHeapSys),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.HeapSys))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsHeapIdle),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.HeapIdle))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsHeapInuse),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.HeapInuse))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsHeapReleased),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.HeapReleased))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsHeapObjects),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.HeapObjects))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsStackInuse),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.StackInuse))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsStackSys),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.StackSys))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsMSpanInuse),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.MSpanInuse))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsMSpanSys),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.MSpanSys))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsMCacheInuse),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.MCacheInuse))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsMCacheSys),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.MCacheSys))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsBuckHashSys),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.BuckHashSys))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsGCSys),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.GCSys))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsOtherSys),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.OtherSys))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsNextGC),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.NextGC))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsLastGC),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.LastGC))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsPauseTotalNs),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.PauseTotalNs))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsNumGC),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(float64(s.NumGC))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsGCCPUFraction),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					m.(snitch.Gauge).Set(r.GCCPUFraction(s))
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsEnableGC),
				collect: func(m snitch.Metric, s *runtime.MemStats) {
					if s.EnableGC {
						m.(snitch.Gauge).Set(1)
					} else {
						m.(snitch.Gauge).Set(0)
					}
				},
			}, {
				metric: snitch.NewGauge(MetricMemStatsDebugGC),
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
