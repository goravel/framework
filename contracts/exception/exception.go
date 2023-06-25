package exception

import "github.com/goravel/framework/contracts/http"

//go:generate mockery --name=Exception
type Exception interface {
	Report(err error)
	Render(ctx http.Context, err error)
}
