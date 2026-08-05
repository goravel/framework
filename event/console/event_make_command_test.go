package console

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/file"
)

func TestEventMakeCommand(t *testing.T) {
	eventMakeCommand := &EventMakeCommand{}
	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Argument(0).Return("").Once()
	mockContext.EXPECT().Ask("Enter the event name", mock.Anything).Return("", errors.New("the event name cannot be empty")).Once()
	mockContext.EXPECT().Error("the event name cannot be empty").Once()
	assert.Nil(t, eventMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("GoravelEvent").Once()
	mockContext.EXPECT().OptionBool("broadcast").Return(false).Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Event created successfully").Once()
	assert.Nil(t, eventMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/events/goravel_event.go"))

	mockContext.EXPECT().Argument(0).Return("GoravelEvent").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Error("the event already exists. Use the --force or -f flag to overwrite").Once()
	assert.Nil(t, eventMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("Goravel/Event").Once()
	mockContext.EXPECT().OptionBool("broadcast").Return(false).Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Event created successfully").Once()
	assert.Nil(t, eventMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/events/Goravel/event.go"))
	assert.True(t, file.Contain("app/events/Goravel/event.go", "package Goravel"))
	assert.True(t, file.Contain("app/events/Goravel/event.go", "type Event struct {"))

	mockContext.EXPECT().Argument(0).Return("GoravelBroadcastEvent").Once()
	mockContext.EXPECT().OptionBool("broadcast").Return(true).Once()
	mockContext.EXPECT().OptionBool("now").Return(false).Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Event created successfully").Once()
	assert.Nil(t, eventMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/events/goravel_broadcast_event.go"))
	assert.True(t, file.Contain("app/events/goravel_broadcast_event.go", `"github.com/goravel/framework/contracts/broadcasting"`))
	assert.False(t, file.Contain("app/events/goravel_broadcast_event.go", `"github.com/goravel/framework/contracts/event"`))
	assert.True(t, file.Contain("app/events/goravel_broadcast_event.go", "func (receiver *GoravelBroadcastEvent) BroadcastOn() []string {"))
	assert.True(t, file.Contain("app/events/goravel_broadcast_event.go", "func (receiver *GoravelBroadcastEvent) BroadcastAs() string {"))
	assert.True(t, file.Contain("app/events/goravel_broadcast_event.go", "func (receiver *GoravelBroadcastEvent) BroadcastWith() map[string]any {"))
	assert.True(t, file.Contain("app/events/goravel_broadcast_event.go", "func (receiver *GoravelBroadcastEvent) BroadcastWhen() bool {"))
	assert.False(t, file.Contain("app/events/goravel_broadcast_event.go", "Handle("))

	mockContext.EXPECT().Argument(0).Return("GoravelBroadcastEventNow").Once()
	mockContext.EXPECT().OptionBool("broadcast").Return(true).Once()
	mockContext.EXPECT().OptionBool("now").Return(true).Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Event created successfully").Once()
	assert.Nil(t, eventMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/events/goravel_broadcast_event_now.go"))
	assert.True(t, file.Contain("app/events/goravel_broadcast_event_now.go", `"github.com/goravel/framework/contracts/broadcasting"`))
	assert.False(t, file.Contain("app/events/goravel_broadcast_event_now.go", `"github.com/goravel/framework/contracts/event"`))
	assert.True(t, file.Contain("app/events/goravel_broadcast_event_now.go", "func (receiver *GoravelBroadcastEventNow) BroadcastOn() []string {"))
	assert.True(t, file.Contain("app/events/goravel_broadcast_event_now.go", "func (receiver *GoravelBroadcastEventNow) BroadcastAs() string {"))
	assert.True(t, file.Contain("app/events/goravel_broadcast_event_now.go", "func (receiver *GoravelBroadcastEventNow) BroadcastWith() map[string]any {"))
	assert.True(t, file.Contain("app/events/goravel_broadcast_event_now.go", "func (receiver *GoravelBroadcastEventNow) BroadcastWhen() bool {"))
	assert.True(t, file.Contain("app/events/goravel_broadcast_event_now.go", "func (receiver *GoravelBroadcastEventNow) BroadcastNow() bool {"))
	assert.False(t, file.Contain("app/events/goravel_broadcast_event_now.go", "Handle("))

	assert.Nil(t, file.Remove("app"))
}
