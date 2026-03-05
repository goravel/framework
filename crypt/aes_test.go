package crypt

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/foundation/json"
	configmock "github.com/goravel/framework/mocks/config"
	"github.com/goravel/framework/support"
)

const (
	testKeyAES128 = "1234567890123456"
	testKeyAES192 = "123456789012345678901234"
	testKeyAES256 = "12345678901234567890123456789012"
)

type AesTestSuite struct {
	suite.Suite
	aes *AES
}

func TestAesTestSuite(t *testing.T) {
	mockConfig := configmock.NewConfig(t)
	mockConfig.EXPECT().GetString("app.key").Return(testKeyAES256).Once()
	aes, err := NewAES(mockConfig, json.New())

	assert.NoError(t, err)

	suite.Run(t, &AesTestSuite{
		aes: aes,
	})
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
	mockConfig := configmock.NewConfig(b)
	mockConfig.EXPECT().GetString("app.key").Return(testKeyAES256).Once()
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
	mockConfig := configmock.NewConfig(b)
	mockConfig.EXPECT().GetString("app.key").Return(testKeyAES256).Once()
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

func setRuntimeMode(t *testing.T, mode string) {
	runtimeMode := support.RuntimeMode
	support.RuntimeMode = mode
	t.Cleanup(func() {
		support.RuntimeMode = runtimeMode
	})
}

func TestNewAES(t *testing.T) {
	t.Run("valid key lengths", func(t *testing.T) {
		cases := []struct {
			name string
			key  string
		}{
			{name: "aes-128", key: testKeyAES128},
			{name: "aes-192", key: testKeyAES192},
			{name: "aes-256", key: testKeyAES256},
		}

		for _, testCase := range cases {
			mockConfig := configmock.NewConfig(t)
			mockConfig.EXPECT().GetString("app.key").Return(testCase.key).Once()
			aes, err := NewAES(mockConfig, json.New())
			assert.NoError(t, err, testCase.name)
			assert.NotNil(t, aes, testCase.name)
		}
	})

	t.Run("empty key in artisan mode", func(t *testing.T) {
		setRuntimeMode(t, support.RuntimeArtisan)

		mockConfig := configmock.NewConfig(t)
		mockConfig.EXPECT().GetString("app.key").Return("").Once()
		aes, err := NewAES(mockConfig, json.New())
		assert.Nil(t, aes)
		assert.Equal(t, errors.CryptAppKeyNotSet, err)
	})

	t.Run("invalid key length", func(t *testing.T) {
		mockConfig := configmock.NewConfig(t)
		mockConfig.EXPECT().GetString("app.key").Return("invalid").Once()
		aes, err := NewAES(mockConfig, json.New())
		assert.Nil(t, aes)
		assert.True(t, errors.Is(err, errors.CryptInvalidAppKeyLength))
	})
}

func TestEncryptString(t *testing.T) {
	cases := []struct {
		name       string
		inputValue string
	}{
		{name: "normal value", inputValue: "Goravel"},
		{name: "empty value", inputValue: ""},
		{name: "unicode value", inputValue: "你好, Goravel 👋"},
	}

	for _, testCase := range cases {
		t.Run("invalid key with "+testCase.name, func(t *testing.T) {
			aes := &AES{
				json: json.New(),
				key:  []byte("invalid"),
			}
			_, err := aes.EncryptString(testCase.inputValue)
			assert.Error(t, err)
			assert.ErrorContains(t, err, "invalid key size")
		})
	}

	t.Run("encrypt/decrypt unicode round trip", func(t *testing.T) {
		aes := &AES{
			json: json.New(),
			key:  []byte(testKeyAES256),
		}

		payload, err := aes.EncryptString("你好, Goravel 👋")
		assert.NoError(t, err)

		value, err := aes.DecryptString(payload)
		assert.NoError(t, err)
		assert.Equal(t, "你好, Goravel 👋", value)
	})
}

func TestDecryptString(t *testing.T) {
	t.Run("invalid key", func(t *testing.T) {
		validAES := &AES{
			json: json.New(),
			key:  []byte(testKeyAES256),
		}
		payload, err := validAES.EncryptString("Goravel")
		assert.NoError(t, err)

		invalidAES := &AES{
			json: json.New(),
			key:  []byte("invalid"),
		}
		_, err = invalidAES.DecryptString(payload)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "invalid key size")
	})

	t.Run("json unmarshal error", func(t *testing.T) {
		aes := &AES{
			json: json.New(),
			key:  []byte(testKeyAES256),
		}

		malformedJSONPayload := base64.StdEncoding.EncodeToString([]byte("{"))
		_, err := aes.DecryptString(malformedJSONPayload)
		assert.Error(t, err)
	})
}
