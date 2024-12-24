package console

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/foundation"
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
	AddAboutInformation("Custom", foundation.AboutInfo{Key: "Test Info", Value: "<fg=cyan>OK</>"})
	color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, aboutCommand.Handle(mockContext))
	})
	appInformation.Range("", func(section string, details []foundation.AboutInfo) {
		assert.Contains(t, []string{"Environment", "Drivers", "Custom"}, section)
		assert.NotEmpty(t, details)
	})
}

func TestAddToSection(t *testing.T) {
	appInformation = &information{section: make(map[string]int)}
	appInformation.addToSection("Test", []foundation.AboutInfo{{Key: "Test Info", Value: "OK"}})
	assert.Equal(t, appInformation.section, map[string]int{"Test": 0})
	assert.Len(t, appInformation.details, 1)
}

func TestInformationRange(t *testing.T) {
	appInformation = &information{section: make(map[string]int)}
	appInformation.addToSection("Test", []foundation.AboutInfo{{Key: "Test Info", Value: "OK"}, {Key: "Test Info", Value: "OK"}})
	appInformation.Range("Test", func(section string, details []foundation.AboutInfo) {
		assert.Equal(t, "Test", section)
		assert.Len(t, details, 2)
		assert.Subset(t, details, []foundation.AboutInfo{{Key: "Test Info", Value: "OK"}})
	})
}
