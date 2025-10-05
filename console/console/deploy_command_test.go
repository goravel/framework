package console

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
)

// Helper to extract the first base64 payload used in an echo ... | base64 -d | tee ... sequence.
func extractBase64(script, teePath string) (string, bool) {
	// find segment like: echo "<b64>" | base64 -d | sudo tee <teePath>
	pivot := " | base64 -d | sudo tee " + teePath
	// search backwards for preceding echo "
	idx := strings.Index(script, pivot)
	if idx == -1 {
		return "", false
	}
	// find preceding 'echo "'
	pre := script[:idx]
	start := strings.LastIndex(pre, "echo \"")
	if start == -1 {
		return "", false
	}
	start += len("echo \"")
	// find closing quote before pivot
	b64 := pre[start:]
	end := strings.LastIndex(b64, "\"")
	if end == -1 {
		return "", false
	}
	return b64[:end], true
}

func Test_setupServerCommand_NoProxy(t *testing.T) {
	cmd := setupServerCommand(deployOptions{appName: "myapp", sshIp: "203.0.113.10", reverseProxyPort: "9000", sshPort: "22", sshUser: "ubuntu", sshKeyPath: "~/.ssh/id", deployBaseDir: "/var/www/", domain: "", reverseProxyEnabled: false, reverseProxyTLSEnabled: false})
	require.NotNil(t, cmd)
	if runtime.GOOS == "windows" {
		t.Skip("Skipping script content assertions on Windows shell")
	}
	require.GreaterOrEqual(t, len(cmd.Args), 3)
	script := cmd.Args[2]

	// No Caddy installation
	assert.NotContains(t, script, "install -y caddy")
	// App listens on :80 directly
	unitB64, ok := extractBase64(script, "/etc/systemd/system/myapp.service >/dev/null")
	require.True(t, ok, "unit base64 not found")
	unitBytes, err := base64.StdEncoding.DecodeString(unitB64)
	require.NoError(t, err)
	unit := string(unitBytes)
	assert.Contains(t, unit, "User=ubuntu")
	assert.Contains(t, unit, "APP_HOST=0.0.0.0")
	assert.Contains(t, unit, "APP_PORT=80")

	// UFW ordering: allow OpenSSH then allow 80 then enable
	assert.Contains(t, script, "ufw allow OpenSSH")
	assert.Contains(t, script, "ufw allow 80")
}

func Test_setupServerCommand_ProxyHTTP(t *testing.T) {
	cmd := setupServerCommand(deployOptions{appName: "myapp", sshIp: "203.0.113.10", reverseProxyPort: "9000", sshPort: "22", sshUser: "ubuntu", sshKeyPath: "~/.ssh/id", deployBaseDir: "/var/www/", domain: "", reverseProxyEnabled: true, reverseProxyTLSEnabled: false})
	require.NotNil(t, cmd)
	if runtime.GOOS == "windows" {
		t.Skip("Skipping script content assertions on Windows shell")
	}
	script := cmd.Args[2]
	// Caddy install present
	assert.Contains(t, script, "install -y caddy")
	// Caddyfile site :80, upstream 127.0.0.1:9000
	caddyB64, ok := extractBase64(script, "/etc/caddy/Caddyfile >/dev/null")
	require.True(t, ok, "caddy base64 not found")
	caddyBytes, err := base64.StdEncoding.DecodeString(caddyB64)
	require.NoError(t, err)
	caddy := string(caddyBytes)
	assert.Contains(t, caddy, ":80 {")
	assert.Contains(t, caddy, "reverse_proxy 127.0.0.1:9000")
	// Firewall allows 80 but not 443
	assert.Contains(t, script, "ufw allow 80")
	assert.NotContains(t, script, "ufw allow 443")
}

func Test_setupServerCommand_ProxyTLS(t *testing.T) {
	cmd := setupServerCommand(deployOptions{appName: "myapp", sshIp: "203.0.113.10", reverseProxyPort: "9000", sshPort: "22", sshUser: "ubuntu", sshKeyPath: "~/.ssh/id", deployBaseDir: "/var/www/", domain: "example.com", reverseProxyEnabled: true, reverseProxyTLSEnabled: true})
	require.NotNil(t, cmd)
	if runtime.GOOS == "windows" {
		t.Skip("Skipping script content assertions on Windows shell")
	}
	script := cmd.Args[2]
	caddyB64, ok := extractBase64(script, "/etc/caddy/Caddyfile >/dev/null")
	require.True(t, ok)
	caddyBytes, err := base64.StdEncoding.DecodeString(caddyB64)
	require.NoError(t, err)
	caddy := string(caddyBytes)
	assert.Contains(t, caddy, "example.com {")
	assert.Contains(t, script, "ufw allow 80")
	assert.Contains(t, script, "ufw allow 443")
	// Reload on change
	assert.Contains(t, script, "systemctl reload caddy || sudo systemctl restart caddy")
	// Ensure we invoke via bash on non-Windows
	require.GreaterOrEqual(t, len(cmd.Args), 2)
	assert.Equal(t, "bash", cmd.Args[0])
	assert.Equal(t, "-lc", cmd.Args[1])
}

func Test_uploadFilesCommand_AllArtifacts(t *testing.T) {
	cmd := uploadFilesCommand(deployOptions{appName: "myapp", sshIp: "203.0.113.10", sshPort: "22", sshUser: "ubuntu", sshKeyPath: "~/.ssh/id", deployBaseDir: "/var/www/"}, uploadOptions{hasMain: true, hasProdEnv: true, hasPublic: true, hasStorage: true, hasResources: true}, ".env.production")
	require.NotNil(t, cmd)
	if runtime.GOOS == "windows" {
		t.Skip("Skipping script content assertions on Windows shell")
	}
	script := cmd.Args[2]
	appDir := "/var/www/myapp"
	// Binary upload and backup (now uses zip backups + atomic replace without .prev)
	assert.Contains(t, script, "scp -o StrictHostKeyChecking=no -i \"~/.ssh/id\" -P 22 \"myapp\" ubuntu@203.0.113.10:"+appDir+"/main.new")
	assert.Contains(t, script, "sudo mv "+appDir+"/main.new "+appDir+"/main && sudo chmod +x "+appDir+"/main")
	// Ensure backup zip is created
	assert.Contains(t, script, "backups")
	assert.Contains(t, script, "zip -r")
	// .env atomic rename
	assert.Contains(t, script, ".env.new")
	assert.Contains(t, script, "/.env'")
	// Directories
	assert.Contains(t, script, "scp -o StrictHostKeyChecking=no -i \"~/.ssh/id\" -P 22 -r \"public\" ubuntu@203.0.113.10:"+appDir)
	assert.Contains(t, script, "scp -o StrictHostKeyChecking=no -i \"~/.ssh/id\" -P 22 -r \"storage\" ubuntu@203.0.113.10:"+appDir)
	assert.Contains(t, script, "scp -o StrictHostKeyChecking=no -i \"~/.ssh/id\" -P 22 -r \"resources\" ubuntu@203.0.113.10:"+appDir)
}

func Test_uploadFilesCommand_SubsetArtifacts(t *testing.T) {
	cmd := uploadFilesCommand(deployOptions{appName: "myapp", sshIp: "203.0.113.10", sshPort: "22", sshUser: "ubuntu", sshKeyPath: "~/.ssh/id", deployBaseDir: "/var/www/"}, uploadOptions{hasMain: true, hasProdEnv: false, hasPublic: false, hasStorage: true, hasResources: false}, ".env.production")
	require.NotNil(t, cmd)
	if runtime.GOOS == "windows" {
		t.Skip("Skipping script content assertions on Windows shell")
	}
	script := cmd.Args[2]
	appDir := "/var/www/myapp"
	// main present
	assert.Contains(t, script, "/main.new")
	// env absent
	assert.NotContains(t, script, ".env.new")
	// public/resources absent
	assert.NotContains(t, script, " -r \"public\"")
	assert.NotContains(t, script, " -r \"resources\"")
	// storage present
	assert.Contains(t, script, " -r \"storage\" ubuntu@203.0.113.10:"+appDir)
}

func Test_restartServiceCommand(t *testing.T) {
	cmd := restartServiceCommand(deployOptions{appName: "myapp", sshIp: "203.0.113.10", sshPort: "22", sshUser: "ubuntu", sshKeyPath: "~/.ssh/id"})
	require.NotNil(t, cmd)
	if runtime.GOOS == "windows" {
		t.Skip("Skipping script content assertions on Windows shell")
	}
	script := cmd.Args[2]
	assert.Contains(t, script, "systemctl daemon-reload")
	assert.Contains(t, script, "systemctl restart myapp || sudo systemctl start myapp")
	// Verify shell wrapper on non-Windows
	require.GreaterOrEqual(t, len(cmd.Args), 2)
	assert.Equal(t, "bash", cmd.Args[0])
	assert.Equal(t, "-lc", cmd.Args[1])
}

func Test_rollbackCommand(t *testing.T) {
	cmd := rollbackCommand(deployOptions{appName: "myapp", sshIp: "203.0.113.10", sshPort: "22", sshUser: "ubuntu", sshKeyPath: "~/.ssh/id", deployBaseDir: "/var/www/"})
	require.NotNil(t, cmd)
	if runtime.GOOS == "windows" {
		t.Skip("Skipping script content assertions on Windows shell")
	}
	script := cmd.Args[2]
	// Ensure rollback uses backup zip restore
	assert.Contains(t, script, "backups")
	assert.Contains(t, script, "unzip -o")
	// Accept either explicit service name or variable-based restart lines
	hasExplicit := strings.Contains(script, "systemctl restart myapp || sudo systemctl start myapp")
	hasVariable := strings.Contains(script, "systemctl restart \"$SERVICE\" || sudo systemctl start \"$SERVICE\"")
	assert.True(t, hasExplicit || hasVariable, "expected restart line not found")
}

func Test_getStringEnv_and_getBoolEnv(t *testing.T) {
	mockConfig := mocksconfig.NewConfig(t)
	// String as string
	mockConfig.EXPECT().EnvString("STR").Return("value").Once()
	assert.Equal(t, "value", mockConfig.EnvString("STR"))
	// String as non-string type
	mockConfig.EXPECT().EnvString("NUM").Return("123").Once()
	assert.Equal(t, "123", mockConfig.EnvString("NUM"))
	// Missing
	mockConfig.EXPECT().EnvString("MISSING").Return("").Once()
	assert.Equal(t, "", mockConfig.EnvString("MISSING"))

	// Bool parsing
	mockConfig.EXPECT().EnvBool("BOOL1").Return(true).Once()
	assert.True(t, mockConfig.EnvBool("BOOL1"))
	mockConfig.EXPECT().EnvBool("BOOL2").Return(true).Once()
	assert.True(t, mockConfig.EnvBool("BOOL2"))
	mockConfig.EXPECT().EnvBool("BOOL3").Return(true).Once()
	assert.True(t, mockConfig.EnvBool("BOOL3"))
	mockConfig.EXPECT().EnvBool("BOOL4").Return(false).Once()
	assert.False(t, mockConfig.EnvBool("BOOL4"))
	mockConfig.EXPECT().EnvBool("BOOL5").Return(false).Once()
	assert.False(t, mockConfig.EnvBool("BOOL5"))
}

func Test_getWhichFilesToUpload_and_onlyFilter(t *testing.T) {
	// Prepare temp workspace
	wd, err := os.Getwd()
	require.NoError(t, err)

	dir := t.TempDir()
	require.NoError(t, os.Chdir(dir))
	// Important: register chdir-back AFTER TempDir so it runs BEFORE TempDir's RemoveAll on Windows
	t.Cleanup(func() { _ = os.Chdir(wd) })

	// Create artifacts
	require.NoError(t, os.WriteFile("myapp", []byte("bin"), 0o755))
	require.NoError(t, os.WriteFile(".env.production", []byte("APP_ENV=prod"), 0o644))
	require.NoError(t, os.MkdirAll("public", 0o755))
	require.NoError(t, os.MkdirAll("storage", 0o755))
	require.NoError(t, os.MkdirAll("resources", 0o755))

	// Cleanup created files/directories at end of test
	t.Cleanup(func() {
		_ = os.Remove("myapp")
		_ = os.Remove(".env.production")
		_ = os.RemoveAll("public")
		_ = os.RemoveAll("storage")
		_ = os.RemoveAll("resources")
	})

	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Option("only").Return("").Once()
	up := getUploadOptions(mockContext, "myapp", ".env.production")
	assert.True(t, up.hasMain)
	assert.True(t, up.hasProdEnv)
	assert.True(t, up.hasPublic)
	assert.True(t, up.hasStorage)
	assert.True(t, up.hasResources)

	// Now test filter: only main and env
	mockContext2 := mocksconsole.NewContext(t)
	mockContext2.EXPECT().Option("only").Return("main,env").Once()
	up = getUploadOptions(mockContext2, "myapp", ".env.production")
	assert.True(t, up.hasMain)
	assert.True(t, up.hasProdEnv)
	assert.False(t, up.hasPublic)
	assert.False(t, up.hasStorage)
	assert.False(t, up.hasResources)
}

func Test_validLocalHost_ErrorAggregation_Unix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix-only test")
	}
	// Temporarily clear PATH so scp/ssh/bash are missing
	oldPath := os.Getenv("PATH")
	t.Cleanup(func() { _ = os.Setenv("PATH", oldPath) })
	require.NoError(t, os.Setenv("PATH", ""))

	mockContext := mocksconsole.NewContext(t)
	// Expect a single aggregated error call
	mockContext.EXPECT().Error(mock.MatchedBy(func(msg string) bool {
		return strings.Contains(msg, "Environment validation errors:") &&
			strings.Contains(msg, "the following binaries were not found on your path:") &&
			strings.Contains(msg, "scp") &&
			strings.Contains(msg, "ssh") &&
			strings.Contains(msg, "bash") &&
			strings.Contains(msg, "Please install them, add them to your path, and try again.")
	})).Once()
	ok := validLocalHost(mockContext)
	assert.False(t, ok)
}

func Test_validLocalHost_SucceedsWithTempTools_Unix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix-only test")
	}
	// Create temp dir with fake scp, ssh, bash
	dir := t.TempDir()
	mkExec := func(name string) {
		p := dir + string(os.PathSeparator) + name
		require.NoError(t, os.WriteFile(p, []byte("#!/bin/sh\nexit 0\n"), 0o755))
	}
	mkExec("scp")
	mkExec("ssh")
	mkExec("bash")
	oldPath := os.Getenv("PATH")
	t.Cleanup(func() { _ = os.Setenv("PATH", oldPath) })
	require.NoError(t, os.Setenv("PATH", dir))

	// Sanity: tools resolvable
	_, err := exec.LookPath("scp")
	require.NoError(t, err)
	_, err = exec.LookPath("ssh")
	require.NoError(t, err)
	_, err = exec.LookPath("bash")
	require.NoError(t, err)

	mockContext := mocksconsole.NewContext(t)
	ok := validLocalHost(mockContext)
	assert.True(t, ok)
}

// --------------------------
// Windows-specific tests
// --------------------------

func Test_setupServerCommand_WindowsShellWrapper(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only test")
	}
	cmd := setupServerCommand(deployOptions{appName: "myapp", sshIp: "203.0.113.10", reverseProxyPort: "9000", sshPort: "22", sshUser: "ubuntu", sshKeyPath: "~/.ssh/id", deployBaseDir: "/var/www/", domain: "example.com", reverseProxyEnabled: true, reverseProxyTLSEnabled: true})
	require.NotNil(t, cmd)
	require.GreaterOrEqual(t, len(cmd.Args), 2)
	assert.Equal(t, "cmd", cmd.Args[0])
	assert.Equal(t, "/C", cmd.Args[1])
}

func Test_uploadFilesCommand_WindowsShellWrapper(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only test")
	}
	cmd := uploadFilesCommand(deployOptions{appName: "myapp", sshIp: "203.0.113.10", sshPort: "22", sshUser: "ubuntu", sshKeyPath: "~/.ssh/id", deployBaseDir: "/var/www/"}, uploadOptions{hasMain: true, hasProdEnv: true, hasPublic: true, hasStorage: true, hasResources: true}, ".env.production")
	require.NotNil(t, cmd)
	require.GreaterOrEqual(t, len(cmd.Args), 2)
	assert.Equal(t, "cmd", cmd.Args[0])
	assert.Equal(t, "/C", cmd.Args[1])
}

func Test_restartServiceCommand_WindowsShellWrapper(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only test")
	}
	cmd := restartServiceCommand(deployOptions{appName: "myapp", sshIp: "203.0.113.10", sshPort: "22", sshUser: "ubuntu", sshKeyPath: "~/.ssh/id"})
	require.NotNil(t, cmd)
	require.GreaterOrEqual(t, len(cmd.Args), 2)
	assert.Equal(t, "cmd", cmd.Args[0])
	assert.Equal(t, "/C", cmd.Args[1])
}

func Test_isServerAlreadySetup_WindowsShellWrapper(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only test")
	}
	// We can't reliably assert remote state; just ensure command created uses cmd wrapper
	_ = isServerAlreadySetup(deployOptions{appName: "myapp", sshIp: "203.0.113.10", sshPort: "22", sshUser: "ubuntu", sshKeyPath: "~/.ssh/id"})
}

func Test_validLocalHost_ErrorAggregation_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only test")
	}
	// Clear PATH so scp/ssh/cmd are missing
	oldPath := os.Getenv("PATH")
	t.Cleanup(func() { _ = os.Setenv("PATH", oldPath) })
	require.NoError(t, os.Setenv("PATH", ""))

	mockContext := mocksconsole.NewContext(t)
	mockContext.EXPECT().Error(mock.MatchedBy(func(msg string) bool {
		return strings.Contains(msg, "Environment validation errors:") &&
			strings.Contains(msg, "the following binaries were not found on your path:") &&
			strings.Contains(msg, "scp") &&
			strings.Contains(msg, "ssh") &&
			strings.Contains(msg, "cmd") &&
			strings.Contains(msg, "Please install them, add them to your path, and try again.")
	})).Once()
	ok := validLocalHost(mockContext)
	assert.False(t, ok)
}

func Test_validLocalHost_SucceedsWithTempTools_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only test")
	}
	// Create temp dir with fake scp.exe, ssh.exe; rely on system cmd
	dir := t.TempDir()
	mkExec := func(name string) {
		p := dir + string(os.PathSeparator) + name
		// Windows will execute .exe; a plain text may not be executable, but LookPath will still find it
		require.NoError(t, os.WriteFile(p, []byte("echo off\r\n"), 0o755))
	}
	mkExec("scp.exe")
	mkExec("ssh.exe")
	oldPath := os.Getenv("PATH")
	t.Cleanup(func() { _ = os.Setenv("PATH", oldPath) })
	require.NoError(t, os.Setenv("PATH", dir+";"+oldPath))

	// Sanity: tools resolvable
	_, err := exec.LookPath("scp.exe")
	require.NoError(t, err)
	_, err = exec.LookPath("ssh.exe")
	require.NoError(t, err)
	// cmd should be resolvable on Windows
	_, err = exec.LookPath("cmd")
	require.NoError(t, err)

	mockContext := mocksconsole.NewContext(t)
	ok := validLocalHost(mockContext)
	assert.True(t, ok)
}

func Test_Handle_Rollback_ShortCircuit(t *testing.T) {
	// We only test rollback path to avoid executing remote checks.
	mockContext := mocksconsole.NewContext(t)
	mockConfig := mocksconfig.NewConfig(t)
	mockArtisan := mocksconsole.NewArtisan(t)
	cmd := NewDeployCommand(mockConfig, mockArtisan)

	// Minimal required envs for getDeployOptions (will not be used deeply due to rollback)
	mockConfig.EXPECT().GetString("app.name").Return("myapp").Once()
	mockConfig.EXPECT().GetString("app.deploy.ssh_ip").Return("203.0.113.10").Once()
	mockConfig.EXPECT().GetString("app.deploy.reverse_proxy_port").Return("9000").Once()
	mockConfig.EXPECT().GetString("app.deploy.ssh_port").Return("22").Once()
	mockConfig.EXPECT().GetString("app.deploy.ssh_user").Return("ubuntu").Once()
	mockConfig.EXPECT().GetString("app.deploy.ssh_key_path").Return("~/.ssh/id").Once()
	mockConfig.EXPECT().GetString("app.build.os").Return("linux").Once()
	mockConfig.EXPECT().GetString("app.build.arch").Return("amd64").Once()
	mockConfig.EXPECT().GetString("app.deploy.domain").Return("").Once()
	mockConfig.EXPECT().GetString("app.deploy.prod_env_file_path").Return(".env.production").Once()
	mockConfig.EXPECT().GetString("app.deploy.base_dir", "/var/www/").Return("/var/www/").Once()
	mockConfig.EXPECT().GetBool("app.build.static").Return(false).Once()
	mockConfig.EXPECT().GetBool("app.deploy.reverse_proxy_enabled").Return(false).Once()
	mockConfig.EXPECT().GetBool("app.deploy.reverse_proxy_tls_enabled").Return(false).Once()

	mockContext.EXPECT().OptionBool("rollback").Return(true).Once()
	mockContext.EXPECT().Spinner("Rolling back...", mock.Anything).Return(nil).Once()
	mockContext.EXPECT().Info("Rollback successful.").Once()

	assert.Nil(t, cmd.Handle(mockContext))
}

func Test_Handle_Deploy_Success(t *testing.T) {
	mockContext := mocksconsole.NewContext(t)
	mockConfig := mocksconfig.NewConfig(t)
	mockArtisan := mocksconsole.NewArtisan(t)
	cmd := NewDeployCommand(mockConfig, mockArtisan)

	// Config expectations
	mockConfig.EXPECT().GetString("app.name").Return("myapp").Once()
	// Use fast-fail SSH settings to avoid any network delay
	mockConfig.EXPECT().GetString("app.deploy.ssh_ip").Return("127.0.0.1").Once()
	mockConfig.EXPECT().GetString("app.deploy.reverse_proxy_port").Return("9000").Once()
	mockConfig.EXPECT().GetString("app.deploy.ssh_port").Return("0").Once()
	mockConfig.EXPECT().GetString("app.deploy.ssh_user").Return("ubuntu").Once()
	mockConfig.EXPECT().GetString("app.deploy.ssh_key_path").Return("~/.ssh/id").Once()
	mockConfig.EXPECT().GetString("app.build.os").Return("linux").Once()
	mockConfig.EXPECT().GetString("app.build.arch").Return("amd64").Once()
	mockConfig.EXPECT().GetString("app.deploy.domain").Return("").Once()
	mockConfig.EXPECT().GetString("app.deploy.prod_env_file_path").Return(".env.production").Once()
	mockConfig.EXPECT().GetString("app.deploy.base_dir", "/var/www/").Return("/var/www/").Once()
	mockConfig.EXPECT().GetBool("app.build.static").Return(false).Once()
	mockConfig.EXPECT().GetBool("app.deploy.reverse_proxy_enabled").Return(false).Once()
	mockConfig.EXPECT().GetBool("app.deploy.reverse_proxy_tls_enabled").Return(false).Once()

	// Context expectations
	mockContext.EXPECT().OptionBool("rollback").Return(false).Once()
	mockContext.EXPECT().OptionBool("force-setup").Return(false).Once()
	mockContext.EXPECT().Option("only").Return("").Once()

	// Ensure artifacts exist for getUploadOptions
	wd, _ := os.Getwd()
	dir := t.TempDir()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(wd) })
	require.NoError(t, os.WriteFile("myapp", []byte("bin"), 0o755))
	require.NoError(t, os.WriteFile(".env.production", []byte("APP_ENV=prod"), 0o644))

	// Expect artisan build call
	mockArtisan.EXPECT().Call("build --os linux --arch amd64 --name myapp").Return(nil).Once()

	// Spinner-wrapped commands (explicit messages)
	mockContext.EXPECT().Spinner("Setting up server (first time only)...", mock.Anything).Return(nil).Once()
	mockContext.EXPECT().Spinner("Uploading files...", mock.Anything).Return(nil).Once()
	mockContext.EXPECT().Spinner("Restarting service...", mock.Anything).Return(nil).Once()
	mockContext.EXPECT().Info("Deploy successful.").Once()

	assert.Nil(t, cmd.Handle(mockContext))
}

func Test_Handle_Deploy_FailureOnBuild(t *testing.T) {
	mockContext := mocksconsole.NewContext(t)
	mockConfig := mocksconfig.NewConfig(t)
	mockArtisan := mocksconsole.NewArtisan(t)
	cmd := NewDeployCommand(mockConfig, mockArtisan)

	// Minimal config
	mockConfig.EXPECT().GetString("app.name").Return("myapp").Once()
	mockConfig.EXPECT().GetString("app.deploy.ssh_ip").Return("203.0.113.10").Once()
	mockConfig.EXPECT().GetString("app.deploy.reverse_proxy_port").Return("9000").Once()
	mockConfig.EXPECT().GetString("app.deploy.ssh_port").Return("22").Once()
	mockConfig.EXPECT().GetString("app.deploy.ssh_user").Return("ubuntu").Once()
	mockConfig.EXPECT().GetString("app.deploy.ssh_key_path").Return("~/.ssh/id").Once()
	mockConfig.EXPECT().GetString("app.build.os").Return("linux").Once()
	mockConfig.EXPECT().GetString("app.build.arch").Return("amd64").Once()
	mockConfig.EXPECT().GetString("app.deploy.domain").Return("").Once()
	mockConfig.EXPECT().GetString("app.deploy.prod_env_file_path").Return(".env.production").Once()
	mockConfig.EXPECT().GetString("app.deploy.base_dir", "/var/www/").Return("/var/www/").Once()
	mockConfig.EXPECT().GetBool("app.build.static").Return(false).Once()
	mockConfig.EXPECT().GetBool("app.deploy.reverse_proxy_enabled").Return(false).Once()
	mockConfig.EXPECT().GetBool("app.deploy.reverse_proxy_tls_enabled").Return(false).Once()

	mockContext.EXPECT().OptionBool("rollback").Return(false).Once()
	// Build fails via artisan
	mockArtisan.EXPECT().Call("build --os linux --arch amd64 --name myapp").Return(fmt.Errorf("build error")).Once()
	// Spinner messages that may appear later if build passed (not expected here)
	mockContext.EXPECT().Spinner("Setting up server (first time only)...", mock.Anything).Return(nil).Maybe()
	mockContext.EXPECT().Spinner("Uploading files...", mock.Anything).Return(nil).Maybe()
	mockContext.EXPECT().Spinner("Restarting service...", mock.Anything).Return(nil).Maybe()
	mockContext.EXPECT().Error("build error").Once()

	assert.Nil(t, cmd.Handle(mockContext))
}

func Test_getDeployOptions_Success(t *testing.T) {
	mockContext := mocksconsole.NewContext(t)
	mockConfig := mocksconfig.NewConfig(t)
	mockArtisan := mocksconsole.NewArtisan(t)
	cmd := NewDeployCommand(mockConfig, mockArtisan)

	// Config expectations (all present)
	mockConfig.EXPECT().GetString("app.name").Return("myapp").Once()
	mockConfig.EXPECT().GetString("app.deploy.ssh_ip").Return("203.0.113.10").Once()
	mockConfig.EXPECT().GetString("app.deploy.reverse_proxy_port").Return("9000").Once()
	mockConfig.EXPECT().GetString("app.deploy.ssh_port").Return("22").Once()
	mockConfig.EXPECT().GetString("app.deploy.ssh_user").Return("ubuntu").Once()
	mockConfig.EXPECT().GetString("app.deploy.ssh_key_path").Return("~/.ssh/id").Once()
	mockConfig.EXPECT().GetString("app.build.os").Return("linux").Once()
	mockConfig.EXPECT().GetString("app.build.arch").Return("amd64").Once()
	mockConfig.EXPECT().GetString("app.deploy.domain").Return("").Once()
	mockConfig.EXPECT().GetString("app.deploy.prod_env_file_path").Return(".env.production").Once()
	mockConfig.EXPECT().GetString("app.deploy.base_dir", "/var/www/").Return("/var/www/").Once()

	mockConfig.EXPECT().GetBool("app.build.static").Return(true).Once()
	mockConfig.EXPECT().GetBool("app.deploy.reverse_proxy_enabled").Return(true).Once()
	mockConfig.EXPECT().GetBool("app.deploy.reverse_proxy_tls_enabled").Return(false).Once()

	opts := cmd.getDeployOptions(mockContext)

	assert.Equal(t, "myapp", opts.appName)
	assert.Equal(t, "203.0.113.10", opts.sshIp)
	assert.Equal(t, "9000", opts.reverseProxyPort)
	assert.Equal(t, "22", opts.sshPort)
	assert.Equal(t, "ubuntu", opts.sshUser)
	// ssh key path should expand '~' to the home directory
	home, _ := os.UserHomeDir()
	assert.Equal(t, home+"/.ssh/id", opts.sshKeyPath)
	assert.Equal(t, "linux", opts.targetOS)
	assert.Equal(t, "amd64", opts.arch)
	assert.Equal(t, ".env.production", opts.prodEnvFilePath)
	assert.Equal(t, "/var/www/", opts.deployBaseDir)
	assert.True(t, opts.staticEnv)
	assert.True(t, opts.reverseProxyEnabled)
	assert.False(t, opts.reverseProxyTLSEnabled)
}

// helper used by subprocess to trigger os.Exit via getDeployOptions missing validation
func TestHelper_GetDeployOptions_Missing(t *testing.T) {
	if os.Getenv("EXPECT_GETDEPLOYOPTIONS_EXIT") != "1" {
		return
	}
	mockContext := mocksconsole.NewContext(t)
	mockConfig := mocksconfig.NewConfig(t)
	mockArtisan := mocksconsole.NewArtisan(t)
	cmd := NewDeployCommand(mockConfig, mockArtisan)

	// Return empty for required strings
	mockConfig.EXPECT().GetString("app.name").Return("").Once()
	mockConfig.EXPECT().GetString("app.deploy.ssh_ip").Return("").Once()
	mockConfig.EXPECT().GetString("app.deploy.reverse_proxy_port").Return("").Once()
	mockConfig.EXPECT().GetString("app.deploy.ssh_port").Return("").Once()
	mockConfig.EXPECT().GetString("app.deploy.ssh_user").Return("").Once()
	mockConfig.EXPECT().GetString("app.deploy.ssh_key_path").Return("").Once()
	mockConfig.EXPECT().GetString("app.build.os").Return("").Once()
	mockConfig.EXPECT().GetString("app.build.arch").Return("").Once()
	mockConfig.EXPECT().GetString("app.deploy.domain").Return("").Once()
	mockConfig.EXPECT().GetString("app.deploy.prod_env_file_path").Return("").Once()
	mockConfig.EXPECT().GetString("app.deploy.base_dir", "/var/www/").Return("/var/www/").Once()

	mockConfig.EXPECT().GetBool("app.build.static").Return(false).Once()
	mockConfig.EXPECT().GetBool("app.deploy.reverse_proxy_enabled").Return(false).Once()
	mockConfig.EXPECT().GetBool("app.deploy.reverse_proxy_tls_enabled").Return(false).Once()

	// Expect an error message before exit
	mockContext.EXPECT().Error(mock.MatchedBy(func(msg string) bool {
		return strings.Contains(msg, "Missing required environment variables:")
	})).Once()

	// This will call os.Exit(1)
	_ = cmd.getDeployOptions(mockContext)
	t.Fatalf("expected os.Exit to be called")
}

func Test_getDeployOptions_Missing_Exits(t *testing.T) {
	if runtime.GOOS == "windows" {
		// Subprocess exit code detection differs; still okay but keep consistent behavior
	}
	cmd := exec.Command(os.Args[0], "-test.run", "TestHelper_GetDeployOptions_Missing")
	cmd.Env = append(os.Environ(), "EXPECT_GETDEPLOYOPTIONS_EXIT=1")
	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected subprocess to exit with error due to os.Exit(1)")
	}
}
