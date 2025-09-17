package console

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	supportconsole "github.com/goravel/framework/support/console"
)

type DeployCommand struct {
	config config.Config
}

func NewDeployCommand(config config.Config) *DeployCommand {
	return &DeployCommand{
		config: config,
	}
}

// Signature The name and signature of the console command.
func (r *DeployCommand) Signature() string {
	return "deploy"
}

// Description The console command description.
func (r *DeployCommand) Description() string {
	return "Deploy the application"
}

// Extend The console command extend.
func (r *DeployCommand) Extend() command.Extend {
	return command.Extend{}
}

// Handle Execute the console command.
func (r *DeployCommand) Handle(ctx console.Context) error {
	var err error

	// get all environment variables
	app_name := r.config.Env("APP_NAME")
	ip_address := r.config.Env("DEPLOY_IP_ADDRESS")
	ssh_port := r.config.Env("DEPLOY_SSH_PORT")
	ssh_user := r.config.Env("DEPLOY_SSH_USER")
	ssh_key_path := r.config.Env("DEPLOY_SSH_KEY_PATH")
	os := r.config.Env("DEPLOY_OS")
	arch := r.config.Env("DEPLOY_ARCH")
	static := r.config.Env("DEPLOY_STATIC")

	// if any of the required environment variables are missing, prompt the user to enter them
	if app_name == "" {
		if app_name, err = ctx.Ask("Enter the app name", console.AskOption{Default: "app"}); err != nil {
			ctx.Error(fmt.Sprintf("Enter the app name error: %v", err))
			return nil
		}
	}

	if ip_address == "" {
		if ip_address, err = ctx.Ask("Enter the server IP address"); err != nil {
			ctx.Error(fmt.Sprintf("Enter the server IP address error: %v", err))
			return nil
		}
	}

	if ssh_port == "" {
		if ssh_port, err = ctx.Ask("Enter the SSH port", console.AskOption{Default: "22"}); err != nil {
			ctx.Error(fmt.Sprintf("Enter the SSH port error: %v", err))
			return nil
		}
	}

	if ssh_user == "" {
		if ssh_user, err = ctx.Ask("Enter the SSH user", console.AskOption{Default: "root"}); err != nil {
			ctx.Error(fmt.Sprintf("Enter the SSH user error: %v", err))
			return nil
		}
	}

	if ssh_key_path == "" {
		if ssh_key_path, err = ctx.Ask("Enter the SSH key path", console.AskOption{Default: "~/.ssh/id_rsa"}); err != nil {
			ctx.Error(fmt.Sprintf("Enter the SSH key path error: %v", err))
			return nil
		}
	}

	if os == "" {
		if os, err = ctx.Choice("Select target os", []console.Choice{
			{Key: "Linux", Value: "linux"},
			{Key: "Windows", Value: "windows"},
			{Key: "Darwin", Value: "darwin"},
		}, console.ChoiceOption{Default: runtime.GOOS}); err != nil {
			ctx.Error(fmt.Sprintf("Select target os error: %v", err))
			return nil
		}
	}

	if arch == "" {
		if arch, err = ctx.Choice("Select target arch", []console.Choice{
			{Key: "amd64", Value: "amd64"},
			{Key: "arm64", Value: "arm64"},
			{Key: "386", Value: "386"},
		}, console.ChoiceOption{Default: "amd64"}); err != nil {
			ctx.Error(fmt.Sprintf("Select target arch error: %v", err))
			return nil
		}
	}

	// if static is not set, prompt the user to enter it
	if static == "" {
		if !ctx.Confirm("Do you want to build a static binary?") {
			static = false
		} else {
			static = true
		}
	}

	// build the application
	if err = supportconsole.ExecuteCommand(ctx, generateCommand(fmt.Sprintf("%v", app_name), fmt.Sprintf("%v", os), fmt.Sprintf("%v", arch), static.(bool)), "Building..."); err != nil {
		ctx.Error(err.Error())
		return nil
	}

	// deploy the application
	if err = supportconsole.ExecuteCommand(ctx, deployCommand(
		fmt.Sprintf("%v", app_name),
		fmt.Sprintf("%v", ip_address),
		fmt.Sprintf("%v", ssh_port),
		fmt.Sprintf("%v", ssh_user),
		fmt.Sprintf("%v", ssh_key_path),
	), "Deploying..."); err != nil {
		ctx.Error(err.Error())
		return nil
	}

	ctx.Info("Deploy successful.")

	return nil
}

// generate the deploy command
func deployCommand(binary_name, ip_address, ssh_port, ssh_user, ssh_key_path string) *exec.Cmd {
	// TODO: implement deploy command
	return exec.Command("echo", "Deploy command not implemented yet... exiting...")
}
