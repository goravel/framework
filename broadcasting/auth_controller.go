package broadcasting

import (
	"fmt"

	"github.com/goravel/framework/contracts/broadcasting"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

type AuthController struct{}

func (c *AuthController) Authenticate(ctx http.Context) http.Response {
	socketID := ctx.Request().Query("socket_id")
	channelName := ctx.Request().Query("channel_name")

	auth := facades.Auth(ctx)
	if !auth.Check() {
		return ctx.Response().Json(403, http.Json{"error": "Unauthenticated"})
	}

	ch := broadcasting.Channel{Name: channelName}
	if !ch.IsPrivate() && !ch.IsPresence() {
		return ctx.Response().Json(200, broadcasting.AuthResponse{})
	}

	userID, _ := auth.ID()
	user := map[string]any{"id": userID}

	app := facades.Broadcast().(*Application)
	if !app.resolveAuth(ch.BaseName(), user) {
		return ctx.Response().Json(403, http.Json{"error": "Forbidden"})
	}

	channelData := ctx.Request().Input("channel_data")
	cfg := NewConfig(facades.Config())
	conn, err := cfg.Connection(cfg.DefaultConnection())
	if err != nil {
		return ctx.Response().Json(500, http.Json{"error": err.Error()})
	}

	signature := computeAuthSignature(conn.Secret, socketID, channelName, channelData)
	resp := broadcasting.AuthResponse{
		Auth: fmt.Sprintf("%s:%s", conn.Key, signature),
	}
	if ch.IsPresence() && channelData != "" {
		resp.ChannelData = channelData
	}
	return ctx.Response().Json(200, resp)
}
