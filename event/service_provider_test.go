package event

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/binding"
	contractsconsole "github.com/goravel/framework/contracts/console"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	eventConsole "github.com/goravel/framework/event/console"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mocksqueue "github.com/goravel/framework/mocks/queue"
)

func TestServiceProviderRelationship(t *testing.T) {
	serviceProvider := &ServiceProvider{}

	relationship := serviceProvider.Relationship()
	assert.Equal(t, []string{binding.Event}, relationship.Bindings)
	assert.Equal(t, binding.Bindings[binding.Event].Dependencies, relationship.Dependencies)
	assert.Empty(t, relationship.ProvideFor)
}

func TestServiceProviderRegister(t *testing.T) {
	serviceProvider := &ServiceProvider{}
	app := mocksfoundation.NewApplication(t)
	var callback func(contractsfoundation.Application) (any, error)

	app.EXPECT().Singleton(binding.Event, mock.Anything).Run(func(key interface{}, cb func(contractsfoundation.Application) (any, error)) {
		callback = cb
	}).Once()

	serviceProvider.Register(app)
	assert.NotNil(t, callback)

	errorApp := mocksfoundation.NewApplication(t)
	errorApp.EXPECT().MakeQueue().Return(nil).Once()
	instance, err := callback(errorApp)
	assert.Nil(t, instance)
	assert.Equal(t, errors.QueueFacadeNotSet.SetModule(errors.ModuleEvent), err)

	queue := mocksqueue.NewQueue(t)
	successApp := mocksfoundation.NewApplication(t)
	successApp.EXPECT().MakeQueue().Return(queue).Once()
	instance, err = callback(successApp)
	assert.NoError(t, err)
	application, ok := instance.(*Application)
	assert.True(t, ok)
	assert.NotNil(t, application)
}

func TestServiceProviderBoot(t *testing.T) {
	serviceProvider := &ServiceProvider{}
	app := mocksfoundation.NewApplication(t)

	app.EXPECT().Commands(mock.MatchedBy(func(commands []contractsconsole.Command) bool {
		if len(commands) != 2 {
			return false
		}
		var (
			hasEventCmd    bool
			hasListenerCmd bool
		)
		for _, command := range commands {
			if _, ok := command.(*eventConsole.EventMakeCommand); ok {
				hasEventCmd = true
			}
			if _, ok := command.(*eventConsole.ListenerMakeCommand); ok {
				hasListenerCmd = true
			}
		}

		return hasEventCmd && hasListenerCmd
	})).Once()

	serviceProvider.Boot(app)
}
