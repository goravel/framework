package notification

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/binding"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mockslog "github.com/goravel/framework/mocks/log"
	mocksmail "github.com/goravel/framework/mocks/mail"
	mocksnotification "github.com/goravel/framework/mocks/notification"
	mocksqueue "github.com/goravel/framework/mocks/queue"
)

func TestServiceProviderRelationship(t *testing.T) {
	provider := &ServiceProvider{}

	relationship := provider.Relationship()

	assert.Equal(t, []string{binding.Notification}, relationship.Bindings)
	assert.Equal(t, binding.Bindings[binding.Notification].Dependencies, relationship.Dependencies)
	assert.Empty(t, relationship.ProvideFor)
}

func TestServiceProviderRegister(t *testing.T) {
	provider := &ServiceProvider{}

	t.Run("log facade not set", func(t *testing.T) {
		app := mocksfoundation.NewApplication(t)
		app.EXPECT().Bind(binding.Notification, mock.AnythingOfType("func(foundation.Application) (interface {}, error)")).Run(func(_ any, callback func(contractsfoundation.Application) (any, error)) {
			callbackApp := mocksfoundation.NewApplication(t)
			callbackApp.EXPECT().MakeLog().Return(nil).Once()

			instance, err := callback(callbackApp)

			assert.Nil(t, instance)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), errors.LogFacadeNotSet.Error())
		}).Once()

		provider.Register(app)
	})

	t.Run("mail facade not set", func(t *testing.T) {
		app := mocksfoundation.NewApplication(t)
		app.EXPECT().Bind(binding.Notification, mock.AnythingOfType("func(foundation.Application) (interface {}, error)")).Run(func(_ any, callback func(contractsfoundation.Application) (any, error)) {
			callbackApp := mocksfoundation.NewApplication(t)
			logger := mockslog.NewLog(t)
			callbackApp.EXPECT().MakeLog().Return(logger).Once()
			callbackApp.EXPECT().MakeMail().Return(nil).Once()

			instance, err := callback(callbackApp)

			assert.Nil(t, instance)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), errors.MailFacadeNotSet.Error())
		}).Once()

		provider.Register(app)
	})

	t.Run("orm facade not set", func(t *testing.T) {
		app := mocksfoundation.NewApplication(t)
		app.EXPECT().Bind(binding.Notification, mock.AnythingOfType("func(foundation.Application) (interface {}, error)")).Run(func(_ any, callback func(contractsfoundation.Application) (any, error)) {
			callbackApp := mocksfoundation.NewApplication(t)
			logger := mockslog.NewLog(t)
			mailer := mocksmail.NewMail(t)
			callbackApp.EXPECT().MakeLog().Return(logger).Once()
			callbackApp.EXPECT().MakeMail().Return(mailer).Once()
			callbackApp.EXPECT().MakeOrm().Return(nil).Once()

			instance, err := callback(callbackApp)

			assert.Nil(t, instance)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), errors.OrmFacadeNotSet.Error())
		}).Once()

		provider.Register(app)
	})

	t.Run("queue facade not set", func(t *testing.T) {
		app := mocksfoundation.NewApplication(t)
		app.EXPECT().Bind(binding.Notification, mock.AnythingOfType("func(foundation.Application) (interface {}, error)")).Run(func(_ any, callback func(contractsfoundation.Application) (any, error)) {
			callbackApp := mocksfoundation.NewApplication(t)
			logger := mockslog.NewLog(t)
			mailer := mocksmail.NewMail(t)
			o := mocksorm.NewOrm(t)
			callbackApp.EXPECT().MakeLog().Return(logger).Once()
			callbackApp.EXPECT().MakeMail().Return(mailer).Once()
			callbackApp.EXPECT().MakeOrm().Return(o).Once()
			callbackApp.EXPECT().MakeQueue().Return(nil).Once()

			instance, err := callback(callbackApp)

			assert.Nil(t, instance)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), errors.QueueFacadeNotSet.Error())
		}).Once()

		provider.Register(app)
	})

	t.Run("config facade not set", func(t *testing.T) {
		app := mocksfoundation.NewApplication(t)
		app.EXPECT().Bind(binding.Notification, mock.AnythingOfType("func(foundation.Application) (interface {}, error)")).Run(func(_ any, callback func(contractsfoundation.Application) (any, error)) {
			callbackApp := mocksfoundation.NewApplication(t)
			logger := mockslog.NewLog(t)
			mailer := mocksmail.NewMail(t)
			o := mocksorm.NewOrm(t)
			q := mocksqueue.NewQueue(t)
			callbackApp.EXPECT().MakeLog().Return(logger).Once()
			callbackApp.EXPECT().MakeMail().Return(mailer).Once()
			callbackApp.EXPECT().MakeOrm().Return(o).Once()
			callbackApp.EXPECT().MakeQueue().Return(q).Once()
			callbackApp.EXPECT().MakeConfig().Return(nil).Once()

			instance, err := callback(callbackApp)

			assert.Nil(t, instance)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), errors.ConfigFacadeNotSet.Error())
		}).Once()

		provider.Register(app)
	})

	t.Run("register notification manager", func(t *testing.T) {
		app := mocksfoundation.NewApplication(t)
		app.EXPECT().Bind(binding.Notification, mock.AnythingOfType("func(foundation.Application) (interface {}, error)")).Run(func(_ any, callback func(contractsfoundation.Application) (any, error)) {
			callbackApp := mocksfoundation.NewApplication(t)
			logger := mockslog.NewLog(t)
			mailer := mocksmail.NewMail(t)
			o := mocksorm.NewOrm(t)
			q := mocksqueue.NewQueue(t)
			config := mocksconfig.NewConfig(t)
			callbackApp.EXPECT().MakeLog().Return(logger).Once()
			callbackApp.EXPECT().MakeMail().Return(mailer).Once()
			callbackApp.EXPECT().MakeOrm().Return(o).Once()
			callbackApp.EXPECT().MakeQueue().Return(q).Once()
			callbackApp.EXPECT().MakeConfig().Return(config).Once()

			instance, err := callback(callbackApp)

			assert.NoError(t, err)
			assert.IsType(t, &Manager{}, instance)
		}).Once()

		provider.Register(app)
	})
}

func TestServiceProviderBoot(t *testing.T) {
	provider := &ServiceProvider{}

	t.Run("queue facade not set", func(t *testing.T) {
		app := mocksfoundation.NewApplication(t)
		app.EXPECT().Commands(mock.Anything).Once()
		app.EXPECT().MakeQueue().Return(nil).Once()

		assert.NotPanics(t, func() {
			provider.Boot(app)
		})
	})

	t.Run("notification facade not set", func(t *testing.T) {
		app := mocksfoundation.NewApplication(t)
		q := mocksqueue.NewQueue(t)
		app.EXPECT().Commands(mock.Anything).Once()
		app.EXPECT().MakeQueue().Return(q).Once()
		app.EXPECT().MakeNotification().Return(nil).Once()

		assert.NotPanics(t, func() {
			provider.Boot(app)
		})
	})

	t.Run("registers the dispatch job", func(t *testing.T) {
		app := mocksfoundation.NewApplication(t)
		q := mocksqueue.NewQueue(t)
		manager := NewManager(mockslog.NewLog(t), q)
		app.EXPECT().Commands(mock.Anything).Once()
		app.EXPECT().MakeQueue().Return(q).Once()
		app.EXPECT().MakeNotification().Return(manager).Once()
		q.EXPECT().Register(mock.Anything).Once()

		assert.NotPanics(t, func() {
			provider.Boot(app)
		})
	})

	t.Run("notification facade is not a *Manager", func(t *testing.T) {
		app := mocksfoundation.NewApplication(t)
		q := mocksqueue.NewQueue(t)
		// Satisfies contractsnotification.Manager but isn't the
		// concrete *Manager type registerJobs type-asserts against —
		// exercises the "Notification Facade is not a *Manager" branch,
		// which nothing else in this file reaches.
		notAManager := mocksnotification.NewManager(t)
		app.EXPECT().Commands(mock.Anything).Once()
		app.EXPECT().MakeQueue().Return(q).Once()
		app.EXPECT().MakeNotification().Return(notAManager).Once()

		assert.NotPanics(t, func() {
			provider.Boot(app)
		})
		// q.Register should never be reached on this branch.
		q.AssertNotCalled(t, "Register", mock.Anything)
	})
}
