package snitch

type Storage interface {
	Write(Measures)
}
