package factory

type Factory interface {
	// Definition Defines the model's default state.
	Definition() map[string]any
}

type Model interface {
	// Factory Creates a new factory instance for the model.
	Factory() Factory
}
