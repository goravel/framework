package mock

import (
	accessmocks "github.com/goravel/framework/contracts/auth/access/mocks"
	authmocks "github.com/goravel/framework/contracts/auth/mocks"
	cachemocks "github.com/goravel/framework/contracts/cache/mocks"
	configmocks "github.com/goravel/framework/contracts/config/mocks"
	consolemocks "github.com/goravel/framework/contracts/console/mocks"
	cryptmocks "github.com/goravel/framework/contracts/crypt/mocks"
	ormmocks "github.com/goravel/framework/contracts/database/orm/mocks"
	seedermocks "github.com/goravel/framework/contracts/database/seeder/mocks"
	eventmocks "github.com/goravel/framework/contracts/event/mocks"
	filesystemmocks "github.com/goravel/framework/contracts/filesystem/mocks"
	foundationmocks "github.com/goravel/framework/contracts/foundation/mocks"
	grpcmocks "github.com/goravel/framework/contracts/grpc/mocks"
	hashmocks "github.com/goravel/framework/contracts/hash/mocks"
	httpmocks "github.com/goravel/framework/contracts/http/mocks"
	mailmocks "github.com/goravel/framework/contracts/mail/mocks"
	queuemocks "github.com/goravel/framework/contracts/queue/mocks"
	validatemocks "github.com/goravel/framework/contracts/validation/mocks"
	"github.com/goravel/framework/foundation"
)

var app *foundationmocks.Application

func App() *foundationmocks.Application {
	if app == nil {
		app = &foundationmocks.Application{}
		foundation.App = app
	}

	return app
}

func Artisan() *consolemocks.Artisan {
	mockArtisan := &consolemocks.Artisan{}
	App().On("MakeArtisan").Return(mockArtisan)

	return mockArtisan
}

func Auth() *authmocks.Auth {
	mockAuth := &authmocks.Auth{}
	App().On("MakeAuth").Return(mockAuth)

	return mockAuth
}

func Cache() (*cachemocks.Cache, *cachemocks.Driver, *cachemocks.Lock) {
	mockCache := &cachemocks.Cache{}
	App().On("MakeCache").Return(mockCache)

	return mockCache, &cachemocks.Driver{}, &cachemocks.Lock{}
}

func Config() *configmocks.Config {
	mockConfig := &configmocks.Config{}
	App().On("MakeConfig").Return(mockConfig)

	return mockConfig
}

func Crypt() *cryptmocks.Crypt {
	mockCrypt := &cryptmocks.Crypt{}
	App().On("MakeCrypt").Return(mockCrypt)

	return mockCrypt
}

func Event() (*eventmocks.Instance, *eventmocks.Task) {
	mockEvent := &eventmocks.Instance{}
	App().On("MakeEvent").Return(mockEvent)

	return mockEvent, &eventmocks.Task{}
}

func Gate() *accessmocks.Gate {
	mockGate := &accessmocks.Gate{}
	App().On("MakeGate").Return(mockGate)

	return mockGate
}

func Grpc() *grpcmocks.Grpc {
	mockGrpc := &grpcmocks.Grpc{}
	App().On("MakeGrpc").Return(mockGrpc)

	return mockGrpc
}

func Hash() *hashmocks.Hash {
	mockHash := &hashmocks.Hash{}
	App().On("MakeHash").Return(mockHash)

	return mockHash
}

func Log() {
	App().On("MakeLog").Return(NewTestLog())
}

func Mail() *mailmocks.Mail {
	mockMail := &mailmocks.Mail{}
	App().On("MakeMail").Return(mockMail)

	return mockMail
}

func Orm() (*ormmocks.Orm, *ormmocks.Query, *ormmocks.Transaction, *ormmocks.Association) {
	mockOrm := &ormmocks.Orm{}
	App().On("MakeOrm").Return(mockOrm)

	return mockOrm, &ormmocks.Query{}, &ormmocks.Transaction{}, &ormmocks.Association{}
}

func Queue() (*queuemocks.Queue, *queuemocks.Task) {
	mockQueue := &queuemocks.Queue{}
	App().On("MakeQueue").Return(mockQueue)

	return mockQueue, &queuemocks.Task{}
}

func RateLimiter() *httpmocks.RateLimiter {
	mockRateLimiter := &httpmocks.RateLimiter{}
	App().On("MakeRateLimiter").Return(mockRateLimiter)

	return mockRateLimiter
}

func Storage() (*filesystemmocks.Storage, *filesystemmocks.Driver, *filesystemmocks.File) {
	mockStorage := &filesystemmocks.Storage{}
	mockDriver := &filesystemmocks.Driver{}
	mockFile := &filesystemmocks.File{}
	App().On("MakeStorage").Return(mockStorage)

	return mockStorage, mockDriver, mockFile
}

func Validation() (*validatemocks.Validation, *validatemocks.Validator, *validatemocks.Errors) {
	mockValidation := &validatemocks.Validation{}
	mockValidator := &validatemocks.Validator{}
	mockErrors := &validatemocks.Errors{}
	App().On("MakeValidation").Return(mockValidation)

	return mockValidation, mockValidator, mockErrors
}

func View() *httpmocks.View {
	mockView := &httpmocks.View{}
	App().On("MakeView").Return(mockView)

	return mockView
}

func Seeder() *seedermocks.Facade {
	mockSeeder := &seedermocks.Facade{}
	App().On("MakeSeeder").Return(mockSeeder)

	return mockSeeder
}
