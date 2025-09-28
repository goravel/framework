package console

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	supportconsole "github.com/goravel/framework/support/console"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/file"
)

/*
DeployCommand
===============

Overview
--------
This command implements a simple, opinionated deployment pipeline for Goravel applications.
It builds the application locally, performs a one-time remote server setup, uploads the
required artifacts to the server, restarts a systemd service, and supports rollback to the
previous binary. The goal is to provide a pragmatic, single-command deploy for small-to-medium
workloads.


Architecture assumptions
------------------------
Two primary deployment topologies are supported:
1) Reverse proxy in front of the app (recommended)
   - reverseProxyEnabled=true
   - App listens on 127.0.0.1:<DEPLOY_REVERSE_PROXY_PORT> (e.g. 9000)
   - Caddy proxies public HTTP(S) traffic to the app
   - If reverseProxyTLSEnabled=true and a valid domain is configured, Caddy terminates TLS
     and automatically provisions certificates; otherwise Caddy serves plain HTTP on :80

2) No reverse proxy
   - reverseProxyEnabled=false
   - App listens directly on :80 (APP_HOST=0.0.0.0, APP_PORT=80)

Artifacts & layout on server
----------------------------
Remote base directory: /var/www/<APP_NAME>
Files managed by this command on the remote host:
  - main        : current binary (running)
  - main.prev   : previous binary (standby for rollback)
  - .env        : environment file (uploaded from DEPLOY_PROD_ENV_FILE_PATH)
  - public/     : optional static assets
  - storage/    : optional storage directory
  - resources/  : optional resources directory

Idempotency & first-time setup
------------------------------
The initial server setup is performed exactly once per server (per app name). The command first
checks if /etc/systemd/system/<APP_NAME>.service exists over SSH. If it exists, setup is skipped.
Otherwise, the command:
  - Installs and configures Caddy (only when reverseProxyEnabled=true)
  - Creates the app directory and sets ownership
  - Writes the systemd unit for <APP_NAME>
  - Enables the service and configures the firewall (ufw)

Subsequent deploys skip the setup entirely for speed and safety (unless --force-setup is used).
Note: If you change proxy/TLS/domain settings later, pass --force-setup to re-apply provisioning
changes (e.g., regenerate Caddyfile, adjust firewall rules, rewrite the unit file).

Rollback model
--------------
Every deployment that uploads a new binary preserves the previous one as main.prev. A rollback
simply swaps main and main.prev atomically and restarts the service. Non-binary assets (.env,
public, storage, resources) are not rolled back by this command.

Build & artifacts (local)
-------------------------
The command builds the binary (name: APP_NAME) using the configured target OS/ARCH and static
linking preference. See Goravel docs for compiling guidance, artifacts, and what to upload:
https://www.goravel.dev/getting-started/compile.html

Configuration (env)
-------------------
Required:
  - app.name                               : Application name (used in remote paths/service name)
  - DEPLOY_SSH_IP                      : Target server IP
  - DEPLOY_REVERSE_PROXY_PORT                        : Backend app port when reverse proxy is used (e.g. 9000)
  - DEPLOY_SSH_PORT                        : SSH port (e.g. 22)
  - DEPLOY_SSH_USER                        : SSH username (user must have sudo privileges)
  - DEPLOY_SSH_KEY_PATH                    : Path to SSH private key (e.g. ~/.ssh/id_rsa)
  - DEPLOY_OS                              : Target OS for build (e.g. linux)
  - DEPLOY_ARCH                            : Target arch for build (e.g. amd64)
  - DEPLOY_PROD_ENV_FILE_PATH              : Local path to production .env file to upload

Optional / boolean flags (default false if unset):
  - DEPLOY_STATIC                          : Build statically when true
  - DEPLOY_REVERSE_PROXY_ENABLED           : Use Caddy reverse proxy when true
  - DEPLOY_REVERSE_PROXY_TLS_ENABLED       : Enable TLS (requires domain) when true
  - DEPLOY_DOMAIN                          : Domain name for TLS or HTTP vhost when using Caddy
                                            (required only if TLS is enabled)

CLI flags
---------
  - --only                                : Comma-separated subset to deploy: main,env,public,storage,resources
  - -r, --rollback                        : Rollback to previous binary
  - -f, --force-setup                     : Force re-run of provisioning even if already set up

Security & firewall
-------------------
The command uses SSH with StrictHostKeyChecking=no for convenience. For production, consider
manually trusting the host key to avoid MITM risks. Firewall rules are applied via ufw with
safe ordering: allow OpenSSH and required HTTP(S) ports first, then enable ufw to avoid losing
SSH connectivity.

Systemd service
---------------
The unit runs under DEPLOY_SSH_USER. Environment variables are provided via the unit for host/port,
and the working directory points to /var/www/<APP_NAME>. Service restarts are used (brief downtime).
For zero-downtime swaps, a more advanced process manager or socket activation would be required.

High-level deployment flow
--------------------------
1) Build: compile the binary for the specified target (OS/ARCH, static optional) with name APP_NAME
2) Determine artifacts to upload: main, .env, public, storage, resources (filter via --only)
3) Setup (first deploy only, or when --force-setup):
   - Create directories and permissions
   - Install/configure Caddy based on reverse proxy + TLS settings
   - Write systemd unit and enable service
   - Configure ufw rules (OpenSSH, 80, and 443 as needed)
4) Upload:
   - Binary: upload to main.new, move previous main to main.prev (if exists), atomically move main.new to main
   - .env:   upload to .env.new, atomically move to .env
   - public, storage, resources: recursively upload if they exist locally
5) Restart service: systemctl daemon-reload, then restart (or start) the service

Known limitations
-----------------
  - No migrations or database orchestration
  - Rollback covers only the binary; assets/env are not rolled back
  - StrictHostKeyChecking is disabled by default for convenience
  - Changing proxy/TLS/domain requires --force-setup to re-apply provisioning
  - Assumes Debian/Ubuntu with apt-get and ufw available

Usage examples
--------------

Usage example (1 - with reverse proxy):

Assuming you have the following .env file stored in the root of your project as .env.production:
```
APP_NAME=my-app
DEPLOY_SSH_IP=127.0.0.1
DEPLOY_REVERSE_PROXY_PORT=9000
DEPLOY_SSH_PORT=22
DEPLOY_SSH_USER=deploy
DEPLOY_SSH_KEY_PATH=~/.ssh/id_rsa
DEPLOY_OS=linux
DEPLOY_ARCH=amd64
DEPLOY_PROD_ENV_FILE_PATH=.env.production
DEPLOY_STATIC=true
DEPLOY_REVERSE_PROXY_ENABLED=true
DEPLOY_REVERSE_PROXY_TLS_ENABLED=true
DEPLOY_DOMAIN=my-app.com
```
You can then deploy your application to the server with the following command:
```
go run . artisan deploy
```
This will:
1. Build the application
2. On the remote server: install Caddy as a reverse proxy, support TLS, configure Caddy to proxy traffic to the application on port 9000, and only allow traffic from the domain my-app.com.
3. On the remote server: install ufw, and set up the firewall to allow traffic to the application.
4. On the remote server: create the systemd unit file and enable it
5. Upload the application binary, environment file, public directory, storage directory, and resources directory to the server
6. Restart the systemd service that manages the application


Usage example (2 - without reverse proxy):

You can also deploy without a reverse proxy by setting the DEPLOY_REVERSE_PROXY_ENABLED environment variable to false. For example,
assuming you have the following .env file stored in the root of your project as .env.production and you want to deploy your application to the server without a reverse proxy:
```
APP_NAME=my-app
DEPLOY_SSH_IP=127.0.0.1
DEPLOY_REVERSE_PROXY_PORT=80
DEPLOY_SSH_PORT=22
DEPLOY_SSH_USER=deploy
DEPLOY_SSH_KEY_PATH=~/.ssh/id_rsa
DEPLOY_OS=linux
DEPLOY_ARCH=amd64
DEPLOY_PROD_ENV_FILE_PATH=.env.production
DEPLOY_STATIC=true
DEPLOY_REVERSE_PROXY_ENABLED=false
DEPLOY_REVERSE_PROXY_TLS_ENABLED=false
DEPLOY_DOMAIN=
```

You can then deploy your application to the server with the following command:
```
go run . artisan deploy
```

This will:
1. Build the application
2. On the remote server: install ufw, and set up the firewall to allow traffic to the application that is listening on port 80 (http).
3. On the remote server: create the systemd unit file and enable it
4. Upload the application binary, environment file, public directory, storage directory, and resources directory to the server
5. Restart the systemd service that manages the application
```

Usage example (3 - rollback):

You can also rollback a deployment to the previous binary by running the following command:
```
go run . artisan deploy --rollback
```


Usage example (4 - force setup):

You can also force the setup of the server by running the following command:
```
go run . artisan deploy --force-setup
```


Usage example (5 - only deploy subset of files):

You can also deploy only a subset of the files (such as only the main binary and the environment file) by running the following command:
```
go run . artisan deploy --only main,env
```
*/

// deployOptions is a struct that contains all the options for the deploy command
type deployOptions struct {
	appName                string
	ipAddress              string
	appPort                string
	sshPort                string
	sshUser                string
	sshKeyPath             string
	targetOS               string
	arch                   string
	domain                 string
	prodEnvFilePath        string
	staticEnv              bool
	reverseProxyEnabled    bool
	reverseProxyTLSEnabled bool
}

type uploadOptions struct {
	hasMain      bool
	hasProdEnv   bool
	hasPublic    bool
	hasStorage   bool
	hasResources bool
}

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
	return command.Extend{
		Flags: []command.Flag{
			&command.StringFlag{
				Name:  "only",
				Usage: "Comma-separated subset to deploy: main,public,storage,resources,env. For example, to only deploy the main binary and the environment file, you can use 'main,env'",
			},
			&command.BoolFlag{
				Name:               "rollback",
				Aliases:            []string{"r"},
				Value:              false,
				Usage:              "Rollback to previous deployment",
				DisableDefaultText: true,
			},
			&command.BoolFlag{
				Name:               "force-setup",
				Aliases:            []string{"f"},
				Value:              false,
				Usage:              "Force re-run server setup even if already configured",
				DisableDefaultText: true,
			},
		},
	}
}

// Handle Execute the console command.
func (r *DeployCommand) Handle(ctx console.Context) error {
	// Rollback check first: allow rollback without validating local host tools
	// (tests can short-circuit Spinner; real runs will still use ssh remotely)
	if ctx.OptionBool("rollback") {
		opts := r.getAllOptions(ctx)
		if err := supportconsole.ExecuteCommand(ctx, rollbackCommand(
			opts.appName, opts.ipAddress, opts.sshPort, opts.sshUser, opts.sshKeyPath,
		), "Rolling back..."); err != nil {
			ctx.Error(err.Error())
			return nil
		}
		ctx.Info("Rollback successful.")
		return nil
	}

	// check if the local host is valid, currently only support macos and linux. Also requires scp, ssh, and bash to be installed and in your path.
	if !validLocalHost(ctx) {
		return nil
	}

	// get all options
	opts := r.getAllOptions(ctx)

	// continue normal deploy flow
	var err error

	// Step 1: build the application
	// Build the binary for target OS/arch
	if err = supportconsole.ExecuteCommand(ctx, generateCommand(opts.appName, opts.targetOS, opts.arch, opts.staticEnv), "Building..."); err != nil {
		ctx.Error(err.Error())
		return nil
	}

	// Step 2: verify which files to upload (main, env, public, storage, resources)
	upload := getWhichFilesToUpload(ctx, opts.appName, opts.prodEnvFilePath)

	// Step 3: set up server on first run —- skip if already set up
	if !isServerAlreadySetup(opts.appName, opts.ipAddress, opts.sshPort, opts.sshUser, opts.sshKeyPath) {
		if err = supportconsole.ExecuteCommand(ctx, setupServerCommand(
			fmt.Sprintf("%v", opts.appName),
			fmt.Sprintf("%v", opts.ipAddress),
			fmt.Sprintf("%v", opts.appPort),
			fmt.Sprintf("%v", opts.sshPort),
			fmt.Sprintf("%v", opts.sshUser),
			fmt.Sprintf("%v", opts.sshKeyPath),
			strings.TrimSpace(opts.domain),
			opts.reverseProxyEnabled,
			opts.reverseProxyTLSEnabled,
		), "Setting up server (first time only)..."); err != nil {
			ctx.Error(err.Error())
			return nil
		}
	} else {
		ctx.Info("Server already set up. Skipping setup.")
	}

	// Step 4: upload files
	if err = supportconsole.ExecuteCommand(ctx, uploadFilesCommand(
		fmt.Sprintf("%v", opts.appName),
		fmt.Sprintf("%v", opts.ipAddress),
		fmt.Sprintf("%v", opts.sshPort),
		fmt.Sprintf("%v", opts.sshUser),
		fmt.Sprintf("%v", opts.sshKeyPath),
		fmt.Sprintf("%v", opts.prodEnvFilePath),
		upload.hasMain, upload.hasProdEnv, upload.hasPublic, upload.hasStorage, upload.hasResources,
	), "Uploading files..."); err != nil {
		ctx.Error(err.Error())
		return nil
	}

	// Step 5: restart service
	if err = supportconsole.ExecuteCommand(ctx, restartServiceCommand(
		fmt.Sprintf("%v", opts.appName),
		fmt.Sprintf("%v", opts.ipAddress),
		fmt.Sprintf("%v", opts.sshPort),
		fmt.Sprintf("%v", opts.sshUser),
		fmt.Sprintf("%v", opts.sshKeyPath),
	), "Restarting service..."); err != nil {
		ctx.Error(err.Error())
		return nil
	}

	ctx.Info("Deploy successful.")

	return nil
}

func (r *DeployCommand) getAllOptions(ctx console.Context) deployOptions {
	opts := deployOptions{}
	opts.appName = r.config.GetString("app.name")
	opts.ipAddress = r.config.GetString("DEPLOY_SSH_IP")
	opts.appPort = r.config.GetString("DEPLOY_REVERSE_PROXY_PORT")
	opts.sshPort = r.config.GetString("DEPLOY_SSH_PORT")
	opts.sshUser = r.config.GetString("DEPLOY_SSH_USER")
	opts.sshKeyPath = r.config.GetString("DEPLOY_SSH_KEY_PATH")
	opts.targetOS = r.config.GetString("DEPLOY_OS")
	opts.arch = r.config.GetString("DEPLOY_ARCH")
	opts.domain = r.config.GetString("DEPLOY_DOMAIN")
	opts.prodEnvFilePath = r.config.GetString("DEPLOY_PROD_ENV_FILE_PATH")

	opts.staticEnv = r.config.GetBool("DEPLOY_STATIC")
	opts.reverseProxyEnabled = r.config.GetBool("DEPLOY_REVERSE_PROXY_ENABLED")
	opts.reverseProxyTLSEnabled = r.config.GetBool("DEPLOY_REVERSE_PROXY_TLS_ENABLED")

	// Validate required options and report all missing at once
	var missing []string
	if opts.appName == "" {
		missing = append(missing, "APP_NAME")
	}
	if opts.ipAddress == "" {
		missing = append(missing, "DEPLOY_SSH_IP")
	}
	if opts.appPort == "" {
		missing = append(missing, "DEPLOY_REVERSE_PROXY_PORT")
	}
	if opts.sshPort == "" {
		missing = append(missing, "DEPLOY_SSH_PORT")
	}
	if opts.sshUser == "" {
		missing = append(missing, "DEPLOY_SSH_USER")
	}
	if opts.sshKeyPath == "" {
		missing = append(missing, "DEPLOY_SSH_KEY_PATH")
	}
	if opts.targetOS == "" {
		missing = append(missing, "DEPLOY_OS")
	}
	if opts.arch == "" {
		missing = append(missing, "DEPLOY_ARCH")
	}
	// domain is only required if reverse proxy TLS is enabled
	if opts.reverseProxyEnabled && opts.reverseProxyTLSEnabled && opts.domain == "" {
		missing = append(missing, "DEPLOY_DOMAIN")
	}
	if opts.prodEnvFilePath == "" {
		missing = append(missing, "DEPLOY_PROD_ENV_FILE_PATH")
	}
	if len(missing) > 0 {
		ctx.Error(fmt.Sprintf("Missing required environment variables: %s. Please set them in the .env file. Deployment cancelled. Exiting...", strings.Join(missing, ", ")))
		os.Exit(1)
	}

	// expand ssh key ~ path if needed
	if after, ok := strings.CutPrefix(opts.sshKeyPath, "~"); ok {
		if home, herr := os.UserHomeDir(); herr == nil {
			opts.sshKeyPath = filepath.Join(home, after)
		}
	}

	return opts
}

func getWhichFilesToUpload(ctx console.Context, appName, prodEnvFilePath string) uploadOptions {
	res := uploadOptions{}
	res.hasMain = file.Exists(appName)
	res.hasProdEnv = file.Exists(prodEnvFilePath)
	res.hasPublic = file.Exists("public")
	res.hasStorage = file.Exists("storage")
	res.hasResources = file.Exists("resources")

	// Allow subset selection via --only
	only := strings.TrimSpace(ctx.Option("only"))
	if only != "" {
		parts := strings.Split(only, ",")
		include := map[string]bool{}
		for _, p := range parts {
			include[strings.TrimSpace(strings.ToLower(p))] = true
		}
		if !include["main"] {
			res.hasMain = false
		}
		if !include["env"] {
			res.hasProdEnv = false
		}
		if !include["public"] {
			res.hasPublic = false
		}
		if !include["storage"] {
			res.hasStorage = false
		}
		if !include["resources"] {
			res.hasResources = false
		}
	}
	return res
}

// validLocalHost checks if the local host is valid, currently only support macos and linux. Also requires scp, ssh, and bash to be installed and in your path.
func validLocalHost(ctx console.Context) bool {
	var errs []string

	if !env.IsDarwin() && !env.IsLinux() && !env.IsWindows() {
		errs = append(errs, "only macos, linux, and windows are supported. Please use a supported machine to deploy.")
	}

	if _, err := exec.LookPath("scp"); err != nil {
		errs = append(errs, "scp is not installed. Please install it, add it to your path, and try again.")
	}

	if _, err := exec.LookPath("ssh"); err != nil {
		errs = append(errs, "ssh is not installed. Please install it, add it to your path, and try again.")
	}

	// Shell requirements depend on OS
	if env.IsWindows() {
		if _, err := exec.LookPath("cmd"); err != nil {
			errs = append(errs, "cmd is not available. Please ensure Windows command processor is accessible and try again.")
		}
	} else {
		if _, err := exec.LookPath("bash"); err != nil {
			errs = append(errs, "bash is not installed. Please install it, add it to your path, and try again.")
		}
	}

	if len(errs) > 0 {
		ctx.Error("Environment validation errors:\n - " + strings.Join(errs, "\n - "))
		return false
	}

	return true
}

// makeLocalCommand chooses the appropriate local shell to execute the composed script.
func makeLocalCommand(script string) *exec.Cmd {
	if env.IsWindows() {
		return exec.Command("cmd", "/C", script)
	}
	return exec.Command("bash", "-lc", script)
}

// setupServerCommand ensures Caddy and a systemd service are installed; no-op on subsequent runs
func setupServerCommand(appName, ip, appPort, sshPort, sshUser, keyPath, domain string, reverseProxyEnabled, reverseProxyTLSEnabled bool) *exec.Cmd {
	// Directories and service
	appDir := fmt.Sprintf("/var/www/%s", appName)
	binCurrent := fmt.Sprintf("%s/main", appDir)

	// Build systemd unit based on whether reverse proxy is used
	listenHost := "127.0.0.1"
	if !reverseProxyEnabled {
		// App listens on port 80 directly
		appPort = "80"
		listenHost = "0.0.0.0"
	}

	unit := fmt.Sprintf(`[Unit]
Description=Goravel App %s
After=network.target

[Service]
User=%s
WorkingDirectory=%s
ExecStart=%s
Environment=APP_ENV=production
Environment=APP_HOST=%s
Environment=APP_PORT=%s
Restart=always
RestartSec=5
KillSignal=SIGINT
SyslogIdentifier=%s

[Install]
WantedBy=multi-user.target
`, appName, sshUser, appDir, binCurrent, listenHost, appPort, appName)

	// Build Caddyfile if reverse proxy enabled
	caddyfile := ""
	if reverseProxyEnabled {
		site := ":80"
		if reverseProxyTLSEnabled && strings.TrimSpace(domain) != "" {
			site = domain
		}
		upstream := fmt.Sprintf("127.0.0.1:%s", appPort)
		caddyfile = fmt.Sprintf(`%s {
    reverse_proxy %s {
        lb_try_duration 30s
        lb_try_interval 250ms
    }
    encode gzip
}
`, site, upstream)
	}

	unitB64 := base64.StdEncoding.EncodeToString([]byte(unit))
	var caddyB64 string
	if caddyfile != "" {
		caddyB64 = base64.StdEncoding.EncodeToString([]byte(caddyfile))
	}

	// Firewall commands based on configuration
	ufwCmds := []string{"sudo apt-get update -y && sudo apt-get install -y ufw", "sudo ufw --force enable"}
	if reverseProxyEnabled {
		ufwCmds = append(ufwCmds, "sudo ufw allow 80")
		if reverseProxyTLSEnabled {
			ufwCmds = append(ufwCmds, "sudo ufw allow 443")
		}
	} else {
		// App listens on 80 directly
		ufwCmds = append(ufwCmds, "sudo ufw allow 80")
	}

	// Remote setup script: create directories, install Caddy optionally, write configs
	script := fmt.Sprintf(`ssh -o StrictHostKeyChecking=no -i %q -p %s %s@%s '
set -e
if [ ! -d %s ]; then
  sudo mkdir -p %s
  sudo chown -R %s:%s %s
fi
%s
if [ ! -f /etc/systemd/system/%s.service ]; then
  echo %q | base64 -d | sudo tee /etc/systemd/system/%s.service >/dev/null
  sudo systemctl daemon-reload
  sudo systemctl enable %s
fi
%s
%s'
`, keyPath, sshPort, sshUser, ip,
		appDir, appDir, sshUser, sshUser, appDir,
		// caddy install and config
		func() string {
			if !reverseProxyEnabled {
				return ""
			}
			install := "sudo apt-get update -y && sudo apt-get install -y caddy"
			writeCfg := fmt.Sprintf("echo %q | base64 -d | sudo tee /etc/caddy/Caddyfile >/dev/null && sudo systemctl enable --now caddy && sudo systemctl reload caddy || sudo systemctl restart caddy", caddyB64)
			return install + " && " + writeCfg
		}(),
		appName, unitB64, appName, appName,
		// Firewall: open before enabling to avoid SSH lockout
		func() string {
			cmds := append([]string{"sudo ufw allow OpenSSH"}, ufwCmds...)
			return strings.Join(cmds, " && ")
		}(),
		"true",
	)

	return makeLocalCommand(script)
}

// uploadFilesCommand uploads available artifacts to remote server
func uploadFilesCommand(appName, ip, sshPort, sshUser, keyPath, prodEnvFilePath string, hasMain, hasProdEnv, hasPublic, hasStorage, hasResources bool) *exec.Cmd {
	appDir := fmt.Sprintf("/var/www/%s", appName)
	remoteBase := fmt.Sprintf("%s@%s:%s", sshUser, ip, appDir)
	// ensure remote base exists and permissions
	cmds := []string{
		fmt.Sprintf("ssh -o StrictHostKeyChecking=no -i %q -p %s %s@%s 'sudo mkdir -p %s && sudo chown -R %s:%s %s'", keyPath, sshPort, sshUser, ip, appDir, sshUser, sshUser, appDir),
	}

	// main binary with previous backup
	if hasMain {
		// upload to temp and atomically move, keeping previous as main.prev
		cmds = append(cmds,
			fmt.Sprintf("scp -o StrictHostKeyChecking=no -i %q -P %s %q %s/main.new", keyPath, sshPort, filepath.Clean(appName), remoteBase),
			fmt.Sprintf("ssh -o StrictHostKeyChecking=no -i %q -p %s %s@%s 'if [ -f %s/main ]; then sudo mv %s/main %s/main.prev; fi; sudo mv %s/main.new %s/main && sudo chmod +x %s/main'", keyPath, sshPort, sshUser, ip, appDir, appDir, appDir, appDir, appDir, appDir),
		)
	}

	if hasProdEnv {
		// Upload env to a temp path, then atomically place as .env; backup previous as .env.prev if exists
		cmds = append(cmds,
			fmt.Sprintf("scp -o StrictHostKeyChecking=no -i %q -P %s %q %s/.env.new", keyPath, sshPort, filepath.Clean(prodEnvFilePath), remoteBase),
			fmt.Sprintf("ssh -o StrictHostKeyChecking=no -i %q -p %s %s@%s 'if [ -f %s/.env ]; then sudo mv %s/.env %s/.env.prev; fi; sudo mv %s/.env.new %s/.env'", keyPath, sshPort, sshUser, ip, appDir, appDir, appDir, appDir, appDir),
		)
	}
	if hasPublic {
		cmds = append(cmds,
			fmt.Sprintf("ssh -o StrictHostKeyChecking=no -i %q -p %s %s@%s 'if [ -d %s/public ]; then sudo rm -rf %s/public.prev; sudo mv %s/public %s/public.prev; fi'", keyPath, sshPort, sshUser, ip, appDir, appDir, appDir, appDir),
			fmt.Sprintf("scp -o StrictHostKeyChecking=no -i %q -P %s -r %q %s", keyPath, sshPort, filepath.Clean("public"), remoteBase),
		)
	}
	if hasStorage {
		cmds = append(cmds,
			fmt.Sprintf("ssh -o StrictHostKeyChecking=no -i %q -p %s %s@%s 'if [ -d %s/storage ]; then sudo rm -rf %s/storage.prev; sudo mv %s/storage %s/storage.prev; fi'", keyPath, sshPort, sshUser, ip, appDir, appDir, appDir, appDir),
			fmt.Sprintf("scp -o StrictHostKeyChecking=no -i %q -P %s -r %q %s", keyPath, sshPort, filepath.Clean("storage"), remoteBase),
		)
	}
	if hasResources {
		cmds = append(cmds,
			fmt.Sprintf("ssh -o StrictHostKeyChecking=no -i %q -p %s %s@%s 'if [ -d %s/resources ]; then sudo rm -rf %s/resources.prev; sudo mv %s/resources %s/resources.prev; fi'", keyPath, sshPort, sshUser, ip, appDir, appDir, appDir, appDir),
			fmt.Sprintf("scp -o StrictHostKeyChecking=no -i %q -P %s -r %q %s", keyPath, sshPort, filepath.Clean("resources"), remoteBase),
		)
	}

	script := strings.Join(cmds, " && ")
	return makeLocalCommand(script)
}

func restartServiceCommand(appName, ip, sshPort, sshUser, keyPath string) *exec.Cmd {
	script := fmt.Sprintf("ssh -o StrictHostKeyChecking=no -i %q -p %s %s@%s 'sudo systemctl daemon-reload && sudo systemctl restart %s || sudo systemctl start %s'", keyPath, sshPort, sshUser, ip, appName, appName)
	return makeLocalCommand(script)
}

// rollbackCommand swaps main and main.prev if available, then restarts the service
func rollbackCommand(appName, ip, sshPort, sshUser, keyPath string) *exec.Cmd {
	appDir := fmt.Sprintf("/var/www/%s", appName)
	script := fmt.Sprintf(`ssh -o StrictHostKeyChecking=no -i %q -p %s %s@%s '
set -e
APP_DIR=%q
SERVICE=%q
if [ ! -f "$APP_DIR/main.prev" ]; then
  echo "No previous deployment to rollback to." >&2
  exit 1
fi
sudo mv "$APP_DIR/main" "$APP_DIR/main.newcurrent" || true
sudo mv "$APP_DIR/main.prev" "$APP_DIR/main"
sudo mv "$APP_DIR/main.newcurrent" "$APP_DIR/main.prev" || true
sudo chmod +x "$APP_DIR/main"
if [ -f "$APP_DIR/.env.prev" ]; then
  sudo mv "$APP_DIR/.env" "$APP_DIR/.env.newcurrent" || true
  sudo mv "$APP_DIR/.env.prev" "$APP_DIR/.env"
  sudo mv "$APP_DIR/.env.newcurrent" "$APP_DIR/.env.prev" || true
fi
if [ -d "$APP_DIR/public.prev" ]; then
  sudo mv "$APP_DIR/public" "$APP_DIR/public.newcurrent" || true
  sudo mv "$APP_DIR/public.prev" "$APP_DIR/public"
  sudo mv "$APP_DIR/public.newcurrent" "$APP_DIR/public.prev" || true
fi
if [ -d "$APP_DIR/resources.prev" ]; then
  sudo mv "$APP_DIR/resources" "$APP_DIR/resources.newcurrent" || true
  sudo mv "$APP_DIR/resources.prev" "$APP_DIR/resources"
  sudo mv "$APP_DIR/resources.newcurrent" "$APP_DIR/resources.prev" || true
fi
if [ -d "$APP_DIR/storage.prev" ]; then
  sudo mv "$APP_DIR/storage" "$APP_DIR/storage.newcurrent" || true
  sudo mv "$APP_DIR/storage.prev" "$APP_DIR/storage"
  sudo mv "$APP_DIR/storage.newcurrent" "$APP_DIR/storage.prev" || true
fi
sudo systemctl daemon-reload
sudo systemctl restart "$SERVICE" || sudo systemctl start "$SERVICE"
 '`, keyPath, sshPort, sshUser, ip, appDir, appName)
	return exec.Command("bash", "-lc", script)
}

// isServerAlreadySetup checks if the systemd unit already exists on remote host
func isServerAlreadySetup(appName, ip, sshPort, sshUser, keyPath string) bool {
	checkCmd := fmt.Sprintf("ssh -o StrictHostKeyChecking=no -i %q -p %s %s@%s 'test -f /etc/systemd/system/%s.service'", keyPath, sshPort, sshUser, ip, appName)
	cmd := makeLocalCommand(checkCmd)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}
