package route

import (
	"testing"

	"github.com/stretchr/testify/assert"

	contractsroute "github.com/goravel/framework/contracts/route"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksroute "github.com/goravel/framework/mocks/route"
)

func TestNewDriver(t *testing.T) {
	t.Run("route instance", func(t *testing.T) {
		mockConfig := mocksconfig.NewConfig(t)
		mockRoute := mocksroute.NewRoute(t)
		mockConfig.EXPECT().Get("http.drivers.gin.route").Return(mockRoute).Once()

		driver, err := NewDriver(mockConfig, "gin")

		assert.NoError(t, err)
		assert.Equal(t, mockRoute, driver)
	})

	t.Run("route callback", func(t *testing.T) {
		mockConfig := mocksconfig.NewConfig(t)
		mockRoute := mocksroute.NewRoute(t)
		mockConfig.EXPECT().Get("http.drivers.gin.route").Return(func() (contractsroute.Route, error) {
			return mockRoute, nil
		}).Twice()

		driver, err := NewDriver(mockConfig, "gin")

		assert.NoError(t, err)
		assert.Equal(t, mockRoute, driver)
	})

	t.Run("callback returns error", func(t *testing.T) {
		mockConfig := mocksconfig.NewConfig(t)
		mockConfig.EXPECT().Get("http.drivers.gin.route").Return(func() (contractsroute.Route, error) {
			return nil, assert.AnError
		}).Twice()

		driver, err := NewDriver(mockConfig, "gin")

		assert.ErrorIs(t, err, assert.AnError)
		assert.Nil(t, driver)
	})

	t.Run("invalid driver", func(t *testing.T) {
		mockConfig := mocksconfig.NewConfig(t)
		mockConfig.EXPECT().Get("http.drivers.gin.route").Return(nil).Twice()

		driver, err := NewDriver(mockConfig, "gin")

		assert.Error(t, err)
		assert.Nil(t, driver)
		assert.Contains(t, err.Error(), "init gin route driver fail")
	})
}

func TestNewRoute(t *testing.T) {
	t.Run("default driver empty", func(t *testing.T) {
		mockConfig := mocksconfig.NewConfig(t)
		mockConfig.EXPECT().GetString("http.default").Return("").Once()

		router, err := NewRoute(mockConfig)

		assert.NoError(t, err)
		assert.NotNil(t, router)
		assert.Nil(t, router.Route)
		assert.Equal(t, mockConfig, router.config)
	})

	t.Run("default driver set", func(t *testing.T) {
		mockConfig := mocksconfig.NewConfig(t)
		mockRoute := mocksroute.NewRoute(t)
		mockConfig.EXPECT().GetString("http.default").Return("gin").Once()
		mockConfig.EXPECT().Get("http.drivers.gin.route").Return(mockRoute).Once()

		router, err := NewRoute(mockConfig)

		assert.NoError(t, err)
		assert.NotNil(t, router)
		assert.Equal(t, mockRoute, router.Route)
		assert.Equal(t, mockConfig, router.config)
	})

	t.Run("driver init fails", func(t *testing.T) {
		mockConfig := mocksconfig.NewConfig(t)
		mockConfig.EXPECT().GetString("http.default").Return("gin").Once()
		mockConfig.EXPECT().Get("http.drivers.gin.route").Return(nil).Twice()

		router, err := NewRoute(mockConfig)

		assert.Nil(t, router)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "init gin route driver fail")
	})
}

// compilerDriver is a route driver that implements view.Compiler.
type compilerDriver struct {
	contractsroute.Route
	err    error
	called int
}

func (d *compilerDriver) CompileViews() error {
	d.called++

	return d.err
}

func TestCompileViews(t *testing.T) {
	t.Run("no driver", func(t *testing.T) {
		route := &Route{}
		assert.NoError(t, route.CompileViews())
	})

	t.Run("driver without view.Compiler", func(t *testing.T) {
		route := &Route{Route: mocksroute.NewRoute(t)}
		assert.NoError(t, route.CompileViews())
	})

	t.Run("driver with view.Compiler is forwarded to once per call", func(t *testing.T) {
		driver := &compilerDriver{Route: mocksroute.NewRoute(t)}
		route := &Route{Route: driver}
		assert.NoError(t, route.CompileViews())
		assert.NoError(t, route.CompileViews())
		assert.Equal(t, 2, driver.called)
	})

	t.Run("driver error is returned", func(t *testing.T) {
		driver := &compilerDriver{Route: mocksroute.NewRoute(t), err: assert.AnError}
		route := &Route{Route: driver}
		assert.ErrorIs(t, route.CompileViews(), assert.AnError)
	})
}
