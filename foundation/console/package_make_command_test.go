package console

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	consolemocks "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/file"
)

func TestPackageMakeCommand(t *testing.T) {
	var (
		mockContext *consolemocks.Context
	)

	beforeEach := func() {
		mockContext = &consolemocks.Context{}
	}

	tests := []struct {
		name   string
		setup  func()
		assert func()
	}{
		{
			name: "name is empty",
			setup: func() {
				mockContext.On("Argument", 0).Return("").Once()
				mockContext.On("Ask", "Enter the package name", mock.Anything).Return("", errors.New("the package name cannot be empty")).Once()
			},
			assert: func() {
				err := NewPackageMakeCommand().Handle(mockContext)
				assert.EqualError(t, err, "the package name cannot be empty")
			},
		},
		{
			name: "name is sms and use default root",
			setup: func() {
				mockContext.On("Argument", 0).Return("sms").Once()
				mockContext.On("OptionBool", "force").Return(false).Once()
				mockContext.On("Option", "root").Return("packages").Once()
			},
			assert: func() {
				err := NewPackageMakeCommand().Handle(mockContext)
				assert.Nil(t, err)
				assert.True(t, file.Exists("packages/sms/README.md"))
				assert.True(t, file.Exists("packages/sms/service_provider.go"))
				assert.True(t, file.Exists("packages/sms/sms.go"))
				assert.True(t, file.Exists("packages/sms/config/sms.go"))
				assert.True(t, file.Exists("packages/sms/contracts/sms.go"))
				assert.True(t, file.Exists("packages/sms/facades/sms.go"))
				assert.True(t, file.Contain("packages/sms/facades/sms.go", "goravel/packages/sms"))
				assert.True(t, file.Contain("packages/sms/facades/sms.go", "goravel/packages/sms/contracts"))
				assert.Nil(t, file.Remove("packages"))
			},
		},
		{
			name: "name is github.com/goravel/sms and use other root",
			setup: func() {
				mockContext.On("Argument", 0).Return("github.com/goravel/sms-aws").Once()
				mockContext.On("OptionBool", "force").Return(true).Once()
				mockContext.On("Option", "root").Return("package").Once()
			},
			assert: func() {
				_ = NewPackageMakeCommand().Handle(mockContext)
				assert.True(t, file.Exists("package/github_com_goravel_sms_aws/README.md"))
				assert.True(t, file.Exists("package/github_com_goravel_sms_aws/service_provider.go"))
				assert.True(t, file.Exists("package/github_com_goravel_sms_aws/github_com_goravel_sms_aws.go"))
				assert.True(t, file.Exists("package/github_com_goravel_sms_aws/config/github_com_goravel_sms_aws.go"))
				assert.True(t, file.Exists("package/github_com_goravel_sms_aws/contracts/github_com_goravel_sms_aws.go"))
				assert.True(t, file.Exists("package/github_com_goravel_sms_aws/facades/github_com_goravel_sms_aws.go"))
				assert.Nil(t, file.Remove("package"))
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beforeEach()
			test.setup()
			if test.assert != nil {
				test.assert()
			}
			mockContext.AssertExpectations(t)
		})
	}
}

func TestPackageName(t *testing.T) {
	input := "github.com/example/package-name"
	expected := "package_name"
	assert.Equal(t, expected, packageName(input))

	input2 := "example.com/another_package.name"
	expected2 := "another_package_name"
	assert.Equal(t, expected2, packageName(input2))
}
