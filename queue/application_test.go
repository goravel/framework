package queue

import (
	"testing"

	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/queue/support"

	"github.com/stretchr/testify/assert"
)

func TestWorker(t *testing.T) {
	app := NewApplication()

	tests := []struct {
		description  string
		setup        func()
		args         *queue.Args
		expectWorker queue.Worker
	}{
		{
			description: "args is nil",
			setup: func() {

			},
			expectWorker: &support.Worker{
				Connection: "redis",
				Queue:      "goravel_queues:default",
				Concurrent: 1,
			},
		},
	}

	for _, test := range tests {
		test.setup()
		worker := app.Worker(test.args)
		assert.Equal(t, test.expectWorker, worker, test.description)
	}
}
