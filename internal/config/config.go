package config

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

// Config represents the complete gateway configuration
type Config struct {
	mu       sync.RWMutex
	raw      *RawConfig
	listener []chan<- *RawConfig
}

// RawConfig represents the raw YAML configuration structure
type RawConfig struct {
	Gateway       GatewayConfig       `yaml:"gateway"`
	Server        ServerConfig        `yaml:"server"`
	Admin         AdminConfig         `yaml:"admin"`
	Database      DatabaseConfig      `yaml:"database"`
	Cache         CacheConfig         `yaml:"cache"`
	Logging       LoggingConfig       `yaml:"logging"`
	Tracing       TracingConfig       `yaml:"tracing"`
	Metrics       MetricsConfig       `yaml:"metrics"`
	Health        HealthConfig        `yaml:"health"`
	RateLimit     RateLimitConfig     `yaml:"rate_limit"`
	CircuitBreaker CircuitBreakerConfig `yaml:"circuit_breaker"`
	Proxy         ProxyConfig         `yaml:"proxy"`
	Retry         RetryConfig         `yaml:"retry"`
	Plugins       PluginsConfig       `yaml:"plugins"`
}

// GatewayConfig holds gateway metadata
type GatewayConfig struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host           string        `yaml:"host"`
	Port           int           `yaml:"port"`
	ReadTimeout    time.Duration `yaml:"read_timeout"`
	WriteTimeout   time.Duration `yaml:"write_timeout"`
	IdleTimeout    time.Duration `yaml:"idle_timeout"`
	MaxHeaderBytes int           `yaml:"max_header_bytes"`
	TLS            TLSConfig     `yaml:"tls"`
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	Enabled   bool   `yaml:"enabled"`
	CertFile  string `yaml:"cert_file"`
	KeyFile   string `yaml:"key_file"`
	MinVersion string `yaml:"min_version"`
}

// AdminConfig holds admin API configuration
type AdminConfig struct {
	Enabled   bool     `yaml:"enabled"`
	Port      int      `yaml:"port"`
	Host      string   `yaml:"host"`
	APIKey    string   `yaml:"api_key"`
	AllowedIPs []string `yaml:"allowed_ips"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Postgres PostgresConfig `yaml:"postgres"`
}

// PostgresConfig holds PostgreSQL connection configuration
type PostgresConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	Name            string        `yaml:"name"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	SSLMode         string        `yaml:"ssl_mode"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Enabled  bool          `yaml:"enabled"`
	Type     string        `yaml:"type"`
	TTL      time.Duration `yaml:"ttl"`
	MaxItems int           `yaml:"max_items"`
	Redis    RedisConfig   `yaml:"redis"`
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"pool_size"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level          string `yaml:"level"`
	Format         string `yaml:"format"`
	Output         string `yaml:"output"`
	AccessLog      bool   `yaml:"access_log"`
	AccessLogFormat string `yaml:"access_log_format"`
}

// TracingConfig holds distributed tracing configuration
type TracingConfig struct {
	Enabled     bool    `yaml:"enabled"`
	ServiceName string  `yaml:"service_name"`
	Exporter    string  `yaml:"exporter"`
	Endpoint    string  `yaml:"endpoint"`
	SamplingRatio float64 `yaml:"sampling_ratio"`
}

// MetricsConfig holds Prometheus metrics configuration
type MetricsConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Port      int    `yaml:"port"`
	Path      string `yaml:"path"`
	Namespace string `yaml:"namespace"`
	Subsystem string `yaml:"subsystem"`
}

// HealthConfig holds health check configuration
type HealthConfig struct {
	Enabled                 bool          `yaml:"enabled"`
	Path                    string        `yaml:"path"`
	ReadinessPath           string        `yaml:"readiness_path"`
	LivenessPath            string        `yaml:"liveness_path"`
	UpstreamCheckInterval   time.Duration `yaml:"upstream_check_interval"`
	UpstreamCheckTimeout    time.Duration `yaml:"upstream_check_timeout"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled      bool   `yaml:"enabled"`
	DefaultRPS   int    `yaml:"default_rps"`
	DefaultBurst int    `yaml:"default_burst"`
	StoreType    string `yaml:"store_type"`
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	Enabled          bool          `yaml:"enabled"`
	FailureThreshold int           `yaml:"failure_threshold"`
	SuccessThreshold int           `yaml:"success_threshold"`
	Timeout          time.Duration `yaml:"timeout"`
	HalfOpenMax      int           `yaml:"half_open_max"`
}

// ProxyConfig holds reverse proxy configuration
type ProxyConfig struct {
	Transport  TransportConfig  `yaml:"transport"`
	WebSocket  WebSocketConfig  `yaml:"websocket"`
	GRPC       GRPCConfig       `yaml:"grpc"`
}

// TransportConfig holds HTTP transport configuration
type TransportConfig struct {
	MaxIdleConns          int           `yaml:"max_idle_conns"`
	MaxIdleConnsPerHost   int           `yaml:"max_idle_conns_per_host"`
	MaxConnsPerHost       int           `yaml:"max_conns_per_host"`
	IdleConnTimeout       time.Duration `yaml:"idle_conn_timeout"`
	TLSHandshakeTimeout   time.Duration `yaml:"tls_handshake_timeout"`
	ResponseHeaderTimeout time.Duration `yaml:"response_header_timeout"`
	ExpectContinueTimeout time.Duration `yaml:"expect_continue_timeout"`
	DisableKeepAlives     bool          `yaml:"disable_keep_alives"`
	DisableCompression    bool          `yaml:"disable_compression"`
}

// WebSocketConfig holds WebSocket proxy configuration
type WebSocketConfig struct {
	Enabled          bool   `yaml:"enabled"`
	ReadBufferSize   int    `yaml:"read_buffer_size"`
	WriteBufferSize  int    `yaml:"write_buffer_size"`
	HandshakeTimeout time.Duration `yaml:"handshake_timeout"`
}

// GRPCConfig holds gRPC proxy configuration
type GRPCConfig struct {
	Enabled                  bool `yaml:"enabled"`
	MaxRequestHeaderListSize uint32 `yaml:"max_request_header_list_size"`
}

// RetryConfig holds retry configuration
type RetryConfig struct {
	Enabled             bool          `yaml:"enabled"`
	MaxRetries          int           `yaml:"max_retries"`
	BackoffBase         time.Duration `yaml:"backoff_base"`
	BackoffMax          time.Duration `yaml:"backoff_max"`
	BackoffMultiplier   float64       `yaml:"backoff_multiplier"`
	RetryableStatusCodes []int        `yaml:"retryable_status_codes"`
}

// PluginsConfig holds plugin system configuration
type PluginsConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Directory  string `yaml:"directory"`
	WASMEnabled bool  `yaml:"wasm_enabled"`
}

// New creates a new Config instance
func New() *Config {
	return &Config{
		raw: &RawConfig{},
	}
}

// Load loads configuration from a YAML file
func (c *Config) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var raw RawConfig
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Override with environment variables if set
	c.applyEnvOverrides(&raw)

	if err := raw.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	c.mu.Lock()
	c.raw = &raw
	c.mu.Unlock()

	return nil
}

// Watch watches for configuration file changes and hot-reloads
func (c *Config) Watch(path string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	go func() {
		defer watcher.Close()

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					if err := c.Reload(path); err != nil {
						// Log error but don't fail - continue with old config
						fmt.Printf("failed to reload config: %v\n", err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("config watcher error: %v\n", err)
			}
		}
	}()

	if err := watcher.Add(path); err != nil {
		return fmt.Errorf("failed to watch config file: %w", err)
	}

	return nil
}

// Reload reloads configuration from file
func (c *Config) Reload(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var raw RawConfig
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := raw.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	c.mu.Lock()
	c.raw = &raw
	listeners := make([]chan<- *RawConfig, len(c.listener))
	copy(listeners, c.listener)
	c.mu.Unlock()

	// Notify listeners
	for _, ch := range listeners {
		select {
		case ch <- &raw:
		default:
		}
	}

	return nil
}

// Get returns the current raw configuration (thread-safe)
func (c *Config) Get() *RawConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.raw
}

// OnChange registers a listener for configuration changes
func (c *Config) OnChange(ch chan<- *RawConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.listener = append(c.listener, ch)
}

// Validate validates the configuration
func (rc *RawConfig) Validate() error {
	if rc.Server.Port <= 0 || rc.Server.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535")
	}

	if rc.Admin.Enabled && (rc.Admin.Port <= 0 || rc.Admin.Port > 65535) {
		return fmt.Errorf("admin port must be between 1 and 65535")
	}

	if rc.Admin.Enabled && rc.Admin.APIKey == "" {
		return fmt.Errorf("admin api_key cannot be empty when admin is enabled")
	}

	if rc.Metrics.Enabled && (rc.Metrics.Port <= 0 || rc.Metrics.Port > 65535) {
		return fmt.Errorf("metrics port must be between 1 and 65535")
	}

	return nil
}

// applyEnvOverrides overrides config values with environment variables
func (c *Config) applyEnvOverrides(raw *RawConfig) {
	// Database overrides
	if val := os.Getenv("SETU_DB_HOST"); val != "" {
		raw.Database.Postgres.Host = val
	}
	if val := os.Getenv("SETU_DB_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			raw.Database.Postgres.Port = port
		}
	}
	if val := os.Getenv("SETU_DB_NAME"); val != "" {
		raw.Database.Postgres.Name = val
	}
	if val := os.Getenv("SETU_DB_USER"); val != "" {
		raw.Database.Postgres.User = val
	}
	if val := os.Getenv("SETU_DB_PASSWORD"); val != "" {
		raw.Database.Postgres.Password = val
	}
	if val := os.Getenv("SETU_DB_SSL_MODE"); val != "" {
		raw.Database.Postgres.SSLMode = val
	}

	// Server overrides
	if val := os.Getenv("SETU_SERVER_HOST"); val != "" {
		raw.Server.Host = val
	}
	if val := os.Getenv("SETU_SERVER_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			raw.Server.Port = port
		}
	}

	// Admin overrides
	if val := os.Getenv("SETU_ADMIN_API_KEY"); val != "" {
		raw.Admin.APIKey = val
	}

	// Logging overrides
	if val := os.Getenv("SETU_LOG_LEVEL"); val != "" {
		raw.Logging.Level = val
	}
	if val := os.Getenv("SETU_LOG_FORMAT"); val != "" {
		raw.Logging.Format = val
	}

	// Redis overrides
	if val := os.Getenv("SETU_REDIS_HOST"); val != "" {
		raw.Cache.Redis.Host = val
	}
	if val := os.Getenv("SETU_REDIS_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			raw.Cache.Redis.Port = port
		}
	}
	if val := os.Getenv("SETU_REDIS_PASSWORD"); val != "" {
		raw.Cache.Redis.Password = val
	}
}
