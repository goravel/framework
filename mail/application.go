package mail

import (
	"crypto/tls"
	"fmt"
	"net/smtp"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/mail"
	queuecontract "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/mail/template"
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

// Params represents all parameters needed for sending mail
type Params struct {
	Subject     string            `json:"subject"`
	HTML        string            `json:"html"`
	Text        string            `json:"text"`
	FromAddress string            `json:"from_address"`
	FromName    string            `json:"from_name"`
	To          []string          `json:"to"`
	CC          []string          `json:"cc"`
	BCC         []string          `json:"bcc"`
	Attachments []string          `json:"attachments"`
	Headers     map[string]string `json:"headers"`
}

func NewApplication(config config.Config, queue queuecontract.Queue) *Application {
	templateEngine, _ := template.Get(config)

	return &Application{
		config:   config,
		queue:    queue,
		template: templateEngine,
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

	params := Params{
		Subject:     r.subject,
		HTML:        r.html,
		Text:        r.text,
		FromAddress: r.from.Address,
		FromName:    r.from.Name,
		To:          r.to,
		CC:          r.cc,
		BCC:         r.bcc,
		Attachments: r.attachments,
		Headers:     r.headers,
	}

	job := r.queue.Job(NewSendMailJob(r.config), []queuecontract.Arg{
		{
			Type:  "mail.Params",
			Value: params,
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

	params := Params{
		Subject:     r.subject,
		HTML:        r.html,
		Text:        r.text,
		FromAddress: r.from.Address,
		FromName:    r.from.Name,
		To:          r.to,
		CC:          r.cc,
		BCC:         r.bcc,
		Attachments: r.attachments,
		Headers:     r.headers,
	}

	return SendMail(r.config, params)
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

func SendMail(config config.Config, params Params) error {
	e := NewEmail()
	fromAddress, fromName := params.FromAddress, params.FromName
	if fromAddress == "" {
		fromName, fromAddress = config.GetString("mail.from.name"), config.GetString("mail.from.address")
	}

	e.From = fmt.Sprintf("%s <%s>", fromName, fromAddress)
	e.To = params.To
	if len(params.BCC) > 0 {
		e.Bcc = params.BCC
	}
	if len(params.CC) > 0 {
		e.Cc = params.CC
	}
	e.Subject = params.Subject

	if len(params.HTML) > 0 {
		e.HTML = []byte(params.HTML)
	}

	if len(params.Text) > 0 {
		e.Text = []byte(params.Text)
	}

	for _, attach := range params.Attachments {
		if _, err := e.AttachFile(attach); err != nil {
			return err
		}
	}

	for key, val := range params.Headers {
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
