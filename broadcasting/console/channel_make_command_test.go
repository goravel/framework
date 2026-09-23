package console

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/file"
)

func TestChannelMakeCommand(t *testing.T) {
	defer func() {
		assert.Nil(t, file.Remove("app"))
	}()

	cmd := &ChannelMakeCommand{}

	t.Run("empty name", func(t *testing.T) {
		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Arguments().Return([]string{""}).Once()
		mockContext.EXPECT().Ask("Enter the channel name", mock.Anything).Return("", errors.New("the channel name cannot be empty")).Once()
		mockContext.EXPECT().Error("the channel name cannot be empty").Once()
		assert.Nil(t, cmd.Handle(mockContext))
	})

	t.Run("success", func(t *testing.T) {
		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Arguments().Return([]string{"TestChannel"}).Once()
		mockContext.EXPECT().OptionBool("force").Return(false).Once()
		mockContext.EXPECT().Success("Channel created successfully").Once()
		assert.Nil(t, cmd.Handle(mockContext))

		channelPath := filepath.Join("app", "broadcasting", "test_channel.go")
		assert.True(t, file.Exists(channelPath))
		assert.True(t, file.Contain(channelPath, "package broadcasting"))
		assert.True(t, file.Contain(channelPath, "func TestChannel("))
		assert.True(t, file.Contain(channelPath, "var _ broadcasting.ChannelAuthFunc = TestChannel"))
	})
}

func TestChannelMakeCommand_MultipleNames(t *testing.T) {
	defer func() {
		assert.Nil(t, file.Remove("app"))
	}()

	cmd := &ChannelMakeCommand{}
	mockContext := mocksconsole.NewContext(t)

	mockContext.EXPECT().Arguments().Return([]string{"OrderChannel", "UserChannel"}).Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Times(2)
	mockContext.EXPECT().Success("Channel created successfully").Times(2)
	assert.Nil(t, cmd.Handle(mockContext))

	assert.True(t, file.Exists(filepath.Join("app", "broadcasting", "order_channel.go")))
	assert.True(t, file.Exists(filepath.Join("app", "broadcasting", "user_channel.go")))
}
