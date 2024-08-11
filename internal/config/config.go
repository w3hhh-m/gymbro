package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Env           string `yaml:"env" env-required:"true"`
	StoragePath   string `yaml:"storage_path" env-required:"true" env:"STORAGE_PATH"`
	SessionsCfg   `yaml:"sessions_cfg"`
	JWTCfg        `yaml:"jwt_cfg"`
	RedisCfg      `yaml:"redis_cfg"`
	OAuthCfg      `yaml:"oauth_cfg"`
	HTTPServerCfg `yaml:"http_server_cfg"`
}

type SessionsCfg struct {
	SessionLifetime   time.Duration `yaml:"session_lifetime" env-required:"true"`
	SchedulerInterval time.Duration `yaml:"scheduler_interval" env-required:"true"`
}

type JWTCfg struct {
	JWTLifetime time.Duration `yaml:"jwt_lifetime" env-required:"true"`
	SecretKey   string        `yaml:"secret_key" env-required:"true" env:"SECRET_KEY"`
}

type OAuthCfg struct {
	GoogleKey    string `yaml:"google_key" env-required:"true" env:"GOOGLE_KEY"`
	GoogleSecret string `yaml:"google_secret" env-required:"true" env:"GOOGLE_SECRET"`
}

type RedisCfg struct {
	RedisPath     string `yaml:"redis_path" env-required:"true" env:"REDIS_PATH"`
	RedisPassword string `yaml:"redis_password" env-required:"true" env:"REDIS_PASSWORD"`
}

type HTTPServerCfg struct {
	Address     string        `yaml:"address" env-required:"true"`
	Timeout     time.Duration `yaml:"timeout" env-required:"true"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-required:"true"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH environment variable is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("%s: CONFIG_PATH does not exist", configPath)
	}

	var config Config
	if err := cleanenv.ReadConfig(configPath, &config); err != nil {
		log.Fatalf("%s: CONFIG_PATH read error: %v", configPath, err)
	}

	return &config
}
