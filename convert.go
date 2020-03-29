package snitch

func Float64(v float64) *float64 {
	return &v
}

func Uint64(v uint64) *uint64 {
	return &v
}

func Float64Map(src map[float64]float64) map[float64]*float64 {
	dst := make(map[float64]*float64)

	for k, val := range src {
		v := val
		dst[k] = &v
	}

	return dst
}
