package queue

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/event"
	contractsqueue "github.com/goravel/framework/contracts/queue"
)

type TestJob struct {
}

func (receiver *TestJob) Signature() string {
	return "TestName"
}

func (receiver *TestJob) Handle(args ...any) error {
	return nil
}

type TestJobDuplicate struct {
}

func (receiver *TestJobDuplicate) Signature() string {
	return "TestName"
}

func (receiver *TestJobDuplicate) Handle(args ...any) error {
	return nil
}

type TestJobEmpty struct {
}

func (receiver *TestJobEmpty) Signature() string {
	return ""
}

func (receiver *TestJobEmpty) Handle(args ...any) error {
	return nil
}

func TestJobs2Tasks(t *testing.T) {
	_, err := jobs2Tasks([]contractsqueue.Job{
		&TestJob{},
	})

	assert.Nil(t, err, "success")

	_, err = jobs2Tasks([]contractsqueue.Job{
		&TestJob{},
		&TestJobDuplicate{},
	})

	assert.NotNil(t, err, "Signature duplicate")

	_, err = jobs2Tasks([]contractsqueue.Job{
		&TestJobEmpty{},
	})

	assert.NotNil(t, err, "Signature empty")
}

type TestEvent struct {
}

func (receiver *TestEvent) Signature() string {
	return "TestName"
}

func (receiver *TestEvent) Handle(args []event.Arg) ([]event.Arg, error) {
	return args, nil
}

type TestListener struct {
}

func (receiver *TestListener) Signature() string {
	return "TestName"
}

func (receiver *TestListener) Queue(args ...any) event.Queue {
	return event.Queue{
		Enable:     false,
		Connection: "",
		Queue:      "",
	}
}

func (receiver *TestListener) Handle(args ...any) error {
	return nil
}

type TestListenerDuplicate struct {
}

func (receiver *TestListenerDuplicate) Signature() string {
	return "TestName"
}

func (receiver *TestListenerDuplicate) Queue(args ...any) event.Queue {
	return event.Queue{
		Enable:     false,
		Connection: "",
		Queue:      "",
	}
}

func (receiver *TestListenerDuplicate) Handle(args ...any) error {
	return nil
}

type TestListenerEmpty struct {
}

func (receiver *TestListenerEmpty) Signature() string {
	return ""
}

func (receiver *TestListenerEmpty) Queue(args ...any) event.Queue {
	return event.Queue{
		Enable:     false,
		Connection: "",
		Queue:      "",
	}
}

func (receiver *TestListenerEmpty) Handle(args ...any) error {
	return nil
}

func TestEvents2Tasks(t *testing.T) {
	_, err := eventsToTasks(map[event.Event][]event.Listener{
		&TestEvent{}: {
			&TestListener{},
		},
	})
	assert.Nil(t, err)

	_, err = eventsToTasks(map[event.Event][]event.Listener{
		&TestEvent{}: {
			&TestListener{},
			&TestListenerDuplicate{},
		},
	})
	assert.Nil(t, err)

	_, err = eventsToTasks(map[event.Event][]event.Listener{
		&TestEvent{}: {
			&TestListenerEmpty{},
		},
	})

	assert.NotNil(t, err)
}

func TestArgsToValuesWithValidArgs(t *testing.T) {
	args := []contractsqueue.Arg{
		{Type: "string", Value: "test"},
		{Type: "int", Value: json.Number("1")},
	}

	values, err := argsToValues(args)

	assert.Nil(t, err)
	assert.Equal(t, 2, len(values))
	assert.True(t, reflect.ValueOf("test").Equal(values[0]))
	assert.True(t, reflect.ValueOf(1).Equal(values[1]))
}

func TestArgsToValuesWithInvalidType(t *testing.T) {
	args := []contractsqueue.Arg{
		{Type: "invalidType", Value: "test"},
	}

	_, err := argsToValues(args)

	assert.Error(t, err)
}

func TestArgsToValuesWithEmptyArgs(t *testing.T) {
	var args []contractsqueue.Arg

	values, err := argsToValues(args)

	assert.Nil(t, err)
	assert.Equal(t, 0, len(values))
}
