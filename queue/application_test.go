package queue

import (
	"testing"

	configmocks "github.com/goravel/framework/contracts/config/mocks"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/queue/support"
	"github.com/goravel/framework/testing/mock"

	"github.com/stretchr/testify/assert"
)

func TestWorker(t *testing.T) {
	var (
		mockConfig *configmocks.Config
		app        = NewApplication()
	)

	beforeEach := func() {
		mockConfig = mock.Config()
	}

	tests := []struct {
		description  string
		setup        func()
		args         *queue.Args
		expectWorker queue.Worker
	}{
		{
			description: "success when args is nil",
			setup: func() {
				mockConfig.On("GetString", "queue.default").Return("redis").Once()
				mockConfig.On("GetString", "app.name").Return("app").Once()
				mockConfig.On("GetString", "queue.connections.redis.queue", "default").Return("queue").Once()

			},
			expectWorker: &support.Worker{
				Connection: "redis",
				Queue:      "app_queues:queue",
				Concurrent: 1,
			},
		},
		{
			description: "success when args isn't nil",
			setup: func() {
				mockConfig.On("GetString", "app.name").Return("app").Once()
			},
			args: &queue.Args{
				Connection: "redis",
				Queue:      "queue",
				Concurrent: 2,
			},
			expectWorker: &support.Worker{
				Connection: "redis",
				Queue:      "app_queues:queue",
				Concurrent: 2,
			},
		},
	}

	for _, test := range tests {
		beforeEach()
		test.setup()
		worker := app.Worker(test.args)
		assert.Equal(t, test.expectWorker, worker, test.description)
	}
}
