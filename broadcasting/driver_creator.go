package broadcasting

import (
	"github.com/goravel/framework/broadcasting/broadcasters"
	"github.com/goravel/framework/contracts/broadcasting"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
)

func CreateDriver(conn broadcasting.ConnectionConfig, app foundation.Application) (broadcasting.Driver, error) {
	switch conn.Driver {
	case "pusher":
		httpClient := app.MakeHttp()
		if httpClient == nil {
			return nil, errors.HttpFacadeNotSet.SetModule(errors.ModuleBroadcast)
		}
		return broadcasters.NewPusherDriver(conn, httpClient)
	case "log":
		logFacade := app.MakeLog()
		if logFacade == nil {
			return nil, errors.LogFacadeNotSet.SetModule(errors.ModuleBroadcast)
		}
		return broadcasters.NewLogDriver(logFacade), nil
	case "null":
		return broadcasters.NewNullDriver(), nil
	default:
		return nil, errors.BroadcastDriverNotSupported.Args(conn.Driver)
	}
}
