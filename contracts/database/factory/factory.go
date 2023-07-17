package factory

type Factory interface {
	Definition() map[string]any
}
