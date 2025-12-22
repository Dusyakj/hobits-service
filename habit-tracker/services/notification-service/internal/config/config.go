package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"go.uber.org/config"
)

type Config struct {
	Service  ServiceConfig  `yaml:"service"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Kafka    KafkaConfig    `yaml:"kafka"`
	SMTP     SMTPConfig     `yaml:"smtp"`
	Email    EmailConfig    `yaml:"email"`
	Logging  LoggingConfig  `yaml:"logging"`
	Metrics  MetricsConfig  `yaml:"metrics"`
}

type ServiceConfig struct {
	Name        string `yaml:"name"`
	Environment string `yaml:"environment"`
	Version     string `yaml:"version"`
}

type DatabaseConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	Database        string        `yaml:"database"`
	SSLMode         string        `yaml:"ssl_mode"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

type RedisConfig struct {
	Addr         string        `yaml:"addr"`
	Password     string        `yaml:"password"`
	DB           int           `yaml:"db"`
	MaxRetries   int           `yaml:"max_retries"`
	PoolSize     int           `yaml:"pool_size"`
	MinIdleConns int           `yaml:"min_idle_conns"`
	DialTimeout  time.Duration `yaml:"dial_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

type KafkaConfig struct {
	Brokers       []string      `yaml:"brokers"`
	Topics        []string      `yaml:"topics"`
	GroupID       string        `yaml:"group_id"`
	ConsumerGroup string        `yaml:"consumer_group"`
	MaxRetries    int           `yaml:"max_retries"`
	RetryBackoff  time.Duration `yaml:"retry_backoff"`
}

type SMTPConfig struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	FromEmail string `yaml:"from_email"`
	FromName  string `yaml:"from_name"`
	UseTLS    bool   `yaml:"use_tls"`
}

type EmailConfig struct {
	VerificationURL string `yaml:"verification_url"`
	TemplatesPath   string `yaml:"templates_path"`
}

type LoggingConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	OutputPath string `yaml:"output_path"`
}

type MetricsConfig struct {
	Port int    `yaml:"port"`
	Path string `yaml:"path"`
}

// Load loads configuration from YAML file with environment variable overrides
func Load() (*Config, error) {
	configPath := getEnv("CONFIG_PATH", "./config/base.yaml")

	provider, err := config.NewYAML(
		config.File(configPath),
		config.Expand(os.LookupEnv),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create config provider: %w", err)
	}

	var cfg Config
	if err := provider.Get(config.Root).Populate(&cfg); err != nil {
		return nil, fmt.Errorf("failed to populate config: %w", err)
	}

	// Override with environment variables
	cfg.overrideFromEnv()

	return &cfg, nil
}

// overrideFromEnv overrides config values with environment variables if present
func (c *Config) overrideFromEnv() {
	if val := os.Getenv("SERVICE_NAME"); val != "" {
		c.Service.Name = val
	}
	if val := os.Getenv("SERVICE_ENVIRONMENT"); val != "" {
		c.Service.Environment = val
	}
	if val := os.Getenv("DATABASE_HOST"); val != "" {
		c.Database.Host = val
	}
	if val := os.Getenv("DATABASE_PORT"); val != "" {
		fmt.Sscanf(val, "%d", &c.Database.Port)
	}
	if val := os.Getenv("DATABASE_USER"); val != "" {
		c.Database.User = val
	}
	if val := os.Getenv("DATABASE_PASSWORD"); val != "" {
		c.Database.Password = val
	}
	if val := os.Getenv("DATABASE_NAME"); val != "" {
		c.Database.Database = val
	}
	if val := os.Getenv("DATABASE_SSL_MODE"); val != "" {
		c.Database.SSLMode = val
	}
	if val := os.Getenv("REDIS_ADDR"); val != "" {
		c.Redis.Addr = val
	}
	if val := os.Getenv("REDIS_PASSWORD"); val != "" {
		c.Redis.Password = val
	}
	if val := os.Getenv("REDIS_DB"); val != "" {
		fmt.Sscanf(val, "%d", &c.Redis.DB)
	}
	if val := os.Getenv("SMTP_HOST"); val != "" {
		c.SMTP.Host = val
	}
	if val := os.Getenv("SMTP_PORT"); val != "" {
		fmt.Sscanf(val, "%d", &c.SMTP.Port)
	}
	if val := os.Getenv("SMTP_USERNAME"); val != "" {
		c.SMTP.Username = val
	}
	if val := os.Getenv("SMTP_PASSWORD"); val != "" {
		c.SMTP.Password = val
	}
	if val := os.Getenv("SMTP_FROM_EMAIL"); val != "" {
		c.SMTP.FromEmail = val
	}
	if val := os.Getenv("SMTP_FROM_NAME"); val != "" {
		c.SMTP.FromName = val
	}
	if val := os.Getenv("SMTP_USE_TLS"); val != "" {
		if useTLS, err := strconv.ParseBool(val); err == nil {
			c.SMTP.UseTLS = useTLS
		}
	}
	if val := os.Getenv("EMAIL_VERIFICATION_URL"); val != "" {
		c.Email.VerificationURL = val
	}
	if val := os.Getenv("KAFKA_BROKER"); val != "" {
		c.Kafka.Brokers = []string{val}
	}
	if val := os.Getenv("LOG_LEVEL"); val != "" {
		c.Logging.Level = val
	}
}

// GetDSN returns PostgreSQL connection string in URL format for pgx/v5
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.SSLMode,
	)
}

// GetRedisAddr returns Redis address
func (c *RedisConfig) GetRedisAddr() string {
	return c.Addr
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
