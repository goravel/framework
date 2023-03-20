package console

type Stubs struct {
}

func (r Stubs) Request() string {
	return `package requests

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type DummyRequest struct {
	DummyField
}

func (r *DummyRequest) Authorize(ctx http.Context) error {
	return nil
}

func (r *DummyRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{}
}

func (r *DummyRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{}
}

func (r *DummyRequest) Attributes(ctx http.Context) map[string]string {
	return map[string]string{}
}

func (r *DummyRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return nil
}
`
}

func (r Stubs) Controller() string {
	return `package controllers

import (
	"github.com/goravel/framework/contracts/http"
)

type DummyController struct {
	//Dependent services
}

func NewDummyController() *DummyController {
	return &DummyController{
		//Inject services
	}
}

func (r *DummyController) Index(ctx http.Context) {
}	
`
}

func (r Stubs) Middleware() string {
	return `package middleware

import (
	"github.com/goravel/framework/contracts/http"
)

func DummyMiddleware() http.Middleware {
	return func(ctx http.Context) {
		ctx.Request().Next()
	}
}
`
}
