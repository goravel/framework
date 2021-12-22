package support

import (
	"github.com/goravel/framework/contracts/cache"
)

type Store interface {
	Handle() cache.Store
}
