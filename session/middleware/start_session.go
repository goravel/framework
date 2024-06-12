package middleware

import (
	"math/rand"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/session"
	"github.com/goravel/framework/support/carbon"
)

func StartSession() http.Middleware {
	return func(ctx http.Context) {
		req := ctx.Request()

		// Check if session exists
		if req.HasSession() || session.ConfigFacade.GetString("session.driver") == "" {
			req.Next()
			return
		}

		// Retrieve session driver
		driver, err := session.SessionFacade.Driver()
		if err != nil {
			panic(err)
		}

		// Build session
		s := session.SessionFacade.BuildSession(driver)
		s.SetID(req.Cookie(s.GetName()))

		// Start session
		s.Start()
		req.SetSession(s)

		// Perform garbage collection based on lottery
		lottery := session.ConfigFacade.Get("session.lottery").([]int)
		if len(lottery) == 2 {
			if rand.Intn(lottery[1])+1 <= lottery[0] {
				lifetime := session.ConfigFacade.GetInt("session.lifetime") * 60
				if err := driver.Gc(lifetime); err != nil {
					panic(err)
				}
			}
		}

		// Set session cookie in response
		config := session.ConfigFacade
		ctx.Response().Cookie(http.Cookie{
			Name:     s.GetName(),
			Value:    s.GetID(),
			Expires:  carbon.Now().AddMinutes(config.GetInt("session.lifetime")).StdTime(),
			Path:     config.GetString("session.path"),
			Domain:   config.GetString("session.domain"),
			Secure:   config.GetBool("session.secure"),
			HttpOnly: config.GetBool("session.http_only"),
			SameSite: config.GetString("session.same_site"),
		})

		// Continue processing request
		req.Next()

		// Save session
		if err := s.Save(); err != nil {
			panic(err)
		}
	}
}
