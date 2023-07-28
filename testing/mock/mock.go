package mock

import (
	accessmock "github.com/goravel/framework/contracts/auth/access/mocks"
	authmock "github.com/goravel/framework/contracts/auth/mocks"
	cachemock "github.com/goravel/framework/contracts/cache/mocks"
	configmock "github.com/goravel/framework/contracts/config/mocks"
	consolemock "github.com/goravel/framework/contracts/console/mocks"
	ormmock "github.com/goravel/framework/contracts/database/orm/mocks"
	seedermocks "github.com/goravel/framework/contracts/database/seeder/mocks"
	eventmock "github.com/goravel/framework/contracts/event/mocks"
	filesystemmock "github.com/goravel/framework/contracts/filesystem/mocks"
	foundationmock "github.com/goravel/framework/contracts/foundation/mocks"
	grpcmock "github.com/goravel/framework/contracts/grpc/mocks"
	mailmock "github.com/goravel/framework/contracts/mail/mocks"
	queuemock "github.com/goravel/framework/contracts/queue/mocks"
	validatemock "github.com/goravel/framework/contracts/validation/mocks"
	"github.com/goravel/framework/foundation"
)

var app *foundationmock.Application

func App() *foundationmock.Application {
	if app == nil {
		app = &foundationmock.Application{}
		foundation.App = app
	}

	return app
}

func Artisan() *consolemock.Artisan {
	mockArtisan := &consolemock.Artisan{}
	App().On("MakeArtisan").Return(mockArtisan)

	return mockArtisan
}

func Auth() *authmock.Auth {
	mockAuth := &authmock.Auth{}
	App().On("MakeAuth").Return(mockAuth)

	return mockAuth
}

func Cache() (*cachemock.Cache, *cachemock.Driver, *cachemock.Lock) {
	mockCache := &cachemock.Cache{}
	App().On("MakeCache").Return(mockCache)

	return mockCache, &cachemock.Driver{}, &cachemock.Lock{}
}

func Config() *configmock.Config {
	mockConfig := &configmock.Config{}
	App().On("MakeConfig").Return(mockConfig)

	return mockConfig
}

func Event() (*eventmock.Instance, *eventmock.Task) {
	mockEvent := &eventmock.Instance{}
	App().On("MakeEvent").Return(mockEvent)

	return mockEvent, &eventmock.Task{}
}

func Gate() *accessmock.Gate {
	mockGate := &accessmock.Gate{}
	App().On("MakeGate").Return(mockGate)

	return mockGate
}

func Grpc() *grpcmock.Grpc {
	mockGrpc := &grpcmock.Grpc{}
	App().On("MakeGrpc").Return(mockGrpc)

	return mockGrpc
}

func Log() {
	App().On("MakeLog").Return(NewTestLog())
}

func Mail() *mailmock.Mail {
	mockMail := &mailmock.Mail{}
	App().On("MakeMail").Return(mockMail)

	return mockMail
}

func Orm() (*ormmock.Orm, *ormmock.Query, *ormmock.Transaction, *ormmock.Association) {
	mockOrm := &ormmock.Orm{}
	App().On("MakeOrm").Return(mockOrm)

	return mockOrm, &ormmock.Query{}, &ormmock.Transaction{}, &ormmock.Association{}
}

func Queue() (*queuemock.Queue, *queuemock.Task) {
	mockQueue := &queuemock.Queue{}
	App().On("MakeQueue").Return(mockQueue)

	return mockQueue, &queuemock.Task{}
}

func Storage() (*filesystemmock.Storage, *filesystemmock.Driver, *filesystemmock.File) {
	mockStorage := &filesystemmock.Storage{}
	mockDriver := &filesystemmock.Driver{}
	mockFile := &filesystemmock.File{}
	App().On("MakeStorage").Return(mockStorage)

	return mockStorage, mockDriver, mockFile
}

func Validation() (*validatemock.Validation, *validatemock.Validator, *validatemock.Errors) {
	mockValidation := &validatemock.Validation{}
	mockValidator := &validatemock.Validator{}
	mockErrors := &validatemock.Errors{}
	App().On("MakeValidation").Return(mockValidation)

	return mockValidation, mockValidator, mockErrors
}

func Seeder() *seedermocks.Facade {
	mockSeeder := &seedermocks.Facade{}
	App().On("MakeSeeder").Return(mockSeeder)

	return mockSeeder
}
