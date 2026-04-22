package middleware

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/session"
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
			session.WriteCookie(ctx)
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
