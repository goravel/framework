package middleware

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/session"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/color"
)

func StartSession() http.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(ctx http.Context) http.Response {
			req := ctx.Request()

			// Check if session exists
			if req.HasSession() || session.ConfigFacade.GetString("session.driver") == "" {
				return next.ServeHTTP(ctx)
			}

			// Retrieve session driver
			driver, err := session.SessionFacade.Driver()
			if err != nil {
				color.Errorln(err)
				return next.ServeHTTP(ctx)
			}

			// Build session
			s, err := session.SessionFacade.BuildSession(driver)
			if err != nil {
				color.Errorln(err)
				return next.ServeHTTP(ctx)
			}

			s.SetID(req.Cookie(s.GetName()))

			// Start session
			s.Start()
			req.SetSession(s)

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
			resp := next.ServeHTTP(ctx)

			// Save session
			if err = s.Save(); err != nil {
				color.Errorf("Error saving session: %s\n", err)
			}

			// Release session
			session.SessionFacade.ReleaseSession(s)

			return resp
		})
	}
}
