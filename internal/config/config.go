/*
SPDX-License-Identifier: GPL-3.0-or-later

Copyright (C) 2025 Aaron Mathis aaron.mathis@gmail.com

This file is part of GoSight.

GoSight is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

GoSight is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with GoSight. If not, see https://www.gnu.org/licenses/.
*/

// server/internal/config/config.go

// Package config provides configuration loading and management for the GoSight server.
// It supports loading configuration from a YAML file and allows for environment variable overrides.
// The configuration includes settings for server address, storage engine, database path,
// logging, TLS settings, and debug options.
package config

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gopkg.in/yaml.v3"
)

// Config represents the comprehensive configuration structure for the GoSight server.
// This is the root configuration object that encompasses all server settings including:
//   - Server networking and protocol settings (gRPC, HTTP)
//   - Storage engine configurations for different data types
//   - Authentication and authorization providers
//   - TLS/SSL certificate settings
//   - Caching backend configurations
//   - Buffer engine settings for data batching
//   - Monitoring and telemetry configurations
//   - Debug and development options
//
// The configuration is typically loaded from a YAML file and can be overridden
// by environment variables for deployment flexibility.
//
// Example usage:
//
//	cfg, err := LoadConfig("config.yaml")
//	if err != nil {
//		log.Fatal(err)
//	}
//	ApplyEnvOverrides(cfg)
//
// The configuration supports multiple storage backends including:
//   - PostgreSQL for persistent storage
//   - SQLite for lightweight deployments
//   - Memory storage for testing
//   - Redis for high-performance caching
//   - File-based storage for simple deployments
type Config struct {
	Server struct {
		GRPCAddr     string        `yaml:"grpc_addr"`
		HTTPAddr     string        `yaml:"http_addr"`
		Environment  string        `yaml:"environment"`
		SyncInterval time.Duration `yaml:"sync_interval" default:"5m"`
	} `yaml:"server"`

	Logs struct {
		ErrorLogFile  string `yaml:"error_log_file"`
		AppLogFile    string `yaml:"app_log_file"`
		AccessLogFile string `yaml:"access_log_file"`
		DebugLogFile  string `yaml:"debug_log_file"`
		LogLevel      string `yaml:"log_level"`
	}
	Web struct {
		StaticDir     string   `yaml:"static_dir"`
		TemplateDir   string   `yaml:"template_dir"`
		DefaultTitle  string   `yaml:"default_title"`
		AuthProviders []string `yaml:"auth_providers"`
	} `yaml:"web"`

	TLS struct {
		HttpsCertFile string `yaml:"https_cert_file"`
		HttpsKeyFile  string `yaml:"https_key_file"`
		CertFile      string `yaml:"cert_file"`
		KeyFile       string `yaml:"key_file"`
		ClientCAFile  string `yaml:"client_ca_file"`
	} `yaml:"tls"`

	Debug struct {
		Enabled          bool `yaml:"enabled"`
		EnableReflection bool `yaml:"enable_reflection"`
	} `yaml:"debug"`

	MetricStore struct {
		Engine        string `yaml:"engine"`
		URL           string `yaml:"url"`
		Workers       int    `yaml:"workers"`
		QueueSize     int    `yaml:"queue_size"`
		BatchSize     int    `yaml:"batch_size"`
		BatchTimeout  int    `yaml:"batch_timeout"`
		BatchRetry    int    `yaml:"batch_retry"`
		BatchInterval int    `yaml:"batch_interval"`
	} `yaml:"metricstore"`

	LogStore struct {
		Engine        string `yaml:"engine"`          // file, victoriametric etc
		Table         string `yaml:"table,omitempty"` // optional table name for PostgreSQL
		Dir           string `yaml:"dir"`
		Url           string `yaml:"url,omitempty"` // optional URL for remote storage
		Workers       int    `yaml:"workers"`
		QueueSize     int    `yaml:"queue_size"`
		BatchSize     int    `yaml:"batch_size"`
		BatchTimeout  int    `yaml:"batch_timeout"`
		BatchRetry    int    `yaml:"batch_retry"`
		BatchInterval int    `yaml:"batch_interval"`
	}

	UserStore struct {
		Engine   string `yaml:"engine"`    // e.g. "postgres", "memory", "ldap"
		DSN      string `yaml:"dsn"`       // e.g. PostgreSQL connection string
		LDAPBase string `yaml:"ldap_base"` // optional: LDAP-specific config
	} `yaml:"userstore"`

	DataStore struct {
		Engine string `yaml:"engine"`         // "memory", "json", or "postgres"
		Path   string `yaml:"path,omitempty"` // optional path for JSON file
		DSN    string `yaml:"dsn,omitempty"`  // optional DSN for PostgreSQL
	} `yaml:"datastore"`

	ResourceStore struct {
		Engine string `yaml:"engine"`         // "memory", "json", or "postgres"
		Path   string `yaml:"path,omitempty"` // optional path for JSON file
		DSN    string `yaml:"dsn,omitempty"`  // optional DSN for PostgreSQL
	} `yaml:"resourcestore"`

	AlertStore struct {
		Engine string `yaml:"engine"`         // "memory", "json", or "postgres"
		Path   string `yaml:"path,omitempty"` // optional path for JSON file
		DSN    string `yaml:"dsn,omitempty"`  // optional DSN for PostgreSQL
	} `yaml:"alertstore"`

	EventStore struct {
		Engine string `yaml:"engine"`        // "memory", "json", or "postgres"
		Path   string `yaml:"path"`          // optional path for JSON file
		DSN    string `yaml:"dsn,omitempty"` // optional DSN for PostgreSQL
	} `yaml:"eventstore"`

	RuleStore struct {
		Engine string `yaml:"engine"` // "memory", "json", or "postgres"
		Path   string `yaml:"path"`   // optional path for JSON file
	} `yaml:"rulestore"`

	RouteStore struct {
		Path string `yaml:"path"` // path for YAML file
	} `yaml:"routestore"`

	BufferEngine BufferEngineConfig `yaml:"buffer_engine"`

	SyslogCollection SyslogCollectionConfig `yaml:"syslog_collection"`

	Auth struct {
		SSOEnabled bool         `yaml:"sso_enabled"`
		MFASecret  string       `yaml:"mfa_secret_key"`
		JWTSecret  string       `yaml:"jwt_secret"`
		Google     GoogleConfig `yaml:"google"`
		AWS        AWSConfig    `yaml:"aws"`
		Azure      AzureConfig  `yaml:"azure"`
		GitHub     GitHubConfig `yaml:"github"`
	} `yaml:"auth"`

	OpenTelemetry OpenTelemetryConfig `yaml:"opentelemetry"`

	Cache CacheConfig `yaml:"cache"`

	API APIConfig `yaml:"api"`
}

// OpenTelemetryConfig defines the configuration for OpenTelemetry observability.
// OpenTelemetry provides distributed tracing, metrics, and logging capabilities
// for monitoring and debugging the GoSight server in production environments.
//
// The configuration supports both HTTP and gRPC protocols for the OTLP (OpenTelemetry
// Protocol) receiver, allowing integration with various observability backends like
// Jaeger, Zipkin, Prometheus, and cloud-native solutions.
//
// Example YAML configuration:
//
//	opentelemetry:
//	  enabled: true
//	  http:
//	    enabled: true
//	    addr: ":4318"
//	  grpc:
//	    enabled: true
//	    addr: ":4317"
type OpenTelemetryConfig struct {
	Enabled bool               `yaml:"enabled"`
	HTTP    OTLPProtocolConfig `yaml:"http"`
	GRPC    OTLPProtocolConfig `yaml:"grpc"`
}

// OTLPProtocolConfig defines the configuration for an OTLP protocol endpoint.
// This configuration specifies whether a particular OTLP protocol (HTTP or gRPC)
// is enabled and the network address it should bind to for receiving telemetry data.
//
// The address format should be ":port" for binding to all interfaces on the specified
// port, or "host:port" for binding to a specific interface.
type OTLPProtocolConfig struct {
	Enabled bool   `yaml:"enabled"`
	Addr    string `yaml:"addr"`
}

// SyslogCollectionConfig defines the configuration for syslog message collection.
// The syslog collection subsystem enables GoSight to receive and process syslog
// messages from network devices, servers, and other infrastructure components.
//
// It supports both TCP and UDP protocols with configurable buffer sizes and
// connection limits to handle high-volume syslog ingestion while maintaining
// system stability and performance.
//
// Key features:
//   - Dual protocol support (TCP/UDP) for maximum compatibility
//   - Configurable buffer sizes for optimal memory usage
//   - Connection limiting to prevent resource exhaustion
//   - Per-IP rate limiting for DDoS protection
//
// Example configuration:
//
//	syslog_collection:
//	  tcp_enabled: true
//	  udp_enabled: true
//	  tcp_port: 514
//	  udp_port: 514
//	  max_connections: 1000
//	  default_ip_limit: 100
type SyslogCollectionConfig struct {
	TCPEnabled     bool `yaml:"tcp_enabled"`
	UDPEnabled     bool `yaml:"udp_enabled"`
	TCPPort        int  `yaml:"tcp_port"`
	UDPPort        int  `yaml:"udp_port"`
	TCPBufferSize  int  `yaml:"tcp_buffer_size"`
	UDPBufferSize  int  `yaml:"udp_buffer_size"`
	MaxConnections int  `yaml:"max_connections"`
	DefaultIPLimit int  `yaml:"default_ip_limit"`
}

// GoogleConfig represents the configuration for Google OAuth2 authentication.
// This configuration enables Single Sign-On (SSO) authentication through Google's
// OAuth2 service, allowing users to authenticate using their Google accounts.
//
// To set up Google OAuth2:
//  1. Create a project in Google Cloud Console
//  2. Enable the Google+ API
//  3. Create OAuth2 credentials
//  4. Configure authorized redirect URIs
//
// The redirect URI should match the callback endpoint in your GoSight server
// (typically: https://yourdomain.com/auth/google/callback).
//
// Required OAuth2 scopes: "openid", "email", "profile"
type GoogleConfig struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectURI  string `yaml:"redirect_uri"`
}

// AWSConfig represents the configuration for AWS Cognito authentication.
// AWS Cognito provides managed user authentication and authorization services
// with support for social identity providers, enterprise identity providers,
// and custom authentication flows.
//
// Configuration requires:
//   - AWS region where the Cognito User Pool is deployed
//   - User Pool ID from the Cognito service
//   - App Client credentials configured in the User Pool
//   - Redirect URI matching your GoSight callback endpoint
//
// The OAuth2 flow uses the standard authorization code grant with PKCE
// for enhanced security in web applications.
type AWSConfig struct {
	Region       string `yaml:"region"`
	UserPoolID   string `yaml:"user_pool_id"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectURI  string `yaml:"redirect_uri"`
}

// AzureConfig represents the configuration for Azure Active Directory authentication.
// Azure AD (now Microsoft Entra ID) provides enterprise-grade identity and
// access management with support for conditional access, multi-factor authentication,
// and integration with Microsoft 365 and other enterprise applications.
//
// Configuration requires:
//   - Tenant ID from your Azure AD tenant
//   - Application (client) ID from app registration
//   - Client secret for server-to-server authentication
//   - Redirect URI configured in the app registration
//
// Supports Microsoft Graph API access for user profile information
// and organizational data integration.
type AzureConfig struct {
	TenantID     string `yaml:"tenant_id"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectURI  string `yaml:"redirect_uri"`
}

// GitHubConfig represents the configuration for GitHub OAuth authentication.
// GitHub OAuth provides authentication using GitHub user accounts, making it
// ideal for development teams and organizations already using GitHub for
// source code management.
//
// Setup process:
//  1. Register a new OAuth App in GitHub Settings > Developer settings
//  2. Configure Authorization callback URL
//  3. Obtain Client ID and Client Secret
//  4. Set appropriate scopes for user data access
//
// Required scopes:
//   - "read:user" for basic user profile information
//   - "user:email" for accessing user email addresses
//
// GitHub OAuth is particularly useful for teams wanting to leverage
// existing GitHub organization membership for access control.
type GitHubConfig struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectURI  string `yaml:"redirect_uri"`
}

// CacheConfig represents the configuration for the caching subsystem.
// The cache system provides high-performance data caching to reduce database
// load and improve response times for frequently accessed data.
//
// Supported cache backends:
//   - "memory": In-memory LRU cache (default, fastest, single-node)
//   - "redis": Redis backend (distributed, persistent)
//   - "memcached": Memcached backend (distributed, volatile)
//
// The cache system includes:
//   - Automatic expiration and cleanup
//   - Configurable size limits
//   - Resource-specific flush intervals
//   - Backend-specific configuration options
//
// Performance considerations:
//   - Memory cache: Fastest access, limited to single node
//   - Redis: Network overhead, but supports clustering and persistence
//   - Memcached: Network overhead, optimized for high throughput
//
// Example configuration:
//
//	cache:
//	  enabled: true
//	  engine: "redis"
//	  expiration: "10m"
//	  size: 10000
//	  redis:
//	    addr: "localhost:6379"
//	    db: 0
type CacheConfig struct {
	Enabled               bool            `yaml:"enabled"`
	Engine                string          `yaml:"engine"`
	Redis                 RedisConfig     `yaml:"redis"`
	Memcached             MemcachedConfig `yaml:"memcached"`
	Expiration            time.Duration   `yaml:"expiration" default:"5m"`
	CleanupInterval       time.Duration   `yaml:"cleanup_interval" default:"1m"`
	Size                  int             `yaml:"size" default:"1000"`
	ResourceFlushInterval time.Duration   `yaml:"resource_flush_interval" env:"CACHE_RESOURCE_FLUSH_INTERVAL" envDefault:"30"`
}

// RedisConfig represents the configuration for Redis caching backend.
// Redis provides distributed caching with persistence, clustering, and
// advanced data structure support beyond simple key-value operations.
//
// Configuration options:
//   - Addr: Redis server address in "host:port" format
//   - Password: Authentication password (leave empty if not required)
//   - DB: Redis database number (0-15, default databases)
//
// Redis features utilized:
//   - Automatic key expiration
//   - Memory-efficient data structures
//   - Persistence options (RDB/AOF)
//   - High availability through clustering
type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// MemcachedConfig represents the configuration for Memcached caching backend.
// Memcached is a high-performance, distributed memory object caching system
// optimized for speed and simplicity, ideal for web applications with
// high read/write ratios.
//
// Configuration options:
//   - Addr: Memcached server address in "host:port" format
//   - Username: SASL authentication username (if SASL is enabled)
//   - Password: SASL authentication password (if SASL is enabled)
//
// Memcached characteristics:
//   - Volatile storage (data lost on restart)
//   - Simple key-value operations
//   - Excellent performance for read-heavy workloads
//   - Easy horizontal scaling
type MemcachedConfig struct {
	Addr     string `yaml:"addr"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// APIConfig represents the configuration for API versioning and management.
// This configuration enables sophisticated API version management including
// backward compatibility, deprecation handling, and client migration support.
//
// Key features:
//   - Multiple concurrent API versions
//   - Automatic version detection via headers or path prefixes
//   - Version-specific routing and middleware
//   - Deprecation warnings and sunset notifications
//   - Client migration assistance through redirects
//
// Version resolution order:
//  1. X-API-Version header (or custom header)
//  2. Path prefix (/v1/, /v2/, etc.)
//  3. Default version fallback
//
// Example configuration:
//
//	api:
//	  default_version: "v1"
//	  version_header: "X-API-Version"
//	  enable_version_redirect: true
//	  supported_versions:
//	    - version: "v1"
//	      enabled: true
//	      path_prefix: "/v1"
//	    - version: "v2"
//	      enabled: true
//	      deprecated: false
//	      path_prefix: "/v2"
type APIConfig struct {
	DefaultVersion        string             `yaml:"default_version" default:"v1"`
	SupportedVersions     []APIVersionConfig `yaml:"supported_versions"`
	EnableVersionRedirect bool               `yaml:"enable_version_redirect" default:"true"`
	VersionHeader         string             `yaml:"version_header" default:"X-API-Version"`
}

// APIVersionConfig represents configuration for a specific API version.
// Each API version can be independently enabled, deprecated, and configured
// with its own routing rules and metadata.
//
// Version lifecycle management:
//   - Enabled: Controls whether the version accepts new requests
//   - Deprecated: Indicates the version is deprecated (adds warning headers)
//   - Description: Human-readable version information for documentation
//   - PathPrefix: URL path prefix for version-specific routing
//
// Deprecation handling:
//   - Deprecated versions include Sunset and Deprecation headers
//   - Clients receive migration guidance through response headers
//   - Monitoring can track usage of deprecated versions
type APIVersionConfig struct {
	Version     string `yaml:"version"`
	Enabled     bool   `yaml:"enabled" default:"true"`
	Deprecated  bool   `yaml:"deprecated" default:"false"`
	Description string `yaml:"description"`
	PathPrefix  string `yaml:"path_prefix"`
}

// LoadConfig loads the configuration from a YAML file with comprehensive error handling.
// This function reads and parses a YAML configuration file, returning a fully
// populated Config struct. It performs basic validation during unmarshaling
// and provides detailed error information for troubleshooting.
//
// Parameters:
//   - path: Absolute or relative path to the YAML configuration file
//
// Returns:
//   - *Config: Populated configuration struct with all settings
//   - error: Detailed error information including file access and parsing errors
//
// Error handling:
//   - File not found: Returns clear indication of missing file
//   - Permission denied: Indicates file access permission issues
//   - YAML syntax errors: Provides line number and syntax details
//   - Type validation: Reports field type mismatches
//
// Example usage:
//
//	cfg, err := LoadConfig("/etc/gosight/server.yaml")
//	if err != nil {
//		log.Fatalf("Failed to load config: %v", err)
//	}
//
// The loaded configuration should be validated and have environment
// overrides applied using ApplyEnvOverrides() before use.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// ApplyEnvOverrides applies environment variable overrides to the configuration.
// This function allows runtime configuration modification without editing YAML files,
// making it ideal for containerized deployments, CI/CD pipelines, and different
// deployment environments (development, staging, production).
//
// Supported environment variables:
//   - GOSIGHT_GRPC_LISTEN: Override gRPC server listen address
//   - GOSIGHT_HTTP_LISTEN: Override HTTP server listen address
//   - GOSIGHT_ERROR_LOG_FILE: Override error log file path
//   - GOSIGHT_APP_LOG_FILE: Override application log file path
//   - GOSIGHT_ACCESS_LOG_FILE: Override access log file path
//   - GOSIGHT_LOG_LEVEL: Override logging level (debug, info, warn, error)
//   - GOSIGHT_TLS_CERT_FILE: Override TLS certificate file path
//   - GOSIGHT_TLS_KEY_FILE: Override TLS private key file path
//   - GOSIGHT_TLS_CLIENT_CA_FILE: Override client CA certificate file path
//   - GOSIGHT_DEBUG_ENABLE_REFLECTION: Enable/disable gRPC reflection (true/false)
//   - GOSIGHT_USERSTORE_TYPE: Override user store backend type
//   - GOSIGHT_USERSTORE_DSN: Override user store connection string
//   - GOSIGHT_USERSTORE_LDAP_BASE: Override LDAP base DN
//
// Environment variable precedence:
//  1. Environment variables (highest priority)
//  2. YAML configuration file
//  3. Default values (lowest priority)
//
// Example usage:
//
//	cfg, _ := LoadConfig("config.yaml")
//	ApplyEnvOverrides(cfg)  // Apply environment overrides
//
// This function modifies the passed configuration struct in-place
// and should be called after LoadConfig() but before using the configuration.
func ApplyEnvOverrides(cfg *Config) {
	if val := os.Getenv("GOSIGHT_GRPC_LISTEN"); val != "" {
		cfg.Server.GRPCAddr = val
	}
	if val := os.Getenv("GOSIGHT_HTTP_LISTEN"); val != "" {
		cfg.Server.HTTPAddr = val
	}
	if val := os.Getenv("GOSIGHT_ERROR_LOG_FILE"); val != "" {
		cfg.Logs.ErrorLogFile = val
	}
	if val := os.Getenv("GOSIGHT_APP_LOG_FILE"); val != "" {
		cfg.Logs.ErrorLogFile = val
	}
	if val := os.Getenv("GOSIGHT_ACCESS_LOG_FILE"); val != "" {
		cfg.Logs.AccessLogFile = val
	}
	if val := os.Getenv("GOSIGHT_LOG_LEVEL"); val != "" {
		cfg.Logs.LogLevel = val
	}
	if val := os.Getenv("GOSIGHT_TLS_CERT_FILE"); val != "" {
		cfg.TLS.CertFile = val
	}
	if val := os.Getenv("GOSIGHT_TLS_KEY_FILE"); val != "" {
		cfg.TLS.KeyFile = val
	}
	if val := os.Getenv("GOSIGHT_TLS_CLIENT_CA_FILE"); val != "" {
		cfg.TLS.ClientCAFile = val
	}
	if val := os.Getenv("GOSIGHT_DEBUG_ENABLE_REFLECTION"); val != "" {
		cfg.Debug.EnableReflection = val == "true"
	}
	if val := os.Getenv("GOSIGHT_USERSTORE_TYPE"); val != "" {
		cfg.UserStore.Engine = val
	}
	if val := os.Getenv("GOSIGHT_USERSTORE_DSN"); val != "" {
		cfg.UserStore.DSN = val
	}
	if val := os.Getenv("GOSIGHT_USERSTORE_LDAP_BASE"); val != "" {
		cfg.UserStore.LDAPBase = val
	}
}

// ToOAuthConfig converts the GoogleConfig to an OAuth2 configuration.
// This method transforms the Google-specific configuration into a standard
// OAuth2 configuration that can be used with the golang.org/x/oauth2 library
// for implementing the OAuth2 authorization code flow.
//
// The conversion includes:
//   - Setting up Google's OAuth2 endpoints
//   - Configuring required scopes for user information
//   - Preparing redirect URI for callback handling
//
// Required OAuth2 scopes:
//   - "openid": Enables OpenID Connect authentication
//   - "email": Access to user's email address
//   - "profile": Access to user's basic profile information
//
// Returns:
//   - *oauth2.Config: Ready-to-use OAuth2 configuration for Google authentication
//
// Example usage:
//
//	googleCfg := &GoogleConfig{
//		ClientID: "your-client-id",
//		ClientSecret: "your-client-secret",
//		RedirectURI: "https://yourapp.com/auth/google/callback",
//	}
//	oauth2Cfg := googleCfg.ToOAuthConfig()
//	authURL := oauth2Cfg.AuthCodeURL("state")
func (g *GoogleConfig) ToOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     g.ClientID,
		ClientSecret: g.ClientSecret,
		RedirectURL:  g.RedirectURI,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

// ToOAuthConfig converts the AWSConfig to an OAuth2 configuration.
// This method creates a standard OAuth2 configuration for AWS Cognito
// authentication, automatically constructing the appropriate endpoints
// based on the configured region and user pool ID.
//
// AWS Cognito OAuth2 endpoints are constructed as:
//   - AuthURL: https://{user_pool_id}.auth.{region}.amazoncognito.com/oauth2/authorize
//   - TokenURL: https://{user_pool_id}.auth.{region}.amazoncognito.com/oauth2/token
//
// The configuration includes standard OpenID Connect scopes for
// accessing user profile information and authentication status.
//
// Returns:
//   - *oauth2.Config: OAuth2 configuration for AWS Cognito authentication
//
// Note: Requires valid AWS Cognito User Pool configuration with
// OAuth2 enabled and appropriate app client settings.
func (a *AWSConfig) ToOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     a.ClientID,
		ClientSecret: a.ClientSecret,
		RedirectURL:  a.RedirectURI,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("https://%s.auth.%s.amazoncognito.com/oauth2/authorize", a.UserPoolID, a.Region),
			TokenURL: fmt.Sprintf("https://%s.auth.%s.amazoncognito.com/oauth2/token", a.UserPoolID, a.Region),
		},
	}
}

// ToOAuthConfig converts the AzureConfig to an OAuth2 configuration.
// This method creates a standard OAuth2 configuration for Azure Active Directory
// (Microsoft Entra ID) authentication using the Microsoft identity platform v2.0 endpoints.
//
// Azure AD OAuth2 endpoints are constructed as:
//   - AuthURL: https://login.microsoftonline.com/{tenant_id}/oauth2/v2.0/authorize
//   - TokenURL: https://login.microsoftonline.com/{tenant_id}/oauth2/v2.0/token
//
// Configured scopes include:
//   - "openid": OpenID Connect authentication
//   - "email": User email address access
//   - "profile": Basic user profile information
//   - "offline_access": Refresh token support for long-lived sessions
//   - "https://graph.microsoft.com/User.Read": Microsoft Graph API access for user data
//
// Returns:
//   - *oauth2.Config: OAuth2 configuration for Azure AD authentication
//
// The configuration supports both personal Microsoft accounts and
// organizational accounts depending on the tenant configuration.
func (a *AzureConfig) ToOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     a.ClientID,
		ClientSecret: a.ClientSecret,
		RedirectURL:  a.RedirectURI,
		Scopes: []string{
			"openid",
			"email",
			"profile",
			"offline_access",
			"https://graph.microsoft.com/User.Read",
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/authorize", a.TenantID),
			TokenURL: fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", a.TenantID),
		},
	}
}

// ToOAuthConfig converts the GitHubConfig to an OAuth2 configuration.
// This method creates a standard OAuth2 configuration for GitHub OAuth
// authentication, enabling users to authenticate using their GitHub accounts.
//
// GitHub OAuth2 endpoints:
//   - AuthURL: https://github.com/login/oauth/authorize
//   - TokenURL: https://github.com/login/oauth/access_token
//
// Configured scopes:
//   - "read:user": Access to user's public profile information
//   - "user:email": Access to user's email addresses (including private emails)
//
// Returns:
//   - *oauth2.Config: OAuth2 configuration for GitHub authentication
//
// The configuration enables access to user profile data and email information
// necessary for user identification and account linking. Additional scopes
// can be added based on application requirements (e.g., repository access).
//
// GitHub OAuth is particularly useful for developer tools and applications
// that want to integrate with existing GitHub workflows and permissions.
func (g *GitHubConfig) ToOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     g.ClientID,
		ClientSecret: g.ClientSecret,
		RedirectURL:  g.RedirectURI,
		Scopes: []string{
			"read:user",
			"user:email",
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		},
	}
}
