package mail

type Mail interface {
	Content(content Content) Mail
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
