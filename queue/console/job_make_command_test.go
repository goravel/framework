package console

import (
	"testing"

	"github.com/stretchr/testify/assert"

	consolemocks "github.com/goravel/framework/contracts/console/mocks"
	"github.com/goravel/framework/support/file"
)

func TestJobMakeCommand(t *testing.T) {
	jobMakeCommand := &JobMakeCommand{}
	mockContext := &consolemocks.Context{}
	mockContext.On("Argument", 0).Return("").Once()
	err := jobMakeCommand.Handle(mockContext)
	assert.EqualError(t, err, "Not enough arguments (missing: name) ")

	mockContext.On("Argument", 0).Return("GoravelJob").Once()
	err = jobMakeCommand.Handle(mockContext)
	assert.Nil(t, err)
	assert.True(t, file.Exists("app/jobs/goravel_job.go"))

	mockContext.On("Argument", 0).Return("Goravel/Job").Once()
	err = jobMakeCommand.Handle(mockContext)
	assert.Nil(t, err)
	assert.True(t, file.Exists("app/jobs/Goravel/job.go"))
	assert.True(t, file.Contain("app/jobs/Goravel/job.go", "package Goravel"))
	assert.True(t, file.Contain("app/jobs/Goravel/job.go", "type Job struct"))
	assert.Nil(t, file.Remove("app"))
}
