package core

import (
	"time"

	"github.com/meraf00/swytch/core/lib/env"
	"github.com/meraf00/swytch/core/lib/logger"
)

type AppConfig struct {
	AppName     string
	Environment string
	StartedAt   time.Time
	Swagger     SwaggerConfig
	HTTP        ServerConfig
	Database    DatabaseConfig
	Encryption  EncryptionConfig
	Redis       RedisConfig
	Storage     StorageConfig
	RabbitMQ    RabbitMQConfig
}

type SwaggerConfig struct {
	Title       string
	Version     string
	BasePath    string
	DocEndpoint string
}

type EncryptionConfig struct {
	JWTKey   string
	HashSalt string
	HashCost int
}

type ServerConfig struct {
	Host string
	Port int
	Cors bool
}

type DatabaseConfig struct {
	Database string
	Host     string
	Username string
	Password string
	Port     int
	SSLMode  string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type RabbitMQConfig struct {
	Addr string
}

type StorageConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
	PresignedUrlTTL time.Duration
}

func LoadConfig(logger logger.Log) *AppConfig {
	env.LoadEnv(logger)

	return &AppConfig{
		AppName:     "swytch",
		Environment: env.GetEnvironment("development"),
		StartedAt:   time.Now(),
		Swagger: SwaggerConfig{
			Title:       "Swytch",
			Version:     "1.0.0",
			BasePath:    "/api",
			DocEndpoint: "/docs",
		},
		HTTP: ServerConfig{
			Host: env.GetEnvString("HTTP_HOST", "0.0.0.0", false),
			Port: env.GetEnvNumber("HTTP_PORT", 9090, false),
			Cors: true,
		},
		Database: DatabaseConfig{
			Database: env.GetEnvString("DB_NAME", "", true),
			Host:     env.GetEnvString("DB_HOST", "127.0.0.1", false),
			Username: env.GetEnvString("DB_USER", "", true),
			Password: env.GetEnvString("DB_PASS", "", true),
			Port:     env.GetEnvNumber("DB_PORT", 5432, false),
			SSLMode:  env.GetEnvString("DB_SSL_MODE", "disable", false),
		},
		Encryption: EncryptionConfig{
			JWTKey:   env.GetEnvString("JWT_SECRET_KEY", "jwt-secret", true),
			HashSalt: env.GetEnvString("HASH_SALT", "hash-salt", true),
			HashCost: env.GetEnvNumber("HASH_COST", 12, true),
		},
		Redis: RedisConfig{
			Addr:     env.GetEnvString("REDIS_ADDR", "127.0.0.1:6379", false),
			Password: env.GetEnvString("REDIS_PASSWORD", "", false),
			DB:       env.GetEnvNumber("REDIS_DB", 0, false),
		},
		Storage: StorageConfig{
			Endpoint:        env.GetEnvString("MINIO_ENDPOINT", "", true),
			AccessKeyID:     env.GetEnvString("MINIO_ACCESS_KEY", "", true),
			SecretAccessKey: env.GetEnvString("MINIO_SECRET_KEY", "", true),
			UseSSL:          env.GetEnvString("MINIO_USE_SSL", "false", false) == "true",
			BucketName:      env.GetEnvString("MINIO_BUCKET_NAME", "", true),
			PresignedUrlTTL: time.Duration(env.GetEnvNumber("MINIO_PRESIGNED_URL_TTL", 600, false)) * time.Second,
		},
		RabbitMQ: RabbitMQConfig{
			Addr: env.GetEnvString("RABBITMQ_ADDR", "amqp://guest:guest@127.0.0.1:5672/", false),
		},
	}
}
