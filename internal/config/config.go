package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

// Config is a structure with fields responsible for configuration parameters. The
// fields match the yaml configuration file
type Config struct {
	Env          string `yaml:"env" env-required:"true"`
	StoragePath  string `yaml:"storage_path" env-required:"true" env:"STORAGE_PATH"`
	SecretKey    string `yaml:"secret_key" env-required:"true" env:"SECRET_KEY"`
	GoogleKey    string `yaml:"google_key" env-required:"true" env:"GOOGLE_KEY"`
	GoogleSecret string `yaml:"google_secret" env-required:"true" env:"GOOGLE_SECRET"`
	HTTPServer   `yaml:"http_server"`
}

// HTTPServer is a structure with fields representing parameters of server such as Address and timeouts
type HTTPServer struct {
	Address     string        `yaml:"address" env-required:"true"`
	Timeout     time.Duration `yaml:"timeout" env-required:"true"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-required:"true"`
}

// MustLoad is a function that gets parameters from yaml config file and fills Config structure with it.
// Start with Must mean that function will panic if something goes wrong. So the function must be executed for the application to work.
// Returns config pointer
func MustLoad() *Config {

	//Getting path from env
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
