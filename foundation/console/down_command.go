package console

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
)

type DownCommand struct {
	app foundation.Application
}

type DownOptions struct {
	Reason   string `json:"reason,omitempty"`
	Redirect string `json:"redirect,omitempty"`
	Render   string `json:"render,omitempty"`
	Secret   string `json:"secret,omitempty"`
	Status   int    `json:"status"`
}

func NewDownCommand(app foundation.Application) *DownCommand {
	return &DownCommand{app}
}

// Signature The name and signature of the console command.
func (r *DownCommand) Signature() string {
	return "down"
}

// Description The console command description.
func (r *DownCommand) Description() string {
	return "Put the application into maintenance mode"
}

// Extend The console command extend.
func (r *DownCommand) Extend() command.Extend {
	return command.Extend{
		Flags: []command.Flag{
			&command.StringFlag{
				Name:  "reason",
				Usage: "The reason for maintenance to show in the response",
				Value: "The application is under maintenance",
			},
			&command.StringFlag{
				Name:  "redirect",
				Usage: "The path that the user should be redirected to",
			},
			&command.StringFlag{
				Name:  "render",
				Usage: "The view should be prerendered for display during maintenance mode",
			},
			&command.StringFlag{
				Name:  "secret",
				Usage: "The secret phrase that may be used to bypass the maintenance mode",
			},
			&command.BoolFlag{
				Name:  "with-secret",
				Usage: "Generate a random secret phrase that may be used to bypass the maintenance mode",
			},
			&command.IntFlag{
				Name:  "status",
				Usage: "The status code that should be used when returning the maintenance mode response",
				Value: http.StatusServiceUnavailable,
			},
		},
	}
}

// Handle Execute the console command.
func (r *DownCommand) Handle(ctx console.Context) error {
	path := r.app.StoragePath("framework/maintenance")

	if ok := file.Exists(path); ok {
		ctx.Error("The application is in maintenance mode already!")

		return nil
	}

	options := DownOptions{}

	options.Status = ctx.OptionInt("status")

	if redirect := ctx.Option("redirect"); redirect != "" {
		options.Redirect = redirect
	}

	if render := ctx.Option("render"); render != "" {
		if r.app.MakeView().Exists(render) {
			options.Render = render
		} else {
			ctx.Error("Unable to find the view template")
			return nil
		}
	}

	if options.Redirect == "" && options.Render == "" {
		options.Reason = ctx.Option("reason")
	}

	if secret := ctx.Option("secret"); secret != "" {
		hash, err := r.app.MakeHash().Make(secret)
		if err != nil {
			ctx.Error("Unable to generate and hash the secret")
		} else {
			options.Secret = hash
		}
	}

	if withSecret := ctx.OptionBool("with-secret"); withSecret {
		secret := str.Random(40)
		hash, err := r.app.MakeHash().Make(secret)

		if err != nil {
			ctx.Error("Unable to generate and hash the secret")
			return nil
		} else {
			options.Secret = hash
			ctx.Info(fmt.Sprintf("Using secret: %s", secret))
		}
	}

	jsonBytes, err := json.Marshal(options)

	if err != nil {
		return err
	}

	if err := file.PutContent(path, string(jsonBytes)); err != nil {
		return err
	}

	ctx.Info("The application is in maintenance mode now")

	return nil
}
