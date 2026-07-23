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
		return broadcasters.NewPusherDriver(conn, app.MakeHttp())
	case "log":
		return broadcasters.NewLogDriver(app.MakeLog()), nil
	case "null":
		return broadcasters.NewNullDriver(), nil
	default:
		return nil, errors.BroadcastDriverNotSupported.Args(conn.Driver)
	}
}
