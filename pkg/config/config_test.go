package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/activatedio/apiinfra/pkg/config"
	"github.com/activatedio/cs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sample struct {
	Host string
	Port int
	TLS  bool
}

func writeFile(t *testing.T, dir, name, body string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(p, []byte(body), 0o600))
	return p
}

func TestNewConfig_YAML(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(config.KeyConfigPaths, "")
	p := writeFile(t, dir, "cfg.yaml", "server:\n  host: 0.0.0.0\n  port: 8080\n  tls: true\n")

	c := config.NewConfig(p)
	got := cs.MustGet[sample](c, config.PrefixServer)
	assert.Equal(t, "0.0.0.0", got.Host)
	assert.Equal(t, 8080, got.Port)
	assert.True(t, got.TLS)
}

func TestNewConfig_JSON(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(config.KeyConfigPaths, "")
	p := writeFile(t, dir, "cfg.json", `{"server":{"host":"127.0.0.1","port":9090}}`)

	c := config.NewConfig(p)
	got := cs.MustGet[sample](c, config.PrefixServer)
	assert.Equal(t, "127.0.0.1", got.Host)
	assert.Equal(t, 9090, got.Port)
}

func TestNewConfig_CONFIG_PATHS_Env(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "cfg.yaml", "server:\n  host: viaenv\n  port: 1234\n")
	t.Setenv(config.KeyConfigPaths, p)

	c := config.NewConfig()
	got := cs.MustGet[sample](c, config.PrefixServer)
	assert.Equal(t, "viaenv", got.Host)
	assert.Equal(t, 1234, got.Port)
}

func TestNewConfig_EnvOverridesFile(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "cfg.yaml", "server:\n  host: fromfile\n  port: 1\n")
	t.Setenv(config.KeyConfigPaths, "")
	t.Setenv("SERVER_HOST", "fromenv")
	t.Setenv("SERVER_PORT", "9999")

	c := config.NewConfig(p)
	got := cs.MustGet[sample](c, config.PrefixServer)
	assert.Equal(t, "fromenv", got.Host)
	assert.Equal(t, 9999, got.Port)
}

func TestNewConfig_UnknownExtensionIgnored(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(config.KeyConfigPaths, "")
	yamlPath := writeFile(t, dir, "cfg.yaml", "server:\n  host: yaml-wins\n  port: 1\n")
	txtPath := writeFile(t, dir, "ignored.txt", "this should not be parsed")

	c := config.NewConfig(yamlPath, txtPath)
	got := cs.MustGet[sample](c, config.PrefixServer)
	assert.Equal(t, "yaml-wins", got.Host)
}
