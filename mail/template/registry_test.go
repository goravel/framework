package template

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	contractsmail "github.com/goravel/framework/contracts/mail"
	"github.com/goravel/framework/errors"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksmail "github.com/goravel/framework/mocks/mail"
)

func TestGet_HtmlEngine(t *testing.T) {
	engines = sync.Map{}

	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("mail.template.default", "html").Return("html").Once()
	mockConfig.EXPECT().GetString("mail.template.engines.html.driver", "html").Return("html").Once()
	mockConfig.EXPECT().GetString("mail.template.engines.html.path", "resources/views/mail").Return("/test/views").Once()

	engine, err := Get(mockConfig)
	assert.NoError(t, err)
	assert.NotNil(t, engine)

	htmlEngine, ok := engine.(*Html)
	assert.True(t, ok)
	assert.Equal(t, "/test/views", htmlEngine.viewsPath)
}
func TestGet_CachedEngine(t *testing.T) {
	engines = sync.Map{}

	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("mail.template.default", "html").Return("html").Times(2)
	mockConfig.EXPECT().GetString("mail.template.engines.html.driver", "html").Return("html").Once()
	mockConfig.EXPECT().GetString("mail.template.engines.html.path", "resources/views/mail").Return("/test/views").Once()

	engine1, err := Get(mockConfig)
	assert.NoError(t, err)
	assert.NotNil(t, engine1)

	engine2, err := Get(mockConfig)
	assert.NoError(t, err)
	assert.NotNil(t, engine2)

	assert.Equal(t, engine1, engine2)
}
func TestGet_CustomEngineViaInstance(t *testing.T) {
	engines = sync.Map{}

	mockTemplate := mocksmail.NewTemplate(t)
	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("mail.template.default", "html").Return("custom").Once()
	mockConfig.EXPECT().GetString("mail.template.engines.custom.driver", "html").Return("custom").Once()
	mockConfig.EXPECT().Get("mail.template.engines.custom.via", "").Return(mockTemplate).Once()

	engine, err := Get(mockConfig)
	assert.NoError(t, err)
	assert.Equal(t, mockTemplate, engine)
}

func TestGet_CustomEngineViaFactory(t *testing.T) {
	engines = sync.Map{}

	mockTemplate := mocksmail.NewTemplate(t)
	factory := func() (contractsmail.Template, error) {
		return mockTemplate, nil
	}

	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("mail.template.default", "html").Return("custom").Once()
	mockConfig.EXPECT().GetString("mail.template.engines.custom.driver", "html").Return("custom").Once()
	mockConfig.EXPECT().Get("mail.template.engines.custom.via", "").Return(factory).Once()

	engine, err := Get(mockConfig)
	assert.NoError(t, err)
	assert.Equal(t, mockTemplate, engine)
}

func TestGet_CustomEngineFactoryError(t *testing.T) {
	engines = sync.Map{}

	factory := func() (contractsmail.Template, error) {
		return nil, fmt.Errorf("factory error")
	}

	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("mail.template.default", "html").Return("custom").Once()
	mockConfig.EXPECT().GetString("mail.template.engines.custom.driver", "html").Return("custom").Once()
	mockConfig.EXPECT().Get("mail.template.engines.custom.via", "").Return(factory).Once()

	engine, err := Get(mockConfig)
	assert.ErrorIs(t, err, errors.MailTemplateEngineFactoryFailed)
	assert.Nil(t, engine)
}

func TestGet_CustomEngineViaRequired(t *testing.T) {
	engines = sync.Map{}

	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("mail.template.default", "html").Return("custom").Once()
	mockConfig.EXPECT().GetString("mail.template.engines.custom.driver", "html").Return("custom").Once()
	mockConfig.EXPECT().Get("mail.template.engines.custom.via", "").Return("").Once()

	engine, err := Get(mockConfig)
	assert.ErrorIs(t, err, errors.MailTemplateEngineViaRequired)
	assert.Nil(t, engine)
}

func TestGet_CustomEngineViaInvalid(t *testing.T) {
	engines = sync.Map{}

	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("mail.template.default", "html").Return("custom").Once()
	mockConfig.EXPECT().GetString("mail.template.engines.custom.driver", "html").Return("custom").Once()
	mockConfig.EXPECT().Get("mail.template.engines.custom.via", "").Return("invalid string").Once()

	engine, err := Get(mockConfig)
	assert.ErrorIs(t, err, errors.MailTemplateEngineViaInvalid)
	assert.Nil(t, engine)
}

func TestGet_UnsupportedEngine(t *testing.T) {
	engines = sync.Map{}

	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("mail.template.default", "html").Return("unsupported").Once()
	mockConfig.EXPECT().GetString("mail.template.engines.unsupported.driver", "html").Return("unknown").Once()

	engine, err := Get(mockConfig)
	assert.ErrorIs(t, err, errors.MailTemplateEngineNotSupported)
	assert.Nil(t, engine)
}
