package console

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/file"
)

var queueServiceProvider = `package providers

import (
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/facades"
)

type QueueServiceProvider struct {
}

func (receiver *QueueServiceProvider) Register(app foundation.Application) {
	facades.Queue().Register(receiver.Jobs())
}

func (receiver *QueueServiceProvider) Boot(app foundation.Application) {

}

func (receiver *QueueServiceProvider) Jobs() []queue.Job {
	return []queue.Job{}
}
`

func TestJobMakeCommand(t *testing.T) {
	jobMakeCommand := &JobMakeCommand{}
	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Argument(0).Return("").Once()
	mockContext.EXPECT().Ask("Enter the job name", mock.Anything).Return("", errors.New("the job name cannot be empty")).Once()
	mockContext.EXPECT().Error("the job name cannot be empty").Once()
	assert.NoError(t, jobMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("GoravelJob").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Job created successfully").Once()
	mockContext.EXPECT().Warning(mock.MatchedBy(func(msg string) bool {
		return strings.HasPrefix(msg, "job register failed:")
	})).Once()
	assert.NoError(t, jobMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/jobs/goravel_job.go"))

	mockContext.On("Argument", 0).Return("GoravelJob").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	mockContext.EXPECT().Error("the job already exists. Use the --force or -f flag to overwrite").Once()
	assert.NoError(t, jobMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("Goravel/Job").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Job created successfully").Once()
	mockContext.EXPECT().Success("Job registered successfully").Once()
	assert.NoError(t, file.PutContent("app/providers/queue_service_provider.go", queueServiceProvider))
	assert.NoError(t, jobMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/jobs/Goravel/job.go"))
	assert.True(t, file.Contain("app/jobs/Goravel/job.go", "package Goravel"))
	assert.True(t, file.Contain("app/jobs/Goravel/job.go", "type Job struct"))
	assert.True(t, file.Contain("app/jobs/Goravel/job.go", "goravel_job"))
	assert.True(t, file.Contain("app/providers/queue_service_provider.go", "app/jobs/Goravel"))
	assert.True(t, file.Contain("app/providers/queue_service_provider.go", "&Goravel.Job{}"))
	assert.Nil(t, file.Remove("app"))
}
