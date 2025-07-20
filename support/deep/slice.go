package deep

func Append[T any](slice []T, items ...T) []T {
	return append([]T(nil), append(slice, items...)...)
}
