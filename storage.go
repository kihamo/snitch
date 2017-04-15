package snitch

type Storage interface {
	Id() string
	Write(Measures) error
	SetLabels(Labels)
}
