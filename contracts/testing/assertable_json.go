package testing

type AssertableJSON interface {
	Json() map[string]any
}
