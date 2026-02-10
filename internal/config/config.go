package config

import "os"

type Config struct {
	Port             string
	Environment      string
	ElasticsearchURL string
	RedisHost        string
	RedisPort        string
	RedisPassword    string
	JWTSecret        string
}

func Load() *Config {
	return &Config{
		Port:             getEnv("PORT", "5006"),
		Environment:      getEnv("ENVIRONMENT", "development"),
		ElasticsearchURL: getEnv("ELASTICSEARCH_URL", "http://localhost:9200"),
		RedisHost:        getEnv("REDIS_HOST", "localhost"),
		RedisPort:        getEnv("REDIS_PORT", "6379"),
		RedisPassword:    getEnv("REDIS_PASSWORD", ""),
		JWTSecret:        getEnv("JWT_SECRET", "dev-secret"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
