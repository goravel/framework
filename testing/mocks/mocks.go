package mocks

import (
	mockcache "github.com/goravel/framework/contracts/cache/mocks"
	mockconfig "github.com/goravel/framework/contracts/config/mocks"
	mockconsole "github.com/goravel/framework/contracts/console/mocks"
	mockorm "github.com/goravel/framework/contracts/database/orm/mocks"
	mockevent "github.com/goravel/framework/contracts/events/mocks"
	mockqueue "github.com/goravel/framework/contracts/queue/mocks"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/log"
)

func Cache() *mockcache.Store {
	mockCache := &mockcache.Store{}
	facades.Cache = mockCache

	return mockCache
}

func Config() *mockconfig.Config {
	mockConfig := &mockconfig.Config{}
	facades.Config = mockConfig

	return mockConfig
}

func Console() *mockconsole.Artisan {
	mockArtisan := &mockconsole.Artisan{}
	facades.Artisan = mockArtisan

	return mockArtisan
}

func Orm() (*mockorm.Orm, *mockorm.DB, *mockorm.Transaction) {
	mockOrm := &mockorm.Orm{}
	mockOrmDB := &mockorm.DB{}
	mockOrmTransaction := &mockorm.Transaction{}

	facades.Orm = mockOrm

	return mockOrm, mockOrmDB, mockOrmTransaction
}

func Event() (*mockevent.Instance, *mockevent.Task) {
	mockEvent := &mockevent.Instance{}
	mockTask := &mockevent.Task{}
	facades.Event = mockEvent

	return mockEvent, mockTask
}

func Log() {
	facades.Log = &log.Log{Instance: nil, Test: true}
}

func Queue() (*mockqueue.Queue, *mockqueue.Task) {
	mockQueue := &mockqueue.Queue{}
	mockTask := &mockqueue.Task{}
	facades.Queue = mockQueue

	return mockQueue, mockTask
}
