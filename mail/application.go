package mail

import (
	"github.com/goravel/framework/contracts/mail"
)

type Application struct {
}

func (app *Application) Init() mail.Mail {
	return NewEmail()
}
