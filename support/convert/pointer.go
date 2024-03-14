package convert

func Pointer[T any](t T) *T {
	return &t
}
