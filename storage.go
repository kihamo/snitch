package snitch

type Storage interface {
	ID() string
	Write(Measures) error
	SetLabels(Labels)
}

type StorageRealtime interface {
	SetCallback(func() (Measures, error))
}
