package console

type Stubs struct {
}

func (s Stubs) Mail() string {
	return `package DummyPackage

import (
	"github.com/goravel/framework/contracts/mail"
)

type DummyMail struct {
}

// Attachments attach files to the mail
func (receiver *DummyMail) Attachments() []string{
	return []string{}
}

// Content set the content of the mail
func (receiver *DummyMail) Content() *mail.Content {
	return &mail.Content{}
}

// Envelope set the envelope of the mail
func (receiver *DummyMail) Envelope() *mail.Envelope {
    return &mail.Envelope{}
}

// Queue set the queue of the mail
func (receiver *DummyMail) Queue() *mail.Queue {
    return &mail.Queue{}
}
`
}
