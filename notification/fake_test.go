package notification

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	contractsnotification "github.com/goravel/framework/contracts/notification"
	mockslog "github.com/goravel/framework/mocks/log"
	mocksqueue "github.com/goravel/framework/mocks/queue"
)

// fakeUser is a notifiable identified by its primary key, like a model.
type fakeUser struct {
	ID    string
	Email string
}

func (u *fakeUser) RouteNotificationFor(string) any      { return u.Email }
func (u *fakeUser) RouteNotificationForDatabase() string { return u.ID }

func TestFake_RecordsInsteadOfDelivering(t *testing.T) {
	mgr := NewManager(mockslog.NewLog(t), mocksqueue.NewQueue(t))
	channel := &fakeChannel{name: "a"}
	mgr.Extend(channel)

	fake := mgr.Fake()
	user := &fakeUser{ID: "1", Email: "ana@example.com"}

	assert.NoError(t, mgr.Send(user, &fakeNotification{channels: []string{"a"}}))
	assert.NoError(t, mgr.SendNow(user, &shouldQueueNotification{channels: []string{"a"}}))
	// Queued notifications are recorded without reaching the queue.
	assert.NoError(t, mgr.Send(user, &shouldQueueNotification{channels: []string{"b"}}))

	assert.Zero(t, channel.calls)
	assert.True(t, fake.AssertCount(3))
	assert.False(t, fake.AssertNothingSent())
	// A model loaded again matches by its primary key.
	assert.True(t, fake.AssertSentTo(&fakeUser{ID: "1"}, (*fakeNotification)(nil)))
	assert.True(t, fake.AssertSentTo(user, &fakeNotification{}, func(sent contractsnotification.Sent) bool {
		return sent.HasChannel("a") && !sent.Queued()
	}))
	assert.False(t, fake.AssertSentTo(user, &fakeNotification{}, func(sent contractsnotification.Sent) bool {
		return sent.HasChannel("b")
	}))
	assert.True(t, fake.AssertSentTo(user, &shouldQueueNotification{}, func(sent contractsnotification.Sent) bool {
		return sent.Queued() && sent.HasChannel("b")
	}))
	assert.True(t, fake.AssertSentTo(user, &shouldQueueNotification{}, func(sent contractsnotification.Sent) bool {
		return !sent.Queued() && sent.HasChannel("a")
	}))
	assert.True(t, fake.AssertNotSentTo(&fakeUser{ID: "2"}, &fakeNotification{}))
	assert.False(t, fake.AssertNotSentTo(user, &fakeNotification{}))
	assert.False(t, fake.AssertNotSentTo(user, &fakeNotification{}, func(sent contractsnotification.Sent) bool {
		return sent.HasChannel("a")
	}))
	assert.True(t, fake.AssertNotSentTo(user, &fakeNotification{}, func(sent contractsnotification.Sent) bool {
		return sent.HasChannel("b")
	}))
	assert.True(t, fake.AssertNotSentTo(user, &afterSendingNotification{}))
	assert.True(t, fake.AssertNothingSentTo(&fakeUser{ID: "2"}))
	assert.False(t, fake.AssertNothingSentTo(user))
	assert.True(t, fake.AssertSentTimes(&shouldQueueNotification{}, 2))
	assert.False(t, fake.AssertSentTimes(&shouldQueueNotification{}, 1))
	assert.False(t, fake.AssertCount(2))
	assert.Len(t, fake.Sent(&fakeNotification{}), 1)
	assert.Nil(t, fake.Sent(nil))
}

func TestFake_ComparesOtherNotifiablesDeeply(t *testing.T) {
	mgr := NewManager(mockslog.NewLog(t), nil)
	fake := mgr.Fake()

	assert.NoError(t, mgr.Send(&fakeNotifiable{email: "ana@example.com"}, &fakeNotification{channels: []string{"a"}}))
	assert.NoError(t, mgr.Send(queueTestNotifiable{}, &fakeNotification{channels: []string{"a"}}))

	assert.True(t, fake.AssertSentTo(&fakeNotifiable{email: "ana@example.com"}, &fakeNotification{}))
	assert.False(t, fake.AssertSentTo(&fakeNotifiable{email: "bob@example.com"}, &fakeNotification{}))
	assert.True(t, fake.AssertSentTo(queueTestNotifiable{}, &fakeNotification{}))
	assert.False(t, fake.AssertSentTo(nil, &fakeNotification{}))
}

func TestFake_ComparesUnsavedModelsDeeply(t *testing.T) {
	mgr := NewManager(mockslog.NewLog(t), nil)
	fake := mgr.Fake()

	// Without a primary key there is no route to compare, so the whole user is.
	assert.NoError(t, mgr.Send(&fakeUser{Email: "ana@example.com"}, &fakeNotification{channels: []string{"a"}}))

	assert.True(t, fake.AssertSentTo(&fakeUser{Email: "ana@example.com"}, &fakeNotification{}))
	assert.False(t, fake.AssertSentTo(&fakeUser{Email: "bob@example.com"}, &fakeNotification{}))
	assert.False(t, fake.AssertSentTo(&fakeUser{ID: "1", Email: "ana@example.com"}, &fakeNotification{}))
	assert.False(t, fake.AssertSentTo(nil, &fakeNotification{}))
}

func TestFake_LeavesOutChannelsShouldSendRejects(t *testing.T) {
	mgr := NewManager(mockslog.NewLog(t), nil)
	fake := mgr.Fake()
	user := &fakeUser{ID: "1"}

	assert.NoError(t, mgr.Send(user, &shouldSendNotification{channels: []string{"a", "b"}, skip: map[string]bool{"b": true}}))
	// Nothing is left to send through, so nothing is recorded, like Laravel.
	assert.NoError(t, mgr.Send(user, &shouldSendNotification{channels: []string{"a"}, skip: map[string]bool{"a": true}}))
	assert.NoError(t, mgr.Send(user, &fakeNotification{}))

	sent := fake.Sent(&shouldSendNotification{})
	assert.Len(t, sent, 1)
	assert.Equal(t, []string{"a"}, sent[0].Channels())
	assert.True(t, fake.AssertCount(1))
}

func TestFake_OnDemand(t *testing.T) {
	mgr := NewManager(mockslog.NewLog(t), nil)
	fake := mgr.Fake()

	assert.NoError(t, mgr.Route("mail", "ana@example.com").Notify(&fakeNotification{channels: []string{"mail"}}))
	assert.NoError(t, mgr.Send(&fakeUser{ID: "1"}, &shouldQueueNotification{channels: []string{"mail"}}))

	assert.True(t, fake.AssertSentOnDemand(&fakeNotification{}, func(sent contractsnotification.Sent) bool {
		return sent.Notifiable().RouteNotificationFor("mail") == "ana@example.com"
	}))
	assert.False(t, fake.AssertSentOnDemand(&fakeNotification{}, func(sent contractsnotification.Sent) bool {
		return sent.Notifiable().RouteNotificationFor("mail") == "bob@example.com"
	}))
	assert.False(t, fake.AssertSentOnDemand(&shouldQueueNotification{}))
}

func TestFake_QueuedNeedsAQueue(t *testing.T) {
	mgr := NewManager(mockslog.NewLog(t), nil)
	fake := mgr.Fake()

	// Without a queue a ShouldQueue notification is delivered right away.
	assert.NoError(t, mgr.Send(&fakeUser{ID: "1"}, &shouldQueueNotification{channels: []string{"a"}}))

	assert.True(t, fake.AssertSentTo(&fakeUser{ID: "1"}, &shouldQueueNotification{}, func(sent contractsnotification.Sent) bool {
		return !sent.Queued()
	}))
}

func TestFake_CallingFakeAgainKeepsTheRecords(t *testing.T) {
	mgr := NewManager(mockslog.NewLog(t), nil)
	fake := mgr.Fake()
	assert.NoError(t, mgr.Send(&fakeUser{ID: "1"}, &fakeNotification{channels: []string{"a"}}))

	assert.Same(t, fake, mgr.Fake())
	assert.True(t, fake.AssertCount(1))
}

func TestFake_Reset(t *testing.T) {
	mgr := NewManager(mockslog.NewLog(t), nil)
	channel := &fakeChannel{name: "a"}
	mgr.Extend(channel)

	fake := mgr.Fake()
	assert.NoError(t, mgr.Send(&fakeUser{ID: "1"}, &fakeNotification{channels: []string{"a"}}))
	fake.Reset()

	assert.True(t, fake.AssertNothingSent())

	assert.NoError(t, mgr.SendNow(&fakeUser{ID: "1"}, &fakeNotification{channels: []string{"a"}}))
	assert.Equal(t, 1, channel.calls)
	assert.True(t, fake.AssertNothingSent())
}

func TestFake_RecordsAreCopies(t *testing.T) {
	mgr := NewManager(mockslog.NewLog(t), nil)
	fake := mgr.Fake()
	notification := &fakeNotification{channels: []string{"a"}}
	assert.NoError(t, mgr.Send(&fakeUser{ID: "1"}, notification))

	notification.channels[0] = "changed"
	fake.Sent(notification)[0].Channels()[0] = "changed"

	assert.True(t, fake.AssertSentTo(&fakeUser{ID: "1"}, notification, func(sent contractsnotification.Sent) bool {
		return sent.HasChannel("a")
	}))
}

func TestFake_FiltersMayCallTheFake(t *testing.T) {
	mgr := NewManager(mockslog.NewLog(t), nil)
	fake := mgr.Fake()
	assert.NoError(t, mgr.Send(&fakeUser{ID: "1"}, &fakeNotification{channels: []string{"a"}}))

	assert.True(t, fake.AssertSentTo(&fakeUser{ID: "1"}, &fakeNotification{}, func(contractsnotification.Sent) bool {
		return fake.AssertCount(1)
	}))
}

func TestFake_Concurrent(t *testing.T) {
	mgr := NewManager(mockslog.NewLog(t), nil)
	fake := mgr.Fake()

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			assert.NoError(t, mgr.Send(&fakeUser{ID: "1"}, &fakeNotification{channels: []string{"a"}}))
		}()
		go func() {
			defer wg.Done()
			fake.Sent(&fakeNotification{})
		}()
	}
	wg.Wait()

	assert.True(t, fake.AssertSentTimes(&fakeNotification{}, 50))
}
