package console

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
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

var bootstrapApp = `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().WithConfig(config.Boot).Run()
}
`

var bootstrapAppWithJobs = `package bootstrap

import (
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() {
	foundation.Setup().
		WithJobs(Jobs()).WithConfig(config.Boot).Run()
}
`

var bootstrapJobs = `package bootstrap

import (
	"github.com/goravel/framework/contracts/queue"

	"goravel/app/jobs"
)

func Jobs() []queue.Job {
	return []queue.Job{
		&jobs.ExistingJob{},
	}
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
	mockContext.EXPECT().Error(mock.MatchedBy(func(msg string) bool {
		return strings.Contains(msg, "job register failed:")
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

func TestJobMakeCommand_WithBootstrapSetup(t *testing.T) {
	tests := []struct {
		name               string
		jobName            string
		bootstrapAppSetup  string
		bootstrapJobsSetup string
		expectJobsFile     bool
		expectSuccess      bool
	}{
		{
			name:              "creates job and registers in bootstrap without existing jobs.go",
			jobName:           "SendEmail",
			bootstrapAppSetup: bootstrapApp,
			expectJobsFile:    true,
			expectSuccess:     true,
		},
		{
			name:               "creates job and appends to existing jobs.go",
			jobName:            "ProcessImage",
			bootstrapAppSetup:  bootstrapAppWithJobs,
			bootstrapJobsSetup: bootstrapJobs,
			expectJobsFile:     true,
			expectSuccess:      true,
		},
		{
			name:              "creates nested job in bootstrap setup",
			jobName:           "Email/SendWelcome",
			bootstrapAppSetup: bootstrapApp,
			expectJobsFile:    true,
			expectSuccess:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup bootstrap directory
			bootstrapDir := support.Config.Paths.Bootstrap
			appFile := filepath.Join(bootstrapDir, "app.go")
			jobsFile := filepath.Join(bootstrapDir, "jobs.go")

			// Create bootstrap/app.go
			assert.NoError(t, file.PutContent(appFile, tt.bootstrapAppSetup))
			defer func() {
				assert.NoError(t, file.Remove(bootstrapDir))
				assert.NoError(t, file.Remove("app"))
			}()

			// Create bootstrap/jobs.go if provided
			if tt.bootstrapJobsSetup != "" {
				assert.NoError(t, file.PutContent(jobsFile, tt.bootstrapJobsSetup))
			}

			// Setup mock context
			jobMakeCommand := &JobMakeCommand{}
			mockContext := mocksconsole.NewContext(t)
			mockContext.EXPECT().Argument(0).Return(tt.jobName).Once()
			mockContext.EXPECT().OptionBool("force").Return(false).Once()
			mockContext.EXPECT().Success("Job created successfully").Once()

			if tt.expectSuccess {
				mockContext.EXPECT().Success("Job registered successfully").Once()
			} else {
				mockContext.EXPECT().Error(mock.MatchedBy(func(msg string) bool {
					return strings.Contains(msg, "job register failed:")
				})).Once()
			}

			// Execute command
			assert.NoError(t, jobMakeCommand.Handle(mockContext))

			// Verify job file was created using the same logic as Make.GetFilePath()
			var jobPath string
			if strings.Contains(tt.jobName, "/") {
				parts := strings.Split(tt.jobName, "/")
				folderPath := filepath.Join(parts[:len(parts)-1]...)
				structName := str.Of(parts[len(parts)-1]).Studly().String()
				fileName := str.Of(structName).Snake().String() + ".go"
				jobPath = filepath.Join("app/jobs", folderPath, fileName)
			} else {
				structName := str.Of(tt.jobName).Studly().String()
				fileName := str.Of(structName).Snake().String() + ".go"
				jobPath = filepath.Join("app/jobs", fileName)
			}
			assert.True(t, file.Exists(jobPath), "Job file should exist at %s", jobPath)

			// Verify bootstrap/jobs.go was created/updated
			if tt.expectJobsFile {
				assert.True(t, file.Exists(jobsFile), "bootstrap/jobs.go should exist")
				jobsContent, err := file.GetContent(jobsFile)
				assert.NoError(t, err)
				assert.Contains(t, jobsContent, "func Jobs() []queue.Job")
			}

			// Verify bootstrap/app.go contains WithJobs
			appContent, err := file.GetContent(appFile)
			assert.NoError(t, err)
			if tt.expectSuccess {
				assert.Contains(t, appContent, "WithJobs(Jobs())")
			}
		})
	}
}

func TestJobMakeCommand_WithNonBootstrapSetup(t *testing.T) {
	tests := []struct {
		name          string
		jobName       string
		expectSuccess bool
	}{
		{
			name:          "creates job and registers in queue service provider",
			jobName:       "SendNotification",
			expectSuccess: true,
		},
		{
			name:          "creates nested job in queue service provider",
			jobName:       "Reports/GenerateMonthly",
			expectSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup queue service provider without bootstrap setup
			providerPath := filepath.Join("app", "providers", "queue_service_provider.go")
			assert.NoError(t, file.PutContent(providerPath, queueServiceProvider))
			defer func() {
				assert.NoError(t, file.Remove("app"))
			}()

			// Setup mock context
			jobMakeCommand := &JobMakeCommand{}
			mockContext := mocksconsole.NewContext(t)
			mockContext.EXPECT().Argument(0).Return(tt.jobName).Once()
			mockContext.EXPECT().OptionBool("force").Return(false).Once()
			mockContext.EXPECT().Success("Job created successfully").Once()

			if tt.expectSuccess {
				mockContext.EXPECT().Success("Job registered successfully").Once()
			} else {
				mockContext.EXPECT().Error(mock.MatchedBy(func(msg string) bool {
					return strings.Contains(msg, "job register failed:")
				})).Once()
			}

			// Execute command
			assert.NoError(t, jobMakeCommand.Handle(mockContext))

			// Verify job file was created using the same logic as Make.GetFilePath()
			var jobPath string
			if strings.Contains(tt.jobName, "/") {
				parts := strings.Split(tt.jobName, "/")
				folderPath := filepath.Join(parts[:len(parts)-1]...)
				structName := str.Of(parts[len(parts)-1]).Studly().String()
				fileName := str.Of(structName).Snake().String() + ".go"
				jobPath = filepath.Join("app/jobs", folderPath, fileName)
			} else {
				structName := str.Of(tt.jobName).Studly().String()
				fileName := str.Of(structName).Snake().String() + ".go"
				jobPath = filepath.Join("app/jobs", fileName)
			}
			assert.True(t, file.Exists(jobPath))

			// Verify registration in queue service provider
			if tt.expectSuccess {
				providerContent, err := file.GetContent(providerPath)
				assert.NoError(t, err)
				assert.Contains(t, providerContent, "app/jobs")
			}
		})
	}
}

func TestJobMakeCommand_RegistrationError(t *testing.T) {
	t.Run("handles error when bootstrap jobs.go exists but WithJobs not registered", func(t *testing.T) {
		// Setup bootstrap directory with app.go but no WithJobs
		bootstrapDir := support.Config.Paths.Bootstrap
		appFile := filepath.Join(bootstrapDir, "app.go")
		jobsFile := filepath.Join(bootstrapDir, "jobs.go")

		// Create bootstrap/app.go without WithJobs
		assert.NoError(t, file.PutContent(appFile, bootstrapApp))
		// Create jobs.go that shouldn't exist without WithJobs
		assert.NoError(t, file.PutContent(jobsFile, bootstrapJobs))
		defer func() {
			assert.NoError(t, file.Remove(bootstrapDir))
			assert.NoError(t, file.Remove("app"))
		}()

		// Setup mock context
		jobMakeCommand := &JobMakeCommand{}
		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Argument(0).Return("FailJob").Once()
		mockContext.EXPECT().OptionBool("force").Return(false).Once()
		mockContext.EXPECT().Success("Job created successfully").Once()
		mockContext.EXPECT().Error(mock.MatchedBy(func(msg string) bool {
			return strings.Contains(msg, "job register failed:")
		})).Once()

		// Execute command
		assert.NoError(t, jobMakeCommand.Handle(mockContext))
	})

	t.Run("handles error when queue service provider is malformed", func(t *testing.T) {
		// Setup malformed queue service provider
		providerPath := filepath.Join("app", "providers", "queue_service_provider.go")
		malformedProvider := `package providers
// This is a malformed file without proper structure
`
		assert.NoError(t, file.PutContent(providerPath, malformedProvider))
		defer func() {
			assert.NoError(t, file.Remove("app"))
		}()

		// Setup mock context
		jobMakeCommand := &JobMakeCommand{}
		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Argument(0).Return("ErrorJob").Once()
		mockContext.EXPECT().OptionBool("force").Return(false).Once()
		mockContext.EXPECT().Success("Job created successfully").Once()
		mockContext.EXPECT().Error(mock.MatchedBy(func(msg string) bool {
			return strings.Contains(msg, "job register failed:")
		})).Once()

		// Execute command
		assert.NoError(t, jobMakeCommand.Handle(mockContext))
	})
}
