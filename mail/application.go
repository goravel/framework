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
	config      config.Config
	queue       queuecontract.Queue
	template    mail.Template
	headers     map[string]string
	from        mail.Address
	html        string
	text        string
	subject     string
	attachments []string
	bcc         []string
	cc          []string
	to          []string
	clone       int

	viewPath string
	textPath string
	with     map[string]any
}

func NewApplication(config config.Config, queue queuecontract.Queue) *Application {
	return &Application{
		config: config,
		queue:  queue,
	}
}

func (r *Application) Attach(attachments []string) mail.Mail {
	instance := r.instance()
	instance.attachments = attachments

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
	instance.viewPath = content.View
	instance.textPath = content.Text
	instance.with = content.With

	return instance
}

func (r *Application) From(address mail.Address) mail.Mail {
	instance := r.instance()
	instance.from = address

	return instance
}

func (r *Application) Headers(headers map[string]string) mail.Mail {
	instance := r.instance()
	instance.headers = headers

	return instance
}

func (r *Application) Queue(mailable ...mail.Mailable) error {
	if len(mailable) > 0 {
		r.setUsingMailable(mailable[0])
	}

	if err := r.renderViewTemplate(); err != nil {
		return err
	}

	job := r.queue.Job(NewSendMailJob(r.config), []queuecontract.Arg{
		{
			Type:  "string",
			Value: r.subject,
		},
		{
			Type:  "string",
			Value: r.html,
		},
		{
			Type:  "string",
			Value: r.from.Address,
		},
		{
			Type:  "string",
			Value: r.from.Name,
		},
		{
			Type:  "[]string",
			Value: r.to,
		},
		{
			Type:  "[]string",
			Value: r.cc,
		},
		{
			Type:  "[]string",
			Value: r.bcc,
		},
		{
			Type:  "[]string",
			Value: r.attachments,
		},
		{
			Type:  "[]string",
			Value: convertMapHeadersToSlice(r.headers),
		},
		{
			Type:  "string",
			Value: r.text,
		},
	})

	if len(mailable) > 0 {
		if queue := mailable[0].Queue(); queue != nil {
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

	if err := r.renderViewTemplate(); err != nil {
		return err
	}

	return SendMail(r.config, r.subject, r.text, r.html, r.from.Address, r.from.Name, r.to, r.cc, r.bcc, r.attachments, r.headers)
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
			clone:    1,
			config:   r.config,
			queue:    r.queue,
			template: r.template,
		}
	}

	return r
}

func (r *Application) setUsingMailable(mailable mail.Mailable) {
	if content := mailable.Content(); content != nil {
		if content.Html != "" {
			r.html = content.Html
		}
		r.viewPath = content.View
		r.textPath = content.Text
		r.with = content.With
	}

	if attachments := mailable.Attachments(); len(attachments) > 0 {
		r.attachments = attachments
	}

	if headers := mailable.Headers(); len(headers) > 0 {
		r.headers = headers
	}

	if envelope := mailable.Envelope(); envelope != nil {
		if envelope.From.Address != "" {
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
		if envelope.Subject != "" {
			r.subject = envelope.Subject
		}
	}
}

func (r *Application) renderViewTemplate() error {
	if r.viewPath != "" && r.template != nil {
		renderedHtml, err := r.template.Render(r.viewPath, r.with)
		if err != nil {
			return err
		}
		r.html = renderedHtml
	}

	if r.textPath != "" && r.template != nil {
		renderedText, err := r.template.Render(r.textPath, r.with)
		if err != nil {
			return err
		}
		r.text = renderedText
	}

	return nil
}

func SendMail(config config.Config, subject, text, html, fromAddress, fromName string, to, cc, bcc, attaches []string, headers map[string]string) error {
	e := NewEmail()
	if fromAddress == "" {
		e.From = fmt.Sprintf("%s <%s>", config.GetString("mail.from.name"), config.GetString("mail.from.address"))
	} else {
		e.From = fmt.Sprintf("%s <%s>", fromName, fromAddress)
	}

	e.To = to
	if len(bcc) > 0 {
		e.Bcc = bcc
	}
	if len(cc) > 0 {
		e.Cc = cc
	}
	e.Subject = subject

	if len(html) > 0 {
		e.HTML = []byte(html)
	}

	if len(text) > 0 {
		e.Text = []byte(text)
	}

	for _, attach := range attaches {
		if _, err := e.AttachFile(attach); err != nil {
			return err
		}
	}

	for key, val := range headers {
		e.Headers.Add(key, val)
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

func (a *loginAuth) Start(*smtp.ServerInfo) (string, []byte, error) {
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
