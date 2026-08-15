package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Environment string         `mapstructure:"ENVIRONMENT"`
	Server      ServerConfig   `mapstructure:"SERVER"`
	Database    DatabaseConfig `mapstructure:"DATABASE"`
	Redis       RedisConfig    `mapstructure:"REDIS"`
	JWT         JWTConfig      `mapstructure:"JWT"`
	Storage     StorageConfig  `mapstructure:"STORAGE"`
	Metrics     MetricsConfig  `mapstructure:"METRICS"`
	Provider    ProviderConfig `mapstructure:"PROVIDER"`
}

type ServerConfig struct {
	Port         int           `mapstructure:"PORT"`
	GRPCPort     int           `mapstructure:"GRPC_PORT"`
	ReadTimeout  time.Duration `mapstructure:"READ_TIMEOUT"`
	WriteTimeout time.Duration `mapstructure:"WRITE_TIMEOUT"`
}

type DatabaseConfig struct {
	Host            string        `mapstructure:"HOST"`
	Port            int           `mapstructure:"PORT"`
	User            string        `mapstructure:"USER"`
	Password        string        `mapstructure:"PASSWORD"`
	DBName          string        `mapstructure:"DB_NAME"`
	SSLMode         string        `mapstructure:"SSL_MODE"`
	MaxOpenConns    int           `mapstructure:"MAX_OPEN_CONNS"`
	MaxIdleConns    int           `mapstructure:"MAX_IDLE_CONNS"`
	ConnMaxLifetime time.Duration `mapstructure:"CONN_MAX_LIFETIME"`
}

func (db DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		db.Host, db.Port, db.User, db.Password, db.DBName, db.SSLMode)
}

type RedisConfig struct {
	Host     string `mapstructure:"HOST"`
	Port     int    `mapstructure:"PORT"`
	Password string `mapstructure:"PASSWORD"`
	DB       int    `mapstructure:"DB"`
}

type JWTConfig struct {
	Secret        string        `mapstructure:"SECRET"`
	AccessExpiry  time.Duration `mapstructure:"ACCESS_EXPIRY"`
	RefreshExpiry time.Duration `mapstructure:"REFRESH_EXPIRY"`
	Issuer        string        `mapstructure:"ISSUER"`
}

type StorageConfig struct {
	Driver          string `mapstructure:"DRIVER"` // local or s3
	LocalPath       string `mapstructure:"LOCAL_PATH"`
	S3Endpoint      string `mapstructure:"S3_ENDPOINT"`
	S3Bucket        string `mapstructure:"S3_BUCKET"`
	S3AccessKey     string `mapstructure:"S3_ACCESS_KEY"`
	S3SecretKey     string `mapstructure:"S3_SECRET_KEY"`
	S3UseSSL        bool   `mapstructure:"S3_USE_SSL"`
}

type MetricsConfig struct {
	Enabled bool   `mapstructure:"ENABLED"`
	Path    string `mapstructure:"PATH"`
	Port    int    `mapstructure:"PORT"`
}

type ProviderConfig struct {
	Mode               string `mapstructure:"MODE"` // local or real
	AWSAccessKeyID     string `mapstructure:"AWS_ACCESS_KEY_ID"`
	AWSSecretAccessKey string `mapstructure:"AWS_SECRET_ACCESS_KEY"`
	AWSRegion          string `mapstructure:"AWS_REGION"`
}

// LoadConfig loads configuration from path or environment variables.
func LoadConfig(path string) (*Config, error) {
	v := viper.New()

	v.SetDefault("ENVIRONMENT", "development")
	v.SetDefault("SERVER.PORT", 8080)
	v.SetDefault("SERVER.GRPC_PORT", 9090)
	v.SetDefault("SERVER.READ_TIMEOUT", 15*time.Second)
	v.SetDefault("SERVER.WRITE_TIMEOUT", 15*time.Second)

	v.SetDefault("DATABASE.HOST", "localhost")
	v.SetDefault("DATABASE.PORT", 5432)
	v.SetDefault("DATABASE.USER", "anarva_admin")
	v.SetDefault("DATABASE.PASSWORD", "anarva_password")
	v.SetDefault("DATABASE.DB_NAME", "anarva_cloud_db")
	v.SetDefault("DATABASE.SSL_MODE", "disable")
	v.SetDefault("DATABASE.MAX_OPEN_CONNS", 25)
	v.SetDefault("DATABASE.MAX_IDLE_CONNS", 5)
	v.SetDefault("DATABASE.CONN_MAX_LIFETIME", 15*time.Minute)

	v.SetDefault("REDIS.HOST", "localhost")
	v.SetDefault("REDIS.PORT", 6379)
	v.SetDefault("REDIS.PASSWORD", "")
	v.SetDefault("REDIS.DB", 0)

	v.SetDefault("JWT.SECRET", "anarva_cloud_db_super_secret_jwt_key_2026")
	v.SetDefault("JWT.ACCESS_EXPIRY", 24*time.Hour)
	v.SetDefault("JWT.REFRESH_EXPIRY", 7*24*time.Hour)
	v.SetDefault("JWT.ISSUER", "anarva-cloud-db")

	v.SetDefault("STORAGE.DRIVER", "local")
	v.SetDefault("STORAGE.LOCAL_PATH", "./data/storage")

	v.SetDefault("METRICS.ENABLED", true)
	v.SetDefault("METRICS.PATH", "/metrics")
	v.SetDefault("METRICS.PORT", 9091)

	v.SetDefault("PROVIDER.MODE", "local")
	v.SetDefault("PROVIDER.AWS_REGION", "us-east-1")

	if path != "" {
		v.AddConfigPath(path)
		v.SetConfigName("config")
		v.SetConfigType("yaml")

		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, fmt.Errorf("error reading config file: %w", err)
			}
		}
	}

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into config struct: %w", err)
	}

	if strings.ToLower(cfg.Environment) == "production" && strings.ToLower(cfg.Provider.Mode) == "real" {
		if cfg.Provider.AWSAccessKeyID == "" || cfg.Provider.AWSSecretAccessKey == "" {
			return nil, fmt.Errorf("PROVIDER_INVALID_CONFIGURATION: Production real provider mode requires valid AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables")
		}
	}

	return &cfg, nil
}
