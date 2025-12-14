package packages

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/packages"
	mockmodify "github.com/goravel/framework/mocks/packages/modify"
	"github.com/goravel/framework/support/color"
)

type PackagesSetupTestSuite struct {
	suite.Suite
}

func TestPackagesSetupTestSuite(t *testing.T) {
	suite.Run(t, new(PackagesSetupTestSuite))
}

func (s *PackagesSetupTestSuite) SetupTest() {
	osExit = func(code int) { panic(code) }
}

func (s *PackagesSetupTestSuite) TearDownTest() {
	osExit = os.Exit
}

func (s *PackagesSetupTestSuite) TestExecute() {
	tests := []struct {
		name    string
		command string
		force   bool
		setup   func(st packages.Setup) packages.Setup
		err     error
		output  string
	}{
		{
			name:    "install failed",
			command: "install",
			setup: func(st packages.Setup) packages.Setup {
				mockModify := mockmodify.NewGoFile(s.T())
				mockModify.EXPECT().Apply(mock.AnythingOfType("modify.Option"), mock.AnythingOfType("modify.Option"), mock.AnythingOfType("modify.Option")).Return(assert.AnError).Once()
				return st.Install(mockModify)
			},
			err:    assert.AnError,
			output: "ERROR",
		},
		{
			name:    "install success",
			command: "install",
			setup: func(st packages.Setup) packages.Setup {
				mockModify := mockmodify.NewGoFile(s.T())
				mockModify.EXPECT().Apply(mock.AnythingOfType("modify.Option"), mock.AnythingOfType("modify.Option"), mock.AnythingOfType("modify.Option")).Return(nil).Once()
				return st.Install(mockModify)
			},
			output: "package installed successfully",
		},
		{
			name:    "uninstall failed",
			command: "uninstall",
			setup: func(st packages.Setup) packages.Setup {
				mockModify := mockmodify.NewGoFile(s.T())
				mockModify.EXPECT().Apply(mock.AnythingOfType("modify.Option"), mock.AnythingOfType("modify.Option"), mock.AnythingOfType("modify.Option")).Return(assert.AnError).Once()
				return st.Uninstall(mockModify)
			},
			err:    assert.AnError,
			output: "ERROR",
		},
		{
			name:    "uninstall failed with force",
			command: "uninstall",
			force:   true,
			setup: func(st packages.Setup) packages.Setup {
				mockModify := mockmodify.NewGoFile(s.T())
				mockModify.EXPECT().Apply(mock.AnythingOfType("modify.Option"), mock.AnythingOfType("modify.Option"), mock.AnythingOfType("modify.Option")).Return(assert.AnError).Once()
				return st.Uninstall(mockModify)
			},
			err:    assert.AnError,
			output: "WARNING",
		},
		{
			name:    "uninstall success",
			command: "uninstall",
			setup: func(st packages.Setup) packages.Setup {
				mockModify := mockmodify.NewGoFile(s.T())
				mockModify.EXPECT().Apply(mock.AnythingOfType("modify.Option"), mock.AnythingOfType("modify.Option"), mock.AnythingOfType("modify.Option")).Return(nil).Once()
				return st.Uninstall(mockModify)
			},
			output: "package uninstalled successfully",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			args := []string{tt.command}
			if tt.force {
				args = append(args, "--force")
			}

			output := color.CaptureOutput(func(w io.Writer) {
				func() {
					defer func() { _ = recover() }()
					st := Setup(args)
					tt.setup(st).Execute()
				}()
			})

			s.Contains(output, tt.output)
			if tt.err != nil {
				s.Contains(output, tt.err.Error())
			}
		})
	}
}

func TestSetup(t *testing.T) {
	s := Setup([]string{"install", "--force", "--facade=test", "--driver=database"})
	assert.Equal(t, &setup{
		command: "install",
		driver:  "database",
		facade:  "test",
		force:   true,
		paths:   Paths("goravel"),
	}, s.(*setup))

	s = Setup([]string{"uninstall", "-f", "--facade=test", "--driver=database"})
	assert.Equal(t, &setup{
		command: "uninstall",
		driver:  "database",
		facade:  "test",
		force:   true,
		paths:   Paths("goravel"),
	}, s.(*setup))

	s = Setup([]string{"install", "--package-name=custom-package", "--facade=test"})
	assert.Equal(t, &setup{
		command: "install",
		facade:  "test",
		paths:   Paths("custom-package"),
	}, s.(*setup))
}
