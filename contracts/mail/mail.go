package mail

//go:generate mockery --name=Mail
type Mail interface {
	Content(content Content) Mail
	From(address From) Mail
	To(addresses []string) Mail
	Cc(addresses []string) Mail
	Bcc(addresses []string) Mail
	Attach(files []string) Mail
	Send() error
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
