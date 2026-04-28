package session

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/support/carbon"
)

// WriteCookie emits the current session ID as a Set-Cookie header on the
// response, using the standard session.* config keys for cookie
// attributes. The StartSession middleware uses it to issue the session
// cookie, and application code can call it after rotating the session
// ID (for example, after auth.Login or auth.Logout) so the rotated ID
// reaches the client. Safe to call with a partially initialised
// context — returns without effect if the context, response, session,
// or config facade is nil.
func WriteCookie(ctx http.Context) {
	if ctx == nil || ConfigFacade == nil {
		return
	}
	req := ctx.Request()
	if req == nil {
		return
	}
	s := req.Session()
	if s == nil {
		return
	}
	resp := ctx.Response()
	if resp == nil {
		return
	}

	resp.Cookie(http.Cookie{
		Name:     s.GetName(),
		Value:    s.GetID(),
		Expires:  carbon.Now().AddMinutes(ConfigFacade.GetInt("session.lifetime", 120)).StdTime(),
		Path:     ConfigFacade.GetString("session.path"),
		Domain:   ConfigFacade.GetString("session.domain"),
		Secure:   ConfigFacade.GetBool("session.secure"),
		HttpOnly: ConfigFacade.GetBool("session.http_only"),
		SameSite: ConfigFacade.GetString("session.same_site"),
	})
}
