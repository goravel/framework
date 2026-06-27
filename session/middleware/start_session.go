package middleware

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/session"
	"github.com/goravel/framework/support/color"
)

type startSessionMiddleware struct{}

func (s *startSessionMiddleware) Signature() string {
	return "goravel:start_session"
}

func (s *startSessionMiddleware) Handle(ctx http.Context) {
	req := ctx.Request()

	if req.HasSession() || session.ConfigFacade.GetString("session.default") == "" {
		req.Next()
		return
	}

	driver, err := session.SessionFacade.Driver()
	if err != nil {
		color.Errorln(err)
		req.Next()
		return
	}

	sess, err := session.SessionFacade.BuildSession(driver)
	if err != nil {
		color.Errorln(err)
		req.Next()
		return
	}

	sess.SetID(req.Cookie(sess.GetName()))

	sess.Start()
	req.SetSession(sess)

	session.WriteCookie(ctx, sess)

	req.Next()

	if err = sess.Save(); err != nil {
		color.Errorf("Error saving session: %s\n", err)
	}

	session.SessionFacade.ReleaseSession(sess)
}

func StartSession() http.Middleware {
	return &startSessionMiddleware{}
}
