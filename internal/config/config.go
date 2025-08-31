package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Database      DatabaseConfig
	Elasticsearch ElasticsearchConfig
	Server        ServerConfig
	Logger        LoggerConfig
	Performance   PerformanceConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type ElasticsearchConfig struct {
	URL string
}

type ServerConfig struct {
	Port    int
	GinMode string
}

type LoggerConfig struct {
	Level string
}

type PerformanceConfig struct {
	ConcurrentUsers int
	Duration        string
}

func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found, using environment variables")
	}

	config := &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "ecommerce"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Elasticsearch: ElasticsearchConfig{
			URL: getEnv("ELASTICSEARCH_URL", "http://localhost:9200"),
		},
		Server: ServerConfig{
			Port:    getEnvInt("SERVER_PORT", 8080),
			GinMode: getEnv("GIN_MODE", "debug"),
		},
		Logger: LoggerConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
		Performance: PerformanceConfig{
			ConcurrentUsers: getEnvInt("PERF_TEST_CONCURRENT_USERS", 100),
			Duration:        getEnv("PERF_TEST_DURATION", "30s"),
		},
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}