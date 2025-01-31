package config

import (
	"os"
	"sort"
	"strconv"
	"strings"
)

type Config struct {
	HTTP               HTTPConfig
	DB                 DBConfig
	OIDC               OIDCConfig
	AuthToken          string
	Environment        string
	LogLevel           string
	SentryDSN          string
	ResourceManagement ResourceConfig
	EDA                EDAConfig
	RedisURL           string
	Cleaner            CleanerConfig
  Blacklist          string
}

type CleanerConfig struct {
	ScheduleTime    string
	RetentionPeriod string
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

// ResourceConfig represents a resource configuration for each code execution in a project cgroup
type ResourceConfig struct {
	Enabled bool
	CPU     CPUConfig
	Memory  MemoryConfig
}

// CPUConfig represents a resource configuration for each code execution in a project cgroup related to cpu resource management
type CPUConfig struct {
	Shares *uint64
	Quota  *int64
}

// MemoryConfig represents a resource configuration for each code execution in a project cgroup related to memory resource management
type MemoryConfig struct {
	Limit       *int64
	Reservation *int64
}

type EDAConfig struct {
	RabbitmqURL            string
	ProjectExchangeName    string
	ProjectQueueName       string
	PermissionExchangeName string
	PermissionQueueName    string
}

func NewConfig() *Config {
	return &Config{
		HTTP:        LoadHTTPConfig(),
		DB:          LoadDBConfig(),
		OIDC:        LoadOIDCConfig(),
		AuthToken:   Getenv("FLOWS_CODE_ACTIONS_AUTH_TOKEN", ""),
		Environment: Getenv("FLOWS_CODE_ACTIONS_ENVIRONMENT", "local"),
		LogLevel:    Getenv("FLOWS_CODE_ACTIONS_LOG_LEVEL", "debug"),
		SentryDSN:   Getenv("FLOWS_CODE_ACTIONS_SENTRY_DSN", ""),
		EDA:         LoadEDAConfig(),
		RedisURL:    Getenv("FLOWS_CODE_ACTIONS_REDIS_URL", "redis://localhost:6379/15"),
		Cleaner:     NewCleanerConfig(),
		Blacklist:   Getenv("FLOWS_CODE_ACTIONS_BLACKLIST", ""),
	}
}

func NewCleanerConfig() CleanerConfig {
	scheduleTime := Getenv("FLOWS_CODE_ACTIONS_CLEANER_SCHEDULE_TIME", "01:00")
	retentionPeriod := Getenv("FLOWS_CODE_ACTIONS_CLEANER_RETENTION_PERIOD", "30")
	return CleanerConfig{
		ScheduleTime:    scheduleTime,
		RetentionPeriod: retentionPeriod,
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

func LoadResourcesConfig() ResourceConfig {
	enabled, err := strconv.ParseBool(Getenv("FLOWS_CODE_ACTIONS_RESOURCE_ENABLED", "false"))
	if err != nil {
		enabled = false
	}

	cpuShares, _ := strconv.ParseUint(Getenv("FLOWS_CODE_ACTIONS_CPU_SHARES", "0"), 10, 64)
	cpuQuota, _ := strconv.ParseInt(Getenv("FLOWS_CODE_ACTIONS_CPU_QUOTA", "0"), 10, 64)
	cpu := CPUConfig{}
	if cpuShares != 0 {
		cpu.Shares = &cpuShares
	}
	if cpuQuota != 0 {
		cpu.Quota = &cpuQuota
	}

	memLimit, _ := strconv.ParseInt(Getenv("FLOWS_CODE_ACTIONS_MEMORY_LIMIT", "0"), 10, 64)
	memRes, _ := strconv.ParseInt(Getenv("FLOWS_CODE_ACTIONS_MEMORY_RESERVATION", "0"), 10, 64)
	memory := MemoryConfig{}
	if memLimit != 0 {
		memory.Limit = &memLimit
	}
	if memRes != 0 {
		memory.Reservation = &memRes
	}
	return ResourceConfig{
		Enabled: enabled,
		CPU:     cpu,
		Memory:  memory,
	}
}

func LoadEDAConfig() EDAConfig {
	rabbitmqURL := Getenv("FLOWS_CODE_ACTIONS_RABBITMQ_URL", "")
	projectExchangeName := Getenv("FLOWS_CODE_ACTIONS_PROJECT_EXCHANGE", "")
	projectQueueName := Getenv("FLOWS_CODE_ACTIONS_PROJECT_QUEUE", "")
	permissionExchangeName := Getenv("FLOWS_CODE_ACTIONS_PERMISSION_EXCHANGE", "")
	permissionQueueName := Getenv("FLOW_CODE_ACTIONS_PERMISSION_QUEUE", "")
	return EDAConfig{
		RabbitmqURL:            rabbitmqURL,
		ProjectExchangeName:    projectExchangeName,
		ProjectQueueName:       projectQueueName,
		PermissionExchangeName: permissionExchangeName,
		PermissionQueueName:    permissionQueueName,
	}
}

func Getenv(key string, defval string) string {
	val := os.Getenv(key)
	if val == "" {
		return defval
	}
	return val
}

func (c *Config) GetBlackListTerms() []string {
	var blackListTerms []string
	blacklist := strings.Split(c.Blacklist, ",")
	for _, term := range blacklist {
		if term != "" {
			blackListTerms = append(blackListTerms, strings.TrimSpace(term))
		}
	}
	sort.Strings(blackListTerms)
	return blackListTerms
}
