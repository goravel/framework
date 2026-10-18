package mail

import "strings"

type Fake interface {
	// AssertNotQueued reports whether no queued mail matches the assertion.
	AssertNotQueued(assertion func(message Message) bool) bool
	// AssertNotSent reports whether no sent mail matches the assertion.
	AssertNotSent(assertion func(message Message) bool) bool
	// AssertNothingOutgoing reports whether no mail was sent or queued.
	AssertNothingOutgoing() bool
	// AssertNothingQueued reports whether no mail was queued.
	AssertNothingQueued() bool
	// AssertNothingSent reports whether no mail was sent.
	AssertNothingSent() bool
	// AssertQueued reports whether at least one queued mail matches the assertion.
	AssertQueued(assertion func(message Message) bool) bool
	// AssertQueuedCount reports whether exactly count mails were queued.
	AssertQueuedCount(count int) bool
	// AssertSent reports whether at least one sent mail matches the assertion.
	AssertSent(assertion func(message Message) bool) bool
	// AssertSentCount reports whether exactly count mails were sent.
	AssertSentCount(count int) bool
	// Queued returns the queued mails, only those matching the filter if given.
	Queued(filter ...func(message Message) bool) []Message
	// Reset stops faking: mails are delivered again and the records are cleared.
	Reset()
	// Sent returns the sent mails, only those matching the filter if given.
	Sent(filter ...func(message Message) bool) []Message
}

// Message is a mail recorded by the fake, as it would have been delivered:
// the Mailable applied and the views rendered.
type Message struct {
	// Mailable is the mailable passed to Send or Queue, nil when only the
	// builder methods were used.
	Mailable    Mailable
	Subject     string
	Html        string
	Text        string
	From        Address
	To          []string
	Cc          []string
	Bcc         []string
	Attachments []string
	Headers     map[string]string
	// Queue is the connection and queue requested by the mailable, for
	// queued mails.
	Queue *Queue
}

// HasTo reports whether the address is one of the recipients.
func (r Message) HasTo(address string) bool {
	return containsAddress(r.To, address)
}

// HasCc reports whether the address is one of the carbon copies.
func (r Message) HasCc(address string) bool {
	return containsAddress(r.Cc, address)
}

// HasBcc reports whether the address is one of the blind carbon copies.
func (r Message) HasBcc(address string) bool {
	return containsAddress(r.Bcc, address)
}

// HasAttachment reports whether the file is attached.
func (r Message) HasAttachment(file string) bool {
	for _, attachment := range r.Attachments {
		if attachment == file {
			return true
		}
	}

	return false
}

func containsAddress(addresses []string, address string) bool {
	for _, item := range addresses {
		if strings.EqualFold(strings.TrimSpace(item), strings.TrimSpace(address)) {
			return true
		}
	}

	return false
}
