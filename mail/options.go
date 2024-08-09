package mail

import (
	"github.com/goravel/framework/contracts/mail"
)

func Address(address, name string) mail.Address {
	return mail.Address{
		Address: address,
		Name:    name,
	}
}

func Html(html string) mail.Content {
	return mail.Content{
		Html: html,
	}
}

type QueueMail struct {
	queue *mail.Queue
}

func Queue() *QueueMail {
	return &QueueMail{
		queue: &mail.Queue{},
	}
}

// Attachments attach files to the mail
func (receiver *QueueMail) Attachments() []string {
	return []string{}
}

// Content set the content of the mail
func (receiver *QueueMail) Content() *mail.Content {
	return &mail.Content{}
}

// Envelope set the envelope of the mail
func (receiver *QueueMail) Envelope() *mail.Envelope {
	return &mail.Envelope{}
}

// Queue set the queue of the mail
func (receiver *QueueMail) Queue() *mail.Queue {
	return receiver.queue
}

func (receiver *QueueMail) OnConnection(connection string) *QueueMail {
	receiver.queue.Connection = connection

	return receiver
}

func (receiver *QueueMail) OnQueue(queue string) *QueueMail {
	receiver.queue.Queue = queue

	return receiver
}
