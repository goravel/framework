package sms

import (
	"errors"
	"sync"
	"testing"

	"github.com/goravel/framework/contracts"
	"github.com/goravel/framework/facades"
)

func TestSms(t *testing.T) {
	container := new(Container)
	container.Bind("sms", &Sms{})
	_ = container.MakeFacade(facades.Sms)

	facades.Sms.Send()
}

type Container struct {
	Bindings sync.Map
}

func (c *Container) Bind(key any, value any) {
	c.Bindings.Store(key, value)
}

func (c *Container) Make(key any) (any, error) {
	binding, ok := c.Bindings.Load(key)
	if !ok {
		return nil, errors.New("binding not found")
	}

	return binding, nil
}

func (c *Container) MakeFacade(facade contracts.Facade) error {
	binding, ok := c.Bindings.Load(facade.GetFacadeAccessor())
	if !ok {
		return errors.New("binding not found")
	}

	return facade.ResolveFacadeInstance(binding)
}
