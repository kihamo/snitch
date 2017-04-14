// +build go1.5

package metrics

import (
	base "runtime"
)

func GCCPUFraction(memStats *base.MemStats) float64 {
	return memStats.GCCPUFraction
}
