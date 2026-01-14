package middleware

import (
	"crypto/subtle"
	"net/url"
	"path"
	"strings"

	contractshttp "github.com/goravel/framework/contracts/http"
)

const HeaderCsrfKey = "X-CSRF-TOKEN"

func VerifyCsrfToken(excepts ...[]string) contractshttp.Middleware {
	var exceptPaths []string
	if len(excepts) > 0 {
		exceptPaths = parseExceptPaths(excepts[0])
	}

	return func(ctx contractshttp.Context) {
		if isReading(ctx.Request().Method()) || isExcept(exceptPaths, ctx.Request().Path()) || isTokenMatch(ctx) {
			ctx.Response().Header(HeaderCsrfKey, ctx.Request().Session().Token())
			ctx.Request().Next()
		} else {
			ctx.Request().Abort(contractshttp.StatusTokenMismatch)
		}
	}
}

func isTokenMatch(ctx contractshttp.Context) bool {
	if !ctx.Request().HasSession() {
		return false
	}
	sessionCsrfToken := ctx.Request().Session().Token()
	requestCsrfToken := ctx.Request().Header(HeaderCsrfKey)
	if requestCsrfToken == "" {
		requestCsrfToken = ctx.Request().Input("_token")
	}
	if requestCsrfToken == "" || subtle.ConstantTimeCompare([]byte(requestCsrfToken), []byte(sessionCsrfToken)) == 0 {
		return false
	}
	return true
}

func isExcept(excepts []string, currentPath string) bool {
	currentPath = strings.Trim(currentPath, "/")
	for _, pattern := range excepts {
		if matched, err := path.Match(pattern, currentPath); err == nil && matched {
			return true
		}
	}
	return false
}

func isReading(method string) bool {
	return method == contractshttp.MethodGet || method == contractshttp.MethodHead || method == contractshttp.MethodOptions
}

func parseExceptPaths(rawExcepts []string) []string {
	var paths []string
	for _, except := range rawExcepts {
		if u, err := url.Parse(except); err == nil && u.Path != "" {
			paths = append(paths, strings.Trim(u.Path, "/"))
		} else {
			paths = append(paths, strings.Trim(except, "/"))
		}
	}
	return paths
}
