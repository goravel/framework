package channels

import (
	"fmt"
	contractsmail "github.com/goravel/framework/contracts/mail"
	"github.com/goravel/framework/contracts/notification"
	"github.com/goravel/framework/mail"
	"github.com/goravel/framework/notification/utils"
)

// MailChannel 默认邮件通道
type MailChannel struct {
	mail contractsmail.Mail
}

func (c *MailChannel) Send(notifiable notification.Notifiable, notif interface{}) error {
    data, err := utils.CallToMethod(notif, "ToMail", notifiable)
    if err != nil {
        return fmt.Errorf("[MailChannel] %s", err.Error())
    }
    params := notifiable.NotificationParams()
    var email string
    if v, ok := params["mail"]; ok {
        if s, ok := v.(string); ok {
            email = s
        }
    }
    if email == "" {
        if v, ok := params["email"]; ok {
            if s, ok := v.(string); ok {
                email = s
            }
        }
    }
    if email == "" {
        return fmt.Errorf("[MailChannel] notifiable has no mail")
    }

	content := data["content"].(string)
	subject := data["subject"].(string)

	if err := c.mail.To([]string{email}).
		Content(mail.Html(content)).
		Subject(subject).Send(); err != nil {
		return err
	}

	return nil
}

func (c *MailChannel) SetMail(mail contractsmail.Mail) {
	c.mail = mail
}
