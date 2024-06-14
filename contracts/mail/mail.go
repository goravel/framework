package mail

type Mail interface {
	// Content set the content of Mail.
	Content(content Content) Mail
	// From set the sender of Mail.
	From(address From) Mail
	// To set the recipients of Mail.
	To(addresses []string) Mail
	// Cc adds a "carbon copy" address to the Mail.
	Cc(addresses []string) Mail
	// Bcc adds a "blind carbon copy" address to the Mail.
	Bcc(addresses []string) Mail
	// Attach attaches files to the Mail.
	Attach(files []string) Mail
	// Subject set the subject of Mail.
	Subject(subject string) Mail
	// Send the Mail
	Send(mailable ...Mailable) error
	// Queue a given Mail
	Queue(queue ...ShouldQueue) error
}

type Mailable interface {
	// Envelope set the envelope of Mailable.
	Envelope() *Envelope
	// Content set the content of Mailable.
	Content() *Content
	// Attachments set the attachments of Mailable.
	Attachments() []string
}

type ShouldQueue interface {
	Mailable
	// Queue set the queue of Mailable.
	Queue() *Queue
}

type Content struct {
	Html string
}

type Queue struct {
	Connection string
	Queue      string
}

type From struct {
	Address string
	Name    string
}

type Envelope struct {
	From     From
	To       []string
	Cc       []string
	Bcc      []string
	ReplyTo  []string
	Tags     []string
	Metadata map[string]any
	Subject  string
}
