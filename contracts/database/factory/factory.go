package factory

type Factory interface {
	Definition() map[string]any
}

type Model interface {
	Factory() Factory
}
