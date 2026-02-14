package errors

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorString_Args(t *testing.T) {
	t.Run("Args creates a new instance", func(t *testing.T) {
		original := New("test error with %s and %d")
		derived := original.Args("arg1", 123)

		// Verify they are different instances
		assert.NotSame(t, original, derived)

		// Verify original is unchanged
		assert.Equal(t, "test error with %s and %d", original.Error())

		// Verify derived has args applied
		assert.Equal(t, "test error with arg1 and 123", derived.Error())
	})

	t.Run("Multiple Args calls don't interfere", func(t *testing.T) {
		base := New("error: %s")
		err1 := base.Args("first")
		err2 := base.Args("second")

		assert.Equal(t, "error: first", err1.Error())
		assert.Equal(t, "error: second", err2.Error())
		assert.Equal(t, "error: %s", base.Error())
	})

	t.Run("Concurrent Args calls are safe", func(t *testing.T) {
		base := New("Processing jobs from [%s] connection and [%s] queue")
		var wg sync.WaitGroup
		results := make([]string, 100)

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				var conn, queue string
				if index%2 == 0 {
					conn, queue = "redis", "default"
				} else {
					conn, queue = "database", "test"
				}
				results[index] = base.Args(conn, queue).Error()
			}(i)
		}

		wg.Wait()

		// Verify all results are correct
		for i := 0; i < 100; i++ {
			if i%2 == 0 {
				assert.Equal(t, "Processing jobs from [redis] connection and [default] queue", results[i])
			} else {
				assert.Equal(t, "Processing jobs from [database] connection and [test] queue", results[i])
			}
		}
	})
}

func TestErrorString_SetModule(t *testing.T) {
	t.Run("SetModule creates a new instance", func(t *testing.T) {
		original := New("test error")
		derived := original.SetModule("TestModule")

		// Verify they are different instances
		assert.NotSame(t, original, derived)

		// Verify original is unchanged
		assert.Equal(t, "test error", original.Error())

		// Verify derived has module set
		assert.Equal(t, "[TestModule] test error", derived.Error())
	})

	t.Run("Multiple SetModule calls don't interfere", func(t *testing.T) {
		base := New("error message")
		err1 := base.SetModule("Module1")
		err2 := base.SetModule("Module2")

		assert.Equal(t, "[Module1] error message", err1.Error())
		assert.Equal(t, "[Module2] error message", err2.Error())
		assert.Equal(t, "error message", base.Error())
	})

	t.Run("Concurrent SetModule calls are safe", func(t *testing.T) {
		base := New("test error")
		var wg sync.WaitGroup
		results := make([]string, 100)

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				module := "Module1"
				if index%2 == 1 {
					module = "Module2"
				}
				results[index] = base.SetModule(module).Error()
			}(i)
		}

		wg.Wait()

		// Verify all results are correct
		for i := 0; i < 100; i++ {
			if i%2 == 0 {
				assert.Equal(t, "[Module1] test error", results[i])
			} else {
				assert.Equal(t, "[Module2] test error", results[i])
			}
		}
	})
}

func TestErrorString_ArgsAndSetModule(t *testing.T) {
	t.Run("Chaining Args and SetModule", func(t *testing.T) {
		base := New("error: %s")

		err1 := base.Args("test").SetModule("Module1")
		err2 := base.SetModule("Module2").Args("another")

		assert.Equal(t, "[Module1] error: test", err1.Error())
		assert.Equal(t, "[Module2] error: another", err2.Error())
		assert.Equal(t, "error: %s", base.Error())
	})

	t.Run("Concurrent chained calls are safe", func(t *testing.T) {
		base := New("job failed: %s")
		var wg sync.WaitGroup
		results := make([]string, 100)

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				if index%2 == 0 {
					results[index] = base.Args("timeout").SetModule("Queue").Error()
				} else {
					results[index] = base.SetModule("Worker").Args("crashed").Error()
				}
			}(i)
		}

		wg.Wait()

		// Verify all results are correct
		for i := 0; i < 100; i++ {
			if i%2 == 0 {
				assert.Equal(t, "[Queue] job failed: timeout", results[i])
			} else {
				assert.Equal(t, "[Worker] job failed: crashed", results[i])
			}
		}
	})
}

func TestErrorString_Is(t *testing.T) {
	t.Run("Is works with base error", func(t *testing.T) {
		err1 := New("test error")
		err2 := New("test error")
		err3 := New("different error")

		assert.True(t, Is(err1, err2))
		assert.False(t, Is(err1, err3))
	})

	t.Run("Is works with Args", func(t *testing.T) {
		base := New("error: %s")
		err1 := base.Args("arg1")
		err2 := base.Args("arg2")

		// Should be considered the same error (same template)
		assert.True(t, Is(err1, base))
		assert.True(t, Is(err2, base))
		assert.True(t, Is(err1, err2))
	})

	t.Run("Is works with SetModule", func(t *testing.T) {
		base := New("test error")
		err1 := base.SetModule("Module1")
		err2 := base.SetModule("Module2")

		// Should be considered the same error (same text)
		assert.True(t, Is(err1, base))
		assert.True(t, Is(err2, base))
		assert.True(t, Is(err1, err2))
	})

	t.Run("Is works with chained calls", func(t *testing.T) {
		base := New("error: %s")
		err := base.Args("test").SetModule("Module")

		assert.True(t, Is(err, base))
	})

	t.Run("Is returns false for different error types", func(t *testing.T) {
		customErr := New("test error")
		standardErr := assert.AnError

		assert.False(t, Is(customErr, standardErr))
	})
}

func TestNew(t *testing.T) {
	t.Run("New without module", func(t *testing.T) {
		err := New("error")
		assert.Equal(t, "error", err.Error())
	})

	t.Run("New with module", func(t *testing.T) {
		err := New("test error", "TestModule")
		assert.Equal(t, "[TestModule] test error", err.Error())
	})
}
