package channels_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	contractsnotification "github.com/goravel/framework/contracts/notification"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
	"github.com/goravel/framework/notification/channels"
)

// ---- Fakes ----

type dbNotifiable struct{ id string }

func (d *dbNotifiable) RouteNotificationFor(channel string) any {
	if channel == contractsnotification.ChannelDatabase {
		return d.id
	}
	return ""
}

// numericRouteNotifiable returns an int route — tests that the database
// channel's cast.ToString handles numeric primary keys, not just strings.
type numericRouteNotifiable struct{ id int }

func (n numericRouteNotifiable) RouteNotificationFor(channel string) any {
	if channel == contractsnotification.ChannelDatabase {
		return n.id
	}
	return nil
}

// typedRouteNotifiable implements DatabaseRoutable in addition to
// Notifiable. typed is what RouteNotificationForDatabase returns and
// fallback is what RouteNotificationFor returns, so tests can pin which
// route Resolve prefers and what happens when either is empty.
type typedRouteNotifiable struct {
	typed    string
	fallback any
}

func (t typedRouteNotifiable) RouteNotificationFor(_ string) any { return t.fallback }
func (t typedRouteNotifiable) RouteNotificationForDatabase() string {
	return t.typed
}

// dbNotification does NOT implement DatabaseNotification — tests the fallback payload.
type dbNotification struct{}

func (d *dbNotification) Via(_ contractsnotification.Notifiable) []string {
	return []string{contractsnotification.ChannelDatabase}
}
func (d *dbNotification) ID() string { return "fixed-uuid-1234" }

// richDbNotification implements DatabaseNotification and NotificationWithID.
type richDbNotification struct{}

func (r *richDbNotification) Via(_ contractsnotification.Notifiable) []string {
	return []string{contractsnotification.ChannelDatabase}
}
func (r *richDbNotification) ID() string { return "" }
func (r *richDbNotification) ToDatabase(_ contractsnotification.Notifiable) map[string]any {
	return map[string]any{"invoice_id": 99, "amount": "250.00"}
}

// routedDbNotification implements NotificationWithDatabaseConnection,
// selecting a non-default connection.
type routedDbNotification struct{ richDbNotification }

func (r *routedDbNotification) DatabaseConnection() string { return "reporting" }

// unmarshalableNotification implements DatabaseNotification but returns
// data json.Marshal can't encode — deterministically exercises Resolve's
// otherwise-unreachable marshal-error branch.
type unmarshalableNotification struct{}

func (unmarshalableNotification) Via(_ contractsnotification.Notifiable) []string {
	return []string{contractsnotification.ChannelDatabase}
}
func (unmarshalableNotification) ToDatabase(_ contractsnotification.Notifiable) map[string]any {
	return map[string]any{"bad": make(chan int)} // channels aren't JSON-marshalable
}

// ---- Tests ----

func TestDatabaseChannel_Name(t *testing.T) {
	ch := channels.NewDatabaseChannel(nil)
	assert.Equal(t, contractsnotification.ChannelDatabase, ch.Name())
}

func TestDatabaseNotificationModel_TableName(t *testing.T) {
	assert.Equal(t, "notifications", channels.DatabaseNotificationModel{}.TableName())
}

func TestDatabaseChannel_Send_InsertsRecord_WithDefaultPayload(t *testing.T) {

	query := mocksorm.NewQuery(t)
	query.EXPECT().Create(mock.AnythingOfType("*channels.DatabaseNotificationModel")).
		Return(nil).Once()

	o := mocksorm.NewOrm(t)
	o.EXPECT().Query().Return(query).Once()

	ch := channels.NewDatabaseChannel(o)
	notifiable := &dbNotifiable{id: "42"}
	n := &dbNotification{}

	err := ch.Send(notifiable, n)
	assert.NoError(t, err)
}

func TestDatabaseChannel_Send_InsertsRecord_WithCustomPayload(t *testing.T) {

	query := mocksorm.NewQuery(t)
	query.EXPECT().Create(mock.MatchedBy(func(r *channels.DatabaseNotificationModel) bool {
		return r.NotifiableID == "42" && r.Data != "" && r.Type != ""
	})).Return(nil).Once()

	o := mocksorm.NewOrm(t)
	o.EXPECT().Query().Return(query).Once()

	ch := channels.NewDatabaseChannel(o)
	notifiable := &dbNotifiable{id: "42"}
	n := &richDbNotification{}

	err := ch.Send(notifiable, n)
	assert.NoError(t, err)
}

func TestDatabaseChannel_Send_UsesCallerAssignedID_WhenNotificationWithID(t *testing.T) {

	query := mocksorm.NewQuery(t)
	query.EXPECT().Create(mock.MatchedBy(func(r *channels.DatabaseNotificationModel) bool {
		return r.ID == "fixed-uuid-1234"
	})).Return(nil).Once()

	o := mocksorm.NewOrm(t)
	o.EXPECT().Query().Return(query).Once()

	ch := channels.NewDatabaseChannel(o)
	notifiable := &dbNotifiable{id: "42"}
	n := &dbNotification{}

	err := ch.Send(notifiable, n)
	assert.NoError(t, err)
}

func TestDatabaseChannel_Send_GeneratesUUID_WhenNoNotificationWithID(t *testing.T) {

	var capturedID string
	query := mocksorm.NewQuery(t)
	query.EXPECT().Create(mock.MatchedBy(func(r *channels.DatabaseNotificationModel) bool {
		capturedID = r.ID
		return true
	})).Return(nil).Once()

	o := mocksorm.NewOrm(t)
	o.EXPECT().Query().Return(query).Once()

	ch := channels.NewDatabaseChannel(o)
	notifiable := &dbNotifiable{id: "42"}
	n := &richDbNotification{} // ID() returns "" — no NotificationWithID benefit

	err := ch.Send(notifiable, n)
	assert.NoError(t, err)
	_, uuidErr := uuid.Parse(capturedID)
	assert.NoError(t, uuidErr, "expected a generated UUID, got %q", capturedID)
}

func TestDatabaseChannel_Send_ReturnsError_WhenEmptyID(t *testing.T) {
	ch := channels.NewDatabaseChannel(nil) // no orm call expected

	err := ch.Send(&dbNotifiable{id: ""}, &dbNotification{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty ID")
}

// RouteNotificationFor returns any (see contracts/notification.Notifiable).
// cast.ToString handles numbers (see the test below) as well as
// strings; a genuinely unsupported route type is treated the same as an
// empty one rather than panicking on a failed type assertion.
func TestDatabaseChannel_Send_ReturnsError_WhenRouteTypeIsUnsupported(t *testing.T) {
	ch := channels.NewDatabaseChannel(nil) // no orm call expected

	err := ch.Send(unsupportedRouteNotifiable{}, &dbNotification{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty ID")
}

// The database route is cast.ToString'd, not strictly type-asserted, so
// a numeric primary key (a common case — most notifiable models key on
// an integer ID) works without the app needing to stringify it itself.
func TestDatabaseChannel_Send_CastsNumericID_ToString(t *testing.T) {
	query := mocksorm.NewQuery(t)
	query.EXPECT().Create(mock.MatchedBy(func(r *channels.DatabaseNotificationModel) bool {
		return r.NotifiableID == "42"
	})).Return(nil).Once()

	o := mocksorm.NewOrm(t)
	o.EXPECT().Query().Return(query).Once()

	ch := channels.NewDatabaseChannel(o)
	err := ch.Send(numericRouteNotifiable{id: 42}, &dbNotification{})
	assert.NoError(t, err)
}

// The typed DatabaseRoutable route wins over RouteNotificationFor — the
// channel prefers RouteNotificationForDatabase so notifiables can route
// the database channel without string-matching channel names.
func TestDatabaseChannel_Send_PrefersTypedRoute_WhenDatabaseRoutable(t *testing.T) {
	query := mocksorm.NewQuery(t)
	query.EXPECT().Create(mock.MatchedBy(func(r *channels.DatabaseNotificationModel) bool {
		return r.NotifiableID == "42"
	})).Return(nil).Once()

	o := mocksorm.NewOrm(t)
	o.EXPECT().Query().Return(query).Once()

	ch := channels.NewDatabaseChannel(o)
	err := ch.Send(typedRouteNotifiable{typed: "42", fallback: "wrong-route"}, &dbNotification{})
	assert.NoError(t, err)
}

// An empty typed route is not an error by itself: like MailRoutable's
// empty-result fallback, Resolve falls through to RouteNotificationFor.
func TestDatabaseChannel_Send_FallsBackToGenericRoute_WhenTypedRouteEmpty(t *testing.T) {
	query := mocksorm.NewQuery(t)
	query.EXPECT().Create(mock.MatchedBy(func(r *channels.DatabaseNotificationModel) bool {
		return r.NotifiableID == "42"
	})).Return(nil).Once()

	o := mocksorm.NewOrm(t)
	o.EXPECT().Query().Return(query).Once()

	ch := channels.NewDatabaseChannel(o)
	err := ch.Send(typedRouteNotifiable{typed: "", fallback: "42"}, &dbNotification{})
	assert.NoError(t, err)
}

// Only an empty result from both the typed route and RouteNotificationFor
// is an error.
func TestDatabaseChannel_Send_ReturnsError_WhenTypedRouteAndGenericRouteEmpty(t *testing.T) {
	ch := channels.NewDatabaseChannel(nil) // no orm call expected

	err := ch.Send(typedRouteNotifiable{typed: "", fallback: ""}, &dbNotification{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty ID")
}

func TestDatabaseChannel_Send_WrapsOrmError(t *testing.T) {

	query := mocksorm.NewQuery(t)
	query.EXPECT().Create(mock.Anything).Return(errors.New("unique constraint violation")).Once()

	o := mocksorm.NewOrm(t)
	o.EXPECT().Query().Return(query).Once()

	ch := channels.NewDatabaseChannel(o)
	err := ch.Send(&dbNotifiable{id: "42"}, &dbNotification{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unique constraint violation")
}

func TestDatabaseChannel_Send_UsesConfiguredConnection_WhenNotificationWithDatabaseConnection(t *testing.T) {

	query := mocksorm.NewQuery(t)
	query.EXPECT().Create(mock.AnythingOfType("*channels.DatabaseNotificationModel")).
		Return(nil).Once()

	reportingOrm := mocksorm.NewOrm(t)
	reportingOrm.EXPECT().Query().Return(query).Once()

	o := mocksorm.NewOrm(t)
	o.EXPECT().Connection("reporting").Return(reportingOrm).Once()

	ch := channels.NewDatabaseChannel(o)
	notifiable := &dbNotifiable{id: "42"}
	n := &routedDbNotification{}

	err := ch.Send(notifiable, n)
	assert.NoError(t, err)
}

func TestDatabaseChannel_Send_UsesDefaultConnection_WhenNotNotificationWithDatabaseConnection(t *testing.T) {

	query := mocksorm.NewQuery(t)
	query.EXPECT().Create(mock.AnythingOfType("*channels.DatabaseNotificationModel")).
		Return(nil).Once()

	o := mocksorm.NewOrm(t)
	o.EXPECT().Query().Return(query).Once() // Connection() never called

	ch := channels.NewDatabaseChannel(o)
	notifiable := &dbNotifiable{id: "42"}
	n := &richDbNotification{}

	err := ch.Send(notifiable, n)
	assert.NoError(t, err)
}

func TestDatabaseChannel_Resolve_ReturnsError_WhenDataNotMarshalable(t *testing.T) {
	ch := channels.NewDatabaseChannel(nil)

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
	ch := channels.NewDatabaseChannel(nil) // no orm call expected
	err := ch.Deliver("", []byte(`{}`))
	assert.NoError(t, err)
}

func TestDatabaseChannel_Deliver_ReturnsError_WhenMalformedPayload(t *testing.T) {
	ch := channels.NewDatabaseChannel(nil) // no orm call expected
	err := ch.Deliver("1", []byte(`{not valid json`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal")
}
