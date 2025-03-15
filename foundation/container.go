package foundation

import (
	"context"
	"fmt"
	"sync"

	"github.com/goravel/framework/contracts"
	contractsauth "github.com/goravel/framework/contracts/auth"
	contractsaccess "github.com/goravel/framework/contracts/auth/access"
	contractscache "github.com/goravel/framework/contracts/cache"
	contractsconfig "github.com/goravel/framework/contracts/config"
	contractsconsole "github.com/goravel/framework/contracts/console"
	contractscrypt "github.com/goravel/framework/contracts/crypt"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	contractsmigration "github.com/goravel/framework/contracts/database/schema"
	contractsseerder "github.com/goravel/framework/contracts/database/seeder"
	contractsevent "github.com/goravel/framework/contracts/event"
	contractsfilesystem "github.com/goravel/framework/contracts/filesystem"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	contractsgrpc "github.com/goravel/framework/contracts/grpc"
	contractshash "github.com/goravel/framework/contracts/hash"
	contractshttp "github.com/goravel/framework/contracts/http"
	contractslog "github.com/goravel/framework/contracts/log"
	contractsmail "github.com/goravel/framework/contracts/mail"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	contractsroute "github.com/goravel/framework/contracts/route"
	contractsschedule "github.com/goravel/framework/contracts/schedule"
	contractsession "github.com/goravel/framework/contracts/session"
	contractstesting "github.com/goravel/framework/contracts/testing"
	contractstranslation "github.com/goravel/framework/contracts/translation"
	contractsvalidation "github.com/goravel/framework/contracts/validation"
	"github.com/goravel/framework/support/color"
)

type instance struct {
	concrete any
	shared   bool
}

type Container struct {
	bindings  sync.Map
	instances sync.Map
}

func NewContainer() *Container {
	return &Container{}
}

func (c *Container) Bind(key any, callback func(app contractsfoundation.Application) (any, error)) {
	c.bindings.Store(key, instance{concrete: callback, shared: false})
}

func (c *Container) BindWith(key any, callback func(app contractsfoundation.Application, parameters map[string]any) (any, error)) {
	c.bindings.Store(key, instance{concrete: callback, shared: false})
}

func (c *Container) Instance(key any, ins any) {
	c.bindings.Store(key, instance{concrete: ins, shared: true})
}

func (c *Container) Make(key any) (any, error) {
	return c.make(key, nil)
}

func (c *Container) MakeArtisan() contractsconsole.Artisan {
	instance, err := c.Make(contracts.BindingConsole)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractsconsole.Artisan)
}

func (c *Container) MakeAuth(ctx contractshttp.Context) contractsauth.Auth {
	instance, err := c.MakeWith(contracts.BindingAuth, map[string]any{
		"ctx": ctx,
	})
	if err != nil {
		color.Errorln(err)
		return nil
	}
	if instance == nil {
		return nil
	}

	return instance.(contractsauth.Auth)
}

func (c *Container) MakeCache() contractscache.Cache {
	instance, err := c.Make(contracts.BindingCache)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractscache.Cache)
}

func (c *Container) MakeConfig() contractsconfig.Config {
	instance, err := c.Make(contracts.BindingConfig)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractsconfig.Config)
}

func (c *Container) MakeCrypt() contractscrypt.Crypt {
	instance, err := c.Make(contracts.BindingCrypt)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractscrypt.Crypt)
}

func (c *Container) MakeEvent() contractsevent.Instance {
	instance, err := c.Make(contracts.BindingEvent)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractsevent.Instance)
}

func (c *Container) MakeGate() contractsaccess.Gate {
	instance, err := c.Make(contracts.BindingGate)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractsaccess.Gate)
}

func (c *Container) MakeGrpc() contractsgrpc.Grpc {
	instance, err := c.Make(contracts.BindingGrpc)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractsgrpc.Grpc)
}

func (c *Container) MakeHash() contractshash.Hash {
	instance, err := c.Make(contracts.BindingHash)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractshash.Hash)
}

func (c *Container) MakeLang(ctx context.Context) contractstranslation.Translator {
	instance, err := c.MakeWith(contracts.BindingTranslation, map[string]any{
		"ctx": ctx,
	})
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractstranslation.Translator)
}

func (c *Container) MakeLog() contractslog.Log {
	instance, err := c.Make(contracts.BindingLog)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractslog.Log)
}

func (c *Container) MakeMail() contractsmail.Mail {
	instance, err := c.Make(contracts.BindingMail)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractsmail.Mail)
}

func (c *Container) MakeOrm() contractsorm.Orm {
	instance, err := c.Make(contracts.BindingOrm)
	if err != nil {
		color.Errorln(err)
		return nil
	}
	if instance == nil {
		return nil
	}

	return instance.(contractsorm.Orm)
}

func (c *Container) MakeQueue() contractsqueue.Queue {
	instance, err := c.Make(contracts.BindingQueue)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractsqueue.Queue)
}

func (c *Container) MakeRateLimiter() contractshttp.RateLimiter {
	instance, err := c.Make(contracts.BindingRateLimiter)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractshttp.RateLimiter)
}

func (c *Container) MakeRoute() contractsroute.Route {
	instance, err := c.Make(contracts.BindingRoute)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractsroute.Route)
}

func (c *Container) MakeSchedule() contractsschedule.Schedule {
	instance, err := c.Make(contracts.BindingSchedule)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractsschedule.Schedule)
}

func (c *Container) MakeSchema() contractsmigration.Schema {
	instance, err := c.Make(contracts.BindingSchema)
	if err != nil {
		color.Errorln(err)
		return nil
	}
	if instance == nil {
		return nil
	}

	return instance.(contractsmigration.Schema)
}

func (c *Container) MakeSession() contractsession.Manager {
	instance, err := c.Make(contracts.BindingSession)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractsession.Manager)
}

func (c *Container) MakeStorage() contractsfilesystem.Storage {
	instance, err := c.Make(contracts.BindingFilesystem)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractsfilesystem.Storage)
}

func (c *Container) MakeTesting() contractstesting.Testing {
	instance, err := c.Make(contracts.BindingTesting)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractstesting.Testing)
}

func (c *Container) MakeValidation() contractsvalidation.Validation {
	instance, err := c.Make(contracts.BindingValidation)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractsvalidation.Validation)
}

func (c *Container) MakeView() contractshttp.View {
	instance, err := c.Make(contracts.BindingView)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractshttp.View)
}

func (c *Container) MakeSeeder() contractsseerder.Facade {
	instance, err := c.Make(contracts.BindingSeeder)

	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractsseerder.Facade)
}

func (c *Container) MakeWith(key any, parameters map[string]any) (any, error) {
	return c.make(key, parameters)
}

func (c *Container) Refresh(bindings ...any) {
	if len(bindings) == 0 {
		// configInstance, ok := c.instances.Load(contracts.BindingConfig)

		c.instances.Range(func(key, value any) bool {
			if key != contracts.BindingConfig {
				c.instances.Delete(key)
			}

			return true
		})
	} else {
		for _, binding := range bindings {
			c.instances.Delete(binding)
		}
	}
}

func (c *Container) Singleton(key any, callback func(app contractsfoundation.Application) (any, error)) {
	c.bindings.Store(key, instance{concrete: callback, shared: true})
}

func (c *Container) make(key any, parameters map[string]any) (any, error) {
	binding, ok := c.bindings.Load(key)
	if !ok {
		return nil, fmt.Errorf("binding not found: %+v", key)
	}

	if parameters == nil {
		instance, ok := c.instances.Load(key)
		if ok {
			return instance, nil
		}
	}

	bindingImpl := binding.(instance)
	switch concrete := bindingImpl.concrete.(type) {
	case func(app contractsfoundation.Application) (any, error):
		concreteImpl, err := concrete(App)
		if err != nil {
			return nil, err
		}
		if bindingImpl.shared {
			c.instances.Store(key, concreteImpl)
		}

		return concreteImpl, nil
	case func(app contractsfoundation.Application, parameters map[string]any) (any, error):
		concreteImpl, err := concrete(App, parameters)
		if err != nil {
			return nil, err
		}

		return concreteImpl, nil
	default:
		c.instances.Store(key, concrete)

		return concrete, nil
	}
}
