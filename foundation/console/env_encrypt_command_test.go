package console

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/console/command"
	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/file"
)

const EnvFileEncryptInvalidKey = "xxxx"
const EnvFileEncryptValidKey = "BgcELROHL8sAV568T7Fiki7krjLHOkUc"
const EnvFileEncryptPlaintext = "APP_KEY=12345"
const EnvFileEncryptCiphertext = "QmdjRUxST0hMOHNBVjU2OKtnzDsyCUjWjNdNa2OVn5w="

type EnvEncryptCommandTestSuite struct {
	suite.Suite
}

func TestEnvEncryptCommandTestSuite(t *testing.T) {
	suite.Run(t, new(EnvEncryptCommandTestSuite))
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
				got      any
				expected any
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
	cmd := NewEnvEncryptCommand()
	mockContext := mocksconsole.NewContext(s.T())

	s.Run(fmt.Sprintf("%s not exists", support.EnvFilePath), func() {
		mockContext.EXPECT().Option("key").Return(EnvFileEncryptValidKey).Once()
		mockContext.EXPECT().Option("name").Return(support.EnvFileEncryptPath).Once()
		mockContext.EXPECT().Error("Environment file not found.").Once()

		s.Nil(cmd.Handle(mockContext))
	})

	s.Run(fmt.Sprintf("%s exists and confirm failed", support.EnvFileEncryptPath), func() {
		s.Nil(file.PutContent(support.EnvFilePath, EnvFileEncryptPlaintext))
		s.Nil(file.PutContent(support.EnvFileEncryptPath, EnvFileEncryptCiphertext))
		defer func() {
			s.Nil(file.Remove(support.EnvFilePath))
			s.Nil(file.Remove(support.EnvFileEncryptPath))
		}()

		mockContext.EXPECT().Option("key").Return(EnvFileEncryptValidKey).Once()
		mockContext.EXPECT().Option("name").Return(support.EnvFileEncryptPath).Once()
		mockContext.EXPECT().Confirm("Encrypted environment file already exists, are you sure to overwrite?").Return(false).Once()

		s.Nil(cmd.Handle(mockContext))
	})

	s.Run("invalid key", func() {
		s.Nil(file.PutContent(support.EnvFilePath, EnvFileEncryptPlaintext))
		defer func() {
			s.Nil(file.Remove(support.EnvFilePath))
		}()

		mockContext.EXPECT().Option("key").Return(EnvFileEncryptInvalidKey).Once()
		mockContext.EXPECT().Option("name").Return(support.EnvFileEncryptPath).Once()
		mockContext.EXPECT().Error("Encrypt error: crypto/aes: invalid key size 4").Once()

		s.Nil(cmd.Handle(mockContext))
	})

	s.Run(fmt.Sprintf("success when %s exists", support.EnvFileEncryptPath), func() {
		s.Nil(file.PutContent(support.EnvFilePath, EnvFileEncryptPlaintext))
		s.Nil(file.PutContent(support.EnvFileEncryptPath, EnvFileEncryptCiphertext))
		defer func() {
			s.Nil(file.Remove(support.EnvFilePath))
			s.Nil(file.Remove(support.EnvFileEncryptPath))
		}()

		mockContext.EXPECT().Option("key").Return(EnvFileEncryptValidKey).Once()
		mockContext.EXPECT().Option("name").Return(support.EnvFileEncryptPath).Once()
		mockContext.EXPECT().Confirm("Encrypted environment file already exists, are you sure to overwrite?").Return(true).Once()
		mockContext.EXPECT().Success("Environment successfully encrypted.").Once()
		mockContext.EXPECT().TwoColumnDetail("Key", EnvFileEncryptValidKey).Once()
		mockContext.EXPECT().TwoColumnDetail("Cipher", support.EnvFileEncryptCipher).Once()
		mockContext.EXPECT().TwoColumnDetail("Encrypted file", ".env.encrypted").Once()

		s.Nil(cmd.Handle(mockContext))
		content, err := file.GetContent(support.EnvFileEncryptPath)
		s.Nil(err)
		s.Equal(EnvFileEncryptCiphertext, content)
	})

	s.Run(fmt.Sprintf("success when %s not exists", support.EnvFileEncryptPath), func() {
		s.Nil(file.PutContent(support.EnvFilePath, EnvFileEncryptPlaintext))
		defer func() {
			s.Nil(file.Remove(support.EnvFilePath))
			s.Nil(file.Remove(support.EnvFileEncryptPath))
		}()

		mockContext.EXPECT().Option("key").Return(EnvFileEncryptValidKey).Once()
		mockContext.EXPECT().Option("name").Return(support.EnvFileEncryptPath).Once()
		mockContext.EXPECT().Success("Environment successfully encrypted.").Once()
		mockContext.EXPECT().TwoColumnDetail("Key", EnvFileEncryptValidKey).Once()
		mockContext.EXPECT().TwoColumnDetail("Cipher", support.EnvFileEncryptCipher).Once()
		mockContext.EXPECT().TwoColumnDetail("Encrypted file", support.EnvFileEncryptPath).Once()

		s.Nil(cmd.Handle(mockContext))
		content, err := file.GetContent(support.EnvFileEncryptPath)
		s.Nil(err)
		s.Equal(EnvFileEncryptCiphertext, content)
	})
}

func (s *EnvEncryptCommandTestSuite) TestEncrypt() {
	s.Run("valid key", func() {
		ciphertext, err := NewEnvEncryptCommand().encrypt([]byte(EnvFileEncryptPlaintext), []byte(EnvFileEncryptValidKey))
		base64Data := base64.StdEncoding.EncodeToString(ciphertext)
		s.Equal(EnvFileEncryptCiphertext, base64Data)
		s.Nil(err)
	})

	s.Run("invalid key", func() {
		_, err := NewEnvEncryptCommand().encrypt([]byte(EnvFileEncryptPlaintext), []byte(EnvFileEncryptInvalidKey))
		s.Error(err)
	})
}
