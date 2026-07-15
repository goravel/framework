package channels_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	contractsnotification "github.com/goravel/framework/contracts/notification"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mockorm "github.com/goravel/framework/mocks/database/orm"
	mocklog "github.com/goravel/framework/mocks/log"
	"github.com/goravel/framework/notification/channels"
)

// ---- Fakes ----

type dbNotifiable struct{ id string }

func (d *dbNotifiable) RouteNotificationFor(channel string) string {
	if channel == "database" {
		return d.id
	}
	return ""
}

// dbNotification does NOT implement DatabaseNotification — tests the fallback payload.
type dbNotification struct{}

func (d *dbNotification) Via(_ contractsnotification.Notifiable) []string {
	return []string{"database"}
}
func (d *dbNotification) ID() string { return "fixed-uuid-1234" }

// plainDbNotification implements only Notification (no ID() at all),
// proving NotificationWithID is truly optional, not just "return empty".
type plainDbNotification struct{}

func (p *plainDbNotification) Via(_ contractsnotification.Notifiable) []string {
	return []string{"database"}
}

// richDbNotification implements DatabaseNotification.
type richDbNotification struct{}

func (r *richDbNotification) Via(_ contractsnotification.Notifiable) []string {
	return []string{"database"}
}
func (r *richDbNotification) ID() string { return "" }
func (r *richDbNotification) ToDatabase(_ contractsnotification.Notifiable) map[string]any {
	return map[string]any{"invoice_id": 99, "amount": "250.00"}
}

// ---- Tests ----

func TestDatabaseChannel_Name(t *testing.T) {
	ch := channels.NewDatabaseChannel(nil, nil, nil)
	assert.Equal(t, "database", ch.Name())
}

func TestDatabaseChannel_Send_InsertsRecord_WithDefaultPayload(t *testing.T) {
	logger := mocklog.NewLog(t)
	logger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Maybe()

	query := mockorm.NewQuery(t)
	query.On("Create", mock.AnythingOfType("*channels.DatabaseNotificationModel")).
		Return(nil).Once()

	o := mockorm.NewOrm(t)
	o.On("Query").Return(query)

	ch := channels.NewDatabaseChannel(o, logger, nil)
	notifiable := &dbNotifiable{id: "42"}
	n := &dbNotification{}

	err := ch.Send(notifiable, n)
	assert.NoError(t, err)
	query.AssertExpectations(t)
}

func TestDatabaseChannel_Send_InsertsRecord_WithCustomPayload(t *testing.T) {
	logger := mocklog.NewLog(t)
	logger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Maybe()

	query := mockorm.NewQuery(t)
	query.On("Create", mock.MatchedBy(func(r *channels.DatabaseNotificationModel) bool {
		// Verify the JSON payload contains the custom fields.
		return r.NotifiableID == "42" && len(r.Data) > 0
	})).Return(nil).Once()

	o := mockorm.NewOrm(t)
	o.On("Query").Return(query)

	ch := channels.NewDatabaseChannel(o, logger, nil)
	notifiable := &dbNotifiable{id: "42"}
	n := &richDbNotification{}

	err := ch.Send(notifiable, n)
	assert.NoError(t, err)
	query.AssertExpectations(t)
}

func TestDatabaseChannel_Send_UsesCallerAssignedID_WhenNotificationWithIDImplemented(t *testing.T) {
	logger := mocklog.NewLog(t)
	logger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Maybe()

	query := mockorm.NewQuery(t)
	query.On("Create", mock.MatchedBy(func(r *channels.DatabaseNotificationModel) bool {
		return r.ID == "fixed-uuid-1234"
	})).Return(nil).Once()

	o := mockorm.NewOrm(t)
	o.On("Query").Return(query)

	ch := channels.NewDatabaseChannel(o, logger, nil)
	err := ch.Send(&dbNotifiable{id: "42"}, &dbNotification{})
	assert.NoError(t, err)
	query.AssertExpectations(t)
}

func TestDatabaseChannel_Send_GeneratesUUID_WhenNotificationWithIDNotImplemented(t *testing.T) {
	logger := mocklog.NewLog(t)
	logger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Maybe()

	var generatedID string
	query := mockorm.NewQuery(t)
	query.On("Create", mock.MatchedBy(func(r *channels.DatabaseNotificationModel) bool {
		generatedID = r.ID
		return r.ID != ""
	})).Return(nil).Once()

	o := mockorm.NewOrm(t)
	o.On("Query").Return(query)

	ch := channels.NewDatabaseChannel(o, logger, nil)
	err := ch.Send(&dbNotifiable{id: "42"}, &plainDbNotification{})
	assert.NoError(t, err)

	_, parseErr := uuid.Parse(generatedID)
	assert.NoError(t, parseErr, "expected a generated UUID when NotificationWithID isn't implemented")
	query.AssertExpectations(t)
}

func TestDatabaseChannel_Send_ReturnsError_WhenEmptyID(t *testing.T) {
	logger := mocklog.NewLog(t)
	o := mockorm.NewOrm(t)

	ch := channels.NewDatabaseChannel(o, logger, nil)
	notifiable := &dbNotifiable{id: ""} // no routing ID
	n := &dbNotification{}

	err := ch.Send(notifiable, n)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty ID")
}

func TestDatabaseChannel_Send_WrapsOrmError(t *testing.T) {
	logger := mocklog.NewLog(t)

	query := mockorm.NewQuery(t)
	query.On("Create", mock.Anything).Return(errors.New("unique constraint violation")).Once()

	o := mockorm.NewOrm(t)
	o.On("Query").Return(query)

	ch := channels.NewDatabaseChannel(o, logger, nil)
	err := ch.Send(&dbNotifiable{id: "1"}, &dbNotification{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unique constraint violation")
}

func TestDatabaseChannel_Send_UsesConfiguredConnection_WhenSet(t *testing.T) {
	logger := mocklog.NewLog(t)
	logger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Maybe()

	query := mockorm.NewQuery(t)
	query.On("Create", mock.AnythingOfType("*channels.DatabaseNotificationModel")).
		Return(nil).Once()

	// The "reporting" connection's Orm is what should actually receive Query().
	reportingOrm := mockorm.NewOrm(t)
	reportingOrm.On("Query").Return(query)

	o := mockorm.NewOrm(t)
	o.On("Connection", "reporting").Return(reportingOrm)

	config := mocksconfig.NewConfig(t)
	config.On("GetString", "notification.channels.database.connection").Return("reporting")

	ch := channels.NewDatabaseChannel(o, logger, config)
	err := ch.Send(&dbNotifiable{id: "42"}, &dbNotification{})

	assert.NoError(t, err)
	query.AssertExpectations(t)
}

func TestDatabaseChannel_Send_UsesDefaultConnection_WhenConfigConnectionEmpty(t *testing.T) {
	logger := mocklog.NewLog(t)
	logger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Maybe()

	query := mockorm.NewQuery(t)
	query.On("Create", mock.AnythingOfType("*channels.DatabaseNotificationModel")).
		Return(nil).Once()

	o := mockorm.NewOrm(t)
	o.On("Query").Return(query)

	config := mocksconfig.NewConfig(t)
	config.On("GetString", "notification.channels.database.connection").Return("")

	ch := channels.NewDatabaseChannel(o, logger, config)
	err := ch.Send(&dbNotifiable{id: "42"}, &dbNotification{})

	assert.NoError(t, err)
	query.AssertExpectations(t)
}

// TestDatabaseChannel_SendNow_BehavesIdenticallyToSend confirms the
// Send-delegates-to-SendNow relationship for the database channel.
func TestDatabaseChannel_SendNow_BehavesIdenticallyToSend(t *testing.T) {
	logger := mocklog.NewLog(t)
	logger.On("Debugf", mock.Anything, mock.Anything, mock.Anything).Maybe()

	query := mockorm.NewQuery(t)
	query.On("Create", mock.AnythingOfType("*channels.DatabaseNotificationModel")).
		Return(nil).Once()

	o := mockorm.NewOrm(t)
	o.On("Query").Return(query)

	ch := channels.NewDatabaseChannel(o, logger, nil)
	notifiable := &dbNotifiable{id: "42"}
	n := &dbNotification{}

	err := ch.SendNow(notifiable, n)
	assert.NoError(t, err)
	query.AssertExpectations(t)
}

func TestDatabaseNotificationModel_TableName(t *testing.T) {
	assert.Equal(t, "notifications", channels.DatabaseNotificationModel{}.TableName())
}

// unmarshalableNotification implements DatabaseNotification but returns
// data json.Marshal can't encode — deterministically exercises Resolve's
// otherwise-unreachable marshal-error branch.
type unmarshalableNotification struct{}

func (unmarshalableNotification) Via(_ contractsnotification.Notifiable) []string {
	return []string{"database"}
}
func (unmarshalableNotification) ToDatabase(_ contractsnotification.Notifiable) map[string]any {
	return map[string]any{"bad": make(chan int)} // channels aren't JSON-marshalable
}

func TestDatabaseChannel_Resolve_ReturnsError_WhenDataNotMarshalable(t *testing.T) {
	ch := channels.NewDatabaseChannel(nil, nil, nil)

	_, _, err := ch.Resolve(&dbNotifiable{id: "1"}, unmarshalableNotification{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal payload")
}

// ---- Deliver-specific branch coverage ----
//
// Deliver is exported and callable directly (DispatchJob does exactly
// this on the queued path), so these exercise branches Resolve→Deliver
// via Send() never reaches.

func TestDatabaseChannel_Deliver_NoOp_WhenEmptyRoute(t *testing.T) {
	ch := channels.NewDatabaseChannel(nil, nil, nil) // no orm call expected
	err := ch.Deliver("", []byte(`{}`))
	assert.NoError(t, err)
}

func TestDatabaseChannel_Deliver_ReturnsError_WhenMalformedPayload(t *testing.T) {
	ch := channels.NewDatabaseChannel(nil, nil, nil) // no orm call expected
	err := ch.Deliver("1", []byte(`{not valid json`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal")
}
