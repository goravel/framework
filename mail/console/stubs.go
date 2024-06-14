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

// Envelope 
func (receiver *DummyMail) Envelope() *mail.Envelope {
    return &mail.Envelope{}
}

// Attachments
func (receiver *DummyMail) Attachments() []string{
	return []string{}
}

// Content
func (receiver *DummyMail) Content() *Content {
	return &mail.Content{}
}
`
}
