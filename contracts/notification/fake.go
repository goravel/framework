package notification

// Fake records the notifications sent while faking instead of delivering
// them through their channels.
//
// The notification parameter of the methods below is the notification type
// to look for, given as an instance of it: a nil pointer such as
// (*InvoicePaid)(nil) is enough.
//
// A notifiable matches when it has the same type and, if it implements
// DatabaseRoutable and its route (its primary key) is not empty, the same
// route, so a model loaded again from the database still matches. Other
// notifiables, an unsaved model included, are compared with
// reflect.DeepEqual.
type Fake interface {
	// AssertCount reports whether the given number of notifications was
	// sent, to any notifiable.
	AssertCount(count int) bool
	// AssertNothingSent reports whether no notification was sent.
	AssertNothingSent() bool
	// AssertNothingSentTo reports whether no notification was sent to the
	// notifiable.
	AssertNothingSentTo(notifiable Notifiable) bool
	// AssertNotSentTo reports whether the notification was not sent to the
	// notifiable, and when assertions are given, whether none of those sends
	// matches all of them.
	AssertNotSentTo(notifiable Notifiable, notification Notification, assertion ...func(sent Sent) bool) bool
	// AssertSentOnDemand reports whether the notification was sent to an
	// on-demand notifiable created by Route, and when assertions are given,
	// whether one of those sends matches all of them.
	AssertSentOnDemand(notification Notification, assertion ...func(sent Sent) bool) bool
	// AssertSentTimes reports whether the notification was sent the given
	// number of times, to any notifiable.
	AssertSentTimes(notification Notification, times int) bool
	// AssertSentTo reports whether the notification was sent to the
	// notifiable, and when assertions are given, whether one of those sends
	// matches all of them.
	AssertSentTo(notifiable Notifiable, notification Notification, assertion ...func(sent Sent) bool) bool
	// Sent returns the sends of the notification, only those matching every
	// given filter.
	Sent(notification Notification, filter ...func(sent Sent) bool) []Sent
	// Reset stops faking: notifications are delivered again and the records
	// are cleared.
	Reset()
}

// Sent is a notification recorded by the fake.
type Sent interface {
	// Channels returns a copy of the channels the notification would have
	// been sent through: the ones returned by Via, without those ShouldSend
	// rejects.
	Channels() []string
	// HasChannel reports whether the notification would have been sent
	// through the channel.
	HasChannel(channel string) bool
	// Notifiable returns the notifiable passed to Send or SendNow, or the
	// on-demand notifiable created by Route.
	Notifiable() Notifiable
	// Notification returns the notification passed to Send or SendNow.
	Notification() Notification
	// Queued reports whether the notification would have been queued: it
	// implements ShouldQueue, was sent with Send rather than SendNow, and
	// the manager has a queue.
	Queued() bool
}
