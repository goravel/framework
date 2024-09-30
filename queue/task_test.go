package queue

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/support/file"
	testingfile "github.com/goravel/framework/testing/file"
)

type Test struct {
}

// Signature The name and signature of the job.
func (receiver *Test) Signature() string {
	return "test"
}

// Handle Execute the job.
func (receiver *Test) Handle(args ...any) error {
	if len(args) == 0 {
		return fmt.Errorf("no arguments provided")
	}

	arg, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("expected a string argument")
	}

	return file.Create("test.txt", arg)
}

func TestDispatchSync(t *testing.T) {
	task := &Task{
		jobs: []queue.Jobs{
			{
				Job: &Test{},
				Args: []queue.Arg{
					{Type: "uint64", Value: "test"},
				},
			},
		},
	}

	err := task.DispatchSync()
	assert.Nil(t, err)
	assert.True(t, file.Exists("test.txt"))
	assert.True(t, testingfile.GetLineNum("test.txt") == 1)
	assert.Nil(t, file.Remove("test.txt"))
}
