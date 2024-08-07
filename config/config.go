package config

import (
	"os"
	"strconv"
)

type Config struct {
	HTTP        HTTPConfig
	DB          DBConfig
	OIDC        OIDCConfig
	AuthToken   string
	Environment string
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

type OIDCConfig struct {
	AuthEnabled bool
	Realm       string
	Host        string
}

func NewConfig() *Config {
	return &Config{
		HTTP:        LoadHTTPConfig(),
		DB:          LoadDBConfig(),
		OIDC:        LoadOIDCConfig(),
		AuthToken:   Getenv("FLOWS_CODE_ACTIONS_AUTH_TOKEN", ""),
		Environment: Getenv("FLOWS_CODE_ACTIONS_ENVIRONMENT", "local"),
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

func LoadOIDCConfig() OIDCConfig {
	Realm := Getenv("FLOWS_CODE_ACTIONS_OIDC_REALM", "")
	Host := Getenv("FLOWS_CODE_ACTIONS_OIDC_HOST", "")
	Enabled, err := strconv.ParseBool(Getenv("FLOWS_CODE_ACTIONS_OIDC_AUTH_ENABLED", "false"))
	if err != nil {
		Enabled = false
	}
	return OIDCConfig{
		Realm:       Realm,
		Host:        Host,
		AuthEnabled: Enabled,
	}
}

func Getenv(key string, defval string) string {
	val := os.Getenv(key)
	if val == "" {
		return defval
	}
	return val
}
