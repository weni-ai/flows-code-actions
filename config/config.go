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
	Redis              string
	RateLimiterCode    RateLimiterConfig
	Cleaner            CleanerConfig
	Blacklist          string
	Skiplist           string

	HealthCheckCacheTime int64
}

type RateLimiterConfig struct {
	MaxRequests int
	Window      int
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
	URI                      string
	Name                     string
	Timeout                  int64 // Mongodb start timeout in seconds, this timeout is only for the initial connection, after the connection is established, the timeout is not applied anymore and for retry is used the retry options of the mongodb driver
	MaxRetries               int   // Maximum number of connection retry attempts
	MaxPoolSize              int   // Maximum number of connections in the connection pool
	MinPoolSize              int   // Minimum number of connections in the connection pool
	ConnectTimeoutSeconds    int   // Connection timeout in seconds
	ServerSelectionTimeout   int   // Server selection timeout in seconds
	SocketTimeoutSeconds     int   // Socket timeout in seconds
	HeartbeatIntervalSeconds int   // Heartbeat interval in seconds for topology monitoring
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
		HTTP:            LoadHTTPConfig(),
		DB:              LoadDBConfig(),
		OIDC:            LoadOIDCConfig(),
		AuthToken:       Getenv("FLOWS_CODE_ACTIONS_AUTH_TOKEN", ""),
		Environment:     Getenv("FLOWS_CODE_ACTIONS_ENVIRONMENT", "local"),
		LogLevel:        Getenv("FLOWS_CODE_ACTIONS_LOG_LEVEL", "debug"),
		SentryDSN:       Getenv("FLOWS_CODE_ACTIONS_SENTRY_DSN", ""),
		EDA:             LoadEDAConfig(),
		Redis:           Getenv("FLOWS_CODE_ACTIONS_REDIS", "redis://localhost:6379/10"),
		RateLimiterCode: LoadRateLimiterCodeConfig(),
		Cleaner:         NewCleanerConfig(),
		Blacklist:       Getenv("FLOWS_CODE_ACTIONS_BLACKLIST", ""),
		Skiplist:        Getenv("FLOWS_CODE_ACTIONS_SKIPLIST", ""),

		HealthCheckCacheTime: GetenvInt64("FLOWS_CODE_ACTIONS_HEALTH_CHECK_CACHE_TIME", 3),
	}
}

func LoadRateLimiterCodeConfig() RateLimiterConfig {
	maxRequests, err := strconv.Atoi(Getenv("FLOWS_CODE_ACTIONS_CODE_LIMITER_MAX_REQUESTS", "600"))
	if err != nil {
		maxRequests = 600
	}
	window, err := strconv.Atoi(Getenv("FLOWS_CODE_ACTIONS_CODE_LIMITER_WINDOW_WINDOW", "60"))
	if err != nil {
		window = 60
	}
	return RateLimiterConfig{
		MaxRequests: maxRequests,
		Window:      window,
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
		Getenv("FLOWS_CODE_ACTIONS_MONGO_DB_TIMEOUT", "35"), 10, 64)
	if timeout == 0 {
		timeout = 35
	}

	maxRetries, err := strconv.Atoi(Getenv("FLOWS_CODE_ACTIONS_MONGO_DB_MAX_RETRIES", "3"))
	if err != nil {
		maxRetries = 3
	}

	maxPoolSize, err := strconv.Atoi(Getenv("FLOWS_CODE_ACTIONS_MONGO_DB_MAX_POOL_SIZE", "100"))
	if err != nil {
		maxPoolSize = 100
	}

	minPoolSize, err := strconv.Atoi(Getenv("FLOWS_CODE_ACTIONS_MONGO_DB_MIN_POOL_SIZE", "5"))
	if err != nil {
		minPoolSize = 5
	}

	connectTimeout, err := strconv.Atoi(Getenv("FLOWS_CODE_ACTIONS_MONGO_DB_CONNECT_TIMEOUT", "30"))
	if err != nil {
		connectTimeout = 30
	}

	serverSelectionTimeout, err := strconv.Atoi(Getenv("FLOWS_CODE_ACTIONS_MONGO_DB_SERVER_SELECTION_TIMEOUT", "30"))
	if err != nil {
		serverSelectionTimeout = 30
	}

	socketTimeout, err := strconv.Atoi(Getenv("FLOWS_CODE_ACTIONS_MONGO_DB_SOCKET_TIMEOUT", "30"))
	if err != nil {
		socketTimeout = 30
	}

	heartbeatInterval, err := strconv.Atoi(Getenv("FLOWS_CODE_ACTIONS_MONGO_DB_HEARTBEAT_INTERVAL", "10"))
	if err != nil {
		heartbeatInterval = 10
	}

	return DBConfig{
		URI:                      Getenv("FLOWS_CODE_ACTIONS_MONGO_DB_URI", "mongodb://localhost:27017"),
		Name:                     Getenv("FLOWS_CODE_ACTIONS_MONGO_DB_NAME", "code-actions"),
		Timeout:                  timeout,
		MaxRetries:               maxRetries,
		MaxPoolSize:              maxPoolSize,
		MinPoolSize:              minPoolSize,
		ConnectTimeoutSeconds:    connectTimeout,
		ServerSelectionTimeout:   serverSelectionTimeout,
		SocketTimeoutSeconds:     socketTimeout,
		HeartbeatIntervalSeconds: heartbeatInterval,
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

func GetenvInt64(key string, defval int64) int64 {
	val := os.Getenv(key)
	if val == "" {
		return defval
	}
	intval, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return defval
	}
	return intval
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

func (c *Config) GetSkipListTerms() []string {
	var skipListTerms []string
	skiplist := strings.Split(c.Skiplist, ",")
	for _, term := range skiplist {
		if term != "" {
			skipListTerms = append(skipListTerms, strings.TrimSpace(term))
		}
	}
	sort.Strings(skipListTerms)
	return skipListTerms
}
