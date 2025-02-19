package console

import (
	"os"
	"reflect"
	"testing"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/file"
	"github.com/stretchr/testify/suite"
)

type EnvEncryptCommandTestSuite struct {
	suite.Suite
}

func TestEnvEncryptCommandTestSuite(t *testing.T) {
	suite.Run(t, new(EnvEncryptCommandTestSuite))
}

func (s *EnvEncryptCommandTestSuite) SetupTest() {

}
func (s *EnvEncryptCommandTestSuite) TestSignature() {
	expected := "env:encrypt"
	s.Require().Equal(expected, NewEnvEncryptCommand().Signature())
}

func (s *EnvEncryptCommandTestSuite) TestDescription() {
	expected := "Encrypt an environment file"
	s.Require().Equal(expected, NewEnvEncryptCommand().Description())
}

func (s *EnvEncryptCommandTestSuite) TestExtend() {
	cmd := NewEnvEncryptCommand()
	got := cmd.Extend()

	s.Run("should return correct category", func() {
		expected := "env"
		s.Require().Equal(expected, got.Category)
	})

	if len(got.Flags) > 0 {
		s.Run("should have correctly configured StringFlag", func() {
			flag, ok := got.Flags[0].(*command.StringFlag)
			if !ok {
				s.Fail("First flag is not StringFlag (got type: %T)", got.Flags[0])
			}

			testCases := []struct {
				name     string
				got      interface{}
				expected interface{}
			}{
				{"Name", flag.Name, "key"},
				{"Aliases", flag.Aliases, []string{"k"}},
				{"Value", flag.Value, ""},
				{"Usage", flag.Usage, "Encryption key"},
			}

			for _, tc := range testCases {
				if !reflect.DeepEqual(tc.got, tc.expected) {
					s.Require().Equal(tc.expected, tc.got)
				}
			}
		})
	}
}

func (s *EnvEncryptCommandTestSuite) TestHandle() {
	envEncryptCommand := NewEnvEncryptCommand()
	mockContext := mocksconsole.NewContext(s.T())

	key := "BgcELROHL8sAV568T7Fiki7krjLHOkUc"

	err := file.Create(".env", "APP_KEY=12345\n")
	s.Nil(err)

	_, err = os.ReadFile(".env")
	s.Nil(err)

	_, err = os.ReadFile(".env.encrypted")
	if err == nil {
		mockContext.EXPECT().Confirm("Encrypted environment file already exists, are you sure to overwrite?", console.ConfirmOption{
			Default:     true,
			Affirmative: "Yes",
			Negative:    "No",
		}).Return(true, nil).Once()
	}

	mockContext.EXPECT().Option("key").Return(key).Once()
	mockContext.EXPECT().Success("Environment successfully encrypted.").Once()
	mockContext.EXPECT().TwoColumnDetail("Key", key).Once()
	mockContext.EXPECT().TwoColumnDetail("Cipher", "AES-256-CBC").Once()
	mockContext.EXPECT().TwoColumnDetail("Encrypted file", ".env.encrypted").Once()

	s.Nil(envEncryptCommand.Handle(mockContext))
}
