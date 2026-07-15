package mailmessage

import contractsnotification "github.com/goravel/framework/contracts/notification"

// NewMailMessage returns a fluent builder for MailMessage. Every method
// returns the same builder so calls can be chained; call Build() to
// finish.
//
//	func (n *InvoicePaid) ToMail(_ notification.Notifiable) notification.MailMessage {
//	    return mailmessage.NewMailMessage().
//	        Subject("Invoice Paid").
//	        Html("<p>Thanks!</p>").
//	        Build()
//	}
func NewMailMessage() *Builder {
	return &Builder{}
}

// Builder is the fluent builder returned by NewMailMessage.
type Builder struct {
	msg contractsnotification.MailMessage
}

func (b *Builder) Subject(subject string) *Builder {
	b.msg.Subject = subject
	return b
}

func (b *Builder) To(addresses ...string) *Builder {
	b.msg.To = addresses
	return b
}

func (b *Builder) From(address string) *Builder {
	b.msg.From = address
	return b
}

func (b *Builder) ReplyTo(address string) *Builder {
	b.msg.ReplyTo = address
	return b
}

func (b *Builder) Html(html string) *Builder {
	b.msg.Content.Html = html
	return b
}

func (b *Builder) Text(text string) *Builder {
	b.msg.Content.Text = text
	return b
}

func (b *Builder) HtmlView(view string, with map[string]any) *Builder {
	b.msg.Content.HtmlView = view
	b.msg.Content.With = with
	return b
}

func (b *Builder) TextView(view string, with map[string]any) *Builder {
	b.msg.Content.TextView = view
	b.msg.Content.With = with
	return b
}

func (b *Builder) Attach(paths ...string) *Builder {
	b.msg.Attachments = append(b.msg.Attachments, paths...)
	return b
}

func (b *Builder) Header(key, value string) *Builder {
	if b.msg.Headers == nil {
		b.msg.Headers = map[string]string{}
	}
	b.msg.Headers[key] = value
	return b
}

// Build returns the finished MailMessage.
func (b *Builder) Build() contractsnotification.MailMessage {
	return b.msg
}
