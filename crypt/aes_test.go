package crypt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/foundation/json"
	configmock "github.com/goravel/framework/mocks/config"
	foundationmock "github.com/goravel/framework/mocks/foundation"
	"github.com/goravel/framework/support"
)

type AesTestSuite struct {
	suite.Suite
	aes *AES
}

func TestAesTestSuite(t *testing.T) {
	mockConfig := &configmock.Config{}
	mockConfig.On("GetString", "app.key").Return("11111111111111111111111111111111").Once()
	aes, err := NewAES(mockConfig, json.New())

	assert.NoError(t, err)

	suite.Run(t, &AesTestSuite{
		aes: aes,
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

func Benchmark_EncryptString(b *testing.B) {
	mockConfig := &configmock.Config{}
	mockConfig.On("GetString", "app.key").Return("11111111111111111111111111111111").Once()
	aes, err := NewAES(mockConfig, json.New())
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := aes.EncryptString("Goravel")
		if err != nil {
			b.Error(err)
		}
	}
}

func Benchmark_DecryptString(b *testing.B) {
	mockConfig := &configmock.Config{}
	mockConfig.On("GetString", "app.key").Return("11111111111111111111111111111111").Once()
	aes, err := NewAES(mockConfig, json.New())
	if err != nil {
		b.Fatal(err)
	}

	payload, err := aes.EncryptString("Goravel")
	if err != nil {
		b.Error(err)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := aes.DecryptString(payload)
		if err != nil {
			b.Error(err)
		}
	}
}

func TestNewAES(t *testing.T) {
	t.Run("valid key lengths", func(t *testing.T) {
		cases := []struct {
			name string
			key  string
		}{
			{name: "aes-128", key: "1111111111111111"},
			{name: "aes-192", key: "111111111111111111111111"},
			{name: "aes-256", key: "11111111111111111111111111111111"},
		}

		for _, testCase := range cases {
			mockConfig := &configmock.Config{}
			mockConfig.On("GetString", "app.key").Return(testCase.key).Once()
			aes, err := NewAES(mockConfig, json.New())
			assert.NoError(t, err, testCase.name)
			assert.NotNil(t, aes, testCase.name)
			mockConfig.AssertExpectations(t)
		}
	})

	t.Run("empty key in artisan mode", func(t *testing.T) {
		runtimeMode := support.RuntimeMode
		support.RuntimeMode = support.RuntimeArtisan
		t.Cleanup(func() {
			support.RuntimeMode = runtimeMode
		})

		mockConfig := &configmock.Config{}
		mockConfig.On("GetString", "app.key").Return("").Once()
		aes, err := NewAES(mockConfig, json.New())
		assert.Nil(t, aes)
		assert.Equal(t, errors.CryptAppKeyNotSet, err)
		mockConfig.AssertExpectations(t)
	})

	t.Run("invalid key length", func(t *testing.T) {
		mockConfig := &configmock.Config{}
		mockConfig.On("GetString", "app.key").Return("invalid").Once()
		aes, err := NewAES(mockConfig, json.New())
		assert.Nil(t, aes)
		assert.True(t, errors.Is(err, errors.CryptInvalidAppKeyLength))
		mockConfig.AssertExpectations(t)
	})
}

func TestEncryptString(t *testing.T) {
	t.Run("invalid key", func(t *testing.T) {
		aes := &AES{
			json: json.New(),
			key:  []byte("invalid"),
		}
		_, err := aes.EncryptString("Goravel")
		assert.Error(t, err)
	})

	t.Run("json marshal error", func(t *testing.T) {
		jsonMock := foundationmock.NewJson(t)
		jsonMock.EXPECT().Marshal(mock.Anything).Return(nil, assert.AnError).Once()

		aes := &AES{
			json: jsonMock,
			key:  []byte("11111111111111111111111111111111"),
		}

		_, err := aes.EncryptString("Goravel")
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestDecryptString(t *testing.T) {
	t.Run("invalid key", func(t *testing.T) {
		validAES := &AES{
			json: json.New(),
			key:  []byte("11111111111111111111111111111111"),
		}
		payload, err := validAES.EncryptString("Goravel")
		assert.NoError(t, err)

		invalidAES := &AES{
			json: json.New(),
			key:  []byte("invalid"),
		}
		_, err = invalidAES.DecryptString(payload)
		assert.Error(t, err)
	})

	t.Run("json unmarshal error", func(t *testing.T) {
		jsonMock := foundationmock.NewJson(t)
		jsonMock.EXPECT().Unmarshal(mock.Anything, mock.Anything).Return(assert.AnError).Once()

		aes := &AES{
			json: jsonMock,
			key:  []byte("11111111111111111111111111111111"),
		}

		_, err := aes.DecryptString("e30=")
		assert.ErrorIs(t, err, assert.AnError)
	})
}
