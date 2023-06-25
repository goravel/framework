package exception

import (
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http"
	"github.com/spf13/cast"
	"reflect"
	"runtime/debug"
)

const Binding = "goravel.exception"

type Handler struct {
	DontReport []error
	DontFlash  []error

	Config config.Config
}

func (h *Handler) Boot(app foundation.Application) {

}

func (h *Handler) Report(e error) {

	if h.shouldntReport(e) {
		return
	}

	// check if e has function Report using reflect
	if reflect.ValueOf(e).IsValid() && reflect.ValueOf(e).MethodByName("Report").IsValid() {
		reflect.ValueOf(e).MethodByName("Report").Call([]reflect.Value{
			reflect.ValueOf(e),
		})
		return
	}

	// will not report

}

func (h *Handler) shouldReport(e error) bool {
	return !h.shouldntReport(e)
}

func (h *Handler) shouldntReport(e error) bool {
	for _, err := range h.DontReport {
		if err == e {
			return true
		}
	}
	return false
}

func (h *Handler) exceptionContext(e error) map[string]interface{} {
	if reflect.ValueOf(e).IsValid() && reflect.ValueOf(e).MethodByName("Context").IsValid() {
		return e.(interface {
			Context() map[string]interface{}
		}).Context()
	}

	return map[string]interface{}{}
}

func (h *Handler) getStatusCode(e error) int {
	if reflect.ValueOf(e).IsValid() && reflect.ValueOf(e).MethodByName("StatusCode").IsValid() {
		return e.(interface {
			StatusCode() int
		}).StatusCode()
	}

	return 500
}

func (h *Handler) Context(e error) map[string]interface{} {
	return map[string]interface{}{
		// "userId": 1,
	}
}

func (h *Handler) Render(ctx http.Context, e error) {
	if reflect.ValueOf(e).IsValid() && reflect.ValueOf(e).MethodByName("Render").IsValid() {
		reflect.ValueOf(e).MethodByName("Render").Call([]reflect.Value{
			reflect.ValueOf(ctx),
			reflect.ValueOf(e),
		})
		return
	}

	// default render
	if h.shouldReturnJson(ctx, e) {
		ctx.Response().Status(h.getStatusCode(e)).Json(h.prepareJson(ctx, e))
		return
	}

	// TODO: render html
	ctx.Response().Status(h.getStatusCode(e)).Json(h.prepareJson(ctx, e))
}

func (h *Handler) shouldReturnJson(ctx http.Context, e error) bool {
	return ctx.Request().ExpectsJson()
}

func (h *Handler) prepareJson(ctx http.Context, e error) http.Json {
	if h.Config.GetBool("app.debug") {
		return http.Json{
			"message":   cast.ToString(e),
			"exception": e,
			"trace":     string(debug.Stack()),
		}
	}
	return http.Json{
		"message": e.Error(),
	}
}
