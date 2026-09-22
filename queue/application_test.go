package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	mocksqueue "github.com/goravel/framework/mocks/queue"
)

func TestApplication_WorkerUsesConfiguredQueueList(t *testing.T) {
	t.Run("default connection", func(t *testing.T) {
		mockConfig := mocksqueue.NewConfig(t)
		mockConfig.EXPECT().DefaultConnection().Return("sync").Once()
		mockConfig.EXPECT().DefaultConcurrent().Return(2).Once()
		mockConfig.EXPECT().GetString("queue.connections.sync.queue", "default").Return("high,default").Once()
		mockConfig.EXPECT().Driver("sync").Return(contractsqueue.DriverSync).Once()
		mockConfig.EXPECT().Debug().Return(false).Once()

		worker := NewApplication(mockConfig, nil, nil, nil, nil, nil).Worker().(*Worker)

		assert.Equal(t, "high,default", worker.queue)
		assert.Equal(t, 2, worker.concurrent)
	})

	t.Run("selected connection", func(t *testing.T) {
		mockConfig := mocksqueue.NewConfig(t)
		mockConfig.EXPECT().DefaultConnection().Return("sync").Once()
		mockConfig.EXPECT().DefaultConcurrent().Return(1).Once()
		mockConfig.EXPECT().GetString("queue.connections.redis.queue", "default").Return("high,default").Once()
		mockConfig.EXPECT().GetInt("queue.connections.redis.concurrent", 1).Return(3).Once()
		mockConfig.EXPECT().Driver("redis").Return(contractsqueue.DriverSync).Once()
		mockConfig.EXPECT().Debug().Return(false).Once()

		worker := NewApplication(mockConfig, nil, nil, nil, nil, nil).Worker(contractsqueue.Args{
			Connection: "redis",
		}).(*Worker)

		assert.Equal(t, "high,default", worker.queue)
		assert.Equal(t, 3, worker.concurrent)
	})
}
