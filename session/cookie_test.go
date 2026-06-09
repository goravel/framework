package session

import (
	"testing"

	"github.com/goravel/framework/contracts/http"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mockshttp "github.com/goravel/framework/mocks/http"
	mockssession "github.com/goravel/framework/mocks/session"
	"github.com/goravel/framework/support/carbon"
)

func TestWriteCookie(t *testing.T) {
	now := carbon.Now()
	carbon.SetTestNow(now)
	defer carbon.ClearTestNow()

	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetInt("session.lifetime", 120).Return(120).Once()
	mockConfig.EXPECT().GetString("session.path").Return("/").Once()
	mockConfig.EXPECT().GetString("session.domain").Return("example.com").Once()
	mockConfig.EXPECT().GetBool("session.secure").Return(true).Once()
	mockConfig.EXPECT().GetBool("session.http_only").Return(true).Once()
	mockConfig.EXPECT().GetString("session.same_site").Return("lax").Once()
	ConfigFacade = mockConfig

	mockSession := mockssession.NewSession(t)
	mockSession.EXPECT().GetName().Return("goravel_session").Once()
	mockSession.EXPECT().GetID().Return("session-id").Once()

	mockResponse := mockshttp.NewContextResponse(t)
	mockResponse.EXPECT().Cookie(http.Cookie{
		Name:     "goravel_session",
		Value:    "session-id",
		Expires:  now.Copy().AddMinutes(120).StdTime(),
		Path:     "/",
		Domain:   "example.com",
		Secure:   true,
		HttpOnly: true,
		SameSite: "lax",
	}).Return(mockResponse).Once()

	mockContext := mockshttp.NewContext(t)
	mockContext.EXPECT().Response().Return(mockResponse).Once()

	WriteCookie(mockContext, mockSession)
}
