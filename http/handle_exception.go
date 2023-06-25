package http

import (
	"github.com/goravel/framework/contracts/http"
)

func handleException(ctx http.Context, err error) {
	ExceptionFacade.Report(err)
	ExceptionFacade.Render(ctx, err)
}
