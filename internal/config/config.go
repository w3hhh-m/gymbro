package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

// Config holds the configuration parameters from the YAML file.
type Config struct {
	Env               string        `yaml:"env" env-required:"true"`
	JWTLifetime       time.Duration `yaml:"jwt_lifetime" env-required:"true"`
	SessionLifetime   time.Duration `yaml:"session_lifetime" env-required:"true"`
	SchedulerInterval time.Duration `yaml:"scheduler_interval" env-required:"true"`
	StoragePath       string        `yaml:"storage_path" env-required:"true" env:"STORAGE_PATH"`
	SecretKey         string        `yaml:"secret_key" env-required:"true" env:"SECRET_KEY"`
	GoogleKey         string        `yaml:"google_key" env-required:"true" env:"GOOGLE_KEY"`
	GoogleSecret      string        `yaml:"google_secret" env-required:"true" env:"GOOGLE_SECRET"`
	HTTPServer        `yaml:"http_server"`
}

// HTTPServer holds the server parameters such as address and timeouts.
type HTTPServer struct {
	Address     string        `yaml:"address" env-required:"true"`
	Timeout     time.Duration `yaml:"timeout" env-required:"true"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-required:"true"`
}

// MustLoad loads configuration parameters from a YAML file and returns a Config pointer.
// The function panics if any error occurs during the loading process.
func MustLoad() *Config {
	// Get configuration path from environment variable
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH environment variable is not set")
	}

	// Check if the configuration path exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("%s: CONFIG_PATH does not exist", configPath)
	}

	// Initialize Config structure
	var config Config

	// Read configuration from YAML file
	if err := cleanenv.ReadConfig(configPath, &config); err != nil {
		log.Fatalf("%s: CONFIG_PATH read error: %v", configPath, err)
	}

	return &config
}
