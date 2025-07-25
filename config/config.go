package config

import (
	"errors"
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Environment string         `yaml:"environment"`
	StorageType string         `yaml:"storage_type"`
	Server      ServerConfig   `yaml:"server"`
	File        FileConfig     `yaml:"file"`
	Logging     LoggingConfig  `yaml:"logging"`
	Otel        OtelConfig     `yaml:"otel"`
	Database    DatabaseConfig `yaml:"database"`
	AWS         AWSConfig      `yaml:"aws"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
}

type FileConfig struct {
	MaxSize      int64    `yaml:"maxSize"`
	AllowedTypes []string `yaml:"allowedTypes"`
	Path         string   `yaml:"path"`
	Timeout      int      `yaml:"timeout"`
	Unit         string   `yaml:"unit"`
	ChunkSize    int      `yaml:"chunkSize"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
}

type OtelConfig struct {
	OtelServiceName string  `yaml:"service_name"`
	OtelEndpoint    string  `yaml:"endpoint"`
	OtelSampleRatio float64 `yaml:"otelSampleRatio"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Dbname   string `yaml:"dbname"`
}

type S3Config struct {
	BucketName         string `yaml:"bucket_name"`
	PresignedURLExpiry int    `yaml:"presigned_url_expiry"`
}

type AWSConfig struct {
	Region          string   `yaml:"region"`
	AccessKeyID     string   `yaml:"-"`
	SecretAccessKey string   `yaml:"-"`
	S3              S3Config `yaml:"s3"`
}

func NewConfig(configPath string) (*Config, error) {
	if err := godotenv.Load(); err != nil {
		slog.Warn("No .env file found")
	}

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

	// Allow environment variable override for the environment setting
	if env := os.Getenv("APP_ENV"); env != "" {
		config.Environment = env
	}

	config.AWS.AccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	config.AWS.SecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")

	return config, nil
}

// ValidateConfig checks if the configuration is valid.
// It now includes an environment check to only validate AWS keys in non-production environments.
func ValidateConfig(config *Config) error {
	if config.Server.Port == "" {
		return errors.New("server port is not set")
	}
	if config.File.MaxSize == 0 {
		return errors.New("File max size is not set")
	}
	if config.File.Path == "" {
		return errors.New("File path is not set")
	}
	if config.Logging.Level == "" {
		return errors.New("Logging level is not set")
	}
	if config.AWS.Region == "" {
		return errors.New("AWS region is not set")
	}

	// Only validate AWS keys if the environment is not "production".
	// This allows local development with .env files while using IAM roles in production.
	if strings.ToLower(config.Environment) != "production" {
		slog.Info("Non-production environment detected, validating AWS keys", "environment", config.Environment)
		if config.AWS.AccessKeyID == "" {
			return errors.New("AWS access key ID is not set for non-production environment")
		}
		if config.AWS.SecretAccessKey == "" {
			return errors.New("AWS secret access key is not set for non-production environment")
		}
	}

	if config.AWS.S3.BucketName == "" {
		return errors.New("S3 bucket name is not set")
	}
	if config.AWS.S3.PresignedURLExpiry == 0 {
		return errors.New("S3 presigned URL expiry is not set")
	}
	return nil
}
