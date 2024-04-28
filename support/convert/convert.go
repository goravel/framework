package convert

// Tap calls the given callback with the given value then returns the value.
//
//	Tap("foo", func(s string) {
//		fmt.Println(s) // "foo" and os.Stdout will print "foo"
//	}, func(s string) {
//		// more callbacks
//	}...)
func Tap[T any](value T, callbacks ...func(T)) T {
	for _, callback := range callbacks {
		if callback != nil {
			callback(value)
		}
	}

	return value
}

// With calls the given callbacks with the given value then return the value.
//
//	With("foo", func(s string) string {
//		return s + "bar"
//	}, func(s string) string {
//		return s + "baz"
//	}) // "foobarbaz"
func With[T any](value T, callbacks ...func(T) T) T {
	for _, callback := range callbacks {
		if callback != nil {
			value = callback(value)
		}
	}

	return value
}

// Transform calls the given callback with the given value then return the result.
//
//	Transform(1, strconv.Itoa) // "1"
//	Transform("foo", func(s string) *foo {
//		return &foo{Name: s}
//	}) // &foo{Name: "foo"}
func Transform[T, R any](value T, callback func(T) R) R {
	return callback(value)
}

// Default returns the first non-zero value.
// If all values are zero, return the zero value.
//
//	Default("", "foo") // "foo"
//	Default("bar", "foo") // "bar"
//	Default("", "", "foo") // "foo"
func Default[T comparable](values ...T) T {
	var zero T
	for _, value := range values {
		if value != zero {
			return value
		}
	}
	return zero
}

// Pointer returns a pointer to the value.
//
//	Pointer("foo") // *string("foo")
//	Pointer(1) // *int(1)
func Pointer[T any](value T) *T {
	return &value
}
