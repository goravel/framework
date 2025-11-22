package configuration

import (
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/support"
)

type Paths struct {
}

func NewPaths() *Paths {
	return &Paths{}
}

func (r *Paths) Bootstrap(path string) configuration.Paths {
	support.Config.Paths.Bootstrap = path

	return r
}

func (r *Paths) Command(path string) configuration.Paths {
	support.Config.Paths.Command = path

	return r
}

func (r *Paths) Controller(path string) configuration.Paths {
	support.Config.Paths.Controller = path

	return r
}

func (r *Paths) Event(path string) configuration.Paths {
	support.Config.Paths.Event = path

	return r
}

func (r *Paths) Factory(path string) configuration.Paths {
	support.Config.Paths.Factory = path

	return r
}

func (r *Paths) Filter(path string) configuration.Paths {
	support.Config.Paths.Filter = path

	return r
}

func (r *Paths) Job(path string) configuration.Paths {
	support.Config.Paths.Job = path

	return r
}

func (r *Paths) Listener(path string) configuration.Paths {
	support.Config.Paths.Listener = path

	return r
}

func (r *Paths) Mail(path string) configuration.Paths {
	support.Config.Paths.Mail = path

	return r
}

func (r *Paths) Middleware(path string) configuration.Paths {
	support.Config.Paths.Middleware = path

	return r
}

func (r *Paths) Migration(path string) configuration.Paths {
	support.Config.Paths.Migration = path

	return r
}

func (r *Paths) Model(path string) configuration.Paths {
	support.Config.Paths.Model = path

	return r
}

func (r *Paths) Observer(path string) configuration.Paths {
	support.Config.Paths.Observer = path

	return r
}

func (r *Paths) Package(path string) configuration.Paths {
	support.Config.Paths.Package = path

	return r
}

func (r *Paths) Policy(path string) configuration.Paths {
	support.Config.Paths.Policy = path

	return r
}

func (r *Paths) Provider(path string) configuration.Paths {
	support.Config.Paths.Provider = path

	return r
}

func (r *Paths) Request(path string) configuration.Paths {
	support.Config.Paths.Request = path

	return r
}

func (r *Paths) Rule(path string) configuration.Paths {
	support.Config.Paths.Rule = path

	return r
}

func (r *Paths) Seeder(path string) configuration.Paths {
	support.Config.Paths.Seeder = path

	return r
}

func (r *Paths) Test(path string) configuration.Paths {
	support.Config.Paths.Test = path

	return r
}
