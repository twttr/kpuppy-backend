package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Admin     AdminConfig
	RateLimit RateLimitConfig
	Sentry    SentryConfig
}

type ServerConfig struct {
	Port           int
	Host           string
	BasePath       string
	AllowedOrigins []string
}

type DatabaseConfig struct {
	Path string
}

type AdminConfig struct {
	Username     string
	PasswordHash string
}

type RateLimitConfig struct {
	Enabled bool
}

type SentryConfig struct {
	DSN         string
	Environment string
}

func Load() *Config {
	loadEnvFile(".env")

	return &Config{
		Server: ServerConfig{
			Port:           getEnvInt("PORT", 8080),
			Host:           getEnv("HOST", "0.0.0.0"),
			BasePath:       getEnv("BASE_PATH", ""),
			AllowedOrigins: getEnvSlice("ALLOWED_ORIGINS", nil),
		},
		Database: DatabaseConfig{
			Path: getEnv("DB_PATH", "./data/kpuppy.db"),
		},
		Admin: AdminConfig{
			Username:     getEnv("ADMIN_USER", "admin"),
			PasswordHash: getEnv("ADMIN_PASS_HASH", ""),
		},
		RateLimit: RateLimitConfig{
			Enabled: getEnvBool("RATE_LIMIT", true),
		},
		Sentry: SentryConfig{
			DSN:         getEnv("SENTRY_DSN", ""),
			Environment: getEnv("SENTRY_ENVIRONMENT", "development"),
		},
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvSlice(key string, defaultVal []string) []string {
	if val := os.Getenv(key); val != "" {
		parts := strings.Split(val, ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			if s := strings.TrimSpace(p); s != "" {
				result = append(result, s)
			}
		}
		return result
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return defaultVal
}

func loadEnvFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		value = strings.Trim(value, `"'`)

		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}
