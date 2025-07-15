package console

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/contracts/foundation"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	"github.com/goravel/framework/support/color"
)

type AboutCommandTestSuite struct {
	suite.Suite
}

func (s *AboutCommandTestSuite) SetupTest() {
}

func (s *AboutCommandTestSuite) TearDownTest() {
	appInformation = &information{section: make(map[string]int)}
}

func TestAboutCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AboutCommandTestSuite))
}

func (s *AboutCommandTestSuite) TestSignature() {
	cmd := &AboutCommand{}
	expected := "about"
	s.Require().Equal(expected, cmd.Signature())
}

func (s *AboutCommandTestSuite) TestDescription() {
	cmd := &AboutCommand{}
	expected := "Display basic information about your application"
	s.Require().Equal(expected, cmd.Description())
}

func (s *AboutCommandTestSuite) TestExtend() {
	cmd := &AboutCommand{}
	got := cmd.Extend()

	if len(got.Flags) > 0 {
		s.Run("should have correctly configured StringFlag", func() {
			flag, ok := got.Flags[0].(*command.StringFlag)
			if !ok {
				s.Fail("First flag is not StringFlag (got type: %T)", got.Flags[0])
			}

			testCases := []struct {
				name     string
				got      any
				expected any
			}{
				{"Name", flag.Name, "only"},
				{"Usage", flag.Usage, "The section to display"},
			}

			for _, tc := range testCases {
				if !reflect.DeepEqual(tc.got, tc.expected) {
					s.Require().Equal(tc.expected, tc.got)
				}
			}
		})
	}
}

func (s *AboutCommandTestSuite) TestHandle() {
	mockApp := mocksfoundation.NewApplication(s.T())
	mockConfig := mocksconfig.NewConfig(s.T())
	mockContext := mocksconsole.NewContext(s.T())

	cmd := NewAboutCommand(mockApp)

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
	mockConfig.EXPECT().GetString("session.default").Return("test_session").Once()

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
	s.Contains(color.CaptureOutput(func(w io.Writer) {
		s.Nil(cmd.Handle(mockContext))
	}), strings.Join(expected, ""), "output should contain expected lines in order")
}

func (s *AboutCommandTestSuite) TestAddToSection() {
	appInformation = &information{section: make(map[string]int)}
	appInformation.addToSection("Test", []foundation.AboutItem{{Key: "Test Info", Value: "OK"}})
	s.Equal(appInformation.section, map[string]int{"Test": 0})
	s.Len(appInformation.details, 1)
}

func (s *AboutCommandTestSuite) TestInformationRange() {
	appInformation = &information{section: make(map[string]int)}
	appInformation.addToSection("Test", []foundation.AboutItem{{Key: "Test Info", Value: "OK"}, {Key: "Test Info", Value: "OK"}})
	appInformation.Range("Test", func(section string, details []foundation.AboutItem) {
		s.Equal("Test", section)
		s.Len(details, 2)
		s.Subset(details, []foundation.AboutItem{{Key: "Test Info", Value: "OK"}})
	})
}
