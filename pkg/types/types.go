package types

import (
	"context"
	"net/http"
	"time"
)

// Route represents a gateway route configuration
type Route struct {
	ID              string            `json:"id" db:"id"`
	Name            string            `json:"name" db:"name"`
	Description     string            `json:"description" db:"description"`
	Path            string            `json:"path" db:"path"`
	PathType        PathType          `json:"path_type" db:"path_type"`
	Methods         []string          `json:"methods" db:"methods"`
	StripPath       bool              `json:"strip_path" db:"strip_path"`
	PreserveHost    bool              `json:"preserve_host" db:"preserve_host"`
	Enabled         bool              `json:"enabled" db:"enabled"`
	Priority        int               `json:"priority" db:"priority"`
	UpstreamID      string            `json:"upstream_id" db:"upstream_id"`
	AuthChain       []string          `json:"auth_chain" db:"auth_chain"`
	Plugins         []string          `json:"plugins" db:"plugins"`
	RateLimitID     *string           `json:"rate_limit_id,omitempty" db:"rate_limit_id"`
	Transform       TransformConfig   `json:"transform,omitempty" db:"-"`
	Timeout         time.Duration     `json:"timeout" db:"timeout"`
	RetryEnabled    bool              `json:"retry_enabled" db:"retry_enabled"`
	CircuitBreaker  bool              `json:"circuit_breaker" db:"circuit_breaker"`
	CreatedAt       time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at" db:"updated_at"`
}

// PathType defines the type of path matching
type PathType string

const (
	PathTypeExact   PathType = "exact"
	PathTypePrefix  PathType = "prefix"
	PathTypeRegex   PathType = "regex"
	PathTypeWildcard PathType = "wildcard"
)

// TransformConfig holds request/response transformation rules
type TransformConfig struct {
	PathRewrite    *PathRewrite    `json:"path_rewrite,omitempty"`
	HeaderRewrite  *HeaderRewrite  `json:"header_rewrite,omitempty"`
	MethodRewrite  *string         `json:"method_rewrite,omitempty"`
	QueryRewrite   *QueryRewrite   `json:"query_rewrite,omitempty"`
}

// PathRewrite holds path transformation rules
type PathRewrite struct {
	From        string `json:"from"`
	To          string `json:"to"`
	StripPrefix string `json:"strip_prefix,omitempty"`
}

// HeaderRewrite holds header transformation rules
type HeaderRewrite struct {
	Add    map[string]string   `json:"add,omitempty"`
	Remove []string            `json:"remove,omitempty"`
	Rename map[string]string   `json:"rename,omitempty"`
}

// QueryRewrite holds query parameter transformation rules
type QueryRewrite struct {
	Add    map[string]string `json:"add,omitempty"`
	Remove []string          `json:"remove,omitempty"`
}

// Upstream represents a backend service
type Upstream struct {
	ID          string        `json:"id" db:"id"`
	Name        string        `json:"name" db:"name"`
	Description string        `json:"description" db:"description"`
	Targets     []Target      `json:"targets" db:"-"`
	Algorithm   LBAlgorithm   `json:"algorithm" db:"algorithm"`
	HealthCheck HealthCheck   `json:"health_check" db:"-"`
	StickySession bool        `json:"sticky_session" db:"sticky_session"`
	Enabled     bool          `json:"enabled" db:"enabled"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at" db:"updated_at"`
}

// Target represents an upstream backend target
type Target struct {
	ID       string `json:"id" db:"id"`
	UpstreamID string `json:"upstream_id" db:"upstream_id"`
	Host     string `json:"host" db:"host"`
	Port     int    `json:"port" db:"port"`
	Weight   int    `json:"weight" db:"weight"`
	Enabled  bool   `json:"enabled" db:"enabled"`
	Healthy  bool   `json:"healthy" db:"healthy"`
	Metadata map[string]string `json:"metadata,omitempty" db:"-"`
}

// LBAlgorithm defines the load balancing algorithm
type LBAlgorithm string

const (
	LBAlgorithmRoundRobin    LBAlgorithm = "round_robin"
	LBAlgorithmWeightedRR    LBAlgorithm = "weighted_round_robin"
	LBAlgorithmLeastConn     LBAlgorithm = "least_connection"
	LBAlgorithmRandom        LBAlgorithm = "random"
)

// HealthCheck holds upstream health check configuration
type HealthCheck struct {
	Enabled  bool          `json:"enabled"`
	Path     string        `json:"path"`
	Interval time.Duration `json:"interval"`
	Timeout  time.Duration `json:"timeout"`
	HealthyThreshold   int `json:"healthy_threshold"`
	UnhealthyThreshold int `json:"unhealthy_threshold"`
}

// AuthProvider represents an authentication provider
type AuthProvider struct {
	ID          string         `json:"id" db:"id"`
	Name        string         `json:"name" db:"name"`
	Type        AuthProviderType `json:"type" db:"type"`
	Config      AuthConfig     `json:"config" db:"-"`
	Enabled     bool           `json:"enabled" db:"enabled"`
	Priority    int            `json:"priority" db:"priority"`
	Timeout     time.Duration  `json:"timeout" db:"timeout"`
	CacheEnabled bool          `json:"cache_enabled" db:"cache_enabled"`
	CacheTTL    time.Duration  `json:"cache_ttl" db:"cache_ttl"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
}

// AuthProviderType defines the type of authentication provider
type AuthProviderType string

const (
	AuthProviderJWT      AuthProviderType = "jwt"
	AuthProviderAPIKey   AuthProviderType = "api_key"
	AuthProviderOAuth2   AuthProviderType = "oauth2"
	AuthProviderBasic    AuthProviderType = "basic"
	AuthProviderMTLS     AuthProviderType = "mtls"
	AuthProviderHMAC     AuthProviderType = "hmac"
	AuthProviderExternal AuthProviderType = "external"
)

// AuthConfig holds authentication provider configuration
type AuthConfig struct {
	// JWT Config
	JWKSEndpoint string   `json:"jwks_endpoint,omitempty"`
	Secret       string   `json:"secret,omitempty"`
	Algorithms   []string `json:"algorithms,omitempty"`
	Issuer       string   `json:"issuer,omitempty"`
	Audience     string   `json:"audience,omitempty"`
	
	// API Key Config
	HeaderName   string `json:"header_name,omitempty"`
	QueryParam   string `json:"query_param,omitempty"`
	Prefix       string `json:"prefix,omitempty"`
	
	// OAuth2 Config
	IntrospectionURL string `json:"introspection_url,omitempty"`
	ClientID         string `json:"client_id,omitempty"`
	ClientSecret     string `json:"client_secret,omitempty"`
	
	// External Auth Config
	ExternalURL      string `json:"external_url,omitempty"`
	ExternalMethod   string `json:"external_method,omitempty"`
	
	// mTLS Config
	CACertFile     string `json:"ca_cert_file,omitempty"`
	
	// HMAC Config
	SecretKey      string `json:"secret_key,omitempty"`
	HeaderDate     string `json:"header_date,omitempty"`
	HeaderSignature string `json:"header_signature,omitempty"`
}

// Plugin represents a gateway plugin
type Plugin struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	RouteID   *string   `json:"route_id,omitempty" db:"route_id"`
	Config    PluginConfig `json:"config" db:"-"`
	Enabled   bool      `json:"enabled" db:"enabled"`
	Priority  int       `json:"priority" db:"priority"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// PluginConfig holds plugin configuration
type PluginConfig map[string]interface{}

// RateLimit represents rate limiting configuration
type RateLimit struct {
	ID               string    `json:"id" db:"id"`
	RouteID          *string   `json:"route_id,omitempty" db:"route_id"`
	RequestsPerSecond int      `json:"requests_per_second" db:"requests_per_second"`
	BurstSize        int       `json:"burst_size" db:"burst_size"`
	KeyType          string    `json:"key_type" db:"key_type"`
	KeyHeader        string    `json:"key_header,omitempty" db:"key_header"`
	Enabled          bool      `json:"enabled" db:"enabled"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// RequestContext holds per-request context data
type RequestContext struct {
	Route      *Route
	Upstream   *Upstream
	Target     *Target
	RequestID  string
	StartTime  time.Time
	AuthResult *AuthResult
	Plugins    map[string]PluginConfig
	Metadata   map[string]interface{}
}

// AuthResult holds authentication result
type AuthResult struct {
	Authenticated bool
	ProviderID    string
	Subject       string
	Claims        map[string]interface{}
	Metadata      map[string]string
}

// GatewayContext extends context.Context with request-specific data
type GatewayContext interface {
	context.Context
	RequestContext() *RequestContext
	SetRequestContext(*RequestContext)
	ResponseWriter() http.ResponseWriter
	Request() *http.Request
}
