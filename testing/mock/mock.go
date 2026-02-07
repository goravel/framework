package mock

import (
	"context"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/foundation"
	mocksauth "github.com/goravel/framework/mocks/auth"
	mocksaccess "github.com/goravel/framework/mocks/auth/access"
	mockscache "github.com/goravel/framework/mocks/cache"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mockscrypt "github.com/goravel/framework/mocks/crypt"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
	mocksseeder "github.com/goravel/framework/mocks/database/seeder"
	mocksevent "github.com/goravel/framework/mocks/event"
	mocksfilesystem "github.com/goravel/framework/mocks/filesystem"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mocksgrpc "github.com/goravel/framework/mocks/grpc"
	mockshash "github.com/goravel/framework/mocks/hash"
	mockshttp "github.com/goravel/framework/mocks/http"
	mocksmail "github.com/goravel/framework/mocks/mail"
	mocksprocess "github.com/goravel/framework/mocks/process"
	mocksqueue "github.com/goravel/framework/mocks/queue"
	mockstelemetry "github.com/goravel/framework/mocks/telemetry"
	mockstranslation "github.com/goravel/framework/mocks/translation"
	mocksvalidate "github.com/goravel/framework/mocks/validation"
	mocksview "github.com/goravel/framework/mocks/view"
	"github.com/goravel/framework/testing/utils"
)

func Factory() *factory {
	app := &mocksfoundation.Application{}
	foundation.App = app

	return &factory{
		app: app,
	}
}

type factory struct {
	app *mocksfoundation.Application
}

func (r *factory) App() *mocksfoundation.Application {
	return r.app
}

func (r *factory) Artisan() *mocksconsole.Artisan {
	mockArtisan := &mocksconsole.Artisan{}
	r.app.On("MakeArtisan").Return(mockArtisan)

	return mockArtisan
}

func (r *factory) Auth(ctx contractshttp.Context) *mocksauth.Auth {
	mockAuth := &mocksauth.Auth{}
	r.app.On("MakeAuth", ctx).Return(mockAuth)

	return mockAuth
}

func (r *factory) Cache() *mockscache.Cache {
	mockCache := &mockscache.Cache{}
	r.app.On("MakeCache").Return(mockCache)

	return mockCache
}

func (r *factory) CacheDriver() *mockscache.Driver {
	return &mockscache.Driver{}
}

func (r *factory) CacheLock() *mockscache.Lock {
	return &mockscache.Lock{}
}

func (r *factory) Context() *mockshttp.Context {
	return &mockshttp.Context{}
}

func (r *factory) ContextRequest() *mockshttp.ContextRequest {
	return &mockshttp.ContextRequest{}
}

func (r *factory) ContextResponse() *mockshttp.ContextResponse {
	return &mockshttp.ContextResponse{}
}

func (r *factory) Config() *mocksconfig.Config {
	mockConfig := &mocksconfig.Config{}
	r.app.On("MakeConfig").Return(mockConfig)

	return mockConfig
}

func (r *factory) Crypt() *mockscrypt.Crypt {
	mockCrypt := &mockscrypt.Crypt{}
	r.app.On("MakeCrypt").Return(mockCrypt)

	return mockCrypt
}

func (r *factory) Event() *mocksevent.Instance {
	mockEvent := &mocksevent.Instance{}
	r.app.On("MakeEvent").Return(mockEvent)

	return mockEvent
}

func (r *factory) EventTask() *mocksevent.Task {
	return &mocksevent.Task{}
}

func (r *factory) Gate() *mocksaccess.Gate {
	mockGate := &mocksaccess.Gate{}
	r.app.On("MakeGate").Return(mockGate)

	return mockGate
}

func (r *factory) Grpc() *mocksgrpc.Grpc {
	mockGrpc := &mocksgrpc.Grpc{}
	r.app.On("MakeGrpc").Return(mockGrpc)

	return mockGrpc
}

func (r *factory) Hash() *mockshash.Hash {
	mockHash := &mockshash.Hash{}
	r.app.On("MakeHash").Return(mockHash)

	return mockHash
}

func (r *factory) Lang(ctx context.Context) *mockstranslation.Translator {
	mockTranslator := &mockstranslation.Translator{}
	r.app.On("MakeLang", ctx).Return(mockTranslator)

	return mockTranslator
}

func (r *factory) Log() {
	r.app.On("MakeLog").Return(utils.NewTestLog())
}

func (r *factory) Mail() *mocksmail.Mail {
	mockMail := &mocksmail.Mail{}
	r.app.On("MakeMail").Return(mockMail)

	return mockMail
}

func (r *factory) Orm() *mocksorm.Orm {
	mockOrm := &mocksorm.Orm{}
	r.app.On("MakeOrm").Return(mockOrm)

	return mockOrm
}

func (r *factory) OrmAssociation() *mocksorm.Association {
	return &mocksorm.Association{}
}

func (r *factory) OrmQuery() *mocksorm.Query {
	return &mocksorm.Query{}
}

func (r *factory) OrmToSql() *mocksorm.ToSql {
	return &mocksorm.ToSql{}
}

func (r *factory) Process() *mocksprocess.Process {
	mockProcess := &mocksprocess.Process{}
	r.app.EXPECT().MakeProcess().Return(mockProcess)

	return mockProcess
}

func (r *factory) Queue() *mocksqueue.Queue {
	mockQueue := &mocksqueue.Queue{}
	r.app.On("MakeQueue").Return(mockQueue)

	return mockQueue
}

func (r *factory) QueueTask() *mocksqueue.Task {
	return &mocksqueue.Task{}
}

func (r *factory) RateLimiter() *mockshttp.RateLimiter {
	mockRateLimiter := &mockshttp.RateLimiter{}
	r.app.On("MakeRateLimiter").Return(mockRateLimiter)

	return mockRateLimiter
}

func (r *factory) Response() *mockshttp.Response {
	return &mockshttp.Response{}
}

func (r *factory) ResponseStatus() *mockshttp.ResponseStatus {
	return &mockshttp.ResponseStatus{}
}

func (r *factory) ResponseOrigin() *mockshttp.ResponseOrigin {
	return &mockshttp.ResponseOrigin{}
}

func (r *factory) ResponseView() *mockshttp.ResponseView {
	return &mockshttp.ResponseView{}
}

func (r *factory) Seeder() *mocksseeder.Facade {
	mockSeeder := &mocksseeder.Facade{}
	r.app.On("MakeSeeder").Return(mockSeeder)

	return mockSeeder
}

func (r *factory) Storage() *mocksfilesystem.Storage {
	mockStorage := &mocksfilesystem.Storage{}
	r.app.On("MakeStorage").Return(mockStorage)

	return mockStorage
}

func (r *factory) StorageDriver() *mocksfilesystem.Driver {
	return &mocksfilesystem.Driver{}
}

func (r *factory) StorageFile() *mocksfilesystem.File {
	return &mocksfilesystem.File{}
}

func (r *factory) Telemetry() *mockstelemetry.Telemetry {
	mockTelemetry := &mockstelemetry.Telemetry{}
	r.app.EXPECT().MakeTelemetry().Return(mockTelemetry)

	return mockTelemetry
}

func (r *factory) Validation() *mocksvalidate.Validation {
	mockValidation := &mocksvalidate.Validation{}
	r.app.On("MakeValidation").Return(mockValidation)

	return mockValidation
}

func (r *factory) ValidationValidator() *mocksvalidate.Validator {
	return &mocksvalidate.Validator{}
}

func (r *factory) ValidationErrors() *mocksvalidate.Errors {
	return &mocksvalidate.Errors{}
}

func (r *factory) View() *mocksview.View {
	mockView := &mocksview.View{}
	r.app.On("MakeView").Return(mockView)

	return mockView
}
