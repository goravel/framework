package support

import (
	"testing"

	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/support"
	"github.com/stretchr/testify/assert"
)

type Test struct {
}

//Signature The name and signature of the job.
func (receiver *Test) Signature() string {
	return "test"
}

//Handle Execute the job.
func (receiver *Test) Handle(args ...interface{}) error {
	support.Helpers{}.CreateFile("test.txt", args[0].(string))

	return nil
}

func TestDispatchSync(t *testing.T) {
	task := &Task{
		Job: &Test{},
		Args: []queue.Arg{
			{Type: "uint64", Value: "test"},
		},
	}

	err := task.DispatchSync()
	assert.Nil(t, err)
	assert.True(t, support.Helpers{}.ExistFile("test.txt"))
	assert.True(t, support.Helpers{}.GetLineNum("test.txt") == 1)
	res := support.Helpers{}.RemoveFile("test.txt")
	assert.True(t, res)
}
