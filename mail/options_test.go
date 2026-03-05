package mail

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddress(t *testing.T) {
	address := Address("mail@example.com", "Mailer")
	assert.Equal(t, "mail@example.com", address.Address)
	assert.Equal(t, "Mailer", address.Name)
}

func TestHtml(t *testing.T) {
	content := Html("<h1>Hello</h1>")
	assert.Equal(t, "<h1>Hello</h1>", content.Html)
}

func TestQueueMail(t *testing.T) {
	mail := Queue()
	assert.NotNil(t, mail.Queue())
	assert.Empty(t, mail.Attachments())
	assert.Empty(t, mail.Content().Html)
	assert.Empty(t, mail.Content().Text)
	assert.Empty(t, mail.Content().View)
	assert.Nil(t, mail.Content().With)
	assert.Equal(t, "", mail.Envelope().Subject)
	assert.Empty(t, mail.Envelope().To)
	assert.Empty(t, mail.Headers())

	returned := mail.OnConnection("redis").OnQueue("emails")
	assert.Same(t, mail, returned)
	assert.Equal(t, "redis", mail.Queue().Connection)
	assert.Equal(t, "emails", mail.Queue().Queue)
}
