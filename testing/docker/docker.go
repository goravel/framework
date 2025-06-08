package docker

import (
	contractsconfig "github.com/goravel/framework/contracts/config"
	contractsconsole "github.com/goravel/framework/contracts/console"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/testing/docker"
)

type Docker struct {
	artisan contractsconsole.Artisan
	config  contractsconfig.Config
	orm     contractsorm.Orm
}

func NewDocker(artisan contractsconsole.Artisan, config contractsconfig.Config, orm contractsorm.Orm) *Docker {
	return &Docker{
		artisan: artisan,
		config:  config,
		orm:     orm,
	}
}

func (r *Docker) Cache(connection string) (docker.Cache, error) {
	return nil, nil
}

func (r *Docker) Database(connection ...string) (docker.Database, error) {
	if len(connection) == 0 {
		return NewDatabase(r.artisan, r.config, r.orm, "")
	} else {
		return NewDatabase(r.artisan, r.config, r.orm, connection[0])
	}
}
