package snitch

type Storage interface {
	Id() string
	Write(Measures) error
	SetLabels(Labels)
}

type StorageRealtime interface {
	SetCallback(func() (Measures, error))
}
