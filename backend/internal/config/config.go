package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Environment represents the application environment.
type Environment string

const (
	EnvLocal      Environment = "local"
	EnvDev        Environment = "development"
	EnvStaging    Environment = "staging"
	EnvProduction Environment = "production"
)

// IsProduction returns true if the environment is production.
func (e Environment) IsProduction() bool {
	return e == EnvProduction
}

// IsDevelopment returns true if the environment is local or development.
func (e Environment) IsDevelopment() bool {
	return e == EnvLocal || e == EnvDev
}

// Config holds all application configuration.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Storage  StorageConfig
	Services ExternalServicesConfig
}

type ServerConfig struct {
	Port           int
	Environment    Environment
	AllowedOrigins []string
}

type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	URL string
}

type JWTConfig struct {
	AccessTokenSecret  string
	RefreshTokenSecret string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

type StorageConfig struct {
	S3Bucket   string
	S3Region   string
	S3Endpoint string
	AWSKey     string
	AWSSecret  string
}

type ExternalServicesConfig struct {
	BkashAppKey     string
	BkashAppSecret  string
	BkashBaseURL    string
	AamarPayStoreID string
	AamarPayAPIKey  string
	AamarPayBaseURL string
	FirebaseProject string
	FirebaseKey     string
	FirebaseEmail   string
	SMSAPIKey       string
	SMSBaseURL      string
	BarikoiAPIKey   string
	SentryDSN       string
}

// Load reads configuration from environment variables and .env file.
func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigFile(".env")
	v.SetConfigType("env")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Defaults
	v.SetDefault("PORT", 8080)
	v.SetDefault("ENVIRONMENT", "local")
	v.SetDefault("DB_MAX_OPEN_CONNS", 25)
	v.SetDefault("DB_MAX_IDLE_CONNS", 5)
	v.SetDefault("DB_CONN_MAX_LIFETIME", "5m")
	v.SetDefault("JWT_ACCESS_EXPIRY", "15m")
	v.SetDefault("JWT_REFRESH_EXPIRY", "168h")

	// Read .env file (ignore if not found)
	_ = v.ReadInConfig()

	connMaxLifetime, err := time.ParseDuration(v.GetString("DB_CONN_MAX_LIFETIME"))
	if err != nil {
		connMaxLifetime = 5 * time.Minute
	}

	accessExpiry, err := time.ParseDuration(v.GetString("JWT_ACCESS_EXPIRY"))
	if err != nil {
		accessExpiry = 15 * time.Minute
	}

	refreshExpiry, err := time.ParseDuration(v.GetString("JWT_REFRESH_EXPIRY"))
	if err != nil {
		refreshExpiry = 7 * 24 * time.Hour
	}

	env := parseEnvironment(v.GetString("ENVIRONMENT"))
	origins := parseOrigins(v.GetString("ALLOWED_ORIGINS"))

	cfg := &Config{
		Server: ServerConfig{
			Port:           v.GetInt("PORT"),
			Environment:    env,
			AllowedOrigins: origins,
		},
		Database: DatabaseConfig{
			URL:             v.GetString("DATABASE_URL"),
			MaxOpenConns:    v.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns:    v.GetInt("DB_MAX_IDLE_CONNS"),
			ConnMaxLifetime: connMaxLifetime,
		},
		Redis: RedisConfig{
			URL: v.GetString("REDIS_URL"),
		},
		JWT: JWTConfig{
			AccessTokenSecret:  v.GetString("JWT_ACCESS_SECRET"),
			RefreshTokenSecret: v.GetString("JWT_REFRESH_SECRET"),
			AccessTokenExpiry:  accessExpiry,
			RefreshTokenExpiry: refreshExpiry,
		},
		Storage: StorageConfig{
			S3Bucket:   v.GetString("S3_BUCKET"),
			S3Region:   v.GetString("S3_REGION"),
			S3Endpoint: v.GetString("S3_ENDPOINT"),
			AWSKey:     v.GetString("AWS_ACCESS_KEY_ID"),
			AWSSecret:  v.GetString("AWS_SECRET_ACCESS_KEY"),
		},
		Services: ExternalServicesConfig{
			BkashAppKey:     v.GetString("BKASH_APP_KEY"),
			BkashAppSecret:  v.GetString("BKASH_APP_SECRET"),
			BkashBaseURL:    v.GetString("BKASH_BASE_URL"),
			AamarPayStoreID: v.GetString("AAMARPAY_STORE_ID"),
			AamarPayAPIKey:  v.GetString("AAMARPAY_API_KEY"),
			AamarPayBaseURL: v.GetString("AAMARPAY_BASE_URL"),
			FirebaseProject: v.GetString("FIREBASE_PROJECT_ID"),
			FirebaseKey:     v.GetString("FIREBASE_PRIVATE_KEY"),
			FirebaseEmail:   v.GetString("FIREBASE_CLIENT_EMAIL"),
			SMSAPIKey:       v.GetString("SMS_API_KEY"),
			SMSBaseURL:      v.GetString("SMS_BASE_URL"),
			BarikoiAPIKey:   v.GetString("BARIKOI_API_KEY"),
			SentryDSN:       v.GetString("SENTRY_DSN"),
		},
	}

	return cfg, nil
}

func parseEnvironment(env string) Environment {
	switch strings.ToLower(strings.TrimSpace(env)) {
	case "development", "dev":
		return EnvDev
	case "staging":
		return EnvStaging
	case "production", "prod":
		return EnvProduction
	default:
		return EnvLocal
	}
}

func parseOrigins(raw string) []string {
	if raw == "" {
		return []string{"http://localhost:3000"}
	}
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}
