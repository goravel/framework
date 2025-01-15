package console

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/foundation"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	"github.com/goravel/framework/support/color"
)

func TestAboutCommand(t *testing.T) {
	mockApp := mocksfoundation.NewApplication(t)
	mockConfig := mocksconfig.NewConfig(t)
	mockApp.EXPECT().MakeConfig().Return(mockConfig).Once()
	mockApp.EXPECT().Version().Return("test_version").Once()
	mockConfig.EXPECT().GetString("app.name").Return("test").Once()
	mockConfig.EXPECT().GetString("app.env").Return("test").Once()
	mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
	mockConfig.EXPECT().GetString("http.url").Return("test_url").Once()
	mockConfig.EXPECT().GetString("http.host").Return("test_host").Once()
	mockConfig.EXPECT().GetString("http.port").Return("test_port").Once()
	mockConfig.EXPECT().GetString("grpc.host").Return("test_host").Once()
	mockConfig.EXPECT().GetString("grpc.port").Return("test_port").Once()
	mockConfig.EXPECT().GetString("cache.default").Return("test_cache").Once()
	mockConfig.EXPECT().GetString("database.default").Return("test_database").Once()
	mockConfig.EXPECT().GetString("hashing.driver").Return("test_hashing").Once()
	mockConfig.EXPECT().GetString("http.default").Return("test_http").Once()
	mockConfig.EXPECT().GetString("logging.default").Return("stack").Once()
	mockConfig.EXPECT().GetString("logging.channels.stack.driver").Return("stack").Once()
	mockConfig.EXPECT().Get("logging.channels.stack.channels").Return([]string{"test"}).Once()
	mockConfig.EXPECT().GetString("mail.default", "smtp").Return("test_mail").Once()
	mockConfig.EXPECT().GetString("queue.default").Return("test_queue").Once()
	mockConfig.EXPECT().GetString("session.driver").Return("test_session").Once()
	aboutCommand := NewAboutCommand(mockApp)
	mockContext := &mocksconsole.Context{}
	mockContext.EXPECT().NewLine().Return().Times(4)
	mockContext.EXPECT().Option("only").Return("").Once()
	getGoVersion = func() string {
		return "test_version"
	}
	var expected []string
	for _, ex := range [][2]string{
		{"<fg=green;op=bold>Environment</>", ""},
		{"Application Name", "test"},
		{"Goravel Version", "test_version"},
		{"Go Version", "test_version"},
		{"Environment", "test"},
		{"Debug Mode", "<fg=yellow;op=bold>ENABLED</>"},
		{"URL", "test_url"},
		{"HTTP Host", "test_host"},
		{"HTTP Port", "test_port"},
		{"GRPC Host", "test_host"},
		{"GRPC Port", "test_port"},
		{"<fg=green;op=bold>Drivers</>", ""},
		{"Cache", "test_cache"},
		{"Database", "test_database"},
		{"Hashing", "test_hashing"},
		{"Http", "test_http"},
		{"Logs", "<fg=yellow;op=bold>stack</> <fg=gray;op=bold>/</> test"},
		{"Mail", "test_mail"},
		{"Queue", "test_queue"},
		{"Session", "test_session"},
		{"<fg=green;op=bold>Custom</>", ""},
		{"Test Info", "<fg=cyan>OK</>"},
	} {
		mockContext.EXPECT().TwoColumnDetail(ex[0], ex[1]).
			Run(func(first string, second string, _ ...rune) {
				expected = append(expected, color.Default().Sprintf("%s %s\n", first, second))
				color.Default().Printf("%s %s\n", first, second)
			}).Return().Once()
	}
	AddAboutInformation("Custom", foundation.AboutItem{Key: "Test Info", Value: "<fg=cyan>OK</>"})
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, aboutCommand.Handle(mockContext))
	}), strings.Join(expected, ""))
}

func TestAddToSection(t *testing.T) {
	appInformation = &information{section: make(map[string]int)}
	appInformation.addToSection("Test", []foundation.AboutItem{{Key: "Test Info", Value: "OK"}})
	assert.Equal(t, appInformation.section, map[string]int{"Test": 0})
	assert.Len(t, appInformation.details, 1)
}

func TestInformationRange(t *testing.T) {
	appInformation = &information{section: make(map[string]int)}
	appInformation.addToSection("Test", []foundation.AboutItem{{Key: "Test Info", Value: "OK"}, {Key: "Test Info", Value: "OK"}})
	appInformation.Range("Test", func(section string, details []foundation.AboutItem) {
		assert.Equal(t, "Test", section)
		assert.Len(t, details, 2)
		assert.Subset(t, details, []foundation.AboutItem{{Key: "Test Info", Value: "OK"}})
	})
}
