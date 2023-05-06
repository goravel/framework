package foundation

import (
	"fmt"
	"log"
	"sync"

	"github.com/goravel/framework/config"
	"github.com/goravel/framework/console"
	configcontract "github.com/goravel/framework/contracts/config"
	consolecontract "github.com/goravel/framework/contracts/console"
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

func (c *Container) Bind(key any, callback func() (any, error)) {
	c.bindings.Store(key, instance{concrete: callback, shared: false})
}

func (c *Container) BindWith(key any, callback func(parameters map[string]any) (any, error)) {
	c.bindings.Store(key, instance{concrete: callback, shared: false})
}

func (c *Container) Make(key any) (any, error) {
	return c.make(key, nil)
}

func (c *Container) MakeArtisan() consolecontract.Artisan {
	instance, err := c.Make(console.Binding)
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	return instance.(consolecontract.Artisan)
}

func (c *Container) MakeConfig() configcontract.Config {
	instance, err := c.Make(config.Binding)
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	return instance.(configcontract.Config)
}

func (c *Container) MakeWith(key any, parameters map[string]any) (any, error) {
	return c.make(key, parameters)
}

func (c *Container) Singleton(key any, callback func() (any, error)) {
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
	case func() (any, error):
		concreteImpl, err := concrete()
		if err != nil {
			return nil, err
		}
		if bindingImpl.shared {
			c.instances.Store(key, concreteImpl)
		}

		return concreteImpl, nil
	case func(parameters map[string]any) (any, error):
		concreteImpl, err := concrete(parameters)
		if err != nil {
			return nil, err
		}

		return concreteImpl, nil
	default:
		return nil, fmt.Errorf("binding type error: %+v", binding)
	}
}
