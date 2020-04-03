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
	MetricMemStatsNumForcedGC   = "go_memstats_forced_gc"
	MetricMemStatsGCCPUFraction = "go_memstats_gc_cpu_fraction"
	MetricMemStatsEnableGC      = "go_memstats_gc_enabled"
	MetricMemStatsDebugGC       = "go_memstats_gc_debug"

	MetricRuntimeNumCPU       = "go_cpu"
	MetricRuntimeNumCgoCall   = "go_cgo_calls"
	MetricRuntimeNumGoroutine = "go_goroutines"
	MetricRuntimeNumThread    = "go_threads"
	MetricRuntimeGoMaxProc    = "go_max_proc"
)

var (
	runtimeNumGC               uint32
	runtimeThreadCreateProfile = pprof.Lookup("threadcreate")
)

type runtimeCollector struct {
	mutex sync.Mutex

	readMemStats snitch.Timer
	pauseNs      snitch.Histogram
	numThread    snitch.Gauge
	numCPU       snitch.Gauge
	numCgoCall   snitch.Gauge
	numGoroutine snitch.Gauge
	goMaxProc    snitch.Gauge

	memStatAlloc      snitch.Gauge
	memStatTotalAlloc snitch.Counter
	memStatSys        snitch.Gauge
	memStatsLookups   snitch.Counter
	memStatsMallocs   snitch.Counter
	memStatsFrees     snitch.Counter

	memStatsHeapAlloc    snitch.Gauge
	memStatsHeapSys      snitch.Gauge
	memStatsHeapIdle     snitch.Gauge
	memStatsHeapInuse    snitch.Gauge
	memStatsHeapReleased snitch.Gauge
	memStatsHeapObjects  snitch.Gauge

	memStatsStackInuse    snitch.Gauge
	memStatsStackSys      snitch.Gauge
	memStatsMSpanInuse    snitch.Gauge
	memStatsMSpanSys      snitch.Gauge
	memStatsMCacheInuse   snitch.Gauge
	memStatsMCacheSys     snitch.Gauge
	memStatsBuckHashSys   snitch.Gauge
	memStatsGCSys         snitch.Gauge
	memStatsOtherSys      snitch.Gauge
	memStatsNextGC        snitch.Gauge
	memStatsLastGC        snitch.Gauge
	memStatsPauseTotalNs  snitch.Gauge
	memStatsNumGC         snitch.Gauge
	memStatsNumForcedGC   snitch.Gauge
	memStatsGCCPUFraction snitch.Gauge
	memStatsEnableGC      snitch.Gauge
	memStatsDebugGC       snitch.Gauge
}

func NewRuntimeCollector() snitch.Collector {
	return &runtimeCollector{
		readMemStats: snitch.NewTimer(MetricRuntimeReadMemStats, "Lead time of ReadMemStats"),
		pauseNs:      snitch.NewHistogram(MetricMemStatsPauseNs, ""),
		numThread:    snitch.NewGauge(MetricRuntimeNumThread, "Number of OS threads created"),
		numCPU:       snitch.NewGauge(MetricRuntimeNumCPU, "Number of logical CPUs usable by the current process"),
		numCgoCall:   snitch.NewGauge(MetricRuntimeNumCgoCall, "Number of CGO calls"),
		numGoroutine: snitch.NewGauge(MetricRuntimeNumGoroutine, "Number of goroutines that currently exist"),
		goMaxProc:    snitch.NewGauge(MetricRuntimeGoMaxProc, "Maximum number of CPUs that can be executing simultaneously"),

		memStatAlloc:      snitch.NewGauge(MetricMemStatsAlloc, "Number of bytes allocated and still in use"),
		memStatTotalAlloc: snitch.NewCounter(MetricMemStatsTotalAlloc, "Total number of bytes allocated, even if freed"),
		memStatSys:        snitch.NewGauge(MetricMemStatsSys, "Number of bytes obtained from system"),
		memStatsLookups:   snitch.NewCounter(MetricMemStatsLookups, "Total number of pointer lookups"),
		memStatsMallocs:   snitch.NewCounter(MetricMemStatsMallocs, "Total number of mallocs"),
		memStatsFrees:     snitch.NewCounter(MetricMemStatsFrees, "Total number of frees"),

		memStatsHeapAlloc:    snitch.NewGauge(MetricMemStatsHeapAlloc, "Number of heap bytes allocated and still in use"),
		memStatsHeapSys:      snitch.NewGauge(MetricMemStatsHeapSys, "Number of heap bytes obtained from system"),
		memStatsHeapIdle:     snitch.NewGauge(MetricMemStatsHeapIdle, "Number of heap bytes waiting to be used"),
		memStatsHeapInuse:    snitch.NewGauge(MetricMemStatsHeapInuse, "Number of heap bytes that are in use"),
		memStatsHeapReleased: snitch.NewGauge(MetricMemStatsHeapReleased, "Number of heap bytes released to OS"),
		memStatsHeapObjects:  snitch.NewGauge(MetricMemStatsHeapObjects, "Number of allocated objects"),

		memStatsStackInuse:  snitch.NewGauge(MetricMemStatsStackInuse, "Number of bytes in use by the stack allocator"),
		memStatsStackSys:    snitch.NewGauge(MetricMemStatsStackSys, "Number of bytes obtained from system for stack allocator"),
		memStatsMSpanInuse:  snitch.NewGauge(MetricMemStatsMSpanInuse, "Number of bytes in use by mspan structures"),
		memStatsMSpanSys:    snitch.NewGauge(MetricMemStatsMSpanSys, "Number of bytes used for mspan structures obtained from system"),
		memStatsMCacheInuse: snitch.NewGauge(MetricMemStatsMCacheInuse, "Number of bytes in use by mcache structures"),
		memStatsMCacheSys:   snitch.NewGauge(MetricMemStatsMCacheSys, "Number of bytes used for mcache structures obtained from system"),
		memStatsBuckHashSys: snitch.NewGauge(MetricMemStatsBuckHashSys, "Number of bytes used by the profiling bucket hash table"),
		memStatsGCSys:       snitch.NewGauge(MetricMemStatsGCSys, "Number of bytes used for garbage collection system metadata"),
		memStatsOtherSys:    snitch.NewGauge(MetricMemStatsOtherSys, "Number of bytes used for other system allocations"),

		memStatsNextGC:        snitch.NewGauge(MetricMemStatsNextGC, "Number of heap bytes when next garbage collection will take place"),
		memStatsLastGC:        snitch.NewGauge(MetricMemStatsLastGC, "Number of seconds since 1970 of last garbage collection"),
		memStatsPauseTotalNs:  snitch.NewGauge(MetricMemStatsPauseTotalNs, "Cumulative nanoseconds in GC stop-the-world pauses since the program started"),
		memStatsNumGC:         snitch.NewGauge(MetricMemStatsNumGC, "Number of completed GC cycles"),
		memStatsNumForcedGC:   snitch.NewGauge(MetricMemStatsNumForcedGC, "Number of GC cycles that were forced by the application calling the GC function"),
		memStatsGCCPUFraction: snitch.NewGauge(MetricMemStatsGCCPUFraction, "The fraction of this program's available CPU time used by the GC since the program started"),
		memStatsEnableGC:      snitch.NewGauge(MetricMemStatsEnableGC, "Indicates that GC is enabled"),
		memStatsDebugGC:       snitch.NewGauge(MetricMemStatsDebugGC, "Debug GC is currently unused"),
	}
}

func (c *runtimeCollector) Describe(ch chan<- *snitch.Description) {
	c.memStatAlloc.Describe(ch)
	c.memStatTotalAlloc.Describe(ch)
	c.memStatSys.Describe(ch)
	c.memStatsLookups.Describe(ch)
	c.memStatsMallocs.Describe(ch)
	c.memStatsFrees.Describe(ch)

	c.memStatsHeapAlloc.Describe(ch)
	c.memStatsHeapSys.Describe(ch)
	c.memStatsHeapIdle.Describe(ch)
	c.memStatsHeapInuse.Describe(ch)
	c.memStatsHeapReleased.Describe(ch)
	c.memStatsHeapObjects.Describe(ch)

	c.memStatsStackInuse.Describe(ch)
	c.memStatsStackSys.Describe(ch)
	c.memStatsMSpanInuse.Describe(ch)
	c.memStatsMSpanSys.Describe(ch)
	c.memStatsMCacheInuse.Describe(ch)
	c.memStatsMCacheSys.Describe(ch)
	c.memStatsBuckHashSys.Describe(ch)
	c.memStatsGCSys.Describe(ch)
	c.memStatsOtherSys.Describe(ch)

	c.memStatsNextGC.Describe(ch)
	c.memStatsLastGC.Describe(ch)
	c.memStatsPauseTotalNs.Describe(ch)
	c.memStatsNumGC.Describe(ch)
	c.memStatsNumForcedGC.Describe(ch)
	c.memStatsGCCPUFraction.Describe(ch)
	c.memStatsEnableGC.Describe(ch)
	c.memStatsDebugGC.Describe(ch)

	c.readMemStats.Describe(ch)
	c.pauseNs.Describe(ch)
	c.numThread.Describe(ch)
	c.numCPU.Describe(ch)
	c.numCgoCall.Describe(ch)
	c.numGoroutine.Describe(ch)
	c.goMaxProc.Describe(ch)
}

func (c *runtimeCollector) Collect(ch chan<- snitch.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	t := time.Now()
	ms := &runtime.MemStats{}
	runtime.ReadMemStats(ms)

	c.readMemStats.UpdateSince(t)

	// memstat
	c.memStatAlloc.Set(float64(ms.Alloc))
	c.memStatTotalAlloc.Add(float64(ms.TotalAlloc))
	c.memStatSys.Set(float64(ms.Sys))
	c.memStatsLookups.Add(float64(ms.Lookups))
	c.memStatsMallocs.Add(float64(ms.Mallocs))
	c.memStatsFrees.Add(float64(ms.Frees))

	c.memStatsHeapAlloc.Set(float64(ms.HeapAlloc))
	c.memStatsHeapSys.Set(float64(ms.HeapSys))
	c.memStatsHeapIdle.Set(float64(ms.HeapIdle))
	c.memStatsHeapInuse.Set(float64(ms.HeapInuse))
	c.memStatsHeapReleased.Set(float64(ms.HeapReleased))
	c.memStatsHeapObjects.Set(float64(ms.HeapObjects))

	c.memStatsStackInuse.Set(float64(ms.StackInuse))
	c.memStatsStackSys.Set(float64(ms.StackSys))

	c.memStatsMSpanInuse.Set(float64(ms.MSpanInuse))
	c.memStatsMSpanSys.Set(float64(ms.MSpanSys))
	c.memStatsMCacheInuse.Set(float64(ms.MCacheInuse))
	c.memStatsMCacheSys.Set(float64(ms.MCacheSys))
	c.memStatsBuckHashSys.Set(float64(ms.BuckHashSys))
	c.memStatsGCSys.Set(float64(ms.GCSys))
	c.memStatsOtherSys.Set(float64(ms.OtherSys))

	c.memStatsNextGC.Set(float64(ms.NextGC))
	c.memStatsLastGC.Set(float64(ms.LastGC))
	c.memStatsPauseTotalNs.Set(float64(ms.PauseTotalNs))
	c.memStatsNumGC.Set(float64(ms.NumGC))
	c.memStatsNumForcedGC.Set(float64(ms.NumForcedGC))
	c.memStatsGCCPUFraction.Set(r.GCCPUFraction(ms))

	if ms.EnableGC {
		c.memStatsEnableGC.Set(1)
	} else {
		c.memStatsEnableGC.Set(0)
	}

	if ms.DebugGC {
		c.memStatsDebugGC.Set(1)
	} else {
		c.memStatsDebugGC.Set(0)
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

	c.numThread.Set(float64(runtimeThreadCreateProfile.Count()))
	c.numCPU.Set(float64(runtime.NumCPU()))
	c.numCgoCall.Set(float64(r.GetNumCgoCall()))
	c.numGoroutine.Set(float64(runtime.NumGoroutine()))
	c.goMaxProc.Set(float64(runtime.GOMAXPROCS(-1)))

	runtimeNumGC = ms.NumGC

	// send metrics
	c.readMemStats.Collect(ch)
	c.pauseNs.Collect(ch)
	c.numThread.Collect(ch)
	c.numCPU.Collect(ch)
	c.numCgoCall.Collect(ch)
	c.numGoroutine.Collect(ch)
	c.goMaxProc.Collect(ch)

	c.memStatAlloc.Collect(ch)
	c.memStatTotalAlloc.Collect(ch)
	c.memStatSys.Collect(ch)
	c.memStatsLookups.Collect(ch)
	c.memStatsMallocs.Collect(ch)
	c.memStatsFrees.Collect(ch)

	c.memStatsHeapAlloc.Collect(ch)
	c.memStatsHeapSys.Collect(ch)
	c.memStatsHeapIdle.Collect(ch)
	c.memStatsHeapInuse.Collect(ch)
	c.memStatsHeapReleased.Collect(ch)
	c.memStatsHeapObjects.Collect(ch)

	c.memStatsStackInuse.Collect(ch)
	c.memStatsStackSys.Collect(ch)
	c.memStatsMSpanInuse.Collect(ch)
	c.memStatsMSpanSys.Collect(ch)
	c.memStatsMCacheInuse.Collect(ch)
	c.memStatsMCacheSys.Collect(ch)
	c.memStatsBuckHashSys.Collect(ch)
	c.memStatsGCSys.Collect(ch)
	c.memStatsOtherSys.Collect(ch)
	c.memStatsNextGC.Collect(ch)
	c.memStatsLastGC.Collect(ch)
	c.memStatsPauseTotalNs.Collect(ch)
	c.memStatsNumGC.Collect(ch)
	c.memStatsNumForcedGC.Collect(ch)
	c.memStatsGCCPUFraction.Collect(ch)
	c.memStatsEnableGC.Collect(ch)
	c.memStatsDebugGC.Collect(ch)
}
