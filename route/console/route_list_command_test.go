package commands

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/route"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksroute "github.com/goravel/framework/mocks/route"
	"github.com/goravel/framework/support/color"
	supportconsole "github.com/goravel/framework/support/console"
)

func TestRouteListCommand(t *testing.T) {
	var (
		mockContext *mocksconsole.Context
		mockRoute   *mocksroute.Route
	)

	beforeEach := func() {
		mockContext = mocksconsole.NewContext(t)
		mockRoute = mocksroute.NewRoute(t)
	}
	tests := []struct {
		name     string
		setup    func()
		expected string
	}{
		{
			name: "no routes",
			setup: func() {
				mockContext.EXPECT().NewLine().Return().Once()
				mockRoute.EXPECT().GetRoutes().Return(nil).Once()
				mockContext.EXPECT().Warning("Your application doesn't have any routes.").
					Run(func(msg string) {
						color.Warningln(msg)
					}).Once()
			},
			expected: "\x1b[30;43m\x1b[30;43m WARNING \x1b[0m\x1b[0m \x1b[33m\x1b[33mYour application doesn't have any routes.\x1b[0m\x1b[0m\n",
		},
		{
			name: "no routes matching criteria",
			setup: func() {
				mockContext.EXPECT().NewLine().Return().Once()
				mockRoute.EXPECT().GetRoutes().Return([]route.Info{
					{Name: "test", Method: "GET", Path: "/test"},
					{Name: "test2", Method: "POST", Path: "/test2"},
				}).Once()
				mockContext.EXPECT().Option("method").Return("").Once()
				mockContext.EXPECT().Option("name").Return("").Once()
				mockContext.EXPECT().Option("path").Return("").Once()
				mockContext.EXPECT().OptionSlice("except-path").Return([]string{"test"}).Once()
				mockContext.EXPECT().Warning("Your application doesn't have any routes matching the given criteria.").
					Run(func(msg string) {
						color.Warningln(msg)
					}).Once()
			},
			expected: "\x1b[30;43m\x1b[30;43m WARNING \x1b[0m\x1b[0m \x1b[33m\x1b[33mYour application doesn't have any routes matching the given criteria.\x1b[0m\x1b[0m\n",
		},
		{
			name: "filter by method",
			setup: func() {
				mockContext.EXPECT().NewLine().Return().Once()
				mockRoute.EXPECT().GetRoutes().Return([]route.Info{
					{Name: "test", Method: "GET", Path: "/test"},
					{Name: "test2", Method: "POST", Path: "/test2"},
				}).Once()
				mockContext.EXPECT().Option("method").Return("POST").Once()
				mockContext.EXPECT().Option("name").Return("").Once()
				mockContext.EXPECT().Option("path").Return("").Once()
				mockContext.EXPECT().OptionSlice("except-path").Return(nil).Once()
				mockContext.EXPECT().TwoColumnDetail("<fg=yellow>POST</>         test2", "test2").
					Run(func(first string, second string, filler ...rune) {
						color.Default().Println(supportconsole.TwoColumnDetail(first, second, filler...))
					}).Once()
				mockContext.EXPECT().NewLine().Return().Once()
				mockContext.EXPECT().TwoColumnDetail("", "<fg=blue;op=bold>Showing [1] routes</>", ' ').
					Run(func(first string, second string, filler ...rune) {
						color.Default().Println(supportconsole.TwoColumnDetail(first, second, filler...))
					}).Once()
			},
			expected: "\x1b[39m  \x1b[33mPOST\x1b[0m\x1b[39m         test2 \x1b[90m...................................................\x1b[0m\x1b[39m test2  " +
				"\x1b[0m\n\x1b[39m\x1b[0m\x1b[39m   \x1b[90m                                                        " +
				"\x1b[0m\x1b[39m \x1b[34;1mShowing [1] routes\x1b[0m\x1b[39m  \x1b[0m\n\x1b[39m\x1b[0m",
		},
		{
			name: "filter by path",
			setup: func() {
				mockContext.EXPECT().NewLine().Return().Once()
				mockRoute.EXPECT().GetRoutes().Return([]route.Info{
					{Name: "test", Method: "GET", Path: "/test"},
					{Name: "test2", Method: "POST", Path: "/test2"},
				}).Once()
				mockContext.EXPECT().Option("method").Return("").Once()
				mockContext.EXPECT().Option("name").Return("").Once()
				mockContext.EXPECT().Option("path").Return("test").Once()
				mockContext.EXPECT().OptionSlice("except-path").Return(nil).Once()
				mockContext.EXPECT().TwoColumnDetail("<fg=blue>GET</>          test", "test").
					Run(func(first string, second string, filler ...rune) {
						color.Default().Println(supportconsole.TwoColumnDetail(first, second, filler...))
					}).Once()
				mockContext.EXPECT().TwoColumnDetail("<fg=yellow>POST</>         test2", "test2").
					Run(func(first string, second string, filler ...rune) {
						color.Default().Println(supportconsole.TwoColumnDetail(first, second, filler...))
					}).Once()
				mockContext.EXPECT().NewLine().Return().Once()
				mockContext.EXPECT().TwoColumnDetail("", "<fg=blue;op=bold>Showing [2] routes</>", ' ').
					Run(func(first string, second string, filler ...rune) {
						color.Default().Println(supportconsole.TwoColumnDetail(first, second, filler...))
					}).Once()
			},
			expected: "\x1b[39m  \x1b[34mGET\x1b[0m\x1b[39m          test \x1b[90m.....................................................\x1b[0m\x1b[39m test  " +
				"\x1b[0m\n\x1b[39m\x1b[0m\x1b[39m  \x1b[33mPOST\x1b[0m\x1b[39m         test2 " +
				"\x1b[90m...................................................\x1b[0m\x1b[39m test2  " +
				"\x1b[0m\n\x1b[39m\x1b[0m\x1b[39m   \x1b[90m                                                        " +
				"\x1b[0m\x1b[39m \x1b[34;1mShowing [2] routes\x1b[0m\x1b[39m  \x1b[0m\n\x1b[39m\x1b[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeEach()
			tt.setup()
			command := NewList(mockRoute)
			assert.Equal(t, tt.expected, color.CaptureOutput(func(w io.Writer) {
				assert.NoError(t, command.Handle(mockContext))
			}))
		})
	}
}
