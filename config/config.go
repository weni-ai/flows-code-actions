package config

import (
	"os"
	"strconv"
)

type Config struct {
	HTTP HTTPConfig
	DB   DBConfig
}

type HTTPConfig struct {
	Host string
	Port string
}

type DBConfig struct {
	URI     string
	Name    string
	Timeout int64
}

func NewConfig() *Config {
	return &Config{
		HTTP: LoadHTTPConfig(),
		DB:   LoadDBConfig(),
	}
}

func LoadHTTPConfig() HTTPConfig {
	return HTTPConfig{
		Host: Getenv("FLOWS_CODE_ACTIONS_HOST", ":"),
		Port: Getenv("FLOWS_CODE_ACTIONS_PORT", "8050"),
	}
}

func LoadDBConfig() DBConfig {
	timeout, _ := strconv.ParseInt(
		Getenv("FLOWS_CODE_ACTIONS_MONGO_DB_TIMEOUT", "15"), 10, 64)
	if timeout == 0 {
		timeout = 15
	}
	return DBConfig{
		URI:     Getenv("FLOWS_CODE_ACTIONS_MONGO_DB_URI", "mongodb://localhost:27017"),
		Name:    Getenv("FLOWS_CODE_ACTIONS_MONGO_DB_NAME", "code-actions"),
		Timeout: timeout,
	}
}

func Getenv(key string, defval string) string {
	val := os.Getenv(key)
	if val == "" {
		return defval
	}
	return val
}
