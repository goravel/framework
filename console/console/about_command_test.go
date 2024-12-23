package console

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocksconfig "github.com/goravel/framework/mocks/config"
	consolemocks "github.com/goravel/framework/mocks/console"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	"github.com/goravel/framework/support/color"
)

func TestAboutCommand(t *testing.T) {
	mockApp := mocksfoundation.NewApplication(t)
	mockConfig := mocksconfig.NewConfig(t)
	mockApp.EXPECT().MakeConfig().Return(mockConfig).Once()
	mockApp.EXPECT().Version().Return("")
	mockConfig.EXPECT().GetString("logging.default").Return("stack").Once()
	mockConfig.EXPECT().GetString("logging.channels.stack.driver").Return("stack").Once()
	mockConfig.EXPECT().Get("logging.channels.stack.channels").Return([]string{"test"}).Once()
	mockConfig.EXPECT().GetString(mock.Anything).Return("")
	mockConfig.EXPECT().GetString(mock.Anything, mock.Anything).Return("")
	mockConfig.EXPECT().GetBool(mock.Anything).Return(true)
	aboutCommand := NewAboutCommand(mockApp)
	mockContext := &consolemocks.Context{}
	mockContext.EXPECT().NewLine().Return()
	mockContext.EXPECT().Option("only").Return("").Once()
	mockContext.EXPECT().TwoColumnDetail(mock.Anything, mock.Anything).Return()
	AddAboutInformation("Custom", "Test Info", "<fg=cyan>OK</>")
	color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, aboutCommand.Handle(mockContext))
	})
	appInformation.Range("", func(section string, details []kv) {
		assert.Contains(t, []string{"Environment", "Drivers", "Custom"}, section)
		assert.NotEmpty(t, details)
	})
}
