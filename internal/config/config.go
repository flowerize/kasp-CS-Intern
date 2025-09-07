package config

import (
	"os"
	"strconv"
)

type Config struct {
	Workers   int
	QueueSize int
}

func getIntEnv(key string, defaultVal int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}
	return val
}

func NewConfig() *Config {
	return &Config{
		Workers:   getIntEnv("WORKERS", 4),
		QueueSize: getIntEnv("QUEUE_SIZE", 64),
	}
}
