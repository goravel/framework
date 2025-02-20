package console

import (
	"encoding/base64"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/file"
)

const EnvEncryptInvalidKey = "xxxx"
const EnvEncryptValidKey = "BgcELROHL8sAV568T7Fiki7krjLHOkUc"
const EnvEncryptPlaintext = "APP_KEY=12345"
const EnvEncryptCiphertext = "QmdjRUxST0hMOHNBVjU2OKtnzDsyCUjWjNdNa2OVn5w="

type EnvEncryptCommandTestSuite struct {
	suite.Suite
}

func TestEnvEncryptCommandTestSuite(t *testing.T) {
	suite.Run(t, new(EnvEncryptCommandTestSuite))
}

func (s *EnvEncryptCommandTestSuite) SetupTest() {
	s.Nil(file.Create(".env", EnvEncryptPlaintext))
}

func (s *EnvEncryptCommandTestSuite) TearDownSuite() {
	s.Nil(file.Remove(".env"))
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

	_, err := os.ReadFile(".env")
	if err != nil {
		mockContext.EXPECT().Error("Environment file not found.").Once()
	}

	_, err = os.ReadFile(".env.encrypted")
	if err == nil {
		mockContext.EXPECT().Confirm("Encrypted environment file already exists, are you sure to overwrite?", console.ConfirmOption{
			Default:     true,
			Affirmative: "Yes",
			Negative:    "No",
		}).Return(true, nil).Once()
	}

	s.Run("valid key", func() {
		mockContext.EXPECT().Option("key").Return(EnvEncryptValidKey).Once()
		mockContext.EXPECT().Success("Environment successfully encrypted.").Once()
		mockContext.EXPECT().TwoColumnDetail("Key", EnvEncryptValidKey).Once()
		mockContext.EXPECT().TwoColumnDetail("Cipher", "AES-256-CBC").Once()
		mockContext.EXPECT().TwoColumnDetail("Encrypted file", ".env.encrypted").Once()

		s.Nil(envEncryptCommand.Handle(mockContext))
	})

	s.Run("invalid key", func() {
		mockContext.EXPECT().Option("key").Return(EnvEncryptInvalidKey).Once()
		mockContext.EXPECT().Confirm("Encrypted environment file already exists, are you sure to overwrite?", console.ConfirmOption{
			Default:     true,
			Affirmative: "Yes",
			Negative:    "No",
		}).Return(true, nil).Once()
		mockContext.EXPECT().Error("Encrypt error: crypto/aes: invalid key size 4").Once()
		s.Nil(envEncryptCommand.Handle(mockContext))
	})

}

func (s *EnvDecryptCommandTestSuite) TestEncrypt() {
	envEncryptCommand := NewEnvEncryptCommand()
	s.Run("valid key", func() {
		ciphertext, err := envEncryptCommand.encrypt([]byte(EnvEncryptPlaintext), []byte(EnvEncryptValidKey))
		base64Data := base64.StdEncoding.EncodeToString(ciphertext)
		s.Equal(EnvEncryptCiphertext, base64Data)
		s.Nil(err)
	})
	s.Run("invalid key", func() {
		_, err := envEncryptCommand.encrypt([]byte(EnvEncryptPlaintext), []byte(EnvEncryptInvalidKey))
		s.Error(err)
	})

}
