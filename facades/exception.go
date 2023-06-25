package facades

import "github.com/goravel/framework/contracts/exception"

func Exception() exception.Exception {
	return App().MakeException()
}
