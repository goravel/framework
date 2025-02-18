package facades

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/http/client"
)

func RateLimiter() http.RateLimiter {
	return App().MakeRateLimiter()
}

func View() http.View {
	return App().MakeView()
}

func Http() client.Request {
	return App().MakeHttp()
}
