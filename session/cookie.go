package session

import (
	"github.com/goravel/framework/contracts/http"
	sessioncontract "github.com/goravel/framework/contracts/session"
	"github.com/goravel/framework/support/carbon"
)

// WriteCookie sends the session ID to the client as a cookie. Call it again
// after rotating the session ID (e.g. Regenerate or Invalidate) so the new ID
// reaches the client; the response may already carry a cookie with the stale ID.
func WriteCookie(ctx http.Context, session sessioncontract.Session) {
	ctx.Response().Cookie(http.Cookie{
		Name:     session.GetName(),
		Value:    session.GetID(),
		Expires:  carbon.Now().AddMinutes(ConfigFacade.GetInt("session.lifetime", 120)).StdTime(),
		Path:     ConfigFacade.GetString("session.path"),
		Domain:   ConfigFacade.GetString("session.domain"),
		Secure:   ConfigFacade.GetBool("session.secure"),
		HttpOnly: ConfigFacade.GetBool("session.http_only"),
		SameSite: ConfigFacade.GetString("session.same_site"),
	})
}
