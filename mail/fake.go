package mail

import (
	"maps"
	"slices"
	"sync"
	"sync/atomic"

	"github.com/goravel/framework/contracts/mail"
)

var _ mail.Fake = &Fake{}

// Fake records the mails sent or queued while faking. One instance is
// shared by every Application created by the service provider, since the
// Mail facade resolves a new Application on each call.
type Fake struct {
	mu      sync.RWMutex
	enabled atomic.Bool
	sent    []mail.Message
	queued  []mail.Message
}

func NewFake() *Fake {
	return &Fake{}
}

func (r *Fake) AssertNotQueued(assertion func(message mail.Message) bool) bool {
	return len(r.Queued(assertion)) == 0
}

func (r *Fake) AssertNotSent(assertion func(message mail.Message) bool) bool {
	return len(r.Sent(assertion)) == 0
}

func (r *Fake) AssertNothingOutgoing() bool {
	return r.AssertNothingSent() && r.AssertNothingQueued()
}

func (r *Fake) AssertNothingQueued() bool {
	return len(r.Queued()) == 0
}

func (r *Fake) AssertNothingSent() bool {
	return len(r.Sent()) == 0
}

func (r *Fake) AssertQueued(assertion func(message mail.Message) bool) bool {
	return len(r.Queued(assertion)) > 0
}

func (r *Fake) AssertQueuedCount(count int) bool {
	return len(r.Queued()) == count
}

func (r *Fake) AssertSent(assertion func(message mail.Message) bool) bool {
	return len(r.Sent(assertion)) > 0
}

func (r *Fake) AssertSentCount(count int) bool {
	return len(r.Sent()) == count
}

func (r *Fake) Queued(filter ...func(message mail.Message) bool) []mail.Message {
	r.mu.RLock()
	messages := slices.Clone(r.queued)
	r.mu.RUnlock()

	// Filters run without the lock, so they may call the fake again.
	return filterMessages(messages, filter)
}

func (r *Fake) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.enabled.Store(false)
	r.sent = nil
	r.queued = nil
}

func (r *Fake) Sent(filter ...func(message mail.Message) bool) []mail.Message {
	r.mu.RLock()
	messages := slices.Clone(r.sent)
	r.mu.RUnlock()

	// Filters run without the lock, so they may call the fake again.
	return filterMessages(messages, filter)
}

// enable starts faking. It is idempotent: while faking, the records are kept,
// so calling Fake again does not hide mails that were already recorded.
func (r *Fake) enable() {
	r.enabled.Store(true)
}

func (r *Fake) isEnabled() bool {
	return r != nil && r.enabled.Load()
}

func (r *Fake) recordSent(message mail.Message) {
	message = cloneMessage(message)

	r.mu.Lock()
	defer r.mu.Unlock()

	r.sent = append(r.sent, message)
}

func (r *Fake) recordQueued(message mail.Message) {
	message = cloneMessage(message)

	r.mu.Lock()
	defer r.mu.Unlock()

	r.queued = append(r.queued, message)
}

// cloneMessage copies the slices and the map of the message, so the records
// and the messages handed out never share them with the caller.
func cloneMessage(message mail.Message) mail.Message {
	message.To = slices.Clone(message.To)
	message.Cc = slices.Clone(message.Cc)
	message.Bcc = slices.Clone(message.Bcc)
	message.Attachments = slices.Clone(message.Attachments)
	message.Headers = maps.Clone(message.Headers)
	if message.Queue != nil {
		queue := *message.Queue
		message.Queue = &queue
	}

	return message
}

// filterMessages returns copies of the messages matching every non-nil filter.
func filterMessages(messages []mail.Message, filters []func(message mail.Message) bool) []mail.Message {
	result := make([]mail.Message, 0, len(messages))
	for _, message := range messages {
		message = cloneMessage(message)
		if matchesAll(message, filters) {
			result = append(result, message)
		}
	}

	return slices.Clip(result)
}

func matchesAll(message mail.Message, filters []func(message mail.Message) bool) bool {
	for _, filter := range filters {
		if filter != nil && !filter(message) {
			return false
		}
	}

	return true
}
