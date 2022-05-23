package support

import (
	"fmt"
	"github.com/RichardKnop/machinery/v2"
	redisBackend "github.com/RichardKnop/machinery/v2/backends/redis"
	redisBroker "github.com/RichardKnop/machinery/v2/brokers/redis"
	"github.com/RichardKnop/machinery/v2/config"
	"github.com/RichardKnop/machinery/v2/locks/eager"
	"github.com/gookit/color"
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/support/facades"
)

func getServer(connection string, queue string) (*machinery.Server, error) {
	driver := getDriver(connection)

	switch driver {
	case DriverSync:
		color.Yellowln("Queue sync driver doesn't need to be run")

		return nil, nil
	case DriverRedis:
		return getRedisServer(connection, queue), nil
	}

	return nil, fmt.Errorf("unknow queue driver: %s", driver)
}

func getDriver(connection string) string {
	return facades.Config.GetString(fmt.Sprintf("queue.connections.%s.driver", connection))
}

func getRedisServer(connection string, queue string) *machinery.Server {
	redisConfig, database, defaultQueue := getRedisConfig(connection)
	if queue == "" {
		queue = defaultQueue
	}

	cnf := &config.Config{
		DefaultQueue: queue,
		Redis:        &config.RedisConfig{},
	}

	broker := redisBroker.NewGR(cnf, []string{redisConfig}, database)
	backend := redisBackend.NewGR(cnf, []string{redisConfig}, database)
	lock := eager.New()

	return machinery.NewServer(cnf, broker, backend, lock)
}

func getRedisConfig(queueConnection string) (config string, database int, queue string) {
	connection := facades.Config.GetString(fmt.Sprintf("queue.connections.%s.connection", queueConnection))
	queue = facades.Config.GetString(fmt.Sprintf("queue.connections.%s.queue", queueConnection), "default")
	host := facades.Config.GetString(fmt.Sprintf("database.redis.%s.host", connection))
	password := facades.Config.GetString(fmt.Sprintf("database.redis.%s.password", connection))
	port := facades.Config.GetString(fmt.Sprintf("database.redis.%s.port", connection))
	database = facades.Config.GetInt(fmt.Sprintf("database.redis.%s.database", connection))

	if password == "" {
		config = host + ":" + port
	} else {
		config = password + "@" + host + ":" + port
	}

	return
}

func jobs2Tasks(jobs []queue.Job) map[string]interface{} {
	tasks := make(map[string]interface{}, len(jobs))

	for _, job := range jobs {
		tasks[job.Signature()] = job.Handle
	}

	return tasks
}
