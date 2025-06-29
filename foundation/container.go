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
	contractsdb "github.com/goravel/framework/contracts/database/db"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	contractsmigration "github.com/goravel/framework/contracts/database/schema"
	contractsseerder "github.com/goravel/framework/contracts/database/seeder"
	contractsevent "github.com/goravel/framework/contracts/event"
	contractsfilesystem "github.com/goravel/framework/contracts/filesystem"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	contractsgrpc "github.com/goravel/framework/contracts/grpc"
	contractshash "github.com/goravel/framework/contracts/hash"
	contractshttp "github.com/goravel/framework/contracts/http"
	contractshttpclient "github.com/goravel/framework/contracts/http/client"
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

func (r *Container) Bind(key any, callback func(app contractsfoundation.Application) (any, error)) {
	r.bindings.Store(key, instance{concrete: callback, shared: false})
}

func (r *Container) BindWith(key any, callback func(app contractsfoundation.Application, parameters map[string]any) (any, error)) {
	r.bindings.Store(key, instance{concrete: callback, shared: false})
}

func (r *Container) Fresh(bindings ...any) {
	if len(bindings) == 0 {
		r.instances.Range(func(key, value any) bool {
			if key != contracts.BindingConfig {
				r.instances.Delete(key)
			}

			return true
		})
	} else {
		for _, binding := range bindings {
			r.instances.Delete(binding)
		}
	}
}

func (r *Container) Instance(key any, ins any) {
	r.bindings.Store(key, instance{concrete: ins, shared: true})
}

func (r *Container) Make(key any) (any, error) {
	return r.make(key, nil)
}

func (r *Container) MakeArtisan() contractsconsole.Artisan {
	instance, err := r.Make(contracts.BindingArtisan)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractsconsole.Artisan)
}

func (r *Container) MakeAuth(ctx ...contractshttp.Context) contractsauth.Auth {
	parameters := map[string]any{}
	if len(ctx) > 0 {
		parameters["ctx"] = ctx[0]
	}

	instance, err := r.MakeWith(contracts.BindingAuth, parameters)
	if err != nil {
		color.Errorln(err)
		return nil
	}
	if instance == nil {
		return nil
	}

	return instance.(contractsauth.Auth)
}

func (r *Container) MakeCache() contractscache.Cache {
	instance, err := r.Make(contracts.BindingCache)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractscache.Cache)
}

func (r *Container) MakeConfig() contractsconfig.Config {
	instance, err := r.Make(contracts.BindingConfig)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractsconfig.Config)
}

func (r *Container) MakeCrypt() contractscrypt.Crypt {
	instance, err := r.Make(contracts.BindingCrypt)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractscrypt.Crypt)
}

func (r *Container) MakeDB() contractsdb.DB {
	instance, err := r.Make(contracts.BindingDB)
	if err != nil {
		color.Errorln(err)
		return nil
	}
	if instance == nil {
		return nil
	}

	return instance.(contractsdb.DB)
}

func (r *Container) MakeEvent() contractsevent.Instance {
	instance, err := r.Make(contracts.BindingEvent)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractsevent.Instance)
}

func (r *Container) MakeGate() contractsaccess.Gate {
	instance, err := r.Make(contracts.BindingGate)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractsaccess.Gate)
}

func (r *Container) MakeGrpc() contractsgrpc.Grpc {
	instance, err := r.Make(contracts.BindingGrpc)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractsgrpc.Grpc)
}

func (r *Container) MakeHash() contractshash.Hash {
	instance, err := r.Make(contracts.BindingHash)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractshash.Hash)
}

func (r *Container) MakeHttp() contractshttpclient.Request {
	instance, err := r.Make(contracts.BindingHttp)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractshttpclient.Request)
}

func (r *Container) MakeLang(ctx context.Context) contractstranslation.Translator {
	instance, err := r.MakeWith(contracts.BindingLang, map[string]any{
		"ctx": ctx,
	})
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractstranslation.Translator)
}

func (r *Container) MakeLog() contractslog.Log {
	instance, err := r.Make(contracts.BindingLog)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractslog.Log)
}

func (r *Container) MakeMail() contractsmail.Mail {
	instance, err := r.Make(contracts.BindingMail)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractsmail.Mail)
}

func (r *Container) MakeOrm() contractsorm.Orm {
	instance, err := r.Make(contracts.BindingOrm)
	if err != nil {
		color.Errorln(err)
		return nil
	}
	if instance == nil {
		return nil
	}

	return instance.(contractsorm.Orm)
}

func (r *Container) MakeQueue() contractsqueue.Queue {
	instance, err := r.Make(contracts.BindingQueue)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractsqueue.Queue)
}

func (r *Container) MakeRateLimiter() contractshttp.RateLimiter {
	instance, err := r.Make(contracts.BindingRateLimiter)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractshttp.RateLimiter)
}

func (r *Container) MakeRoute() contractsroute.Route {
	instance, err := r.Make(contracts.BindingRoute)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractsroute.Route)
}

func (r *Container) MakeSchedule() contractsschedule.Schedule {
	instance, err := r.Make(contracts.BindingSchedule)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractsschedule.Schedule)
}

func (r *Container) MakeSchema() contractsmigration.Schema {
	instance, err := r.Make(contracts.BindingSchema)
	if err != nil {
		color.Errorln(err)
		return nil
	}
	if instance == nil {
		return nil
	}

	return instance.(contractsmigration.Schema)
}

func (r *Container) MakeSession() contractsession.Manager {
	instance, err := r.Make(contracts.BindingSession)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractsession.Manager)
}

func (r *Container) MakeStorage() contractsfilesystem.Storage {
	instance, err := r.Make(contracts.BindingStorage)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractsfilesystem.Storage)
}

func (r *Container) MakeTesting() contractstesting.Testing {
	instance, err := r.Make(contracts.BindingTesting)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractstesting.Testing)
}

func (r *Container) MakeValidation() contractsvalidation.Validation {
	instance, err := r.Make(contracts.BindingValidation)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractsvalidation.Validation)
}

func (r *Container) MakeView() contractshttp.View {
	instance, err := r.Make(contracts.BindingView)
	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractshttp.View)
}

func (r *Container) MakeSeeder() contractsseerder.Facade {
	instance, err := r.Make(contracts.BindingSeeder)

	if err != nil {
		color.Errorln(err)
		return nil
	}

	return instance.(contractsseerder.Facade)
}

func (r *Container) MakeWith(key any, parameters map[string]any) (any, error) {
	return r.make(key, parameters)
}

func (r *Container) Singleton(key any, callback func(app contractsfoundation.Application) (any, error)) {
	r.bindings.Store(key, instance{concrete: callback, shared: true})
}

func (r *Container) make(key any, parameters map[string]any) (any, error) {
	binding, ok := r.bindings.Load(key)
	if !ok {
		return nil, fmt.Errorf("binding not found: %+v", key)
	}

	if parameters == nil {
		instance, ok := r.instances.Load(key)
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
			r.instances.Store(key, concreteImpl)
		}

		return concreteImpl, nil
	case func(app contractsfoundation.Application, parameters map[string]any) (any, error):
		concreteImpl, err := concrete(App, parameters)
		if err != nil {
			return nil, err
		}

		return concreteImpl, nil
	default:
		r.instances.Store(key, concrete)

		return concrete, nil
	}
}
