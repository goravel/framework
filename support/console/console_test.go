package console

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/errors"
	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/file"
)

type MakeTestSuite struct {
	suite.Suite
	make *Make
}

func TestMakeTestSuite(t *testing.T) {
	suite.Run(t, new(MakeTestSuite))
}

func (s *MakeTestSuite) SetupTest() {
	s.make = &Make{
		name: "Lowercase",
		root: filepath.Join("app", "rules"),
	}
}

func (s *MakeTestSuite) TestGetFilePath() {
	pwd, _ := os.Getwd()
	s.Equal(filepath.Join(pwd, s.make.root, "lowercase.go"), s.make.GetFilePath())

	s.make.name = "user/Lowercase"
	s.Equal(filepath.Join(pwd, s.make.root, "user", "lowercase.go"), s.make.GetFilePath())
}

func (s *MakeTestSuite) TestGetStructName() {
	s.Equal("Lowercase", s.make.GetStructName())

	s.make.name = "lowercase"
	s.Equal("Lowercase", s.make.GetStructName())

	s.make.name = "user/Lowercase"
	s.Equal("Lowercase", s.make.GetStructName())
}

func (s *MakeTestSuite) TestGetPackageName() {
	s.Equal("rules", s.make.GetPackageName())

	s.make.name = "user/Lowercase"
	s.Equal("user", s.make.GetPackageName())
}

func (s *MakeTestSuite) TestGetFolderPath() {
	s.Empty(s.make.GetFolderPath())

	s.make.name = "user/Lowercase"
	s.Equal("user", s.make.GetFolderPath())
}

func TestNewMake(t *testing.T) {
	var (
		name string

		mockCtx = &consolemocks.Context{}
		ttype   = "rule"
		root    = filepath.Join("app", "rules")
	)

	tests := []struct {
		name        string
		setup       func()
		expectMake  *Make
		expectError error
	}{
		{
			name: "Sad path - name is empty",
			setup: func() {
				name = ""
				mockCtx.EXPECT().Ask("Enter the rule name", mock.Anything).Return("", errors.ConsoleEmptyFieldValue.Args("rule")).Once()
			},
			expectMake:  nil,
			expectError: errors.ConsoleEmptyFieldValue.Args("rule"),
		},
		{
			name: "Sad path - name already exists",
			setup: func() {
				name = "Uppercase"
				assert.Nil(t, file.Create(filepath.Join(root, "uppercase.go"), ""))
				mockCtx.EXPECT().OptionBool("force").Return(false).Once()
			},
			expectMake:  nil,
			expectError: errors.ConsoleFileAlreadyExists.Args("rule"),
		},
		{
			name: "Happy path - name already exists, but force is true",
			setup: func() {
				name = "Uppercase"
				assert.Nil(t, file.Create(filepath.Join(root, "uppercase.go"), ""))
				mockCtx.EXPECT().OptionBool("force").Return(true).Once()
			},
			expectMake:  &Make{name: "Lowercase", root: root},
			expectError: nil,
		},
		{
			name: "Happy path - name is not empty",
			setup: func() {
				name = "Lowercase"
				mockCtx.On("OptionBool", "force").Return(false).Once()
			},
			expectMake:  &Make{name: "Lowercase", root: root},
			expectError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setup()
			m, err := NewMake(mockCtx, ttype, name, root)
			if test.expectError != nil {
				assert.Equal(t, test.expectError, err)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, m)
				assert.Nil(t, file.Remove("app"))
			}

			mockCtx.AssertExpectations(t)
		})
	}
}

func TestConfirmToProceed(t *testing.T) {
	var (
		mockCtx *consolemocks.Context
	)

	beforeEach := func() {
		mockCtx = &consolemocks.Context{}
	}

	tests := []struct {
		name         string
		env          string
		setup        func()
		expectResult bool
	}{
		{
			name:         "env is not production",
			setup:        func() {},
			expectResult: true,
		},
		{
			name: "the force option is true",
			env:  "production",
			setup: func() {
				mockCtx.EXPECT().OptionBool("force").Return(true).Once()
			},
			expectResult: true,
		},
		{
			name: "confirm returns err",
			env:  "production",
			setup: func() {
				mockCtx.EXPECT().OptionBool("force").Return(false).Once()
				mockCtx.EXPECT().Confirm("Are you sure you want to run this command?").Return(false, assert.AnError).Once()
				mockCtx.EXPECT().Error(errors.ConsoleFailedToConfirm.Args(assert.AnError).Error()).Once()
			},
			expectResult: false,
		},
		{
			name: "confirm returns true",
			env:  "production",
			setup: func() {
				mockCtx.EXPECT().OptionBool("force").Return(false).Once()
				mockCtx.EXPECT().Confirm("Are you sure you want to run this command?").Return(true, nil).Once()
			},
			expectResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeEach()
			tt.setup()
			result := ConfirmToProceed(mockCtx, tt.env)

			assert.Equal(t, tt.expectResult, result)
		})
	}
}
