package console

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/file"
)

func TestJobMakeCommand(t *testing.T) {
	jobMakeCommand := &JobMakeCommand{}
	mockContext := &consolemocks.Context{}
	mockContext.On("Argument", 0).Return("").Once()
	mockContext.On("Ask", "Enter the job name", mock.Anything).Return("", errors.New("the job name cannot be empty")).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, jobMakeCommand.Handle(mockContext))
	}), "the job name cannot be empty")

	mockContext.On("Argument", 0).Return("GoravelJob").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Nil(t, jobMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/jobs/goravel_job.go"))

	mockContext.On("Argument", 0).Return("GoravelJob").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, jobMakeCommand.Handle(mockContext))
	}), "the job already exists. Use the --force or -f flag to overwrite")

	mockContext.On("Argument", 0).Return("Goravel/Job").Once()
	mockContext.On("OptionBool", "force").Return(false).Once()
	assert.Nil(t, jobMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/jobs/Goravel/job.go"))
	assert.True(t, file.Contain("app/jobs/Goravel/job.go", "package Goravel"))
	assert.True(t, file.Contain("app/jobs/Goravel/job.go", "type Job struct"))
	assert.Nil(t, file.Remove("app"))
}
