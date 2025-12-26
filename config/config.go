package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	Redis    RedisConfig    `yaml:"redis"`
	Crypto   CryptoConfig   `yaml:"crypto"`
	Logger   LoggerConfig   `yaml:"logger"`
}

type ServerConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type DatabaseConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	Database        string `yaml:"database"`
	SSLMode         string `yaml:"ssl_mode"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"` // in minutes
}

type JWTConfig struct {
	SecretKey       string        `yaml:"secret_key"`
	TokenDuration   time.Duration `yaml:"token_duration"`
	RefreshDuration time.Duration `yaml:"refresh_duration"`
}

type RedisConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type CryptoConfig struct {
	EncryptionKey string `yaml:"encryption_key"`
}

type LoggerConfig struct {
	Environment string `yaml:"environment"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(file, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Override dengan environment variables jika ada
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		config.Database.Host = dbHost
	}
	if dbUser := os.Getenv("DB_USER"); dbUser != "" {
		config.Database.User = dbUser
	}
	if dbPass := os.Getenv("DB_PASSWORD"); dbPass != "" {
		config.Database.Password = dbPass
	}
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		config.JWT.SecretKey = jwtSecret
	}
	if encKey := os.Getenv("ENCRYPTION_KEY"); encKey != "" {
		config.Crypto.EncryptionKey = encKey
	}
	if redisHost := os.Getenv("REDIS_HOST"); redisHost != "" {
		config.Redis.Host = redisHost
	}
	if redisPort := os.Getenv("REDIS_PORT"); redisPort != "" {
		if port, err := strconv.Atoi(redisPort); err == nil {
			config.Redis.Port = port
		}
	}

	// ===== CRITICAL DEBUG LOGGING =====
	fmt.Println("========================================")
	fmt.Println("📋 CONFIG LOADED SUCCESSFULLY")
	fmt.Println("========================================")
	fmt.Printf("JWT Token Duration (Raw): %v\n", config.JWT.TokenDuration)
	fmt.Printf("JWT Token Duration (Hours): %.2f hours\n", config.JWT.TokenDuration.Hours())
	fmt.Printf("JWT Token Duration (Days): %.2f days\n", config.JWT.TokenDuration.Hours()/24)
	fmt.Printf("JWT Refresh Duration (Raw): %v\n", config.JWT.RefreshDuration)
	fmt.Printf("JWT Refresh Duration (Hours): %.2f hours\n", config.JWT.RefreshDuration.Hours())
	fmt.Printf("JWT Refresh Duration (Days): %.2f days\n", config.JWT.RefreshDuration.Hours()/24)
	fmt.Println("========================================")

	// Validasi durasi minimal
	if config.JWT.TokenDuration < time.Hour {
		fmt.Printf("⚠️  WARNING: Token duration sangat pendek: %v\n", config.JWT.TokenDuration)
	}

	return &config, nil
}

// Helper function untuk mendapatkan connection string database
func (c *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode)
}

// Helper function untuk mendapatkan Redis address
func (c *RedisConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
