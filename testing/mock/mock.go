package mock

import (
	"context"

	httpcontract "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/foundation"
	authmock "github.com/goravel/framework/mocks/auth"
	accessmock "github.com/goravel/framework/mocks/auth/access"
	cachemock "github.com/goravel/framework/mocks/cache"
	configmock "github.com/goravel/framework/mocks/config"
	consolemock "github.com/goravel/framework/mocks/console"
	cryptmock "github.com/goravel/framework/mocks/crypt"
	ormmock "github.com/goravel/framework/mocks/database/orm"
	seedermock "github.com/goravel/framework/mocks/database/seeder"
	eventmock "github.com/goravel/framework/mocks/event"
	filesystemmock "github.com/goravel/framework/mocks/filesystem"
	foundationmock "github.com/goravel/framework/mocks/foundation"
	grpcmock "github.com/goravel/framework/mocks/grpc"
	hashmock "github.com/goravel/framework/mocks/hash"
	httpmock "github.com/goravel/framework/mocks/http"
	mailmock "github.com/goravel/framework/mocks/mail"
	queuemock "github.com/goravel/framework/mocks/queue"
	translationmock "github.com/goravel/framework/mocks/translation"
	validatemock "github.com/goravel/framework/mocks/validation"
)

func Factory() *factory {
	app := &foundationmock.Application{}
	foundation.App = app

	return &factory{
		app: app,
	}
}

type factory struct {
	app *foundationmock.Application
}

func (r *factory) App() *foundationmock.Application {
	return r.app
}

func (r *factory) Artisan() *consolemock.Artisan {
	mockArtisan := &consolemock.Artisan{}
	r.app.On("MakeArtisan").Return(mockArtisan)

	return mockArtisan
}

func (r *factory) Auth(ctx httpcontract.Context) *authmock.Auth {
	mockAuth := &authmock.Auth{}
	r.app.On("MakeAuth", ctx).Return(mockAuth)

	return mockAuth
}

func (r *factory) Cache() *cachemock.Cache {
	mockCache := &cachemock.Cache{}
	r.app.On("MakeCache").Return(mockCache)

	return mockCache
}

func (r *factory) CacheDriver() *cachemock.Driver {
	return &cachemock.Driver{}
}

func (r *factory) CacheLock() *cachemock.Lock {
	return &cachemock.Lock{}
}

func (r *factory) Context() *httpmock.Context {
	return &httpmock.Context{}
}

func (r *factory) ContextRequest() *httpmock.ContextRequest {
	return &httpmock.ContextRequest{}
}

func (r *factory) ContextResponse() *httpmock.ContextResponse {
	return &httpmock.ContextResponse{}
}

func (r *factory) Config() *configmock.Config {
	mockConfig := &configmock.Config{}
	r.app.On("MakeConfig").Return(mockConfig)

	return mockConfig
}

func (r *factory) Crypt() *cryptmock.Crypt {
	mockCrypt := &cryptmock.Crypt{}
	r.app.On("MakeCrypt").Return(mockCrypt)

	return mockCrypt
}

func (r *factory) Event() *eventmock.Instance {
	mockEvent := &eventmock.Instance{}
	r.app.On("MakeEvent").Return(mockEvent)

	return mockEvent
}

func (r *factory) EventTask() *eventmock.Task {
	return &eventmock.Task{}
}

func (r *factory) Gate() *accessmock.Gate {
	mockGate := &accessmock.Gate{}
	r.app.On("MakeGate").Return(mockGate)

	return mockGate
}

func (r *factory) Grpc() *grpcmock.Grpc {
	mockGrpc := &grpcmock.Grpc{}
	r.app.On("MakeGrpc").Return(mockGrpc)

	return mockGrpc
}

func (r *factory) Hash() *hashmock.Hash {
	mockHash := &hashmock.Hash{}
	r.app.On("MakeHash").Return(mockHash)

	return mockHash
}

func (r *factory) Lang(ctx context.Context) *translationmock.Translator {
	mockTranslator := &translationmock.Translator{}
	r.app.On("MakeLang", ctx).Return(mockTranslator)

	return mockTranslator
}

func (r *factory) Log() {
	r.app.On("MakeLog").Return(NewTestLog())
}

func (r *factory) Mail() *mailmock.Mail {
	mockMail := &mailmock.Mail{}
	r.app.On("MakeMail").Return(mockMail)

	return mockMail
}

func (r *factory) Orm() *ormmock.Orm {
	mockOrm := &ormmock.Orm{}
	r.app.On("MakeOrm").Return(mockOrm)

	return mockOrm
}

func (r *factory) OrmAssociation() *ormmock.Association {
	return &ormmock.Association{}
}

func (r *factory) OrmQuery() *ormmock.Query {
	return &ormmock.Query{}
}

func (r *factory) OrmToSql() *ormmock.ToSql {
	return &ormmock.ToSql{}
}

func (r *factory) Queue() *queuemock.Queue {
	mockQueue := &queuemock.Queue{}
	r.app.On("MakeQueue").Return(mockQueue)

	return mockQueue
}

func (r *factory) QueueTask() *queuemock.Task {
	return &queuemock.Task{}
}

func (r *factory) RateLimiter() *httpmock.RateLimiter {
	mockRateLimiter := &httpmock.RateLimiter{}
	r.app.On("MakeRateLimiter").Return(mockRateLimiter)

	return mockRateLimiter
}

func (r *factory) Response() *httpmock.Response {
	return &httpmock.Response{}
}

func (r *factory) ResponseStatus() *httpmock.ResponseStatus {
	return &httpmock.ResponseStatus{}
}

func (r *factory) ResponseOrigin() *httpmock.ResponseOrigin {
	return &httpmock.ResponseOrigin{}
}

func (r *factory) ResponseView() *httpmock.ResponseView {
	return &httpmock.ResponseView{}
}

func (r *factory) Seeder() *seedermock.Facade {
	mockSeeder := &seedermock.Facade{}
	r.app.On("MakeSeeder").Return(mockSeeder)

	return mockSeeder
}

func (r *factory) Storage() *filesystemmock.Storage {
	mockStorage := &filesystemmock.Storage{}
	r.app.On("MakeStorage").Return(mockStorage)

	return mockStorage
}

func (r *factory) StorageDriver() *filesystemmock.Driver {
	return &filesystemmock.Driver{}
}

func (r *factory) StorageFile() *filesystemmock.File {
	return &filesystemmock.File{}
}

func (r *factory) Validation() *validatemock.Validation {
	mockValidation := &validatemock.Validation{}
	r.app.On("MakeValidation").Return(mockValidation)

	return mockValidation
}

func (r *factory) ValidationValidator() *validatemock.Validator {
	return &validatemock.Validator{}
}

func (r *factory) ValidationErrors() *validatemock.Errors {
	return &validatemock.Errors{}
}

func (r *factory) View() *httpmock.View {
	mockView := &httpmock.View{}
	r.app.On("MakeView").Return(mockView)

	return mockView
}
