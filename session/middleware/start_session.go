package middleware

import (
	"math/rand"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/session"
	"github.com/goravel/framework/support/carbon"
)

func StartSession() http.Middleware {
	return func(ctx http.Context) {
		if ctx.Request().HasSession() {
			ctx.Request().Next()
			return
		}

		d, err := session.Facade.Driver()
		if err != nil {
			return
		}

		s := session.Facade.BuildSession(d)

		s.SetID(ctx.Request().Cookie(s.GetName()))
		s.Start()
		ctx.Request().SetSession(s)

		lottery := session.ConfigFacade.Get("lottery").([]int)
		if len(lottery) == 2 {
			randInt := rand.Intn(lottery[1]) + 1
			if randInt <= lottery[0] {
				err = d.Gc(300)
				if err != nil {
					return
				}
			}
		}

		ctx.Request().Next()

		s = ctx.Request().Session()

		config := session.ConfigFacade
		ctx.Response().Cookie(http.Cookie{
			Name:     s.GetName(),
			Value:    s.GetID(),
			Expires:  carbon.Now().AddMinutes(config.GetInt("lifetime")).StdTime(),
			Path:     config.GetString("path"),
			Domain:   config.GetString("domain"),
			Secure:   config.GetBool("secure"),
			HttpOnly: config.GetBool("http_only"),
			SameSite: config.GetString("same_site"),
		})

		err = s.Save()
		if err != nil {
			return
		}
	}
}
