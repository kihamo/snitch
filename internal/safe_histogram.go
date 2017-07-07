package internal

import (
	"sync"

	"github.com/bsm/histogram"
)

type SafeHistogram struct {
	sync.RWMutex
	histogram.Histogram
}

func NewSafeHistogram() *SafeHistogram {
	return &SafeHistogram{
		Histogram: *histogram.New(50),
	}
}

func (h *SafeHistogram) Copy() *SafeHistogram {
	h.RLock()
	defer h.RUnlock()

	return &SafeHistogram{
		Histogram: *h.Histogram.Copy(nil),
	}
}

func (h *SafeHistogram) Quantiles(quantiles []float64) map[float64]float64 {
	ret := make(map[float64]float64, len(quantiles))
	for _, q := range quantiles {
		ret[q] = h.Quantile(q)
	}

	return ret
}
