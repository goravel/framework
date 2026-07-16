package notification

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	contractsnotification "github.com/goravel/framework/contracts/notification"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	mockslog "github.com/goravel/framework/mocks/log"
	mocksmail "github.com/goravel/framework/mocks/mail"
	mocksqueue "github.com/goravel/framework/mocks/queue"
	"github.com/goravel/framework/notification/channels"
	"github.com/goravel/framework/notification/mail"
)

// ---- Fakes ----

type fakeNotifiable struct{ email string }

func (f *fakeNotifiable) RouteNotificationFor(channel string) any {
	if channel == "mail" {
		return f.email
	}
	return ""
}

type fakeNotification struct {
	channels []string
}

func (f *fakeNotification) Via(_ contractsnotification.Notifiable) []string { return f.channels }

type fakeChannel struct {
	name    string
	sendErr error
	calls   int
}

func (c *fakeChannel) Name() string { return c.name }
func (c *fakeChannel) Send(_ contractsnotification.Notifiable, _ contractsnotification.Notification) error {
	c.calls++
	return c.sendErr
}

// shouldQueueNotification implements ShouldQueue with configurable
// channels, for exercising dispatchQueued's error paths.
type shouldQueueNotification struct {
	channels []string
}

func (n *shouldQueueNotification) Via(_ contractsnotification.Notifiable) []string { return n.channels }
func (n *shouldQueueNotification) OnQueue() string                                 { return "" }
func (n *shouldQueueNotification) OnConnection() string                            { return "" }

// fakeResolvableChannel implements ResolvableChannel with configurable
// Resolve/Deliver errors, for exercising dispatchQueued and DispatchJob
// without a real channel driver.
type fakeResolvableChannel struct {
	name       string
	resolveErr error
	deliverErr error
	delivered  []deliveredCall
}

type deliveredCall struct {
	route   string
	payload []byte
}

func (c *fakeResolvableChannel) Name() string { return c.name }
func (c *fakeResolvableChannel) Send(_ contractsnotification.Notifiable, _ contractsnotification.Notification) error {
	return nil
}
func (c *fakeResolvableChannel) Resolve(_ contractsnotification.Notifiable, _ contractsnotification.Notification) (string, []byte, error) {
	if c.resolveErr != nil {
		return "", nil, c.resolveErr
	}
	return "route", []byte("{}"), nil
}
func (c *fakeResolvableChannel) Deliver(route string, payload []byte) error {
	c.delivered = append(c.delivered, deliveredCall{route: route, payload: payload})
	return c.deliverErr
}

var _ contractsnotification.ResolvableChannel = (*fakeResolvableChannel)(nil)

// shouldSendNotification implements NotificationWithShouldSend, vetoing
// delivery to specific channels.
type shouldSendNotification struct {
	channels []string
	skip     map[string]bool
}

func (n *shouldSendNotification) Via(_ contractsnotification.Notifiable) []string { return n.channels }
func (n *shouldSendNotification) ShouldSend(_ contractsnotification.Notifiable, channel string) bool {
	return !n.skip[channel]
}

// afterSendingNotification implements NotificationWithAfterSending,
// recording which channels it was called for and optionally erroring.
type afterSendingNotification struct {
	channels []string
	called   []string
	err      error
}

func (n *afterSendingNotification) Via(_ contractsnotification.Notifiable) []string {
	return n.channels
}
func (n *afterSendingNotification) AfterSending(_ contractsnotification.Notifiable, channel string) error {
	n.called = append(n.called, channel)
	return n.err
}

// queueableShouldSendNotification combines ShouldQueue + ShouldSend for
// TestManager_Send_QueuedNotification_SkipsChannel_WhenShouldSendReturnsFalse.
type queueableShouldSendNotification struct {
	channels []string
	skip     map[string]bool
}

func (n *queueableShouldSendNotification) Via(_ contractsnotification.Notifiable) []string {
	return n.channels
}
func (n *queueableShouldSendNotification) OnQueue() string      { return "" }
func (n *queueableShouldSendNotification) OnConnection() string { return "" }
func (n *queueableShouldSendNotification) ShouldSend(_ contractsnotification.Notifiable, channel string) bool {
	return !n.skip[channel]
}

type queueTestNotifiable struct{}

func (queueTestNotifiable) RouteNotificationFor(channel string) any {
	if channel == "mail" {
		return "user@example.com"
	}
	return ""
}

type queueableNotification struct{}

func (n *queueableNotification) Via(_ contractsnotification.Notifiable) []string {
	return []string{"mail"}
}
func (n *queueableNotification) ToMail(_ contractsnotification.Notifiable) contractsnotification.MailMessage {
	return mail.NewMessage().
		Subject("Queued").
		Html("<p>hi</p>").
		Build()
}
func (n *queueableNotification) OnQueue() string      { return "" }
func (n *queueableNotification) OnConnection() string { return "" }

// queueableNotificationWithRouting exercises OnConnection()/OnQueue()
// actually being invoked and their values passed through.
type queueableNotificationWithRouting struct{ queueableNotification }

func (n *queueableNotificationWithRouting) OnQueue() string      { return "notifications" }
func (n *queueableNotificationWithRouting) OnConnection() string { return "redis" }

// ---- Manager: SendNow / dispatchSync ----

func TestManager_SendNow_CallsCorrectChannels(t *testing.T) {
	logger := mockslog.NewLog(t)

	mgr := NewManager(logger, nil)

	chA := &fakeChannel{name: "a"}
	chB := &fakeChannel{name: "b"}
	mgr.Extend(chA)
	mgr.Extend(chB)

	n := &fakeNotification{channels: []string{"a", "b"}}
	notifiable := &fakeNotifiable{email: "user@example.com"}

	err := mgr.SendNow(notifiable, n)
	assert.NoError(t, err)
	assert.Equal(t, 1, chA.calls)
	assert.Equal(t, 1, chB.calls)
}

func TestManager_SendNow_SkipsUnregisteredChannel(t *testing.T) {
	logger := mockslog.NewLog(t)
	logger.EXPECT().Errorf("%s", mock.MatchedBy(func(s string) bool {
		return strings.Contains(s, "nonexistent")
	})).Once()

	mgr := NewManager(logger, nil)

	n := &fakeNotification{channels: []string{"nonexistent"}}
	notifiable := &fakeNotifiable{}

	err := mgr.SendNow(notifiable, n)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestManager_SendNow_LogsChannelError_ContinuesOthers(t *testing.T) {
	logger := mockslog.NewLog(t)
	logger.EXPECT().Errorf(
		"notifications: channel %q failed for %T: %v",
		"fail", mock.AnythingOfType("*notification.fakeNotification"), errors.New("smtp down"),
	).Once()

	mgr := NewManager(logger, nil)

	chFail := &fakeChannel{name: "fail", sendErr: errors.New("smtp down")}
	chOK := &fakeChannel{name: "ok"}
	mgr.Extend(chFail)
	mgr.Extend(chOK)

	n := &fakeNotification{channels: []string{"fail", "ok"}}
	notifiable := &fakeNotifiable{}

	err := mgr.SendNow(notifiable, n)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "smtp down")

	// The "ok" channel must still be called even though "fail" errored.
	assert.Equal(t, 1, chOK.calls)
}

func TestManager_SendNow_NoChannels_Warns(t *testing.T) {
	logger := mockslog.NewLog(t)
	logger.EXPECT().Errorf(
		"notifications: %T.Via() returned no channels for %T — nothing sent",
		mock.AnythingOfType("*notification.fakeNotification"), mock.AnythingOfType("*notification.fakeNotifiable"),
	).Once()

	mgr := NewManager(logger, nil)

	n := &fakeNotification{channels: []string{}}
	err := mgr.SendNow(&fakeNotifiable{}, n)
	assert.NoError(t, err)
}

func TestManager_SendNow_SkipsChannel_WhenShouldSendReturnsFalse(t *testing.T) {
	logger := mockslog.NewLog(t)
	mgr := NewManager(logger, nil)

	ch := &fakeChannel{name: "a"}
	mgr.Extend(ch)

	n := &shouldSendNotification{channels: []string{"a"}, skip: map[string]bool{"a": true}}
	err := mgr.SendNow(&fakeNotifiable{}, n)

	assert.NoError(t, err)
	assert.Equal(t, 0, ch.calls, "ShouldSend returning false should skip the channel entirely")
}

func TestManager_SendNow_CallsAfterSending_OnSuccess(t *testing.T) {
	logger := mockslog.NewLog(t)
	mgr := NewManager(logger, nil)

	ch := &fakeChannel{name: "a"}
	mgr.Extend(ch)

	n := &afterSendingNotification{channels: []string{"a"}}
	err := mgr.SendNow(&fakeNotifiable{}, n)

	assert.NoError(t, err)
	assert.Equal(t, []string{"a"}, n.called)
}

func TestManager_SendNow_LogsAndJoinsError_WhenAfterSendingFails(t *testing.T) {
	logger := mockslog.NewLog(t)
	logger.EXPECT().Errorf(
		"notifications: AfterSending hook failed for channel %q, %T: %v",
		"a", mock.AnythingOfType("*notification.afterSendingNotification"), errors.New("webhook failed"),
	).Once()

	mgr := NewManager(logger, nil)
	ch := &fakeChannel{name: "a"}
	mgr.Extend(ch)

	n := &afterSendingNotification{channels: []string{"a"}, err: errors.New("webhook failed")}
	err := mgr.SendNow(&fakeNotifiable{}, n)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "webhook failed")
}

func TestManager_Channel_ReturnsNil_WhenNotRegistered(t *testing.T) {
	logger := mockslog.NewLog(t)
	logger.EXPECT().Errorf("%s", mock.MatchedBy(func(s string) bool {
		return strings.Contains(s, "missing")
	})).Once()
	mgr := NewManager(logger, nil)

	got := mgr.Channel("missing")
	assert.Nil(t, got)
}

func TestManager_Extend_RegistersChannel(t *testing.T) {
	logger := mockslog.NewLog(t)
	mgr := NewManager(logger, nil)

	ch := &fakeChannel{name: "custom"}
	mgr.Extend(ch)

	got := mgr.Channel("custom")
	assert.Equal(t, ch, got)
}

// ---- Manager.Send: routing to sync vs. queued ----

func TestManager_Send_NonQueueableNotification_UsesDispatchSync(t *testing.T) {
	logger := mockslog.NewLog(t)

	mgr := NewManager(logger, nil)
	ch := &fakeChannel{name: "a"}
	mgr.Extend(ch)

	n := &fakeNotification{channels: []string{"a"}}
	err := mgr.Send(&fakeNotifiable{}, n)

	assert.NoError(t, err)
	assert.Equal(t, 1, ch.calls)
}

// ---- dispatchQueued error paths (only reachable via Send() + ShouldQueue + a configured queue) ----

func TestManager_Send_QueuedNotification_UnregisteredChannel_ReturnsError(t *testing.T) {
	logger := mockslog.NewLog(t)
	logger.EXPECT().Errorf("%s", mock.MatchedBy(func(s string) bool {
		return strings.Contains(s, "missing")
	})).Once()

	q := mocksqueue.NewQueue(t) // no calls expected — should fail before reaching the queue
	mgr := NewManager(logger, q)

	n := &shouldQueueNotification{channels: []string{"missing"}}
	err := mgr.Send(&fakeNotifiable{}, n)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing")
}

func TestManager_Send_QueuedNotification_ChannelNotResolvable_ReturnsError(t *testing.T) {
	logger := mockslog.NewLog(t)
	logger.EXPECT().Errorf("notifications: %v", mock.MatchedBy(func(err error) bool {
		return strings.Contains(err.Error(), "does not support queued dispatch")
	})).Once()

	mgr := NewManager(logger, mocksqueue.NewQueue(t))
	mgr.Extend(&fakeChannel{name: "plain"}) // implements Channel only, not ResolvableChannel

	n := &shouldQueueNotification{channels: []string{"plain"}}
	err := mgr.Send(&fakeNotifiable{}, n)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not support queued dispatch")
}

func TestManager_Send_QueuedNotification_ResolveError_ReturnsError(t *testing.T) {
	logger := mockslog.NewLog(t)
	logger.EXPECT().Errorf(
		"notifications: failed to resolve channel %q for %T: %v",
		"broken", mock.AnythingOfType("*notification.shouldQueueNotification"), errors.New("boom"),
	).Once()

	mgr := NewManager(logger, mocksqueue.NewQueue(t))
	mgr.Extend(&fakeResolvableChannel{name: "broken", resolveErr: errors.New("boom")})

	n := &shouldQueueNotification{channels: []string{"broken"}}
	err := mgr.Send(&fakeNotifiable{}, n)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
}

func TestManager_Send_QueuedNotification_DispatchError_ReturnsError(t *testing.T) {
	logger := mockslog.NewLog(t)

	q := mocksqueue.NewQueue(t)
	pending := mocksqueue.NewPendingJob(t)
	q.EXPECT().Job(mock.AnythingOfType("*notification.DispatchJob"), mock.Anything).Return(pending).Once()
	pending.EXPECT().Dispatch().Return(errors.New("queue connection refused")).Once()

	mgr := NewManager(logger, q)
	mgr.Extend(&fakeResolvableChannel{name: "a"})

	n := &shouldQueueNotification{channels: []string{"a"}}
	err := mgr.Send(&fakeNotifiable{}, n)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "queue connection refused")
}

func TestManager_Send_QueuedNotification_SkipsChannel_WhenShouldSendReturnsFalse(t *testing.T) {
	logger := mockslog.NewLog(t)
	q := mocksqueue.NewQueue(t) // no Job() call expected

	mgr := NewManager(logger, q)
	mgr.Extend(&fakeResolvableChannel{name: "a"})

	n := &queueableShouldSendNotification{channels: []string{"a"}, skip: map[string]bool{"a": true}}
	err := mgr.Send(&fakeNotifiable{}, n)
	assert.NoError(t, err)
}

// ---- Route (on-demand notifications) ----

func TestManager_Route_RouteNotificationForReturnsConfiguredAddress(t *testing.T) {
	logger := mockslog.NewLog(t)
	mgr := NewManager(logger, nil)

	target := mgr.Route("mail", "user@example.com").Route("sms", "+15551234567")

	assert.Equal(t, "user@example.com", target.RouteNotificationFor("mail"))
	assert.Equal(t, "+15551234567", target.RouteNotificationFor("sms"))
	// No route was ever set for "database" — the underlying map lookup
	// misses, returning the any zero value (nil), not an empty string.
	assert.Nil(t, target.RouteNotificationFor("database"))
}

func TestManager_Route_RouteNotificationFor_ReturnsRawValue_WhenRouteIsNotAString(t *testing.T) {
	logger := mockslog.NewLog(t)
	mgr := NewManager(logger, nil)

	route := struct{ ID int }{ID: 1}
	target := mgr.Route("custom", route)
	assert.Equal(t, route, target.RouteNotificationFor("custom"))
}

func TestManager_Route_NotifyNow_DeliversToChainedChannels(t *testing.T) {
	logger := mockslog.NewLog(t)
	mgr := NewManager(logger, nil)

	chA := &fakeChannel{name: "a"}
	chB := &fakeChannel{name: "b"}
	mgr.Extend(chA)
	mgr.Extend(chB)

	n := &fakeNotification{channels: []string{"a", "b"}}

	err := mgr.Route("a", "route-a").Route("b", "route-b").NotifyNow(n)
	assert.NoError(t, err)
	assert.Equal(t, 1, chA.calls)
	assert.Equal(t, 1, chB.calls)
}

func TestManager_Route_Notify_RespectsShouldQueue(t *testing.T) {
	logger := mockslog.NewLog(t)
	q := mocksqueue.NewQueue(t)
	pending := mocksqueue.NewPendingJob(t)

	q.EXPECT().Job(mock.AnythingOfType("*notification.DispatchJob"), mock.Anything).Return(pending).Once()
	pending.EXPECT().Dispatch().Return(nil).Once()

	mgr := NewManager(logger, q)
	mgr.Extend(&fakeResolvableChannel{name: "a"})

	n := &shouldQueueNotification{channels: []string{"a"}}
	err := mgr.Route("a", "route-a").Notify(n)
	assert.NoError(t, err)
}

// ---- Full queue round-trip, using the real mail channel ----

func TestManager_Send_QueuedNotification_SurvivesWorkerRoundTrip(t *testing.T) {
	logger := mockslog.NewLog(t)

	// The "dispatch side": what Send() sees.
	dispatchMailer := mocksmail.NewMail(t) // no calls expected — dispatch only queues
	dispatchQueue := mocksqueue.NewQueue(t)
	dispatchPending := mocksqueue.NewPendingJob(t)

	var capturedArgs []contractsqueue.Arg

	dispatchQueue.EXPECT().
		Job(mock.AnythingOfType("*notification.DispatchJob"), mock.Anything).
		Run(func(_ contractsqueue.Job, args ...[]contractsqueue.Arg) {
			if len(args) > 0 {
				capturedArgs = args[0]
			}
		}).
		Return(dispatchPending).
		Once()

	dispatchPending.EXPECT().Dispatch().Return(nil).Once()

	dispatchMgr := NewManager(logger, dispatchQueue)
	dispatchMgr.Extend(channels.NewMailChannel(dispatchMailer))

	err := dispatchMgr.Send(queueTestNotifiable{}, &queueableNotification{})
	assert.NoError(t, err)
	assert.NotEmpty(t, capturedArgs, "expected Send() to have queued at least one resolved channel item")

	// The "worker side": a deliberately separate Manager/DispatchJob,
	// standing in for a different process. Only capturedArgs crosses
	// from one side to the other — nothing else.
	workerMailer := mocksmail.NewMail(t)
	workerMailer.EXPECT().Send(mock.AnythingOfType("*channels.NotificationMailable")).
		Return(nil).Once()

	workerMgr := NewManager(logger, nil)
	workerMgr.Extend(channels.NewMailChannel(workerMailer))

	workerJob := NewDispatchJob(workerMgr)

	argsAny := make([]any, len(capturedArgs))
	for i, a := range capturedArgs {
		argsAny[i] = a.Value
	}

	err = workerJob.Handle(argsAny...)
	assert.NoError(t, err)
}

// Covers OnConnection()/OnQueue() actually being invoked and their
// values passed through when the notification specifies them.
func TestManager_Send_QueuedNotification_PassesConnectionAndQueue(t *testing.T) {
	logger := mockslog.NewLog(t)
	mailer := mocksmail.NewMail(t)
	q := mocksqueue.NewQueue(t)
	pending := mocksqueue.NewPendingJob(t)

	q.EXPECT().
		Job(mock.AnythingOfType("*notification.DispatchJob"), mock.Anything).
		Return(pending).
		Once()

	pending.EXPECT().OnConnection("redis").Return(pending).Once()
	pending.EXPECT().OnQueue("notifications").Return(pending).Once()
	pending.EXPECT().Dispatch().Return(nil).Once()

	mgr := NewManager(logger, q)
	mgr.Extend(channels.NewMailChannel(mailer))

	err := mgr.Send(queueTestNotifiable{}, &queueableNotificationWithRouting{})
	assert.NoError(t, err)
}
