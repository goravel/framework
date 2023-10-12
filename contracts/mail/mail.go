package mail

//go:generate mockery --name=Mail
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
	// Send the Mail
	Send() error
	// Queue a given Mail
	Queue(queue *Queue) error
}

type Content struct {
	Subject string
	Html    string
}

type Queue struct {
	Connection string
	Queue      string
}

type From struct {
	Address string
	Name    string
}
