package mail

import (
	"crypto/tls"
	"fmt"
	"net/smtp"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/mail"
	queuecontract "github.com/goravel/framework/contracts/queue"
)

type Application struct {
	attaches []string
	bcc      []string
	cc       []string
	clone    int
	config   config.Config
	from     mail.From
	html     string
	queue    queuecontract.Queue
	subject  string
	to       []string
}

func NewApplication(config config.Config, queue queuecontract.Queue) *Application {
	return &Application{
		config: config,
		queue:  queue,
	}
}

func (r *Application) Attach(files []string) mail.Mail {
	instance := r.instance()
	instance.attaches = files

	return instance
}

func (r *Application) Bcc(bcc []string) mail.Mail {
	instance := r.instance()
	instance.bcc = bcc

	return instance
}

func (r *Application) Cc(cc []string) mail.Mail {
	instance := r.instance()
	instance.cc = cc

	return instance
}

func (r *Application) Content(content mail.Content) mail.Mail {
	instance := r.instance()
	instance.html = content.Html

	return instance
}

func (r *Application) From(from mail.From) mail.Mail {
	instance := r.instance()
	instance.from = from

	return instance
}

func (r *Application) Queue(queueMail ...mail.ShouldQueue) error {

	if len(queueMail) > 0 {
		r.setUsingMailable(queueMail[0])
	}

	job := r.queue.Job(NewSendMailJob(r.config), []queuecontract.Arg{
		{Value: r.subject, Type: "string"},
		{Value: r.html, Type: "string"},
		{Value: r.from.Address, Type: "string"},
		{Value: r.from.Name, Type: "string"},
		{Value: r.to, Type: "[]string"},
		{Value: r.cc, Type: "[]string"},
		{Value: r.bcc, Type: "[]string"},
		{Value: r.attaches, Type: "[]string"},
	})

	if len(queueMail) > 0 {
		if queue := queueMail[0].Queue(); queue != nil {
			if queue.Connection != "" {
				job.OnConnection(queue.Connection)
			}
			if queue.Queue != "" {
				job.OnQueue(queue.Queue)
			}
		}
	}

	return job.Dispatch()
}

func (r *Application) Send(mailable ...mail.Mailable) error {
	if len(mailable) > 0 {
		r.setUsingMailable(mailable[0])
	}
	return SendMail(r.config, r.subject, r.html, r.from.Address, r.from.Name, r.to, r.cc, r.bcc, r.attaches)
}

func (r *Application) Subject(subject string) mail.Mail {
	instance := r.instance()
	instance.subject = subject

	return instance
}

func (r *Application) To(to []string) mail.Mail {
	instance := r.instance()
	instance.to = to

	return instance
}

func (r *Application) instance() *Application {
	if r.clone == 0 {
		return &Application{
			clone:  1,
			config: r.config,
			queue:  r.queue,
		}
	}

	return r
}

func (r *Application) setUsingMailable(mailable mail.Mailable) {
	content := mailable.Content()
	if content != nil {
		if content.Html != "" {
			r.html = content.Html
		}
	}

	envelope := mailable.Envelope()
	if envelope != nil {
		if envelope.From != (mail.From{}) {
			r.from = envelope.From
		}

		if len(envelope.To) > 0 {
			r.to = envelope.To
		}

		if len(envelope.Cc) > 0 {
			r.cc = envelope.Cc
		}

		if len(envelope.Bcc) > 0 {
			r.bcc = envelope.Bcc
		}

		if len(mailable.Attachments()) > 0 {
			r.attaches = mailable.Attachments()
		}

		if envelope.Subject != "" {
			r.subject = envelope.Subject
		}
	}
}

func SendMail(config config.Config, subject, html, fromAddress, fromName string, to, cc, bcc, attaches []string) error {
	e := NewEmail()
	if fromAddress == "" {
		e.From = fmt.Sprintf("%s <%s>", config.GetString("mail.from.name"), config.GetString("mail.from.address"))
	} else {
		e.From = fmt.Sprintf("%s <%s>", fromName, fromAddress)
	}

	e.To = to
	if len(e.Bcc) > 0 {
		e.Bcc = bcc
	}
	if len(e.Cc) > 0 {
		e.Cc = cc
	}
	e.Subject = subject
	e.HTML = []byte(html)

	for _, attach := range attaches {
		if _, err := e.AttachFile(attach); err != nil {
			return err
		}
	}

	port := config.GetInt("mail.port")
	switch port {
	case 465:
		return e.SendWithTLS(fmt.Sprintf("%s:%d", config.GetString("mail.host"), config.GetInt("mail.port")),
			LoginAuth(config.GetString("mail.username"), config.GetString("mail.password")),
			&tls.Config{ServerName: config.GetString("mail.host")})
	case 587:
		return e.SendWithStartTLS(fmt.Sprintf("%s:%d", config.GetString("mail.host"), config.GetInt("mail.port")),
			LoginAuth(config.GetString("mail.username"), config.GetString("mail.password")),
			&tls.Config{ServerName: config.GetString("mail.host")})
	default:
		return e.Send(fmt.Sprintf("%s:%d", config.GetString("mail.host"), port),
			LoginAuth(config.GetString("mail.username"), config.GetString("mail.password")))
	}
}

type loginAuth struct {
	username, password string
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte(a.username), nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		}
	}
	return nil, nil
}
