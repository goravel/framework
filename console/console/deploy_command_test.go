package console

import (
	"encoding/base64"
	"os"
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
	cmd := setupServerCommand("myapp", "203.0.113.10", "9000", "22", "ubuntu", "~/.ssh/id", "", false, false)
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
	cmd := setupServerCommand("myapp", "203.0.113.10", "9000", "22", "ubuntu", "~/.ssh/id", "", true, false)
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
	cmd := setupServerCommand("myapp", "203.0.113.10", "9000", "22", "ubuntu", "~/.ssh/id", "example.com", true, true)
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
}

func Test_uploadFilesCommand_AllArtifacts(t *testing.T) {
	cmd := uploadFilesCommand("myapp", "203.0.113.10", "22", "ubuntu", "~/.ssh/id", ".env.production", true, true, true, true, true)
	require.NotNil(t, cmd)
	if runtime.GOOS == "windows" {
		t.Skip("Skipping script content assertions on Windows shell")
	}
	script := cmd.Args[2]
	appDir := "/var/www/myapp"
	// Binary upload and backup
	assert.Contains(t, script, "scp -o StrictHostKeyChecking=no -i \"~/.ssh/id\" -P 22 \"myapp\" ubuntu@203.0.113.10:"+appDir+"/main.new")
	assert.Contains(t, script, "if [ -f "+appDir+"/main ]; then sudo mv "+appDir+"/main "+appDir+"/main.prev; fi; sudo mv "+appDir+"/main.new "+appDir+"/main && sudo chmod +x "+appDir+"/main")
	// .env atomic rename
	assert.Contains(t, script, ".env.new")
	assert.Contains(t, script, "/.env'")
	// Directories
	assert.Contains(t, script, "scp -o StrictHostKeyChecking=no -i \"~/.ssh/id\" -P 22 -r \"public\" ubuntu@203.0.113.10:"+appDir)
	assert.Contains(t, script, "scp -o StrictHostKeyChecking=no -i \"~/.ssh/id\" -P 22 -r \"storage\" ubuntu@203.0.113.10:"+appDir)
	assert.Contains(t, script, "scp -o StrictHostKeyChecking=no -i \"~/.ssh/id\" -P 22 -r \"resources\" ubuntu@203.0.113.10:"+appDir)
}

func Test_restartServiceCommand(t *testing.T) {
	cmd := restartServiceCommand("myapp", "203.0.113.10", "22", "ubuntu", "~/.ssh/id")
	require.NotNil(t, cmd)
	if runtime.GOOS == "windows" {
		t.Skip("Skipping script content assertions on Windows shell")
	}
	script := cmd.Args[2]
	assert.Contains(t, script, "systemctl daemon-reload")
	assert.Contains(t, script, "systemctl restart myapp || sudo systemctl start myapp")
}

func Test_rollbackCommand(t *testing.T) {
	cmd := rollbackCommand("myapp", "203.0.113.10", "22", "ubuntu", "~/.ssh/id")
	require.NotNil(t, cmd)
	if runtime.GOOS == "windows" {
		t.Skip("Skipping script content assertions on Windows shell")
	}
	script := cmd.Args[2]
	assert.Contains(t, script, "main.prev")
	assert.Contains(t, script, "systemctl restart myapp || sudo systemctl start myapp")
}

func Test_getStringEnv_and_getBoolEnv(t *testing.T) {
	mc := &mocksconfig.Config{}
	// String as string
	mc.EXPECT().Env("STR").Return("value").Once()
	assert.Equal(t, "value", getStringEnv(mc, "STR"))
	// String as non-string type
	mc.EXPECT().Env("NUM").Return(123).Once()
	assert.Equal(t, "123", getStringEnv(mc, "NUM"))
	// Missing
	mc.EXPECT().Env("MISSING").Return(nil).Once()
	assert.Equal(t, "", getStringEnv(mc, "MISSING"))

	// Bool parsing
	mc.EXPECT().Env("BOOL1").Return(true).Once()
	assert.True(t, getBoolEnv(mc, "BOOL1"))
	mc.EXPECT().Env("BOOL2").Return("true").Once()
	assert.True(t, getBoolEnv(mc, "BOOL2"))
	mc.EXPECT().Env("BOOL3").Return("1").Once()
	assert.True(t, getBoolEnv(mc, "BOOL3"))
	mc.EXPECT().Env("BOOL4").Return("no").Once()
	assert.False(t, getBoolEnv(mc, "BOOL4"))
	mc.EXPECT().Env("BOOL5").Return(nil).Once()
	assert.False(t, getBoolEnv(mc, "BOOL5"))
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

	mc := &mocksconsole.Context{}
	mc.EXPECT().Option("only").Return("").Once()
	hasMain, hasEnv, hasPub, hasStor, hasRes := getWhichFilesToUpload(mc, "myapp", ".env.production")
	assert.True(t, hasMain)
	assert.True(t, hasEnv)
	assert.True(t, hasPub)
	assert.True(t, hasStor)
	assert.True(t, hasRes)

	// Now test filter: only main and env
	mc2 := &mocksconsole.Context{}
	mc2.EXPECT().Option("only").Return("main,env").Once()
	hasMain, hasEnv, hasPub, hasStor, hasRes = getWhichFilesToUpload(mc2, "myapp", ".env.production")
	assert.True(t, hasMain)
	assert.True(t, hasEnv)
	assert.False(t, hasPub)
	assert.False(t, hasStor)
	assert.False(t, hasRes)
}

func Test_Handle_Rollback_ShortCircuit(t *testing.T) {
	// We only test rollback path to avoid executing remote checks.
	mc := &mocksconsole.Context{}
	cfg := &mocksconfig.Config{}
	cmd := NewDeployCommand(cfg)

	// Minimal required envs for getAllOptions (will not be used deeply due to rollback)
	cfg.EXPECT().GetString("app.name").Return("myapp").Once()
	cfg.EXPECT().Env("DEPLOY_IP_ADDRESS").Return("203.0.113.10").Once()
	cfg.EXPECT().Env("DEPLOY_APP_PORT").Return("9000").Once()
	cfg.EXPECT().Env("DEPLOY_SSH_PORT").Return("22").Once()
	cfg.EXPECT().Env("DEPLOY_SSH_USER").Return("ubuntu").Once()
	cfg.EXPECT().Env("DEPLOY_SSH_KEY_PATH").Return("~/.ssh/id").Once()
	cfg.EXPECT().Env("DEPLOY_OS").Return("linux").Once()
	cfg.EXPECT().Env("DEPLOY_ARCH").Return("amd64").Once()
	cfg.EXPECT().Env("DEPLOY_DOMAIN").Return("").Once()
	cfg.EXPECT().Env("DEPLOY_PROD_ENV_FILE_PATH").Return(".env.production").Once()
	cfg.EXPECT().Env("DEPLOY_STATIC").Return(false).Once()
	cfg.EXPECT().Env("DEPLOY_REVERSE_PROXY_ENABLED").Return(false).Once()
	cfg.EXPECT().Env("DEPLOY_REVERSE_PROXY_TLS_ENABLED").Return(false).Once()

	mc.EXPECT().OptionBool("rollback").Return(true).Once()
	mc.EXPECT().Spinner("Rolling back...", mock.Anything).Return(nil).Once()
	mc.EXPECT().Info("Rollback successful.").Once()

	assert.Nil(t, cmd.Handle(mc))
}
