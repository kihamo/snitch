package snitch

type Metric interface {
	Description() *Description
	Measure() (*MeasureValue, error)
}
