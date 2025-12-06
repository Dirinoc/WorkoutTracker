package config

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

// TODO: AppSecret (auth)
type Config struct {
	Env        string `yaml:"env" env-default:"local"`
	HTTPServer `yaml:"http_server"`
	DB         `yaml:"db"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8085"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type Client struct {
	Address      string        `yaml:"address"`
	Timeout      time.Duration `yaml:"timeout"`
	RetriesCount int           `yaml:"retries_count"`
	Insecure     bool          `yaml:"insecure"`
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
