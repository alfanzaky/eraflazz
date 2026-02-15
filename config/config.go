package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	App       AppConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	JWT       JWTConfig
	Auth      AuthConfig
	SMTP      SMTPConfig
	API       APIConfig
	Suppliers SupplierConfig
	H2H       H2HConfig
}

// AppConfig holds application configuration
type AppConfig struct {
	Name        string
	Environment string
	Port        string
	Debug       bool
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSLMode  string
	MaxIdle  int
	MaxOpen  int
	MaxLife  time.Duration
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	PoolSize int
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret         string
	ExpirationTime time.Duration
	RefreshTime    time.Duration
}

// AuthConfig holds authentication related configuration
type AuthConfig struct {
	AccessSecret    string
	RefreshSecret   string
	Issuer          string
	Audience        string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	H2HAPIKey       string
	H2HAPISecret    string
	H2HAllowedIPs   []string
}

// SMTPConfig holds SMTP configuration
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// APIConfig holds API configuration
type APIConfig struct {
	RateLimitPerMinute int
	TimeoutSeconds     int
	MaxRequestSize     int64
}

// SupplierConfig holds external supplier configurations
type SupplierConfig struct {
	Digiflazz DigiflazzConfig
}

// DigiflazzConfig holds Digiflazz supplier specific configuration
type DigiflazzConfig struct {
	BaseURL        string
	Username       string
	APIKey         string
	Testing        bool
	TimeoutSeconds int
}

// H2HConfig holds H2H API configuration
type H2HConfig struct {
	APIKey     string
	APISecret  string
	AllowedIPs []string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// .env file not found, continue with environment variables
		fmt.Println("No .env file found, using environment variables")
	}

	config := &Config{
		App: AppConfig{
			Name:        getEnv("APP_NAME", "Eraflazz"),
			Environment: getEnv("APP_ENV", "development"),
			Port:        getEnv("APP_PORT", "8080"),
			Debug:       getEnvBool("APP_DEBUG", true),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			Name:     getEnv("DB_NAME", "eraflazz_db"),
			User:     getEnv("DB_USER", "eraflazz_user"),
			Password: getEnv("DB_PASSWORD", "eraflazz_password"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			MaxIdle:  getEnvInt("DB_MAX_IDLE", 10),
			MaxOpen:  getEnvInt("DB_MAX_OPEN", 100),
			MaxLife:  getEnvDuration("DB_MAX_LIFE", time.Hour),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
			PoolSize: getEnvInt("REDIS_POOL_SIZE", 10),
		},
		JWT: JWTConfig{
			Secret:         getEnv("JWT_SECRET", "your-secret-key"),
			ExpirationTime: getEnvDuration("JWT_EXPIRATION", 24*time.Hour),
			RefreshTime:    getEnvDuration("JWT_REFRESH", 7*24*time.Hour),
		},
		Auth: AuthConfig{
			AccessSecret:    getEnv("AUTH_ACCESS_SECRET", getEnv("JWT_SECRET", "your-secret-key")),
			RefreshSecret:   getEnv("AUTH_REFRESH_SECRET", getEnv("JWT_SECRET", "your-secret-key")),
			Issuer:          getEnv("AUTH_ISSUER", "eraflazz"),
			Audience:        getEnv("AUTH_AUDIENCE", "eraflazz-clients"),
			AccessTokenTTL:  getEnvDuration("AUTH_ACCESS_TTL", 24*time.Hour),
			RefreshTokenTTL: getEnvDuration("AUTH_REFRESH_TTL", 7*24*time.Hour),
			H2HAPIKey:       getEnv("H2H_API_KEY", ""),
			H2HAPISecret:    getEnv("H2H_API_SECRET", ""),
			H2HAllowedIPs:   getEnvSlice("H2H_ALLOWED_IPS", []string{}),
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", "smtp.gmail.com"),
			Port:     getEnvInt("SMTP_PORT", 587),
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", "noreply@eraflazz.com"),
		},
		API: APIConfig{
			RateLimitPerMinute: getEnvInt("API_RATE_LIMIT", 100),
			TimeoutSeconds:     getEnvInt("API_TIMEOUT", 30),
			MaxRequestSize:     getEnvInt64("API_MAX_REQUEST_SIZE", 1048576), // 1MB
		},
		Suppliers: SupplierConfig{
			Digiflazz: DigiflazzConfig{
				BaseURL:        getEnv("DIGIFLAZZ_BASE_URL", "https://api.digiflazz.com/v1"),
				Username:       getEnv("DIGIFLAZZ_USERNAME", ""),
				APIKey:         getEnv("DIGIFLAZZ_API_KEY", ""),
				Testing:        getEnvBool("DIGIFLAZZ_TESTING", true),
				TimeoutSeconds: getEnvInt("DIGIFLAZZ_TIMEOUT", 30),
			},
		},
		H2H: H2HConfig{
			APIKey:     getEnv("H2H_API_KEY", ""),
			APISecret:  getEnv("H2H_API_SECRET", ""),
			AllowedIPs: getEnvSlice("H2H_ALLOWED_IPS", []string{}),
		},
	}

	return config, nil
}

// GetDSN returns database connection string
func (d *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode)
}

// GetRedisAddr returns Redis connection address
func (r *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

// IsDevelopment returns true if environment is development
func (a *AppConfig) IsDevelopment() bool {
	return a.Environment == "development"
}

// IsProduction returns true if environment is production
func (a *AppConfig) IsProduction() bool {
	return a.Environment == "production"
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		parts := strings.Split(value, ",")
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultValue
}

// Validate validates configuration
func (c *Config) Validate() error {
	// Validate required fields
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if c.JWT.Secret == "" || c.JWT.Secret == "your-secret-key" {
		return fmt.Errorf("JWT secret must be set and not use default value")
	}

	return nil
}

// Print prints configuration (excluding sensitive data)
func (c *Config) Print() {
	fmt.Printf("=== Configuration ===\n")
	fmt.Printf("App Name: %s\n", c.App.Name)
	fmt.Printf("Environment: %s\n", c.App.Environment)
	fmt.Printf("Port: %s\n", c.App.Port)
	fmt.Printf("Debug: %v\n", c.App.Debug)
	fmt.Printf("Database: %s:%s/%s\n", c.Database.Host, c.Database.Port, c.Database.Name)
	fmt.Printf("Redis: %s:%s/%d\n", c.Redis.Host, c.Redis.Port, c.Redis.DB)
	fmt.Printf("JWT Expiration: %v\n", c.JWT.ExpirationTime)
	fmt.Printf("====================\n")
}
