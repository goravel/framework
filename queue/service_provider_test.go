package queue

import (
	"testing"

	"github.com/stretchr/testify/mock"

	contractsconsole "github.com/goravel/framework/contracts/console"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mocksqueue "github.com/goravel/framework/mocks/queue"
)

func TestServiceProviderBoot(t *testing.T) {
	provider := &ServiceProvider{}
	app := mocksfoundation.NewApplication(t)
	queue := mocksqueue.NewQueue(t)

	app.EXPECT().MakeQueue().Return(queue).Twice()
	app.EXPECT().GetJson().Return(nil).Once()
	app.EXPECT().Commands(mock.MatchedBy(func(commands []contractsconsole.Command) bool {
		if len(commands) != 3 {
			return false
		}

		return commands[0].Signature() == "make:job" &&
			commands[1].Signature() == "queue:retry" &&
			commands[2].Signature() == "queue:failed"
	})).Once()

	provider.Boot(app)
}
