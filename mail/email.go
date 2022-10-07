package mail

import (
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"

	"github.com/goravel/framework/contracts/mail"
	contractqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/facades"
)

type Email struct {
	content  mail.Content
	from     mail.From
	to       []string
	cc       []string
	bcc      []string
	attaches []string
}

func NewEmail() mail.Mail {
	return &Email{}
}

func (r *Email) Content(content mail.Content) mail.Mail {
	r.content = content

	return r
}

func (r *Email) From(from mail.From) mail.Mail {
	r.from = from

	return r
}

func (r *Email) To(addresses []string) mail.Mail {
	r.to = addresses

	return r
}

func (r *Email) Cc(addresses []string) mail.Mail {
	r.cc = addresses

	return r
}

func (r *Email) Bcc(addresses []string) mail.Mail {
	r.bcc = addresses

	return r
}

func (r *Email) Attach(files []string) mail.Mail {
	//todo Test multi file
	r.attaches = files

	return r
}

func (r *Email) Send() error {
	return SendMail(r.content.Subject, r.content.Html, r.from.Address, r.from.Name, r.to, r.cc, r.bcc, r.attaches)
}

func (r *Email) Queue(queue *mail.Queue) error {
	job := facades.Queue.Job(&SendMailJob{}, []contractqueue.Arg{
		{Value: r.content.Subject, Type: "string"},
		{Value: r.content.Html, Type: "string"},
		{Value: r.from.Address, Type: "string"},
		{Value: r.from.Name, Type: "string"},
		{Value: r.to, Type: "[]string"},
		{Value: r.cc, Type: "[]string"},
		{Value: r.bcc, Type: "[]string"},
		{Value: r.attaches, Type: "[]string"},
	})
	if queue != nil {
		if queue.Connection != "" {
			job.OnConnection(queue.Connection)
		}
		if queue.Queue != "" {
			job.OnQueue(queue.Queue)
		}
	}

	return job.Dispatch()
}

func SendMail(subject, html string, fromAddress, fromName string, to, cc, bcc, attaches []string) error {
	e := email.NewEmail()
	if fromAddress == "" {
		e.From = fmt.Sprintf("%s <%s>", facades.Config.GetString("mail.from.name"), facades.Config.GetString("mail.from.address"))
	} else {
		e.From = fmt.Sprintf("%s <%s>", fromName, fromAddress)
	}

	e.To = to
	e.Bcc = bcc
	e.Cc = cc
	e.Subject = subject
	e.HTML = []byte(html)

	for _, attach := range attaches {
		if _, err := e.AttachFile(attach); err != nil {
			return err
		}
	}

	return e.Send(fmt.Sprintf("%s:%s", facades.Config.GetString("mail.host"),
		facades.Config.GetString("mail.port")),
		smtp.PlainAuth("", facades.Config.GetString("mail.username"),
			facades.Config.GetString("mail.password"),
			facades.Config.GetString("mail.host")))
}
