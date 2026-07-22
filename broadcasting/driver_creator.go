package broadcasting

import (
	"github.com/goravel/framework/contracts/broadcasting"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
)

func CreateDriver(conn broadcasting.ConnectionConfig, app foundation.Application) (Driver, error) {
	switch conn.Driver {
	case "pusher":
		return NewPusherDriver(conn), nil
	case "log":
		return NewLogDriver(app.MakeLog()), nil
	case "null":
		return NewNullDriver(), nil
	default:
		return nil, errors.BroadcastDriverNotSupported.Args(conn.Driver)
	}
}
