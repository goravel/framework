package console

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/file"
)

func TestJobMakeCommand(t *testing.T) {
	jobMakeCommand := &JobMakeCommand{}
	mockContext := consolemocks.NewContext(t)
	mockContext.EXPECT().Argument(0).Return("").Once()
	mockContext.EXPECT().Ask("Enter the job name", mock.Anything).Return("", errors.New("the job name cannot be empty")).Once()
	mockContext.EXPECT().Error("the job name cannot be empty").Once()
	assert.NoError(t, jobMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("GoravelJob").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Job created successfully").Once()
	assert.NoError(t, jobMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/jobs/goravel_job.go"))

	mockContext.On("Argument", 0).Return("GoravelJob").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	mockContext.EXPECT().Error("the job already exists. Use the --force or -f flag to overwrite").Once()
	assert.NoError(t, jobMakeCommand.Handle(mockContext))

	mockContext.EXPECT().Argument(0).Return("Goravel/Job").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Success("Job created successfully").Once()
	assert.NoError(t, jobMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/jobs/Goravel/job.go"))
	assert.True(t, file.Contain("app/jobs/Goravel/job.go", "package Goravel"))
	assert.True(t, file.Contain("app/jobs/Goravel/job.go", "type Job struct"))
	assert.Nil(t, file.Remove("app"))
}
