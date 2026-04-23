package console

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/console/command"
	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/file"
)

const EnvFileDecryptInvalidKey = "xxxx"
const EnvFileDecryptValidKey = "BgcELROHL8sAV568T7Fiki7krjLHOkUc"
const EnvFileDecryptPlaintext = "APP_KEY=12345"
const EnvFileDecryptCiphertext = "QmdjRUxST0hMOHNBVjU2OKtnzDsyCUjWjNdNa2OVn5w="

type EnvDecryptCommandTestSuite struct {
	suite.Suite
}

func TestEnvDecryptCommandTestSuite(t *testing.T) {
	suite.Run(t, new(EnvDecryptCommandTestSuite))
}

func (s *EnvDecryptCommandTestSuite) SetupSuite() {
}

func (s *EnvDecryptCommandTestSuite) TearDownSuite() {
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
				got      any
				expected any
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
	cmd := NewEnvDecryptCommand()
	mockContext := mocksconsole.NewContext(s.T())

	s.Run("empty key", func() {
		mockContext.EXPECT().Option("key").Return("").Once()
		mockContext.EXPECT().Option("name").Return(support.EnvFileEncryptPath).Once()
		mockContext.EXPECT().Error("A decryption key is required.").Once()
		s.Nil(cmd.Handle(mockContext))
	})

	s.Run("invalid key", func() {
		s.Nil(file.PutContent(support.EnvFileEncryptPath, EnvFileDecryptCiphertext))
		defer func() {
			s.Nil(file.Remove(support.EnvFileEncryptPath))
		}()

		mockContext.EXPECT().Option("key").Return(EnvFileDecryptInvalidKey).Once()
		mockContext.EXPECT().Option("name").Return(support.EnvFileEncryptPath).Once()
		mockContext.EXPECT().Error("Decrypt error: crypto/aes: invalid key size 4").Once()

		s.Nil(cmd.Handle(mockContext))
	})

	s.Run(fmt.Sprintf("%s is not found", support.EnvFileEncryptPath), func() {
		mockContext.EXPECT().Option("key").Return(EnvFileDecryptValidKey).Once()
		mockContext.EXPECT().Option("name").Return(support.EnvFileEncryptPath).Once()
		mockContext.EXPECT().Error("Encrypted environment file not found.").Once()
		s.Nil(cmd.Handle(mockContext))
	})

	s.Run(fmt.Sprintf("%s exists and confirm failed", support.EnvFilePath), func() {
		s.Nil(file.PutContent(support.EnvFileEncryptPath, EnvFileDecryptCiphertext))
		s.Nil(file.PutContent(support.EnvFilePath, EnvFileDecryptPlaintext))
		defer func() {
			s.Nil(file.Remove(support.EnvFileEncryptPath))
			s.Nil(file.Remove(support.EnvFilePath))
		}()

		mockContext.EXPECT().Option("key").Return(EnvFileDecryptValidKey).Once()
		mockContext.EXPECT().Option("name").Return(support.EnvFileEncryptPath).Once()
		mockContext.EXPECT().Confirm("Environment file already exists, are you sure to overwrite?").Return(false).Once()
		s.Nil(cmd.Handle(mockContext))
	})

	s.Run(fmt.Sprintf("success when %s exists", support.EnvFilePath), func() {
		s.Nil(file.PutContent(support.EnvFileEncryptPath, EnvFileDecryptCiphertext))
		s.Nil(file.PutContent(support.EnvFilePath, EnvFileDecryptPlaintext))
		defer func() {
			s.Nil(file.Remove(support.EnvFilePath))
			s.Nil(file.Remove(support.EnvFileEncryptPath))
		}()

		mockContext.EXPECT().Option("key").Return(EnvFileDecryptValidKey).Once()
		mockContext.EXPECT().Option("name").Return(support.EnvFileEncryptPath).Once()
		mockContext.EXPECT().Confirm("Environment file already exists, are you sure to overwrite?").Return(true).Once()
		mockContext.EXPECT().Success("Encrypted environment successfully decrypted.").Once()

		s.Nil(cmd.Handle(mockContext))
		s.True(file.Exists(support.EnvFilePath))
		content, err := file.GetContent(support.EnvFilePath)
		s.Nil(err)
		s.Equal(EnvFileDecryptPlaintext, content)
	})

	s.Run(fmt.Sprintf("success when %s not exists", support.EnvFilePath), func() {
		s.Nil(file.PutContent(support.EnvFileEncryptPath, EnvFileDecryptCiphertext))
		defer func() {
			s.Nil(file.Remove(support.EnvFileEncryptPath))
			s.Nil(file.Remove(support.EnvFilePath))
		}()

		mockContext.EXPECT().Option("key").Return(EnvFileDecryptValidKey).Once()
		mockContext.EXPECT().Option("name").Return(support.EnvFileEncryptPath).Once()
		mockContext.EXPECT().Success("Encrypted environment successfully decrypted.").Once()

		s.Nil(cmd.Handle(mockContext))
		s.True(file.Exists(support.EnvFilePath))
		content, err := file.GetContent(support.EnvFilePath)
		s.Nil(err)
		s.Equal(EnvFileDecryptPlaintext, content)
	})
}

func (s *EnvDecryptCommandTestSuite) TestDecrypt() {
	s.Run("valid key", func() {
		decrypted, err := NewEnvDecryptCommand().decrypt([]byte(EnvFileDecryptCiphertext), []byte(EnvFileDecryptValidKey))
		s.Nil(err)
		s.Equal(EnvFileDecryptPlaintext, string(decrypted))
		s.Nil(err)
	})

	s.Run("invalid key", func() {
		_, err := NewEnvDecryptCommand().decrypt([]byte(EnvFileDecryptCiphertext), []byte(EnvFileDecryptInvalidKey))
		s.Error(err)
	})
}
