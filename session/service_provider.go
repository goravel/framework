package session

import (
	"time"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/session"
	"github.com/goravel/framework/support/color"
)

var (
	SessionFacade session.Manager
	ConfigFacade  config.Config
)

const Binding = "goravel.session"

type ServiceProvider struct {
}

func (receiver *ServiceProvider) Register(app foundation.Application) {
	app.Singleton(Binding, func(app foundation.Application) (any, error) {
		c := app.MakeConfig()
		j := app.GetJson()
		return NewManager(c, j), nil
	})
}

func (receiver *ServiceProvider) Boot(app foundation.Application) {
	SessionFacade = app.MakeSession()
	ConfigFacade = app.MakeConfig()

	driver, err := SessionFacade.Driver()
	if err != nil {
		color.Red().Println(err)
		return
	}

	startGcTimer(driver)
}

// startGcTimer starts a garbage collection timer for the session driver.
func startGcTimer(driver session.Driver) {
	interval := ConfigFacade.GetInt("session.gc_interval", 30)
	if interval <= 0 {
		// No need to start the timer if the interval is zero or negative
		return
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Minute)

	go func() {
		for range ticker.C {
			lifetime := ConfigFacade.GetInt("session.lifetime") * 60
			if err := driver.Gc(lifetime); err != nil {
				color.Red().Printf("Error performing garbage collection: %s\n", err)
			}
		}
	}()
}
