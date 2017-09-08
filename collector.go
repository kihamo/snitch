package snitch

type Collector interface {
	Describe(chan<- *Description)
	Collect(chan<- Metric)
}
