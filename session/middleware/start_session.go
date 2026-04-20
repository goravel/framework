package middleware

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/session"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/color"
)

func StartSession() http.Middleware {
	return func(ctx http.Context) {
		req := ctx.Request()

		// Check if session exists
		if req.HasSession() || session.ConfigFacade.GetString("session.default") == "" {
			req.Next()
			return
		}

		// Retrieve session driver
		driver, err := session.SessionFacade.Driver()
		if err != nil {
			color.Errorln(err)
			req.Next()
			return
		}

		// Build session
		s, err := session.SessionFacade.BuildSession(driver)
		if err != nil {
			color.Errorln(err)
			req.Next()
			return
		}

		incomingCookie := req.Cookie(s.GetName())
		s.SetID(incomingCookie)

		// Start session
		s.Start()
		req.SetSession(s)

		// Only write the session cookie when the client doesn't already have
		// a matching one. Rotations triggered in the handler (e.g. login /
		// logout) reissue the cookie themselves, so this avoids emitting a
		// duplicate Set-Cookie header in that flow.
		if incomingCookie != s.GetID() {
			config := session.ConfigFacade
			ctx.Response().Cookie(http.Cookie{
				Name:     s.GetName(),
				Value:    s.GetID(),
				Expires:  carbon.Now().AddMinutes(config.GetInt("session.lifetime", 120)).StdTime(),
				Path:     config.GetString("session.path"),
				Domain:   config.GetString("session.domain"),
				Secure:   config.GetBool("session.secure"),
				HttpOnly: config.GetBool("session.http_only"),
				SameSite: config.GetString("session.same_site"),
			})
		}

		// Continue processing request
		req.Next()

		// Save session
		if err = s.Save(); err != nil {
			color.Errorf("Error saving session: %s\n", err)
		}

		// Release session
		session.SessionFacade.ReleaseSession(s)
	}
}
