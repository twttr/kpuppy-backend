package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Defaults(t *testing.T) {
	os.Clearenv()

	cfg := Load()

	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, "./data/kpuppy.db", cfg.Database.Path)
	assert.Equal(t, "admin", cfg.Admin.Username)
	assert.Equal(t, "", cfg.Admin.PasswordHash)
	assert.True(t, cfg.RateLimit.Enabled)
	assert.Equal(t, "", cfg.Sentry.DSN)
	assert.Equal(t, "development", cfg.Sentry.Environment)
}

func TestLoad_FromEnv(t *testing.T) {
	os.Clearenv()
	os.Setenv("PORT", "9000")
	os.Setenv("HOST", "127.0.0.1")
	os.Setenv("DB_PATH", "/tmp/test.db")
	os.Setenv("ADMIN_USER", "superadmin")
	os.Setenv("ADMIN_PASS_HASH", "hashedpassword")
	os.Setenv("RATE_LIMIT", "false")
	os.Setenv("SENTRY_DSN", "https://key@sentry.io/123")
	os.Setenv("SENTRY_ENVIRONMENT", "production")
	defer os.Clearenv()

	cfg := Load()

	assert.Equal(t, 9000, cfg.Server.Port)
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, "/tmp/test.db", cfg.Database.Path)
	assert.Equal(t, "superadmin", cfg.Admin.Username)
	assert.Equal(t, "hashedpassword", cfg.Admin.PasswordHash)
	assert.False(t, cfg.RateLimit.Enabled)
	assert.Equal(t, "https://key@sentry.io/123", cfg.Sentry.DSN)
	assert.Equal(t, "production", cfg.Sentry.Environment)
}

func TestGetEnv(t *testing.T) {
	os.Clearenv()

	assert.Equal(t, "default", getEnv("NONEXISTENT", "default"))

	os.Setenv("TEST_KEY", "value")
	assert.Equal(t, "value", getEnv("TEST_KEY", "default"))
}

func TestGetEnvInt(t *testing.T) {
	os.Clearenv()

	assert.Equal(t, 42, getEnvInt("NONEXISTENT", 42))

	os.Setenv("TEST_INT", "123")
	assert.Equal(t, 123, getEnvInt("TEST_INT", 42))

	os.Setenv("TEST_INT_INVALID", "notanumber")
	assert.Equal(t, 42, getEnvInt("TEST_INT_INVALID", 42))
}

func TestGetEnvBool(t *testing.T) {
	os.Clearenv()

	assert.True(t, getEnvBool("NONEXISTENT", true))
	assert.False(t, getEnvBool("NONEXISTENT", false))

	os.Setenv("TEST_BOOL_TRUE", "true")
	assert.True(t, getEnvBool("TEST_BOOL_TRUE", false))

	os.Setenv("TEST_BOOL_FALSE", "false")
	assert.False(t, getEnvBool("TEST_BOOL_FALSE", true))

	os.Setenv("TEST_BOOL_1", "1")
	assert.True(t, getEnvBool("TEST_BOOL_1", false))

	os.Setenv("TEST_BOOL_INVALID", "notabool")
	assert.True(t, getEnvBool("TEST_BOOL_INVALID", true))
}

func TestLoadEnvFile(t *testing.T) {
	os.Clearenv()

	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")

	envContent := `
# Comment line
PORT=3000
HOST=localhost

DB_PATH=/var/lib/app/db.sqlite
ADMIN_USER="quoted_user"
ADMIN_PASS_HASH='single_quoted'
`
	err := os.WriteFile(envPath, []byte(envContent), 0644)
	require.NoError(t, err)

	loadEnvFile(envPath)

	assert.Equal(t, "3000", os.Getenv("PORT"))
	assert.Equal(t, "localhost", os.Getenv("HOST"))
	assert.Equal(t, "/var/lib/app/db.sqlite", os.Getenv("DB_PATH"))
	assert.Equal(t, "quoted_user", os.Getenv("ADMIN_USER"))
	assert.Equal(t, "single_quoted", os.Getenv("ADMIN_PASS_HASH"))
}

func TestLoadEnvFile_DoesNotOverrideExisting(t *testing.T) {
	os.Clearenv()
	os.Setenv("PORT", "9999")

	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")

	envContent := `PORT=3000`
	err := os.WriteFile(envPath, []byte(envContent), 0644)
	require.NoError(t, err)

	loadEnvFile(envPath)

	assert.Equal(t, "9999", os.Getenv("PORT"))
}

func TestLoadEnvFile_NonExistentFile(t *testing.T) {
	os.Clearenv()
	loadEnvFile("/nonexistent/path/.env")
}

func TestLoadEnvFile_InvalidLines(t *testing.T) {
	os.Clearenv()

	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")

	envContent := `
VALID_KEY=valid_value
invalid_line_without_equals
=no_key
ANOTHER_VALID=value
`
	err := os.WriteFile(envPath, []byte(envContent), 0644)
	require.NoError(t, err)

	loadEnvFile(envPath)

	assert.Equal(t, "valid_value", os.Getenv("VALID_KEY"))
	assert.Equal(t, "value", os.Getenv("ANOTHER_VALID"))
}
