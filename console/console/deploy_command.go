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
			&command.StringFlag{Name: "ip", Usage: "Server IP address"},
			&command.StringFlag{Name: "port", Usage: "SSH port", Value: "22"},
			&command.StringFlag{Name: "user", Usage: "SSH user", Value: "root"},
			&command.StringFlag{Name: "key", Usage: "SSH private key path", Value: "~/.ssh/id_rsa"},
			&command.StringFlag{Name: "os", Usage: "Target OS", Value: "linux"},
			&command.StringFlag{Name: "arch", Usage: "Target arch", Value: "amd64"},
			&command.StringFlag{Name: "domain", Usage: "Domain for Caddy reverse proxy"},
			&command.StringFlag{Name: "only", Usage: "Comma-separated subset to deploy: main,public,storage,resources"},
			&command.BoolFlag{
				Name:               "rollback",
				Aliases:            []string{"r"},
				Value:              false,
				Usage:              "Rollback to previous deployment",
				DisableDefaultText: true,
			},
			&command.BoolFlag{
				Name:               "static",
				Aliases:            []string{"s"},
				Value:              false,
				Usage:              "Static compilation",
				DisableDefaultText: true,
			},
			&command.BoolFlag{
				Name:               "zero-downtime",
				Aliases:            []string{"z"},
				Value:              false,
				Usage:              "Zero downtime deployment",
				DisableDefaultText: true,
			},
		},
	}
}

func getAllOptions(config config.Config) (appName, ipAddress, sshPort, sshUser, sshKeyPath, targetOS, arch, domain string, zeroDowntime bool, staticEnv bool, reverseProxyEnabled bool, reverseProxyTLSEnabled bool) {
	appName = config.GetString("appName")
	ipAddress = config.GetString("DEPLOY_IP_ADDRESS")
	sshPort = config.GetString("DEPLOY_SSH_PORT")
	sshUser = config.GetString("DEPLOY_SSH_USER")
	sshKeyPath = config.GetString("DEPLOY_SSH_KEY_PATH")
	targetOS = config.GetString("DEPLOY_OS")
	arch = config.GetString("DEPLOY_ARCH")
	domain = config.GetString("DEPLOY_DOMAIN")

	zeroDowntime = config.GetBool("DEPLOY_ZERO_DOWNTIME")
	staticEnv = config.GetBool("DEPLOY_STATIC")
	reverseProxyEnabled = config.GetBool("DEPLOY_REVERSE_PROXY_ENABLED")
	reverseProxyTLSEnabled = config.GetBool("DEPLOY_REVERSE_PROXY_TLS_ENABLED")

	// expand ssh key ~ path if needed
	if after, ok := strings.CutPrefix(sshKeyPath, "~"); ok {
		if home, herr := os.UserHomeDir(); herr == nil {
			sshKeyPath = filepath.Join(home, after)
		}
	}

	return appName, ipAddress, sshPort, sshUser, sshKeyPath, targetOS, arch, domain, zeroDowntime, staticEnv, reverseProxyEnabled, reverseProxyTLSEnabled
}

// Handle Execute the console command.
func (r *DeployCommand) Handle(ctx console.Context) error {
	var err error

	// get all options
	appName, ipAddress, sshPort, sshUser, sshKeyPath, targetOS, arch, domain, zeroDowntime, staticEnv, reverseProxyEnabled, reverseProxyTLSEnabled := getAllOptions(r.config)

	// Rollback flow
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

	// Step 3: verify artifacts to determine which to upload
	hasMain := fileExists(appName)
	hasPublic := dirExists("public")
	hasStorage := dirExists("storage")
	hasResources := dirExists("resources")

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

	// Step 2: set up server on first run (idempotent)
	if err = supportconsole.ExecuteCommand(ctx, setupServerCommand(
		fmt.Sprintf("%v", appName),
		fmt.Sprintf("%v", ipAddress),
		fmt.Sprintf("%v", sshPort),
		fmt.Sprintf("%v", sshUser),
		fmt.Sprintf("%v", sshKeyPath),
		strings.TrimSpace(domain),
		zeroDowntime,
		reverseProxyEnabled,
		reverseProxyTLSEnabled,
	), "Setting up server (first time only)..."); err != nil {
		ctx.Error(err.Error())
		return nil
	}

	// Step 3: upload files
	if err = supportconsole.ExecuteCommand(ctx, uploadFilesCommand(
		fmt.Sprintf("%v", appName),
		fmt.Sprintf("%v", ipAddress),
		fmt.Sprintf("%v", sshPort),
		fmt.Sprintf("%v", sshUser),
		fmt.Sprintf("%v", sshKeyPath),
		hasMain, hasPublic, hasStorage, hasResources,
	), "Uploading files..."); err != nil {
		ctx.Error(err.Error())
		return nil
	}

	// Step 4: restart service
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

// setupServerCommand ensures Caddy and a systemd service are installed; no-op on subsequent runs
func setupServerCommand(appName, ip, port, user, keyPath, domain string, zeroDowntime, reverseProxyEnabled, reverseProxyTLSEnabled bool) *exec.Cmd {
	// Directories and service
	appDir := fmt.Sprintf("/var/www/%s", appName)
	binCurrent := fmt.Sprintf("%s/main", appDir)

	// Ports
	appPort := "9000"
	httpPort := "80"
	// httpsPort := "443" // only used when TLS is enabled via Caddy

	// Build systemd unit based on whether reverse proxy is used
	listenHost := "127.0.0.1"
	if !reverseProxyEnabled {
		// App listens on port 80 directly
		appPort = httpPort
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
`, appName, user, appDir, binCurrent, listenHost, appPort, appName)

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
`, keyPath, port, user, ip,
		appDir, appDir, user, user, appDir,
		// caddy install and config
		func() string {
			if !reverseProxyEnabled {
				return ""
			}
			install := "sudo apt-get update -y && sudo apt-get install -y caddy"
			writeCfg := fmt.Sprintf("echo %q | base64 -d | sudo tee /etc/caddy/Caddyfile >/dev/null && sudo systemctl enable --now caddy", caddyB64)
			return install + " && " + writeCfg
		}(),
		appName, unitB64, appName, appName,
		strings.Join(ufwCmds, " && "),
		"true",
	)

	return exec.Command("bash", "-lc", script)
}

// uploadFilesCommand uploads available artifacts to remote server
func uploadFilesCommand(appName, ip, port, user, keyPath string, hasMain, hasPublic, hasStorage, hasResources bool) *exec.Cmd {
	appDir := fmt.Sprintf("/var/www/%s", appName)
	remoteBase := fmt.Sprintf("%s@%s:%s", user, ip, appDir)
	// ensure remote base exists and permissions
	cmds := []string{
		fmt.Sprintf("ssh -o StrictHostKeyChecking=no -i %q -p %s %s@%s 'sudo mkdir -p %s && sudo chown -R %s:%s %s'", keyPath, port, user, ip, appDir, user, user, appDir),
	}

	// main binary with previous backup
	if hasMain {
		localMain := appName
		if !fileExists(localMain) && fileExists("main") {
			localMain = "main"
		}
		// upload to temp and atomically move, keeping previous as main.prev
		cmds = append(cmds,
			fmt.Sprintf("scp -o StrictHostKeyChecking=no -i %q -P %s %q %s/main.new", keyPath, port, filepath.Clean(localMain), remoteBase),
			fmt.Sprintf("ssh -o StrictHostKeyChecking=no -i %q -p %s %s@%s 'if [ -f %s/main ]; then sudo mv %s/main %s/main.prev; fi; sudo mv %s/main.new %s/main && sudo chmod +x %s/main'", keyPath, port, user, ip, appDir, appDir, appDir, appDir, appDir, appDir),
		)
	}

	if hasPublic {
		cmds = append(cmds, fmt.Sprintf("scp -o StrictHostKeyChecking=no -i %q -P %s -r %q %s", keyPath, port, filepath.Clean("public"), remoteBase))
	}
	if hasStorage {
		cmds = append(cmds, fmt.Sprintf("scp -o StrictHostKeyChecking=no -i %q -P %s -r %q %s", keyPath, port, filepath.Clean("storage"), remoteBase))
	}
	if hasResources {
		cmds = append(cmds, fmt.Sprintf("scp -o StrictHostKeyChecking=no -i %q -P %s -r %q %s", keyPath, port, filepath.Clean("resources"), remoteBase))
	}

	script := strings.Join(cmds, " && ")
	return exec.Command("bash", "-lc", script)
}

func restartServiceCommand(appName, ip, port, user, keyPath string) *exec.Cmd {
	script := fmt.Sprintf("ssh -o StrictHostKeyChecking=no -i %q -p %s %s@%s 'sudo systemctl daemon-reload && sudo systemctl restart %s || sudo systemctl start %s'", keyPath, port, user, ip, appName, appName)
	return exec.Command("bash", "-lc", script)
}

// rollbackCommand swaps main and main.prev if available, then restarts the service
func rollbackCommand(appName, ip, port, user, keyPath string) *exec.Cmd {
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
'`, keyPath, port, user, ip,
		appDir, appDir, appDir, appDir, appDir, appDir, appDir, appDir, appName, appName)
	return exec.Command("bash", "-lc", script)
}
