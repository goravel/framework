package crypt

import (
	"testing"

	"github.com/stretchr/testify/suite"

	configmock "github.com/goravel/framework/contracts/config/mocks"
)

type AesTestSuite struct {
	suite.Suite
	aes *AES
}

func TestAesTestSuite(t *testing.T) {
	mockConfig := &configmock.Config{}
	mockConfig.On("GetString", "app.key").Return("11111111111111111111111111111111").Once()
	suite.Run(t, &AesTestSuite{
		aes: NewAES(mockConfig),
	})
	mockConfig.AssertExpectations(t)
}

func (s *AesTestSuite) SetupTest() {

}

func (s *AesTestSuite) TestEncryptString() {
	encryptString, err := s.aes.EncryptString("Goravel")
	s.NoError(err)
	s.NotEmpty(encryptString)
}

func (s *AesTestSuite) TestDecryptString() {
	payload, err := s.aes.EncryptString("Goravel")
	s.NoError(err)
	s.NotEmpty(payload)

	value, err := s.aes.DecryptString(payload)
	s.NoError(err)
	s.Equal("Goravel", value)

	_, err = s.aes.DecryptString("Goravel")
	s.Error(err)

	_, err = s.aes.DecryptString("R29yYXZlbA==")
	s.Error(err)

	_, err = s.aes.DecryptString("eyJpIjoiMTIzNDUiLCJ2YWx1ZSI6IjEyMzQ1In0=")
	s.Error(err)

	_, err = s.aes.DecryptString("eyJpdiI6IjEyMzQ1IiwidiI6IjEyMzQ1In0=")
	s.Error(err)

	_, err = s.aes.DecryptString("eyJpdiI6IjEyMzQ1IiwidmFsdWUiOiIxMjM0NSJ9")
	s.Error(err)
}
