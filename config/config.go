package config

import (
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	File     FileConfig     `yaml:"file"`
	Logging  LoggingConfig  `yaml:"logging"`
	Database DatabaseConfig `yaml:"database"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
}

type FileConfig struct {
	MaxSize   int64  `yaml:"maxSize"`
	Path      string `yaml:"path"`
	Timeout   int    `yaml:"timeout"`
	Unit      string `yaml:"unit"`
	ChunkSize int    `yaml:"chunkSize"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Dbname   string `yaml:"dbname"`
}

func NewConfig(configPath string) (*Config, error) {
	config := &Config{}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)

	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}

func ValidateConfig(config *Config) bool {
	if config.Server.Port == "" {
		slog.Error("Server port is not set")
		return false
	}
	if config.File.MaxSize == 0 {
		slog.Error("File max size is not set")
		return false
	}
	if config.File.Path == "" {
		slog.Error("File path is not set")
		return false
	}
	if config.Logging.Level == "" {
		slog.Error("Logging level is not set")
		return false
	}
	return true
}
