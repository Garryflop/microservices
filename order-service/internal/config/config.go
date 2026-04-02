package config

import (
	"bufio"
	"os"
	"strings"
)

type Config struct {
	DBDSN             string
	PaymentServiceURL string
	Port              string
}

func Load() *Config {
	loadEnvFile(".env")

	return &Config{
		DBDSN:             getEnv("DB_DSN", "postgres://postgres:postgres@localhost:5432/orders_db?sslmode=disable"),
		PaymentServiceURL: getEnv("PAYMENT_SERVICE_URL", "http://localhost:8082"),
		Port:              getEnv("PORT", "8081"),
	}
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
