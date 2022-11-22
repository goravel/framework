package mock

import (
	cachemocks "github.com/goravel/framework/contracts/cache/mocks"
	configmocks "github.com/goravel/framework/contracts/config/mocks"
	consolemocks "github.com/goravel/framework/contracts/console/mocks"
	ormmocks "github.com/goravel/framework/contracts/database/orm/mocks"
	eventmocks "github.com/goravel/framework/contracts/event/mocks"
	filesystemmocks "github.com/goravel/framework/contracts/filesystem/mocks"
	mailmocks "github.com/goravel/framework/contracts/mail/mocks"
	queuemocks "github.com/goravel/framework/contracts/queue/mocks"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/log"
)

func Cache() *cachemocks.Store {
	mockCache := &cachemocks.Store{}
	facades.Cache = mockCache

	return mockCache
}

func Config() *configmocks.Config {
	mockConfig := &configmocks.Config{}
	facades.Config = mockConfig

	return mockConfig
}

func Artisan() *consolemocks.Artisan {
	mockArtisan := &consolemocks.Artisan{}
	facades.Artisan = mockArtisan

	return mockArtisan
}

func Orm() (*ormmocks.Orm, *ormmocks.DB, *ormmocks.Transaction) {
	mockOrm := &ormmocks.Orm{}
	facades.Orm = mockOrm

	return mockOrm, &ormmocks.DB{}, &ormmocks.Transaction{}
}

func Event() (*eventmocks.Instance, *eventmocks.Task) {
	mockEvent := &eventmocks.Instance{}
	facades.Event = mockEvent

	return mockEvent, &eventmocks.Task{}
}

func Log() {
	facades.Log = log.NewLogrus(nil, log.NewTestWriter())
}

func Mail() *mailmocks.Mail {
	mockMail := &mailmocks.Mail{}
	facades.Mail = mockMail

	return mockMail
}

func Queue() (*queuemocks.Queue, *queuemocks.Task) {
	mockQueue := &queuemocks.Queue{}
	facades.Queue = mockQueue

	return mockQueue, &queuemocks.Task{}
}

func Storage() (*filesystemmocks.Storage, *filesystemmocks.Driver, *filesystemmocks.File) {
	mockStorage := &filesystemmocks.Storage{}
	mockDriver := &filesystemmocks.Driver{}
	mockFile := &filesystemmocks.File{}
	facades.Storage = mockStorage

	return mockStorage, mockDriver, mockFile
}
