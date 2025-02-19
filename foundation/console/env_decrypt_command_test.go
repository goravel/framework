package console

import (
	"github.com/goravel/framework/support/file"
	"os"
	"reflect"
	"testing"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/stretchr/testify/suite"
)

type EnvDecryptCommandTestSuite struct {
	suite.Suite
}

func TestEnvDecryptCommandTestSuite(t *testing.T) {
	suite.Run(t, new(EnvDecryptCommandTestSuite))
}

func (s *EnvDecryptCommandTestSuite) SetupTest() {

}
func (s *EnvDecryptCommandTestSuite) TestSignature() {
	expected := "env:decrypt"
	s.Require().Equal(expected, NewEnvDecryptCommand().Signature())
}

func (s *EnvDecryptCommandTestSuite) TestDescription() {
	expected := "Decrypt an environment file"
	s.Require().Equal(expected, NewEnvDecryptCommand().Description())
}

func (s *EnvDecryptCommandTestSuite) TestExtend() {
	cmd := NewEnvDecryptCommand()
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
				{"Usage", flag.Usage, "Decryption key"},
			}

			for _, tc := range testCases {
				if !reflect.DeepEqual(tc.got, tc.expected) {
					s.Require().Equal(tc.expected, tc.got)
				}
			}
		})
	}
}

func (s *EnvDecryptCommandTestSuite) TestHandle() {
	envDecryptCommand := NewEnvDecryptCommand()
	mockContext := mocksconsole.NewContext(s.T())

	key := "BgcELROHL8sAV568T7Fiki7krjLHOkUc"

	_, err := os.ReadFile(".env.encrypted")
	if err != nil {
		encryptCommandTestSuite := &EnvEncryptCommandTestSuite{}
		TestEnvEncryptCommandTestSuite(s.T())
		encryptCommandTestSuite.SetupTest()
	}

	env, err := os.ReadFile(".env")
	if err == nil {
		mockContext.EXPECT().Confirm("Environment file already exists, are you sure to overwrite?", console.ConfirmOption{
			Default:     true,
			Affirmative: "Yes",
			Negative:    "No",
		}).Return(true, nil).Once()
		s.Require().Equal("APP_KEY=12345\n", string(env))
	}

	mockContext.EXPECT().Option("key").Return(key).Once()
	mockContext.EXPECT().Success("Encrypted environment successfully decrypted.").Once()
	s.Nil(envDecryptCommand.Handle(mockContext))
	s.Nil(file.Remove(".env"))
	s.Nil(file.Remove(".env.encrypted"))
}
