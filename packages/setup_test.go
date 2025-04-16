package packages

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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
		name   string
		setup  func() packages.Setup
		assert func(output string)
	}{
		{
			name: "module name is empty",
			setup: func() packages.Setup {
				return Setup([]string{"uninstall", "--force"})
			},
			assert: func(output string) {
				s.Contains(output, "package module name is empty")
				s.Contains(output, "please run command with module name")
			},
		},
		{
			name: "install failed",
			setup: func() packages.Setup {
				var (
					mockModify = mockmodify.NewGoFile(s.T())
					set        = &setup{
						module:  "test",
						command: "install",
					}
				)
				mockModify.EXPECT().Apply().Return(assert.AnError).Once()
				set.Install(mockModify)

				return set
			},
			assert: func(output string) {
				s.Contains(output, "ERROR")
				s.Contains(output, "assert.AnError general error for testing")
			},
		},
		{
			name: "install success",
			setup: func() packages.Setup {
				var (
					mockModify = mockmodify.NewGoFile(s.T())
					set        = &setup{
						module:  "test",
						command: "install",
					}
				)
				mockModify.EXPECT().Apply().Return(nil).Once()
				set.Install(mockModify)

				return set
			},
			assert: func(output string) {
				s.Contains(output, "package installed successfully")
			},
		},
		{
			name: "uninstall failed",
			setup: func() packages.Setup {
				var (
					mockModify = mockmodify.NewGoFile(s.T())
					set        = &setup{
						module:  "test",
						command: "uninstall",
					}
				)
				mockModify.EXPECT().Apply().Return(assert.AnError).Once()
				set.Uninstall(mockModify)

				return set
			},
			assert: func(output string) {
				s.Contains(output, "ERROR")
				s.Contains(output, "assert.AnError general error for testing")
			},
		},
		{
			name: "uninstall failed with force",
			setup: func() packages.Setup {
				var (
					mockModify = mockmodify.NewGoFile(s.T())
					set        = &setup{
						module:  "test",
						command: "uninstall",
						force:   true,
					}
				)
				mockModify.EXPECT().Apply().Return(assert.AnError).Once()
				set.Uninstall(mockModify)

				return set
			},
			assert: func(output string) {
				s.Contains(output, "WARNING")
				s.Contains(output, "assert.AnError general error for testing")
			},
		},
		{
			name: "uninstall success",
			setup: func() packages.Setup {
				var (
					mockModify = mockmodify.NewGoFile(s.T())
					set        = &setup{
						module:  "test",
						command: "uninstall",
					}
				)
				mockModify.EXPECT().Apply().Return(nil).Once()
				set.Uninstall(mockModify)

				return set
			},
			assert: func(output string) {
				s.Contains(output, "package uninstalled successfully")
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.assert(color.CaptureOutput(func(w io.Writer) {
				func() {
					defer func() { _ = recover() }()
					tt.setup().Execute()
				}()
			}))
		})
	}
}
