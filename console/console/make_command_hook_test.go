package console

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mocksconsole "github.com/goravel/framework/mocks/console"
	"github.com/goravel/framework/support/file"
)

const (
	commandHookBootstrapApp = `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().Create()
}
`

	commandHookBootstrapAppWithHooks = `package bootstrap

import (
	"github.com/goravel/framework/app/console/hooks"
	"github.com/goravel/framework/contracts/console"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithCommandHooks(func() map[string][]console.CommandHook {
			return map[string][]console.CommandHook{
				"migrate": {
					&hooks.Existing{},
				},
			}
		}).Create()
}
`

	commandHookBootstrapAppWithNamedHooks = `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func commandHooks() map[string][]console.CommandHook {
	return nil
}

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithCommandHooks(commandHooks()).Create()
}
`

	commandHookBootstrapAppWithResolvableNamedHooks = `package bootstrap

import (
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().WithCommandHooks(CommandHooks).Create()
}
`

	commandHookBootstrapNamedHooks = `package bootstrap

import (
	"github.com/goravel/framework/contracts/console"
)

func CommandHooks() map[string][]console.CommandHook {
	return map[string][]console.CommandHook{}
}
`
)

func TestMakeCommandHook(t *testing.T) {
	defer func() {
		assert.Nil(t, file.Remove("app"))
		assert.Nil(t, file.Remove("bootstrap"))
	}()

	hookPath := filepath.Join("app", "console", "hooks", "audit_log.go")
	makeCommandHook := &MakeCommandHook{}

	// Cases that must generate a file start from a clean output tree.
	resetHooks := func() {
		assert.NoError(t, file.Remove(filepath.Join("app", "console", "hooks")))
	}

	t.Run("empty name", func(t *testing.T) {
		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Option("command").Return("").Once()
		mockContext.EXPECT().Arguments().Return(nil).Once()
		mockContext.EXPECT().Ask("Enter the command hook name", mock.Anything).Return("", errors.New("the command hook name cannot be empty")).Once()
		mockContext.EXPECT().Error("the command hook name cannot be empty").Once()

		assert.Nil(t, makeCommandHook.Handle(mockContext))
	})

	t.Run("hook already exists", func(t *testing.T) {
		assert.NoError(t, file.PutContent(hookPath, "package hooks\n"))

		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Option("command").Return("").Once()
		mockContext.EXPECT().Arguments().Return([]string{"AuditLog"}).Once()
		mockContext.EXPECT().OptionBool("force").Return(false).Once()
		mockContext.EXPECT().Error("the command hook already exists. Use the --force or -f flag to overwrite").Once()

		assert.Nil(t, makeCommandHook.Handle(mockContext))
	})

	t.Run("force overwrites the generated file", func(t *testing.T) {
		assert.NoError(t, file.PutContent(hookPath, "package hooks\n"))

		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Option("command").Return("*").Once()
		mockContext.EXPECT().Arguments().Return([]string{"AuditLog"}).Once()
		mockContext.EXPECT().OptionBool("force").Return(true).Once()
		mockContext.EXPECT().Success("Command hook created successfully").Once()
		mockContext.EXPECT().Warning(mock.Anything).Once()

		assert.Nil(t, makeCommandHook.Handle(mockContext))
		assert.True(t, file.Contain(hookPath, "type AuditLog struct"))
	})

	t.Run("no bootstrap setup keeps the file and warns", func(t *testing.T) {
		resetHooks()

		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Option("command").Return("").Once()
		mockContext.EXPECT().Arguments().Return([]string{"AuditLog"}).Once()
		mockContext.EXPECT().Ask("Enter the command signature the hook runs for", mock.Anything).Return("*", nil).Once()
		mockContext.EXPECT().OptionBool("force").Return(false).Once()
		mockContext.EXPECT().Success("Command hook created successfully").Once()
		mockContext.EXPECT().Warning("Command hook created but not registered, add it to WithCommandHooks in bootstrap/app.go manually").Once()

		assert.Nil(t, makeCommandHook.Handle(mockContext))
		assert.True(t, file.Exists(hookPath))
		assert.True(t, file.Contain(hookPath, "package hooks"))
		assert.True(t, file.Contain(hookPath, "type AuditLog struct"))
		assert.True(t, file.Contain(hookPath, "func (r *AuditLog) Before(ctx console.Context) (bool, error)"))
		assert.True(t, file.Contain(hookPath, "func (r *AuditLog) After(ctx console.Context, handleErr error) error"))
		assert.True(t, file.Contain(hookPath, "return handleErr"))
		assert.True(t, file.Contain(hookPath, "func (r *AuditLog) Flags() []command.Flag"))
		assert.True(t, file.Contain(hookPath, "return nil"))
	})

	t.Run("single --command key creates WithCommandHooks", func(t *testing.T) {
		resetHooks()
		bootstrapPath := filepath.Join("bootstrap", "app.go")
		assert.NoError(t, file.PutContent(bootstrapPath, commandHookBootstrapApp))

		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Option("command").Return("migrate").Once()
		mockContext.EXPECT().Arguments().Return([]string{"AuditLog"}).Once()
		mockContext.EXPECT().OptionBool("force").Return(false).Once()
		mockContext.EXPECT().Success("Command hook created successfully").Once()
		mockContext.EXPECT().Success("Command hook registered successfully").Once()

		assert.Nil(t, makeCommandHook.Handle(mockContext))

		appContent, err := file.GetContent(bootstrapPath)
		assert.NoError(t, err)
		assert.Contains(t, appContent, "WithCommandHooks(func() map[string][]console.CommandHook {")
		assert.Contains(t, appContent, `"migrate": {`)
		assert.Contains(t, appContent, "&hooks.AuditLog{}")
		assert.Contains(t, appContent, "github.com/goravel/framework/contracts/console")
		assert.Contains(t, appContent, "github.com/goravel/framework/app/console/hooks")
	})

	t.Run("single --command key appends to an existing WithCommandHooks", func(t *testing.T) {
		resetHooks()
		bootstrapPath := filepath.Join("bootstrap", "app.go")
		assert.NoError(t, file.PutContent(bootstrapPath, commandHookBootstrapAppWithHooks))

		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Option("command").Return("migrate").Once()
		mockContext.EXPECT().Arguments().Return([]string{"AuditLog"}).Once()
		mockContext.EXPECT().OptionBool("force").Return(false).Once()
		mockContext.EXPECT().Success("Command hook created successfully").Once()
		mockContext.EXPECT().Success("Command hook registered successfully").Once()

		assert.Nil(t, makeCommandHook.Handle(mockContext))

		appContent, err := file.GetContent(bootstrapPath)
		assert.NoError(t, err)
		assert.Contains(t, appContent, "&hooks.Existing{}")
		assert.Contains(t, appContent, "&hooks.AuditLog{}")
		assert.Equal(t, 1, strings.Count(appContent, "WithCommandHooks("))
	})

	t.Run("named CommandHooks function in a sibling file is registered", func(t *testing.T) {
		resetHooks()
		bootstrapPath := filepath.Join("bootstrap", "app.go")
		namedHooksPath := filepath.Join("bootstrap", "command_hooks.go")
		assert.NoError(t, file.PutContent(bootstrapPath, commandHookBootstrapAppWithResolvableNamedHooks))
		assert.NoError(t, file.PutContent(namedHooksPath, commandHookBootstrapNamedHooks))
		defer func() {
			assert.NoError(t, file.Remove(namedHooksPath))
		}()

		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Option("command").Return("migrate").Once()
		mockContext.EXPECT().Arguments().Return([]string{"AuditLog"}).Once()
		mockContext.EXPECT().OptionBool("force").Return(false).Once()
		mockContext.EXPECT().Success("Command hook created successfully").Once()
		mockContext.EXPECT().Success("Command hook registered successfully").Once()

		assert.Nil(t, makeCommandHook.Handle(mockContext))

		appContent, err := file.GetContent(bootstrapPath)
		assert.NoError(t, err)
		assert.Equal(t, commandHookBootstrapAppWithResolvableNamedHooks, appContent)

		namedHooksContent, err := file.GetContent(namedHooksPath)
		assert.NoError(t, err)
		assert.Contains(t, namedHooksContent, `"migrate": {`)
		assert.Contains(t, namedHooksContent, "&hooks.AuditLog{}")
		assert.Contains(t, namedHooksContent, "github.com/goravel/framework/app/console/hooks")
	})

	t.Run("hooks call that cannot be edited reports an error", func(t *testing.T) {
		resetHooks()
		bootstrapPath := filepath.Join("bootstrap", "app.go")
		assert.NoError(t, file.PutContent(bootstrapPath, commandHookBootstrapAppWithNamedHooks))

		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Option("command").Return("migrate").Once()
		mockContext.EXPECT().Arguments().Return([]string{"AuditLog"}).Once()
		mockContext.EXPECT().OptionBool("force").Return(false).Once()
		mockContext.EXPECT().Success("Command hook created successfully").Once()
		mockContext.EXPECT().Error(mock.MatchedBy(func(message string) bool {
			return strings.Contains(message, "failed to register command hook") && strings.Contains(message, "could not be edited")
		})).Once()

		assert.Nil(t, makeCommandHook.Handle(mockContext))
	})

	t.Run("several names share the one --command key", func(t *testing.T) {
		resetHooks()
		bootstrapPath := filepath.Join("bootstrap", "app.go")
		assert.NoError(t, file.PutContent(bootstrapPath, commandHookBootstrapApp))

		mockContext := mocksconsole.NewContext(t)
		mockContext.EXPECT().Option("command").Return("migrate").Once()
		mockContext.EXPECT().Arguments().Return([]string{"AuditLog", "SlowQuery"}).Once()
		mockContext.EXPECT().OptionBool("force").Return(false).Times(2)
		mockContext.EXPECT().Success("Command hook created successfully").Times(2)
		mockContext.EXPECT().Success("Command hook registered successfully").Times(2)

		assert.Nil(t, makeCommandHook.Handle(mockContext))

		assert.True(t, file.Exists(hookPath))
		assert.True(t, file.Exists(filepath.Join("app", "console", "hooks", "slow_query.go")))

		appContent, err := file.GetContent(bootstrapPath)
		assert.NoError(t, err)
		assert.Contains(t, appContent, "&hooks.AuditLog{}")
		assert.Contains(t, appContent, "&hooks.SlowQuery{}")
	})
}
