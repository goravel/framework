package mock

import (
	"context"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/foundation"
	authmocks "github.com/goravel/framework/mocks/auth"
	accessmocks "github.com/goravel/framework/mocks/auth/access"
	cachemocks "github.com/goravel/framework/mocks/cache"
	configmocks "github.com/goravel/framework/mocks/config"
	consolemocks "github.com/goravel/framework/mocks/console"
	cryptmocks "github.com/goravel/framework/mocks/crypt"
	ormmocks "github.com/goravel/framework/mocks/database/orm"
	seedermocks "github.com/goravel/framework/mocks/database/seeder"
	eventmocks "github.com/goravel/framework/mocks/event"
	filesystemmocks "github.com/goravel/framework/mocks/filesystem"
	foundationmocks "github.com/goravel/framework/mocks/foundation"
	grpcmocks "github.com/goravel/framework/mocks/grpc"
	hashmocks "github.com/goravel/framework/mocks/hash"
	httpmocks "github.com/goravel/framework/mocks/http"
	mailmocks "github.com/goravel/framework/mocks/mail"
	queuemocks "github.com/goravel/framework/mocks/queue"
	translationmocks "github.com/goravel/framework/mocks/translation"
	validatemocks "github.com/goravel/framework/mocks/validation"
)

func Factory() *factory {
	app := &foundationmocks.Application{}
	foundation.App = app

	return &factory{
		app: app,
	}
}

type factory struct {
	app *foundationmocks.Application
}

func (r *factory) App() *foundationmocks.Application {
	return r.app
}

func (r *factory) Artisan() *consolemocks.Artisan {
	mockArtisan := &consolemocks.Artisan{}
	r.app.On("MakeArtisan").Return(mockArtisan)

	return mockArtisan
}

func (r *factory) Auth(ctx contractshttp.Context) *authmocks.Auth {
	mockAuth := &authmocks.Auth{}
	r.app.On("MakeAuth", ctx).Return(mockAuth)

	return mockAuth
}

func (r *factory) Cache() *cachemocks.Cache {
	mockCache := &cachemocks.Cache{}
	r.app.On("MakeCache").Return(mockCache)

	return mockCache
}

func (r *factory) CacheDriver() *cachemocks.Driver {
	return &cachemocks.Driver{}
}

func (r *factory) CacheLock() *cachemocks.Lock {
	return &cachemocks.Lock{}
}

func (r *factory) Config() *configmocks.Config {
	mockConfig := &configmocks.Config{}
	r.app.On("MakeConfig").Return(mockConfig)

	return mockConfig
}

func (r *factory) Crypt() *cryptmocks.Crypt {
	mockCrypt := &cryptmocks.Crypt{}
	r.app.On("MakeCrypt").Return(mockCrypt)

	return mockCrypt
}

func (r *factory) Event() *eventmocks.Instance {
	mockEvent := &eventmocks.Instance{}
	r.app.On("MakeEvent").Return(mockEvent)

	return mockEvent
}

func (r *factory) EventTask() *eventmocks.Task {
	return &eventmocks.Task{}
}

func (r *factory) Gate() *accessmocks.Gate {
	mockGate := &accessmocks.Gate{}
	r.app.On("MakeGate").Return(mockGate)

	return mockGate
}

func (r *factory) Grpc() *grpcmocks.Grpc {
	mockGrpc := &grpcmocks.Grpc{}
	r.app.On("MakeGrpc").Return(mockGrpc)

	return mockGrpc
}

func (r *factory) Hash() *hashmocks.Hash {
	mockHash := &hashmocks.Hash{}
	r.app.On("MakeHash").Return(mockHash)

	return mockHash
}

func (r *factory) Lang(ctx context.Context) *translationmocks.Translator {
	mockTranslator := &translationmocks.Translator{}
	r.app.On("MakeLang", ctx).Return(mockTranslator)

	return mockTranslator
}

func (r *factory) Log() {
	r.app.On("MakeLog").Return(NewTestLog())
}

func (r *factory) Mail() *mailmocks.Mail {
	mockMail := &mailmocks.Mail{}
	r.app.On("MakeMail").Return(mockMail)

	return mockMail
}

func (r *factory) Orm() *ormmocks.Orm {
	mockOrm := &ormmocks.Orm{}
	r.app.On("MakeOrm").Return(mockOrm)

	return mockOrm
}

func (r *factory) OrmAssociation() *ormmocks.Association {
	return &ormmocks.Association{}
}

func (r *factory) OrmQuery() *ormmocks.Query {
	return &ormmocks.Query{}
}

func (r *factory) OrmTransaction() *ormmocks.Transaction {
	return &ormmocks.Transaction{}
}

func (r *factory) Queue() *queuemocks.Queue {
	mockQueue := &queuemocks.Queue{}
	r.app.On("MakeQueue").Return(mockQueue)

	return mockQueue
}

func (r *factory) QueueTask() *queuemocks.Task {
	return &queuemocks.Task{}
}

func (r *factory) RateLimiter() *httpmocks.RateLimiter {
	mockRateLimiter := &httpmocks.RateLimiter{}
	r.app.On("MakeRateLimiter").Return(mockRateLimiter)

	return mockRateLimiter
}

func (r *factory) Seeder() *seedermocks.Facade {
	mockSeeder := &seedermocks.Facade{}
	r.app.On("MakeSeeder").Return(mockSeeder)

	return mockSeeder
}

func (r *factory) Storage() *filesystemmocks.Storage {
	mockStorage := &filesystemmocks.Storage{}
	r.app.On("MakeStorage").Return(mockStorage)

	return mockStorage
}

func (r *factory) StorageDriver() *filesystemmocks.Driver {
	return &filesystemmocks.Driver{}
}

func (r *factory) StorageFile() *filesystemmocks.File {
	return &filesystemmocks.File{}
}

func (r *factory) Validation() *validatemocks.Validation {
	mockValidation := &validatemocks.Validation{}
	r.app.On("MakeValidation").Return(mockValidation)

	return mockValidation
}

func (r *factory) ValidationValidator() *validatemocks.Validator {
	return &validatemocks.Validator{}
}

func (r *factory) ValidationErrors() *validatemocks.Errors {
	return &validatemocks.Errors{}
}

func (r *factory) View() *httpmocks.View {
	mockView := &httpmocks.View{}
	r.app.On("MakeView").Return(mockView)

	return mockView
}
