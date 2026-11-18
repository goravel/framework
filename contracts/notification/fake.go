package notification

import "slices"

// Fake records the notifications sent while faking instead of delivering
// them through their channels.
//
// The notification parameter of the methods below is the notification type
// to look for, given as an instance of it: a nil pointer such as
// (*InvoicePaid)(nil) is enough.
//
// A notifiable matches when it has the same type and, if it implements
// DatabaseRoutable, the same route (its primary key), so a model loaded
// again from the database still matches; other notifiables are compared
// with reflect.DeepEqual.
type Fake interface {
	// AssertCount reports whether exactly count notifications were sent, to
	// any notifiable.
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
	// AssertSentTimes reports whether the notification was sent exactly
	// times times, to any notifiable.
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
type Sent struct {
	Notifiable   Notifiable
	Notification Notification
	// Channels are the channels the notification would have been sent
	// through: the ones returned by Via, without those ShouldSend rejects.
	Channels []string
	// Queued reports whether the notification would have been queued: it
	// implements ShouldQueue and was sent with Send rather than SendNow.
	Queued bool
}

// HasChannel reports whether the notification would have been sent through
// the channel.
func (r Sent) HasChannel(channel string) bool {
	return slices.Contains(r.Channels, channel)
}
