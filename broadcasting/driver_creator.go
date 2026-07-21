package broadcasting

import (
	"fmt"

	"github.com/goravel/framework/contracts/broadcasting"
	"github.com/goravel/framework/contracts/foundation"
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
		return nil, fmt.Errorf("unknown broadcast driver: %s", conn.Driver)
	}
}
