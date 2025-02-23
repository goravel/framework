package console

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/file"
)

const EnvDecryptInvalidKey = "xxxx"
const EnvDecryptValidKey = "BgcELROHL8sAV568T7Fiki7krjLHOkUc"
const EnvDecryptPlaintext = "APP_KEY=12345"
const EnvDecryptCiphertext = "QmdjRUxST0hMOHNBVjU2OKtnzDsyCUjWjNdNa2OVn5w="

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
	cmd := NewEnvDecryptCommand()
	ctx := mocksconsole.NewContext(s.T())

	s.Run("empty key", func() {
		ctx.EXPECT().Option("key").Return("").Once()
		ctx.EXPECT().Option("name").Return(support.EnvEncryptPath).Once()
		ctx.EXPECT().Error("A decryption key is required.").Once()
		s.Nil(cmd.Handle(ctx))
	})

	s.Run("invalid key", func() {
		s.Nil(file.PutContent(support.EnvEncryptPath, EnvDecryptCiphertext))
		defer func() {
			s.Nil(file.Remove(support.EnvEncryptPath))
		}()

		ctx.EXPECT().Option("key").Return(EnvDecryptInvalidKey).Once()
		ctx.EXPECT().Option("name").Return(support.EnvEncryptPath).Once()
		ctx.EXPECT().Error("Decrypt error: crypto/aes: invalid key size 4").Once()

		s.Nil(cmd.Handle(ctx))
	})

	s.Run(".env.encrypted is not found", func() {
		ctx.EXPECT().Option("key").Return(EnvDecryptValidKey).Once()
		ctx.EXPECT().Option("name").Return(support.EnvEncryptPath).Once()
		ctx.EXPECT().Error("Encrypted environment file not found.").Once()
		s.Nil(cmd.Handle(ctx))
	})

	s.Run(".env exists and confirm failed", func() {
		s.Nil(file.PutContent(support.EnvEncryptPath, EnvDecryptCiphertext))
		s.Nil(file.PutContent(support.EnvPath, EnvDecryptPlaintext))
		defer func() {
			s.Nil(file.Remove(support.EnvEncryptPath))
			s.Nil(file.Remove(support.EnvPath))
		}()

		ctx.EXPECT().Option("key").Return(EnvDecryptValidKey).Once()
		ctx.EXPECT().Option("name").Return(support.EnvEncryptPath).Once()
		ctx.EXPECT().Confirm("Environment file already exists, are you sure to overwrite?", console.ConfirmOption{
			Default:     false,
			Affirmative: "Yes",
			Negative:    "No",
		}).Return(false, nil).Once()
		s.Nil(cmd.Handle(ctx))
	})

	s.Run("success when .env exists", func() {
		s.Nil(file.PutContent(support.EnvEncryptPath, EnvDecryptCiphertext))
		s.Nil(file.PutContent(support.EnvPath, EnvDecryptPlaintext))
		defer func() {
			s.Nil(file.Remove(support.EnvPath))
			s.Nil(file.Remove(support.EnvEncryptPath))
		}()

		ctx.EXPECT().Option("key").Return(EnvDecryptValidKey).Once()
		ctx.EXPECT().Option("name").Return(support.EnvEncryptPath).Once()
		ctx.EXPECT().Confirm("Environment file already exists, are you sure to overwrite?", console.ConfirmOption{
			Default:     false,
			Affirmative: "Yes",
			Negative:    "No",
		}).Return(true, nil).Once()
		ctx.EXPECT().Success("Encrypted environment successfully decrypted.").Once()

		s.Nil(cmd.Handle(ctx))
		s.True(file.Exists(support.EnvPath))
		content, err := file.GetContent(support.EnvPath)
		s.Nil(err)
		s.Equal(EnvDecryptPlaintext, content)
	})

	s.Run("success when .env not exists", func() {
		s.Nil(file.PutContent(support.EnvEncryptPath, EnvDecryptCiphertext))
		defer func() {
			s.Nil(file.Remove(support.EnvEncryptPath))
			s.Nil(file.Remove(support.EnvPath))
		}()

		ctx.EXPECT().Option("key").Return(EnvDecryptValidKey).Once()
		ctx.EXPECT().Option("name").Return(support.EnvEncryptPath).Once()
		ctx.EXPECT().Success("Encrypted environment successfully decrypted.").Once()

		s.Nil(cmd.Handle(ctx))
		s.True(file.Exists(support.EnvPath))
		content, err := file.GetContent(support.EnvPath)
		s.Nil(err)
		s.Equal(EnvDecryptPlaintext, content)
	})
}

func (s *EnvDecryptCommandTestSuite) TestDecrypt() {
	s.Run("valid key", func() {
		decrypted, err := NewEnvDecryptCommand().decrypt([]byte(EnvDecryptCiphertext), []byte(EnvDecryptValidKey))
		s.Nil(err)
		s.Equal(EnvDecryptPlaintext, string(decrypted))
		s.Nil(err)
	})

	s.Run("invalid key", func() {
		_, err := NewEnvDecryptCommand().decrypt([]byte(EnvDecryptCiphertext), []byte(EnvDecryptInvalidKey))
		s.Error(err)
	})
}
