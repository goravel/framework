package template

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	mailcontract "github.com/goravel/framework/contracts/mail"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksmail "github.com/goravel/framework/mocks/mail"
)

func TestGet_DefaultEngine(t *testing.T) {
	engines = sync.Map{}

	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("mail.template.driver", "default").Return("default")
	mockConfig.EXPECT().GetString("mail.template.views_path", "resources/views/mail").Return("/test/views")

	engine, err := Get(mockConfig)
	assert.NoError(t, err)
	assert.NotNil(t, engine)

	defaultEngine, ok := engine.(*DefaultEngine)
	assert.True(t, ok)
	assert.Equal(t, "/test/views", defaultEngine.viewsPath)
}
func TestGet_CachedEngine(t *testing.T) {
	engines = sync.Map{}

	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("mail.template.driver", "default").Return("default").Times(2)
	mockConfig.EXPECT().GetString("mail.template.views_path", "resources/views/mail").Return("/test/views").Once()

	engine1, err := Get(mockConfig)
	assert.NoError(t, err)
	assert.NotNil(t, engine1)

	engine2, err := Get(mockConfig)
	assert.NoError(t, err)
	assert.NotNil(t, engine2)

	assert.Equal(t, engine1, engine2)
}
func TestGet_CustomEngineInstance(t *testing.T) {
	engines = sync.Map{}

	mockTemplate := mocksmail.NewTemplate(t)
	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("mail.template.driver", "default").Return("custom")
	mockConfig.EXPECT().Get("mail.template.drivers.custom.engine").Return(mockTemplate)

	engine, err := Get(mockConfig)
	assert.NoError(t, err)
	assert.Equal(t, mockTemplate, engine)
}

func TestGet_CustomEngineFactory(t *testing.T) {
	engines = sync.Map{}

	mockTemplate := mocksmail.NewTemplate(t)
	factory := func() (mailcontract.Template, error) {
		return mockTemplate, nil
	}

	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("mail.template.driver", "default").Return("custom")
	mockConfig.EXPECT().Get("mail.template.drivers.custom.engine").Return(factory)

	engine, err := Get(mockConfig)
	assert.NoError(t, err)
	assert.Equal(t, mockTemplate, engine)
}

func TestGet_CustomEngineFactoryError(t *testing.T) {
	engines = sync.Map{}

	factory := func() (mailcontract.Template, error) {
		return nil, fmt.Errorf("factory error")
	}

	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("mail.template.driver", "default").Return("custom")
	mockConfig.EXPECT().Get("mail.template.drivers.custom.engine").Return(factory)

	engine, err := Get(mockConfig)
	assert.Error(t, err)
	assert.Nil(t, engine)
	assert.Contains(t, err.Error(), "factory for template engine 'custom' failed")
}

func TestGet_UnsupportedEngine(t *testing.T) {
	engines = sync.Map{}

	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("mail.template.driver", "default").Return("unsupported")
	mockConfig.EXPECT().Get("mail.template.drivers.unsupported.engine").Return(nil)

	engine, err := Get(mockConfig)
	assert.Error(t, err)
	assert.Nil(t, engine)
	assert.Contains(t, err.Error(), "unsupported or misconfigured template engine: unsupported")
}
