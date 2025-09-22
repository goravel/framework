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
   - App listens on 127.0.0.1:<DEPLOY_APP_PORT> (e.g. 9000)
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
  - DEPLOY_IP_ADDRESS                      : Target server IP
  - DEPLOY_APP_PORT                        : Backend app port when reverse proxy is used (e.g. 9000)
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
  - -F, --force-setup                     : Force re-run of provisioning even if already set up

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
DEPLOY_IP_ADDRESS=127.0.0.1
DEPLOY_APP_PORT=9000
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
DEPLOY_IP_ADDRESS=127.0.0.1
DEPLOY_APP_PORT=80
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
				Aliases:            []string{"F"},
				Value:              false,
				Usage:              "Force re-run server setup even if already configured",
				DisableDefaultText: true,
			},
		},
	}
}

func getAllOptions(ctx console.Context, cfg config.Config) (appName, ipAddress, appPort, sshPort, sshUser, sshKeyPath, targetOS, arch, domain, prodEnvFilePath string, staticEnv bool, reverseProxyEnabled bool, reverseProxyTLSEnabled bool) {
	appName = cfg.GetString("app.name")
	ipAddress = getStringEnv(cfg, "DEPLOY_IP_ADDRESS")
	appPort = getStringEnv(cfg, "DEPLOY_APP_PORT")
	sshPort = getStringEnv(cfg, "DEPLOY_SSH_PORT")
	sshUser = getStringEnv(cfg, "DEPLOY_SSH_USER")
	sshKeyPath = getStringEnv(cfg, "DEPLOY_SSH_KEY_PATH")
	targetOS = getStringEnv(cfg, "DEPLOY_OS")
	arch = getStringEnv(cfg, "DEPLOY_ARCH")
	domain = getStringEnv(cfg, "DEPLOY_DOMAIN")
	prodEnvFilePath = getStringEnv(cfg, "DEPLOY_PROD_ENV_FILE_PATH")

	staticEnv = getBoolEnv(cfg, "DEPLOY_STATIC")
	reverseProxyEnabled = getBoolEnv(cfg, "DEPLOY_REVERSE_PROXY_ENABLED")
	reverseProxyTLSEnabled = getBoolEnv(cfg, "DEPLOY_REVERSE_PROXY_TLS_ENABLED")

	// if any of the options is not set, tell the user to set it and exit
	if appName == "" {
		ctx.Error("APP_NAME environment variable is required. Please set it in the .env file. Deployment cancelled. Exiting...")
		os.Exit(1)
	}
	if ipAddress == "" {
		ctx.Error("DEPLOY_IP_ADDRESS environment variable is required. Please set it in the .env file. Deployment cancelled. Exiting...")
		os.Exit(1)
	}
	if appPort == "" {
		ctx.Error("DEPLOY_APP_PORT environment variable is required. Please set it in the .env file. Deployment cancelled. Exiting...")
		os.Exit(1)
	}
	if sshPort == "" {
		ctx.Error("DEPLOY_SSH_PORT environment variable is required. Please set it in the .env file. Deployment cancelled. Exiting...")
		os.Exit(1)
	}
	if sshUser == "" {
		ctx.Error("DEPLOY_SSH_USER environment variable is required. Please set it in the .env file. Deployment cancelled. Exiting...")
		os.Exit(1)
	}
	if sshKeyPath == "" {
		ctx.Error("DEPLOY_SSH_KEY_PATH environment variable is required. Please set it in the .env file. Deployment cancelled. Exiting...")
		os.Exit(1)
	}
	if targetOS == "" {
		ctx.Error("DEPLOY_OS environment variable is required. Please set it in the .env file. Deployment cancelled. Exiting...")
		os.Exit(1)
	}
	if arch == "" {
		ctx.Error("DEPLOY_ARCH environment variable is required. Please set it in the .env file. Deployment cancelled. Exiting...")
		os.Exit(1)
	}

	// domain is only required if reverse proxy TLS is enabled
	if reverseProxyEnabled && reverseProxyTLSEnabled && domain == "" {
		ctx.Error("DEPLOY_DOMAIN environment variable is required when reverse proxy TLS is enabled. Please set it in the .env file. Deployment cancelled. Exiting...")
		os.Exit(1)
	}

	if prodEnvFilePath == "" {
		ctx.Error("DEPLOY_PROD_ENV_FILE_PATH environment variable is required. Please set it in the .env file. Deployment cancelled. Exiting...")
		os.Exit(1)
	}

	// expand ssh key ~ path if needed
	if after, ok := strings.CutPrefix(sshKeyPath, "~"); ok {
		if home, herr := os.UserHomeDir(); herr == nil {
			sshKeyPath = filepath.Join(home, after)
		}
	}

	return appName, ipAddress, appPort, sshPort, sshUser, sshKeyPath, targetOS, arch, domain, prodEnvFilePath, staticEnv, reverseProxyEnabled, reverseProxyTLSEnabled
}

func getWhichFilesToUpload(ctx console.Context, appName, prodEnvFilePath string) (hasMain, hasProdEnv, hasPublic, hasStorage, hasResources bool) {
	hasMain = fileExists(appName)
	hasProdEnv = fileExists(prodEnvFilePath)
	hasPublic = dirExists("public")
	hasStorage = dirExists("storage")
	hasResources = dirExists("resources")

	// Allow subset selection via --only
	only := strings.TrimSpace(ctx.Option("only"))
	if only != "" {
		parts := strings.Split(only, ",")
		include := map[string]bool{}
		for _, p := range parts {
			include[strings.TrimSpace(strings.ToLower(p))] = true
		}
		if !include["main"] {
			hasMain = false
		}
		if !include["env"] {
			hasProdEnv = false
		}
		if !include["public"] {
			hasPublic = false
		}
		if !include["storage"] {
			hasStorage = false
		}
		if !include["resources"] {
			hasResources = false
		}
	}
	return hasMain, hasProdEnv, hasPublic, hasStorage, hasResources
}

// Handle Execute the console command.
func (r *DeployCommand) Handle(ctx console.Context) error {
	var err error

	// get all options
	appName, ipAddress, appPort, sshPort, sshUser, sshKeyPath, targetOS, arch, domain, prodEnvFilePath, staticEnv, reverseProxyEnabled, reverseProxyTLSEnabled := getAllOptions(ctx, r.config)

	// Rollback if needed, then exit
	if ctx.OptionBool("rollback") {
		if err = supportconsole.ExecuteCommand(ctx, rollbackCommand(
			appName, ipAddress, sshPort, sshUser, sshKeyPath,
		), "Rolling back..."); err != nil {
			ctx.Error(err.Error())
			return nil
		}
		ctx.Info("Rollback successful.")
		return nil
	}

	// Step 1: build the application
	// Build the binary for target OS/arch
	if err = supportconsole.ExecuteCommand(ctx, generateCommand(appName, targetOS, arch, staticEnv), "Building..."); err != nil {
		ctx.Error(err.Error())
		return nil
	}

	// Step 2: verify which files to upload (main, env, public, storage, resources)
	hasMain, hasProdEnv, hasPublic, hasStorage, hasResources := getWhichFilesToUpload(ctx, appName, prodEnvFilePath)

	// Step 3: set up server on first run —- skip if already set up
	if !isServerAlreadySetup(appName, ipAddress, sshPort, sshUser, sshKeyPath) {
		if err = supportconsole.ExecuteCommand(ctx, setupServerCommand(
			fmt.Sprintf("%v", appName),
			fmt.Sprintf("%v", ipAddress),
			fmt.Sprintf("%v", appPort),
			fmt.Sprintf("%v", sshPort),
			fmt.Sprintf("%v", sshUser),
			fmt.Sprintf("%v", sshKeyPath),
			strings.TrimSpace(domain),
			reverseProxyEnabled,
			reverseProxyTLSEnabled,
		), "Setting up server (first time only)..."); err != nil {
			ctx.Error(err.Error())
			return nil
		}
	} else {
		ctx.Info("Server already set up. Skipping setup.")
	}

	// Step 4: upload files
	if err = supportconsole.ExecuteCommand(ctx, uploadFilesCommand(
		fmt.Sprintf("%v", appName),
		fmt.Sprintf("%v", ipAddress),
		fmt.Sprintf("%v", sshPort),
		fmt.Sprintf("%v", sshUser),
		fmt.Sprintf("%v", sshKeyPath),
		fmt.Sprintf("%v", prodEnvFilePath),
		hasMain, hasProdEnv, hasPublic, hasStorage, hasResources,
	), "Uploading files..."); err != nil {
		ctx.Error(err.Error())
		return nil
	}

	// Step 5: restart service
	if err = supportconsole.ExecuteCommand(ctx, restartServiceCommand(
		fmt.Sprintf("%v", appName),
		fmt.Sprintf("%v", ipAddress),
		fmt.Sprintf("%v", sshPort),
		fmt.Sprintf("%v", sshUser),
		fmt.Sprintf("%v", sshKeyPath),
	), "Restarting service..."); err != nil {
		ctx.Error(err.Error())
		return nil
	}

	ctx.Info("Deploy successful.")

	return nil
}

// helpers
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// helpers: safe env parsing
func getStringEnv(cfg config.Config, key string) string {
	val := cfg.Env(key)
	if val == nil {
		return ""
	}
	s, ok := val.(string)
	if ok {
		return s
	}
	return fmt.Sprintf("%v", val)
}

func getBoolEnv(cfg config.Config, key string) bool {
	val := cfg.Env(key)
	if val == nil {
		return false
	}
	switch v := val.(type) {
	case bool:
		return v
	case string:
		t := strings.ToLower(strings.TrimSpace(v))
		return t == "1" || t == "true" || t == "t" || t == "yes" || t == "y"
	default:
		return false
	}
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
		if reverseProxyTLSEnabled && strings.TrimSpace(domain) != "" && domain != "<nil>" {
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

	return exec.Command("bash", "-lc", script)
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
		// Upload env to a temp path, then atomically place as .env
		cmds = append(cmds,
			fmt.Sprintf("scp -o StrictHostKeyChecking=no -i %q -P %s %q %s/.env.new", keyPath, sshPort, filepath.Clean(prodEnvFilePath), remoteBase),
			fmt.Sprintf("ssh -o StrictHostKeyChecking=no -i %q -p %s %s@%s 'sudo mv %s/.env.new %s/.env'", keyPath, sshPort, sshUser, ip, appDir, appDir),
		)
	}
	if hasPublic {
		cmds = append(cmds, fmt.Sprintf("scp -o StrictHostKeyChecking=no -i %q -P %s -r %q %s", keyPath, sshPort, filepath.Clean("public"), remoteBase))
	}
	if hasStorage {
		cmds = append(cmds, fmt.Sprintf("scp -o StrictHostKeyChecking=no -i %q -P %s -r %q %s", keyPath, sshPort, filepath.Clean("storage"), remoteBase))
	}
	if hasResources {
		cmds = append(cmds, fmt.Sprintf("scp -o StrictHostKeyChecking=no -i %q -P %s -r %q %s", keyPath, sshPort, filepath.Clean("resources"), remoteBase))
	}

	script := strings.Join(cmds, " && ")
	return exec.Command("bash", "-lc", script)
}

func restartServiceCommand(appName, ip, sshPort, sshUser, keyPath string) *exec.Cmd {
	script := fmt.Sprintf("ssh -o StrictHostKeyChecking=no -i %q -p %s %s@%s 'sudo systemctl daemon-reload && sudo systemctl restart %s || sudo systemctl start %s'", keyPath, sshPort, sshUser, ip, appName, appName)
	return exec.Command("bash", "-lc", script)
}

// rollbackCommand swaps main and main.prev if available, then restarts the service
func rollbackCommand(appName, ip, sshPort, sshUser, keyPath string) *exec.Cmd {
	appDir := fmt.Sprintf("/var/www/%s", appName)
	script := fmt.Sprintf(`ssh -o StrictHostKeyChecking=no -i %q -p %s %s@%s '
set -e
if [ ! -f %s/main.prev ]; then
  echo "No previous deployment to rollback to." >&2
  exit 1
fi
sudo mv %s/main %s/main.newcurrent || true
sudo mv %s/main.prev %s/main
sudo mv %s/main.newcurrent %s/main.prev || true
sudo chmod +x %s/main
sudo systemctl daemon-reload
sudo systemctl restart %s || sudo systemctl start %s
'`, keyPath, sshPort, sshUser, ip,
		appDir, appDir, appDir, appDir, appDir, appDir, appDir, appDir, appName, appName)
	return exec.Command("bash", "-lc", script)
}

// isServerAlreadySetup checks if the systemd unit already exists on remote host
func isServerAlreadySetup(appName, ip, sshPort, sshUser, keyPath string) bool {
	checkCmd := fmt.Sprintf("ssh -o StrictHostKeyChecking=no -i %q -p %s %s@%s 'test -f /etc/systemd/system/%s.service'", keyPath, sshPort, sshUser, ip, appName)
	cmd := exec.Command("bash", "-lc", checkCmd)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}
