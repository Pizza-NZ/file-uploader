package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	File     FileConfig     `yaml:"file"`
	Logging  LoggingConfig  `yaml:"logging"`
	Database DatabaseConfig `yaml:"database"`
	AWS      AWSConfig      `yaml:"aws"`
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

	config.AWS.AccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	config.AWS.SecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")

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
	if config.AWS.Region == "" {
		slog.Error("AWS region is not set")
		return false
	}
	if config.AWS.AccessKeyID == "" {
		slog.Error("AWS access key ID is not set")
		return false
	}
	if config.AWS.SecretAccessKey == "" {
		slog.Error("AWS secret access key is not set")
		return false
	}
	if config.AWS.S3.BucketName == "" {
		slog.Error("S3 bucket name is not set")
		return false
	}
	if config.AWS.S3.PresignedURLExpiry == 0 {
		slog.Error("S3 presigned URL expiry is not set")
		return false
	}
	return true
}
