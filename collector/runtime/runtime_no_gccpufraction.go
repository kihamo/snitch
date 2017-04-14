// +build !go1.5

package runtime

import (
	base "runtime"
)

func GCCPUFraction(_ *base.MemStats) float64 {
	return 0
}
