package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateTask(t *testing.T) {
	t.Parallel()

	type someStruct struct{}
	var (
		taskOfWrongType                   = new(someStruct)
		taskWithoutReturnValue            = func() {}
		taskWithoutErrorAsLastReturnValue = func() int { return 0 }
		validTask                         = func(arg string) error { return nil }
	)

	err := ValidateTask(taskOfWrongType)
	assert.Equal(t, ErrTaskMustBeFunc, err)

	err = ValidateTask(taskWithoutReturnValue)
	assert.Equal(t, ErrTaskReturnsNoValue, err)

	err = ValidateTask(taskWithoutErrorAsLastReturnValue)
	assert.Equal(t, ErrLastReturnValueMustBeError, err)

	err = ValidateTask(validTask)
	assert.NoError(t, err)
}
