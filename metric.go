package snitch

type Metric interface {
	Description() *Description
	Write(*Measure) error
}
