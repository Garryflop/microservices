package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DBDSN                  string
	PaymentGRPCAddr        string
	Port                   string
	GRPCPort               string
	RedisAddr              string
	CacheTTLSeconds        int
	RateLimitMax           int
	RateLimitWindowSeconds int
}

func Load() *Config {
	loadEnvFile(".env")

	return &Config{
		DBDSN:                  getEnv("DB_DSN", "postgres://postgres:postgres@localhost:5432/orders_db?sslmode=disable"),
		PaymentGRPCAddr:        getEnv("PAYMENT_GRPC_ADDR", "localhost:50051"),
		Port:                   getEnv("PORT", "8081"),
		GRPCPort:               getEnv("GRPC_PORT", "50052"),
		RedisAddr:              getEnv("REDIS_ADDR", "localhost:6379"),
		CacheTTLSeconds:        getEnvInt("CACHE_TTL_SECONDS", 300),
		RateLimitMax:           getEnvInt("RATE_LIMIT_MAX", 10),
		RateLimitWindowSeconds: getEnvInt("RATE_LIMIT_WINDOW_SECONDS", 60),
	}
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func loadEnvFile(filename string) {
	file, err := os.Open(filename)
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
