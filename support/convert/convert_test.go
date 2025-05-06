package convert

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type foo struct {
	Name string
	Age  int
}

func TestTap(t *testing.T) {
	// pointer
	f := &foo{Name: "foo"}

	assert.Equal(t, "foo", f.Name)
	assert.Equal(t, 0, f.Age)

	got1 := Tap(f, func(f *foo) {
		f.Name = "bar" //nolint:goconst
		f.Age = 18
	})
	assert.Equal(t, "bar", got1.Name)
	assert.Equal(t, 18, got1.Age)

	// int
	got2 := Tap(10, func(i int) {
		assert.Equal(t, 10, i)
		i = 20
		assert.Equal(t, 20, i)
	})
	assert.Equal(t, 10, got2)

	// string
	got3 := Tap("foo", func(s string) {
		assert.Equal(t, "foo", s)
		s = "bar"
		assert.Equal(t, "bar", s)
	})
	assert.Equal(t, "foo", got3)
}

func TestWith(t *testing.T) {
	// pointer
	f := &foo{Name: "foo"}

	assert.Equal(t, "foo", f.Name)
	assert.Equal(t, 0, f.Age)

	got1 := With(f, func(f *foo) *foo {
		f.Name = "bar" //nolint:goconst
		f.Age = 18
		return f
	})
	assert.Equal(t, "bar", got1.Name)
	assert.Equal(t, 18, got1.Age)

	// int
	got2 := With(10, func(i int) int {
		return i + 10
	})
	assert.Equal(t, 20, got2)

	// string
	got3 := With("foo", func(s string) string {
		return s + "bar"
	})
	assert.Equal(t, "foobar", got3)
}

func TestTransform(t *testing.T) {
	assert.Equal(t, "1", Transform(1, strconv.Itoa))
	assert.Equal(t, &foo{Name: "foo"}, Transform("foo", func(s string) *foo {
		return &foo{Name: s}
	}))
}

func TestDefault(t *testing.T) {
	// string
	assert.Equal(t, "foo", Default("", "foo"))
	assert.Equal(t, "bar", Default("bar", "foo"))
	assert.Equal(t, "foo", Default("", "", "foo"))

	// int
	assert.Equal(t, 1, Default(0, 1))
	assert.Equal(t, 2, Default(2, 1))
	assert.Equal(t, 1, Default(0, 0, 1))

	// pointer
	assert.Equal(t, &foo{Name: "foo"}, Default(nil, &foo{Name: "foo"}))
	assert.Equal(t, &foo{Name: "bar"}, Default(&foo{Name: "bar"}, &foo{Name: "foo"}))

	// struct
	assert.Equal(t, foo{Name: "foo"}, Default(foo{}, foo{Name: "foo"}))
	assert.Equal(t, foo{Name: "bar"}, Default(foo{Name: "bar"}, foo{Name: "foo"}))

	// zero
	assert.Equal(t, 0, Default(0, 0))
}

func TestPointer(t *testing.T) {
	assert.Equal(t, "foo", *Pointer("foo"))
	assert.Equal(t, 1, *Pointer(1))
	assert.Equal(t, &foo{Name: "foo"}, *Pointer(&foo{Name: "foo"}))
	assert.Equal(t, time.Time{}, *Pointer(time.Time{}))
}

func TestUnsafeString(t *testing.T) {
	t.Parallel()
	res := UnsafeString([]byte("Hello, World!"))
	assert.Equal(t, "Hello, World!", res)
}

func TestUnsafeBytes(t *testing.T) {
	t.Parallel()
	res := UnsafeBytes("Hello, World!")
	assert.Equal(t, []byte("Hello, World!"), res)
}

func TestCopyString(t *testing.T) {
	t.Parallel()
	res := CopyString("Hello, World!")
	assert.Equal(t, "Hello, World!", res)
}

func TestCopyBytes(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		input := []byte{}
		copied := CopyBytes(input)
		assert.Equal(t, input, copied)
	})

	t.Run("single element", func(t *testing.T) {
		input := []byte{42}
		copied := CopyBytes(input)
		assert.Equal(t, input, copied)
		input[0] = 0 // Modify the input to ensure the copied slice does not change
		assert.NotEqual(t, input[0], copied[0])
	})

	t.Run("multiple elements", func(t *testing.T) {
		input := []byte{1, 2, 3, 4, 5}
		copied := CopyBytes(input)
		assert.Equal(t, input, copied)
		input[0] = 0 // Modify the input to ensure the copied slice does not change
		assert.NotEqual(t, input, copied)
	})

	t.Run("deep copy validation", func(t *testing.T) {
		input := []byte{1, 2, 3, 4, 5}
		copied := CopyBytes(input)
		input[0] = 0 // Modify the input to ensure the copied slice does not change
		assert.NotEqual(t, input[0], copied[0])
	})

	t.Run("nil slice", func(t *testing.T) {
		copied := CopyBytes(nil)
		assert.NotNil(t, copied)
		assert.Empty(t, copied)
		assert.Equal(t, 0, cap(copied))
	})
}
