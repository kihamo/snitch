package snitch

import (
	"github.com/pborman/uuid"
)

type MetricType int

const (
	MetricTypeCounter MetricType = 1 + iota
	MetricTypeGauge
	MetricTypeHistogram
	MetricTypeTimer
)

var MetricTypeValue = [...]string{
	"counter",
	"gauge",
	"histogram",
	"timer",
}

type Description struct {
	id     string
	name   string
	typ    MetricType
	labels Labels
}

func NewDescription(name string, typ MetricType, labels ...string) *Description {
	return &Description{
		id:     uuid.New(),
		name:   name,
		typ:    typ,
		labels: Labels{}.With(labels...),
	}
}

func (d *Description) Id() string {
	return d.id
}

func (d *Description) Name() string {
	return d.name
}

func (d *Description) Type() MetricType {
	return d.typ
}

func (d *Description) Labels() Labels {
	return d.labels
}

func (t MetricType) String() string {
	return MetricTypeValue[t-1]
}
