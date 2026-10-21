package testing

import (
	"github.com/brianvoe/gofakeit/v7"

	"github.com/goravel/framework/contracts/testing"
)

var _ testing.Fake = (*Fake)(nil)

type Fake struct {
	*gofakeit.Faker
}

func NewFake() *Fake {
	return &Fake{
		Faker: gofakeit.New(0),
	}
}

func (r *Fake) Instance() *gofakeit.Faker {
	return r.Faker
}
