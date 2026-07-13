package notification_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	contractsnotification "github.com/goravel/framework/contracts/notification"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	mocklog "github.com/goravel/framework/mocks/log"
	mockmail "github.com/goravel/framework/mocks/mail"
	mockqueue "github.com/goravel/framework/mocks/queue"
	"github.com/goravel/framework/notification"
	"github.com/goravel/framework/notification/channels"
)

// ---- Fakes ----

type queueTestNotifiable struct{}

func (queueTestNotifiable) RouteNotificationFor(channel string) string {
	if channel == "mail" {
		return "user@example.com"
	}
	return ""
}

type queueableNotification struct{}

func (n *queueableNotification) ID() string { return "" }
func (n *queueableNotification) Via(_ contractsnotification.Notifiable) []string {
	return []string{"mail"}
}
func (n *queueableNotification) ToMail(_ contractsnotification.Notifiable) contractsnotification.MailMessage {
	return contractsnotification.MailMessage{
		Subject: "Queued",
		Content: contractsnotification.MailContent{Html: "<p>hi</p>"},
	}
}
func (n *queueableNotification) OnQueue() string      { return "" }
func (n *queueableNotification) OnConnection() string { return "" }

// queueableNotificationWithRouting exercises the branch
// TestManager_Send_QueuedNotification_SurvivesWorkerRoundTrip's fix just
// revealed was never covered: OnConnection()/OnQueue() actually being
// called when the notification specifies non-empty values.
type queueableNotificationWithRouting struct{ queueableNotification }

func (n *queueableNotificationWithRouting) OnQueue() string      { return "notifications" }
func (n *queueableNotificationWithRouting) OnConnection() string { return "redis" }

// TestManager_Send_QueuedNotification_SurvivesWorkerRoundTrip is the test
// the original standalone package's design would have failed. It
// deliberately does NOT call Handle() on the exact *DispatchJob instance
// Manager.Send() constructs and passes to queue.Job() — that would prove
// nothing, since the old buggy design (unexported live
// manager/notifiable/notification fields) would also "work" if you just
// reuse the same live Go object. Instead it builds a SEPARATE
// DispatchJob backed by a SEPARATE Manager (simulating a different
// process — a real `artisan queue:work` worker booting its own app) and
// calls Handle on that one, using only the []queue.Arg captured from the
// Job() call. If delivery still happens, the fix genuinely doesn't
// depend on object identity surviving a serialization boundary — only
// on the plain (channel, route, payload) data actually making it
// through.
func TestManager_Send_QueuedNotification_SurvivesWorkerRoundTrip(t *testing.T) {
	logger := mocklog.NewLog(t)

	// The "dispatch side": what Send() sees.
	dispatchMailer := mockmail.NewMail(t) // not expected to be called — dispatch only queues
	dispatchQueue := mockqueue.NewQueue(t)
	dispatchPending := mockqueue.NewPendingJob(t)

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

	dispatchMgr := notification.NewManager(logger, dispatchQueue)
	dispatchMgr.Extend(channels.NewMailChannel(dispatchMailer, logger))

	err := dispatchMgr.Send(queueTestNotifiable{}, &queueableNotification{})
	assert.NoError(t, err)
	assert.NotEmpty(t, capturedArgs, "expected Send() to have queued at least one resolved channel item")

	// The "worker side": a deliberately separate Manager/DispatchJob,
	// standing in for a different process. Only capturedArgs crosses
	// from one side to the other — nothing else.
	workerMailer := mockmail.NewMail(t)
	workerMailer.On("Send", mock.AnythingOfType("*channels.NotificationMailable")).
		Return(nil).Once()

	workerMgr := notification.NewManager(logger, nil)
	workerMgr.Extend(channels.NewMailChannel(workerMailer, logger))

	workerJob := notification.NewDispatchJob(workerMgr)

	argsAny := make([]any, len(capturedArgs))
	for i, a := range capturedArgs {
		argsAny[i] = a.Value
	}

	err = workerJob.Handle(argsAny...)
	assert.NoError(t, err)

	workerMailer.AssertExpectations(t)
	dispatchMailer.AssertNotCalled(t, "Send", mock.Anything)
}

// Covers the branch the fix above revealed was untested: OnConnection()/
// OnQueue() actually being invoked and their values passed through when
// the notification specifies them.
func TestManager_Send_QueuedNotification_PassesConnectionAndQueue(t *testing.T) {
	logger := mocklog.NewLog(t)
	mailer := mockmail.NewMail(t)
	q := mockqueue.NewQueue(t)
	pending := mockqueue.NewPendingJob(t)

	q.EXPECT().
		Job(mock.AnythingOfType("*notification.DispatchJob"), mock.Anything).
		Return(pending).
		Once()

	pending.EXPECT().OnConnection("redis").Return(pending).Once()
	pending.EXPECT().OnQueue("notifications").Return(pending).Once()
	pending.EXPECT().Dispatch().Return(nil).Once()

	mgr := notification.NewManager(logger, q)
	mgr.Extend(channels.NewMailChannel(mailer, logger))

	err := mgr.Send(queueTestNotifiable{}, &queueableNotificationWithRouting{})
	assert.NoError(t, err)
}
