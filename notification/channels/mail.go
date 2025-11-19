package channels

import (
    "fmt"
    contractsmail "github.com/goravel/framework/contracts/mail"
    "github.com/goravel/framework/contracts/notification"
    "github.com/goravel/framework/mail"
    "github.com/goravel/framework/notification/utils"
)

// MailChannel is the default mail delivery channel.
type MailChannel struct {
    mail contractsmail.Mail
}

// Send delivers a notification via email using the notifiable's params.
// It expects the notification to implement a ToMail(notifiable) method or PayloadProvider.
func (c *MailChannel) Send(notifiable notification.Notifiable, notif interface{}) error {
    data, err := utils.CallToMethod(notif, "ToMail", notifiable)
    if err != nil {
        return fmt.Errorf("[MailChannel] %s", err.Error())
    }
    params := notifiable.NotificationParams()
    email := getEmail(params)
    if email == "" {
        return fmt.Errorf("[MailChannel] notifiable has no mail")
    }

    contentVal, ok := data["content"]
    if !ok {
        return fmt.Errorf("[MailChannel] content not provided")
    }
    subjectVal, ok := data["subject"]
    if !ok {
        return fmt.Errorf("[MailChannel] subject not provided")
    }
    content, _ := contentVal.(string)
    subject, _ := subjectVal.(string)
    if content == "" || subject == "" {
        return fmt.Errorf("[MailChannel] invalid content or subject")
    }

	if err := c.mail.To([]string{email}).
		Content(mail.Html(content)).
		Subject(subject).Send(); err != nil {
		return err
	}

    return nil
}

// SetMail injects the mail facade into the channel.
func (c *MailChannel) SetMail(mail contractsmail.Mail) {
    c.mail = mail
}

func getEmail(params map[string]interface{}) string {
    if v, ok := params["mail"]; ok {
        if s, ok := v.(string); ok && s != "" {
            return s
        }
    }
    if v, ok := params["email"]; ok {
        if s, ok := v.(string); ok && s != "" {
            return s
        }
    }
    return ""
}
