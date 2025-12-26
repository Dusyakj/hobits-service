package config

import (
	"fmt"
	"os"

	"go.uber.org/config"
)

type Config struct {
	Service ServiceConfig `yaml:"service"`
	HTTP    HTTPConfig    `yaml:"http"`
	GRPC    GRPCConfig    `yaml:"grpc"`
	JWT     JWTConfig     `yaml:"jwt"`
	Logging LoggingConfig `yaml:"logging"`
	Metrics MetricsConfig `yaml:"metrics"`
}

type ServiceConfig struct {
	Name        string `yaml:"name"`
	Environment string `yaml:"environment"`
	Version     string `yaml:"version"`
}

type HTTPConfig struct {
	Port         int `yaml:"port"`
	ReadTimeout  int `yaml:"read_timeout"`
	WriteTimeout int `yaml:"write_timeout"`
}

type GRPCConfig struct {
	UserServiceAddr      string `yaml:"user_service_addr"`
	HabitsServiceAddr    string `yaml:"habits_service_addr"`
	BadHabitsServiceAddr string `yaml:"bad_habits_service_addr"`
	Timeout              int    `yaml:"timeout"`
}

type JWTConfig struct {
	Secret string `yaml:"secret"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
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

	cfg.overrideFromEnv()

	return &cfg, nil
}

// overrideFromEnv overrides config values with environment variables if present
func (c *Config) overrideFromEnv() {
	if val := os.Getenv("HTTP_PORT"); val != "" {
		fmt.Sscanf(val, "%d", &c.HTTP.Port)
	}
	if val := os.Getenv("USER_SERVICE_ADDR"); val != "" {
		c.GRPC.UserServiceAddr = val
	}
	if val := os.Getenv("HABITS_SERVICE_ADDR"); val != "" {
		c.GRPC.HabitsServiceAddr = val
	}
	if val := os.Getenv("BAD_HABITS_SERVICE_ADDR"); val != "" {
		c.GRPC.BadHabitsServiceAddr = val
	}
	if val := os.Getenv("JWT_SECRET"); val != "" {
		c.JWT.Secret = val
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
