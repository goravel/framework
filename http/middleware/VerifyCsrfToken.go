package middleware

import (
	"crypto/subtle"
	"path"

	contractshttp "github.com/goravel/framework/contracts/http"
)

const csrfKey = "X-CSRF-TOKEN"

func VerifyCsrfToken(excepts []string) contractshttp.Middleware {
	return func(ctx contractshttp.Context) {
		if isReading(ctx.Request().Method()) || inExceptArray(excepts, ctx.Request().FullUrl()) || tokenMatch(ctx) {
			ctx.Request().Next()
			ctx.Request().Session().Put(csrfKey, ctx.Request().Session().Token())
		} else {
			ctx.Request().AbortWithStatusJson(contractshttp.StatusTokenMismatch, map[string]string{"message": "CSRF token mismatch."})
		}
	}
}

func tokenMatch(ctx contractshttp.Context) bool {
	if !ctx.Request().HasSession() {
		return false
	}
	sessionCsrfToken := ctx.Request().Session().Token()
	requestCsrfToken := ctx.Request().Header(csrfKey)
	if requestCsrfToken == "" {
		requestCsrfToken = ctx.Request().Input("_token")
	}
	if requestCsrfToken == "" || subtle.ConstantTimeCompare([]byte(requestCsrfToken), []byte(sessionCsrfToken)) == 0 {
		return false
	}
	return true
}

func inExceptArray(excepts []string, url string) bool {
	for _, except := range excepts {
		matched, err := path.Match(except, url)
		if err != nil {
			continue
		}
		if matched {
			return true
		}
	}
	return false
}

func isReading(method string) bool {
	return method == "GET" || method == "HEAD" || method == "OPTIONS"
}
