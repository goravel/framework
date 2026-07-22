package broadcasting

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/broadcasting"
	mockslog "github.com/goravel/framework/mocks/log"
)

func TestLogDriver_Broadcast(t *testing.T) {
	mockLog := mockslog.NewLog(t)
	mockWriter := mockslog.NewWriter(t)

	channels := []broadcasting.Channel{
		{Name: "chan-a"},
		{Name: "chan-b"},
	}
	event := "test.event"
	payload := map[string]any{"key": "value"}

	mockLog.EXPECT().With(mock.MatchedBy(func(data map[string]any) bool {
		return data["event"] == event &&
			data["payload"] != nil &&
			data["channels"] != nil
	})).Return(mockWriter).Once()
	mockWriter.EXPECT().Info("Broadcasting event").Return().Once()

	driver := NewLogDriver(mockLog)
	err := driver.Broadcast(channels, event, payload)

	assert.NoError(t, err)
}
