package notification_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	contractsnotification "github.com/goravel/framework/contracts/notification"
	mocklog "github.com/goravel/framework/mocks/log"
	mockqueue "github.com/goravel/framework/mocks/queue"
	"github.com/goravel/framework/notification"
)

// ---- Fakes ----

type fakeNotifiable struct{ email string }

func (f *fakeNotifiable) RouteNotificationFor(channel string) string {
	if channel == "mail" {
		return f.email
	}
	return ""
}

type fakeNotification struct {
	channels []string
	id       string
}

func (f *fakeNotification) Via(_ contractsnotification.Notifiable) []string { return f.channels }
func (f *fakeNotification) ID() string                                      { return f.id }

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
func (c *fakeChannel) SendNow(notifiable contractsnotification.Notifiable, n contractsnotification.Notification) error {
	return c.Send(notifiable, n)
}

// shouldQueueNotification implements ShouldQueue with configurable
// channels, for exercising Manager.dispatchQueued's error paths.
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
func (c *fakeResolvableChannel) SendNow(notifiable contractsnotification.Notifiable, n contractsnotification.Notification) error {
	return c.Send(notifiable, n)
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

// ---- Tests ----

func TestManager_SendNow_CallsCorrectChannels(t *testing.T) {
	logger := mocklog.NewLog(t)
	logger.On("Debugf", mock.Anything, mock.Anything).Maybe()

	mgr := notification.NewManager(logger, nil)

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
	logger := mocklog.NewLog(t)
	logger.On("Errorf", mock.Anything, mock.Anything).Once()

	mgr := notification.NewManager(logger, nil)

	n := &fakeNotification{channels: []string{"nonexistent"}}
	notifiable := &fakeNotifiable{}

	err := mgr.SendNow(notifiable, n)
	// Returns the error but does not panic
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
	logger.AssertExpectations(t)
}

func TestManager_SendNow_LogsChannelError_ContinuesOthers(t *testing.T) {
	logger := mocklog.NewLog(t)
	logger.On("Errorf", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()

	mgr := notification.NewManager(logger, nil)

	chFail := &fakeChannel{name: "fail", sendErr: errors.New("smtp down")}
	chOK := &fakeChannel{name: "ok"}
	mgr.Extend(chFail)
	mgr.Extend(chOK)

	n := &fakeNotification{channels: []string{"fail", "ok"}}
	notifiable := &fakeNotifiable{}

	_ = mgr.SendNow(notifiable, n)

	// The "ok" channel must still be called even though "fail" errored.
	assert.Equal(t, 1, chOK.calls)
	logger.AssertExpectations(t)
}

func TestManager_SendNow_NoChannels_Warns(t *testing.T) {
	logger := mocklog.NewLog(t)
	logger.On("Errorf", mock.Anything, mock.Anything, mock.Anything).Once()

	mgr := notification.NewManager(logger, nil)

	n := &fakeNotification{channels: []string{}}
	err := mgr.SendNow(&fakeNotifiable{}, n)
	assert.NoError(t, err)
	logger.AssertExpectations(t)
}

func TestManager_Channel_ReturnsNil_WhenNotRegistered(t *testing.T) {
	logger := mocklog.NewLog(t)
	logger.EXPECT().Errorf(mock.Anything, mock.Anything).Once()
	mgr := notification.NewManager(logger, nil)

	got := mgr.Channel("missing")
	assert.Nil(t, got)
}

func TestManager_Extend_RegistersChannel(t *testing.T) {
	logger := mocklog.NewLog(t)
	mgr := notification.NewManager(logger, nil)

	ch := &fakeChannel{name: "custom"}
	mgr.Extend(ch)

	got := mgr.Channel("custom")
	assert.Equal(t, ch, got)
}

// ---- dispatchQueued error paths (only reachable via Send() + ShouldQueue + a configured queue) ----

func TestManager_Send_QueuedNotification_UnregisteredChannel_ReturnsError(t *testing.T) {
	logger := mocklog.NewLog(t)
	logger.On("Errorf", mock.Anything, mock.Anything).Once()

	q := mockqueue.NewQueue(t) // no calls expected — should fail before reaching the queue
	mgr := notification.NewManager(logger, q)

	n := &shouldQueueNotification{channels: []string{"missing"}}
	err := mgr.Send(&fakeNotifiable{}, n)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing")
}

func TestManager_Send_QueuedNotification_ChannelNotResolvable_ReturnsError(t *testing.T) {
	logger := mocklog.NewLog(t)
	logger.On("Errorf", mock.Anything, mock.Anything).Once()

	mgr := notification.NewManager(logger, mockqueue.NewQueue(t))
	mgr.Extend(&fakeChannel{name: "plain"}) // implements Channel only, not ResolvableChannel

	n := &shouldQueueNotification{channels: []string{"plain"}}
	err := mgr.Send(&fakeNotifiable{}, n)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not support queued dispatch")
}

func TestManager_Send_QueuedNotification_ResolveError_ReturnsError(t *testing.T) {
	logger := mocklog.NewLog(t)
	logger.On("Errorf", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()

	mgr := notification.NewManager(logger, mockqueue.NewQueue(t))
	mgr.Extend(&fakeResolvableChannel{name: "broken", resolveErr: errors.New("boom")})

	n := &shouldQueueNotification{channels: []string{"broken"}}
	err := mgr.Send(&fakeNotifiable{}, n)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
}

// ---- Route (on-demand notifications) ----

func TestManager_Route_RouteNotificationForReturnsConfiguredAddress(t *testing.T) {
	logger := mocklog.NewLog(t)
	mgr := notification.NewManager(logger, nil)

	target := mgr.Route("mail", "user@example.com").Route("sms", "+15551234567")

	assert.Equal(t, "user@example.com", target.RouteNotificationFor("mail"))
	assert.Equal(t, "+15551234567", target.RouteNotificationFor("sms"))
	assert.Equal(t, "", target.RouteNotificationFor("database"))
}

func TestManager_Route_NotifyNow_DeliversToChainedChannels(t *testing.T) {
	logger := mocklog.NewLog(t)
	mgr := notification.NewManager(logger, nil)

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
	logger := mocklog.NewLog(t)
	q := mockqueue.NewQueue(t)
	pending := mockqueue.NewPendingJob(t)

	q.EXPECT().Job(mock.AnythingOfType("*notification.DispatchJob"), mock.Anything).Return(pending).Once()
	pending.EXPECT().Dispatch().Return(nil).Once()

	mgr := notification.NewManager(logger, q)
	mgr.Extend(&fakeResolvableChannel{name: "a"})

	n := &shouldQueueNotification{channels: []string{"a"}}
	err := mgr.Route("a", "route-a").Notify(n)
	assert.NoError(t, err)
}

// TestManager_Send_NonQueueableNotification_UsesDispatchSync exercises
// Send()'s fallthrough branch directly — every other Send() test in this
// file uses a ShouldQueue notification, leaving the plain-notification
// path (which SendNow's own tests reach, but Send() itself never had a
// direct test for) uncovered.
func TestManager_Send_NonQueueableNotification_UsesDispatchSync(t *testing.T) {
	logger := mocklog.NewLog(t)
	logger.On("Debugf", mock.Anything, mock.Anything).Maybe()

	mgr := notification.NewManager(logger, nil)
	ch := &fakeChannel{name: "a"}
	mgr.Extend(ch)

	n := &fakeNotification{channels: []string{"a"}}
	err := mgr.Send(&fakeNotifiable{}, n)

	assert.NoError(t, err)
	assert.Equal(t, 1, ch.calls)
}

func TestManager_Send_QueuedNotification_DispatchError_ReturnsError(t *testing.T) {
	logger := mocklog.NewLog(t)

	q := mockqueue.NewQueue(t)
	pending := mockqueue.NewPendingJob(t)
	q.EXPECT().Job(mock.AnythingOfType("*notification.DispatchJob"), mock.Anything).Return(pending).Once()
	pending.EXPECT().Dispatch().Return(errors.New("queue connection refused")).Once()

	mgr := notification.NewManager(logger, q)
	mgr.Extend(&fakeResolvableChannel{name: "a"})

	n := &shouldQueueNotification{channels: []string{"a"}}
	err := mgr.Send(&fakeNotifiable{}, n)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "queue connection refused")
}
