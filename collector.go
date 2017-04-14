package snitch

type Collector interface {
	Describe(chan<- *Description)
	Collect(chan<- Metric)
}

type selfCollector struct {
	self Metric
}

func (c *selfCollector) Describe(ch chan<- *Description) {
	ch <- c.self.Description()
}

func (c *selfCollector) Collect(ch chan<- Metric) {
	ch <- c.self
}
