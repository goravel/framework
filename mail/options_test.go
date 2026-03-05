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
	t.Run("defaults", func(t *testing.T) {
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
	})

	t.Run("set queue values", func(t *testing.T) {
		mail := Queue()
		returned := mail.OnConnection("redis").OnQueue("emails")
		assert.Same(t, mail, returned)
		assert.Equal(t, "redis", mail.Queue().Connection)
		assert.Equal(t, "emails", mail.Queue().Queue)
	})

	t.Run("last value wins", func(t *testing.T) {
		mail := Queue()
		mail.OnConnection("redis").OnConnection("sync")
		mail.OnQueue("emails").OnQueue("default")

		assert.Equal(t, "sync", mail.Queue().Connection)
		assert.Equal(t, "default", mail.Queue().Queue)
	})

	t.Run("empty values supported", func(t *testing.T) {
		mail := Queue()
		mail.OnConnection("redis").OnQueue("emails")
		mail.OnConnection("").OnQueue("")

		assert.Equal(t, "", mail.Queue().Connection)
		assert.Equal(t, "", mail.Queue().Queue)
	})
}
