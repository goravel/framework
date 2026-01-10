package modify

import (
	"bytes"
	"go/token"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support"
	supportfile "github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
)

func TestAddCommand(t *testing.T) {
	tests := []struct {
		name              string
		appContent        string
		commandsContent   string // empty if file doesn't exist
		pkg               string
		command           string
		expectedApp       string
		expectedCommands  string // empty if file shouldn't be created
		wantErr           bool
		expectedErrString string
	}{
		{
			name: "add command when WithCommands doesn't exist and commands.go doesn't exist",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
			pkg:     "goravel/app/console/commands",
			command: "&commands.ExampleCommand{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithCommands(Commands).WithConfig(config.Boot).Start()
}
`,
			expectedCommands: `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"

	"goravel/app/console/commands"
)

func Commands() []console.Command {
	return []console.Command{
		&commands.ExampleCommand{},
	}
}
`,
		},
		{
			name: "add command when WithCommands exists with Commands() and commands.go exists",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithCommands(Commands).WithConfig(config.Boot).Start()
}
`,
			commandsContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"

	"goravel/app/console/commands"
)

func Commands() []console.Command {
	return []console.Command{
		&commands.ExistingCommand{},
	}
}
`,
			pkg:     "goravel/app/console/commands",
			command: "&commands.NewCommand{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithCommands(Commands).WithConfig(config.Boot).Start()
}
`,
			expectedCommands: `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"

	"goravel/app/console/commands"
)

func Commands() []console.Command {
	return []console.Command{
		&commands.ExistingCommand{},
		&commands.NewCommand{},
	}
}
`,
		},
		{
			name: "add command when WithCommands exists with inline array",
			appContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/app/console/commands"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithCommands(func() []console.Command {
			return []console.Command{
				&commands.ExistingCommand{},
			}
		}).WithConfig(config.Boot).Start()
}
`,
			pkg:     "goravel/app/console/commands",
			command: "&commands.NewCommand{}",
			expectedApp: `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/app/console/commands"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithCommands(func() []console.Command {
			return []console.Command{
				&commands.ExistingCommand{},
				&commands.NewCommand{},
			}
		}).WithConfig(config.Boot).Start()
}
`,
		},
		{
			name: "error when commands.go exists but WithCommands doesn't exist",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
			commandsContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"

	"goravel/app/console/commands"
)

func Commands() []console.Command {
	return []console.Command{
		&commands.ExistingCommand{},
	}
}
`,
			pkg:     "goravel/app/console/commands",
			command: "&commands.NewCommand{}",
			wantErr: true,
		},
		{
			name: "add command when WithCommands doesn't exist at the beginning of chain",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().Start()
}
`,
			pkg:     "goravel/app/console/commands",
			command: "&commands.FirstCommand{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithCommands(Commands).Start()
}
`,
			expectedCommands: `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"

	"goravel/app/console/commands"
)

func Commands() []console.Command {
	return []console.Command{
		&commands.FirstCommand{},
	}
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bootstrapDir := support.Config.Paths.Bootstrap
			appFile := filepath.Join(bootstrapDir, "app.go")
			commandsFile := filepath.Join(bootstrapDir, "commands.go")

			assert.NoError(t, supportfile.PutContent(appFile, tt.appContent))
			defer func() {
				assert.NoError(t, supportfile.Remove(bootstrapDir))
			}()

			if tt.commandsContent != "" {
				assert.NoError(t, supportfile.PutContent(commandsFile, tt.commandsContent))
			}

			err := AddCommand(tt.pkg, tt.command)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErrString != "" {
					assert.Contains(t, err.Error(), tt.expectedErrString)
				}
				return
			}

			assert.NoError(t, err)

			// Verify app.go content
			appContent, err := supportfile.GetContent(appFile)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedApp, appContent)

			// Verify commands.go content if expected
			if tt.expectedCommands != "" {
				commandsContent, err := supportfile.GetContent(commandsFile)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCommands, commandsContent)
			}
		})
	}
}

func TestAddFilter(t *testing.T) {
	tests := []struct {
		name              string
		appContent        string
		filtersContent    string // empty if file doesn't exist
		pkg               string
		filter            string
		expectedApp       string
		expectedFilters   string // empty if file shouldn't be created
		wantErr           bool
		expectedErrString string
	}{
		{
			name: "add filter when WithFilters doesn't exist and filters.go doesn't exist",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
			pkg:    "goravel/app/validators",
			filter: "&validators.ExampleFilter{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithFilters(Filters).WithConfig(config.Boot).Start()
}
`,
			expectedFilters: `package bootstrap

import (
	"github.com/goravel/framework/contracts/validation"

	"goravel/app/validators"
)

func Filters() []validation.Filter {
	return []validation.Filter{
		&validators.ExampleFilter{},
	}
}
`,
		},
		{
			name: "add filter when WithFilters exists with Filters() and filters.go exists",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithFilters(Filters).WithConfig(config.Boot).Start()
}
`,
			filtersContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/validation"

	"goravel/app/validators"
)

func Filters() []validation.Filter {
	return []validation.Filter{
		&validators.ExistingFilter{},
	}
}
`,
			pkg:    "goravel/app/validators",
			filter: "&validators.NewFilter{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithFilters(Filters).WithConfig(config.Boot).Start()
}
`,
			expectedFilters: `package bootstrap

import (
	"github.com/goravel/framework/contracts/validation"

	"goravel/app/validators"
)

func Filters() []validation.Filter {
	return []validation.Filter{
		&validators.ExistingFilter{},
		&validators.NewFilter{},
	}
}
`,
		},
		{
			name: "add filter when WithFilters exists with inline array",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/foundation"
	"goravel/app/validators"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithFilters(func() []validation.Filter {
			return []validation.Filter{
				&validators.ExistingFilter{},
			}
		}).WithConfig(config.Boot).Start()
}
`,
			pkg:    "goravel/app/validators",
			filter: "&validators.NewFilter{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/foundation"
	"goravel/app/validators"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithFilters(func() []validation.Filter {
			return []validation.Filter{
				&validators.ExistingFilter{},
				&validators.NewFilter{},
			}
		}).WithConfig(config.Boot).Start()
}
`,
		},
		{
			name: "error when filters.go exists but WithFilters doesn't exist",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
			filtersContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/validation"

	"goravel/app/validators"
)

func Filters() []validation.Filter {
	return []validation.Filter{
		&validators.ExistingFilter{},
	}
}
`,
			pkg:     "goravel/app/validators",
			filter:  "&validators.NewFilter{}",
			wantErr: true,
		},
		{
			name: "add filter when WithFilters doesn't exist at the beginning of chain",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().Start()
}
`,
			pkg:    "goravel/app/validators",
			filter: "&validators.FirstFilter{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithFilters(Filters).Start()
}
`,
			expectedFilters: `package bootstrap

import (
	"github.com/goravel/framework/contracts/validation"

	"goravel/app/validators"
)

func Filters() []validation.Filter {
	return []validation.Filter{
		&validators.FirstFilter{},
	}
}
`,
		},
		{
			name: "add filter from different package",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
			pkg:    "github.com/mycompany/customfilters",
			filter: "&customfilters.SpecialFilter{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithFilters(Filters).WithConfig(config.Boot).Start()
}
`,
			expectedFilters: `package bootstrap

import (
	"github.com/goravel/framework/contracts/validation"
	"github.com/mycompany/customfilters"
)

func Filters() []validation.Filter {
	return []validation.Filter{
		&customfilters.SpecialFilter{},
	}
}
`,
		},
		{
			name: "add multiple filters sequentially",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithFilters(Filters).WithConfig(config.Boot).Start()
}
`,
			filtersContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/validation"

	"goravel/app/validators"
)

func Filters() []validation.Filter {
	return []validation.Filter{
		&validators.EmailFilter{},
	}
}
`,
			pkg:    "goravel/app/validators",
			filter: "&validators.PhoneFilter{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithFilters(Filters).WithConfig(config.Boot).Start()
}
`,
			expectedFilters: `package bootstrap

import (
	"github.com/goravel/framework/contracts/validation"

	"goravel/app/validators"
)

func Filters() []validation.Filter {
	return []validation.Filter{
		&validators.EmailFilter{},
		&validators.PhoneFilter{},
	}
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bootstrapDir := support.Config.Paths.Bootstrap
			appFile := filepath.Join(bootstrapDir, "app.go")
			filtersFile := filepath.Join(bootstrapDir, "filters.go")

			assert.NoError(t, supportfile.PutContent(appFile, tt.appContent))
			defer func() {
				assert.NoError(t, supportfile.Remove(bootstrapDir))
			}()

			if tt.filtersContent != "" {
				assert.NoError(t, supportfile.PutContent(filtersFile, tt.filtersContent))
			}

			err := AddFilter(tt.pkg, tt.filter)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErrString != "" {
					assert.Contains(t, err.Error(), tt.expectedErrString)
				}
				return
			}

			assert.NoError(t, err)

			// Verify app.go content
			appContent, err := supportfile.GetContent(appFile)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedApp, appContent)

			// Verify filters.go content if expected
			if tt.expectedFilters != "" {
				filtersContent, err := supportfile.GetContent(filtersFile)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedFilters, filtersContent)
			}
		})
	}
}

func TestAddJob(t *testing.T) {
	tests := []struct {
		name              string
		appContent        string
		jobsContent       string // empty if file doesn't exist
		pkg               string
		job               string
		expectedApp       string
		expectedJobs      string // empty if file shouldn't be created
		wantErr           bool
		expectedErrString string
	}{
		{
			name: "add job when WithJobs doesn't exist and jobs.go doesn't exist",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
			pkg: "goravel/app/jobs",
			job: "&jobs.ExampleJob{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithJobs(Jobs).WithConfig(config.Boot).Start()
}
`,
			expectedJobs: `package bootstrap

import (
	"github.com/goravel/framework/contracts/queue"

	"goravel/app/jobs"
)

func Jobs() []queue.Job {
	return []queue.Job{
		&jobs.ExampleJob{},
	}
}
`,
		},
		{
			name: "add job when WithJobs exists with Jobs() and jobs.go exists",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithJobs(Jobs).WithConfig(config.Boot).Start()
}
`,
			jobsContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/queue"

	"goravel/app/jobs"
)

func Jobs() []queue.Job {
	return []queue.Job{
		&jobs.ExistingJob{},
	}
}
`,
			pkg: "goravel/app/jobs",
			job: "&jobs.NewJob{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithJobs(Jobs).WithConfig(config.Boot).Start()
}
`,
			expectedJobs: `package bootstrap

import (
	"github.com/goravel/framework/contracts/queue"

	"goravel/app/jobs"
)

func Jobs() []queue.Job {
	return []queue.Job{
		&jobs.ExistingJob{},
		&jobs.NewJob{},
	}
}
`,
		},
		{
			name: "add job when WithJobs exists with inline array",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/foundation"
	"goravel/app/jobs"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithJobs(func() []queue.Job {
			return []queue.Job{
				&jobs.ExistingJob{},
			}
		}).WithConfig(config.Boot).Start()
}
`,
			pkg: "goravel/app/jobs",
			job: "&jobs.NewJob{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/foundation"
	"goravel/app/jobs"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithJobs(func() []queue.Job {
			return []queue.Job{
				&jobs.ExistingJob{},
				&jobs.NewJob{},
			}
		}).WithConfig(config.Boot).Start()
}
`,
		},
		{
			name: "error when jobs.go exists but WithJobs doesn't exist",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
			jobsContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/queue"

	"goravel/app/jobs"
)

func Jobs() []queue.Job {
	return []queue.Job{
		&jobs.ExistingJob{},
	}
}
`,
			pkg:     "goravel/app/jobs",
			job:     "&jobs.NewJob{}",
			wantErr: true,
		},
		{
			name: "add job when WithJobs doesn't exist at the beginning of chain",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().Start()
}
`,
			pkg: "goravel/app/jobs",
			job: "&jobs.FirstJob{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithJobs(Jobs).Start()
}
`,
			expectedJobs: `package bootstrap

import (
	"github.com/goravel/framework/contracts/queue"

	"goravel/app/jobs"
)

func Jobs() []queue.Job {
	return []queue.Job{
		&jobs.FirstJob{},
	}
}
`,
		},
		{
			name: "add job from different package",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
			pkg: "github.com/mycompany/customjobs",
			job: "&customjobs.SpecialJob{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithJobs(Jobs).WithConfig(config.Boot).Start()
}
`,
			expectedJobs: `package bootstrap

import (
	"github.com/goravel/framework/contracts/queue"
	"github.com/mycompany/customjobs"
)

func Jobs() []queue.Job {
	return []queue.Job{
		&customjobs.SpecialJob{},
	}
}
`,
		},
		{
			name: "add multiple jobs sequentially",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithJobs(Jobs).WithConfig(config.Boot).Start()
}
`,
			jobsContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/queue"

	"goravel/app/jobs"
)

func Jobs() []queue.Job {
	return []queue.Job{
		&jobs.SendEmailJob{},
	}
}
`,
			pkg: "goravel/app/jobs",
			job: "&jobs.ProcessImageJob{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithJobs(Jobs).WithConfig(config.Boot).Start()
}
`,
			expectedJobs: `package bootstrap

import (
	"github.com/goravel/framework/contracts/queue"

	"goravel/app/jobs"
)

func Jobs() []queue.Job {
	return []queue.Job{
		&jobs.SendEmailJob{},
		&jobs.ProcessImageJob{},
	}
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bootstrapDir := support.Config.Paths.Bootstrap
			appFile := filepath.Join(bootstrapDir, "app.go")
			jobsFile := filepath.Join(bootstrapDir, "jobs.go")

			assert.NoError(t, supportfile.PutContent(appFile, tt.appContent))
			defer func() {
				assert.NoError(t, supportfile.Remove(bootstrapDir))
			}()

			if tt.jobsContent != "" {
				assert.NoError(t, supportfile.PutContent(jobsFile, tt.jobsContent))
			}

			err := AddJob(tt.pkg, tt.job)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErrString != "" {
					assert.Contains(t, err.Error(), tt.expectedErrString)
				}
				return
			}

			assert.NoError(t, err)

			// Verify app.go content
			appContent, err := supportfile.GetContent(appFile)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedApp, appContent)

			// Verify jobs.go content if expected
			if tt.expectedJobs != "" {
				jobsContent, err := supportfile.GetContent(jobsFile)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedJobs, jobsContent)
			}
		})
	}
}

func TestAddMiddleware(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		pkg      string
		mw       string
		expected string
		wantErr  bool
	}{
		{
			name: "add middleware when WithMiddleware doesn't exist",
			content: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
			pkg: "github.com/goravel/framework/http/middleware",
			mw:  "&middleware.Auth{}",
			expected: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
			handler.Append(
				&middleware.Auth{},
			)
		}).WithConfig(config.Boot).Start()
}
`,
		},
		{
			name: "add middleware when WithMiddleware already exists",
			content: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
			handler.Append(&middleware.Existing{})
		}).WithConfig(config.Boot).Start()
}
`,
			pkg: "github.com/goravel/framework/http/middleware",
			mw:  "&middleware.Auth{}",
			expected: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
			handler.Append(&middleware.Existing{},
				&middleware.Auth{},
			)
		}).WithConfig(config.Boot).Start()
}
`,
		},
		{
			name: "add middleware with complex chain",
			content: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).WithRoute(route.Boot).Start()
}
`,
			pkg: "github.com/goravel/framework/http/middleware",
			mw:  "&middleware.Cors{}",
			expected: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
			handler.Append(
				&middleware.Cors{},
			)
		}).WithConfig(config.Boot).WithRoute(route.Boot).Start()
}
`,
		},
		{
			name: "add middleware when WithMiddleware exists but no Append call",
			content: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
		}).WithConfig(config.Boot).Start()
}
`,
			pkg: "github.com/goravel/framework/http/middleware",
			mw:  "&middleware.Auth{}",
			expected: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
			handler.Append(
				&middleware.Auth{},
			)
		}).WithConfig(config.Boot).Start()
}
`,
		},
		{
			name: "add middleware to Boot function with multiple statements",
			content: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	_ = foundation.NewApplication()
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
			pkg: "github.com/goravel/framework/http/middleware",
			mw:  "&middleware.Throttle{}",
			expected: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/http/middleware"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	_ = foundation.NewApplication()
	return foundation.Setup().
		WithMiddleware(func(handler configuration.Middleware) {
			handler.Append(
				&middleware.Throttle{},
			)
		}).WithConfig(config.Boot).Start()
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bootstrapDir := support.Config.Paths.Bootstrap
			sourceFile := filepath.Join(bootstrapDir, "app.go")

			assert.NoError(t, supportfile.PutContent(sourceFile, tt.content))
			defer func() {
				assert.NoError(t, supportfile.Remove(bootstrapDir))
			}()

			err := AddMiddleware(tt.pkg, tt.mw)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			content, err := supportfile.GetContent(sourceFile)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, content)
		})
	}
}

func TestAddMigration(t *testing.T) {
	tests := []struct {
		name               string
		appContent         string
		migrationsContent  string // empty if file doesn't exist
		pkg                string
		migration          string
		expectedApp        string
		expectedMigrations string // empty if file shouldn't be created
		wantErr            bool
		expectedErrString  string
	}{
		{
			name: "add migration when WithMigrations doesn't exist and migrations.go doesn't exist",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
			pkg:       "goravel/database/migrations",
			migration: "&migrations.CreateUsersTable{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMigrations(Migrations).WithConfig(config.Boot).Start()
}
`,
			expectedMigrations: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.CreateUsersTable{},
	}
}
`,
		},
		{
			name: "add migration when WithMigrations exists with Migrations() and migrations.go exists",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMigrations(Migrations).WithConfig(config.Boot).Start()
}
`,
			migrationsContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.ExistingMigration{},
	}
}
`,
			pkg:       "goravel/database/migrations",
			migration: "&migrations.CreateUsersTable{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMigrations(Migrations).WithConfig(config.Boot).Start()
}
`,
			expectedMigrations: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.ExistingMigration{},
		&migrations.CreateUsersTable{},
	}
}
`,
		},
		{
			name: "add migration when WithMigrations exists with inline array",
			appContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
	"goravel/database/migrations"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMigrations(func() []schema.Migration {
			return []schema.Migration{
				&migrations.ExistingMigration{},
			}
		}).WithConfig(config.Boot).Start()
}
`,
			pkg:       "goravel/database/migrations",
			migration: "&migrations.CreateUsersTable{}",
			expectedApp: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
	"goravel/database/migrations"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMigrations(func() []schema.Migration {
			return []schema.Migration{
				&migrations.ExistingMigration{},
				&migrations.CreateUsersTable{},
			}
		}).WithConfig(config.Boot).Start()
}
`,
		},
		{
			name: "error when migrations.go exists but WithMigrations doesn't exist",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
			migrationsContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.ExistingMigration{},
	}
}
`,
			pkg:       "goravel/database/migrations",
			migration: "&migrations.CreateUsersTable{}",
			wantErr:   true,
		},
		{
			name: "add migration when WithMigrations doesn't exist at the beginning of chain",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().Start()
}
`,
			pkg:       "goravel/database/migrations",
			migration: "&migrations.CreatePostsTable{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMigrations(Migrations).Start()
}
`,
			expectedMigrations: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.CreatePostsTable{},
	}
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bootstrapDir := support.Config.Paths.Bootstrap
			appFile := filepath.Join(bootstrapDir, "app.go")
			migrationsFile := filepath.Join(bootstrapDir, "migrations.go")

			assert.NoError(t, supportfile.PutContent(appFile, tt.appContent))
			defer func() {
				assert.NoError(t, supportfile.Remove(bootstrapDir))
			}()

			if tt.migrationsContent != "" {
				assert.NoError(t, supportfile.PutContent(migrationsFile, tt.migrationsContent))
			}

			err := AddMigration(tt.pkg, tt.migration)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErrString != "" {
					assert.Contains(t, err.Error(), tt.expectedErrString)
				}
				return
			}

			assert.NoError(t, err)

			// Verify app.go content
			appContent, err := supportfile.GetContent(appFile)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedApp, appContent)

			// Verify migrations.go content if expected
			if tt.expectedMigrations != "" {
				migrationsContent, err := supportfile.GetContent(migrationsFile)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedMigrations, migrationsContent)
			}
		})
	}
}

func TestAddProvider(t *testing.T) {
	tests := []struct {
		name              string
		appContent        string
		providersContent  string // empty if file doesn't exist
		pkg               string
		provider          string
		expectedApp       string
		expectedProviders string // empty if file shouldn't be created
		wantErr           bool
		expectedErrString string
	}{
		{
			name: "add provider when WithProviders doesn't exist and providers.go doesn't exist",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
			pkg:      "goravel/app/providers",
			provider: "&providers.AppServiceProvider{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithProviders(Providers).WithConfig(config.Boot).Start()
}
`,
			expectedProviders: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation"

	"goravel/app/providers"
)

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&providers.AppServiceProvider{},
	}
}
`,
		},
		{
			name: "add provider when WithProviders exists with Providers() and providers.go exists",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithProviders(Providers).WithConfig(config.Boot).Start()
}
`,
			providersContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation"

	"goravel/app/providers"
)

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&providers.ExistingProvider{},
	}
}
`,
			pkg:      "goravel/app/providers",
			provider: "&providers.AppServiceProvider{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithProviders(Providers).WithConfig(config.Boot).Start()
}
`,
			expectedProviders: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation"

	"goravel/app/providers"
)

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&providers.ExistingProvider{},
		&providers.AppServiceProvider{},
	}
}
`,
		},
		{
			name: "add provider when WithProviders exists with inline array",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/app/providers"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithProviders(func() []foundation.ServiceProvider {
			return []foundation.ServiceProvider{
				&providers.ExistingProvider{},
			}
		}).WithConfig(config.Boot).Start()
}
`,
			pkg:      "goravel/app/providers",
			provider: "&providers.AppServiceProvider{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/app/providers"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithProviders(func() []foundation.ServiceProvider {
			return []foundation.ServiceProvider{
				&providers.ExistingProvider{},
				&providers.AppServiceProvider{},
			}
		}).WithConfig(config.Boot).Start()
}
`,
		},
		{
			name: "error when providers.go exists but WithProviders doesn't exist",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
			providersContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation"

	"goravel/app/providers"
)

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&providers.ExistingProvider{},
	}
}
`,
			pkg:               "goravel/app/providers",
			provider:          "&providers.AppServiceProvider{}",
			wantErr:           true,
			expectedErrString: "providers.go already exists but WithProviders is not registered in foundation.Setup()",
		},
		{
			name: "add provider when WithProviders doesn't exist at the beginning of chain",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().Start()
}
`,
			pkg:      "goravel/app/providers",
			provider: "&providers.RouteServiceProvider{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithProviders(Providers).Start()
}
`,
			expectedProviders: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation"

	"goravel/app/providers"
)

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&providers.RouteServiceProvider{},
	}
}
`,
		},
		{
			name: "add provider from different package",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
			pkg:      "github.com/goravel/redis",
			provider: "&redis.ServiceProvider{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithProviders(Providers).WithConfig(config.Boot).Start()
}
`,
			expectedProviders: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/redis"
)

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&redis.ServiceProvider{},
	}
}
`,
		},
		{
			name: "add multiple providers sequentially",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithProviders(Providers).WithConfig(config.Boot).Start()
}
`,
			providersContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation"

	"goravel/app/providers"
)

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&providers.AppServiceProvider{},
	}
}
`,
			pkg:      "goravel/app/providers",
			provider: "&providers.RouteServiceProvider{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithProviders(Providers).WithConfig(config.Boot).Start()
}
`,
			expectedProviders: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation"

	"goravel/app/providers"
)

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&providers.AppServiceProvider{},
		&providers.RouteServiceProvider{},
	}
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bootstrapDir := support.Config.Paths.Bootstrap
			appFile := filepath.Join(bootstrapDir, "app.go")
			providersFile := filepath.Join(bootstrapDir, "providers.go")

			require.NoError(t, supportfile.PutContent(appFile, tt.appContent))
			defer func() {
				require.NoError(t, supportfile.Remove(bootstrapDir))
			}()

			if tt.providersContent != "" {
				require.NoError(t, supportfile.PutContent(providersFile, tt.providersContent))
			}

			err := AddProvider(tt.pkg, tt.provider)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErrString != "" {
					require.Contains(t, err.Error(), tt.expectedErrString)
				}
				return
			}

			require.NoError(t, err)

			// Verify app.go content
			appContent, err := supportfile.GetContent(appFile)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedApp, appContent)

			// Verify providers.go content if expected
			if tt.expectedProviders != "" {
				providersContent, err := supportfile.GetContent(providersFile)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedProviders, providersContent)
			}
		})
	}
}

func TestAddRoute(t *testing.T) {
	tests := []struct {
		name              string
		appContent        string
		pkg               string
		route             string
		expectedApp       string
		wantErr           bool
		expectedErrString string
	}{
		{
			name: "add route when WithRouting doesn't exist",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
			pkg:   "goravel/routes",
			route: "routes.Web()",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
	"goravel/routes"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRouting(func() {
			routes.Web()
		}).WithConfig(config.Boot).Start()
}
`,
		},
		{
			name: "add route when WithRouting exists with empty slice",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
	"goravel/routes"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRouting(func() {}).WithConfig(config.Boot).Start()
}
`,
			pkg:   "goravel/routes",
			route: "routes.Web()",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
	"goravel/routes"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRouting(func() {
			routes.Web()
		}).WithConfig(config.Boot).Start()
}
`,
		},
		{
			name: "add route when WithRouting exists with existing routes",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
	"goravel/routes"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRouting(func() {
			routes.Web()
		}).WithConfig(config.Boot).Start()
}
`,
			pkg:   "goravel/routes",
			route: "routes.Api()",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
	"goravel/routes"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRouting(func() {
			routes.Web()
			routes.Api()
		}).WithConfig(config.Boot).Start()
}
`,
		},
		{
			name: "add route with different package",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
	"goravel/routes"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRouting(func() {
			routes.Web()
		}).WithConfig(config.Boot).Start()
}
`,
			pkg:   "goravel/app/routes",
			route: "routes.Admin()",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/app/routes"
	"goravel/config"
	"goravel/routes"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRouting(func() {
			routes.Web()
			routes.Admin()
		}).WithConfig(config.Boot).Start()
}
`,
		},
		{
			name: "add route to chain with multiple methods",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithCommands(Commands).
		WithConfig(config.Boot).Start()
}
`,
			pkg:   "goravel/routes",
			route: "routes.Web()",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
	"goravel/routes"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRouting(func() {
			routes.Web()
		}).
		WithCommands(Commands).
		WithConfig(config.Boot).Start()
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bootstrapDir := support.Config.Paths.Bootstrap
			appFile := filepath.Join(bootstrapDir, "app.go")

			require.NoError(t, supportfile.PutContent(appFile, tt.appContent))
			defer func() {
				require.NoError(t, supportfile.Remove(bootstrapDir))
			}()

			err := AddRoute(tt.pkg, tt.route)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErrString != "" {
					assert.Contains(t, err.Error(), tt.expectedErrString)
				}
			} else {
				require.NoError(t, err)

				// Verify app.go content
				appContent, err := supportfile.GetContent(appFile)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedApp, appContent)

				// Verify no helper file was created (always inline)
				routingFile := filepath.Join(bootstrapDir, "routing.go")
				assert.False(t, supportfile.Exists(routingFile), "routing.go should never be created for routes")
			}
		})
	}
}

func TestAddRule(t *testing.T) {
	tests := []struct {
		name              string
		appContent        string
		rulesContent      string // empty if file doesn't exist
		pkg               string
		rule              string
		expectedApp       string
		expectedRules     string // empty if file shouldn't be created
		wantErr           bool
		expectedErrString string
	}{
		{
			name: "add rule when WithRules doesn't exist and rules.go doesn't exist",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
			pkg:  "goravel/app/rules",
			rule: "&rules.Uppercase{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRules(Rules).WithConfig(config.Boot).Start()
}
`,
			expectedRules: `package bootstrap

import (
	"github.com/goravel/framework/contracts/validation"

	"goravel/app/rules"
)

func Rules() []validation.Rule {
	return []validation.Rule{
		&rules.Uppercase{},
	}
}
`,
		},
		{
			name: "add rule when WithRules exists with Rules() and rules.go exists",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRules(Rules).WithConfig(config.Boot).Start()
}
`,
			rulesContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/validation"

	"goravel/app/rules"
)

func Rules() []validation.Rule {
	return []validation.Rule{
		&rules.ExistingRule{},
	}
}
`,
			pkg:  "goravel/app/rules",
			rule: "&rules.NewRule{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRules(Rules).WithConfig(config.Boot).Start()
}
`,
			expectedRules: `package bootstrap

import (
	"github.com/goravel/framework/contracts/validation"

	"goravel/app/rules"
)

func Rules() []validation.Rule {
	return []validation.Rule{
		&rules.ExistingRule{},
		&rules.NewRule{},
	}
}
`,
		},
		{
			name: "add rule when WithRules exists with inline array",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/foundation"
	"goravel/app/rules"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRules(func() []validation.Rule {
			return []validation.Rule{
				&rules.ExistingRule{},
			}	
		}).WithConfig(config.Boot).Start()
}
`,
			pkg:  "goravel/app/rules",
			rule: "&rules.NewRule{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/foundation"
	"goravel/app/rules"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRules(func() []validation.Rule {
			return []validation.Rule{
				&rules.ExistingRule{},
				&rules.NewRule{},
			}
		}).WithConfig(config.Boot).Start()
}
`,
		},
		{
			name: "error when rules.go exists but WithRules doesn't exist",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
			rulesContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/validation"

	"goravel/app/rules"
)

func Rules() []validation.Rule {
	return []validation.Rule{
		&rules.ExistingRule{},
	}
}
`,
			pkg:     "goravel/app/rules",
			rule:    "&rules.NewRule{}",
			wantErr: true,
		},
		{
			name: "add rule when WithRules doesn't exist at the beginning of chain",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().Start()
}
`,
			pkg:  "goravel/app/rules",
			rule: "&rules.FirstRule{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRules(Rules).Start()
}
`,
			expectedRules: `package bootstrap

import (
	"github.com/goravel/framework/contracts/validation"

	"goravel/app/rules"
)

func Rules() []validation.Rule {
	return []validation.Rule{
		&rules.FirstRule{},
	}
}
`,
		},
		{
			name: "add rule from different package",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
			pkg:  "github.com/mycompany/customrules",
			rule: "&customrules.SpecialRule{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRules(Rules).WithConfig(config.Boot).Start()
}
`,
			expectedRules: `package bootstrap

import (
	"github.com/goravel/framework/contracts/validation"
	"github.com/mycompany/customrules"
)

func Rules() []validation.Rule {
	return []validation.Rule{
		&customrules.SpecialRule{},
	}
}
`,
		},
		{
			name: "add multiple rules sequentially",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
			pkg:  "goravel/app/rules",
			rule: "&rules.FirstRule{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRules(Rules).WithConfig(config.Boot).Start()
}
`,
			expectedRules: `package bootstrap

import (
	"github.com/goravel/framework/contracts/validation"

	"goravel/app/rules"
)

func Rules() []validation.Rule {
	return []validation.Rule{
		&rules.FirstRule{},
	}
}
`,
		},
		{
			name: "skip duplicate rule",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRules(Rules).WithConfig(config.Boot).Start()
}
`,
			rulesContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/validation"

	"goravel/app/rules"
)

func Rules() []validation.Rule {
	return []validation.Rule{
		&rules.ExistingRule{},
	}
}
`,
			pkg:  "goravel/app/rules",
			rule: "&rules.ExistingRule{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRules(Rules).WithConfig(config.Boot).Start()
}
`,
			expectedRules: `package bootstrap

import (
	"github.com/goravel/framework/contracts/validation"

	"goravel/app/rules"
)

func Rules() []validation.Rule {
	return []validation.Rule{
		&rules.ExistingRule{},
	}
}
`,
		},
		{
			name: "add rule with WithConfig at start",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
			pkg:  "goravel/app/rules",
			rule: "&rules.CustomRule{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRules(Rules).WithConfig(config.Boot).Start()
}
`,
			expectedRules: `package bootstrap

import (
	"github.com/goravel/framework/contracts/validation"

	"goravel/app/rules"
)

func Rules() []validation.Rule {
	return []validation.Rule{
		&rules.CustomRule{},
	}
}
`,
		},
		{
			name: "add rule when other With methods exist",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithCommands(Commands).
		WithConfig(config.Boot).Start()
}
`,
			pkg:  "goravel/app/rules",
			rule: "&rules.ExtraRule{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRules(Rules).
		WithCommands(Commands).
		WithConfig(config.Boot).Start()
}
`,
			expectedRules: `package bootstrap

import (
	"github.com/goravel/framework/contracts/validation"

	"goravel/app/rules"
)

func Rules() []validation.Rule {
	return []validation.Rule{
		&rules.ExtraRule{},
	}
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			bootstrapDir := support.Config.Paths.Bootstrap
			appFile := path.Bootstrap("app.go")
			rulesFile := filepath.Join(bootstrapDir, "rules.go")

			// Ensure clean state
			require.NoError(t, supportfile.Remove(bootstrapDir))
			require.NoError(t, supportfile.PutContent(appFile, tt.appContent))

			if tt.rulesContent != "" {
				require.NoError(t, supportfile.PutContent(rulesFile, tt.rulesContent))
			}

			// Execute
			err := AddRule(tt.pkg, tt.rule)

			// Assert error
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErrString != "" {
					assert.Contains(t, err.Error(), tt.expectedErrString)
				} else {
					assert.ErrorIs(t, err, errors.PackageRulesFileExists)
				}
				return
			}

			require.NoError(t, err)

			// Assert app.go content
			actualApp, err := supportfile.GetContent(appFile)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedApp, actualApp)

			// Assert rules.go content
			if tt.expectedRules != "" {
				actualRules, err := supportfile.GetContent(rulesFile)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedRules, actualRules)
			}

			// Cleanup
			require.NoError(t, supportfile.Remove(bootstrapDir))
		})
	}
}

func TestAddSeeder(t *testing.T) {
	tests := []struct {
		name              string
		appContent        string
		seedersContent    string // empty if file doesn't exist
		pkg               string
		seeder            string
		expectedApp       string
		expectedSeeders   string // empty if file shouldn't be created
		wantErr           bool
		expectedErrString string
	}{
		{
			name: "add seeder when WithSeeders doesn't exist and seeders.go doesn't exist",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
			pkg:    "goravel/database/seeders",
			seeder: "&seeders.DatabaseSeeder{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithSeeders(Seeders).WithConfig(config.Boot).Start()
}
`,
			expectedSeeders: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/seeder"

	"goravel/database/seeders"
)

func Seeders() []seeder.Seeder {
	return []seeder.Seeder{
		&seeders.DatabaseSeeder{},
	}
}
`,
		},
		{
			name: "add seeder when WithSeeders exists with Seeders() and seeders.go exists",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithSeeders(Seeders).WithConfig(config.Boot).Start()
}
`,
			seedersContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/seeder"

	"goravel/database/seeders"
)

func Seeders() []seeder.Seeder {
	return []seeder.Seeder{
		&seeders.ExistingSeeder{},
	}
}
`,
			pkg:    "goravel/database/seeders",
			seeder: "&seeders.NewSeeder{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithSeeders(Seeders).WithConfig(config.Boot).Start()
}
`,
			expectedSeeders: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/seeder"

	"goravel/database/seeders"
)

func Seeders() []seeder.Seeder {
	return []seeder.Seeder{
		&seeders.ExistingSeeder{},
		&seeders.NewSeeder{},
	}
}
`,
		},
		{
			name: "add seeder when WithSeeders exists with inline array",
			appContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/seeder"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
	"goravel/database/seeders"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithSeeders(func() []seeder.Seeder {
			return []seeder.Seeder{
				&seeders.ExistingSeeder{},
			}
		}).WithConfig(config.Boot).Start()
}
`,
			pkg:    "goravel/database/seeders",
			seeder: "&seeders.NewSeeder{}",
			expectedApp: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/seeder"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
	"goravel/database/seeders"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithSeeders(func() []seeder.Seeder {
			return []seeder.Seeder{
				&seeders.ExistingSeeder{},
				&seeders.NewSeeder{},
			}
		}).WithConfig(config.Boot).Start()
}
`,
		},
		{
			name: "error when seeders.go exists but WithSeeders doesn't exist",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
			seedersContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/seeder"

	"goravel/database/seeders"
)

func Seeders() []seeder.Seeder {
	return []seeder.Seeder{
		&seeders.ExistingSeeder{},
	}
}
`,
			pkg:               "goravel/database/seeders",
			seeder:            "&seeders.NewSeeder{}",
			wantErr:           true,
			expectedErrString: "seeders.go already exists but WithSeeders is not registered in foundation.Setup()",
		},
		{
			name: "add seeder when WithSeeders doesn't exist at the beginning of chain",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().Start()
}
`,
			pkg:    "goravel/database/seeders",
			seeder: "&seeders.FirstSeeder{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithSeeders(Seeders).Start()
}
`,
			expectedSeeders: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/seeder"

	"goravel/database/seeders"
)

func Seeders() []seeder.Seeder {
	return []seeder.Seeder{
		&seeders.FirstSeeder{},
	}
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bootstrapDir := support.Config.Paths.Bootstrap
			appFile := filepath.Join(bootstrapDir, "app.go")
			seedersFile := filepath.Join(bootstrapDir, "seeders.go")

			assert.NoError(t, supportfile.PutContent(appFile, tt.appContent))
			defer func() {
				assert.NoError(t, supportfile.Remove(bootstrapDir))
			}()

			if tt.seedersContent != "" {
				assert.NoError(t, supportfile.PutContent(seedersFile, tt.seedersContent))
			}

			err := AddSeeder(tt.pkg, tt.seeder)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErrString != "" {
					assert.Contains(t, err.Error(), tt.expectedErrString)
				}
				return
			}

			assert.NoError(t, err)

			// Verify app.go content
			appContent, err := supportfile.GetContent(appFile)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedApp, appContent)

			// Verify seeders.go content if expected
			if tt.expectedSeeders != "" {
				seedersContent, err := supportfile.GetContent(seedersFile)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSeeders, seedersContent)
			}
		})
	}
}

func TestExprExists(t *testing.T) {
	assert.NotPanics(t, func() {
		t.Run("expr exists", func(t *testing.T) {
			assert.True(t,
				ExprExists(
					MustParseExpr("[]any{&some.Struct{}}").(*dst.CompositeLit).Elts,
					MustParseExpr("&some.Struct{}").(dst.Expr),
				),
			)
			assert.NotEqual(t, -1,
				ExprIndex(
					MustParseExpr("[]any{&some.Struct{}}").(*dst.CompositeLit).Elts,
					MustParseExpr("&some.Struct{}").(dst.Expr),
				),
			)
		})
		t.Run("expr does not exist", func(t *testing.T) {
			assert.False(t,
				ExprExists(
					MustParseExpr("[]any{&some.OtherStruct{}}").(*dst.CompositeLit).Elts,
					MustParseExpr("&some.Struct{}").(dst.Expr),
				),
			)
			assert.Equal(t, -1,
				ExprIndex(
					MustParseExpr("[]any{&some.OtherStruct{}}").(*dst.CompositeLit).Elts,
					MustParseExpr("&some.Struct{}").(dst.Expr),
				),
			)
		})
	})
}

func TestKeyExists(t *testing.T) {
	assert.NotPanics(t, func() {
		t.Run("key exists", func(t *testing.T) {
			assert.True(t,
				KeyExists(
					MustParseExpr(`map[string]any{"someKey":"exist"}`).(*dst.CompositeLit).Elts,
					&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote("someKey")},
				),
			)
			assert.NotEqual(t, -1,
				KeyIndex(
					MustParseExpr(`map[string]any{"someKey":"exist"}`).(*dst.CompositeLit).Elts,
					&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote("someKey")},
				),
			)
		})
		t.Run("key does not exist", func(t *testing.T) {
			assert.False(t,
				KeyExists(
					MustParseExpr(`map[string]any{"otherKey":"exist"}`).(*dst.CompositeLit).Elts,
					&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote("someKey")},
				),
			)
			assert.Equal(t, -1,
				KeyIndex(
					MustParseExpr(`map[string]any{"otherKey":"exist"}`).(*dst.CompositeLit).Elts,
					&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote("someKey")},
				),
			)
		})
	})
}

func TestMustParseStatement(t *testing.T) {
	t.Run("parse failed", func(t *testing.T) {
		assert.Panics(t, func() {
			MustParseExpr("var invalid:=syntax")
		})
	})

	t.Run("parse success", func(t *testing.T) {
		assert.NotPanics(t, func() {
			assert.NotNil(t, MustParseExpr(`struct{x *int}`))
		})
	})
}

func TestRemoveProvider(t *testing.T) {
	tests := []struct {
		name              string
		appContent        string
		providersContent  string // empty if file doesn't exist
		pkg               string
		provider          string
		expectedApp       string
		expectedProviders string // expected content after removal, empty if file doesn't exist
	}{
		{
			name: "remove provider from providers.go when multiple providers exist",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithProviders(Providers).WithConfig(config.Boot).Start()
}
`,
			providersContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation"

	"goravel/app/providers"
)

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&providers.AppServiceProvider{},
		&providers.RouteServiceProvider{},
	}
}
`,
			pkg:      "goravel/app/providers",
			provider: "&providers.AppServiceProvider{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithProviders(Providers).WithConfig(config.Boot).Start()
}
`,
			expectedProviders: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation"

	"goravel/app/providers"
)

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&providers.RouteServiceProvider{},
	}
}
`,
		},
		{
			name: "remove provider from inline array when multiple providers exist",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/app/providers"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithProviders(func() []foundation.ServiceProvider {
			return []foundation.ServiceProvider{
				&providers.AppServiceProvider{},
				&providers.RouteServiceProvider{},
			}
		}).WithConfig(config.Boot).Start()
}
`,
			pkg:      "goravel/app/providers",
			provider: "&providers.AppServiceProvider{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/app/providers"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithProviders(func() []foundation.ServiceProvider {
			return []foundation.ServiceProvider{
				&providers.RouteServiceProvider{},
			}
		}).WithConfig(config.Boot).Start()
}
`,
		},
		{
			name: "remove last provider from providers.go",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithProviders(Providers).WithConfig(config.Boot).Start()
}
`,
			providersContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation"

	"goravel/app/providers"
)

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&providers.AppServiceProvider{},
	}
}
`,
			pkg:      "goravel/app/providers",
			provider: "&providers.AppServiceProvider{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithProviders(Providers).WithConfig(config.Boot).Start()
}
`,
			expectedProviders: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation"
)

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{}
}
`,
		},
		{
			name: "remove provider from different package",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithProviders(Providers).WithConfig(config.Boot).Start()
}
`,
			providersContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/redis"

	"goravel/app/providers"
)

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&providers.AppServiceProvider{},
		&redis.ServiceProvider{},
	}
}
`,
			pkg:      "github.com/goravel/redis",
			provider: "&redis.ServiceProvider{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithProviders(Providers).WithConfig(config.Boot).Start()
}
`,
			expectedProviders: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation"

	"goravel/app/providers"
)

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&providers.AppServiceProvider{},
	}
}
`,
		},
		{
			name: "no-op when WithProviders doesn't exist",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
			pkg:      "goravel/app/providers",
			provider: "&providers.AppServiceProvider{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
		},
		{
			name: "no-op when provider doesn't exist in the list",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithProviders(Providers).WithConfig(config.Boot).Start()
}
`,
			providersContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation"

	"goravel/app/providers"
)

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&providers.RouteServiceProvider{},
	}
}
`,
			pkg:      "goravel/app/providers",
			provider: "&providers.AppServiceProvider{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithProviders(Providers).WithConfig(config.Boot).Start()
}
`,
			expectedProviders: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation"

	"goravel/app/providers"
)

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&providers.RouteServiceProvider{},
	}
}
`,
		},
		{
			name: "remove provider but keep import when another provider from same package exists",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithProviders(Providers).WithConfig(config.Boot).Start()
}
`,
			providersContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation"

	"goravel/app/providers"
)

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&providers.AppServiceProvider{},
		&providers.RouteServiceProvider{},
	}
}
`,
			pkg:      "goravel/app/providers",
			provider: "&providers.AppServiceProvider{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithProviders(Providers).WithConfig(config.Boot).Start()
}
`,
			expectedProviders: `package bootstrap

import (
	"github.com/goravel/framework/contracts/foundation"

	"goravel/app/providers"
)

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&providers.RouteServiceProvider{},
	}
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bootstrapDir := support.Config.Paths.Bootstrap
			appFile := filepath.Join(bootstrapDir, "app.go")
			providersFile := filepath.Join(bootstrapDir, "providers.go")

			require.NoError(t, supportfile.PutContent(appFile, tt.appContent))
			defer func() {
				require.NoError(t, supportfile.Remove(bootstrapDir))
			}()

			if tt.providersContent != "" {
				require.NoError(t, supportfile.PutContent(providersFile, tt.providersContent))
			}

			err := RemoveProvider(tt.pkg, tt.provider)
			require.NoError(t, err)

			// Verify app.go content
			appContent, err := supportfile.GetContent(appFile)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedApp, appContent)

			// Verify providers.go content if expected
			if tt.expectedProviders != "" {
				providersContent, err := supportfile.GetContent(providersFile)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedProviders, providersContent)
			}
		})
	}
}

func TestRemoveMigration(t *testing.T) {
	tests := []struct {
		name               string
		appContent         string
		migrationsContent  string // empty if file doesn't exist
		pkg                string
		migration          string
		expectedApp        string
		expectedMigrations string // expected content after removal, empty if file doesn't exist
	}{
		{
			name: "remove migration from migrations.go when multiple migrations exist",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMigrations(Migrations).WithConfig(config.Boot).Start()
}
`,
			migrationsContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.CreateUsersTable{},
		&migrations.CreatePostsTable{},
	}
}
`,
			pkg:       "goravel/database/migrations",
			migration: "&migrations.CreateUsersTable{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMigrations(Migrations).WithConfig(config.Boot).Start()
}
`,
			expectedMigrations: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.CreatePostsTable{},
	}
}
`,
		},
		{
			name: "remove migration from inline array when multiple migrations exist",
			appContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
	"goravel/database/migrations"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMigrations(func() []schema.Migration {
			return []schema.Migration{
				&migrations.CreateUsersTable{},
				&migrations.CreatePostsTable{},
			}
		}).WithConfig(config.Boot).Start()
}
`,
			pkg:       "goravel/database/migrations",
			migration: "&migrations.CreateUsersTable{}",
			expectedApp: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
	"goravel/database/migrations"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMigrations(func() []schema.Migration {
			return []schema.Migration{
				&migrations.CreatePostsTable{},
			}
		}).WithConfig(config.Boot).Start()
}
`,
		},
		{
			name: "remove last migration from migrations.go",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMigrations(Migrations).WithConfig(config.Boot).Start()
}
`,
			migrationsContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.CreateUsersTable{},
	}
}
`,
			pkg:       "goravel/database/migrations",
			migration: "&migrations.CreateUsersTable{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMigrations(Migrations).WithConfig(config.Boot).Start()
}
`,
			expectedMigrations: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"
)

func Migrations() []schema.Migration {
	return []schema.Migration{}
}
`,
		},
		{
			name: "remove migration from different package",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMigrations(Migrations).WithConfig(config.Boot).Start()
}
`,
			migrationsContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/third-party/dbmigrations"

	"goravel/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.CreateUsersTable{},
		&dbmigrations.ThirdPartyMigration{},
	}
}
`,
			pkg:       "github.com/third-party/dbmigrations",
			migration: "&dbmigrations.ThirdPartyMigration{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMigrations(Migrations).WithConfig(config.Boot).Start()
}
`,
			expectedMigrations: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.CreateUsersTable{},
	}
}
`,
		},
		{
			name: "no-op when WithMigrations doesn't exist",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
			pkg:       "goravel/database/migrations",
			migration: "&migrations.CreateUsersTable{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
		},
		{
			name: "no-op when migration doesn't exist in the list",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMigrations(Migrations).WithConfig(config.Boot).Start()
}
`,
			migrationsContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.CreatePostsTable{},
	}
}
`,
			pkg:       "goravel/database/migrations",
			migration: "&migrations.CreateUsersTable{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMigrations(Migrations).WithConfig(config.Boot).Start()
}
`,
			expectedMigrations: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.CreatePostsTable{},
	}
}
`,
		},
		{
			name: "remove migration but keep import when another migration from same package exists",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMigrations(Migrations).WithConfig(config.Boot).Start()
}
`,
			migrationsContent: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.CreateUsersTable{},
		&migrations.CreatePostsTable{},
	}
}
`,
			pkg:       "goravel/database/migrations",
			migration: "&migrations.CreateUsersTable{}",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMigrations(Migrations).WithConfig(config.Boot).Start()
}
`,
			expectedMigrations: `package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.CreatePostsTable{},
	}
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bootstrapDir := support.Config.Paths.Bootstrap
			appFile := filepath.Join(bootstrapDir, "app.go")
			migrationsFile := filepath.Join(bootstrapDir, "migrations.go")

			require.NoError(t, supportfile.PutContent(appFile, tt.appContent))
			defer func() {
				require.NoError(t, supportfile.Remove(bootstrapDir))
			}()

			if tt.migrationsContent != "" {
				require.NoError(t, supportfile.PutContent(migrationsFile, tt.migrationsContent))
			}

			err := RemoveMigration(tt.pkg, tt.migration)
			require.NoError(t, err)

			// Verify app.go content
			appContent, err := supportfile.GetContent(appFile)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedApp, appContent)

			// Verify migrations.go content if expected
			if tt.expectedMigrations != "" {
				migrationsContent, err := supportfile.GetContent(migrationsFile)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedMigrations, migrationsContent)
			}
		})
	}
}

func TestRemoveRoute(t *testing.T) {
	tests := []struct {
		name        string
		appContent  string
		pkg         string
		route       string
		expectedApp string
	}{
		{
			name: "remove route from WithRouting when multiple routes exist",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
	"goravel/routes"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRouting(func() {
			routes.Web()
			routes.Api()
		}).WithConfig(config.Boot).Start()
}
`,
			pkg:   "goravel/routes",
			route: "routes.Web()",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
	"goravel/routes"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRouting(func() {
			routes.Api()
		}).WithConfig(config.Boot).Start()
}
`,
		},
		{
			name: "remove last route from WithRouting",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
	"goravel/routes"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRouting(func() {
			routes.Web()
		}).WithConfig(config.Boot).Start()
}
`,
			pkg:   "goravel/routes",
			route: "routes.Web()",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
	"goravel/routes"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRouting(func() {}).WithConfig(config.Boot).Start()
}
`,
		},
		{
			name: "remove route with different package prefix",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/app/routes"
	"goravel/config"
	"goravel/routes"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRouting(func() {
			routes.Web()
			routes.Admin()
		}).WithConfig(config.Boot).Start()
}
`,
			pkg:   "goravel/app/routes",
			route: "routes.Admin()",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/app/routes"
	"goravel/config"
	"goravel/routes"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRouting(func() {
			routes.Web()
		}).WithConfig(config.Boot).Start()
}
`,
		},
		{
			name: "remove middle route from multiple routes",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
	"goravel/routes"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRouting(func() {
			routes.Web()
			routes.Api()
			routes.Admin()
		}).WithConfig(config.Boot).Start()
}
`,
			pkg:   "goravel/routes",
			route: "routes.Api()",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
	"goravel/routes"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRouting(func() {
			routes.Web()
			routes.Admin()
		}).WithConfig(config.Boot).Start()
}
`,
		},
		{
			name: "no-op when WithRouting doesn't exist",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
			pkg:   "goravel/routes",
			route: "routes.Web()",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithConfig(config.Boot).Start()
}
`,
		},
		{
			name: "no-op when route doesn't exist",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
	"goravel/routes"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRouting(func() {
			routes.Web()
		}).WithConfig(config.Boot).Start()
}
`,
			pkg:   "goravel/routes",
			route: "routes.Api()",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
	"goravel/routes"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRouting(func() {
			routes.Web()
		}).WithConfig(config.Boot).Start()
}
`,
		},
		{
			name: "remove route from chain with multiple methods",
			appContent: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
	"goravel/routes"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRouting(func() {
			routes.Web()
			routes.Api()
		}).
		WithCommands(Commands).
		WithConfig(config.Boot).Start()
}
`,
			pkg:   "goravel/routes",
			route: "routes.Web()",
			expectedApp: `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"goravel/config"
	"goravel/routes"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithRouting(func() {
			routes.Api()
		}).
		WithCommands(Commands).
		WithConfig(config.Boot).Start()
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bootstrapDir := support.Config.Paths.Bootstrap
			appFile := filepath.Join(bootstrapDir, "app.go")

			require.NoError(t, supportfile.PutContent(appFile, tt.appContent))
			defer func() {
				require.NoError(t, supportfile.Remove(bootstrapDir))
			}()

			err := RemoveRoute(tt.pkg, tt.route)
			require.NoError(t, err)

			// Verify app.go content
			appContent, err := supportfile.GetContent(appFile)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedApp, appContent)
		})
	}
}

func TestUsesImport(t *testing.T) {
	df, err := decorator.Parse(`package main
import (
    "fmt"        
    mylog "log"
)

func main() {
    fmt.Println("hello")
}`)
	require.NoError(t, err)
	require.NotNil(t, df)

	assert.True(t, IsUsingImport(df, "fmt"))
	assert.False(t, IsUsingImport(df, "log", "mylog"))
}

func TestWrapNewline(t *testing.T) {
	src := `package main

var value = 1
var _ = map[string]any{"key": &value, "func": func() bool { return true }}
`

	df, err := decorator.Parse(src)
	assert.NoError(t, err)

	// without WrapNewline
	var buf bytes.Buffer
	assert.NoError(t, decorator.Fprint(&buf, df))
	assert.Equal(t, src, buf.String())

	// with WrapNewline
	WrapNewline(df)
	buf.Reset()
	assert.NoError(t, decorator.Fprint(&buf, df))
	assert.NotEqual(t, src, buf.String())
	assert.Equal(t, `package main

var value = 1
var _ = map[string]any{
	"key": &value,
	"func": func() bool {
		return true
	},
}
`, buf.String())

}
