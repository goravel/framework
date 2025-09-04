package middleware

import (
	"encoding/json"
	"io"
	"strings"

	httpcontract "github.com/goravel/framework/contracts/http"
)

const csrfKey = "X-CSRF-TOKEN"

type BodyPayload struct {
	CSRFToken string `json:"csrf_token"`
}

func CSRFToken() httpcontract.Middleware {
	return func(ctx httpcontract.Context) {
		sessionCSRFTokenInterface := ctx.Request().Session().Get(csrfKey)
		sessionCSRFToken, ok := sessionCSRFTokenInterface.(string)
		if !ok || sessionCSRFToken == "" {
			ctx.Request().Abort(httpcontract.StatusUnauthorized)
			return
		}

		requestCSRFToken := ctx.Request().Header(csrfKey)
		if requestCSRFToken == "" {
			requestCSRFToken = ctx.Request().Origin().FormValue("csrf_token")
		}

		if requestCSRFToken == "" {
			httpReq := ctx.Request().Origin()
			if httpReq != nil && httpReq.Body != nil {
				rawBody, err := io.ReadAll(httpReq.Body)
				if err != nil {
					ctx.Request().Abort(httpcontract.StatusBadRequest)
					return
				}

				httpReq.Body = io.NopCloser(strings.NewReader(string(rawBody)))
				var input BodyPayload
				if err := json.Unmarshal(rawBody, &input); err == nil && input.CSRFToken != "" {
					requestCSRFToken = input.CSRFToken
				}
			}
		}

		if requestCSRFToken == "" || requestCSRFToken != sessionCSRFToken {
			ctx.Request().Abort(httpcontract.StatusUnauthorized)
			return
		}

		ctx.Request().Next()
	}
}
