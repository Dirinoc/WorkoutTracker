package config

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env      string        `yaml:"env" env-default:"local"`
	TokenTTL time.Duration `yaml:"token_ttl" env-required:"true"`
	GRPC     GRPCConfig    `yaml:"grpc"`
	DB       `yaml:"db"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type DB struct {
	User     string `env:"DB_USER"     yaml:"user"`
	Password string `env:"DB_PASSWORD" yaml:"password"`
	Name     string `env:"DB_NAME"     yaml:"name"`
	Host     string `env:"DB_HOST"     yaml:"host"`
	Port     string `env:"DB_PORT"     yaml:"port"`
	SSLMode  string `env:"DB_SSLMODE"  yaml:"sslmode"`
}

func MustLoad() (*Config, error) {
	_ = godotenv.Load()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return nil, err
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) ConnStr() string {
	return fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%s sslmode=%s",
		c.DB.User, c.DB.Password, c.DB.Name, c.DB.Host, c.DB.Port, c.DB.SSLMode,
	)
}
