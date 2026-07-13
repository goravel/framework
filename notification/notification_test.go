package notification_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	contractsnotification "github.com/goravel/framework/contracts/notification"
	mocklog "github.com/goravel/framework/mocks/log"
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
	logger.On("Errorf", mock.Anything, mock.Anything, mock.Anything).Once()

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

func TestManager_Channel_ReturnsError_WhenNotRegistered(t *testing.T) {
	logger := mocklog.NewLog(t)
	mgr := notification.NewManager(logger, nil)

	_, err := mgr.Channel("missing")
	assert.Error(t, err)
}

func TestManager_Extend_RegistersChannel(t *testing.T) {
	logger := mocklog.NewLog(t)
	mgr := notification.NewManager(logger, nil)

	ch := &fakeChannel{name: "custom"}
	mgr.Extend(ch)

	got, err := mgr.Channel("custom")
	assert.NoError(t, err)
	assert.Equal(t, ch, got)
}
