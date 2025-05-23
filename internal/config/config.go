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

// Config represents the configuration structure for the GoSight server.
// It includes settings for server address, storage engine, database path,
// logging, TLS settings, and debug options.
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
		EnableReflection bool `yaml:"enable_reflection"`
	} `yaml:"debug"`

	// TODO - split up storage engine configs.
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
		Engine        string `yaml:"engine"` // file, victoriametric etc
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
}

// SyslogCollectionConfig defines the configuration for syslog collection
// The syslog collection can be used to collect syslog messages from network devices.
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
// It includes the client ID, client secret, and redirect URI.
type GoogleConfig struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectURI  string `yaml:"redirect_uri"`
}

// AWSConfig represents the configuration for AWS Cognito authentication
type AWSConfig struct {
	Region       string `yaml:"region"`
	UserPoolID   string `yaml:"user_pool_id"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectURI  string `yaml:"redirect_uri"`
}

// AzureConfig represents the configuration for Azure AD authentication
type AzureConfig struct {
	TenantID     string `yaml:"tenant_id"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectURI  string `yaml:"redirect_uri"`
}

// GitHubConfig represents the configuration for GitHub OAuth authentication
type GitHubConfig struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectURI  string `yaml:"redirect_uri"`
}

// LoadConfig loads the configuration from a YAML file.
// It returns a Config struct and an error if any occurs during loading.
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
// It checks for specific environment variables and updates the corresponding
// fields in the Config struct. This allows for dynamic configuration without
// modifying the YAML file.
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
// It sets the client ID, client secret, redirect URI, scopes, and endpoint.
func (g *GoogleConfig) ToOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     g.ClientID,
		ClientSecret: g.ClientSecret,
		RedirectURL:  g.RedirectURI,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

// ToOAuthConfig converts the AWSConfig to an OAuth2 configuration
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

// ToOAuthConfig converts the AzureConfig to an OAuth2 configuration
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

// ToOAuthConfig converts the GitHubConfig to an OAuth2 configuration
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
