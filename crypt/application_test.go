package crypt

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/testing/mock"
)

type ApplicationTestSuite struct {
	suite.Suite
}

func TestApplicationTestSuite(t *testing.T) {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "app.key").Return("11111111111111111111111111111111").Once()

	facades.Crypt = NewApplication()
	suite.Run(t, new(ApplicationTestSuite))
	mockConfig.AssertExpectations(t)
}

func (s *ApplicationTestSuite) SetupTest() {

}

func (s *ApplicationTestSuite) TestEncryptString() {
	encryptString, err := facades.Crypt.EncryptString("Goravel")
	s.NoError(err)
	s.NotEmpty(encryptString)
}

func (s *ApplicationTestSuite) TestDecryptString() {
	payload, err := facades.Crypt.EncryptString("Goravel")
	s.NoError(err)
	s.NotEmpty(payload)

	value, err := facades.Crypt.DecryptString(payload)
	s.NoError(err)
	s.Equal("Goravel", value)

	_, err = facades.Crypt.DecryptString("Goravel")
	s.Error(err)

	_, err = facades.Crypt.DecryptString("R29yYXZlbA==")
	s.Error(err)

	_, err = facades.Crypt.DecryptString("eyJpIjoiMTIzNDUiLCJ2YWx1ZSI6IjEyMzQ1In0=")
	s.Error(err)

	_, err = facades.Crypt.DecryptString("eyJpdiI6IjEyMzQ1IiwidiI6IjEyMzQ1In0=")
	s.Error(err)

	_, err = facades.Crypt.DecryptString("eyJpdiI6IjEyMzQ1IiwidmFsdWUiOiIxMjM0NSJ9")
	s.Error(err)
}
