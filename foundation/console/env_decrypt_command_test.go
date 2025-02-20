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

var EnvDecryptKey = "BgcELROHL8sAV568T7Fiki7krjLHOkUc"

type EnvDecryptCommandTestSuite struct {
	suite.Suite
}

func TestEnvDecryptCommandTestSuite(t *testing.T) {
	suite.Run(t, new(EnvDecryptCommandTestSuite))
}

func (s *EnvDecryptCommandTestSuite) SetupTest() {
	err := file.Create(".env.encrypted", "QmdjRUxST0hMOHNBVjU2OKtnzDsyCUjWjNdNa2OVn5w=")
	s.Nil(err)
}

func (s *EnvDecryptCommandTestSuite) TearDownSuite() {
	s.Nil(file.Remove(".env.encrypted"))
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

	_, err := os.ReadFile(".env.encrypted")
	if err != nil {
		mockContext.EXPECT().Error("Encrypted environment file not found.").Once()
	}

	env, err := os.ReadFile(".env")
	if err == nil {
		mockContext.EXPECT().Confirm("Environment file already exists, are you sure to overwrite?", console.ConfirmOption{
			Default:     true,
			Affirmative: "Yes",
			Negative:    "No",
		}).Return(true, nil).Once()
		s.Require().Equal("APP_KEY=12345", string(env))
	}

	s.Run("valid key", func() {
		mockContext.EXPECT().Option("key").Return(EnvDecryptKey).Once()
		mockContext.EXPECT().Success("Encrypted environment successfully decrypted.").Once()
		s.Nil(envDecryptCommand.Handle(mockContext))
	})

	s.Run("invalid key", func() {
		key := "xxxx"
		mockContext.EXPECT().Option("key").Return(key).Once()
		mockContext.EXPECT().Confirm("Environment file already exists, are you sure to overwrite?", console.ConfirmOption{
			Default:     true,
			Affirmative: "Yes",
			Negative:    "No",
		}).Return(true, nil).Once()
		mockContext.EXPECT().Error("Decrypt error: crypto/aes: invalid key size 4").Once()
		s.Nil(envDecryptCommand.Handle(mockContext))
	})
}

func (s *EnvDecryptCommandTestSuite) TestDecrypt() {
	envDecryptCommand := NewEnvDecryptCommand()
	s.Run("valid key", func() {
		ciphertext := "QmdjRUxST0hMOHNBVjU2OKtnzDsyCUjWjNdNa2OVn5w="
		decrypted, err := envDecryptCommand.decrypt([]byte(ciphertext), []byte(EnvDecryptKey))
		s.Nil(err)
		s.Equal("APP_KEY=12345", string(decrypted))
		s.Nil(err)
	})
	s.Run("invalid key", func() {
		ciphertext := "QmdjRUxST0hMOHNBVjU2OKtnzDsyCUjWjNdNa2OVn5w="
		key := "xxxx"
		_, err := envDecryptCommand.decrypt([]byte(ciphertext), []byte(key))
		s.Error(err)
	})
}
