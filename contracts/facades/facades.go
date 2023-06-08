package facades

import (
	"github.com/goravel/framework/contracts/config"
)

/*
 * The functions here are faked for a third package, the third package can only use https://github.com/goravel/contracts
 * instead of https://github.com/goravel/framework that doesn't need to install many unused package. If you import
 * `github.com/goravel/framework/contracts/facades` in your package, it will be replaced by `github.com/goravel/framework/facades`
 * when Users run vendor:publish command.
 */

func Config() config.Config {
	return nil
}
